# Caching — Distilled

A cache's value depends on how skewed your access patterns are — the more skewed, the less memory you need to intercept the majority of reads.

## Core Mental Model: The Skew Principle

Most data isn't accessed equally. A small fraction gets nearly all the reads. This is why caching works: you don't need to cache everything, just the hot fraction. If access were truly uniform, a cache would need to hold the entire dataset — which is just a second database.

```
Typical access pattern (product catalogue):

  100% of reads ──────────────────────────────────────
                ██
                ████
                ████████
                ██████████████
                ████████████████████████████████████████
                │         │         │         │
                Top 1%    Top 5%    Top 20%   All products

  Top 20% of products → ~80% of reads
  Bottom 50% of products → rarely accessed
```

Cache the top 20% of items and you intercept 80% of reads. The more skewed the pattern, the higher your hit rate with less memory.

## Hit Rate — The Number That Matters

Hit rate tells you how much work you've moved off the database. But the relationship is non-linear: the higher you go, the more each percentage point is worth.

```
DB traffic remaining (lower is better):

80% hit rate  |████████████████████| 20% of reads still reach DB
90% hit rate  |██████████          | 10%  (2× better than 80%)
95% hit rate  |█████               |  5%  (4× better than 80%)
99% hit rate  |█                   |  1%  (20× better than 80%)
```

**Target ≥ 90%.** Below 80%, caching helps but doesn't transform. At 95%, your database barely handles reads. Monitor with `redis-cli INFO stats` → `keyspace_hits` / `keyspace_misses`.

Hit rate is driven by:

- **Access skew** — the more concentrated your hot data, the higher the rate at any cache size
- **Cache size** — bigger cache = fewer evictions = higher hit rate
- **TTL** — longer TTL = fewer expirations = higher hit rate (but more staleness)
- **Eviction policy** — **LRU is the right default** for most access patterns

## Caching Strategies

The four patterns differ on who does the loading and whether writes go through the cache.

**Cache-aside (default):** the application manages the cache explicitly. On a read, check cache first; on a miss, query the database and store the result. On a write, write to the database and delete the cache entry.

```
Read:   cache hit? → return                        (0.5 ms)
        cache miss? → DB → store in cache → return (5 ms, warm for next time)

Write:  DB write → delete cache key
        (next read repopulates from the fresh DB value)
```

Cache-aside degrades gracefully: if Redis is down, reads fall through to the database. Only data that's actually requested gets cached — no wasted memory.

**Read-through:** same flow, but the loading logic lives in the cache layer, not the application. App only talks to the cache. Keeps application code cleaner; behaviour is identical to cache-aside.

**Write-through:** every write goes through the cache synchronously before confirming. Cache is always fresh. Write latency is higher (cache + DB). Use when reads vastly outnumber writes and you need strong consistency.

**Write-behind:** writes go to the cache, then async to the database in the background. Write latency is minimal. Risk: cache crash before flush = data loss. **Only use for data where losing a few seconds of writes is acceptable** — counters, analytics, metrics.

| Strategy      | Read latency | Write cost | Consistency   | Data loss risk |
| ------------- | ------------ | ---------- | ------------- | -------------- |
| Cache-aside   | Low (on hit) | Normal     | Eventual, TTL | None           |
| Read-through  | Low (on hit) | Normal     | Eventual, TTL | None           |
| Write-through | Low (always) | Higher     | Strong        | None           |
| Write-behind  | Low (always) | Very low   | Strong reads  | Yes            |

**Default: cache-aside with TTL.**

## Cache Invalidation

Getting stale data out of the cache is harder than getting data in. There are three approaches.

**TTL-based expiration** is the floor. Every key has a time-to-live; after it expires, the next read fetches fresh data. Pick TTL based on how much staleness the data can tolerate:

| Data type                       | TTL             |
| ------------------------------- | --------------- |
| Product catalogue, config       | 5–60 minutes    |
| User profile                    | 1–5 minutes     |
| Session data                    | 30 min – 24 hrs |
| Real-time data (prices, scores) | 1–5 seconds     |
| Must-never-be-stale (balance)   | Don't cache     |

**Active invalidation:** when data changes, explicitly delete the cache entry. The next read misses, fetches fresh, and repopulates. Combined with a TTL as a safety net, this is the standard approach.

The critical rule: **delete on write, don't update.** Updating the cache on write creates a race condition — two concurrent writes can arrive at the cache in a different order than they arrived at the database, leaving stale data. Deleting avoids this: the next read always fetches the authoritative value.

**Event-driven (CDC):** database change events (Postgres WAL → Debezium → Kafka) trigger cache invalidation automatically. The application doesn't need to know about caching. Useful in microservices where multiple services write to the same tables. Adds infrastructure and a 100–500 ms invalidation delay.

## Sizing: Cache the Working Set, Not Everything

You don't need to cache the entire dataset to get high hit rates.

```
Example: 1,000,000 products × 5 KB = 5 GB total dataset
         Top 10% (hot products)   = 500 MB
         Hit rate with 500 MB cache: ~85–90%

You don't need 5 GB of cache. 500 MB gets you 90% hit rate.
```

Start small, measure hit rate, and grow only if it's too low. A 100 MB cache that achieves 90% hit rate is better than a 5 GB cache that's 70% idle.

## Cache Levels

Not all caching needs Redis. The fastest cache is in your application's own memory.

```
Scale: log10  | 0.001ms    0.01ms    0.1ms     1ms      20ms |
              0-----+---------+----------+--------+--------+-+

In-process (app memory) |·                   | ~0.001 ms
Local Redis             |████████████        | 0.2–0.5 ms
Remote Redis            |██████████████      | 0.5–2 ms
CDN                     |███████████████████ | 1–20 ms
```

In-process is 500–2000× faster than Redis — no network round trip. Use it for:

- Configuration and feature flags (refresh every 30–60 s)
- Small reference data (country codes, currency rates)
- Extremely hot keys that would create a hot key problem in Redis

The trade-off: each app server instance has its own copy. For reference data, fine. For user-specific or shared state, use Redis so all instances see the same values.

**CDN caching** is geographically distributed. A cache hit at a Tokyo edge node serves a Tokyo user in 2–5 ms instead of 150 ms to a US origin.

| Content type            | TTL           | Cache-Control header          |
| ----------------------- | ------------- | ----------------------------- |
| Static assets (JS, CSS) | 1 year        | `max-age=31536000, immutable` |
| Images, PDFs            | 1 year        | `max-age=31536000`            |
| Public API responses    | 5–60 seconds  | `max-age=30, s-maxage=60`     |
| Public HTML pages       | 10–60 seconds | `s-maxage=30`                 |
| Personalised content    | Don't cache   | `no-store` or `private`       |

Static assets should use content-hashed filenames (`styles.a1b2c3.css`) — the hash changes when the file changes, so you can cache aggressively for a year and still bust the cache on deploy. Build tools do this automatically.

## Redis vs Memcached

Both are fast (~0.5–1 ms). The choice is almost always Redis.

**Redis** is a data structure server: strings, hashes, sorted sets, lists, streams, pub/sub. It has persistence (survives restarts), replication, and Lua scripting for atomic operations. Single-threaded for commands — no lock contention, predictable latency.

**Memcached** is simpler: pure key-value, multi-threaded, no persistence, no data structures. It can reach 500,000+ ops/sec on simple GET/SET on high-core-count machines.

**Use Memcached only when:** you need maximum throughput for simple GET/SET, you don't need persistence, and you don't need any data structures. In every other case, Redis.

## Common Failure Modes

**Thundering herd (cache stampede):** a hot key expires, and hundreds of concurrent requests all miss simultaneously and hammer the database with the same query. Fix: **jitter TTLs** — use `TTL = 300 + random(0, 60)` so keys don't all expire together. For extremely hot keys, use stale-while-revalidate (return stale data, refresh in background) or a mutex so only one request fetches from the DB while others wait.

**Cache penetration:** repeated requests for keys that don't exist (e.g., probing for nonexistent user IDs). Every request misses cache and hits the DB uselessly. Fix: **cache the negative result** — store `key = NULL` with a short TTL (30–60 s). For high-volume attacks, add a bloom filter in front: it can say "definitely not in the database" with zero DB queries and very little memory.

**Cache avalanche:** bulk-populated keys all expire at the same time. Same root cause as thundering herd, same fix: jitter TTLs on population.

**Hot key:** one key (viral product, global config) receives so much traffic that the single Redis node holding it becomes the bottleneck. Fix: **local in-process cache** for the hottest keys (refresh every few seconds), or key splitting — `product:viral:1` through `product:viral:10`, requests randomly pick one, spreading load across Redis Cluster slots.

## When NOT to Cache

Caching adds complexity and trades consistency for speed. Skip it when:

- **Data changes constantly and staleness is unacceptable** — live prices, account balances, auction bids. Read from primary.
- **Every request is unique** — personalised search results, one-time tokens. Hit rate ≈ 0; all cost, no benefit.
- **The database is already fast enough** — if reads are 2 ms and DB headroom is comfortable, adding Redis just adds a network hop and an ops burden.
- **Write-heavy workload** — if you write more than you read, the cache is constantly being invalidated before anyone reads it.
- **Small dataset in a well-resourced DB** — if your entire hot dataset fits in the database's buffer pool (common with Postgres on large instances), the database IS the cache. Adding Redis in front just adds latency.

## Key Mental Models

1. **Cache because access is skewed.** 20% of data gets 80% of reads. Uniform access = caching doesn't help.
2. **Target ≥ 90% hit rate.** Below 80%, caching helps but doesn't transform. At 95%, the DB barely handles reads.
3. **Cache-aside with TTL is the default.** Simple, graceful under cache failure, works with any backend.
4. **Delete on write, never update.** Concurrent writes can corrupt the cache if you update; delete is always safe.
5. **Jitter TTLs.** Prevents thundering herd and cache avalanche with one line of code.
6. **In-process cache is 500–2000× faster than Redis.** Use it for config, feature flags, and hot reference data.
7. **Redis is the default.** Memcached only wins for very high throughput simple GET/SET on multi-core machines.
8. **Cache the working set, not the whole dataset.** 10% of your data often gives 90% hit rate.
