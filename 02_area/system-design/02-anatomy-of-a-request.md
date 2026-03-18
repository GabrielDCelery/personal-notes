# Anatomy of a Request

Every system design discussion starts with a request flowing through the stack. Understanding where time goes — and where it hides — tells you what to optimise and what to leave alone. This lesson traces a single HTTP request from the browser to the database and back, layer by layer.

## The Full Journey

A user clicks "View Order" on an e-commerce site. Here's every hop, in order:

```
User's browser
  │
  ├─ 1. DNS resolution           → find the IP address
  ├─ 2. TCP handshake             → establish connection
  ├─ 3. TLS handshake             → encrypt the connection
  │
  ▼
CDN / Edge
  │
  ├─ 4. Edge check                → cached? serve it. not cached? forward.
  │
  ▼
Load balancer
  │
  ├─ 5. Route to app server       → pick a healthy backend
  │
  ▼
App server
  │
  ├─ 6. Parse request             → read headers, body, auth token
  ├─ 7. Authentication            → validate JWT or session lookup
  ├─ 8. Business logic            → run the actual code
  ├─ 9. Database query            → fetch the order
  ├─ 10. Serialise response       → build JSON
  │
  ▼
Back through load balancer → CDN → user's browser
  │
  ├─ 11. Render                   → browser paints the page
```

Not every request hits every layer. An API call between internal services skips DNS (uses service discovery), TLS (sometimes, within a mesh), and CDN. A cached response skips everything below the edge. But the full path is the worst case you need to understand.

## The Core Mental Model: 0 / 1 / 5 / 50

Before going layer by layer, here's the formula to carry in your head:

| Cost        | What it represents                                  | Examples                                                |
| ----------- | --------------------------------------------------- | ------------------------------------------------------- |
| **~0 ms**   | Cached or reused — pay once, then free              | DNS (cached), TCP (keep-alive), TLS (resumed), JWT auth |
| **~1 ms**   | Infrastructure hop — unavoidable overhead per layer | Load balancer, request parsing, serialisation           |
| **~5 ms**   | A single DB query done right                        | Indexed lookup, simple JOIN                             |
| **~50 ms**  | An unnecessary network call per request             | External auth, OAuth introspection, session DB lookup   |
| **~500 ms** | A broken data access pattern                        | Full table scan, N+1 queries, no connection pool        |

The formula: **a healthy request = 0 (reused) + a few 1s (hops) + a 5 (DB) = ~7 ms total.**

Every 50 ms chunk means you're making a network call you shouldn't be. Every 500 ms chunk means your data access is structurally broken.

## Layer by Layer

### 1–3. DNS, TCP, TLS — The Connection Cost

These three steps only matter on the **first request**. After that, they're either cached or reused and cost nothing.

- **DNS**: cold lookup costs 20–100 ms (walking root → TLD → authority). In steady state it's cached — forget it.
- **TCP handshake**: 1 round trip. Same data centre = ~1 ms. Cross-ocean = 150–300 ms. Paid once per connection.
- **TLS 1.3**: 1 additional round trip on new connections, 0 on resumed sessions.

**The rule:** keep-alive and session resumption turn all three into 0 ms. Most frameworks do this by default. Connection pooling (PgBouncer, HikariCP) does the same for database connections — opening a new Postgres connection costs 5–20 ms; a pool makes it near-zero.

If you ever see these showing up in traces on steady-state traffic, you've accidentally disabled connection reuse somewhere.

### 4. CDN / Edge

A cache hit at the edge (1–20 ms, served from a node close to the user) skips your entire backend. A cache miss just forwards to origin.

CDNs are effective for static assets and public API responses. They don't help with personalised content, writes, or WebSockets.

**The rule:** `Cache-Control: max-age=<seconds>` is the lever. Static assets with content-hashed filenames can be `max-age=31536000`. Missing headers usually mean no caching.

### 5. Load Balancer

Adds 0.5–2 ms. Largely invisible in single-service architectures.

It becomes visible in microservices: a request that passes through 3 services each behind a load balancer adds 3–6 ms just in routing overhead. This is part of the **microservice hop tax** — each service-to-service call costs ~1–3 ms (network + load balancer + serialise/deserialise), before any business logic runs.

### 6. Request Parsing and Middleware

HTTP parsing, JSON deserialisation, middleware chain — individually ~0.1–1 ms each, but they accumulate.

**Where middleware gets expensive:**

- Auth middleware calling an external service: +50 ms (see section 7)
- Rate limiting backed by Redis: +1 ms (a network round trip)
- Logging the full request body on every request: surprising CPU cost at scale

**The rule:** order middleware cheapest-first. Put auth early — don't rate-limit or log a request you're about to reject.

### 7. Authentication

This is where the biggest avoidable cost hides.

| Method                            | Cost      | Why                                     |
| --------------------------------- | --------- | --------------------------------------- |
| JWT (local validation)            | ~0.1 ms   | Self-contained — no network call needed |
| Session lookup in Redis           | ~1 ms     | One network round trip                  |
| Session lookup in DB              | ~5 ms     | Network + query                         |
| OAuth token introspection / Auth0 | 10–100 ms | HTTP call to external service           |

**Why JWT needs no network call:** a JWT encodes the user's identity and permissions directly in the token, signed with a private key. Your server verifies it by decoding the payload (base64) and checking the signature against a public key it already holds locally. That's pure CPU — no network involved. A session token is just an opaque ID that means nothing without a lookup.

The trade-off: JWTs can't be revoked until they expire. Teams handle this with short expiry windows (15 min) plus refresh tokens, rather than a blocklist (which reintroduces the network call).

**The rule for microservices:** verify auth once at the edge (API gateway), then pass a trusted internal token downstream. If each service independently calls Auth0, you pay 50 ms per service per request. Verify once, trust internally.

### 8. Business Logic

For I/O-bound services (most web apps), this is 1–5% of total request time. Rewriting from Node.js to Go to save 0.5 ms when the database takes 5 ms is not the move.

Business logic becomes the bottleneck only for CPU-heavy operations — image resizing (50–500 ms), PDF generation (100–2000 ms), ML inference (10–500 ms). These should always be async: queue the job, process in a worker, notify when done.

The other trap: **N+1 queries disguised as business logic** — a loop that makes a DB call per item. The fix is batching, not faster code.

### 9. Database Query

The bottleneck. In a healthy same-region request, the DB takes ~77% of total time. This is where optimisation effort should go.

| Query type                   | Latency       |
| ---------------------------- | ------------- |
| Primary key / indexed lookup | 1–5 ms        |
| JOIN on indexed columns      | 2–20 ms       |
| Aggregation                  | 5–500 ms      |
| Full table scan              | 100–10,000 ms |
| Transaction (3–5 statements) | 10–30 ms      |

**Where DB latency hides — the real cost formula:**

```
What you think:   1 query = 3 ms
What actually happens:
  query 1 - auth check        2 ms
  query 2 - get order         3 ms
  query 3 - get order items   4 ms
  query 4 - get shipping      2 ms
  query 5 - audit log         3 ms
  Total:                     14 ms
```

Each query pays a network round trip. 5 sequential queries at 3 ms = 15 ms. **Batch them:**

```
Bad:  5 queries × 3 ms = 15 ms  (sequential round trips)
Good: 1 JOIN query     =  5 ms  (one round trip, DB does the work)
Good: 1 IN clause      =  4 ms  (one round trip, multiple rows)
```

**The rule:** missing index = full table scan = 100–10,000 ms. Run `EXPLAIN ANALYZE` on any slow query. The query planner tells you exactly why it's slow.

### 10–11. Serialisation and Network Return

Serialisation is negligible for normal payloads (< 100 KB). It becomes visible at 1 MB+ — a 10 MB JSON response costs 50–500 ms to serialise and dominates the response time.

**The rule:** paginate. Return 50 rows, not 10,000. Gzip/brotli compression is almost always worth it — a 100 KB JSON response compresses to ~15 KB at a cost of ~0.1 ms.

## Putting It All Together — Latency Budget

A typical API request within the same data centre, everything done right:

| Layer                        | Latency   | % of total |
| ---------------------------- | --------- | ---------- |
| DNS, TCP, TLS (reused)       | 0 ms      | 0%         |
| Load balancer                | 0.5 ms    | 4%         |
| Request parsing + middleware | 0.2 ms    | 3%         |
| Auth (JWT)                   | 0.1 ms    | 1%         |
| Business logic               | 0.5 ms    | 7%         |
| **Database query**           | **5 ms**  | **77%**    |
| Serialisation + network      | 0.7 ms    | 10%        |
| **Total**                    | **~7 ms** |            |

The same request with common mistakes:

| Layer         | Good     | Bad        | What went wrong                        |
| ------------- | -------- | ---------- | -------------------------------------- |
| DNS           | 0 ms     | 50 ms      | TTL set to 0                           |
| TCP           | 0 ms     | 5 ms       | No keep-alive                          |
| Auth          | 0.1 ms   | 50 ms      | External OAuth per request             |
| Database      | 5 ms     | 500 ms     | Missing index, N+1, no pool            |
| Serialisation | 0.2 ms   | 50 ms      | 10,000 rows un-paginated               |
| **Total**     | **7 ms** | **666 ms** | **95× slower, no code changes needed** |

Every "bad" entry is something seen regularly in production. The fix is never "use a faster language."

## Multi-Service Requests

When a request fans out to multiple services, the latency formula changes:

```
sequential total = sum of all calls
parallel total   = max of all calls
```

```
API Gateway
  ├─→ User Service         3 ms
  ├─→ Order Service        5 ms
  │     └─→ Inventory      4 ms   (depends on Order, must be sequential)
  └─→ Recommendations     15 ms

Sequential: 3 + 5 + 4 + 15 = 27 ms
Parallel:   max(3, 5+4, 15) = 15 ms
```

Always parallelise independent calls. This is free latency reduction.

**The microservice hop tax:** each service-to-service call costs ~1–3 ms (network + load balancer + serialise/deserialise) before any logic runs. 5 sequential hops = 10 ms overhead before your code does anything. A monolith making 5 function calls pays ~0.001 ms for the same thing. This doesn't make microservices wrong — but prefer wide and shallow call chains over deep and narrow ones.

## Timeouts and Cascading Failure

Without timeouts, a slow downstream service stalls your entire system:

```
Order Service hangs (30s timeout):
  100 users hit the endpoint
  100 threads blocked waiting
  Thread pool exhausted
  API can't serve ANY requests — including unrelated ones
  Everything is down
```

**The rule:** set timeouts on every outbound call — HTTP clients, DB connections, Redis. Set them just above the p99 latency of the downstream service, not at 30 seconds.

The cascade stops with a **circuit breaker**: after N failures, stop calling the service for a cooldown period and return a fallback instead of hanging.

## How to Read a Latency Profile

**Use percentiles, not averages.** A service with "average 10 ms" might have p99 at 2 seconds. Averages hide the tail.

The gap between p50 and p99 tells you what's wrong:

| Gap            | What it means                                                                                    |
| -------------- | ------------------------------------------------------------------------------------------------ |
| p99 ≈ 2× p50   | Healthy. Minor variance from GC pauses, cache misses.                                            |
| p99 ≈ 10× p50  | Something occasionally goes wrong — lock contention, cold cache, rare slow query.                |
| p99 ≈ 100× p50 | Bimodal — some requests hit a completely different code path (cache miss vs hit, index vs scan). |

| Symptom                | Most likely cause                 | First check                    |
| ---------------------- | --------------------------------- | ------------------------------ |
| All requests slow      | DB overloaded, pool exhausted     | DB CPU, active connections     |
| Specific endpoint slow | Missing index, N+1, large payload | `EXPLAIN ANALYZE`, query count |
| Slow after deploy      | Cold cache, new query plan        | Cache hit rate                 |
| Gets slower over time  | Memory leak, table bloat          | Memory usage, disk             |

## Key Takeaways

- **0/1/5/50:** cached = 0 ms, infrastructure hop = 1 ms, DB query = 5 ms, unnecessary network call = 50 ms. A healthy request is ~7 ms.
- **The DB is 77% of your time.** Optimise there first — indexes, connection pooling, batching.
- **JWT = no network call.** The token is self-contained and verified locally. Session tokens require a lookup every time.
- **Verify auth once at the edge.** Don't re-verify at every microservice.
- **Parallel calls = max, not sum.** Parallelise independent downstream calls.
- **Timeouts on everything.** A missing timeout on one call can cascade and take down unrelated services.
- **Measure p99, not average.** Averages hide the tail where users suffer.
- **Most problems are configuration, not code.** Missing indexes, no pooling, no caching, no timeouts — fix these before rewriting anything.

## What's Next

Now that you can trace where latency goes in a request, the next lesson dives into the component that dominates most of that time — databases. How Postgres, MySQL, MongoDB, and DynamoDB actually handle queries, their throughput profiles, and when to reach for read replicas, caching, or sharding.
