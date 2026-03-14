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

## Layer by Layer

### 1. DNS Resolution

The browser needs an IP address before it can connect.

| Scenario                  | Latency   | What happens                                   |
| ------------------------- | --------- | ---------------------------------------------- |
| Cached in browser         | 0 ms      | Browser already knows the IP                   |
| Cached in OS              | 0-1 ms    | OS resolver returns from local cache           |
| Cached at ISP resolver    | 1-5 ms    | ISP's DNS server has it                        |
| Full recursive resolution | 20-100 ms | Walk the DNS hierarchy: root → TLD → authority |

DNS is a one-time cost per domain per TTL period. After the first lookup, it's cached everywhere. This is why DNS almost never matters in steady-state performance — but it dominates cold-start latency (first request after deploy, new domain, short TTL).

**What you control:**

- Set reasonable TTLs — 60-300 seconds is typical. Too short (5s) causes constant re-resolution. Too long (86400s) slows failover.
- Use DNS-based load balancing (Route 53, Cloudflare) for geographic routing — adds 0 ms because the resolution itself is the routing decision.

### 2. TCP Handshake

Before any data flows, client and server must establish a TCP connection.

```
Client                          Server
  │                                │
  │──── SYN ──────────────────────→│   1 round trip
  │←─── SYN-ACK ──────────────────│
  │──── ACK ──────────────────────→│
  │                                │
  Connection established. Data can flow.
```

| Network                    | Handshake cost | Notes                     |
| -------------------------- | -------------- | ------------------------- |
| Same data centre           | 0.5-1 ms       | 1 round trip × 0.5 ms     |
| Same region (e.g. eu-west) | 2-5 ms         | 1 round trip × 2-5 ms     |
| Cross-region               | 50-100 ms      | 1 round trip × 50-100 ms  |
| Cross-ocean                | 150-300 ms     | 1 round trip × 150-300 ms |

This cost is paid once per connection. After that, requests on the same connection skip it entirely. This is why **connection reuse matters enormously**:

```
Without keep-alive (HTTP/1.0 default):
  Request 1: TCP handshake (1 ms) + request (5 ms) = 6 ms
  Request 2: TCP handshake (1 ms) + request (5 ms) = 6 ms
  Request 3: TCP handshake (1 ms) + request (5 ms) = 6 ms
  Total: 18 ms

With keep-alive (HTTP/1.1+ default):
  Request 1: TCP handshake (1 ms) + request (5 ms) = 6 ms
  Request 2: request (5 ms)                         = 5 ms
  Request 3: request (5 ms)                         = 5 ms
  Total: 16 ms

With connection pooling (app server → database):
  Same idea. Open 10-50 connections at startup. Every query reuses one.
  Zero handshake cost per query.
```

**What you control:**

- Enable keep-alive everywhere (HTTP, database connections, Redis connections). Most frameworks do this by default — make sure you haven't disabled it.
- Use connection pooling for databases (PgBouncer, HikariCP, built-in pool). Opening a new Postgres connection costs 5-20 ms (TCP + auth + process fork). A pool amortises this to near-zero.

### 3. TLS Handshake

If the connection is HTTPS (it should be), encryption adds more round trips on top of TCP.

```
TLS 1.2: 2 additional round trips
  Client → Server: ClientHello (supported ciphers)
  Server → Client: ServerHello + certificate
  Client → Server: Key exchange
  Server → Client: Finished

TLS 1.3: 1 additional round trip
  Client → Server: ClientHello + key share
  Server → Client: ServerHello + certificate + finished
  Done. Data flows immediately.

TLS 1.3 with 0-RTT (resumption):
  Client → Server: ClientHello + early data
  0 additional round trips for resumed connections.
```

| Scenario                   | Latency    | Notes                                  |
| -------------------------- | ---------- | -------------------------------------- |
| TLS 1.3, same data centre  | 0.5-1 ms   | 1 round trip                           |
| TLS 1.3, cross-region      | 50-100 ms  | 1 round trip, but the trip is long     |
| TLS 1.2, cross-region      | 100-200 ms | 2 round trips over a long link         |
| TLS 1.3 resumption (0-RTT) | 0 ms       | Resumed session, data sent immediately |

TLS 1.3 is standard everywhere now. The practical cost is 1 round trip for new connections, 0 for resumed ones.

**What you control:**

- Ensure TLS 1.3 is enabled (it's default on modern servers, but legacy configs may force 1.2).
- Terminate TLS at the load balancer or CDN edge — your app servers talk plain HTTP internally, avoiding repeated handshakes between services.
- For internal service-to-service calls within the same data centre, TLS adds ~1 ms. Some teams skip it inside a VPC/mesh — security trade-off.

### 4. CDN / Edge Layer

A CDN (CloudFront, Cloudflare, Fastly) sits between users and your servers. It caches responses at edge locations close to users.

| Scenario           | Latency      | What happens                                           |
| ------------------ | ------------ | ------------------------------------------------------ |
| Cache hit at edge  | 1-20 ms      | Response served from nearest edge, origin not hit      |
| Cache miss at edge | + origin ms  | Edge forwards to origin, caches response for next time |
| Dynamic content    | Pass-through | Not cached, forwarded directly to origin               |

CDNs are most effective for:

- Static assets (JS, CSS, images) — cache hit rate of 90-99%
- Public API responses that don't change per-user — cache for 5-60 seconds
- Media files — images, video, PDFs

CDNs don't help with:

- Personalised content (your cart, your feed) — each user sees different data
- Write operations (POST/PUT/DELETE) — pass straight through
- WebSocket connections — persistent, not request/response

**What you control:**

- Set `Cache-Control` headers correctly. `max-age=3600` caches for 1 hour. `no-store` never caches. Missing headers usually mean no caching.
- Static assets should be cached aggressively (`max-age=31536000`) with a content hash in the filename for cache busting.
- Consider edge compute (Cloudflare Workers, Lambda@Edge) for light transforms at the edge — auth token validation, A/B test routing, header rewriting — before the request ever reaches your origin.

### 5. Load Balancer

The load balancer picks which app server handles the request. In cloud environments, this is usually an ALB (AWS), Cloud Load Balancer (GCP), or an Nginx/HAProxy instance.

| Component                    | Added latency | Notes                                   |
| ---------------------------- | ------------- | --------------------------------------- |
| Cloud load balancer (ALB)    | 0.5-2 ms      | TCP termination + routing decision      |
| Nginx/HAProxy (self-hosted)  | 0.1-0.5 ms    | Lighter weight, fewer features          |
| Service mesh sidecar (Envoy) | 0.5-2 ms      | Per-hop overhead in a mesh architecture |

Load balancers add minimal latency. Their cost only matters when you have many hops — a request that traverses 3 services, each behind a load balancer, adds 3-6 ms just in routing.

**Routing algorithms:**

| Algorithm         | How it works                          | Good for                                |
| ----------------- | ------------------------------------- | --------------------------------------- |
| Round robin       | Next server in the list               | Stateless services, equal instances     |
| Least connections | Server with fewest active connections | Varying request complexity              |
| IP hash           | Same client IP → same server          | Session affinity (but prefer stateless) |
| Weighted          | Heavier servers get more traffic      | Mixed instance sizes, canary deploys    |

**What you control:**

- Use health checks — a load balancer that sends traffic to a dead server turns a partial outage into user-facing errors.
- Keep services stateless so any instance can handle any request. Session data goes in Redis or a cookie, not in server memory.
- For internal services, consider client-side load balancing (gRPC, service mesh) — one fewer network hop.

### 6. Request Parsing and Middleware

The request arrives at your app server. Before your business logic runs, the framework processes the request through a middleware stack.

| Step                 | Latency     | Notes                                 |
| -------------------- | ----------- | ------------------------------------- |
| HTTP parsing         | 0.01-0.1 ms | Read method, path, headers, body      |
| Body deserialisation | 0.05-1 ms   | JSON parse, depends on payload size   |
| Middleware chain     | 0.1-5 ms    | Logging, CORS, rate limiting, tracing |
| Request validation   | 0.05-0.5 ms | Schema validation, input sanitisation |

Individually these are tiny. But middleware chains accumulate. 10 middleware layers at 0.2 ms each = 2 ms before your code even runs. This usually doesn't matter, but it's worth knowing where time goes if you're profiling.

**Where middleware gets expensive:**

- Logging middleware that serialises the full request body to JSON on every request
- Rate limiting backed by Redis — adds a network round trip (~1 ms)
- Distributed tracing that propagates context and sends spans — adds 0.5-2 ms
- Auth middleware that calls an external service (OAuth introspection) — adds 5-50 ms

**What you control:**

- Order middleware from cheapest to most expensive. Put authentication early — no point rate-limiting or logging a request you're about to reject.
- Make rate limiting local when possible (in-memory token bucket) rather than centralised (Redis). Use centralised only when you need global accuracy across multiple servers.

### 7. Authentication and Authorisation

Verifying who the user is and what they're allowed to do.

| Method                        | Latency     | Notes                                         |
| ----------------------------- | ----------- | --------------------------------------------- |
| JWT validation (local)        | 0.01-0.1 ms | Signature check, no network call              |
| JWT + JWKS fetch (cached)     | 0.01-0.1 ms | Key is cached locally, refreshed periodically |
| Session lookup in Redis       | 0.5-2 ms    | Network round trip to Redis                   |
| Session lookup in database    | 1-5 ms      | Network round trip + query                    |
| OAuth token introspection     | 5-50 ms     | HTTP call to auth server                      |
| External auth service (Auth0) | 10-100 ms   | HTTP call to third party                      |

JWTs are fast because they're self-contained — all the information is in the token, verified with a signature. No network call. The trade-off is revocation: you can't invalidate a JWT until it expires unless you check a blocklist (which brings back the network call).

**What you control:**

- Prefer JWTs for stateless services — zero network overhead for auth.
- Cache session/token lookups aggressively. User permissions rarely change mid-request.
- If you must call an external auth service, do it once at the edge (API gateway) and pass a trusted internal token downstream. Don't re-verify auth at every microservice.

### 8. Business Logic

Your actual code. The part you have the most control over and usually the part that matters least for latency.

| Operation             | Latency       | Notes                              |
| --------------------- | ------------- | ---------------------------------- |
| Simple computation    | 0.001-0.1 ms  | Transform data, apply rules        |
| In-memory data lookup | 0.001-0.01 ms | HashMap, cached config             |
| Regex on short string | 0.001-0.01 ms | Compiled regex                     |
| Template rendering    | 0.1-1 ms      | Server-side HTML                   |
| Image resize          | 50-500 ms     | CPU-heavy, consider async          |
| PDF generation        | 100-2000 ms   | CPU-heavy, definitely async        |
| ML model inference    | 10-500 ms     | Depends on model size and hardware |

For I/O-bound services (most web applications), business logic is 1-5% of total request time. The database query dominates. Rewriting your app from Node.js to Go to save 0.5 ms when the database takes 5 ms is usually not worth it.

**When business logic does matter:**

- CPU-heavy operations: image processing, PDF generation, video transcoding, ML inference. These should almost always be async — put a message on a queue, process in a worker, notify when done.
- N+1 query patterns disguised as business logic — a loop that makes a database call per iteration. The fix is usually batching, not faster code.
- Large payload serialisation/deserialisation — transforming 10 MB of JSON takes 50 ms. This is often hidden inside "business logic" time.

### 9. Database Query

Almost always the bottleneck. Where most optimisation effort should go.

| Query type                    | Latency      | Notes                                          |
| ----------------------------- | ------------ | ---------------------------------------------- |
| Primary key lookup (indexed)  | 1-3 ms       | Best case, single row, index in memory         |
| Simple WHERE with index       | 1-5 ms       | Index seek + row fetch                         |
| JOIN on indexed columns       | 2-20 ms      | Depends on result set size                     |
| Complex JOIN, multiple tables | 10-100 ms    | Grows with data size, join strategy            |
| Aggregation (COUNT, SUM, AVG) | 5-500 ms     | Depends on whether index-only scan is possible |
| Full table scan               | 100-10000 ms | Missing index, or genuinely need all rows      |
| Write + index update          | 2-10 ms      | WAL + update every index on the table          |
| Write with FK constraints     | 5-20 ms      | Must validate foreign keys, possible lock wait |
| Transaction (3-5 statements)  | 10-30 ms     | Sum of statements + lock acquisition           |

### Where database latency hides

```
What you think happens:            What actually happens:

  App → DB: 1 query, 3 ms           App → DB: get connection from pool (0.1 ms)
                                     App → DB: query 1 - auth check (2 ms)
                                     App → DB: query 2 - get order (3 ms)
                                     App → DB: query 3 - get order items (4 ms)
                                     App → DB: query 4 - get shipping status (2 ms)
                                     App → DB: query 5 - audit log insert (3 ms)
                                     Total: 14.1 ms (not 3 ms)
```

Each query pays network round trip + execution. 5 sequential queries at 3 ms each = 15 ms. This is why:

**Batching matters:**

```
Bad:  5 queries × 3 ms each = 15 ms (sequential round trips)
Good: 1 query with JOIN    = 5 ms (one round trip, DB does the work)
Good: 1 batch query (IN clause) = 4 ms (one round trip, multiple rows)
```

**Connection pooling matters:**

| Without pool                          | With pool                   |
| ------------------------------------- | --------------------------- |
| Open connection: TCP + auth = 5-20 ms | Grab from pool: 0.05-0.1 ms |
| Run query: 3 ms                       | Run query: 3 ms             |
| Close connection                      | Return to pool              |
| Total: 8-23 ms                        | Total: 3.1 ms               |

**Indexes matter:**

| Without index           | With index             | Difference       |
| ----------------------- | ---------------------- | ---------------- |
| Full table scan: 500 ms | Index lookup: 2 ms     | 250x faster      |
| Scales with table size  | Scales with tree depth | O(n) vs O(log n) |

**What you control:**

- Use EXPLAIN ANALYZE on slow queries. The query planner tells you exactly why it's slow.
- Add indexes on columns you filter, join, or sort by. But not on everything — each index slows writes.
- Use connection pooling always. PgBouncer for Postgres, HikariCP for Java, built-in pools in most ORMs.
- Batch reads with IN clauses or JOINs instead of sequential queries in a loop.
- Consider read replicas when read QPS exceeds what a single node can handle (5,000-15,000 QPS).

### 10. Response Serialisation

Turning your data into JSON (or protobuf, or HTML) to send back.

| Payload size | JSON serialisation | Protobuf serialisation | Notes                           |
| ------------ | ------------------ | ---------------------- | ------------------------------- |
| 1 KB         | 0.01-0.05 ms       | 0.005-0.02 ms          | Trivial either way              |
| 10 KB        | 0.05-0.5 ms        | 0.02-0.1 ms            | Still negligible                |
| 100 KB       | 0.5-5 ms           | 0.1-1 ms               | Starting to show up in traces   |
| 1 MB         | 5-50 ms            | 1-10 ms                | Dominates if you're not careful |
| 10 MB        | 50-500 ms          | 10-100 ms              | Paginate or stream instead      |

Protobuf is 2-10x faster than JSON and produces smaller payloads (30-50% less bytes on wire). The trade-off is readability — JSON is human-readable, protobuf isn't. Use protobuf for internal service-to-service communication where performance matters. Use JSON for external APIs where developer experience matters.

**What you control:**

- Paginate large responses. Return 50 items per page, not 10,000.
- Use field selection (GraphQL, sparse fieldsets) — don't send 50 fields when the client needs 5.
- Compress responses with gzip/brotli. A 100 KB JSON response compresses to ~15 KB. The compression costs 0.1 ms but saves network transfer time.
- For internal services at high throughput, consider protobuf or gRPC.

### 11. Network Transfer and Rendering

The response travels back through load balancer → CDN → internet → browser.

| Phase                           | Latency     | Notes                                 |
| ------------------------------- | ----------- | ------------------------------------- |
| Response on wire (same DC)      | 0.2-0.5 ms  | Speed of light + routing              |
| Response on wire (cross-region) | 50-100 ms   | Physics, not software                 |
| Browser HTML parse              | 1-10 ms     | Depends on document size              |
| Browser render (simple page)    | 10-50 ms    | Layout + paint                        |
| Browser render (complex SPA)    | 50-500 ms   | JS execution + virtual DOM + paint    |
| Time to interactive (heavy SPA) | 500-3000 ms | Download JS bundles + parse + execute |

The network transfer is physics — you can't speed up light. But you can reduce what you send (compression, pagination) and where you send it from (CDN edge nodes close to users).

## Putting It All Together — Latency Budget

A typical API request within the same data centre:

| Layer                  | Latency  | Cumulative | % of total |
| ---------------------- | -------- | ---------- | ---------- |
| DNS (cached)           | 0 ms     | 0 ms       | 0%         |
| TCP (keep-alive)       | 0 ms     | 0 ms       | 0%         |
| TLS (resumed)          | 0 ms     | 0 ms       | 0%         |
| Load balancer          | 0.5 ms   | 0.5 ms     | 4%         |
| Request parsing        | 0.2 ms   | 0.7 ms     | 5%         |
| Auth (JWT)             | 0.1 ms   | 0.8 ms     | 6%         |
| Business logic         | 0.5 ms   | 1.3 ms     | 10%        |
| **Database query**     | **5 ms** | **6.3 ms** | **77%**    |
| Response serialisation | 0.2 ms   | 6.5 ms     | 2%         |
| Network back           | 0.5 ms   | 7 ms       | 4%         |
| **Total**              |          | **~7 ms**  |            |

The database is 77% of the time. In a cross-region request where network adds 100+ ms, the network dominates instead. Either way, your app code is rarely the bottleneck.

### The same request, but with common mistakes

| Layer                  | Good     | Bad        | What went wrong                           |
| ---------------------- | -------- | ---------- | ----------------------------------------- |
| DNS                    | 0 ms     | 50 ms      | TTL set to 0, re-resolves every time      |
| TCP                    | 0 ms     | 5 ms       | No keep-alive, new connection per request |
| TLS                    | 0 ms     | 5 ms       | No session resumption, TLS 1.2            |
| Load balancer          | 0.5 ms   | 0.5 ms     | Same either way                           |
| Request parsing        | 0.2 ms   | 5 ms       | Logging full 1 MB request body            |
| Auth                   | 0.1 ms   | 50 ms      | Calling external OAuth on every request   |
| Business logic         | 0.5 ms   | 0.5 ms     | Same either way                           |
| Database               | 5 ms     | 500 ms     | Missing index, N+1 queries, no pool       |
| Response serialisation | 0.2 ms   | 50 ms      | Returning 10,000 rows un-paginated        |
| Network back           | 0.5 ms   | 0.5 ms     | Same either way                           |
| **Total**              | **7 ms** | **666 ms** | **95x slower from avoidable mistakes**    |

Every "bad" entry is something seen in production regularly. The fix is never "use a faster language" — it's fix the query, add connection pooling, cache the auth check, paginate the response.

## Multi-Service Requests

Modern architectures have multiple services. A single user request might fan out:

```
API Gateway
  │
  ├─→ User Service (get user profile)          3 ms
  ├─→ Order Service (get order details)        5 ms
  │     └─→ Inventory Service (check stock)    4 ms
  └─→ Recommendation Service (get suggestions) 15 ms

Sequential: 3 + 5 + 4 + 15 = 27 ms
Parallel:   max(3, 5+4, 15) = 15 ms
```

### Sequential vs parallel calls

| Pattern    | Total latency                       | When to use                                 |
| ---------- | ----------------------------------- | ------------------------------------------- |
| Sequential | Sum of all calls                    | When call B depends on the result of call A |
| Parallel   | Max of all calls                    | When calls are independent                  |
| Mixed      | Sum of sequential + max of parallel | Most real scenarios                         |

If you're calling 3 independent services at 5 ms each:

- Sequential: 15 ms
- Parallel: 5 ms

Always parallelise independent calls. This is free latency reduction.

### Latency amplification

With microservices, a single user request might touch 5-10 services. Each service adds:

| Per-hop cost          | Latency    |
| --------------------- | ---------- |
| Network round trip    | 0.5-1 ms   |
| Load balancer         | 0.5-1 ms   |
| Serialise/deserialise | 0.1-1 ms   |
| **Total per hop**     | **1-3 ms** |

5 sequential hops × 2 ms overhead each = 10 ms before any business logic runs. This is the hidden tax of microservices. A monolith making 5 function calls pays ~0.001 ms for the same thing.

This doesn't mean microservices are bad — but it means you need to:

- Minimise the depth of call chains (prefer wide and shallow over deep and narrow)
- Parallelise independent calls
- Consider whether two services that always call each other should be one service
- Use async communication (queues) where you don't need a synchronous response

## Timeouts and Failure

Every network call can fail or hang. Without timeouts, a slow downstream service stalls your entire request.

| Timeout type       | Typical value | Purpose                                       |
| ------------------ | ------------- | --------------------------------------------- |
| Connection timeout | 1-3 seconds   | Give up if TCP handshake takes too long       |
| Read timeout       | 5-30 seconds  | Give up if server accepts but doesn't respond |
| Total timeout      | 10-60 seconds | Hard cap on the entire operation              |

### How slow services cascade

```
Normal:
  User → API (50 ms) → Order Service (5 ms) → DB (3 ms)
  Total: 58 ms ✓

Order Service is slow:
  User → API ──→ Order Service (hanging... 30 seconds... timeout)
  API's thread is blocked for 30 seconds.
  100 users hit the same endpoint.
  100 threads blocked.
  Thread pool exhausted.
  API can't serve ANY requests — even ones that don't touch Order Service.
  Everything is down.
```

This is a **cascading failure**. One slow service takes down everything upstream. Defence:

| Defence            | How it works                                                                  |
| ------------------ | ----------------------------------------------------------------------------- |
| Timeouts           | Don't wait forever. Fail fast.                                                |
| Circuit breaker    | After N failures, stop calling the service for a cooldown period.             |
| Bulkhead           | Isolate thread pools per downstream service. Slow DB can't starve API calls.  |
| Retry with backoff | Retry failed calls with exponential delay. Don't hammer a struggling service. |
| Fallback           | Return cached/default data when a dependency is down.                         |

**What you control:**

- Set timeouts on every outbound call. Every HTTP client, every database connection, every Redis call. No exceptions.
- Set the timeout to slightly above the p99 latency of the downstream service. If Order Service is normally 5 ms with p99 at 50 ms, set the timeout to 100-200 ms — not 30 seconds.
- Use circuit breakers for external dependencies. If a service fails 50% of the time, stop calling it for 30 seconds and return a fallback.

## How to Read a Latency Profile

When debugging slow requests, you need to know where the time goes. Here's what to look for:

### p50, p95, p99 — percentile latencies

| Metric | What it tells you                        | Example |
| ------ | ---------------------------------------- | ------- |
| p50    | The median — half of requests are faster | 5 ms    |
| p95    | 1 in 20 requests is slower than this     | 25 ms   |
| p99    | 1 in 100 requests is slower than this    | 200 ms  |
| p999   | 1 in 1000 — the "tail" latency           | 2000 ms |

Averages are useless for latency. A service with "average 10 ms" might have p99 at 2 seconds — 1% of users get a terrible experience, and the average hides it.

The gap between p50 and p99 tells you about consistency:

| Gap            | What it means                                                                                    |
| -------------- | ------------------------------------------------------------------------------------------------ |
| p99 ≈ 2× p50   | Consistent. Minor variance from GC pauses, cache misses.                                         |
| p99 ≈ 10× p50  | Something occasionally goes wrong — lock contention, cold cache, rare slow query.                |
| p99 ≈ 100× p50 | Bimodal — some requests hit a completely different code path (cache miss vs hit, index vs scan). |

### Where to look when it's slow

| Symptom                | Most likely cause                              | First thing to check         |
| ---------------------- | ---------------------------------------------- | ---------------------------- |
| All requests slow      | Database overloaded, connection pool exhausted | DB CPU, active connections   |
| Random requests slow   | GC pauses, noisy neighbour, lock contention    | GC logs, CPU steal, DB locks |
| Specific endpoint slow | Missing index, N+1 queries, large payload      | EXPLAIN ANALYZE, query count |
| Slow after deploy      | Cold cache, new query plan, memory leak        | Cache hit rate, query plans  |
| Gets slower over time  | Memory leak, table bloat, log disk full        | Memory usage, disk space     |
| Slow at certain times  | Traffic spikes, batch jobs, backup running     | Traffic patterns, cron jobs  |

## Key Takeaways

- **The database dominates** — in a typical request, 50-80% of time is in the database query. Optimise there first.
- **Connection reuse is free performance** — keep-alive, connection pooling, and session resumption eliminate repeated handshake costs.
- **Parallelise independent calls** — if you call 3 services sequentially at 5 ms each, that's 15 ms. In parallel, it's 5 ms.
- **Every network hop costs 1-3 ms** — microservice architectures pay this tax per service call. Keep call chains shallow.
- **Set timeouts on everything** — a missing timeout on one call can take down your entire system through cascading failure.
- **Measure percentiles, not averages** — p99 latency is what your worst 1% of users experience. Averages hide problems.
- **Most performance problems are configuration, not code** — missing indexes, no connection pooling, no caching, missing timeouts. Fix these before rewriting anything.

## What's Next

Now that you can trace where latency goes in a request, the next lesson dives deep into the component that dominates most of that time — databases. How Postgres, MySQL, MongoDB, and DynamoDB actually handle queries, their throughput profiles, and when to reach for read replicas, caching, or sharding.
