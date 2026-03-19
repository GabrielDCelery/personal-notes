# Numbers and Quick Math

Estimation isn't about exact numbers. It's about getting within an order of magnitude to make decisions: do I need a cache? Will one database handle this? Should this be async?

## The Three Speed Worlds

Everything lives in one of three worlds. The gap between each is ~1,000x.

| World        | Range      | What lives here                           |
| ------------ | ---------- | ----------------------------------------- |
| Nanoseconds  | 1–300 ns   | L1/L2/L3 cache, RAM, mutex, syscall       |
| Microseconds | 1–1,000 µs | SSD random read, context switch           |
| Milliseconds | 0.5–300 ms | Network (DC), HDD, network (cross-region) |

Memorise the world, not the exact number. If it touches the network, it's milliseconds. If it's in memory, it's nanoseconds.

The gap between RAM and a network call is 4-5 orders of magnitude. This is why caching works.

## Operation Latency Reference

**The simple model: cache ~1 ms, DB read ~1-10 ms, DB write ~2-10 ms, external API ~50-500 ms.**

Everything is dominated by the network. Redis is fast not because of the lookup (microseconds) but because it's in the same DC. External APIs are slow because the internet adds 50-500 ms before your code does anything.

| Operation                          | Latency      | Notes                                   |
| ---------------------------------- | ------------ | --------------------------------------- |
| Redis / Memcached GET or SET       | 0.5–1 ms     | Almost entirely network round trip      |
| Postgres query (indexed)           | 1–5 ms       | Network + parse + plan + execute        |
| Postgres query (complex JOIN)      | 10–100 ms    | Depends on data size and indexes        |
| Postgres full table scan (1M rows) | 100–1,000 ms | Missing index — fix this                |
| Postgres INSERT (single)           | 2–10 ms      | WAL + index update + fsync              |
| Postgres batch INSERT (1K rows)    | 20–100 ms    | ~10x faster per row than individual     |
| Postgres transaction (3-5 stmts)   | 10–30 ms     | Count as 1 tx for throughput estimation |
| MongoDB find by \_id               | 1–5 ms       | Similar to indexed Postgres             |
| DynamoDB GetItem                   | 5–10 ms      | Single-digit ms promise + network       |
| Kafka produce (acks=1)             | 2–5 ms       | Broker writes to local log              |
| Kafka produce (acks=all)           | 5–30 ms      | Waits for all replicas                  |
| SQS SendMessage                    | 10–30 ms     | HTTP API call to AWS                    |
| HTTP call (same DC)                | 5–50 ms      | Serialisation + network + processing    |
| HTTP call (external)               | 50–500 ms    | Internet latency + processing           |
| JSON serialise/deserialise (1 KB)  | 5–50 µs      | CPU-bound, negligible                   |
| JSON serialise/deserialise (1 MB)  | 5–50 ms      | Can dominate if unpagenated             |
| Bcrypt hash (cost 10)              | ~100 ms      | Intentionally slow                      |

## Throughput Reference

**DB writes ~1K/sec, DB reads ~10K/sec, cache ~100K/sec. Each tier is 10x the previous.**

| System                | Throughput               | Notes                                   |
| --------------------- | ------------------------ | --------------------------------------- |
| Postgres/MySQL reads  | 5,000–20,000 QPS         | Indexed, connection pooled              |
| Postgres/MySQL writes | 1,000–10,000 QPS         | Depends on indexes, WAL, fsync          |
| MongoDB reads         | 10,000–50,000 ops/sec    | Depends on working set in RAM           |
| MongoDB writes        | 5,000–25,000 ops/sec     |                                         |
| DynamoDB              | Unlimited (provisioned)  | Pay per RCU/WCU                         |
| Redis                 | 80,000–100,000 ops/sec   | Single-threaded, scale via cluster      |
| Kafka                 | 100,000–500,000 msgs/sec | Depends on message size and replication |
| SQS (standard)        | ~unlimited               | FIFO: 300/sec, 3,000/sec with batching  |
| Node.js               | 10,000–30,000 req/sec    | Simple JSON API                         |
| Go                    | 30,000–100,000 req/sec   | Simple JSON API                         |

The app server is almost never the bottleneck. A single app server handles 10K-30K req/sec. Your DB hits its ceiling first.

**Postgres on RDS — instance sizing:**

| Instance       | vCPU | RAM    | Reads/sec     | Writes/sec   |
| -------------- | ---- | ------ | ------------- | ------------ |
| db.t3.micro    | 2    | 1 GB   | 200–500       | 50–200       |
| db.r6g.large   | 2    | 16 GB  | 2,000–5,000   | 500–3,000    |
| db.r6g.xlarge  | 4    | 32 GB  | 5,000–15,000  | 1,000–8,000  |
| db.r6g.4xlarge | 16   | 128 GB | 15,000–50,000 | 5,000–20,000 |

The "DB writes ~1K, reads ~10K" simple model maps to `db.r6g.xlarge` (~$300/month). Most production systems run here.

## Size Reference

**1 million rows × 500 bytes = 500 MB.** Most people overestimate how much storage they need. 10 million users with profiles is ~20 GB — fits in RAM on one machine.

| Data type              | Size          |
| ---------------------- | ------------- |
| UUID                   | 16 bytes      |
| Timestamp (Unix int64) | 8 bytes       |
| Average DB row         | 100–500 bytes |
| Short text (tweet)     | 200–500 bytes |
| JSON API response      | 1–10 KB       |
| Log line               | 200–500 bytes |
| Photo (JPEG)           | 2–5 MB        |
| 1-min video (720p)     | 10–30 MB      |
| Full movie (1080p)     | 1.5–4 GB      |

Storage shortcuts:

```
1M rows  × 500 bytes = 500 MB
1B rows  × 500 bytes = 500 GB   ← start thinking about sharding
1M users × 1 KB      = 1 GB
1M items × 1 MB      = 1 TB     ← photos, documents → S3
```

## The Quick Math

### 1 day = 10^5 seconds (drop 5 zeros)

```
RPS = (users × actions_per_day) / 100,000
```

Traffic peaks at 2-5x average. Multiply by 3 as default.

| Users      | Actions/day | Avg RPS | Peak (×3) |
| ---------- | ----------- | ------- | --------- |
| 100,000    | 10          | 10      | 30        |
| 100,000    | 100         | 100     | 300       |
| 1,000,000  | 50          | 500     | 1,500     |
| 10,000,000 | 20          | 2,000   | 6,000     |

### RPS → DB QPS

Reads multiply (3-5 queries per request). Writes count as 1 transaction per request (not per row).

```
Read endpoints:  RPS × 3-5 = read QPS
Write endpoints: RPS × 1   = write transactions/sec
```

### Storage per year

```
Storage = users × data_per_user × retention
```

### Bandwidth

```
Bandwidth = RPS × avg_response_size
```

1 Gbps ≈ 100K RPS of 1 KB responses, or 1K RPS of 1 MB responses.

### Cache size

```
Cache size = hot_items × item_size
```

Redis single instance comfortably holds 10-25 GB. Above that, cluster.

## Worked Example: E-Commerce

**Given:** 100K users, 100 orders/day each, 10 page views per order. Traffic concentrated in 8 hours (30K seconds).

**Traffic:**

```
Orders:   100K × 100 / 30,000 = ~333 RPS avg → ~1,000 peak
Browsing: 10 pages/order → 1B views/day / 30,000 = ~3,333 RPS avg → ~10,000 peak
```

**DB load:**

```
Browse:  10,000 RPS × 4 reads = 40,000 read QPS
Orders:   1,000 RPS × 1 tx   =  1,000 write tx/sec
Total: ~41,000 read QPS — exceeds a single Postgres instance
```

**Fix:** Cache top 50K products (50K × 5 KB = 250 MB, trivial Redis). If 80% of views hit cached products, reads drop from 41K to ~9K QPS — comfortable for one instance.

**Storage:** 100K users × 100 orders/day × 365 × 1 KB ≈ 3.6 TB/year → plan to archive old orders.

**Verdict:** 2-3 app servers, one Postgres + PgBouncer, one Redis. Read replicas only if cache hit rate is lower than expected.

## Scale Intuition

| Peak RPS   | What it means                                         |
| ---------- | ----------------------------------------------------- |
| ~100       | One server, one DB, done                              |
| ~1,000     | Still one server — add indexes and connection pooling |
| ~10,000    | Multiple app servers, read replicas or caching        |
| ~100,000   | CDN, sharding, distributed systems                    |
| ~1,000,000 | Custom everything                                     |

Most startups and internal tools never leave 100-1,000 RPS.

## Common Mistakes

- **Peak ≠ average.** Always multiply by 3-5x for peak.
- **DB queries multiply.** 100 RPS ≠ 100 DB queries. It's 300-500.
- **Reads ≠ writes.** 1,000 write QPS is much harder than 1,000 read QPS.
- **Count write transactions, not rows.** 8 writes inside one transaction = 1 tx for throughput purposes.
- **Payload size matters.** 10K RPS × 1 KB = 10 MB/sec (easy). 10K RPS × 1 MB = 10 GB/sec (CDN required).
- **Fix the query, not the language.** If DB takes 50 ms and app logic takes 1 ms, rewriting in Go saves nothing.
