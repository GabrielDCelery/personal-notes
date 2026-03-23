# Logging and Observability — Distilled

Observability is how you find out what your system is doing — not by guessing, but by asking the right question with the right signal.

## The Core Mental Model: Three Signals, Three Questions

Logging is the most instinctive part of observability, but it's only one of three signals. Using the wrong signal for a question is expensive, slow, and brittle.

```
Logs     → "What happened?"
           Events: errors, state changes, external calls.
           High detail, high cost, hard to aggregate.

Metrics  → "How much / how often?"
           Counters, gauges, histograms: error rate, latency, queue depth.
           Low cost, instantly queryable, no detail.

Traces   → "Where did the time go?"
           A request's journey across services, with timing per hop.
           Expensive to collect; invaluable for distributed latency debugging.
```

Metrics tell you _that_ something is wrong. Logs tell you _what_ went wrong. Traces tell you _where_ in the call chain it went wrong. **The most common mistake is using logs as a substitute for metrics** — counting error lines to compute an error rate is parsing text to answer a question a counter answers instantly, at a fraction of the cost.

## RED and USE: What to Actually Monitor

Two frameworks cover almost every monitoring concern. Apply them first before reaching for custom metrics.

**RED** applies to every service that handles requests:

| Signal       | What it measures             | Example                                                         |
| ------------ | ---------------------------- | --------------------------------------------------------------- |
| **Rate**     | Requests per second          | `http_requests_total` (counter)                                 |
| **Errors**   | Fraction of requests failing | `http_requests_errors_total` (counter)                          |
| **Duration** | How long requests take       | `http_request_duration_seconds` (histogram — track p50/p95/p99) |

**USE** applies to every resource (servers, DBs, queues, load balancers):

| Signal          | What it measures                      | Example                                       |
| --------------- | ------------------------------------- | --------------------------------------------- |
| **Utilization** | How busy, as a % of capacity          | CPU %, DB connection pool %                   |
| **Saturation**  | Work queued waiting to be processed   | Run queue length, DB wait events, queue depth |
| **Errors**      | Error events from the resource itself | Disk errors, network drops, OOM kills         |

Saturation is more informative than utilization alone. A CPU at 60% with a growing run queue is already over capacity — work is waiting. A CPU at 80% with no queue is fine.

Alerts come in two severities: **investigate** (low-urgency notification, someone checks it during business hours) and **page** (wakes up an on-call engineer immediately, expects a response within minutes). The thresholds below reflect that split.

```
Typical alert thresholds:
  RED — Error rate     > 1%  → investigate;  > 5%  → page
  RED — p99 latency    > 1s  → investigate;  > 5s  → page
  RED — Rate drop      > 20% vs 5-min average → possible upstream outage
  USE — CPU            > 80% sustained → scale or investigate
  USE — Memory         > 85%           → OOM risk
  USE — DB connections > 80% pool used → exhaustion approaching
  USE — Queue depth    growing > 5 min → consumers can't keep up
  USE — Disk           > 75%  → plan capacity;  > 90% → urgent
```

If a metric doesn't fit RED or USE, ask whether it belongs in a log instead.

## What to Log (and What Not to)

Every log line costs: ingestion, storage, and signal-to-noise ratio. Logging everything is as dangerous as logging nothing.

| Log these                         | Why                                                      |
| --------------------------------- | -------------------------------------------------------- |
| Errors and exceptions             | Stack trace + context is what lets you fix it            |
| Security events                   | Failed login, permission denied — often legally required |
| Slow operations (> 500 ms)        | Performance regression detection                         |
| State transitions                 | Order placed → paid → shipped → debugging stuck orders   |
| External calls (status + latency) | Blame attribution when a third party fails               |

| Never log these                    | Why not                                                   |
| ---------------------------------- | --------------------------------------------------------- |
| Passwords, tokens, API keys        | Logs are often less protected than databases              |
| Full request/response bodies       | PII risk, massive volume, rarely useful                   |
| High-frequency healthy operations  | Every 200 OK is noise that buries real signals            |
| Anything you'd compute as a metric | Use a counter — don't parse logs to answer rate questions |

After an incident, the instinct is to add more logging. The right response is to add _better_ logging — structured, contextual, for events you don't already have metrics for.

## Log Levels

Levels only work if used consistently. The key rule: **ERROR means an alert fires and a human looks at it.** If it doesn't need a human, it's WARN or lower.

| Level   | When to use                                         | Example                                   |
| ------- | --------------------------------------------------- | ----------------------------------------- |
| `ERROR` | Failed, requires attention. Alert should fire.      | DB connection lost, payment failed        |
| `WARN`  | Unexpected but recovered.                           | Retry succeeded, rate limit approaching   |
| `INFO`  | Normal operational events worth recording.          | Request completed, service started        |
| `DEBUG` | Diagnostic detail — not for production by default.  | SQL query text, intermediate state        |
| `TRACE` | Extremely granular — targeted troubleshooting only. | Every function call, every loop iteration |

**Default to INFO in production. Enable DEBUG temporarily and targeted — never globally.** Logging recoverable retries as ERROR trains engineers to ignore alerts. That's how real errors get missed.

## Structured Logging

Plain text logs are human-readable and machine-hostile. Structured logs — JSON — make every field a queryable column.

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
  "duration_ms": 312,
  "trace_id": "abc123def456"
}
```

With structured logs you can ask: "all payment failures in the last hour, grouped by reason." With unstructured logs you cannot — not reliably, not fast, not cheaply. **Always log JSON in production.**

Always include these fields on every log line:

| Field         | Why                                                                   |
| ------------- | --------------------------------------------------------------------- |
| `timestamp`   | UTC, ISO 8601                                                         |
| `level`       | For filtering                                                         |
| `service`     | Which service — essential in multi-service systems                    |
| `trace_id`    | Ties this line to logs in other services and to the distributed trace |
| `user_id`     | Who was affected                                                      |
| `event`       | A stable, machine-readable name for the event type                    |
| `duration_ms` | For any operation that can be slow                                    |

## Correlation IDs

In a multi-service system, one user request produces log lines across four services. Without a shared identifier, debugging means correlating timestamps and guessing.

A correlation ID is a UUID generated at the system boundary and propagated through every downstream call in a header:

```
User → API Gateway → generates trace_id: abc123
                   → passes X-Trace-Id: abc123 to all downstream calls

Auth Service    logs { trace_id: abc123, event: "token_validated" }
Order Service   logs { trace_id: abc123, event: "order_created", order_id: 9981 }
Payment Service logs { trace_id: abc123, event: "payment_failed", reason: "card_declined" }

Query: trace_id = abc123 → full sequence across all services
```

**Generate at the edge. Pass in every header. Log on every line. Never generate mid-request.** If you use OpenTelemetry, Jaeger, or AWS X-Ray, use the same ID for traces and logs — so you can jump from a log line directly into the full distributed trace.

## Log Aggregation

Containers are ephemeral. Logs written to a local file disappear on restart. The production pattern: **write to stdout, let the platform collect it, ship to a central store.**

```
App Container
  └── stdout
        └── Log Driver (awslogs, Fluentd, json-file)
              └── Aggregator (Fluent Bit, Vector)
                    └── Central Store
                          ├── CloudWatch Logs  (AWS-native)
                          ├── Elasticsearch    (full-text search)
                          ├── Loki             (Grafana stack, cheap)
                          └── Datadog          (all-in-one SaaS)
```

| Tool                | Best for                             | Notes                                     |
| ------------------- | ------------------------------------ | ----------------------------------------- |
| CloudWatch Logs     | AWS-native, zero setup               | Expensive at high volume; default for AWS |
| ELK (Elasticsearch) | Full-text search, complex queries    | Powerful; operationally heavy             |
| Loki                | Kubernetes, Grafana already in use   | Much cheaper than ES; limited query power |
| Datadog / New Relic | Logs + metrics + traces + APM in one | Expensive; best cross-signal correlation  |

**On AWS at moderate volume: CloudWatch.** Switch to Loki (cost) or Elasticsearch (query power) when it hurts. Use Datadog when you want everything correlated in one place.

## Cost and Retention

Log costs grow with traffic. Left unchecked they can rival compute costs. Match retention to actual need:

```
Hot   (0–7 days):      Full query, indexed, expensive   →  active debugging, recent incidents
Warm  (7–30 days):     Slower queries, cheaper storage  →  post-incident review
Cold  (30–365 days):   Compressed, S3 Glacier           →  compliance, forensic lookups
Beyond 1 year: delete unless regulation requires longer (GDPR, PCI, SOX: 1–7 years)
```

Metrics: 13 months (year-over-year comparison). Application logs: 7–14 days hot. Audit logs: per regulation.

To control volume without losing signal:

| Technique           | How                                                     | Trade-off                                    |
| ------------------- | ------------------------------------------------------- | -------------------------------------------- |
| Log level filtering | Ship INFO+ to prod; DEBUG only on explicit flag         | Loses detail when you need it most           |
| Sampling            | 1% of healthy requests; 100% of errors                  | Reduces volume 100×; safe for traffic trends |
| Field filtering     | Strip large fields before shipping                      | Loses detail; saves significant storage      |
| Metric extraction   | Extract counters at the aggregator; discard the raw log | Metrics for trends, logs for detail          |

**Sample high-volume healthy paths. Never sample errors.**

## Alerting: Logs vs Metrics

Alerts on log parsing are slow, brittle, and expensive. Alert on metrics; use log alerts only when the content is inherently textual.

| Use case                         | Signal    | Why                                                        |
| -------------------------------- | --------- | ---------------------------------------------------------- |
| Error rate > 1%                  | Metric    | Counter; instant query                                     |
| p99 latency > 500 ms             | Metric    | Histogram percentile                                       |
| Queue depth growing              | Metric    | Native gauge from the queue                                |
| Disk / memory approaching limit  | Metric    | System metric                                              |
| Specific error message appeared  | Log alert | No metric captures message text                            |
| "Payment failed: fraud detected" | Log alert | Specific textual event; warrants immediate human attention |

**Metrics for rates and thresholds. Log alerts only when the text content is the signal.**

## Key Mental Models

1. **Three signals, three questions.** Metrics: how much. Logs: what happened. Traces: where did time go. Use the right one.
2. **RED for services: Rate, Errors, Duration.** Three metrics per service = load, health, user experience.
3. **USE for infrastructure: Utilization, Saturation, Errors.** Saturation is more telling than utilization — a growing queue means you're already over capacity.
4. **Metrics are cheaper than logs for everything they can express.** Counting log lines to compute a rate is doing it wrong.
5. **Always log JSON in production.** Every field queryable. Unstructured logs are grep-or-bust.
6. **Correlation IDs are mandatory in multi-service systems.** Generate at the edge, propagate everywhere, log on every line.
7. **Write to stdout. Never to files in containers.** The platform collects it; you just configure the format.
8. **Sample healthy paths. Never sample errors.** Volume control without losing the signals that matter.
9. **Alert on metrics, not log parsing.** Log alerts only for specific textual events.
10. **Match retention to need.** App logs: 7–14 days. Audit logs: years. Metrics: 13 months.
11. **ERROR means a human looks at it.** If it doesn't need a human, it's WARN. Crying wolf kills alert trust.
12. **Context makes logs useful.** trace_id + user_id + event on every line. Without them, a log is noise.
