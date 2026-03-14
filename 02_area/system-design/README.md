# System Design and Architecture

A mental model for reasoning about system design — not language-specific, not "how to design Twitter", but the underlying numbers, patterns, and trade-offs that let you make decisions about any system.

The series builds from raw numbers → understanding a single request → individual components → combining them into architectures.

## Lessons

| #   | File                           | Status      | What it covers                                                                                                                                                                               |
| --- | ------------------------------ | ----------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 01  | `01-numbers-and-quick-math.md` | Done        | Reference numbers (latencies, sizes, throughput), quick estimation formulas (users → RPS → DB QPS → storage), instance sizing (RDS, ElastiCache, ECS, Lambda), worked e-commerce example     |
| 02  | `02-anatomy-of-a-request.md`   | Done        | Layer-by-layer trace of an HTTP request (DNS → TCP → TLS → CDN → LB → app → DB → response), where latency hides, multi-service fan-out, timeouts and cascading failure, percentile latencies |
| 03  | `03-databases.md`              | Done        | Read vs write paths, why the ratio isn't fixed, B-tree indexes and composite index ordering, connection pooling (PgBouncer), Postgres/MySQL/MongoDB/DynamoDB comparison, scaling flowchart (optimise → vertical → replicas → cache → shard) |
| 04  | `04-caching.md`                | Not started | Redis, Memcached, CDNs — cache-aside vs write-through vs write-behind, hit rate math, cache sizing, invalidation strategies, thundering herd, when caching hurts                             |
| 05  | `05-queues-and-async.md`       | Not started | Kafka, SQS, RabbitMQ — when to go sync vs async, throughput numbers, batch sizing, backpressure, consumer scaling, exactly-once vs at-least-once, dead letter queues                         |
| 06  | `06-scaling-decisions.md`      | Not started | The decision framework tying it all together — vertical vs horizontal, stateless vs stateful, read-heavy vs write-heavy, common architecture patterns at different scale tiers               |

## How to use this

Lesson 01 is the cheat sheet — the numbers and formulas you'd pull out during any estimation. Lessons 02-05 go deep on individual components. Lesson 06 is the playbook for combining them.

For a system design interview or architecture discussion:

1. Start with 01 to estimate the scale of the problem
2. Use 02 to identify where the bottleneck is
3. Dive into 03-05 for the specific component that needs attention
4. Use 06 for the overall scaling strategy
