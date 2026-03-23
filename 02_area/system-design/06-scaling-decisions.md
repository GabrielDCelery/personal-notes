# Scaling Decisions — Distilled

Scale in response to bottlenecks, not in anticipation of them — and always take the cheapest step first.

## The Scaling Progression

Bottlenecks appear in a predictable order. The right move at each stage is well understood. The mistake is jumping ahead.

```
         100   1K    10K   100K    1M    10M
           │    │      │     │      │     │
 Stage 1   ├────┤                          app + db on one box
 Stage 2        ├──────┤                   separate the db
 Stage 3               ├─────┤  ← stop    + redis + cdn
 Stage 4                     ├──────┤     + horizontal scale + replicas
 Stage 5                            ├─────+ queues + service split
 Stage 6                                  ├─→ sharding + specialised storage
           │    │      │     │      │     │
         100   1K    10K   100K    1M    10M
                          users
```

**Most applications never leave stage 3.** The ones that do usually have years between stages. Premature scaling adds complexity that makes everything harder to change.

**Availability vs capacity are different concerns.** The progression above is about capacity — when adding complexity pays off in traffic terms. Availability is separate: a single app server is a single point of failure at any traffic level. Running 2 instances behind a load balancer costs almost nothing and gives you crash isolation, health-check routing, and zero-downtime deploys from day one. If downtime is unacceptable, treat 2+ instances as a production baseline regardless of stage. Lambda sidesteps this entirely — it's inherently multi-instance from the first request.

**Queues and service splits have independent justifications that appear well before Stage 5.** Stage 5 is when you add them to handle write volume or diverging capacity needs — but those aren't the only reasons. A queue belongs in your design as soon as you have work that is async by nature: sending emails, generating PDFs, transcoding video, notifying a warehouse. That operation shouldn't block an HTTP response regardless of how much traffic you have. Similarly, splitting a service is justified by operational signals — two teams stepping on each other, different deploy cadences, different reliability requirements — none of which require 1M users. Don't wait for scale; act on the signal.

## The Decision Order

When something is slow or at capacity, walk through this in order. Stop at the first step that solves the problem.

```
System slow or hitting limits
         │
         ▼
  Is the code inefficient? ─────yes──→ Indexes, N+1 fixes, connection pooling,
         │ no                          payload size. 80% of problems end here.
         ▼
  Can you scale vertically? ────yes──→ Bigger instance, more RAM.
         │ no                          Cheapest and fastest fix.
         ▼
  Is it read-heavy (>70%)? ─────yes──→ Redis cache (cache-aside).
         │ no                          CDN for static. Read replicas if needed.
         ▼
  Write-heavy or bursty? ───────yes──→ Queue to buffer writes.
         │ no                          Move non-critical writes async.
         ▼
  Single app server maxed? ─────yes──→ Horizontal scale behind LB.
         │ no                          Must be stateless first.
         ▼
  DB is the write ceiling? ─────yes──→ Shard by a well-chosen key.
         │ no                          Move workloads to specialised storage.
         ▼
  Services diverging? ──────────yes──→ Split along domain boundaries.
                                       Communicate via queues and events.
```

The order encodes cost. An index costs nothing. A replica costs $200/month. Sharding costs months of engineering. Never jump to step 6 when step 1 would do.

## Stage 3: Caching Is the Highest-Leverage Step

A cache doesn't just speed things up — it multiplies effective capacity. 80–90% of reads in most applications are for the same hot data: product catalogues, user profiles, sessions. Intercepting those reads before they hit the database means a single DB instance handles 5–10× the traffic it would otherwise. This is why stage 3 is where most systems comfortably plateau.

```
Users → LB → App Server → Redis ──hit──→ response
                              │
                            miss (10–20%)
                              ↓
                             DB
```

Monitor hit rate. Below 80% means you're caching the wrong things or TTLs are too short.

## Stage 4: Stateless Is the Prerequisite

You cannot scale horizontally until app servers are stateless. Any request must be routable to any server without loss of context. If sessions, uploads, or config live on the server, a second server breaks things.

| State type       | Wrong                  | Right                        |
| ---------------- | ---------------------- | ---------------------------- |
| Sessions         | In-memory / filesystem | Redis                        |
| File uploads     | Local disk             | S3 / object storage          |
| In-process cache | Local only             | Redis (shared)               |
| Config           | Local files per server | Environment variables / SSM  |
| Scheduled tasks  | Cron on one server     | Dedicated scheduler / Lambda |

Fix these before adding servers. After that, horizontal scaling is straightforward: add servers behind the LB, route reads to replicas, writes to the primary.

```
                ┌─── App Server 1 ───┐
Users → LB ─────┼─── App Server 2 ───┼───→ Redis ───→ Primary (RW)
                └─── App Server 3 ───┘                    │
                    (all stateless)              ┌─────────┴─────────┐
                                              Replica             Replica
                                               (RO)                (RO)
```

## Stage 5: When to Split Services

The wrong reason to split: "microservices are best practice." The right reasons are operational signals:

| Signal                        | Example                                                |
| ----------------------------- | ------------------------------------------------------ |
| Different scaling needs       | Search handles 10× the traffic of checkout             |
| Different deploy cadence      | Marketing updates CMS daily; payments change quarterly |
| Different reliability targets | Payments need 99.99%; recommendations tolerate 99.9%   |
| Team ownership conflict       | Two teams stepping on each other in the same codebase  |

**Service boundaries follow domain boundaries, not technical layers.** "Order service" and "Payment service" are good. "Database service" and "Cache service" are infrastructure, not domains.

**Splitting too early is worse than splitting too late.** A well-structured monolith is easier to split than a poorly-designed set of services is to fix. When in doubt, stay monolithic.

Between services: sync HTTP/gRPC when the caller needs the response to continue; async queues when it doesn't; event bus (SNS/Kafka) when multiple services need to react to the same event.

```
Users → LB → API Service ───→ Redis ───→ Database
                  │
             publish to queue
                  │
            ┌─────┴───────┐
            │ SQS / Kafka │
            └─────┬───────┘
         ┌────────┼────────┐
         ▼        ▼        ▼
      Email    Payment  Analytics
      Worker   Worker    Worker
```

## Stage 6: Sharding Is a Last Resort

Vertical scaling and read replicas handle most systems to 100K+ QPS. Sharding — splitting data across multiple DB instances — is the move when write volume exceeds what a single primary can handle. It is expensive in engineering time.

**Why it's painful:** cross-shard JOINs are hard or impossible, schema migrations must run on every shard, transactions across shards require distributed coordination, and rebalancing when adding shards is complex.

The shard key is the most consequential decision:

| Shard key         | Good for                                     | Bad for                                  |
| ----------------- | -------------------------------------------- | ---------------------------------------- |
| User ID           | User-centric apps — all user data co-located | Queries across users                     |
| Geographic region | Compliance, regional isolation               | Travelling users, global queries         |
| Hash of ID        | Even distribution                            | Range queries ("all orders this month")  |
| Time (date)       | Time-series, logs                            | Latest shard gets all writes — hot shard |

At this stage, specific workloads also get moved to purpose-built storage:

| Workload             | Technology                       | Why not Postgres                                          |
| -------------------- | -------------------------------- | --------------------------------------------------------- |
| Full-text search     | Elasticsearch / OpenSearch       | Postgres FTS doesn't scale to millions of complex queries |
| Time-series          | TimescaleDB / InfluxDB           | Optimised for append-heavy, time-range access patterns    |
| Analytics / OLAP     | Redshift / BigQuery / ClickHouse | OLTP isn't built for scanning billions of rows            |
| Graph relationships  | Neo4j / Neptune                  | Recursive JOINs are slow and awkward in SQL               |
| Large binary objects | S3                               | Databases aren't filesystems                              |

**Use Postgres until a specific workload outgrows it. Move that workload. Don't start with six databases.**

## Anti-Patterns

**Premature microservices.** Network calls add latency and failure modes. Distributed tracing, service discovery, deployment pipelines — all overhead. With two engineers and 1,000 users, a monolith is faster to build, deploy, and debug. Start monolithic; split when forced by operational signals above.

**Premature sharding.** A single Postgres instance handles 50K+ QPS with proper indexing and caching. Most applications will never need to shard. The path is: vertical → cache → read replicas → shard. Most apps stop at cache.

**Caching everything without a strategy.** Cache the hot path — the top 20% of data serving 80% of reads. Caching everything degrades hit rate, wastes memory, and makes invalidation unmanageable.

**Queuing everything because async sounds better.** A 5 ms DB write doesn't need a queue. Queues add latency, consumers to manage, DLQs to monitor. Sync by default; async when the user doesn't need the result.

**Over-provisioning for peak 24/7.** Running 20 servers around the clock for a 2-hour daily spike means paying for 20 when you need 2 for 22 hours. Auto-scale on CPU/memory. Set a minimum for baseline; let it scale up for peaks.

## Key Mental Models

1. **Do the cheapest thing first.** Fix code → vertical scale → cache → replicas → horizontal scale → shard. Never jump steps.
2. **Most systems never leave stage 3.** 100K users, Redis cache, single DB. No microservices needed.
3. **The bottleneck is almost always the database.** Find it with metrics, fix that specific thing.
4. **Stateless app servers, stateful storage.** App servers are disposable — all state lives in DB, cache, or object storage.
5. **Scale when metrics tell you to.** Premature complexity makes everything harder to change.
6. **Service boundaries follow domain boundaries.** Split along "what changes together", not technical layers.
7. **Every component you add is a cost.** Monitoring, debugging, ops, reasoning. The simplest architecture that meets requirements is the best one.
