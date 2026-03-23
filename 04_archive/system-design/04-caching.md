# Caching

Lesson 03 showed that the database is almost always the bottleneck, and that scaling it is progressively harder — from indexes to vertical scaling to read replicas to sharding. Caching sits between your application and your database and eliminates the most common reads entirely. A well-configured cache removes 80-90% of database read traffic, often buying you more headroom than any other single change.

But caching isn't free. You're trading consistency for speed — cached data can be stale. You're adding a new component to operate. And a cache that's poorly designed can make things worse, not better. This lesson covers how caching actually works, when it helps, when it hurts, and how to make the trade-offs.

## Why Caching Works

Caching exploits a fundamental property of most applications: **access patterns are not uniform**. A small fraction of your data gets the vast majority of reads. This is the Pareto principle applied to data access — roughly 20% of data gets 80% of traffic.

```
Typical access pattern for a product catalogue:

  100% of reads ──────────────────────────────────────────────
                ██
                ██
                ████
                ██████
                ████████
                ██████████████
                ████████████████████████████████████████████████
                │         │         │         │         │
                Top 1%    Top 5%    Top 20%   Top 50%   All products

  Top 1% of products get ~30% of reads
  Top 20% of products get ~80% of reads
  Bottom 50% of products are rarely accessed
```

If you cache the top 20% of products in Redis, you intercept 80% of reads before they hit the database. Those reads take 0.5-1 ms (Redis) instead of 2-10 ms (database). Your database QPS drops by 80%, and your average read latency drops by 60-70%.

This only works because access patterns are skewed. If every piece of data were accessed equally (uniform distribution), you'd need to cache everything — which is just a second database. The more skewed your access pattern, the more effective caching is.

## How a Cache Fits in the Architecture

```
Without cache:

  App Server → Database
  Every read hits the DB. DB handles 10,000 QPS.

With cache:

  App Server → Cache (hit?) → yes → return cached data (0.5 ms)
                     │
                     no (miss)
                     │
                     ▼
                  Database → return data (5 ms) → store in cache → return to app
```

The cache sits in the read path. On a cache hit, the database is never touched. On a cache miss, the database is queried and the result is stored in the cache for next time.

## Caching Strategies

There are several patterns for how data flows between your app, cache, and database. Each makes different trade-offs between consistency, complexity, and performance.

### Cache-aside (lazy loading)

The most common pattern. The application manages the cache explicitly.

```
Read:
  1. App checks cache for key
  2. If found (hit): return cached value
  3. If not found (miss): query database, store result in cache, return

Write:
  1. App writes to database
  2. App invalidates (deletes) the cache entry
  3. Next read will miss, fetch from DB, and repopulate cache
```

**Why it's popular:** Simple. The app controls everything. Cache failures are graceful — if Redis is down, reads fall through to the database (slower, but still works). Only data that's actually requested gets cached (no wasted memory on data nobody reads).

**The risk — race condition on write:**

```
Thread A: UPDATE user SET name='Bob'
Thread B:                               SELECT user → gets 'Alice' (old value)
Thread A: DELETE cache key
Thread B:                               SET cache key = 'Alice' (stale!)
```

Thread B reads the old value from the database after Thread A wrote but before the cache was invalidated, then caches the stale value. This is rare in practice (requires exact timing), and the staleness is bounded by the TTL — but it can happen.

**Mitigation:** Set a short TTL (30-300 seconds) so stale data self-corrects. For most applications, this is sufficient. For critical data (account balance), don't cache it — read from primary.

### Read-through

The cache itself is responsible for loading data on a miss.

```
Read:
  1. App asks cache for key
  2. If found (hit): cache returns value
  3. If not found (miss): cache queries database, stores result, returns to app

Write:
  Same as cache-aside — app writes to DB and invalidates cache
```

The difference from cache-aside: the loading logic lives in the cache layer, not in the application. This keeps application code cleaner — the app only talks to the cache, never directly to the database for reads. AWS ElastiCache doesn't support this natively, but you can build it with a thin wrapper.

### Write-through

Every write goes through the cache to the database. The cache is always up to date.

```
Write:
  1. App writes to cache
  2. Cache synchronously writes to database
  3. Cache confirms write to app

Read:
  1. App reads from cache (always has latest data)
```

**Advantage:** Cache is never stale. No race conditions. Reads are always consistent.

**Disadvantage:** Every write is slower — it goes through the cache AND the database. You're adding latency to writes to guarantee freshness on reads. Also, you're caching data that might never be read (wasting memory).

Use this when: consistency matters more than write speed, and you read data much more often than you write it.

### Write-behind (write-back)

The app writes to the cache, and the cache asynchronously writes to the database later.

```
Write:
  1. App writes to cache → returns immediately (fast!)
  2. Cache queues the write
  3. Cache writes to database in the background (batched, async)

Read:
  1. App reads from cache (always has latest data)
```

**Advantage:** Writes are extremely fast (only cache latency, ~0.5 ms). Writes to the database are batched, reducing DB load.

**Disadvantage:** If the cache crashes before flushing to the database, you lose data. This is fundamentally unsafe for data you can't afford to lose (orders, payments). Useful for metrics, counters, analytics — data where losing a few seconds of writes is acceptable.

### Which strategy to use

| Strategy      | Read latency | Write latency         | Consistency            | Complexity | Data loss risk    |
| ------------- | ------------ | --------------------- | ---------------------- | ---------- | ----------------- |
| Cache-aside   | Low (on hit) | Normal (DB write)     | Eventual (TTL-bounded) | Low        | None              |
| Read-through  | Low (on hit) | Normal (DB write)     | Eventual (TTL-bounded) | Medium     | None              |
| Write-through | Low (always) | Higher (cache + DB)   | Strong                 | Medium     | None              |
| Write-behind  | Low (always) | Very low (cache only) | Strong for reads       | High       | Yes (cache crash) |

**Default choice:** cache-aside with TTL. It's the simplest, works with any cache, and degrades gracefully.

## Cache Invalidation

There are only two hard problems in computer science: cache invalidation, naming things, and off-by-one errors.

Cache invalidation is the problem of keeping cached data in sync with the source of truth (the database). There are three approaches, each with trade-offs.

### TTL-based expiration

Every cached entry has a time-to-live. After it expires, the next read triggers a cache miss and fetches fresh data.

```
SET user:42 {"name": "Alice"} TTL 300   ← expires in 5 minutes

Read at T+0:   cache hit → "Alice"
Read at T+200: cache hit → "Alice"
Read at T+301: cache miss → fetch from DB → "Bob" (updated in DB at T+100)
```

**The trade-off is the TTL value:**

| TTL                   | Staleness                | Cache hit rate      | DB load  |
| --------------------- | ------------------------ | ------------------- | -------- |
| 10 seconds            | Very low                 | Lower (more misses) | Higher   |
| 60 seconds            | Up to 1 minute           | Good                | Moderate |
| 300 seconds (5 min)   | Up to 5 minutes          | High                | Low      |
| 3600 seconds (1 hour) | Up to 1 hour             | Very high           | Very low |
| No TTL                | Forever (until eviction) | Highest             | Lowest   |

**How to choose a TTL:**

- Product catalogue, configuration: 5-60 minutes (changes rarely, staleness is fine)
- User profile: 1-5 minutes (changes occasionally, brief staleness acceptable)
- Session data: 30 minutes - 24 hours (set by business rules, not freshness)
- Real-time data (stock prices, live scores): don't cache, or TTL of 1-5 seconds
- Data that must never be stale: don't cache it — read from primary

### Active invalidation

When data changes, explicitly delete or update the cache entry.

```
1. App writes to DB: UPDATE users SET name='Bob' WHERE id=42
2. App deletes cache: DEL user:42
3. Next read: cache miss → fetches "Bob" from DB → caches it
```

This is more consistent than TTL alone — data is stale only for the brief window between the DB write and the cache delete. Combined with a TTL as a safety net, it's the standard approach for most applications.

**Delete vs update:**

| Approach        | What happens                                 | Pro                        | Con                                  |
| --------------- | -------------------------------------------- | -------------------------- | ------------------------------------ |
| Delete on write | Remove cache entry, let next read repopulate | Simple, no race conditions | One cache miss after each write      |
| Update on write | Write new value to cache AND database        | No cache miss after write  | Race condition if two writes overlap |

**Delete is almost always safer.** Two concurrent updates can arrive at the cache in a different order than they arrived at the database, leaving the cache with an older value. Delete avoids this — the next read always fetches the latest from the database.

### Event-driven invalidation

Use database change events (CDC — Change Data Capture) to invalidate or update the cache automatically.

```
1. App writes to DB
2. DB emits a change event (Postgres WAL → Debezium → Kafka)
3. Cache invalidation service consumes the event
4. Deletes the cache entry

App doesn't need to know about caching — it just writes to the DB.
```

This decouples cache management from application code. Useful in microservice architectures where multiple services write to the same database — you don't need every service to remember to invalidate the cache.

The downside is added infrastructure (Debezium/Kafka) and a small delay between the DB write and cache invalidation (typically 100-500 ms).

## Cache Sizing and Hit Rate

### Hit rate — the single most important metric

```
Hit rate = cache hits / (cache hits + cache misses)
```

| Hit rate | What it means                       | Effect on database                           |
| -------- | ----------------------------------- | -------------------------------------------- |
| 50%      | Half of reads still hit DB          | 2x reduction — helps, but not transformative |
| 80%      | 4 out of 5 reads served from cache  | 5x reduction — significant                   |
| 90%      | 9 out of 10 reads served from cache | 10x reduction — DB barely loaded for reads   |
| 95%      | Only 1 in 20 reads hits DB          | 20x reduction — DB mostly handles writes     |
| 99%      | Almost everything from cache        | Extreme — only cold/new data hits DB         |

The jump from 80% to 95% is more valuable than the jump from 0% to 80%. At 80%, your DB handles 20% of reads. At 95%, it handles 5% — a 4x further reduction.

**What determines hit rate:**

- **Access pattern skew** — the more skewed (fewer hot items), the higher the hit rate with less memory
- **Cache size** — a bigger cache holds more items, fewer evictions, higher hit rate
- **TTL** — longer TTL = fewer expirations = higher hit rate (but more staleness)
- **Eviction policy** — LRU (Least Recently Used) is the default and works well for most patterns

### How much memory do you need?

Think of it as: **cache the working set, not the whole dataset.**

```
Example: E-commerce product catalogue
  Total products: 1,000,000
  Total size: 1M × 5 KB = 5 GB
  Hot products (top 10%): 100,000
  Hot products size: 100K × 5 KB = 500 MB
  Hit rate with 500 MB cache: ~85-90% (because top 10% gets ~90% of traffic)

  You DON'T need 5 GB of cache to be effective. 500 MB gets you 90% hit rate.
```

Start small, measure hit rate, and grow the cache only if the hit rate is too low. Redis gives you `INFO stats` which shows `keyspace_hits` and `keyspace_misses` — monitor this.

## Redis vs Memcached

Both are in-memory key-value stores. Both are fast (~0.5-1 ms per operation). The choice comes down to what features you need.

### Redis

Redis is the default choice for most caching use cases and more. It's a data structure server — not just key-value, but strings, hashes, lists, sets, sorted sets, streams, and more.

**Key features:**

- Persistence (RDB snapshots + AOF log) — survives restarts
- Replication — primary/replica for high availability
- Data structures — sorted sets for leaderboards, lists for queues, hashes for objects
- Pub/sub — real-time messaging
- Lua scripting — atomic complex operations
- TTL per key — automatic expiration
- Single-threaded for commands — no lock contention, predictable latency

**Typical deployment:**

```
Application
  │
  ▼
Redis Primary (writes + reads)
  │
  ├─→ Redis Replica 1 (reads, failover)
  └─→ Redis Replica 2 (reads, failover)
```

For larger datasets, Redis Cluster shards data across multiple nodes:

```
Redis Cluster (3 primaries, each with a replica):
  Slots 0-5460     → Primary A (Replica A')
  Slots 5461-10922 → Primary B (Replica B')
  Slots 10923-16383→ Primary C (Replica C')

  Each key hashes to a slot. Client routes to the right primary.
  Total capacity: 3x a single instance.
```

**Typical numbers (ElastiCache, cache.r6g.large — 13 GB):**

| Metric                 | Value                                              |
| ---------------------- | -------------------------------------------------- |
| GET/SET latency        | 0.2-0.5 ms (same AZ), 0.5-1 ms (cross-AZ)          |
| Throughput             | 80,000-120,000 ops/sec                             |
| Max memory (practical) | 10-12 GB (leave headroom for fragmentation + fork) |
| Max connections        | ~65,000                                            |
| Replication lag        | <1 ms typically                                    |

### Memcached

Memcached is simpler and older. It's a pure key-value cache — no data structures, no persistence, no replication.

**Key features:**

- Multi-threaded — uses all CPU cores (Redis is single-threaded for commands)
- Higher throughput for simple GET/SET — can reach 500,000+ ops/sec
- Simpler — less to configure, less to go wrong
- No persistence — cache is empty after restart
- Slab allocation — efficient memory management for uniform-size values

**When to choose Memcached over Redis:**

- You only need simple GET/SET (no data structures, no pub/sub)
- You need maximum throughput for simple operations
- You don't need persistence (cache is disposable)
- Multi-threaded performance matters (very high core count machines)

**When to choose Redis over Memcached:**

- You need data structures (sorted sets, lists, hashes)
- You need persistence (cache warm after restart)
- You need replication and automatic failover
- You need pub/sub or streams
- You need Lua scripting for atomic operations

**For most applications, Redis is the better default.** Memcached wins in the specific case of very high throughput simple caching on multi-core machines, but Redis covers far more use cases.

## CDN Caching

A CDN (Content Delivery Network) is a geographically distributed cache. It sits between your users and your origin servers, caching responses at edge locations around the world.

```
Without CDN:
  User in Tokyo → Origin in eu-west-1 (150 ms RTT) → 200 ms total

With CDN:
  User in Tokyo → CDN edge in Tokyo (2 ms RTT) → cache hit → 5 ms total
  User in Tokyo → CDN edge in Tokyo → cache miss → Origin (150 ms) → cached for next request
```

### What to cache at the CDN

| Content type                   | Cache strategy       | TTL           | Cache-Control header          |
| ------------------------------ | -------------------- | ------------- | ----------------------------- |
| Static assets (JS, CSS, fonts) | Always cache         | 1 year        | `max-age=31536000, immutable` |
| Images, video, PDFs            | Always cache         | 1 year        | `max-age=31536000`            |
| Public API responses           | Cache if possible    | 5-60 seconds  | `max-age=30, s-maxage=60`     |
| HTML pages (public)            | Cache with short TTL | 10-60 seconds | `s-maxage=30`                 |
| Personalised content           | Don't cache          | —             | `no-store` or `private`       |
| API responses with auth        | Don't cache at CDN   | —             | `private, max-age=0`          |

`s-maxage` controls CDN/shared cache TTL separately from `max-age` which controls browser cache TTL. This lets you cache at the CDN for 60 seconds while telling the browser not to cache (or cache for less time).

### Cache busting for static assets

Static assets should be cached aggressively (1 year), but you need a way to force users to get new versions when you deploy.

```
Bad:  /styles.css          → cached for 1 year. Deploy new CSS? Users still see old.
Good: /styles.a1b2c3.css   → cached for 1 year. Deploy? New filename, new file fetched.

The hash in the filename is the content hash. If the file changes, the hash changes,
the URL changes, and the CDN/browser treats it as a new file.
```

Build tools (webpack, vite, esbuild) do this automatically.

## Common Caching Problems

### Thundering herd (cache stampede)

A popular cache key expires. Hundreds of requests simultaneously miss the cache and all query the database at once.

```
T=0:    Cache key "popular_product" expires (TTL reached)
T=0.1:  100 concurrent requests all check cache → all miss
T=0.2:  100 requests all query the database simultaneously
T=0.3:  Database overloaded with 100 identical queries
T=0.5:  All 100 responses come back, all 100 SET the same cache key
```

**Solutions:**

| Solution               | How it works                                                                                                                     | Complexity |
| ---------------------- | -------------------------------------------------------------------------------------------------------------------------------- | ---------- |
| Lock/mutex             | First request to miss acquires a lock, fetches from DB, populates cache. Other requests wait for the lock, then read from cache. | Medium     |
| Early recompute        | Refresh the cache before TTL expires (e.g., at 80% of TTL, a background job refreshes)                                           | Medium     |
| Stale-while-revalidate | Return stale data immediately, refresh in background                                                                             | Low        |
| Jittered TTLs          | Add random variance to TTL (300 ± 30 seconds) so keys don't all expire at once                                                   | Low        |

For most applications, jittered TTLs are sufficient. The thundering herd is mainly a problem for extremely hot keys (millions of reads per minute).

### Cache penetration

Repeated requests for data that doesn't exist. Every request misses the cache AND misses the database, providing no benefit from caching.

```
Request: GET /user/99999999  (doesn't exist)
  → cache miss
  → DB query: SELECT * FROM users WHERE id=99999999 → no rows
  → nothing to cache
  → next request: same thing, cache miss again, DB query again
```

**Solutions:**

- Cache negative results: store `user:99999999 = NULL` with a short TTL (30-60 seconds)
- Bloom filter: a probabilistic data structure that can tell you "definitely not in the database" with no DB query. Uses very little memory.

### Cache avalanche

Many cache keys expire at the same time (e.g., you populated the cache in bulk and all keys got the same TTL). The database is suddenly hit with the full read load.

**Solution:** Jitter TTLs when setting them. Instead of `TTL = 300`, use `TTL = 300 + random(0, 60)`. Keys expire at different times, spreading the load.

### Hot key problem

One specific key gets so much traffic that the single Redis node holding it becomes a bottleneck. This happens with viral content, flash sales, or global configuration.

**Solutions:**

- Local cache: cache the hottest keys in application memory (Go map, Node.js Map). No network round trip. Refresh every few seconds.
- Read replicas: spread reads across multiple Redis replicas.
- Key splitting: instead of one key `product:viral`, use `product:viral:1`, `product:viral:2`, ..., `product:viral:10`. Each request randomly picks one. Spreads the load across multiple slots in Redis Cluster.

## Application-Level Caching

Not all caching requires Redis. Sometimes the most effective cache is in your application's own memory.

| Cache level             | Latency    | Best for                                  | Watch out for                             |
| ----------------------- | ---------- | ----------------------------------------- | ----------------------------------------- |
| In-process (app memory) | ~0.001 ms  | Config, feature flags, hot reference data | Stale across instances, uses app memory   |
| Local Redis/Memcached   | 0.2-0.5 ms | Session data, computed results            | Single point of failure if not replicated |
| Remote Redis/Memcached  | 0.5-2 ms   | Shared state, distributed cache           | Network latency, extra infra              |
| CDN                     | 1-20 ms    | Static assets, public content             | Limited control over invalidation         |

**In-process cache** is 500-2000x faster than Redis because there's no network round trip. Use it for:

- Configuration that changes rarely (refresh every 60 seconds)
- Feature flags (refresh every 30 seconds)
- Small reference data (country codes, currency rates)
- Extremely hot keys (compute once, serve from memory)

The trade-off: each app server instance has its own copy. If you have 10 instances, you have 10 caches that might be slightly out of sync. For reference data and config, this is fine. For user-specific data, use Redis so all instances share the same cache.

## When NOT to Cache

Caching isn't always the right answer. It adds complexity and creates consistency problems. Don't cache when:

- **Data changes constantly and staleness is unacceptable** — real-time stock prices, account balances after transactions, live auction bids. Read from primary.
- **Every request is unique** — personalised search results, one-time tokens, unique reports. Hit rate will be near zero — all cost, no benefit.
- **The database is fast enough** — if your reads take 2 ms and you have 500 QPS on a database that handles 15,000 QPS, adding a cache adds complexity with no meaningful benefit.
- **Write-heavy workload** — if you write more often than you read, the cache is constantly being invalidated before anyone reads it. The overhead of cache management exceeds the benefit.
- **Small dataset that fits in DB memory** — if your entire dataset is 1 GB and your database has 32 GB of RAM, the database IS the cache (data is in the buffer pool). Adding Redis in front just adds a network hop.

## Key Takeaways

**1. Caching works because access patterns are skewed.** 20% of data gets 80% of reads. Cache the hot data, ignore the rest. If access is uniform, caching doesn't help much.

**2. Cache-aside with TTL is the default pattern.** App checks cache, misses fall through to database, results are stored in cache with a TTL. Simple, graceful under failure, works with any cache.

**3. The hit rate is everything.** Below 80%, caching helps but doesn't transform. Above 90%, your database barely handles reads. Monitor `keyspace_hits` and `keyspace_misses` in Redis.

**4. Invalidation is the hard part.** TTL gives you eventual consistency with a bounded staleness window. Active invalidation (delete on write) gives you near-immediate consistency. Pick the right trade-off for each piece of data.

**5. Redis is the default choice.** It's a cache, a data structure server, a pub/sub broker, and a session store. Memcached wins only for very high throughput simple GET/SET on multi-core machines.

**6. Don't cache everything.** If the database is fast enough, if the data changes constantly, or if every request is unique, a cache adds complexity with no benefit. Cache the things that are read often, change rarely, and tolerate brief staleness.

## What's Next

Caching handles the read side. The next lesson covers the write side — queues and async processing. When should you process something synchronously vs put it on a queue? How do Kafka, SQS, and RabbitMQ compare? How do you size batch processing? This is where you decide whether a request gets an immediate response or a "we'll process this shortly."
