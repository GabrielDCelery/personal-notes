# Caching

A well-configured cache removes 80-90% of database read traffic. It's often the single most effective scaling move — but only when used correctly.

## Why Caching Works

Access patterns are skewed. A small fraction of data gets the vast majority of reads.

```
  100% of reads ──────────────────────────────────────────────
                ██
                ██
                ████
                ██████
                ████████
                ██████████████
                ████████████████████████████████████████████████
                │         │         │         │         │
                Top 1%    Top 5%    Top 20%   Top 50%   All items

  Top 20% of data gets ~80% of reads.
  Cache that 20% → intercept 80% of DB queries.
```

If access were uniform, you'd need to cache everything — which is just a second database. The more skewed the pattern, the more effective caching is.

## How a Cache Fits

```
Cache hit:   App → Cache → return (0.5 ms)
Cache miss:  App → Cache → DB → store in cache → return (5 ms)
```

On a hit, the database is never touched.

## Caching Strategies

| Strategy      | Read latency | Write latency       | Consistency            | Data loss risk | Use when                              |
| ------------- | ------------ | ------------------- | ---------------------- | -------------- | ------------------------------------- |
| Cache-aside   | Low (on hit) | Normal (DB write)   | Eventual (TTL-bounded) | None           | Default — simple, degrades gracefully |
| Read-through  | Low (on hit) | Normal (DB write)   | Eventual (TTL-bounded) | None           | Want loading logic out of app code    |
| Write-through | Low (always) | Higher (cache + DB) | Strong                 | None           | Consistency > write speed             |
| Write-behind  | Low (always) | Very low (cache)    | Strong for reads       | Yes            | Metrics, counters — loss is tolerable |

**Default: cache-aside.** App checks cache, miss falls through to DB, result stored in cache with TTL. Simplest, works with any cache, graceful under cache failure.

**Write-behind is dangerous** — if the cache crashes before flushing, writes are lost. Never use it for orders, payments, or anything durable.

## Cache Invalidation

### TTL-based

Every key has a time-to-live. Expired keys trigger a cache miss and refresh from DB.

| TTL     | Staleness      | Hit rate  | DB load  | Good for                        |
| ------- | -------------- | --------- | -------- | ------------------------------- |
| 10 sec  | Very low       | Lower     | Higher   | Near-real-time data             |
| 60 sec  | Up to 1 min    | Good      | Moderate | User profiles                   |
| 5 min   | Up to 5 min    | High      | Low      | Product catalogue, config       |
| 1 hour+ | Up to 1 hour   | Very high | Very low | Reference data, rarely changing |
| No TTL  | Until eviction | Highest   | Lowest   | Immutable data only             |

### Active invalidation

On write: update the DB, then **delete** the cache key. Next read repopulates.

Delete is safer than update on write — two concurrent writes can arrive at the cache in a different order than they hit the DB, leaving stale data. Delete forces the next read to always fetch fresh.

Combine both: active invalidation for correctness + TTL as a safety net.

### Event-driven (CDC)

DB emits change events (Postgres WAL → Debezium → Kafka) → cache invalidation service deletes keys. App code doesn't need to know about caching. Useful in microservices where multiple services write to the same DB. Adds infra complexity and ~100-500 ms invalidation lag.

## Hit Rate

The single most important metric.

```
Hit rate = cache hits / (cache hits + cache misses)
```

| Hit rate | DB read reduction | What it means                 |
| -------- | ----------------- | ----------------------------- |
| 50%      | 2×                | Helps, not transformative     |
| 80%      | 5×                | Significant                   |
| 90%      | 10×               | DB barely loaded for reads    |
| 95%      | 20×               | DB mostly handles writes only |
| 99%      | 100×              | Only cold/new data hits DB    |

The jump from 80% → 95% is more valuable than 0% → 80%.

**What drives hit rate:** access pattern skew, cache size, TTL length, eviction policy (LRU is the right default).

**How to size:** cache the working set, not the whole dataset. If top 10% of products get 90% of traffic and each is 5 KB, you need 10% × total items × 5 KB — not the full catalogue. Start small, measure, grow only if hit rate is too low.

Monitor via Redis `INFO stats` → `keyspace_hits` / `keyspace_misses`.

## Redis vs Memcached

| Feature         | Redis                                              | Memcached      |
| --------------- | -------------------------------------------------- | -------------- |
| Data structures | Strings, hashes, lists, sets, sorted sets, streams | Strings only   |
| Persistence     | Yes (RDB + AOF)                                    | No             |
| Replication     | Yes (primary/replica)                              | No             |
| Pub/sub         | Yes                                                | No             |
| Threading       | Single-threaded (commands)                         | Multi-threaded |
| Throughput      | ~100K ops/sec                                      | ~500K ops/sec  |
| Latency         | 0.2-1 ms                                           | 0.2-1 ms       |

**Default: Redis.** It covers far more use cases. Memcached wins only for very high throughput simple GET/SET on multi-core machines where you need nothing else.

Redis Cluster shards data across nodes when a single instance isn't enough (practical limit ~10-25 GB per node).

## CDN Caching

A CDN is a geographically distributed cache. Cache hits are served from an edge node near the user instead of your origin.

```
Without CDN:  User in Tokyo → Origin in eu-west-1 → ~200 ms
With CDN:     User in Tokyo → Edge in Tokyo → cache hit → ~5 ms
```

| Content type                | Cache? | TTL       | Cache-Control header          |
| --------------------------- | ------ | --------- | ----------------------------- |
| Static assets (JS/CSS)      | Always | 1 year    | `max-age=31536000, immutable` |
| Images, video               | Always | 1 year    | `max-age=31536000`            |
| Public API responses        | Yes    | 5-60 sec  | `s-maxage=60`                 |
| Public HTML pages           | Yes    | 10-60 sec | `s-maxage=30`                 |
| Personalised content        | Never  | —         | `private, no-store`           |
| Authenticated API responses | Never  | —         | `private, max-age=0`          |

Static assets should use content-hashed filenames (`styles.a1b2c3.css`) so you can set a 1-year TTL and still deploy changes — new hash = new URL = cache miss.

## Common Problems

### Thundering herd

A hot key expires. Hundreds of requests simultaneously miss and hammer the DB.

**Fix:** jitter TTLs (`TTL = 300 + random(0, 60)`) so keys don't all expire together. For extremely hot keys: return stale data while refreshing in the background (stale-while-revalidate), or use a mutex so only the first miss queries the DB.

### Cache penetration

Requests for data that doesn't exist bypass the cache every time (nothing to cache on a miss).

**Fix:** cache negative results (`user:99999 = NULL`, short TTL). For high-volume attacks, use a Bloom filter to answer "definitely not in DB" without a query.

### Cache avalanche

Bulk-loaded keys all get the same TTL and expire simultaneously. DB gets hit with full load at once.

**Fix:** jitter TTLs on write.

### Hot key

One key gets so much traffic that the single Redis node holding it is a bottleneck.

**Fix:** in-process cache (app memory, no network hop), or split the key (`product:viral:1` through `:10`) and randomly pick one per request.

## Cache Levels

| Level            | Latency   | Best for                              | Watch out for                           |
| ---------------- | --------- | ------------------------------------- | --------------------------------------- |
| In-process (RAM) | ~0.001 ms | Config, feature flags, hot reference  | Stale across instances, uses app memory |
| Redis/Memcached  | 0.5-2 ms  | Shared state, sessions, computed data | Extra infra, network hop                |
| CDN              | 1-20 ms   | Static assets, public responses       | Limited invalidation control            |

In-process is 500-2000x faster than Redis — no network. Use it for config and reference data that changes rarely and doesn't need to be consistent across instances.

## When NOT to Cache

- **Data that can't be stale** — account balances, live auction bids, real-time prices. Read from primary.
- **Every request is unique** — personalised search, one-time tokens. Hit rate ≈ 0%. All overhead, no benefit.
- **DB is already fast enough** — if reads take 2 ms and you have headroom, adding Redis adds complexity with no meaningful gain.
- **Write-heavy workload** — cache is invalidated before anyone reads it.
- **Dataset fits in DB memory** — if your whole dataset is 1 GB and the DB has 32 GB RAM, the DB buffer pool IS the cache. Redis just adds a network hop.

## Key Mental Models

1. **Caching works because access is skewed.** 20% of data = 80% of reads. Cache the hot slice.
2. **Cache-aside + TTL is the default.** Simple, graceful under failure.
3. **Hit rate is the metric.** Below 80% = helps but not transformative. Above 90% = DB mostly handles writes.
4. **Delete on write, not update.** Avoids race conditions from concurrent writes.
5. **Jitter your TTLs.** Prevents thundering herd and cache avalanche.
6. **Redis is the default.** Memcached only if you need raw GET/SET throughput and nothing else.
7. **Don't cache everything.** Stale-intolerant data, unique requests, and already-fast reads don't benefit.
