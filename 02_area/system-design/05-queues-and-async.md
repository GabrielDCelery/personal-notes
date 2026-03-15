# Queues and Async Processing

Lesson 04 showed how caching reduces read load by intercepting repeated queries before they hit the database. But what about writes? When a user places an order, you need to charge their card, update inventory, send a confirmation email, notify the warehouse, and update analytics. Doing all of that synchronously in one request means the user waits for everything to finish — and if the email service is slow, the whole request is slow.

Queues let you separate "accepting work" from "doing work". The request records the intent (the order), returns immediately, and the rest happens asynchronously. This is the fundamental shift: you're trading immediate consistency for better latency, reliability, and throughput.

## Sync vs Async — The Decision

Not everything should be async. The decision comes down to: **does the user need the result right now?**

```
Sync (do it now, wait for the result):
  - User logs in → validate credentials → return token
  - User views product → query DB → return product
  - User checks balance → query DB → return balance

Async (accept the intent, do it later):
  - User places order → accept order → process payment later
  - User uploads image → accept upload → resize/compress later
  - User requests report → accept request → generate and email later
```

The mental model is a restaurant. Sync is like a food truck — you order, wait, and get your food. Async is like a sit-down restaurant — you order, get a confirmation (your order number), and the kitchen works on it. You don't stand at the kitchen window watching.

### When to go async

| Signal                                            | Why                                                    |
| ------------------------------------------------- | ------------------------------------------------------ |
| The operation takes > 1-2 seconds                 | Users shouldn't wait that long for a page response     |
| The operation can fail and be retried             | Email delivery, payment processing, external API calls |
| The operation isn't needed for the response       | Analytics events, audit logs, notifications            |
| Multiple independent steps follow the trigger     | Order → payment + inventory + email + warehouse        |
| Traffic is spiky and the downstream can't keep up | Flash sale → 10x order volume, DB can't handle it      |

### When to stay sync

| Signal                                      | Why                                                         |
| ------------------------------------------- | ----------------------------------------------------------- |
| User needs the result to continue           | Login, search results, balance check                        |
| The operation is fast (< 100 ms)            | Simple DB read/write — a queue adds overhead, not value     |
| Ordering and consistency matter immediately | "Is this username taken?" must be checked before proceeding |
| The system is simple and low traffic        | A queue adds infrastructure for no benefit                  |

The common mistake is making everything async for the sake of it. Every queue adds operational complexity — a consumer to run, monitoring to set up, dead letters to handle, and a delay the user can perceive. If a simple DB write takes 5 ms, don't put it on a queue.

## How a Queue Fits in the Architecture

```
Without queue (sync):

  User → API → do everything (charge card, update DB, send email, notify warehouse)
                    │
                    └→ total: 50 + 200 + 300 + 150 = 700 ms response time
                       if email service is down, entire request fails

With queue (async):

  User → API → save order to DB → publish message to queue → return 202 Accepted
                    │                                              (50 ms response)
                    │
                    └→ Queue → Consumer 1: charge card (200 ms)
                         │  → Consumer 2: send email (300 ms)
                         │  → Consumer 3: notify warehouse (150 ms)
                         │
                         └→ if email service is down, message stays in queue → retry later
```

The API response dropped from 700 ms to 50 ms. The downstream processing still takes the same total time, but the user doesn't wait for it. And if one consumer fails, the others are unaffected — the failed message gets retried automatically.

## Queue Fundamentals

Before comparing specific technologies, understand what all message queues share.

### Producers and consumers

A **producer** publishes messages to a queue. A **consumer** reads messages from the queue and processes them. This decoupling is the whole point — the producer doesn't know or care who processes the message, how long it takes, or whether the consumer is even running right now.

```
  Producer(s)                    Consumer(s)
  ┌─────────┐    ┌─────────┐    ┌──────────┐
  │ API      │───→│  Queue  │───→│ Worker 1 │
  │ Server   │    │         │    │ Worker 2 │
  │          │    │         │    │ Worker 3 │
  └─────────┘    └─────────┘    └──────────┘
  Fast: accepts    Durable:      Scales:
  work, returns    holds msgs    add/remove
  immediately      until done    as needed
```

### Key concepts

| Concept                     | What it means                                                                           |
| --------------------------- | --------------------------------------------------------------------------------------- |
| **Message**                 | A unit of work — typically a JSON payload (order ID, event data, etc.)                  |
| **Topic / Queue**           | The named channel messages are published to                                             |
| **Partition / Shard**       | A subdivision of a topic for parallelism (Kafka, Kinesis)                               |
| **Consumer group**          | A set of consumers that share the work — each message goes to one consumer in the group |
| **Offset / Receipt handle** | The consumer's position in the queue — "I've processed up to here"                      |
| **Visibility timeout**      | How long a message is hidden after a consumer picks it up (SQS model)                   |
| **Acknowledgement (ack)**   | Consumer tells the queue "I'm done with this message, delete it"                        |
| **Dead letter queue (DLQ)** | Where messages go after failing too many times                                          |
| **Backpressure**            | When consumers can't keep up and the queue grows                                        |

## Technology Comparison

Three technologies dominate: **Kafka** for high-throughput event streaming, **SQS** for simple managed queues on AWS, and **RabbitMQ** for flexible routing. They're built for different problems.

### Kafka

Kafka is a **distributed commit log**. Messages are appended to partitions and stay there — consumers read by offset, like reading a file from a specific position. This means multiple consumer groups can read the same data independently, and you can replay old messages.

```
Topic: "orders" (3 partitions)

  Partition 0:  [msg0] [msg1] [msg2] [msg3] [msg4] ───→
  Partition 1:  [msg0] [msg1] [msg2] [msg3] ───→
  Partition 2:  [msg0] [msg1] [msg2] ───→

  Consumer Group A (order processing):
    Consumer A1 reads Partition 0
    Consumer A2 reads Partition 1
    Consumer A3 reads Partition 2

  Consumer Group B (analytics):
    Consumer B1 reads Partition 0, 1
    Consumer B2 reads Partition 2

  Each group tracks its own offset independently.
  Messages are NOT deleted after reading — they expire after a retention period.
```

**Key properties:**

- **Throughput**: 100K-1M+ messages/sec per broker, scales linearly with brokers and partitions
- **Ordering**: Guaranteed within a partition, not across partitions
- **Retention**: Messages persist for a configurable period (hours, days, forever) — this is what enables replay
- **Consumer scaling**: One consumer per partition per consumer group. 12 partitions = max 12 consumers in a group
- **Latency**: Typically 2-10 ms end-to-end (producer → consumer), can be sub-millisecond with tuning
- **Storage**: Sequential disk writes — Kafka's throughput comes from treating disk like a log, not random access

**When to use Kafka**: High-throughput event streaming (clickstream, logs, metrics), event sourcing, when multiple systems need the same event stream, when you need replay capability.

**When NOT to use Kafka**: Simple task queues (overkill), low message volume (< 1K/sec — operational cost isn't justified), when you need complex routing logic per message.

### SQS (Simple Queue Service)

SQS is a **managed message queue** on AWS. No servers to run, no brokers to configure. You create a queue and start sending messages. AWS handles scaling, durability, and availability.

```
Standard Queue:                        FIFO Queue:

  Producer → [msg3] [msg1] [msg2]       Producer → [msg1] [msg2] [msg3]
             (best-effort order)                    (strict order)
             (at-least-once)                        (exactly-once)
             (unlimited throughput)                  (300 msg/sec, 3,000 with batching)
```

**Key properties:**

- **Throughput**: Standard queue — practically unlimited (AWS scales automatically). FIFO queue — 300 messages/sec per queue (3,000 with batching, 70,000 with high throughput mode)
- **Ordering**: Standard — best effort (mostly ordered, no guarantee). FIFO — strict ordering within a message group
- **Retention**: 1 minute to 14 days (default 4 days)
- **Consumer scaling**: Any number of consumers. SQS distributes messages across them. No partition limit
- **Latency**: Typically 1-10 ms to send, consumers poll (short poll: immediate response, long poll: waits up to 20 seconds for a message)
- **Visibility timeout**: When a consumer picks up a message, it's hidden from other consumers for a configurable period (default 30 seconds). If the consumer doesn't delete it in time, it reappears for another consumer
- **Cost**: Pay per request — ~$0.40 per million requests. Very cheap at low-to-medium volume

**When to use SQS**: Task queues on AWS (job processing, email sending, image resizing), decoupling microservices, when you want zero operational overhead, when throughput requirements are moderate.

**When NOT to use SQS**: When you need multiple consumers reading the same message (SQS deletes after ack — use SNS + SQS fan-out instead), when you need replay, when you're not on AWS.

### RabbitMQ

RabbitMQ is a **message broker** that supports complex routing. Messages go through an exchange, which routes them to queues based on rules. This makes it the most flexible option for message routing patterns.

```
Exchange types:

  Direct:    msg with key="order" → goes to "order" queue
  Fanout:    msg → goes to ALL bound queues (broadcast)
  Topic:     msg with key="order.eu.paid" → matches "order.eu.*" and "order.#"
  Headers:   route based on message headers, not routing key
```

**Key properties:**

- **Throughput**: 20K-50K messages/sec per node (less than Kafka, more than enough for most apps)
- **Ordering**: Guaranteed per queue (single consumer) — with multiple consumers, ordering is per-consumer
- **Retention**: Messages are deleted after ack. No built-in replay. Can persist to disk for durability
- **Consumer scaling**: Multiple consumers per queue. RabbitMQ round-robins messages across them
- **Latency**: Sub-millisecond in many cases — lower than Kafka for small message volumes
- **Routing**: The standout feature. Complex patterns: fanout, topic matching, header-based, priority queues
- **Protocol**: AMQP standard — works with any language, not cloud-vendor locked

**When to use RabbitMQ**: Complex routing requirements (route different events to different consumers), request-reply patterns, priority queues, when you need flexible message patterns, multi-cloud or on-prem.

**When NOT to use RabbitMQ**: Very high throughput (> 50K msg/sec sustained — Kafka handles this better), when you need message replay, when you want zero ops (use SQS instead).

### Comparison at a glance

|                      | Kafka                                     | SQS                                    | RabbitMQ                               |
| -------------------- | ----------------------------------------- | -------------------------------------- | -------------------------------------- |
| **Model**            | Distributed log                           | Managed queue                          | Message broker                         |
| **Throughput**       | 100K-1M+/sec                              | Unlimited (standard)                   | 20K-50K/sec                            |
| **Ordering**         | Per partition                             | Best-effort (FIFO available)           | Per queue                              |
| **Delivery**         | At-least-once (exactly-once within Kafka) | At-least-once (exactly-once with FIFO) | At-least-once (at-most-once available) |
| **Retention**        | Days/weeks/forever                        | Up to 14 days                          | Until ack'd                            |
| **Replay**           | Yes (re-read from offset)                 | No                                     | No                                     |
| **Routing**          | Topics + partitions                       | Single queue (fan-out via SNS)         | Exchanges with complex routing         |
| **Consumer scaling** | Max 1 per partition per group             | Unlimited                              | Unlimited                              |
| **Ops overhead**     | High (brokers, ZooKeeper/KRaft)           | Zero (managed)                         | Medium (Erlang cluster)                |
| **Cost model**       | Infrastructure (brokers + storage)        | Per request (~$0.40/M)                 | Infrastructure (nodes)                 |
| **Best for**         | Event streaming, high volume, replay      | Task queues on AWS, simple decoupling  | Complex routing, multi-protocol        |

### Choosing: the quick decision

```
  "I need a task queue on AWS"                          → SQS
  "I need high-throughput event streaming"              → Kafka
  "I need complex routing (fanout, topic matching)"     → RabbitMQ
  "Multiple systems need the same event"                → Kafka (consumer groups)
                                                          or SNS + SQS (fan-out)
  "I need replay / event sourcing"                      → Kafka
  "I want zero operational overhead"                    → SQS
  "I'm on-prem or multi-cloud"                          → RabbitMQ (or self-hosted Kafka)
  "Message volume is < 1K/sec and I'm on AWS"           → SQS (don't overthink it)
```

## Throughput Numbers

These are per-node/per-broker numbers. All three scale horizontally by adding nodes.

| System              | Messages/sec (single node)                | Typical message size | Max message size                   | Notes                                                  |
| ------------------- | ----------------------------------------- | -------------------- | ---------------------------------- | ------------------------------------------------------ |
| Kafka (per broker)  | 100K-1M+                                  | 1-10 KB              | 1 MB (configurable)                | Throughput scales linearly with partitions and brokers |
| SQS (standard)      | Effectively unlimited                     | 1-10 KB              | 256 KB (up to 2 GB via S3 pointer) | AWS scales transparently                               |
| SQS (FIFO)          | 300/sec (3K batched, 70K high-throughput) | 1-10 KB              | 256 KB                             | Per message group for ordering                         |
| RabbitMQ (per node) | 20K-50K                                   | 1-10 KB              | 128 MB (practical limit ~10 MB)    | Degrades above 50K sustained                           |

Message size matters for throughput. A Kafka broker pushing 1 KB messages at 500K/sec is moving 500 MB/sec of data — you'll hit network bandwidth before message rate limits. The rule of thumb: **keep messages small** (event reference + metadata), put large payloads in S3/blob storage and include a pointer in the message.

## Batch Sizing

Batching is the single biggest throughput lever. Instead of sending messages one at a time, you group them and send in batches. This amortises the per-message overhead (network round trip, serialisation, broker acknowledgement) across many messages.

### Why batching matters

```
Without batching (1 message per request):
  1,000 messages × 5 ms round trip each = 5,000 ms (5 seconds)

With batching (100 messages per request):
  10 batches × 5 ms round trip each = 50 ms
  Same 1,000 messages, 100x faster
```

### Batch size trade-offs

| Batch size      | Throughput | Latency                   | Memory   | When to use                           |
| --------------- | ---------- | ------------------------- | -------- | ------------------------------------- |
| 1 (no batching) | Lowest     | Lowest (immediate)        | Lowest   | Real-time requirements (chat, alerts) |
| 10-50           | Good       | Low (ms-level delay)      | Low      | Default for most use cases            |
| 100-500         | High       | Moderate (10-50 ms delay) | Moderate | High-throughput pipelines             |
| 1,000-10,000    | Highest    | High (100ms+ delay)       | High     | Bulk data ingestion, analytics        |

The trade-off is always **throughput vs latency**. Larger batches mean higher throughput but each individual message waits longer before being sent. Most systems let you configure both a **batch size** (max messages per batch) and a **linger time** (max time to wait before sending a partial batch).

| System         | Batch setting                 | Default                 | Recommendation                                 |
| -------------- | ----------------------------- | ----------------------- | ---------------------------------------------- |
| Kafka producer | `batch.size` + `linger.ms`    | 16 KB / 0 ms            | 64-256 KB / 5-20 ms for throughput             |
| SQS            | `SendMessageBatch`            | N/A (explicit API call) | Always batch when possible (up to 10 messages) |
| RabbitMQ       | Publisher confirms + batching | 1 (no batching)         | Batch confirms every 100-200 messages          |

## Delivery Guarantees

This is where queues get tricky. What happens when things fail? A consumer crashes halfway through processing a message. The network drops between the broker and consumer. A producer sends a message but doesn't get an ack back — did the broker receive it or not?

### The three guarantees

| Guarantee         | What it means                                                   | How it fails                                               | Performance                                    |
| ----------------- | --------------------------------------------------------------- | ---------------------------------------------------------- | ---------------------------------------------- |
| **At-most-once**  | Message is delivered 0 or 1 times. Fire and forget              | If the consumer crashes before processing, message is lost | Fastest — no ack overhead                      |
| **At-least-once** | Message is delivered 1 or more times. Always retried on failure | Consumer might process the same message twice (duplicate)  | Middle — requires ack                          |
| **Exactly-once**  | Message is delivered exactly 1 time                             | Extremely hard to guarantee across distributed systems     | Slowest — requires transactions or idempotency |

**At-least-once is the default for almost everything.** It's the practical sweet spot — you never lose a message, and you handle duplicates on the consumer side. This is what Kafka, SQS standard, and RabbitMQ all default to.

### Why exactly-once is hard

True exactly-once delivery across distributed systems is practically impossible. What systems actually implement is **exactly-once semantics** — the message might be delivered more than once, but the system ensures the effect only happens once.

```
The problem:

  Producer → Broker: "Here's message X"
  Broker: saves message X, sends ack
  Network: drops the ack
  Producer: "I didn't get an ack, let me resend"
  Broker: now has message X twice

  OR:

  Broker → Consumer: "Here's message X"
  Consumer: processes message X, sends ack
  Network: drops the ack
  Broker: "Consumer didn't ack, let me resend"
  Consumer: processes message X again (duplicate)
```

The solution is **idempotency** — design your consumers so that processing the same message twice has the same effect as processing it once.

### Making consumers idempotent

| Pattern                       | How it works                                  | Example                                                                       |
| ----------------------------- | --------------------------------------------- | ----------------------------------------------------------------------------- |
| **Idempotency key**           | Store processed message IDs, skip duplicates  | Before processing order #123, check if you've already processed it            |
| **Database upsert**           | Use INSERT ON CONFLICT or PUT (not increment) | Set balance to $50, not "subtract $10" — same result if run twice             |
| **Idempotent external calls** | Use idempotency keys in external API calls    | Stripe's `Idempotency-Key` header — same key = same charge, not double charge |
| **Version / timestamp check** | Only apply if version is newer                | Update product price only if event timestamp > last update timestamp          |

The mental model: **assume every message will be delivered at least twice, and design accordingly.** If your consumer can safely process the same message multiple times without side effects, you get effectively exactly-once behaviour with at-least-once delivery.

## Consumer Scaling

The number of consumers determines how fast you drain the queue. But scaling consumers isn't as simple as "add more" — it depends on the queue technology and the ordering requirements.

### Kafka consumer scaling

In Kafka, **one partition can only be read by one consumer in a consumer group**. This is a hard constraint.

```
Topic with 6 partitions:

  3 consumers (under-utilised capacity exists):
    Consumer 1 → Partition 0, 1
    Consumer 2 → Partition 2, 3
    Consumer 3 → Partition 4, 5

  6 consumers (ideal — 1:1 mapping):
    Consumer 1 → Partition 0
    Consumer 2 → Partition 1
    ...
    Consumer 6 → Partition 5

  9 consumers (3 are idle, wasted):
    Consumer 1 → Partition 0
    ...
    Consumer 6 → Partition 5
    Consumer 7 → idle
    Consumer 8 → idle
    Consumer 9 → idle
```

This means **you must plan partition count upfront**. If you create a topic with 6 partitions, you can never have more than 6 consumers processing in parallel (per consumer group). Choose a partition count that matches your expected peak parallelism. Common starting points: 6 for low-traffic topics, 12-30 for medium, 50-100+ for high-throughput.

Increasing partitions later is possible but causes a rebalance and breaks ordering guarantees for existing data.

### SQS consumer scaling

SQS has no partition concept — any number of consumers can poll the same queue. AWS distributes messages across consumers automatically. This makes scaling simple: need more throughput? Add more consumers.

```
SQS auto-scaling pattern:

  Queue depth: 0-100      → 2 consumers
  Queue depth: 100-1,000  → 5 consumers
  Queue depth: 1,000+     → 20 consumers

  Metric to watch: ApproximateNumberOfMessagesVisible
  Scale trigger: messages per consumer > threshold
```

The common pattern on AWS is an SQS queue triggering Lambda functions — AWS scales the Lambda invocations automatically based on queue depth (up to 1,000 concurrent executions by default). No consumer management at all.

### RabbitMQ consumer scaling

RabbitMQ round-robins messages across consumers on a queue. Add consumers to increase throughput. Use `prefetch` to control how many unacked messages each consumer holds.

```
prefetch = 1:  Consumer gets 1 message, must ack before getting next
               Slow but fair distribution. Good when processing time varies widely.

prefetch = 10: Consumer gets 10 messages upfront
               Faster throughput, but if consumer crashes, 10 messages need redelivery.
               Good when processing time is predictable.
```

### Consumer scaling rules of thumb

| Situation                                      | Action                                                        |
| ---------------------------------------------- | ------------------------------------------------------------- |
| Queue depth growing steadily                   | Add consumers (or increase batch size)                        |
| Processing time per message is high (> 1 sec)  | Consider optimising processing before adding consumers        |
| One consumer can handle the load               | Don't add more — unnecessary complexity                       |
| Traffic is spiky (high for minutes, then idle) | Auto-scale consumers based on queue depth                     |
| Ordering matters within a group                | Kafka: use same partition key. SQS FIFO: use message group ID |

## Backpressure

Backpressure happens when producers are faster than consumers. The queue grows, latency increases, and eventually something breaks — either the queue runs out of storage, memory fills up, or messages expire before being processed.

```
Normal:       Producer (100 msg/s) → Queue (small) → Consumer (100 msg/s)
Backpressure: Producer (500 msg/s) → Queue (GROWING) → Consumer (100 msg/s)

  Queue depth over time:
  ────────────────────────────────────────
  ▁▁▁▂▂▃▃▄▅▅▆▇▇███████  ← queue keeps growing
  0                       ← this is a problem
```

### Handling backpressure

| Strategy                 | How it works                                                                 | Trade-off                                             |
| ------------------------ | ---------------------------------------------------------------------------- | ----------------------------------------------------- |
| **Scale consumers**      | Add more consumers to match producer rate                                    | Cost, complexity                                      |
| **Increase batch size**  | Process more messages per consumer cycle                                     | Higher per-message latency                            |
| **Rate-limit producers** | Slow down the source (return 429, apply throttling)                          | Pushes backpressure upstream to callers               |
| **Shed load**            | Drop low-priority messages, keep critical ones                               | Some messages are lost by design                      |
| **Absorb the spike**     | Let the queue buffer it — if the spike is temporary, consumers will catch up | Only works if spikes are short and queue has capacity |

The most common approach: **absorb short spikes, auto-scale consumers for sustained load, alert if queue depth exceeds a threshold.** SQS with Lambda handles this automatically. Kafka requires you to monitor consumer lag (the gap between latest offset and consumer offset) and scale accordingly.

### Key metrics to monitor

| Metric                         | What it tells you                            | Alert threshold                           |
| ------------------------------ | -------------------------------------------- | ----------------------------------------- |
| **Queue depth / consumer lag** | How far behind consumers are                 | Growing for > 5 minutes                   |
| **Message age**                | How long the oldest message has been waiting | > your SLA (e.g., 30 seconds, 5 minutes)  |
| **Consumer processing time**   | How long each message takes to process       | Increasing trend (degradation)            |
| **Error rate**                 | How often consumers fail to process messages | > 1-5%                                    |
| **DLQ depth**                  | How many messages have permanently failed    | Any message in DLQ deserves investigation |

## Dead Letter Queues

A dead letter queue (DLQ) is where messages go to die — or more accurately, where messages go when they've failed processing too many times. Without a DLQ, a "poison message" (one that always fails) blocks the queue forever or gets silently dropped.

```
Normal flow:
  Queue → Consumer → success → ack → message deleted

Failure flow:
  Queue → Consumer → fail → message returns to queue → retry
       → Consumer → fail → message returns to queue → retry
       → Consumer → fail → message returns to queue → retry
       → Max retries reached → message moves to DLQ

  DLQ:
    [failed msg 1] [failed msg 2] [failed msg 3]
         │
         └→ Investigate, fix the bug, then replay from DLQ
```

### DLQ configuration by system

| System   | Setting                                                | Typical config                                          |
| -------- | ------------------------------------------------------ | ------------------------------------------------------- |
| SQS      | `maxReceiveCount` + `deadLetterTargetArn`              | 3-5 retries, then DLQ                                   |
| RabbitMQ | `x-dead-letter-exchange` + `x-dead-letter-routing-key` | 3-5 retries, then DLQ                                   |
| Kafka    | No built-in DLQ — implement in consumer code           | Consumer catches exception, publishes to a `.dlq` topic |

### What to do with DLQ messages

DLQ messages need a plan, not just a parking lot.

1. **Alert** — any message in the DLQ should trigger an alert. These are messages that your system couldn't handle
2. **Investigate** — look at the message payload and the consumer error logs. Is it a bug? A downstream service outage? Bad input data?
3. **Fix** — fix the bug or wait for the downstream service to recover
4. **Replay** — re-process the messages from the DLQ. SQS supports "redrive" (move messages back to the source queue). For Kafka, read from the DLQ topic and republish to the original topic

Never leave DLQ messages unattended. They represent lost work or broken assumptions about your system.

## Common Patterns

### Fan-out: one event, many consumers

A single event needs to trigger multiple independent actions. On AWS, the standard pattern is SNS → SQS.

```
  Order placed
       │
       ▼
   SNS Topic
   ┌───┼────────┐
   ▼   ▼        ▼
  SQS  SQS     SQS
  │    │        │
  ▼    ▼        ▼
 Email Inventory Analytics
```

Each SQS queue gets a copy of every message. Each consumer processes independently. If the email consumer is slow, it doesn't affect inventory or analytics.

With Kafka, you achieve fan-out through consumer groups — each group reads the same topic independently.

### Request-reply: async with a response

Sometimes you need an async flow but still need to get the result back. The pattern: send the request with a **correlation ID**, consumer processes it and publishes the result to a reply queue/topic.

```
  Client → Request Queue: { correlationId: "abc", payload: ... }
                                    │
                                    ▼
                               Worker processes
                                    │
                                    ▼
         Client ← Reply Queue: { correlationId: "abc", result: ... }
```

This is common in RabbitMQ. In practice, most systems prefer webhooks or polling instead — the client sends the request, gets back a job ID, and polls a status endpoint.

### Competing consumers: parallel processing

Multiple consumers on the same queue, each taking different messages. The simplest scaling pattern.

```
  Queue: [msg1] [msg2] [msg3] [msg4] [msg5] [msg6]
              │          │          │
              ▼          ▼          ▼
          Worker 1   Worker 2   Worker 3

  Each worker processes different messages in parallel.
  Queue distributes work round-robin.
```

### Delay / scheduled processing

Process a message after a delay. "Send a reminder email 24 hours after signup." "Retry this failed payment in 1 hour."

| System   | How to delay                                                                                      |
| -------- | ------------------------------------------------------------------------------------------------- |
| SQS      | `DelaySeconds` (up to 15 minutes per message) or use a scheduled Lambda                           |
| RabbitMQ | `x-delayed-message` plugin or TTL + dead letter exchange trick                                    |
| Kafka    | No built-in delay — use a separate "scheduled" topic with timestamp checks, or external scheduler |

## When Queues Go Wrong

### Message ordering assumptions

You assumed messages would arrive in order. They don't (in most systems).

**Fix**: If ordering matters, use Kafka with a consistent partition key (same key = same partition = ordered), or SQS FIFO with message group IDs. Accept that ordering only applies within a partition/group, not globally.

### Slow consumers causing memory pressure

The queue grows faster than consumers can drain it. The broker runs out of memory or disk.

**Fix**: Monitor queue depth, auto-scale consumers, set retention limits (SQS: max 14 days, Kafka: configurable retention), and alert when depth exceeds thresholds.

### Poison messages

A message that always fails processing — bad format, references a deleted record, triggers a bug. Without a DLQ, it blocks the queue or cycles forever.

**Fix**: Always configure a DLQ. Set max retries to 3-5. Log the failure with full message content for debugging.

### Consumer rebalancing storms (Kafka)

Adding or removing consumers triggers a rebalance — Kafka reassigns partitions across consumers. During rebalance, all consumers pause. If consumers crash and restart frequently, you get constant rebalancing.

**Fix**: Use incremental cooperative rebalancing (Kafka 2.4+), increase `session.timeout.ms` to avoid false positives, and fix the root cause of consumer crashes.

### Duplicate processing

At-least-once delivery means duplicates. Your consumer charges a customer twice.

**Fix**: Make consumers idempotent. Use idempotency keys, upserts, or version checks. This is the single most important design decision for queue consumers.

## Sizing Mental Model

To estimate what you need:

**1. Calculate message rate:**
From lesson 01 — convert your user activity to messages per second. If 1M users generate 10 events per day: 10M events/day ÷ 10^5 = 100 messages/sec average, 300-500/sec peak.

**2. Pick the technology:**

- < 1K msg/sec on AWS → SQS (don't overthink it)
- 1K-100K msg/sec → SQS or Kafka, depending on whether you need replay/multiple consumers
- 100K+ msg/sec → Kafka

**3. Size consumers:**
If each message takes 50 ms to process, one consumer handles 20 msg/sec. At 500 msg/sec peak, you need 25 consumers. With Kafka, you need at least 25 partitions.

**4. Size the buffer:**
Queue should be able to absorb a traffic spike if consumers go down temporarily. If peak is 500 msg/sec and you want 30 minutes of buffer: 500 × 1,800 = 900K messages. At 5 KB each = 4.5 GB. SQS handles this without thinking. Kafka — ensure broker disk can hold retention period of data.

## Quick Reference

```
                    ┌─────────────────────────────────────────────────┐
  When to use       │  User doesn't need result now                   │
  a queue:          │  Operation takes > 1-2 sec                      │
                    │  Need to retry on failure                       │
                    │  Need to decouple producer from consumer        │
                    │  Traffic is spiky                               │
                    └─────────────────────────────────────────────────┘

                    ┌─────────────────────────────────────────────────┐
  Default choices:  │  On AWS, simple tasks → SQS                     │
                    │  High throughput / replay → Kafka                │
                    │  Complex routing → RabbitMQ                      │
                    └─────────────────────────────────────────────────┘

                    ┌─────────────────────────────────────────────────┐
  Always do:        │  Make consumers idempotent                      │
                    │  Configure a dead letter queue                   │
                    │  Monitor queue depth and consumer lag            │
                    │  Use batching for throughput                     │
                    │  Plan partition count upfront (Kafka)            │
                    └─────────────────────────────────────────────────┘
```

The next lesson ties everything together — databases, caches, and queues — into a scaling decision framework. Given a system at a specific scale, what do you do first? When do you add replicas vs caches vs queues? How do common architectures evolve as traffic grows?
