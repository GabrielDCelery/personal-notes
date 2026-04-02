# Scale Gut-Check — Deep Dive

The gut-check tiers from the quick-math doc tell you _what category_ you're in. This doc explains _what that actually means_ — the dominant failure mode at each tier, what the architecture is solving, and where common solutions actually belong.

## The Core Question at Each Tier

Each tier has one dominant bottleneck. The architecture is just the answer to that bottleneck.

| Scale          | Core bottleneck                                | Scale answer                                                   | Resilience answer                                                                     |
| -------------- | ---------------------------------------------- | -------------------------------------------------------------- | ------------------------------------------------------------------------------------- |
| ~100 RPS       | Nothing yet                                    | Boring stack, ship it                                          | Backups — accept downtime, restore and move on                                        |
| ~1,000 RPS     | DB query performance                           | Indexes, connection pooling                                    | Alerting — know when you're down                                                      |
| ~10,000 RPS    | DB read load                                   | Caching (Redis)                                                | Replica (hot standby) + monitoring — harden each component so it survives what's next |
| ~100,000 RPS   | DB write throughput + single-server app limits | Sharding, CQRS, queues as write buffer, horizontal app scaling | For every component: if this dies, does everything else die with it?                  |
| ~1,000,000 RPS | Everything is the bottleneck                   | Custom everything                                              | Chaos engineering — assume failure, test it deliberately                              |

**The thread:** each tier has a different failure mode. Knowing the failure mode tells you the solution — you don't need to memorise the solutions independently.

---

## ~100 RPS — Nothing yet

A small SaaS, internal tool, or startup MVP.

- **What this is:** ~8.6M requests/day. A decent indie product, B2B app with a few hundred paying customers, a company internal portal.
- **Why a boring stack works:** a single t3.medium + RDS Postgres handles this with CPU at 5%. You don't need to think.
- **Where it breaks:** not from load. It breaks from dumb mistakes — a missing WHERE clause, loading 10K rows in a loop, a synchronous email send on every request.
- **Resilience:** backups. If it goes down you restore from a snapshot and accept the downtime. At this scale that's usually fine. Add Sentry so you know it's down before a user emails you.

---

## ~1,000 RPS — DB query performance

A growing startup, a popular internal tool, a mid-size SaaS.

- **What this is:** ~86M requests/day. You have real users.
- **Why indexes matter here:** at 100 RPS a full table scan on 50K rows is annoying. At 1,000 RPS it's a fire. Each request waiting 200ms on a seq scan means 200 queued requests piling up behind it.
- **Why connection pooling matters here:** Postgres default max connections ~100. At 1,000 RPS with naive connection-per-request you saturate the DB instantly. PgBouncer or an in-app pool keeps you alive.
- **One server still works** — but it's a beefy one and you're watching it.
- **Resilience:** alerting. Health checks + PagerDuty so you know when it's down. Connection pooling also has a resilience angle here — without it, a traffic spike exhausts DB connections and kills the server entirely, not just slows it down.

---

## ~10,000 RPS — Helpers for each pressure point

A well-known consumer app, a high-traffic media site, a popular API.

- **What this is:** ~864M requests/day. You're a real product.
- **The fundamental architecture doesn't change shape.** You still have one app, one DB, one cache layer. But vertical scaling has stopped being the answer — you can't just provision a bigger server. Instead you introduce helpers for each individual pressure point.
- **Redis/Memcached** absorbs hot reads (user sessions, product listings, popular content) without touching Postgres at all. A good cache hit rate (60–80%) means your DB might only see 2–4K actual queries/sec — well within a single instance's capacity.
- **CDN** absorbs static load (JS, CSS, images) before it hits origin at all. Not the dominant bottleneck answer, but it fits the tier: offload the pressure before it reaches the component that can't handle it.
- **A single well-tuned server can handle 10K RPS for most workloads.** You don't add more app servers because one can't cope — you add them because you need zero-downtime deploys and can't afford a single hardware failure to take you down. That's a reliability decision, not a scale decision.
- **Resilience:** this is where you harden each component individually so it survives what's coming next. A read replica gives you a hot standby — if the primary dies, you promote and you're back online in seconds. Monitoring gives visibility into which component is struggling before it becomes an outage. Circuit breakers on external API calls mean a downstream failure doesn't cascade into your failure.

**The key insight:** you're not changing what the system does, you're making sure each component is hardened for the load it will face. Any component you didn't prepare becomes the weak link. The system survives not because it's bigger but because each individual piece is ready.

---

## ~100,000 RPS — The shape of work changes

A major consumer product, a large fintech or logistics platform, a top-50 website.

- **What this is:** ~8.6B requests/day.
- **You're no longer building a traditional application.** The familiar patterns stop working — not because they're badly implemented, but because the load makes them structurally impossible. A synchronous order placement that works fine at 10K will cause cascading failures at 100K. The answer isn't a bigger server or a better cache — it's changing the shape of the work itself.
- **The dominant bottleneck:** two things converge. A single Postgres instance tops out around 1K write transactions/sec. And the app tier itself starts hitting its ceiling — Node.js caps at roughly 10–50K RPS per node, Nginx around 50–100K RPS. Even Go, which can push past 100K RPS, is at its limit. At 10K RPS the app server was fine and only the DB needed help; at 100K RPS both layers need horizontal scaling — not for HA, but just to survive the load.

### Database strategies

- **Sharding** — splits data across multiple DB instances (by user ID, region, etc.) so writes scale horizontally. Each shard handles a fraction of the total write load.
- **CQRS** — separate read and write models entirely. Writes go to Postgres; reads come from a denormalized store (Redis, Elasticsearch, a read-optimized DB). Eliminates contention between the two workloads.
- **NoSQL for specific access patterns** — DynamoDB or Cassandra for high-write, simple-lookup data (user sessions, event logs, activity feeds). Not replacing Postgres — coexisting with it for the workloads it's better suited to.

### Queues as a write buffer

At 100K RPS, queues become a core architectural pattern for writes — not just for async jobs. Instead of every request writing synchronously to the DB, you accept the request, put a message on a queue (SQS, Kafka, RabbitMQ), return immediately, and let workers process it async. This levels out write spikes dramatically.

**The pattern:** placing an order means writing to the DB, emitting an event, and an orchestrator picks it up asynchronously. The app doesn't "handle" the order — it accepts it and hands it off. The rules of a traditional request-response app no longer apply. Every architectural decision matters, and every component without a backup plan becomes the one that takes you down.

**Important distinction:** queues aren't a 100K RPS invention. Any system that has slow or unreliable async work — sending emails, processing payments, resizing images, firing webhooks — needs a queue at 10K RPS, 1K RPS, even 100 RPS. The moment work takes longer than a request should wait, can fail and needs retrying, or spikes unpredictably, you need a queue. The difference at 100K RPS is that queues become a _systemic_ architectural pattern — not just a place to offload slow tasks, but the primary way the system processes writes at all.

### Infrastructure

- **CDN** — a large proportion of traffic is static or cacheable at the edge. Without a CDN, 100K RPS all hitting origin servers is brutal. CloudFront or Fastly absorbs 60–80% of it.
- **Multi-region** — latency from routing global users to a single region becomes measurable and user-visible. Route users to the nearest region.
- **Autoscaling** — you're not manually provisioning servers. You define scaling policies (CPU > 70% for 2 minutes → add instances) and let the platform handle it.
- **Rate limiting** — protect yourself from traffic spikes, bad actors, and runaway clients. Per-user, per-IP, per-API-key limits at the edge.
- **Circuit breakers** — if a downstream service is degraded, stop hammering it. Fail fast, return a cached or degraded response, let it recover.
- **Service decomposition** — not necessarily full microservices, but splitting by scaling profile. Your image upload service has different scaling needs than your auth service. Separating them lets you scale each independently.

### Resilience — no single point of failure

At 100K RPS you have enough moving parts that any component without a backup plan becomes the thing that kills you. The question to ask for every component is: **if this dies, does everything else die with it?**

- One DB with no replica → DB goes down, you're offline
- One region → region has an outage, you're offline
- No autoscaling → traffic spike kills the one server
- No rate limiting → one bad actor takes you down
- One app server → it crashes during a deploy, you're offline

You can execute perfectly everywhere else, but the one component you didn't prepare for is the one that kills you. Multi-region, autoscaling, circuit breakers, and rate limiting aren't just performance tools — they're the answer to "what happens when this specific thing dies?"

---

## ~1,000,000 RPS — Custom everything

Google, Netflix, Instagram, Stripe.

- **What this is:** ~86B requests/day.
- **Why off-the-shelf breaks:** even sharded Postgres can't handle this write load. These companies build or heavily modify their own storage engines, networking stacks, and load balancers.
  - Google built Bigtable and Spanner
  - Facebook built Cassandra, then RocksDB
  - Discord moved from Cassandra to a custom Rust/ScyllaDB setup
- **The real challenge isn't throughput — it's consistency and latency at scale.** Making a tweet appear in 500M timelines within seconds, or ensuring a payment is idempotent across 50 data centres, is the actual hard problem. Throughput is solved by throwing hardware at it; correctness at this scale requires entirely new systems.
- **Resilience:** chaos engineering. At this scale failures are unpredictable, so you test them deliberately. Netflix's Chaos Monkey randomly kills production instances to verify the system survives. You assume failure and design for it — not as an edge case but as the normal operating condition.

---

## What doesn't belong in this table

**Read replicas** and **multiple app servers at 10K RPS** are reliability decisions, not scale answers. At 10K RPS, one app server can genuinely handle the load — you add more for zero-downtime deploys and redundancy, not because one can't cope. Similarly, a read replica at 10K gives you a hot standby for failover, not extra read throughput (if caching is working, you don't need it for that).

At 100K RPS this changes — multiple app servers cross into being a scale answer because a single node is at or beyond its ceiling for most workloads.

The cleaner framing: **if caching is effective, you may not need read replicas for throughput at 10K RPS at all.** The right justification for a read replica at this tier is HA, not performance.

---

## Summary

```
         scale answer                          resilience answer
~100 RPS     ship it                           back it up
~1K RPS      index, pool                       alerting
~10K RPS     cache                             replica + monitoring
~100K RPS    shard, CDN, queues                no single point of failure
~1M RPS      custom everything                 chaos engineering
```

Each jump isn't "more of the same" — it's a fundamentally different problem class. The jump from 1K to 10K is operational. The jump from 10K to 100K is architectural. The jump to 1M is organisational.

**The narrative arc:**

- **10K:** the architecture doesn't change shape, but vertical scaling stops being the answer. You introduce a helper for each pressure point and harden each component individually.
- **100K:** you're not building a traditional app anymore. The shape of work has to change — async, event-driven, orchestrated. Every decision carries real consequences.

The resilience thread follows the same pattern: each tier has a different resilience _philosophy_, not just a checklist. Backups accept downtime. Alerting means you know when you're down. Hardening components means they survive individually. No single point of failure means the system survives any one component dying. Chaos engineering means you've stopped trying to prevent failure and started designing for it.
