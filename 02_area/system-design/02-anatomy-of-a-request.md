# Anatomy of a Request — Distilled

Knowing where time goes in a request tells you what to optimise — and what to leave alone.

## The Core Mental Model: 0 / 1 / 5 / 50

Every layer in a request belongs to one of four cost buckets. A healthy request stays in the first three.

```
Scale: log10  |    1ms         10ms        100ms       500ms      |
              0----+------------+-----------+-----------+----------+

~0 ms  ·                              cached or reused (DNS, TCP, TLS, JWT auth)
~1 ms  ███                            infrastructure hop (load balancer, parsing, serialisation)
~5 ms  ██████████                     one DB query done right (indexed lookup)
~50 ms ████████████████████           unnecessary network call per request
~500ms ██████████████████████████████ broken data access
```

**A healthy request = 0 (reused) + a few 1s (hops) + a 5 (DB) = ~7 ms total.**

Every 50 ms chunk means a network call you shouldn't be making per request. Every 500 ms chunk means structurally broken data access — fix the pattern, not the code.

## The Full Journey

```
User's browser
  ├─ 1. DNS resolution       → find the IP (cached = free)
  ├─ 2. TCP handshake        → establish connection (keep-alive = free)
  ├─ 3. TLS handshake        → encrypt (session resumption = free)
  ▼
CDN / Edge
  ├─ 4. Edge cache check     → hit: serve and stop. miss: forward.
  ▼
Load balancer                → +0.5–2 ms
  ▼
App server
  ├─ 5. Parse request        → ~0.1–1 ms
  ├─ 6. Authenticate         → 0.1 ms (JWT) … 100 ms (OAuth)
  ├─ 7. Business logic       → 1–5% of total for I/O-bound apps
  ├─ 8. Database query       → ~77% of total time in a healthy request
  ├─ 9. Serialise response   → negligible under 100 KB
  ▼
Back through load balancer → CDN → browser
```

Not every request hits every layer. An internal API call skips DNS and CDN. A CDN cache hit skips everything below the edge. The full path is the worst case.

## Connection Cost: DNS, TCP, TLS

These three steps only matter on the **first request**. After that they're cached or reused and cost nothing.

```
Scale: log10  |  1ms       10ms      100ms     300ms |
              0---+---------+----------+---------+---+

DNS cold lookup       |█████████████████   | 20–100 ms   (walking root → TLD → authority)
DNS cached            |                    | ~0 ms
TCP same data centre  |██                  | ~1 ms
TCP cross-ocean       |████████████████████| 150–300 ms
TLS 1.3 new session   |██                  | +1 round trip (~1 ms same DC)
TLS 1.3 resumed       |                    | ~0 ms
```

Keep-alive and session resumption make all three free. Most frameworks do this by default. **If DNS or TCP show up in steady-state traces, connection reuse has been accidentally disabled.** Connection pooling (PgBouncer, HikariCP) does the same for DB connections — opening a new Postgres connection costs 5–20 ms; a pool makes it near-zero.

## Authentication

This is where the biggest avoidable latency hides. The cost difference is three orders of magnitude.

```
Scale: log10  |  0.1ms    1ms       10ms      100ms |
              0----+--------+---------+---------+---+

JWT (local verify)          |██                  | ~0.1 ms   no network call
Session → Redis             |████████            | ~1 ms     one round trip
Session → DB                |████████████        | ~5 ms     network + query
OAuth / Auth0 introspect    |████████████████████| 10–100 ms HTTP call to external service
```

**Why JWT needs no network call:** the token encodes identity and permissions directly, signed with a private key. Verification is pure CPU — decode the payload, check the signature against a locally held public key. A session token is an opaque ID that means nothing without a lookup.

The trade-off: JWTs can't be revoked until expiry. Handle this with short expiry windows (15 min) + refresh tokens, not a blocklist (which reintroduces a network call).

**In microservices: verify once at the edge, pass a trusted internal token downstream.** If each service independently calls Auth0, you pay 50 ms per service per request.

## Database Query

The DB is ~77% of your time in a healthy same-region request. This is where optimisation effort belongs.

```
Scale: log10  |  1ms      10ms      100ms    1,000ms  10,000ms |
              0---+----------+---------+---------+----------+--+

Primary key / indexed     |████                | 1–5 ms
JOIN on indexed cols      |███████             | 2–20 ms
Aggregation               |██████████████      | 5–500 ms
Full table scan           |████████████████████| 100–10,000 ms
```

The real cost is usually hidden across multiple sequential queries, each paying a network round trip:

```
What you think:   1 query ≈ 3 ms

What actually happens:
  auth check        2 ms
  get order         3 ms
  get order items   4 ms
  get shipping      2 ms
  audit log         3 ms
  ─────────────────────
  Total:           14 ms   (5 round trips)
```

**Batch them:**

```
Bad:  5 queries × 3 ms = 15 ms   sequential round trips
Good: 1 JOIN query     =  5 ms   one round trip, DB does the work
Good: 1 IN clause      =  4 ms   one round trip, multiple rows
```

Missing index = full table scan = 100–10,000 ms. Run `EXPLAIN ANALYZE` on any slow query — the planner tells you exactly why it's slow.

## Latency Budget: Good vs Broken

```
Layer                       Good      Bad       What went wrong
─────────────────────────────────────────────────────────────────
DNS, TCP, TLS (reused)       0 ms     50 ms    TTL set to 0, no keep-alive
Load balancer                0.5 ms    0.5 ms  —
Auth                         0.1 ms   50 ms    External OAuth per request
Database                     5 ms    500 ms    Missing index / N+1 / no pool
Serialisation                0.2 ms   50 ms    10,000 rows un-paginated
─────────────────────────────────────────────────────────────────
Total                        ~7 ms   ~666 ms   95× slower, no code changes needed
```

Every "bad" entry is configuration — not a language or framework problem.

## Multi-Service Requests

When a request fans out to multiple services:

```
sequential total = sum of all calls
parallel total   = max of all calls
```

```
API Gateway
  ├─→ User Service       3 ms
  ├─→ Order Service       5 ms
  │     └─→ Inventory     4 ms   (depends on Order — must be sequential)
  └─→ Recommendations    15 ms

Sequential: 3 + 5 + 4 + 15 = 27 ms
Parallel:   max(3, 5+4, 15) = 15 ms   ← free latency reduction
```

Always parallelise independent calls.

**The microservice hop tax:** each service-to-service call costs ~1–3 ms (network + load balancer + serialise/deserialise) before any logic runs. 5 sequential hops = 10 ms overhead before your code does anything. A monolith making 5 function calls pays ~0.001 ms for the same. Prefer wide and shallow call chains over deep and narrow ones.

## Timeouts and Cascading Failure

Without a timeout, a single slow downstream service can stall your entire system:

```
Order Service hangs (no timeout):
  100 users hit the endpoint
  100 threads blocked waiting
  Thread pool exhausted
  API can't serve ANY request — including unrelated ones
```

**Set a timeout on every outbound call.** Safe conservative defaults while you gather p99 data:

| Call type        | Default timeout |
| ---------------- | --------------- |
| Internal service | 500 ms – 1 s    |
| External API     | 2 – 5 s         |
| DB query         | 5 – 10 s        |
| Redis            | 100 – 500 ms    |

Once you have p99 from production, tighten to ~2–3× p99. Even a bad timeout is vastly better than none — most cascading failures happen because no timeout was set at all.

A **circuit breaker** stops the cascade: after N failures, stop calling the service for a cooldown period and return a fallback instead of hanging.

## Reading a Latency Profile

Use percentiles, not averages. A service averaging 10 ms can have p99 at 2 seconds. Averages hide the tail.

**p50** = what a normal request looks like. **p99** = what the worst 1 in 100 users experiences, and what you calibrate timeouts against.

```
p99 ≈ 2×  p50  → healthy — minor variance (GC pauses, occasional cache miss)
p99 ≈ 10× p50  → something occasionally goes wrong (lock contention, rare slow query)
p99 ≈ 100× p50 → bimodal — two distinct code paths (cache hit vs miss, index vs scan)
```

| Symptom                | Most likely cause              | First check                    |
| ---------------------- | ------------------------------ | ------------------------------ |
| All requests slow      | DB overloaded / pool exhausted | DB CPU, active connections     |
| Specific endpoint slow | Missing index / N+1            | `EXPLAIN ANALYZE`, query count |
| Slow after deploy      | Cold cache / new query plan    | Cache hit rate                 |
| Gets slower over time  | Memory leak / table bloat      | Memory usage, disk             |

## Key Mental Models

1. **0/1/5/50:** cached = 0 ms, hop = 1 ms, DB query = 5 ms, unnecessary network call = 50 ms. Healthy request ≈ 7 ms.
2. **DB is 77% of your time.** Optimise there first — indexes, connection pooling, batching.
3. **JWT = no network call.** Self-contained, verified locally. Session tokens require a lookup every time.
4. **Verify auth once at the edge.** Don't re-verify at every downstream service.
5. **Parallel calls = max, not sum.** Parallelise independent downstream calls — it's free latency reduction.
6. **Set timeouts on everything.** A missing timeout can cascade and take down unrelated services.
7. **Measure p50 and p99, not average.** p50 = normal, p99 = worst 1 in 100. The gap reveals what's wrong.
8. **Most problems are configuration, not code.** Missing indexes, no pooling, no caching, no timeouts — fix these before rewriting anything.
