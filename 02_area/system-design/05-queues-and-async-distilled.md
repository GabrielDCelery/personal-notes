# Queues and Async Processing — Distilled

Queues exist to separate "accepting work" from "doing work" so users don't wait for things they don't need to wait for.

## The Core Decision: Does the User Need the Result Now?

This single question determines everything. Sync means block until done. Async means accept the intent, return immediately, process later. The difference in practice is dramatic:

```
Without queue (sync):

  User → API → charge card → update DB → send email → notify warehouse → 200 OK
                                                                          700 ms
  If email is down, the whole request fails.

With queue (async):

  User → API → save order → publish to queue → 202 Accepted
                                                50 ms
                     ↓
               Queue → Worker A: charge card   (200 ms)
                     → Worker B: send email    (300 ms)
                     → Worker C: notify WH     (150 ms)

  If email is down, the message stays in queue and retries. Order is safe.
```

The total work is identical — it just no longer blocks the user. Think of it as the difference between a food truck (you wait at the window) and a sit-down restaurant (you get an order number and sit down).

**Go async when:**

- The result isn't needed to continue (notifications, analytics, audit logs)
- Multiple independent steps follow one trigger (order → payment + inventory + email + warehouse)
- The operation can fail and needs to be retried without the user waiting (warehouse notification, not payment confirmation)
- Traffic is spiky and the downstream can't absorb bursts

**Stay sync when:**

- The user needs the result to continue (login, search, balance check)
- The operation is fast (< 100 ms) — a queue adds overhead, not value
- Consistency must be immediate ("is this username taken?" can't be deferred)

Every queue adds a consumer to run, a DLQ to monitor, and delay the user can perceive. Don't add one unless you need it.

## Queue Fundamentals

A **producer** publishes messages; a **consumer** reads and processes them. The producer doesn't know who the consumer is, how long it takes, or whether the consumer is even running. That decoupling is the point.

| Concept                | What it means                                                                        |
| ---------------------- | ------------------------------------------------------------------------------------ |
| **Consumer group**     | A set of consumers sharing the work — each message goes to one consumer in the group |
| **Acknowledgement**    | Consumer tells the queue "done, delete this"                                         |
| **Visibility timeout** | Message is hidden after pickup (SQS); if no ack arrives in time, it reappears        |
| **Dead letter queue**  | Where messages go after failing too many times                                       |
| **Backpressure**       | Producers faster than consumers — the queue grows                                    |
| **Offset**             | Consumer's read position in a partition (Kafka) — "I've processed up to here"        |

## Technology Comparison

Three technologies dominate, each built for a different problem.

**Kafka** is a distributed commit log. Messages append to partitions and persist — consumers read by offset, not by destructive dequeue. Multiple independent consumer groups can read the same data, and you can replay old messages. The log model is its defining property: data is history, not a task list.

**SQS** is a managed queue on AWS. No brokers, no servers, no configuration. A message is deleted after acknowledgement — no replay, no shared reads across consumer groups. Simplicity is the feature.

**RabbitMQ** is a message broker with a routing layer. Messages pass through an exchange that routes them to queues based on rules (direct, fanout, topic pattern, headers). Most flexible routing; lower throughput ceiling than Kafka.

|                  | Kafka                                | SQS                            | RabbitMQ                     |
| ---------------- | ------------------------------------ | ------------------------------ | ---------------------------- |
| **Model**        | Distributed log                      | Managed queue                  | Message broker               |
| **Throughput**   | 100K–1M+/broker                      | Unlimited (std) / 300/s (FIFO) | 20K–50K/node                 |
| **Ordering**     | Per partition                        | Best-effort (FIFO option)      | Per queue                    |
| **Replay**       | Yes — re-read from any offset        | No                             | No                           |
| **Retention**    | Days / weeks / forever               | Up to 14 days                  | Until ack'd                  |
| **Routing**      | Topics + partitions                  | Single queue (fan-out via SNS) | Exchanges, complex patterns  |
| **Ops overhead** | High (brokers, KRaft)                | Zero (managed)                 | Medium (Erlang cluster)      |
| **Best for**     | Event streaming, high volume, replay | Task queues on AWS             | Complex routing, multi-cloud |

**The quick pick:**

```
On AWS, simple task queue             → SQS (don't overthink it)
High throughput / replay needed       → Kafka
Multiple systems need the same event  → Kafka (consumer groups) or SNS+SQS fan-out
Complex routing (fanout, patterns)    → RabbitMQ
On-prem or multi-cloud                → RabbitMQ or self-hosted Kafka
Volume < 1K/sec on AWS                → SQS
```

## Delivery Guarantees

| Guarantee         | Meaning                                  | Risk                                      | Cost    |
| ----------------- | ---------------------------------------- | ----------------------------------------- | ------- |
| **At-most-once**  | Delivered 0 or 1 times — fire and forget | Message lost if consumer crashes          | Fastest |
| **At-least-once** | Always retried on failure                | Consumer may process same message twice   | Middle  |
| **Exactly-once**  | Delivered exactly once                   | Extremely hard across distributed systems | Slowest |

**At-least-once is the default for almost everything** — Kafka, SQS, and RabbitMQ all default to it. You never lose messages and you handle duplicates on the consumer side.

True exactly-once delivery across distributed systems is not achievable. The problem is inescapable:

```
Producer → Broker: "Here's message X"
Broker saves X, sends ack
Network drops the ack
Producer: "No ack received — resend"
Broker now has X twice
```

The solution isn't exactly-once delivery — it's **idempotent consumers**: design processing so that running the same message twice has the same effect as running it once.

| Pattern                               | How it works                                                               |
| ------------------------------------- | -------------------------------------------------------------------------- |
| **Idempotency key**                   | Store processed message IDs; skip if already seen                          |
| **Upsert**                            | `INSERT ON CONFLICT` / PUT instead of increment — same result if run twice |
| **Idempotency key on external calls** | Stripe's `Idempotency-Key` — same key = same charge, not a double charge   |
| **Version / timestamp check**         | Only apply if event timestamp > last processed timestamp                   |

**Assume every message will be delivered at least twice. Design accordingly.** Idempotent consumers give you effectively exactly-once behaviour with at-least-once delivery.

## Consumer Scaling

### Kafka: partitions cap parallelism

In Kafka, **one partition can only be read by one consumer per consumer group**. This is a hard constraint that you must plan for upfront.

```
6 partitions, 6 consumers (ideal):    6 partitions, 9 consumers (3 idle):

  C1 → P0                               C1 → P0
  C2 → P1                               C2 → P1
  C3 → P2                               ...
  C4 → P3                               C6 → P5
  C5 → P4                               C7 → idle
  C6 → P5                               C8 → idle
                                         C9 → idle
```

6 partitions = max 6 consumers processing in parallel. Choose a partition count that matches your expected peak parallelism. Common starting points: 6 (low traffic), 12–30 (medium), 50–100+ (high throughput). Increasing partitions later causes a rebalance and breaks ordering guarantees for in-flight data.

### SQS: scale freely

SQS has no partition concept. Any number of consumers can poll the same queue. The common AWS pattern is SQS + Lambda — AWS scales Lambda invocations automatically based on queue depth, up to 1,000 concurrent executions by default. No partition planning, no consumer management.

```
Queue depth:  0–100    → 2 consumers
              100–1K   → 5 consumers
              1K+      → 20 consumers

Metric: ApproximateNumberOfMessagesVisible
```

## Backpressure and Monitoring

Backpressure is producers outrunning consumers. The queue grows, messages age, and eventually the broker runs out of disk or messages expire before processing.

```
Normal:       Producer (100/s) → ▁ Queue ▁ → Consumer (100/s)

Backpressure: Producer (500/s) → ▁▂▄▆█ Queue growing → Consumer (100/s)
```

Strategies — in order of preference:

1. **Absorb short spikes** — the queue is a buffer; if the spike is temporary, consumers catch up
2. **Auto-scale consumers** — scale on queue depth, not CPU
3. **Rate-limit producers** — return 429, push backpressure upstream to callers
4. **Shed low-priority messages** — drop what you can afford to lose

| Metric                     | Alert when                             |
| -------------------------- | -------------------------------------- |
| Queue depth / consumer lag | Growing for > 5 minutes                |
| Message age                | Exceeds your SLA (30 sec, 5 min, etc.) |
| Consumer error rate        | > 1–5%                                 |
| DLQ depth                  | Any message — investigate immediately  |

## Dead Letter Queues

A poison message — one that always fails — will cycle indefinitely or block the queue without a DLQ. Always configure one.

```
Normal:  Queue → Consumer → success → ack → deleted

Failure: Queue → Consumer → fail → back to queue
              → Consumer → fail → back to queue  (×3–5 retries)
              → max retries → DLQ

DLQ: alert → investigate → fix the bug → replay
```

| System   | Config                                                 | Typical     |
| -------- | ------------------------------------------------------ | ----------- |
| SQS      | `maxReceiveCount` + `deadLetterTargetArn`              | 3–5 retries |
| RabbitMQ | `x-dead-letter-exchange`                               | 3–5 retries |
| Kafka    | No built-in — publish to `.dlq` topic in consumer code | —           |

DLQ messages represent lost work or broken assumptions. Every message in a DLQ needs a plan: alert → investigate → fix → replay. Never leave them unattended.

## Common Patterns

**Fan-out** — one event, multiple independent consumers. On AWS: SNS → multiple SQS queues. With Kafka: multiple consumer groups reading the same topic.

```
Order placed → SNS Topic → SQS (email)     → Email worker
                         → SQS (inventory) → Inventory worker
                         → SQS (analytics) → Analytics worker

Each queue is independent — a slow email consumer doesn't block inventory.
```

**Competing consumers** — multiple workers drain the same queue in parallel. The default scaling pattern: add consumers, queue distributes work automatically.

**Delayed processing** — SQS supports `DelaySeconds` up to 15 minutes per message. For longer delays: scheduled Lambda, or a separate "scheduled" Kafka topic with consumer-side timestamp checks.

## Sizing

**1. Message rate** — use the day-to-seconds rule. 1M users × 10 events/day = 10M/day ÷ 10^5 = 100/sec average, ~300/sec peak.

**2. Pick the technology:**

```
< 1K/sec on AWS      → SQS
1K–100K/sec          → SQS or Kafka (replay / multi-consumer needs decide)
100K+/sec            → Kafka
```

**3. Size consumers** — if each message takes 50 ms to process, one consumer handles 20/sec. At 300/sec peak, you need 15 consumers. With Kafka, you need at least 15 partitions.

**4. Size the buffer** — 300/sec peak × 1,800 sec (30-min buffer) = 540K messages. At 5 KB each = ~2.7 GB. SQS handles this without configuration. For Kafka, ensure broker disk covers your retention period.

**Keep messages small.** A Kafka broker pushing 1 KB messages at 500K/sec is moving 500 MB/sec — you'll hit network bandwidth before message rate limits. Store large payloads in S3 and put a pointer in the message.

## Key Mental Models

1. **Async = separate accepting work from doing work.** The user returns in 50 ms; the queue absorbs the rest.
2. **Sync if the user needs the result. Async if they don't.** A queue on a 5 ms DB write is overhead, not value.
3. **At-least-once is the default. Design idempotent consumers.** Assume every message arrives at least twice.
4. **Kafka: partitions cap parallelism.** Max consumers per group = partition count. Plan it upfront — changing it later is painful.
5. **SQS: zero ops, unlimited consumers, no replay.** Default for task queues on AWS.
6. **Always configure a DLQ.** A poison message without one blocks the queue or cycles forever.
7. **Monitor queue depth, not just errors.** A growing queue is a silent failure until it isn't.
8. **Keep messages small — pointer, not payload.** Large payloads in S3; message carries the reference.
