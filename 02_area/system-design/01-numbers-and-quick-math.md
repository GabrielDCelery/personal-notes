# Numbers and Quick Math — Distilled

Three memorised numbers and one formula let you answer any capacity question in two minutes — and most systems are far smaller than you think.

## Core Mental Model: Three Speed Worlds

Everything in computing lives in one of three worlds: **nanoseconds** (CPU and memory), **microseconds** (SSD and local I/O), **milliseconds** (network and disk). The jump between each world is ~1,000×. Once this is internalised, you can classify any operation instantly: touches the network → milliseconds. In memory → nanoseconds.

```
              1 ns        10 ns       100 ns       1 us        10 us       100 us       1 ms        10 ms       100 ms
              │           │           │            │           │           │            │           │           │
L1 cache      █           ·           ·            ·           ·           ·            ·           ·           ·
L2 cache      █████       ·           ·            ·           ·           ·            ·           ·           ·
L3 cache      ████████████·           ·            ·           ·           ·            ·           ·           ·
Mutex lock    ····████████████████    ·            ·           ·           ·            ·           ·           ·
RAM           ·           ·███████████·            ·           ·           ·            ·           ·           ·
Syscall       ·           ·███████████████████     ·           ·           ·            ·           ·           ·
SSD random    ·           ·           ·            ·███████████████████████·            ·           ·           ·
SSD seq 1 MB  ·           ·           ·            ·           ·           ·████████████·           ·           ·
Network (DC)  ·           ·           ·            ·           ·        ███████████     ·           ·           ·
HDD random    ·           ·           ·            ·           ·           ·         ███████████████████        ·
Network (US)  ·           ·           ·            ·           ·           ·            ·        ███████████████████
Network (Asia)·           ·           ·            ·           ·           ·            ·           ·  █████████████████
              │           │           │            │           │           │            │           │           │
              1 ns        10 ns       100 ns       1 us        10 us       100 us       1 ms        10 ms       100 ms
            ◄──── CPU-bound work ─────►        ◄──── SSD/local ─────►          ◄───── network/disk ─────►
```

The visceral version: if RAM access took 1 second, an SSD read would take 16 minutes, a same-DC network hop would take 1.4 hours, and a cross-region call would take 5.7 days. This is why caching works, and why optimising in-memory code is almost never the right target.

## Operation Latencies

The practical translation of the three worlds — mapped to Redis, databases, and HTTP calls. **The operation itself is rarely the bottleneck; the network journey to and from it is.** A Redis GET takes microseconds internally; the 1 ms cost is the round trip.

```
                            100 us      1 ms        10 ms       100 ms      1 s         10 s
                            │           │           │           │           │           │
  READS
  Redis/Memcached           ·       ├███████┤       ·           ·           ·           ·
  Postgres (indexed)        ·           ·   ├███████┤           ·           ·           ·
  Postgres (complex join)   ·           ·       ├███████████┤   ·           ·           ·
  Postgres (full scan!)     ·           ·           ·       ├███████████████┤           ·
  MongoDB (by _id)          ·           ·   ├███████┤           ·           ·           ·
  DynamoDB GetItem          ·           ·     ├█████┤           ·           ·           ·
  Elasticsearch             ·           ·     ├█████████┤       ·           ·           ·
                            │           │           │           │           │           │
  WRITES
  Postgres INSERT           ·           ·   ├███████┤           ·           ·           ·
  Postgres batch (1K rows)  ·           ·           ·   ├███████┤           ·           ·
  Postgres transaction      ·           ·           ├███████┤   ·           ·           ·
  DynamoDB PutItem          ·           ·       ├███████┤       ·           ·           ·
  Kafka (acks=1)            ·           ·   ├███┤   ·           ·           ·           ·
  Kafka (acks=all)          ·           ·       ├█████████┤     ·           ·           ·
  SQS SendMessage           ·           ·           ├███████┤   ·           ·           ·
                            │           │           │           │           │           │
  NETWORK
  HTTP (same DC)            ·           ├███████████┤           ·           ·           ·
  HTTP (external)           ·           ·           ·       ├███████████┤   ·           ·
  DNS lookup (uncached)     ·           ·           ├███████████┤           ·           ·
                            │           │           │           │           │           │
  CPU / COMPUTE
  JSON parse (1 MB)         ·           ·       ├███████████┤   ·           ·           ·
  ML inference              ·           ·           ├███████████████████┤   ·           ·
  Bcrypt (cost 10)          ·           ·           ·          ├███┤        ·           ·
  Image resize              ·           ·           ·       ├███████████┤   ·           ·
  PDF generation            ·           ·           ·           ├███████████████┤       ·
                            │           │           │           │           │           │
                            100 us      1 ms        10 ms       100 ms      1 s         10 s

  Pattern: cache (1 ms) → database (2–10 ms) → external call (50–500 ms) → CPU-bound (100 ms–10 s)
           ◄── use more ────────────────────────────────────── offload or make async ───────────────►
```

**The model: cache ~1 ms, database ~1–10 ms, external API ~50–500 ms.**

## Two Bottleneck Types

Databases and external APIs are both slow, but they fail differently.

**Databases (1–10 ms/call, many calls/request):** the problem shows up as throughput. Thousands of concurrent requests saturate connection pools and the DB starts queuing. Fix: indexes → connection pooling → caching hot reads → read replicas.

**External APIs (50–500 ms/call, few calls/request):** the problem shows up as latency. A single request sits waiting half a second; threads are blocked doing nothing. Fix: make calls async, parallelise independent calls, add timeouts and circuit breakers.

The instinct is to optimise the database first — and that's usually right. But if a request includes an external API call, those 200 ms dwarf everything else. No amount of DB tuning helps. Check which one actually dominates before optimising.

## Throughput Model: 1K / 10K / 100K

The 1,000× tiers from latency compress to 10× steps in throughput, because concurrency partially fills the gap:

```
                    1K         5K        10K        50K       100K       500K
                    │          │          │          │          │          │
  DATABASES (reads)
  Postgres/MySQL    · ├────████████████──┤           ·          ·          ·
  MongoDB           ·          · ├───████████████████████──┤    ·          ·
  DynamoDB          ·          ·          ·     <- unlimited (provisioned) ->
                    │          │          │          │          │          │
  DATABASES (writes)
  Postgres/MySQL    ├─████████████──┤     ·          ·          ·          ·
  MongoDB           · ├────████████████████──┤       ·          ·          ·
                    │          │          │          │          │          │
  CACHES
  Redis             ·          ·          ·          ·  ├──█████████████──┤
  Memcached         ·          ·          ·          ·     ├────██████████████████──>
                    │          │          │          │          │          │
  APP SERVERS
  Node.js           ·          ·├──████████████████──┤          ·          ·
  Go                ·          ·          · ├───████████████████████████████──┤
  Nginx (proxy)     ·          ·          ·          ├─████████████████──┤
                    │          │          │          │          │          │
                    1K         5K        10K        50K       100K       500K
                                     ops/sec per single node

  Pattern: database (1K–50K) → app server (10K–100K) → cache/queue (100K–500K+)
           <-- usually the bottleneck                    almost never the bottleneck -->
```

**DB writes ~1K/sec, DB reads ~10K/sec, cache ops ~100K/sec. Each tier is 10×.** This is the only throughput figure you need for most estimates.

Why: a DB read takes ~1–5 ms. With ~40 connections on a 4-core instance each doing ~200 reads/sec, you get ~10K reads/sec. Throughput = latency × concurrency.

| Threshold             | What it signals               | First move                     |
| --------------------- | ----------------------------- | ------------------------------ |
| Approaching 1K writes | WAL pressure, lock contention | Connection pooling (PgBouncer) |
| Approaching 10K reads | DB working hard               | Read replicas or cache         |
| Beyond 100K ops       | Past single-node cache limit  | Redis Cluster or shard         |

These numbers assume ~4 vCPU, 16–32 GB RAM, SSD (`db.r6g.xlarge`, ~$200–400/month). A dev `db.t3.micro` caps at 200–500 QPS. Most startups run `db.r6g.large` to `db.r6g.xlarge` and never need anything bigger.

## The Estimation Chain

Every capacity question follows the same chain:

```
Users → actions/day → ÷ 10^5 → avg RPS × 3 → peak RPS
                                             → × queries/request → DB QPS
                                             → × response size   → bandwidth
                                             → × hot item size   → cache size
```

### Rule 1: Time conversions — drop 5 zeros

Forget exact seconds. Round everything:

| Period  | Estimate       | Actual         | Error | Trick                         |
| ------- | -------------- | -------------- | ----- | ----------------------------- |
| 1 hour  | 4,000 sec      | 3,600 sec      | ~10%  | Drop 3 zeros, ×4              |
| 1 day   | 100,000 sec    | 86,400 sec     | ~15%  | **Drop 5 zeros — the anchor** |
| 1 month | 3,000,000 sec  | 2,592,000 sec  | ~16%  | 30 days × 100K                |
| 1 year  | 30,000,000 sec | 31,536,000 sec | ~5%   | 300 days × 100K               |

**To go from "per day" to "per second": drop 5 zeros.** Month and year follow naturally — just multiply up. The errors are consistent enough that they never change a design decision.

Traffic peaks at 2–5× average. **Use 3× as the default peak multiplier.**

### Rule 2: Count transactions, not rows

A read request touches 3–5 queries. A write request is **one transaction** (BEGIN...COMMIT), regardless of how many rows it modifies. The DB's write limit is concurrent transactions, not individual row operations.

- **Read endpoints:** peak RPS × 3–5 = read QPS
- **Write endpoints:** peak RPS × 1 = write tx/sec (plus 1–2 pre-transaction reads)

Counting every row in a transaction as a separate write is the most common estimation mistake — it makes writes look 5–10× worse than they are.

### Size and bandwidth anchors

```
1M rows  × 500 B  = 500 MB     fits easily in RAM
1B rows  × 500 B  = 500 GB     now you're thinking about sharding
1M users × 1 KB   = 1 GB       trivial
1M items × 1 MB   = 1 TB       photos/docs → object storage (S3)

Bandwidth = RPS × avg response size
1 Gbps ≈ 100K RPS of 1 KB responses, or 1K RPS of 1 MB responses
```

Storage only becomes a problem with binary blobs or high write rates over long retention. A database with 10M users and profiles is ~20 GB — it fits in RAM on a single machine.

## Worked Example

**100K users, 100 orders/day, 10 page views per order, business hours (8 hrs).**

Traffic is concentrated in 8 hours, not spread over 24 — use 30K seconds instead of 100K:

```
Orders:   100K × 100 = 10M/day ÷ 30K = 333 avg RPS → ~1K peak
Browsing: 10M × 10   = 100M/day ÷ 30K = 3,333 avg  → ~10K peak
Total: ~11K peak RPS
```

Split reads from writes — they hit different limits:

```
Browse:          10K RPS × 4 reads = 40K read QPS
Order pre-check:  1K RPS × 1 read  =  1K read QPS
Order writes:     1K RPS × 1 tx    =  1K write tx/sec
──────────────────────────────────────────────────
Total: 41K read QPS + 1K write tx/sec
```

41K reads exceeds a single Postgres instance. **Cache the top products.** Top 20% of products gets ~80% of views. Cache 50K products (~250 MB in Redis) → absorbs ~32K reads → leaves ~9K hitting the DB. Comfortable.

Storage: 100K users × 100 orders/day × 365 × 1 KB = ~3.6 TB/year. Plan for archiving.

**Verdict:** 2–3 app servers, Postgres with PgBouncer, Redis for products + sessions. Add read replicas only if cache hit rate disappoints.

**What if traffic is spread across 24 hours instead?** Use 100K seconds: ~300 peak RPS, ~12K read QPS. One app server, no read replicas needed. The time window makes a 3× difference — always ask "when does traffic happen?" before defaulting to the day rule.

## Scale Gut-Check

Before doing any math, a rough RPS gives you an immediate category:

```
~100 RPS     → one server, one DB, go home early
~1,000 RPS   → one server, indexes and connection pooling
~10,000 RPS  → multiple servers, caching or read replicas, monitoring
~100,000 RPS → distributed systems, CDN, sharding
~1,000,000 RPS → FAANG scale, everything is custom
```

**Most startups and internal tools never leave 100–1,000 RPS.** The instinct to build for scale before you have it is the most expensive engineering mistake.

## Common Mistakes

- **Peak ≠ average.** Traffic follows patterns. Always apply 3× for peak.
- **DB queries multiply.** 100 RPS ≠ 100 DB queries. Auth + main query + related data = 3–5 per request.
- **Reads ≠ writes.** 1,000 write QPS is 10× harder than 1,000 read QPS. Always split them.
- **Transactions ≠ rows.** A transaction touching 8 rows counts as 1 write tx/sec.
- **Fix the query, not the language.** DB takes 50 ms, app logic takes 1 ms → rewriting in Go saves nothing. Add an index.
- **Payload size matters.** 10K RPS × 1 KB = 10 MB/sec (trivial). 10K RPS × 1 MB = 10 GB/sec (CDN required).

## Key Mental Models

1. **Three speed worlds: ns / μs / ms — each 1,000× apart.** Network = ms, memory = ns. Any operation touching the network is in a different league.
2. **1K / 10K / 100K.** DB writes ~1K/sec, reads ~10K/sec, cache ~100K/sec. Each tier is 10×.
3. **Drop 5 zeros.** Per-day to per-second: move decimal 5 left. Apply 3× for peak.
4. **Count transactions, not rows.** A write request = 1 tx/sec regardless of how many rows it touches.
5. **The DB is almost always the bottleneck.** App servers cap at 10K–100K RPS; DB reads cap at 10K.
6. **Most systems are smaller than you think.** 100K users ≈ 100 RPS. One server, one DB.
7. **Two minutes of estimation prevents weeks of over-engineering.**
