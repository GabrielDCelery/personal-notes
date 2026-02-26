# Datadog Lambda Layer — Deep Dive

## What is a Trace?

A trace represents a single end-to-end request as it flows through your system. Think of it as a tree of work.

```
Trace ID: abc-123
│
├── Span: API Gateway (5ms)
│   └── Span: Lambda A — handler (200ms)
│       ├── Span: DynamoDB GetItem (12ms)
│       └── Span: Lambda B invoke (80ms)
│           └── Span: Lambda B — handler (75ms)
│               └── Span: S3 PutObject (30ms)
```

Every box above is a **span**.

---

## What is a Span?

A span is a single unit of work. It records:

- `trace_id` — which trace it belongs to
- `span_id` — its own unique ID
- `parent_span_id` — who called it
- `name` — what it is (e.g. `aws.lambda`)
- `start_time` / `duration`
- `resource` — the specific thing (e.g. `GET /users`)
- `status` — ok / error
- arbitrary `tags` / metadata

Spans are **not log lines**. They are structured objects that live in memory, get sent to the local extension agent over a Unix socket, and are flushed directly to Datadog's trace intake API.

---

## How Does the Trace ID Travel Between Lambdas?

Context propagation via headers or message attributes.

### HTTP calls

The calling service injects trace context into HTTP headers:

```
X-Datadog-Trace-Id: 1234567890
X-Datadog-Parent-Id: 9876543210
X-Datadog-Sampling-Priority: 1
```

The receiving Lambda reads those headers and continues the same trace.

### SQS / SNS

Datadog injects trace context into the message attributes of each message. The consumer Lambda reads those attributes to reconnect the trace.

### Direct Lambda invocation

Context goes into the invocation payload or client context.

If no context is found → a new trace starts (orphaned trace).

---

## What is the Datadog Lambda Layer?

The layer is a ZIP attached to your Lambda at deploy time. It contains two things:

```
/opt/
  extensions/
    datadog-agent        ← Datadog Agent running as a Lambda Extension
  python/ or nodejs/
    datadog_lambda/      ← DD Lambda library (wraps your handler)
```

### 1. The Lambda Extension (the Agent)

Runs as a separate process alongside your function in the same execution environment.

- Collects metrics, traces, logs from your function
- Flushes them to Datadog before the execution environment freezes (uses Lambda's extension lifecycle hooks)
- Avoids needing to make network calls inside your handler's hot path
- Reads Lambda runtime metrics from the Lambda Telemetry API (an AWS-provided local HTTP endpoint)

### 2. The Lambda Library (e.g. `dd-trace`, `datadog-lambda-py`)

Wraps your handler function. It:

- Creates the root span for the Lambda invocation
- Extracts trace context from the incoming event (headers, SQS attributes, etc.)
- Patches AWS SDK, HTTP clients, etc. to auto-instrument outgoing calls (monkey-patching)
- Patches your logging library to inject `dd.trace_id` and `dd.span_id` into log output
- Sends completed spans to the local Agent via Unix socket

---

## How Monkey Patching Works

The library uses monkey-patching at import/require time. Modules are cached singletons in memory — mutate the object once at startup and every piece of code that imported that module sees the replacement.

```python
# Save original
original_make_api_call = botocore.client.BaseClient._make_api_call

# Define wrapper
def patched_make_api_call(self, operation_name, api_params):
    span = tracer.start_span(f"aws.{operation_name}")
    try:
        result = original_make_api_call(self, operation_name, api_params)
        span.finish()
        return result
    except Exception as e:
        span.error(e)
        span.finish()
        raise

# Replace on the class — affects all instances everywhere
botocore.client.BaseClient._make_api_call = patched_make_api_call
```

All AWS SDK clients (DynamoDB, S3, SQS, etc.) inherit from `BaseClient` and funnel through `_make_api_call` — so one patch covers the entire AWS SDK. No wrapping needed in your code.

---

## What the Layer Records

### Spans (via monkey-patching)

Automatically created for every:

- DynamoDB call (GetItem, PutItem, Query, etc.)
- S3 call (GetObject, PutObject, etc.)
- SQS call (SendMessage, ReceiveMessage, etc.)
- HTTP calls
- Any other auto-instrumented library

Each span captures: operation name, table/bucket/queue name, AWS region, HTTP status, duration, error status.

### Enhanced Lambda Metrics (via Lambda Telemetry API)

| Metric                     | What it is                |
| -------------------------- | ------------------------- |
| `aws.lambda.invocations`   | count of invocations      |
| `aws.lambda.duration`      | how long your handler ran |
| `aws.lambda.errors`        | unhandled exceptions      |
| `aws.lambda.cold_starts`   | cold start count          |
| `aws.lambda.timeout`       | timed out invocations     |
| `aws.lambda.out_of_memory` | OOM kills                 |

### Custom Metrics (manual)

```python
from datadog_lambda.metric import lambda_metric
lambda_metric("orders.processed", 1, tags=["env:prod"])
```

---

## Log Enrichment vs Log Forwarding

These are two separate things people often conflate.

**Log enrichment** — the library patches your logging library so every log line gets `dd.trace_id` and `dd.span_id` stamped on it:

```json
{
  "message": "Processing order 42",
  "dd": {
    "trace_id": "1234567890123456789",
    "span_id": "9876543210987654321",
    "service": "order-service",
    "env": "prod"
  }
}
```

This happens regardless of how logs are forwarded. The enrichment is done by the library in your function process.

**Log forwarding** — the extension intercepting stdout and sending logs directly to Datadog. Controlled by `enableDatadogLogs`. If disabled, logs fall back to `lambda → cloudwatch → forwarder → datadog`.

Because enrichment is separate from forwarding, even with `enableDatadogLogs: false`, your logs still contain `dd.trace_id` and Datadog can still correlate them to traces when the forwarder delivers them.

---

## The Full Picture

```
Your Lambda process:
  ├── library patches AWS SDK
  │     → span objects created in memory
  │     → sent to extension via Unix socket
  │
  ├── library patches your logger
  │     → log lines get dd.trace_id stamped on them
  │     → written to stdout
  │
  └── extension (separate process):
        ├── receives spans → flushes to Datadog trace API
        ├── reads Lambda Telemetry API → flushes metrics to Datadog metrics API
        └── (if enableDatadogLogs: true) intercepts stdout → flushes to Datadog log API
```

---

## The CloudWatch Extension Error Problem

### The duplicate log paths problem

When the Datadog Lambda Layer is set up alongside the old log forwarder, logs flow through two paths simultaneously:

```
lambda → cloudwatch → forwarder → datadog   (old, automatic)
lambda → layer extension → datadog          (new)
```

The old forwarder auto-subscribes to every Lambda log group it finds. You can't stop it.

### Why denying CloudWatch breaks things

The workaround is to deny `logs:*` via IAM to kill the CloudWatch path. But the next-gen extension (Rust-based, default from v88+) tries to do CloudWatch operations itself — gets denied — and throws extension errors in Datadog.

Compatibility mode (non-Rust) didn't have this problem, but it was dropped from v88+.

### Scenarios

| Setup                                        | Duplicate logs | Extension errors |
| -------------------------------------------- | -------------- | ---------------- |
| Layer + compatibility mode + deny CloudWatch | No             | No               |
| Layer + next-gen extension + deny CloudWatch | No             | Yes              |
| Layer + next-gen extension + no deny         | Yes            | No               |

### Solutions

**Short-term: `enableDatadogLogs: false`**

Disable layer log forwarding. Logs go via CloudWatch → forwarder only (single path). No duplication, no extension errors. Layer still handles traces and metrics fully. Log enrichment (trace_id injection) still happens in the log lines themselves, so correlation still works via the forwarder.

```typescript
const dataDogLambdaLayer = new DatadogLambda(this, "DataDogLambdaLayer", {
  enableDatadogLogs: false,
});
// remove denyCloudwatchLogging calls
```

**Medium-term: `DD_EXCLUDE_AT_MATCH` on the forwarder**

The Datadog Forwarder Lambda has an env var that takes a regex of log group names to skip:

```
DD_EXCLUDE_AT_MATCH = /aws/lambda/my-service.*
```

Set this on the forwarder to stop it subscribing to layer-enabled services. Re-enable `enableDatadogLogs: true`. Remove the CloudWatch deny policy. Clean single path via the layer.

**Long-term: explicit log groups + tags**

CDK never creates log groups explicitly — Lambda creates them automatically on first run, at which point the auto-subscription has already fired. Create them explicitly in CDK before deploy:

```typescript
const logGroup = new logs.LogGroup(this, "MyLambdaLogs", {
  logGroupName: `/aws/lambda/${myLambda.functionName}`,
  removalPolicy: RemovalPolicy.DESTROY,
});
cdk.Tags.of(logGroup).add("dd-managed-by-layer", "true");
```

Then modify the auto-subscription trigger to check for that tag before subscribing. Most surgical solution but requires diligence across all Lambdas.

### Recommended path

```
Now:   enableDatadogLogs: false + remove deny policy  → stops the bleeding
Soon:  DD_EXCLUDE_AT_MATCH on the forwarder           → proper fix, low risk
Later: explicit log groups + tags everywhere           → clean architecture
```
