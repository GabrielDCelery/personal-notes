# Cockpit — Joker (Anatomy of a Request)

The Cockpit is where Joker flies the ship — obsessed with speed, routing, latency. How fast can we get there? What does each hop cost? Why are we making unnecessary stops?

Maps to: **02-anatomy-of-a-request.md**

## Theme

Latency — where does time go in a single request? Joker knows every millisecond of every route. He complains loudly about unnecessary network calls.

## Content to Anchor

- **0/1/5/50 mental model** — cached = 0ms, hop = 1ms, DB query = 5ms, unnecessary network call = 50ms. Healthy request ≈ 7ms.
- **The full request journey** — DNS, TCP, TLS, CDN, load balancer, app server, DB, back
- **Connection cost** — DNS/TCP/TLS only matter on first request, keep-alive makes them free
- **Authentication costs** — JWT = 0.1ms (no network), Session→Redis = 1ms, OAuth = 10-100ms
- **DB query costs** — indexed = 1-5ms, full scan = 100-10,000ms
- **Two bottleneck types** — DB fails as throughput (pool saturation), external API fails as latency (threads blocked)
- **Multi-service requests** — sequential = sum, parallel = max. Always parallelise independent calls.
- **Timeouts and cascading failure** — no timeout = one slow service takes down everything
- **Reading a latency profile** — p50 vs p99, not averages

## Still To Do

- Design specific anchors for each piece of content
- Write image prompts
