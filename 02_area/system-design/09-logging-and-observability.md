# Logging and Observability

Lesson 08 covered cost and storage lifecycle — how data ages and where to put it as it gets cheaper to store but more expensive to query. Observability is the same problem applied to operational data: what happened in your system, where did time go, and how do you find out when things go wrong.

Logging is the most instinctive part of observability — you print things and read them when something breaks. But raw logs at scale are unmanageable, expensive, and slow to query. The goal of this lesson is to build a complete mental model of observability: what logs are for, what they shouldn't be used for, how they fit alongside metrics and traces, and what good production logging looks like in practice.

## The Observability Triad

Observability isn't just logging. It's three complementary signals, each answering a different question:

```
Logs     → "What happened?"
           Individual events: errors, requests, state changes.
           High detail, high cost, hard to aggregate.

Metrics  → "How much / how often?"
           Counters, gauges, histograms: error rate, p99 latency, queue depth.
           Low cost, easy to aggregate, no detail.

Traces   → "Where did the time go?"
           A single request's journey across services, with timing for each hop.
           Expensive to collect and store, invaluable for latency debugging.
```

The three signals are not interchangeable. Metrics tell you _that_ something is wrong. Logs tell you _what_ went wrong. Traces tell you _where_ in the call chain it went wrong. A mature system uses all three — and each for the right purpose.

The common mistake is using logs as a substitute for metrics. If you're counting error log lines to compute an error rate, you're parsing text to answer a question that a counter would answer instantly. Metrics are orders of magnitude cheaper to query than logs. Use the right signal for the question.

## What to Log (and What Not to)

Every log line has a cost: storage, ingestion, and most importantly, the noise that makes real signals harder to find. Logging everything is as dangerous as logging nothing. The discipline is knowing what belongs in a log.

### Log these

| Event type                        | Example                                             | Why                                            |
| --------------------------------- | --------------------------------------------------- | ---------------------------------------------- |
| Errors and exceptions             | Unhandled exception, DB connection failure          | You need the stack trace and context to fix it |
| Security events                   | Failed login, permission denied, token expired      | Audit trail — often legally required           |
| Slow operations                   | Query > 500 ms, request > 2 s                       | Performance regression detection               |
| State transitions                 | Order placed → paid → shipped → delivered           | Debugging "why is this order stuck?"           |
| External calls (request/response) | Outbound HTTP to payment provider, status + latency | Blame attribution when a third party fails     |
| Application startup/shutdown      | Service started, config loaded, graceful shutdown   | Operational baseline                           |

### Do not log these

| What                               | Why not                                                                   |
| ---------------------------------- | ------------------------------------------------------------------------- |
| Passwords, tokens, API keys        | Security — logs are often less protected than databases                   |
| Full request/response bodies       | PII risk, massive volume, rarely useful                                   |
| High-frequency healthy operations  | Every cache hit, every 200 OK — noise that buries real signals            |
| Anything you'd compute as a metric | Request count, error rate, latency histogram — use a counter or histogram |

The instinct after an incident is to add more logging. The right response is to add _better_ logging — structured, contextual, and for events you don't already have metrics for.

## Log Levels

Log levels exist to filter signal from noise. They're only useful if used consistently.

| Level   | When to use                                                        | Example                                   |
| ------- | ------------------------------------------------------------------ | ----------------------------------------- |
| `ERROR` | Something failed and requires attention. An alert should fire.     | DB connection lost, payment failed        |
| `WARN`  | Something unexpected happened but the system recovered.            | Retry succeeded, rate limit approaching   |
| `INFO`  | Normal operational events worth recording.                         | Request completed, service started        |
| `DEBUG` | Detailed diagnostic information — not for production by default.   | SQL query text, intermediate calculation  |
| `TRACE` | Extremely granular — rarely used outside targeted troubleshooting. | Every function call, every loop iteration |

**Default to INFO in production. Enable DEBUG temporarily and targeted — never globally in production.** A service emitting DEBUG logs at scale will saturate your log aggregator and run up your bill.

Level discipline also means not crying wolf. If an event is `ERROR`, an alert fires. If you log recoverable retries as `ERROR`, on-call engineers stop trusting the alerts. Reserve `ERROR` for things that genuinely need a human.

## Structured Logging

Plain text logs are human-readable and machine-hostile. To search `app.log` for all errors from user `12345` in the last hour, you need grep and regex and hope the format is consistent. Structured logs — JSON by default — make every field a queryable column.

```
# Unstructured — grep or bust
2024-01-15 14:23:01 ERROR Payment failed for user 12345: card declined

# Structured — every field is queryable
{
  "timestamp": "2024-01-15T14:23:01Z",
  "level": "error",
  "event": "payment_failed",
  "user_id": "12345",
  "reason": "card_declined",
  "amount_cents": 4999,
  "currency": "GBP",
  "duration_ms": 312,
  "trace_id": "abc123def456"
}
```

With structured logs you can ask: "all payment failures in the last hour, grouped by reason, for users in the EU." With unstructured logs, you cannot — not reliably, not fast, not cheaply.

**Always log in JSON in production.** The verbosity cost is trivial. The query cost savings are enormous.

### What fields to always include

| Field         | Why                                                                 |
| ------------- | ------------------------------------------------------------------- |
| `timestamp`   | When it happened — in UTC, ISO 8601                                 |
| `level`       | Severity — for filtering                                            |
| `service`     | Which service emitted this — essential in multi-service systems     |
| `trace_id`    | Ties this log line to a request trace and to logs in other services |
| `user_id`     | Who was affected — for incident investigation                       |
| `event`       | A stable, machine-readable name for the event type                  |
| `duration_ms` | For any operation that can be slow                                  |

## Correlation IDs and Request Tracing

In a single-service system, a log line tells you what happened. In a multi-service system, a single user request touches the API service, the auth service, the order service, and the payment service. When payment fails, the log line is in the payment service. The context — what the user did, what the API received, what order service sent — is in four different log streams.

A correlation ID (also called a trace ID or request ID) is a UUID generated at the system boundary — the load balancer or the first service to receive the request — and propagated through every downstream call.

```
User → API Gateway → generates trace_id: abc123
                   → passes X-Trace-Id: abc123 to all downstream calls

Auth Service   logs { trace_id: abc123, event: "token_validated" }
Order Service  logs { trace_id: abc123, event: "order_created", order_id: 9981 }
Payment Service logs { trace_id: abc123, event: "payment_failed", reason: "card_declined" }

Query: trace_id = abc123 → full picture of everything that happened for this request
```

Without correlation IDs, debugging a distributed system means correlating timestamps and guessing. With them, you can reconstruct the exact sequence of events across every service for any request.

**Generate the trace ID at the edge. Pass it in a header (X-Trace-Id or the W3C traceparent standard). Log it on every log line. Never generate it mid-request.**

If you're using a tracing system like OpenTelemetry, Jaeger, or AWS X-Ray, the trace ID is the same ID — your logs and your traces share the same identifier, so you can jump from a log line to the full distributed trace.

## Log Aggregation

A container that writes logs to a local file has already lost the game. Containers are ephemeral — when they restart, the logs are gone. On ECS or Kubernetes, a task can be scheduled on any node; logs scattered across nodes are unqueryable.

The production pattern is: **write to stdout, let the platform collect it, ship it to a central store.**

```
App Container
  └── writes to stdout
        └── Log Driver (Docker json-file, awslogs, Fluentd)
              └── Log Aggregator (Fluentd, Fluent Bit, Vector)
                    └── Central Store
                          ├── Elasticsearch / OpenSearch (ELK stack)
                          ├── CloudWatch Logs (AWS)
                          ├── Loki (Grafana stack)
                          └── Datadog / Splunk / New Relic (SaaS)
```

**Write to stdout. Never write to files in a container.** The platform — ECS, Kubernetes, Lambda — handles collection. You configure the log driver, not the destination path.

### Technology comparison

| Tool                | Best for                                  | Cost model                                 | Notes                                          |
| ------------------- | ----------------------------------------- | ------------------------------------------ | ---------------------------------------------- |
| CloudWatch Logs     | AWS-native, simple setup                  | Per GB ingested + stored                   | Expensive at high volume; fine for low traffic |
| ELK (Elasticsearch) | Full-text search, complex queries         | Self-hosted infra cost                     | Powerful but operationally heavy               |
| Loki (Grafana)      | Kubernetes, already using Grafana         | Storage only (labels, not full text index) | Much cheaper than ES; limited query power      |
| Datadog / New Relic | All-in-one: logs + metrics + traces + APM | Per host or per GB                         | Expensive; excellent UX and correlation        |
| Splunk              | Enterprise, compliance, security          | Per GB ingested                            | Very expensive; very powerful                  |

**On AWS with moderate volume: CloudWatch Logs.** It's already there, requires zero setup, and integrates with Lambda and ECS natively. Switch when the bill hurts or queries become too limited. At high volume: Loki (cheap) or Elasticsearch (powerful). At any scale where you want everything correlated: Datadog.

## Cost and Retention

Log costs are two-dimensional: ingestion (paying per GB sent to the store) and storage (paying per GB retained). Both grow with log volume, which grows with traffic. Left unchecked, logging costs can rival compute costs.

### The retention model

Most logs are only useful for a short window. A request log from 90 days ago is almost never needed. An audit log from 2 years ago may be legally required.

```
Hot (0–7 days):   Full query access. Indexed. Expensive.
                  → Recent incidents, active debugging

Warm (7–30 days): Full access but slower queries, cheaper storage.
                  → Post-incident review, trend analysis

Cold (30–365 days): Compressed, archived (S3 Glacier).
                    → Compliance, rare forensic lookups

Delete or archive beyond 1 year unless compliance requires longer.
```

Match retention to need. Most application logs: 7–14 days hot. Audit/security logs: 1–7 years depending on regulation (GDPR, PCI, SOX). Time-series metrics: 13 months for year-over-year comparison.

### Controlling volume

| Technique           | How                                                           | Trade-off                                          |
| ------------------- | ------------------------------------------------------------- | -------------------------------------------------- |
| Log level filtering | Ship INFO+ to prod; DEBUG only on explicit flag               | Loses debug detail when you need it most           |
| Sampling            | Log 1% of healthy requests; 100% of errors                    | Reduces volume 100×; fine for traffic patterns     |
| Field filtering     | Strip large fields (full request body) before shipping        | Loses detail; saves significant storage            |
| Metric extraction   | Extract counters from logs at the aggregator; discard the log | Best of both — metrics for trends, logs for detail |

**Sample high-volume healthy paths. Never sample errors.** A 1% sample of 200 OK responses gives you accurate traffic patterns. Missing even one error log during an incident is expensive.

## RED and USE: What to Actually Monitor

Knowing that metrics are better than logs for rates and thresholds is half the answer. The other half is knowing which metrics to actually track. Two frameworks cover almost every service and infrastructure concern.

### RED — for services

RED applies to every service that handles requests. For each service, track three things:

| Signal       | What it measures                       | Example metric                                                 |
| ------------ | -------------------------------------- | -------------------------------------------------------------- |
| **Rate**     | How many requests per second           | `http_requests_total` (counter)                                |
| **Errors**   | How many of those requests are failing | `http_requests_errors_total` (counter)                         |
| **Duration** | How long requests take (latency)       | `http_request_duration_seconds` (histogram, track p50/p95/p99) |

Rate tells you load. Errors tell you health. Duration tells you user experience. These three metrics, on every service, give you a complete picture of whether your system is working — without logs.

```
Alert thresholds (typical starting points):
  Error rate   > 1% → investigate; > 5% → page
  p99 latency  > 1s → investigate; > 5s → page
  Rate drop    > 20% vs 5-min average → possible outage upstream
```

### USE — for infrastructure

USE applies to every resource: servers, databases, load balancers, queues. For each resource, track:

| Signal          | What it measures                                | Example                                       |
| --------------- | ----------------------------------------------- | --------------------------------------------- |
| **Utilization** | How busy is it, as a percentage of capacity     | CPU %, disk I/O %, DB connection pool %       |
| **Saturation**  | How much work is queued waiting to be processed | Run queue length, DB wait events, queue depth |
| **Errors**      | Error events from the resource itself           | Disk errors, network drops, OOM kills         |

Utilization tells you how close to the limit you are. Saturation tells you whether you're already over — requests are waiting, not just arriving. A CPU at 80% utilisation with no run queue is fine. A CPU at 60% utilisation with a growing run queue means something is blocking and tasks are piling up.

```
Typical alert thresholds:
  CPU utilisation    > 80% sustained → scale or investigate
  Memory utilisation > 85%          → risk of OOM; investigate
  DB connection pool > 80% used     → connection exhaustion approaching
  Queue saturation   growing > 5m   → consumers can't keep up (see lesson 05)
  Disk utilisation   > 75%          → plan capacity; > 90% → urgent
```

**RED for services, USE for resources.** Together they cover almost everything worth alerting on. If a metric doesn't fit either framework, ask whether it's a metric at all or whether it belongs in a log.

## Alerting: Logs vs Metrics

Alerts should fire on metrics, not log parsing. Parsing log streams to detect error rates introduces latency (you're scanning text), is brittle (log format changes break your alert), and is expensive (querying logs is slow and costly).

| Use case                         | Right signal | Why                                                       |
| -------------------------------- | ------------ | --------------------------------------------------------- |
| Error rate > 1%                  | Metric       | Counter; instant query; doesn't require log parsing       |
| p99 latency > 500 ms             | Metric       | Histogram; accurate percentiles from counters             |
| Specific error message appeared  | Log alert    | You need the content; no metric captures the message text |
| Queue depth growing              | Metric       | The queue exposes this as a gauge natively                |
| "Payment failed: fraud detected" | Log alert    | Specific event that warrants immediate human attention    |
| Disk / memory approaching limit  | Metric       | System metric; no log needed                              |

The mental model: **metrics for rates and thresholds; logs for specific events that are inherently textual.** If you're writing a regex to match log lines to trigger an alert, ask whether a counter or gauge would express the same thing better.

## Common Patterns

### Centralised request logging at the edge

Log every inbound request at the load balancer or API gateway: method, path, status code, latency, user ID, trace ID. This gives you traffic patterns, error rates, and latency distributions without touching application code. Application logs then focus on the _why_, not the _what_.

### Error context enrichment

When an exception is caught, log the full context, not just the message. The stack trace tells you where. The context tells you what the system was doing:

```json
{
  "level": "error",
  "event": "payment_failed",
  "error": "CardDeclinedException: insufficient funds",
  "stack": "...",
  "user_id": "12345",
  "order_id": "9981",
  "amount_cents": 4999,
  "payment_provider": "stripe",
  "attempt": 2,
  "trace_id": "abc123"
}
```

A log line with just `"error": "CardDeclinedException"` tells you nothing about which user, which order, or whether this was a retry. Context is what turns a log into a debugging tool.

### Slow query logging

Configure the database to log queries above a threshold (Postgres: `log_min_duration_statement = 500`). These logs, shipped to your aggregator, let you find slow queries in production without instrumenting application code. Far more reliable than trying to reproduce slow queries in development.

## Key Mental Models

1. **Logs, metrics, traces are not interchangeable.** Metrics for rates and thresholds; logs for events; traces for distributed latency. Use the right signal for the question.
2. **RED for services: Rate, Errors, Duration.** Three metrics per service gives you load, health, and user experience.
3. **USE for infrastructure: Utilization, Saturation, Errors.** Saturation is more informative than utilization alone — a growing queue means you're already over capacity.
4. **Metrics are cheaper than logs for everything they can express.** If you're counting log lines to compute a rate, you're doing it wrong.
5. **Structure everything.** JSON logs in production — every field queryable, every log comparable.
6. **Correlation IDs are mandatory in multi-service systems.** Generate at the edge, propagate everywhere, log on every line.
7. **Write to stdout. Never to files in containers.** The platform handles collection. You handle format.
8. **Sample high-volume healthy paths. Never sample errors.** Volume control without losing the signals that matter.
9. **Alert on metrics, not log parsing.** Log alerts only for specific textual events with no metric equivalent.
10. **Match retention to need.** Application logs: 7–14 days. Audit logs: years. Metrics: 13 months. Delete the rest.
11. **Log level discipline prevents alert fatigue.** ERROR means alert. If it doesn't need a human, it's WARN.
12. **Context makes logs useful.** A log line without user ID, trace ID, and event name is noise. A log line with them is a debugging tool.
