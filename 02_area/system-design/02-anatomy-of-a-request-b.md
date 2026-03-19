# Anatomy of a Request

Knowing where time goes in a request tells you what to optimise — and what to leave alone.

## The Full Journey

```
Browser → DNS → TCP → TLS → CDN/Edge → Load Balancer → App Server → DB → back
```

Not every request hits every layer. A CDN cache hit skips your entire backend. An internal service call skips DNS and CDN. But the full path is the worst case to understand.

## The Core Mental Model: 0 / 1 / 5 / 50

| Cost    | What it means                        | Examples                                         |
| ------- | ------------------------------------ | ------------------------------------------------ |
| ~0 ms   | Cached or reused — pay once, free    | DNS (cached), TCP (keep-alive), JWT auth         |
| ~1 ms   | Infrastructure hop                   | Load balancer, request parsing, serialisation    |
| ~5 ms   | Single DB query done right           | Indexed lookup, simple JOIN                      |
| ~50 ms  | Unnecessary network call per request | External auth, OAuth introspection, session DB   |
| ~500 ms | Broken data access pattern           | Full table scan, N+1 queries, no connection pool |

**A healthy request = ~7 ms.** A few infrastructure hops (~1 ms each) + one DB query (~5 ms).

Every 50 ms chunk = a network call you shouldn't be making. Every 500 ms chunk = structurally broken data access.

## Layer by Layer

### DNS / TCP / TLS

Only paid on the **first request**. After that they're cached or reused — cost is 0 ms.

- DNS cold lookup: 20-100 ms. Cached: free.
- TCP handshake: 1 round trip. Same DC = ~1 ms. Cross-ocean = 150-300 ms.
- TLS 1.3: 1 extra round trip on new connections, 0 on resumed sessions.

Keep-alive and session resumption turn all three into 0 ms. Most frameworks do this by default. If you see these in steady-state traces, you've disabled connection reuse somewhere.

### CDN / Edge

A cache hit (served from an edge node near the user) skips your entire backend. A miss just forwards.

Effective for static assets and public responses. Doesn't help with personalised content, writes, or WebSockets. The lever is `Cache-Control: max-age=<seconds>`.

### Load Balancer

Adds 0.5-2 ms. Invisible in single-service architectures.

In microservices: each service-to-service hop costs ~1-3 ms (network + LB + serialise/deserialise) before any business logic runs. Prefer wide and shallow call chains over deep and narrow ones.

### Authentication

The biggest avoidable cost.

| Method                  | Cost      | Why                                                    |
| ----------------------- | --------- | ------------------------------------------------------ |
| JWT (local validation)  | ~0.1 ms   | Self-contained — decode + verify signature, no network |
| Session lookup in Redis | ~1 ms     | One network round trip                                 |
| Session lookup in DB    | ~5 ms     | Network + query                                        |
| External OAuth / Auth0  | 10-100 ms | HTTP call to external service                          |

```
               0.1 ms    1 ms      10 ms     100 ms
               │         │         │         │
JWT            █·········│·········│·········│
Session/Redis  ·         █·········│·········│
Session/DB     ·         ·    █····│·········│
OAuth/Auth0    ·         ·         │    ├────█
               │         │         │         │
               0.1 ms    1 ms      10 ms     100 ms
```

JWT encodes identity and permissions directly in the token. Your server verifies it locally against a public key it already holds — no network call. A session token is an opaque ID that requires a lookup every time.

JWT trade-off: can't be revoked until expiry. Handle with short expiry (15 min) + refresh tokens.

**In microservices: verify auth once at the edge, pass a trusted internal token downstream.** If each service independently calls Auth0, you pay 50 ms per service per request.

### Business Logic

For I/O-bound services (most web apps), this is 1-5% of total request time. Optimising language choice here is irrelevant when the DB takes 5 ms.

Business logic becomes the bottleneck only for CPU-heavy work: image resizing, PDF generation, ML inference. These should be async — queue the job, process in a worker, notify when done.

Common trap: **N+1 queries disguised as business logic** — a loop that makes a DB call per item. Fix with batching, not faster code.

### Database Query

The bottleneck. In a healthy request, the DB takes ~77% of total time.

| Query type              | Latency       |
| ----------------------- | ------------- |
| Indexed lookup          | 1-5 ms        |
| JOIN on indexed columns | 2-20 ms       |
| Aggregation             | 5-500 ms      |
| Full table scan         | 100-10,000 ms |
| Transaction (3-5 stmts) | 10-30 ms      |

```
               1 ms      10 ms     100 ms    10,000 ms
               │         │         │         │
Indexed        ├──█──┤   ·         ·         ·
JOIN           ├──────────────█──┤ ·         ·
Aggregation    ·         ·   ├─────────────┤ ·   (varies)
Full scan      ·         ·         ·  ├───────────────────►
Transaction    ·      ├──█──┤      ·         ·
               │         │         │         │
               1 ms      10 ms     100 ms    10,000 ms
```

**N+1 is the silent killer:** 5 sequential queries × 3 ms = 15 ms. One JOIN = 5 ms. Each query pays a network round trip — batch them.

Run `EXPLAIN ANALYZE` on any slow query. A missing index is the most common cause.

### Serialisation and Response

Negligible for normal payloads (< 100 KB). At 1 MB+ it starts to matter — a 10 MB JSON response can take 50-500 ms to serialise. Paginate: return 50 rows, not 10,000. Gzip compression is almost always worth it (100 KB → 15 KB at ~0.1 ms cost).

## Latency Budget

Everything done right, same data centre:

| Layer                   | Latency   | % of total |
| ----------------------- | --------- | ---------- |
| DNS, TCP, TLS (reused)  | 0 ms      | 0%         |
| Load balancer           | 0.5 ms    | 4%         |
| Parsing + middleware    | 0.2 ms    | 3%         |
| Auth (JWT)              | 0.1 ms    | 1%         |
| Business logic          | 0.5 ms    | 7%         |
| **Database query**      | **5 ms**  | **77%**    |
| Serialisation + network | 0.7 ms    | 10%        |
| **Total**               | **~7 ms** |            |

```
                         0 ms                              7 ms
                         ├──────────────────────────────────┤
DNS/TCP/TLS (reused)     ·                                  · (0 ms)
Load balancer            ██·                                · (0.5 ms)
Parsing + middleware     █··                                · (0.2 ms)
Auth (JWT)               ·                                  · (0.1 ms)
Business logic           ██·                                · (0.5 ms)
Database query           █████████████████████████████████  · (5 ms — 77%)
Serialisation + network  ████·                              · (0.7 ms)
                         ├──────────────────────────────────┤
                         0 ms                              7 ms
```

The same request with common mistakes:

| Layer         | Good     | Bad         | What went wrong             |
| ------------- | -------- | ----------- | --------------------------- |
| Auth          | 0.1 ms   | 50 ms       | External OAuth per request  |
| Database      | 5 ms     | 500 ms      | Missing index, N+1, no pool |
| Serialisation | 0.2 ms   | 50 ms       | 10,000 rows un-paginated    |
| **Total**     | **7 ms** | **~600 ms** | **Fix is config, not code** |

```
                         0 ms                    300 ms              600 ms
                         ├────────────────────────┼───────────────────┤
GOOD REQUEST (~7 ms)
  Auth (JWT)             · (0.1 ms)
  Database               ███ (5 ms)
  Serialisation          · (0.2 ms)
  Total                  █ (7 ms)

BAD REQUEST (~600 ms)
  Auth (OAuth)           ████ (50 ms)
  Database               ██████████████████████████████████████ (500 ms)
  Serialisation          ████ (50 ms)
  Total                  ████████████████████████████████████████████
                         ├────────────────────────┼───────────────────┤
                         0 ms                    300 ms              600 ms
```

## Multi-Service Requests

```
sequential total = sum of all calls
parallel total   = max of all calls
```

```
API Gateway
  ├─→ User Service       3 ms
  ├─→ Order Service      5 ms
  │     └─→ Inventory    4 ms  (sequential, depends on Order)
  └─→ Recommendations   15 ms

Sequential: 3 + 5 + 4 + 15 = 27 ms
Parallel:   max(3, 9, 15)   = 15 ms
```

Always parallelise independent calls. This is free latency reduction.

## Timeouts and Cascading Failure

Without a timeout, a slow downstream service stalls your entire system — threads pile up waiting, pool exhausts, unrelated endpoints go down.

Set timeouts on every outbound call (HTTP clients, DB connections, Redis). Set them just above p99 latency of the downstream service, not at 30 seconds.

**Circuit breaker:** after N failures, stop calling the service for a cooldown period and return a fallback. Prevents a degraded service from taking everything down with it.

## Reading Latency Profiles

**Use p99, not averages.** "Average 10 ms" can hide a p99 of 2 seconds. The gap between p50 and p99 tells you what's wrong:

| Gap            | What it means                                                        |
| -------------- | -------------------------------------------------------------------- |
| p99 ≈ 2× p50   | Healthy. Minor variance from GC pauses, cache misses.                |
| p99 ≈ 10× p50  | Something occasionally goes wrong — lock contention, rare slow query |
| p99 ≈ 100× p50 | Bimodal — some requests hit a completely different code path         |

| Symptom               | Most likely cause             | First check                    |
| --------------------- | ----------------------------- | ------------------------------ |
| All requests slow     | DB overloaded, pool exhausted | DB CPU, active connections     |
| One endpoint slow     | Missing index, N+1            | `EXPLAIN ANALYZE`, query count |
| Slow after deploy     | Cold cache, new query plan    | Cache hit rate                 |
| Gets slower over time | Memory leak, table bloat      | Memory usage, disk             |

## Key Mental Models

1. **0/1/5/50:** cached = 0 ms, hop = 1 ms, DB query = 5 ms, unnecessary network call = 50 ms. Target ~7 ms total.
2. **DB is 77% of your time.** Optimise there first — indexes, pooling, batching.
3. **JWT = no network call.** Self-contained and verified locally. Session tokens need a lookup every time.
4. **Verify auth once at the edge.** Don't re-verify at every microservice.
5. **Parallel calls = max, not sum.** Parallelise independent downstream calls.
6. **Timeouts on everything.** One missing timeout can cascade and take down unrelated services.
7. **Measure p99, not average.** Averages hide the tail.
8. **Most problems are configuration, not code.** Missing indexes, no pooling, no caching, no timeouts — fix these before rewriting anything.
