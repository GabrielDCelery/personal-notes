# Scaling Decisions

The previous lessons covered individual components — databases, caches, queues. Each one has its own scaling path and trade-offs. But real systems don't live in isolation. The question isn't "how do I scale Postgres?" — it's "my system handles 5,000 users and needs to handle 500,000. What changes, in what order, and what are the trade-offs?"

This lesson is the decision framework. Given a system at a specific scale, what's the cheapest, simplest thing to do next? The core principle: **do the easiest thing that buys the most headroom, and don't do anything until you need to.**

## The Scaling Progression

Almost every web application follows the same growth path. Not because there's one right architecture, but because the bottlenecks appear in a predictable order. At each stage, one component becomes the constraint, and the solution at that stage is well understood.

```
Stage 1: Single server
  Everything on one box. App + DB + maybe Redis.
  Handles: 0-1,000 users, < 100 RPS

Stage 2: Separate the database
  App server(s) + dedicated DB instance.
  Handles: 1,000-10,000 users, 100-1,000 RPS

Stage 3: Add caching + CDN
  Cache hot reads (Redis). Serve static assets from CDN.
  Handles: 10,000-100,000 users, 1,000-10,000 RPS

Stage 4: Horizontal app scaling + read replicas
  Multiple app servers behind LB. DB read replicas.
  Handles: 100,000-1,000,000 users, 10,000-50,000 RPS

Stage 5: Async processing + service separation
  Queues for background work. Split into services by domain.
  Handles: 1,000,000-10,000,000 users, 50,000-200,000 RPS

Stage 6: Database sharding + specialised storage
  Shard writes. Purpose-built databases for specific workloads.
  Handles: 10,000,000+ users, 200,000+ RPS
```

Most applications never get past stage 3. The ones that do usually have years of growth between each stage. **Premature scaling is worse than no scaling** — it adds complexity, operational cost, and makes everything harder to change. Scale when the metrics tell you to, not when you imagine you might need to.

## Stage 1: Single Server (0-1,000 users)

Everything runs on one machine — the application, the database, maybe a Redis instance for sessions.

```
┌─────────────────────────────┐
│         Single Server       │
│                             │
│  ┌─────┐  ┌──────┐  ┌───┐  │
│  │ App │  │ DB   │  │Red│  │
│  │     │  │(PG)  │  │is │  │
│  └─────┘  └──────┘  └───┘  │
│                             │
│  t3.medium / t3.large       │
│  2-4 vCPU, 4-8 GB RAM      │
│  ~$30-60/month              │
└─────────────────────────────┘
```

**This is fine.** A single t3.large can handle a surprising amount of traffic. Postgres on the same box handles thousands of QPS for simple queries. Don't be embarrassed by a single-server architecture — many successful products run this way for months or years.

**What to focus on at this stage:**

- Write clean code that's easy to change later
- Use database indexes from the start
- Use connection pooling (even locally)
- Deploy behind a reverse proxy (nginx/Caddy) for TLS and static file serving
- Set up basic monitoring (CPU, memory, disk, response times)

**When to move to stage 2:** CPU consistently above 70%, database queries competing with app server for memory, or you need independent scaling of app and DB.

## Stage 2: Separate the Database (1,000-10,000 users)

Move the database to its own instance. This is the first real architectural decision, and it's almost always the right one.

```
┌───────────┐        ┌──────────────┐
│ App Server│───────→│   Database   │
│           │        │   (RDS)      │
│ t3.medium │        │ db.t3.medium │
│ ~$30/mo   │        │ ~$50/mo      │
└───────────┘        └──────────────┘
```

**Why separate:**

- App and DB can scale independently
- DB gets dedicated memory (critical for caching data pages)
- App server can be replaced/restarted without touching data
- Managed database (RDS) handles backups, patches, failover

**What to focus on at this stage:**

- Set up proper database monitoring (slow query log, connection count, replication lag if using replicas)
- Add application-level health checks
- Configure automated backups
- Start tracking key metrics: RPS, p95 latency, error rate, DB QPS

**When to move to stage 3:** Database reads are the bottleneck (most queries are repeated reads), static assets are eating bandwidth, or p95 latency is rising.

## Stage 3: Add Caching + CDN (10,000-100,000 users)

This is the highest-leverage stage. A cache (Redis/ElastiCache) in front of your database eliminates 80-90% of read traffic. A CDN serves static assets from edge locations worldwide. Together, they typically buy a 5-10x improvement in effective capacity.

```
                         ┌─────────┐
                    ┌───→│   CDN   │ (static assets, images, JS/CSS)
                    │    └─────────┘
┌──────┐    ┌──────┤
│Users │───→│  LB  │    ┌───────────┐     ┌───────┐     ┌────────┐
│      │    │      │───→│ App Server│────→│ Redis │────→│   DB   │
└──────┘    └──────┘    │           │     │ Cache │     │ (RDS)  │
                        └───────────┘     └───────┘     └────────┘
                                          cache-aside
                                          hit rate: 80-90%
                                          DB load: drops 5-10x
```

**What changes:**

- Add Redis/ElastiCache for hot data (sessions, product catalogue, user profiles)
- Put static assets behind CloudFront/Cloudflare
- Add a load balancer even with one app server (makes adding more seamless later)
- The DB now handles 10-20% of original read traffic — the rest comes from cache

**From lesson 04 (caching):** Cache-aside is the default pattern. TTL of 5-60 minutes for most data. Monitor hit rate — below 80% means your cache isn't configured well.

**Cost at this stage:**
| Component | Instance | Monthly cost |
| --- | --- | --- |
| App server | t3.medium | ~$30 |
| Database | db.r6g.large | ~$200 |
| Redis | cache.r6g.large | ~$150 |
| CDN | CloudFront | ~$20-50 |
| Load balancer | ALB | ~$20 |
| **Total** | | **~$420-450** |

This handles 10,000-100,000 users depending on the workload. For $450/month. The instinct to immediately go multi-server is often premature — a single well-cached app server handles more than people expect.

**When to move to stage 4:** Single app server CPU is consistently above 70%, or you need high availability (one server = single point of failure).

## Stage 4: Horizontal App Scaling + Read Replicas (100,000-1,000,000 users)

The app server becomes the bottleneck. Add more app servers behind the load balancer. If read traffic is still growing, add database read replicas.

```
                    ┌─────────────┐
               ┌───→│ App Server 1│───┐
┌──────┐  ┌───┤    └─────────────┘   │    ┌───────┐    ┌──────────┐
│Users │→ │LB │───→│ App Server 2│───┼───→│ Redis │───→│ Primary  │
└──────┘  └───┤    └─────────────┘   │    │ Cache │    │   (RW)   │
               └───→│ App Server 3│───┘    └───────┘    └────┬─────┘
                    └─────────────┘                          │ replication
                                                     ┌──────┼──────┐
                                                     ▼      ▼      ▼
                                                   Replica Replica Replica
                                                    (RO)    (RO)    (RO)
```

**Critical requirement: stateless app servers.** If your app stores session data, uploaded files, or any state on the local filesystem, you can't scale horizontally. Sessions must be in Redis or a database. Uploads must go to S3. Configuration must come from environment variables or a config service.

**The stateless checklist:**
| State type | Wrong (stateful) | Right (stateless) |
| --- | --- | --- |
| Sessions | In-memory or filesystem | Redis or database |
| File uploads | Local disk | S3 / object storage |
| Cache | In-process only | Redis (shared) |
| Config | Local files that differ per server | Environment variables / SSM |
| Scheduled tasks | Cron on one server | Dedicated scheduler / Lambda |

**Read replicas — from lesson 03:**

- Each replica handles roughly the same read QPS as the primary
- Replication lag: typically 10-100 ms (async)
- Reads go to replicas, writes go to primary
- Application must handle stale reads (user writes, then immediately reads from a replica that hasn't caught up)

**When to move to stage 5:** Synchronous processing is causing latency spikes, background work is competing with request handling, or individual services need independent deploy/scale cycles.

## Stage 5: Async Processing + Service Separation (1,000,000-10,000,000 users)

At this scale, you can't do everything in the request path. Order processing, email, notifications, analytics — these move to queues. And the monolith application starts splitting along domain boundaries, not because microservices are trendy, but because different parts have different scaling needs.

```
┌──────┐    ┌────┐    ┌───────────────┐    ┌───────┐    ┌──────────┐
│Users │───→│ LB │───→│  API Service  │───→│ Redis │───→│ Database │
└──────┘    └────┘    └───────┬───────┘    └───────┘    └──────────┘
                              │
                         publish events
                              │
                        ┌─────┴─────┐
                        │   Queue   │ (SQS / Kafka)
                        └─────┬─────┘
                    ┌─────────┼─────────┐
                    ▼         ▼         ▼
              ┌──────────┐ ┌───────┐ ┌───────────┐
              │ Email    │ │Payment│ │ Analytics │
              │ Service  │ │Service│ │ Service   │
              └──────────┘ └───────┘ └───────────┘
```

**What moves to async (from lesson 05):**

- Anything that takes > 1-2 seconds
- Anything that can fail and be retried (payments, external APIs)
- Anything the user doesn't need to wait for (emails, notifications, analytics)

**When to split services:**

The wrong reason: "microservices are best practice." The right reasons:

| Signal                      | Example                                                                       |
| --------------------------- | ----------------------------------------------------------------------------- |
| Different scaling needs     | Search handles 10x the traffic of checkout                                    |
| Different deploy cadences   | Marketing wants to update the CMS daily; payments change quarterly            |
| Different reliability needs | Payment processing needs 99.99%; the recommendation engine can tolerate 99.9% |
| Team boundaries             | Two teams stepping on each other in the same codebase                         |
| Different technology needs  | ML model serving in Python, API in Go                                         |

**Service boundaries should follow domain boundaries**, not technical layers. "Order service" and "Payment service" are good boundaries. "Database service" and "Cache service" are not — those are infrastructure, not domains.

**Splitting too early is worse than splitting too late.** A well-structured monolith is easier to split later than a poorly-designed set of microservices is to fix. If you're not sure where the boundaries are, keep it monolithic.

**Communication between services:**
| Pattern | Use when | Latency | Coupling |
| --- | --- | --- | --- |
| Synchronous HTTP/gRPC | Need response to continue (get user profile to authorise request) | Low (ms) | High — caller blocks |
| Async queue (SQS/Kafka) | Don't need immediate response (send email after order) | Variable | Low — fire and forget |
| Event bus (SNS/Kafka) | Multiple services need to react to one event | Variable | Lowest — publisher doesn't know consumers |

## Stage 6: Database Sharding + Specialised Storage (10,000,000+ users)

Vertical scaling and read replicas have limits. The biggest RDS instance handles ~100,000 QPS. If your write volume exceeds what a single primary can handle, or your data is too large for one instance, you need sharding — splitting data across multiple database instances.

```
                        ┌──────────────┐
                   ┌───→│  Shard 0     │  Users A-F
                   │    │  (RW)        │
┌──────────┐       │    └──────────────┘
│ Shard    │───────┤    ┌──────────────┐
│ Router   │       ├───→│  Shard 1     │  Users G-N
│          │       │    │  (RW)        │
└──────────┘       │    └──────────────┘
                   │    ┌──────────────┐
                   └───→│  Shard 2     │  Users O-Z
                        │  (RW)        │
                        └──────────────┘
```

**Sharding is hard.** It's the last resort, not the first option.

**Why sharding is painful:**

- Cross-shard queries are expensive or impossible (JOINs across shards)
- Rebalancing data when adding shards is complex
- Application logic must be shard-aware (route queries to the right shard)
- Transactions across shards require distributed consensus
- Schema changes must be applied to every shard

**Choosing a shard key:** The shard key determines which shard holds which data. A bad shard key creates hot spots (one shard gets all the traffic). A good shard key distributes data and queries evenly.

| Shard key         | Good for                                              | Bad for                                        |
| ----------------- | ----------------------------------------------------- | ---------------------------------------------- |
| User ID           | User-centric apps (each user's data on one shard)     | Queries across users ("most popular products") |
| Geographic region | Region-scoped data, compliance (data stays in region) | Users who travel or global queries             |
| Time (date)       | Time-series data, logs                                | Latest shard gets all writes (hot shard)       |
| Hash of ID        | Even distribution                                     | Range queries (all orders from last month)     |

**At this stage, you also start using specialised storage:**

| Workload                 | Technology                          | Why not the main DB                                                                   |
| ------------------------ | ----------------------------------- | ------------------------------------------------------------------------------------- |
| Full-text search         | Elasticsearch / OpenSearch          | Postgres full-text search doesn't scale to millions of documents with complex queries |
| Time-series metrics      | TimescaleDB / InfluxDB / CloudWatch | Optimised for append-heavy, time-range queries                                        |
| Graph relationships      | Neo4j / Neptune                     | "Friends of friends who liked X" is a nightmare in SQL                                |
| Session / ephemeral data | Redis / DynamoDB                    | Don't load your primary database with throwaway data                                  |
| Large binary objects     | S3                                  | Databases aren't filesystems                                                          |
| Analytics / OLAP         | Redshift / BigQuery / ClickHouse    | OLTP databases aren't built for scanning billions of rows                             |

The rule: **use the main database (Postgres) until a specific workload outgrows it, then move that workload to a purpose-built system.** Don't start with 6 databases.

## The Decision Flowchart

When something is slow or at capacity, walk through this in order. Stop at the first step that solves the problem.

```
System is slow or hitting limits
          │
          ▼
┌─────────────────────┐     ┌─────────────────────────────────────┐
│ 1. Is the code      │ yes │ Fix queries, add indexes, fix N+1,  │
│    inefficient?      │────→│ add connection pooling, reduce       │
│                     │     │ payload sizes                        │
└─────────┬───────────┘     └─────────────────────────────────────┘
          │ no
          ▼
┌─────────────────────┐     ┌─────────────────────────────────────┐
│ 2. Can you buy a    │ yes │ Bigger DB instance, bigger app       │
│    bigger instance?  │────→│ server, more memory for Redis.       │
│                     │     │ Cheapest, fastest fix.               │
└─────────┬───────────┘     └─────────────────────────────────────┘
          │ no
          ▼
┌─────────────────────┐     ┌─────────────────────────────────────┐
│ 3. Is it read-heavy?│ yes │ Add Redis cache (cache-aside).       │
│    (> 70% reads)    │────→│ Add CDN for static assets.           │
│                     │     │ Add read replicas if cache isn't     │
└─────────┬───────────┘     │ enough.                              │
          │ no              └─────────────────────────────────────┘
          ▼
┌─────────────────────┐     ┌─────────────────────────────────────┐
│ 4. Is it write-heavy│ yes │ Add queue (SQS/Kafka) to buffer     │
│    or bursty?       │────→│ writes. Batch writes. Move non-      │
│                     │     │ critical writes to async.            │
└─────────┬───────────┘     └─────────────────────────────────────┘
          │ no
          ▼
┌─────────────────────┐     ┌─────────────────────────────────────┐
│ 5. Is a single      │ yes │ Multiple app servers behind LB.      │
│    server maxed?    │────→│ Ensure app is stateless first.       │
│                     │     │                                      │
└─────────┬───────────┘     └─────────────────────────────────────┘
          │ no
          ▼
┌─────────────────────┐     ┌─────────────────────────────────────┐
│ 6. Is the DB the    │ yes │ Shard by a well-chosen key.          │
│    write bottleneck │────→│ Consider DynamoDB or MongoDB for     │
│    at scale?        │     │ native sharding. Move specific       │
└─────────┬───────────┘     │ workloads to specialised storage.    │
          │ no              └─────────────────────────────────────┘
          ▼
┌─────────────────────┐     ┌─────────────────────────────────────┐
│ 7. Are services     │ yes │ Split along domain boundaries.       │
│    coupled or have  │────→│ Independent deploy, scale, fail.     │
│    different needs? │     │ Communicate via queues/events.       │
└─────────────────────┘     └─────────────────────────────────────┘
```

## Worked Example: E-Commerce Platform

Let's trace an e-commerce platform through the stages, tying together all the lessons.

### Starting point (1,000 users)

Single server. Postgres. A Node.js or Go API. Everything is simple and that's correct.

- 10 RPS average (lesson 01: 1,000 users × ~10 requests/day ÷ 10^5)
- Postgres handles this without breaking a sweat
- Total cost: ~$50/month

### Growing (50,000 users)

- 500 RPS average, 1,500-2,500 RPS peak
- Product pages are the hot path — 80% of traffic is browsing the catalogue

**Actions:**

1. Move DB to RDS (db.r6g.large) — $200/month
2. Add Redis for product catalogue and sessions — cache-aside, 5 min TTL
3. Add CloudFront for images and static assets
4. DB now sees ~100-500 QPS instead of 1,500-2,500 (cache absorbs 80%)

Cost: ~$450/month. Handles 5-10x current traffic.

### Scaling (500,000 users)

- 5,000 RPS average, 15,000-25,000 RPS peak
- Flash sales cause 10x spikes above average
- Order processing is slow (payment + inventory + email in one request)

**Actions:**

1. Horizontal app scaling: 3 app servers behind ALB (stateless — sessions in Redis)
2. Read replicas (2) for the DB — handles report queries and read overflow
3. Queue for order processing: API accepts order (50 ms) → SQS → workers handle payment, inventory, email independently
4. Bigger DB instance: db.r6g.xlarge for the primary

```
                 ┌─────────┐
            ┌───→│ CDN     │ (images, JS, CSS)
            │    └─────────┘
┌──────┐  ┌─┴──┐  ┌──────┐    ┌───────┐    ┌─────────┐
│Users │→ │ LB │→ │App x3│───→│ Redis │───→│ Primary │
└──────┘  └────┘  └──┬───┘    └───────┘    └────┬────┘
                     │                          │
                  ┌──┴───┐               ┌──────┼──────┐
                  │ SQS  │               ▼      ▼      ▼
                  └──┬───┘            Replica Replica
            ┌───────┼───────┐
            ▼       ▼       ▼
         Payment  Email  Inventory
         Worker   Worker  Worker
```

Cost: ~$2,000/month. Handles 10x current traffic.

### At scale (5,000,000 users)

- 50,000 RPS average, 150,000+ RPS peak
- Search is slow on Postgres at this data volume
- Analytics queries are killing the primary DB
- Different teams own different domains

**Actions:**

1. Elasticsearch for product search (offloads complex queries from Postgres)
2. Analytics to Redshift (CDC from Postgres via Kafka → S3 → Redshift)
3. Split into services: Product API, Order API, User API, Search API
4. Kafka for event streaming between services (multiple consumers need order events)
5. DynamoDB for shopping cart (key-value access pattern, scales to zero, no connection limits)

Cost: ~$15,000-30,000/month. But you have a team of engineers and revenue to match.

## Anti-Patterns

Patterns that seem reasonable but cause more problems than they solve.

### Premature microservices

**The mistake:** Starting with 10 microservices from day one because "Netflix does it."

**Why it hurts:** Network calls between services add latency and failure modes. Distributed tracing, service discovery, deployment orchestration — all of this is overhead. With 2 engineers and 1,000 users, a monolith is faster to develop, deploy, and debug.

**The rule:** Start monolithic. Split when team size or scaling needs force it, not before.

### Premature sharding

**The mistake:** Sharding the database at 10,000 users because "we might go viral."

**Why it hurts:** Sharding makes every query harder, every migration harder, every debugging session harder. A single Postgres instance handles 50,000+ QPS. Most applications will never need to shard.

**The rule:** Vertical scaling → read replicas → cache → shard. In that order. Most apps stop at cache.

### Cache everything

**The mistake:** Caching every database query with no strategy.

**Why it hurts:** Cache invalidation becomes impossible. Stale data causes bugs. Memory costs grow. Cache hit rate drops because you're caching rarely-accessed data.

**The rule:** Cache the hot path. The top 20% of data that gets 80% of reads (lesson 04). Monitor hit rate. If it's below 80%, you're caching the wrong things.

### Queue everything

**The mistake:** Putting every operation on a queue because "async is faster."

**Why it hurts:** Queues add latency (user doesn't get immediate feedback), complexity (consumers, DLQs, monitoring), and failure modes (consumer lag, message ordering). A 5 ms database write doesn't need a queue.

**The rule:** Sync by default. Async when the operation is slow, can fail, or the user doesn't need the result (lesson 05).

### Over-provisioning for peak

**The mistake:** Running 20 servers 24/7 because traffic peaks at 20x for 2 hours a day.

**Why it hurts:** You're paying for 20 servers when you need 2 for 22 hours a day.

**The rule:** Auto-scale. ECS and Kubernetes scale based on CPU/memory. Lambda scales automatically. Set a minimum for baseline traffic, scale up for peaks, scale down after.

## Key Principles

1. **Measure before you scale.** Without metrics, you're guessing. Know your RPS, p95 latency, DB QPS, cache hit rate, and queue depth before making decisions.

2. **Do the cheapest thing first.** An index is cheaper than a replica. A replica is cheaper than a cache. A cache is cheaper than sharding. Vertical scaling is cheaper than horizontal. Always start simple.

3. **Scale the bottleneck, not everything.** If the database is the bottleneck, adding more app servers doesn't help. Identify the constraint (lesson 02: trace the request), then fix that specific thing.

4. **Stateless app servers, stateful storage.** App servers should be disposable — any request can go to any server. All state lives in the database, cache, or object storage. This is the foundation of horizontal scaling.

5. **Accept trade-offs explicitly.** Caching trades consistency for speed. Queues trade immediacy for reliability. Replicas trade write consistency for read throughput. Sharding trades query flexibility for write scalability. Know what you're giving up.

6. **Complexity is a cost.** Every component you add (cache, queue, service, database) is something to monitor, debug, pay for, and reason about. The simplest architecture that meets your requirements is the best one.
