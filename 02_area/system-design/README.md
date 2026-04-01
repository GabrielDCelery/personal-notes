# System Design and Architecture

A mental model for reasoning about system design — not language-specific, not "how to design Twitter", but the underlying numbers, patterns, and trade-offs that let you make decisions about any system.

The series builds from raw numbers → understanding a single request → individual components → combining them into architectures.

## Lessons

| #   | File                               | Status | What it covers                                                                                                                                                                                                                                                                                           |
| --- | ---------------------------------- | ------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 01  | `01-numbers-and-quick-math.md`     | Done   | Reference numbers (latencies, sizes, throughput), quick estimation formulas (users → RPS → DB QPS → storage), instance sizing (RDS, ElastiCache, ECS, Lambda), worked e-commerce example                                                                                                                 |
| 02  | `02-anatomy-of-a-request.md`       | Done   | Layer-by-layer trace of an HTTP request (DNS → TCP → TLS → CDN → LB → app → DB → response), where latency hides, multi-service fan-out, timeouts and cascading failure, percentile latencies                                                                                                             |
| 03  | `03-databases.md`                  | Done   | Read vs write paths, why the ratio isn't fixed, B-tree indexes and composite index ordering, connection pooling (PgBouncer), Postgres/MySQL/MongoDB/DynamoDB comparison, scaling flowchart (optimise → vertical → replicas → cache → shard)                                                              |
| 04  | `04-caching.md`                    | Done   | Why caching works (access pattern skew), caching strategies (cache-aside, read-through, write-through, write-behind), invalidation (TTL, active, event-driven), hit rate math, Redis vs Memcached, CDN caching, common problems (thundering herd, penetration, avalanche, hot keys), when NOT to cache   |
| 05  | `05-queues-and-async.md`           | Done   | Sync vs async decision framework, Kafka/SQS/RabbitMQ comparison with throughput numbers, batch sizing trade-offs, delivery guarantees (at-least-once, exactly-once, idempotency), consumer scaling, backpressure, dead letter queues, common patterns (fan-out, request-reply, competing consumers)      |
| 06  | `06-scaling-decisions.md`          | Done   | The 6-stage scaling progression (single server → sharding), decision flowchart (what to do first), worked e-commerce example from 1K to 5M users, when to split services, sharding trade-offs, anti-patterns (premature microservices, premature sharding, cache/queue everything), key principles       |
| 07  | `07-large-data-and-migrations.md`  | Draft  | Transfer and scan speed numbers, batch vs streaming trade-offs, migration patterns (offline, dual-write, CDC, shadow reads), long-running workers vs event-driven fan-out, protecting production (throttling, read replicas, backpressure), ETL patterns, worked 200M-row migration example              |
| 08  | `08-cost-and-storage-lifecycle.md` | Draft  | Storage tier costs (RDS → S3 → Glacier, 115x spread), compute pricing models (on-demand, reserved, spot, Lambda), database archival (partition + Parquet + Athena), S3 lifecycle policies, Lambda vs containers crossover, worked journey pipeline cost comparison, common cost mistakes                 |
| 10  | `10-auth-patterns.md`              | Done   | Stateless (JWT) vs stateful (session) decision, JWT anatomy and verification, the revocation problem, access + refresh token rotation, where to store tokens (httpOnly cookie vs memory vs localStorage), sessions with shared storage, OAuth/OIDC flows, microservice auth (verify once at the gateway) |
| 11  | `11-communication-protocols.md`    | Done   | REST as default, short vs long polling (cost at scale), SSE (server push over HTTP, auto-reconnect), WebSockets (bidirectional, upgrade handshake), protocol comparison table, scaling persistent connections (sticky sessions, pub/sub backplane, dedicated gateway), decision framework                |
| 12  | `12-security.md`                   | Done   | Defence in depth (8-layer model), rate limiting (token bucket vs sliding window, Redis counters), DDoS / WAF, network isolation (VPC subnets, security groups, zero trust), secrets management (inject at runtime, rotation, least privilege), encryption in transit and at rest (envelope encryption), input validation (SQLi, XSS, path traversal, SSRF) |

## How to use this

Lesson 01 is the cheat sheet — the numbers and formulas you'd pull out during any estimation. Lessons 02-05 go deep on individual components. Lesson 06 is the playbook for combining them.

For a system design interview or architecture discussion:

1. Start with 01 to estimate the scale of the problem
2. Use 02 to identify where the bottleneck is
3. Dive into 03-05 for the specific component that needs attention
4. Use 06 for the overall scaling strategy
