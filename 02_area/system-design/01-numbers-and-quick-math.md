# Numbers and Quick Math

System design estimation isn't about getting exact numbers. It's about getting within an order of magnitude so you can make decisions: do I need a cache? Will one database handle this? Should this be async? You need a small set of memorised numbers and a few formulas to get there in two minutes.

## The Three Speed Worlds

Everything in computing lives in one of three speed worlds: **nanoseconds** (CPU, memory and syscalls), **microseconds** (SSD, context switches), and **milliseconds** (network, disk). The jump between each world is roughly 1,000x. Once you internalise this, you can reason about any system: if the operation touches the network, it's milliseconds. If it's in memory, it's nanoseconds. If it's on disk, somewhere in between.

These are approximate, but stable across years of hardware. Memorise the order of magnitude, not the exact value.

| Operation                             | Latency    | Notes                             |
| ------------------------------------- | ---------- | --------------------------------- |
| L1 cache reference                    | 1 ns       | CPU-local, fastest memory         |
| L2 cache reference                    | 4 ns       |                                   |
| L3 cache reference                    | 10 ns      | Shared across cores               |
| Mutex lock/unlock                     | 25 ns      | Uncontended                       |
| RAM access                            | 100 ns     | Main memory                       |
| Syscall                               | 100-300 ns | User space to kernel and back     |
| Context switch (OS thread)            | 1-10 us    | Save/restore registers, flush TLB |
| SSD random read                       | 100 us     | 1,000x slower than RAM            |
| SSD sequential read (1 MB)            | 1 ms       |                                   |
| Network round trip (same data centre) | 0.5 ms     | Same region                       |
| HDD random read                       | 5-10 ms    | Seek time dominates               |
| HDD sequential read (1 MB)            | 5-10 ms    |                                   |
| Network round trip (cross-region)     | 50-100 ms  | EU to US, depends on distance     |
| Network round trip (cross-ocean)      | 150-300 ms | EU to Asia, worst case            |

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
HDD seq 1 MB  ·           ·           ·            ·           ·           ·         ███████████    ·           ·
Network (US)  ·           ·           ·            ·           ·           ·            ·        ███████████████████
Network (Asia)·           ·           ·            ·           ·           ·            ·           ·  █████████████████
              │           │           │            │           │           │            │           │           │
              1 ns        10 ns       100 ns       1 us        10 us       100 us       1 ms        10 ms       100 ms
            ◄──── CPU-bound work ─────►        ◄──── SSD/local ─────►          ◄───── network/disk ─────►
```

The gap between "in-memory" and "over the network" is 4-5 orders of magnitude. This is why caching works. To feel the gap viscerally: if RAM access took 1 second, an SSD read would take 16 minutes, a same-DC network hop would take 1.4 hours, and a cross-region call would take 5.7 days.

## Operation Latencies

The hardware numbers explain _why_ things are fast or slow. But when designing a system, you think in terms of Redis calls, database queries, and API calls — not L1 cache hits.

The answer is almost always dominated by the network. A Redis GET is 0.5-1 ms not because Redis is slow (the lookup itself takes microseconds) but because the request has to travel over the network and back. A Postgres query is 1-5 ms because it's network + query planning + disk/cache access. An external API call is 50-500 ms because the internet is slow. The operation itself is rarely the bottleneck — the journey to and from the operation is.

**The simple latency model: cache ~1 ms, database read ~1-10 ms, database write ~2-10 ms, external API ~50-500 ms.**

| Operation                            | Latency     | Why                                             |
| ------------------------------------ | ----------- | ----------------------------------------------- |
| **In-process**                       |             |                                                 |
| gzip compress 1 KB                   | 3-10 us     | CPU-bound                                       |
| JSON serialise/deserialise 1 KB      | 5-50 us     | CPU-bound, language-dependent                   |
| JSON serialise/deserialise 1 MB      | 5-50 ms     | Can dominate if you're not careful              |
| Bcrypt hash (cost factor 10)         | ~100 ms     | Intentionally slow                              |
| **Cache**                            |             |                                                 |
| Redis/Memcached GET or SET           | 0.5-1 ms    | Almost entirely network                         |
| **Database reads**                   |             |                                                 |
| PostgreSQL simple query (indexed)    | 1-5 ms      | Network + parse + plan + execute + return       |
| PostgreSQL complex join              | 10-100 ms   | Depends on data size, indexes, joins            |
| PostgreSQL full table scan (1M rows) | 100-1000 ms | Missing index — fix this                        |
| MongoDB find by \_id                 | 1-5 ms      | Similar to indexed Postgres query               |
| MongoDB query (indexed)              | 2-20 ms     | Depends on index selectivity and doc size       |
| DynamoDB GetItem                     | 5-10 ms     | Single-digit ms promise + network               |
| DynamoDB Query (index, 100 items)    | 10-30 ms    | Pagination helps, filter adds cost              |
| Elasticsearch query                  | 5-20 ms     | Depends on index size and query complexity      |
| **Database writes**                  |             |                                                 |
| PostgreSQL single INSERT             | 2-10 ms     | Parse + WAL write + index update + fsync        |
| PostgreSQL batch (1000 rows)         | 20-100 ms   | ~10x faster per row than individual inserts     |
| PostgreSQL UPDATE/DELETE (indexed)   | 2-10 ms     | Finds row by index, updates, writes WAL         |
| PostgreSQL transaction (3-5 stmts)   | 10-30 ms    | Latency adds up, lock contention makes it worse |
| MongoDB insertOne                    | 2-10 ms     | Journal write + index update                    |
| MongoDB bulkWrite (1000 docs)        | 20-100 ms   | Same batching benefit as Postgres               |
| DynamoDB PutItem                     | 5-15 ms     | Replication cost                                |
| DynamoDB BatchWriteItem (25 items)   | 15-50 ms    | Max 25 items per batch, parallel internally     |
| Elasticsearch index (single doc)     | 5-20 ms     | Indexed on next refresh (default 1 sec)         |
| **Queues**                           |             |                                                 |
| Kafka produce (acks=1)               | 2-5 ms      | Broker writes to local log                      |
| Kafka produce (acks=all)             | 5-30 ms     | Waits for all replicas, safest                  |
| SQS SendMessage                      | 10-30 ms    | HTTP API call to AWS                            |
| **Network**                          |             |                                                 |
| HTTP call (same data centre)         | 5-50 ms     | Serialisation + network + processing + response |
| HTTP call (external)                 | 50-500 ms   | Internet latency + processing                   |
| DNS lookup (uncached)                | 20-100 ms   | Usually cached after first lookup               |
| TLS handshake                        | 2-50 ms     | 1-2 round trips, CPU for crypto                 |
| Send 1 MB over 1 Gbps                | ~10 ms      | Theoretical, actual is higher                   |

```
                1 us       10 us      100 us      1 ms       10 ms      100 ms      1 s
                │          │          │           │          │          │           │
  READS
  Redis/Memcached          ·          ·           ├──█──┤    ·          ·           ·
  Postgres (indexed)       ·          ·           · ├──██──┤ ·          ·           ·
  Postgres (complex join)  ·          ·           · ·  ├────████████──┤ ·           ·
  Postgres (full scan!)    ·          ·           · ·        ·├────────████████──┤  ·
  MongoDB (by _id)         ·          ·           · ├──██──┤ ·          ·           ·
  MongoDB (indexed query)  ·          ·           · · ├─██████──┤       ·           ·
  DynamoDB GetItem         ·          ·           · ·  ├██──┤·          ·           ·
  DynamoDB Query (100)     ·          ·           · ·     ├──██████──┤  ·           ·
  Elasticsearch            ·          ·           · · ├─███──┤          ·           ·
                │          │          │           │          │          │           │
  WRITES
  Redis/Memcached          ·          ·           ├──█──┤    ·          ·           ·
  Postgres INSERT          ·          ·           · · ├─██──┤·          ·           ·
  Postgres UPDATE/DELETE   ·          ·           · · ├─██──┤·          ·           ·
  Postgres batch (1K rows) ·          ·           · ·       ├████████──┤·           ·
  Postgres transaction     ·          ·           · ·     ├──██████──┤  ·           ·
  MongoDB insertOne        ·          ·           · · ├─██──┤·          ·           ·
  MongoDB bulk (1K docs)   ·          ·           · ·       ├████████──┤·           ·
  DynamoDB PutItem         ·          ·           · ·  ├████──┤         ·           ·
  DynamoDB BatchWrite (25) ·          ·           · ·      ├──████████──┤           ·
  Elasticsearch index      ·          ·           · · ├─███──┤          ·           ·
  Kafka (acks=1)           ·          ·           · · ├██──┤ ·          ·           ·
  Kafka (acks=all)         ·          ·           · · ├─████████──┤     ·           ·
  SQS SendMessage          ·          ·           · ·     ├──██████──┤  ·           ·
                │          │          │           │          │          │           │
  PROCESSING
  gzip 1 KB     ├──█──┤    ·          ·           ·          ·          ·           ·
  JSON 1 KB     · ├──██──┤ ·          ·           ·          ·          ·           ·
  JSON 1 MB                ·          ·           · · ├─████████──┤     ·           ·
  Bcrypt hash              ·          ·           · ·          ·     ├──█──┤        ·
  HTTP (same DC)           ·          ·           · · ├─████████──┤     ·           ·
  HTTP (external)          ·          ·           · ·          ├────████████████──┤ ·
                │          │          │           │          │          │           │
                1 us       10 us      100 us      1 ms       10 ms      100 ms      1 s

  Pattern: cache (1 ms) → database (2-10 ms) → external call (50-500 ms)
           ◄── do more ──────────────────────────── do less of these ──►
```

The takeaway: everything between your cache and your database is 2-10 ms. Everything involving the network outside your data centre is 50+ ms. Design to minimise the slow stuff, not speed up the fast stuff.

### Where time goes in a typical API request

To make this concrete, here's a "get user by ID" call broken down:

```
Client                                              Server
  │                                                    │
  │──── DNS lookup (cached: 0 ms, uncached: ~50 ms) ───│
  │──── TCP handshake ────────── 0.5 ms (data centre) ─│
  │──── TLS handshake ──────────────── 2-5 ms ─────────│
  │──── HTTP request flies over wire── 0.2 ms ─────────│
  │                                    │               │
  │                          Load balancer: 0.1 ms     │
  │                          App server:               │
  │                            Parse request: 0.05 ms  │
  │                            Auth check: 0.5-2 ms    │
  │                            DB query: 1-5 ms  ←── usually the bottleneck
  │                            Serialise JSON: 0.1 ms  │
  │                          Load balancer: 0.1 ms     │
  │                                    │               │
  │←── HTTP response ──────────────── 0.2 ms ──────────│
  │                                                    │
  Total: 5-15 ms (same data centre), 60-150 ms (cross-region)
```

The database query almost always dominates. This is why the first optimisation is usually: add an index, then add a cache. Not "rewrite in a faster language."

## Size Numbers

Estimating storage is about knowing what a "typical" piece of data looks like. A JSON API response is 1-10 KB. A database row is 100-500 bytes. A photo is 2-5 MB. Once you know the unit size, everything is multiplication — the hard part is knowing the unit, not the math.

The key mental anchor: **1 million rows at 500 bytes each = 500 MB**. That's small. Most people overestimate how much storage they need. A database with 10 million users and their profiles is ~20 GB — it fits in RAM on a single machine. Storage only becomes a problem with binary blobs (photos, video) or very high write rates over long retention periods.

| Data type           | Size          | Notes                        |
| ------------------- | ------------- | ---------------------------- |
| UUID                | 16 bytes      | 128 bits, 36 chars as string |
| Timestamp (Unix)    | 8 bytes       | int64                        |
| IPv4 / IPv6 address | 4 / 16 bytes  | 32 / 128 bits                |
| Short text (tweet)  | 200-500 bytes | UTF-8                        |
| JSON API response   | 1-10 KB       | Typical REST endpoint        |
| Log line            | 200-500 bytes | Structured JSON              |
| Average DB row      | 100-500 bytes | Depends on schema            |
| Web page (full)     | 2-5 MB        | HTML + CSS + JS + images     |
| Photo (JPEG)        | 2-5 MB        | Reasonable quality           |
| 1-min video (720p)  | 10-30 MB      | Compressed, depends on codec |

The scaling shortcuts to memorise:

```
1 million rows  x 500 bytes = 500 MB
1 billion rows  x 500 bytes = 500 GB   <- now you're thinking about sharding
1 million users x 1 KB each = 1 GB
1 million       x 1 MB each = 1 TB     <- photos, documents
```

## Throughput Numbers

Latency tells you how fast one operation is. Throughput tells you how many operations a system can handle per second. They're related but different — a database query might take 5 ms (latency), but the database can handle 10,000 of them per second (throughput) because it processes many queries concurrently.

### The simple model

The same 10x jumps from the latency tiers show up in throughput:

```
DB writes    ~1,000/sec    (10^3)
DB reads     ~10,000/sec   (10^4)   — 10x writes
Cache ops    ~100,000/sec  (10^5)   — 10x reads
```

**DB writes ~1K, DB reads ~10K, cache ~100K. Each tier is 10x the previous.** This is the only throughput fact you need for most estimations.

Why does this work? A DB read takes ~1-5 ms. With ~10 concurrent connections per core on a 4-core instance, that's ~40-80 connections x ~200 reads/sec each = ~10,000 reads/sec. Throughput falls out of latency and concurrency — the faster the operation, the more you can do per second.

When to worry:

| Threshold             | What it signals               | First move                     |
| --------------------- | ----------------------------- | ------------------------------ |
| Approaching 1K writes | WAL pressure, lock contention | Connection pooling (PgBouncer) |
| Approaching 10K reads | Database working hard         | Read replicas or cache         |
| Beyond 100K ops       | Past single-node cache limit  | Redis Cluster or shard         |

### The hierarchy

**Databases** sit at the bottom: slowest, because every operation potentially touches disk, acquires locks, and maintains consistency guarantees. Writes are slower than reads because they must update the WAL/journal, modify indexes, and wait for replication. **App servers** sit in the middle: they mostly shuffle data between clients and databases. **Caches and queues** sit at the top: in-memory (Redis, Memcached) or sequential I/O (Kafka appending to a log).

This is why the database is almost always the bottleneck. Your app server can handle 30,000 req/sec, but if each request makes 3 database queries, your Postgres instance only needs to hit 10,000 QPS before it's the ceiling.

### Detailed reference

These are ballpark numbers for a single node with decent hardware (4-8 cores, 16-32 GB RAM, SSD):

| System                | Throughput               | Notes                                      |
| --------------------- | ------------------------ | ------------------------------------------ |
| **Databases**         |                          |                                            |
| Postgres/MySQL reads  | 5,000-20,000 queries/sec | Simple queries, indexed, connection pooled |
| Postgres/MySQL writes | 1,000-10,000 inserts/sec | Depends on indexes, WAL, fsync             |
| MongoDB reads         | 10,000-50,000 ops/sec    | Depends on working set in RAM              |
| MongoDB writes        | 5,000-25,000 inserts/sec | Journal + index update                     |
| DynamoDB              | Unlimited (provisioned)  | Pay per RCU/WCU, scales horizontally       |
| Elasticsearch         | 5,000-20,000 ops/sec     | Reads and writes in similar range          |
| **Caches**            |                          |                                            |
| Redis                 | 80,000-100,000 ops/sec   | In-memory, single-threaded                 |
| Memcached             | 100,000-500,000 ops/sec  | In-memory, multi-threaded                  |
| **Queues**            |                          |                                            |
| Kafka (single broker) | 100,000-500,000 msgs/sec | Depends on message size, replication       |
| SQS (standard)        | ~unlimited               | FIFO: 300/sec, 3,000/sec with batching     |
| **App servers**       |                          |                                            |
| Node.js               | 10,000-30,000 req/sec    | Simple JSON API, no heavy computation      |
| Go                    | 30,000-100,000 req/sec   | Simple JSON API                            |
| Java (Spring/Netty)   | 20,000-80,000 req/sec    | Warmed up                                  |
| Nginx (proxy)         | 50,000-100,000 req/sec   | No heavy processing                        |

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
  DynamoDB          ·          ·          ·     <- unlimited (provisioned) ->
                    │          │          │          │          │          │
  CACHES
  Redis             ·          ·          ·          ·  ├──█████████████──┤
  Memcached         ·          ·          ·          ·     ├────██████████████████──>
                    │          │          │          │          │          │
  QUEUES
  Kafka             ·          ·          ·          ·     ├────██████████████████──>
  SQS (standard)    ·          ·          ·     <- nearly unlimited (standard) ->
                    │          │          │          │          │          │
  APP SERVERS
  Node.js           ·          ·├──████████████████──┤          ·          ·
  Java              ·          · ├────████████████████████████████──┤      ·
  Go                ·          ·          · ├───████████████████████████████──┤
  Nginx (proxy)     ·          ·          ·          ├─████████████████──┤
                    │          │          │          │          │          │
                    1K         5K        10K        50K       100K       500K
                                     ops/sec per single node

  Pattern: database (1K-50K) -> app server (10K-100K) -> cache/queue (100K-500K+)
           <-- usually the bottleneck                    almost never the bottleneck -->
```

Most web applications don't need to worry about app server throughput. If you have 100,000 users making 100 requests/day, that's 10 million/day = ~100 RPS. A single app server handles this with 99% of its capacity idle. The bottleneck is almost always the database.

### How instance size shifts these numbers

The throughput numbers above assume a mid-range instance (~4-8 vCPU, 16-32 GB RAM). This matters because "Postgres handles 10,000 QPS" is wrong if you're running on a `db.t3.micro` — that tops out around 200-500 QPS.

The rough scaling rule: **2x cores = 1.5-1.8x throughput** for reads, less for writes. Here's the full range for Postgres on RDS:

| Instance class  | vCPU | RAM    | Reads/sec       | Writes/sec    | Typical use               |
| --------------- | ---- | ------ | --------------- | ------------- | ------------------------- |
| db.t3.micro     | 2    | 1 GB   | 200-500         | 50-200        | Dev/test only             |
| db.t3.medium    | 2    | 4 GB   | 500-2,000       | 200-1,000     | Light prod, side projects |
| db.r6g.large    | 2    | 16 GB  | 2,000-5,000     | 500-3,000     | Small production          |
| db.r6g.xlarge   | 4    | 32 GB  | 5,000-15,000    | 1,000-8,000   | Medium production         |
| db.r6g.4xlarge  | 16   | 128 GB | 15,000-50,000   | 5,000-20,000  | Large production          |
| db.r6g.16xlarge | 64   | 512 GB | 50,000-100,000+ | 15,000-50,000 | Before sharding           |

```
                  db.t3.micro    db.r6g.large   db.r6g.xlarge   db.r6g.4xlarge   db.r6g.16xlarge
                  (dev/test)     (small prod)   (medium prod)   (large prod)     (before sharding)
                  │              │              │               │                │
  DB reads/sec    200            2,000          5,000-15,000    15,000-50,000    50,000-100,000+
  DB writes/sec   50             500            1,000-8,000     5,000-20,000     15,000-50,000
                  │              │              │               │                │
                  <-- costs      │     <-- the "simple model"   │               -- costs $$$$ -->
                  ~$15/mo -->    │         numbers live here    │
                                 │                              │
                           <-- most startups -->    <-- scale-up before sharding -->
```

The simple model numbers (1K writes, 10K reads) map to `db.r6g.xlarge` — a typical production instance at $200-400/month. Most startups run on `db.r6g.large` to `db.r6g.xlarge` and never need anything bigger. MySQL and MongoDB follow similar scaling curves at similar instance sizes.

For other systems: **Redis** scales with bigger instances mainly for more RAM (larger working set), not throughput — it's single-threaded, so you shard via Redis Cluster. A `cache.r6g.large` (13 GB) does ~100K ops/sec. **App servers** scale linearly with vCPU: a 4 vCPU ECS task handles ~10K-25K req/sec (Node.js) or ~30K-80K req/sec (Go). **Lambda** is different entirely — throughput scales by running more instances in parallel (up to 1,000-3,000 concurrent), not by making one faster. A single Lambda handles one request at a time, so 1,000 RPS means 1,000 concurrent Lambdas if each takes 1 second, or 100 if each takes 100 ms.

## The Quick Math

The numbers above are your building blocks. Now you need to go from a product requirement ("100,000 users placing orders") to a system requirement ("how many database queries per second?"). The chain is always the same:

```
Users -> actions per day -> requests per second -> DB queries per second
                                                -> storage per year
                                                -> bandwidth
                                                -> cache size
```

Every estimation starts with "how many requests per second?" and everything else follows.

### The one rule: 1 day = 10^5 seconds

Forget 86,400. These rounded numbers are close enough:

| Time period | Quick estimate     | Actual         | Error | Trick                     |
| ----------- | ------------------ | -------------- | ----- | ------------------------- |
| 1 day       | 100,000 sec (10^5) | 86,400 sec     | ~15%  | The anchor — drop 5 zeros |
| 1 hour      | 4,000 sec          | 3,600 sec      | ~10%  | Drop 3 zeros, x4          |
| 1 month     | 2.5 million sec    | 2,592,000 sec  | ~3%   | day x 25                  |
| 1 year      | 30 million sec     | 31,536,000 sec | ~5%   | day x 300                 |

**To go from "per day" to "per second", drop 5 zeros.** 10 million actions per day? That's 100 per second. The quick number is always ~14% lower than exact — this never changes a design decision.

### Users to requests per second

```
RPS = (users x actions_per_day) / 10^5       <- just drop 5 zeros
```

Traffic isn't evenly distributed — peak is usually 2-5x average. Multiply by 3 as a safe default:

| Users      | Actions/day | Total/day | Avg RPS | Peak (x3) |
| ---------- | ----------- | --------- | ------- | --------- |
| 100,000    | 10          | 1M        | 10      | 30        |
| 100,000    | 100         | 10M       | 100     | 300       |
| 1,000,000  | 50          | 50M       | 500     | 1,500     |
| 10,000,000 | 20          | 200M      | 2,000   | 6,000     |
| 10,000,000 | 100         | 1B        | 10,000  | 30,000    |

### Requests per second to database queries

This is where most people underestimate. A "simple" endpoint that shows an order typically involves: check auth (1 query), fetch the order (1), fetch order items (1), fetch shipping status (1), maybe log the access (1 write). That's 5 queries for what feels like one request.

```
DB QPS = RPS x queries_per_request

300 peak RPS x 3 queries = 900 DB QPS   <- comfortable for a single Postgres
```

### Storage

```
Storage = users x data_per_user x retention_period
```

| Scenario                    | Calculation                        | Per year | Where?         |
| --------------------------- | ---------------------------------- | -------- | -------------- |
| User profiles (1M users)    | 1M x 2 KB                          | 2 GB     | Database       |
| Activity logs (1M users)    | 1M x 50 events/day x 500 bytes     | 9 GB     | Database       |
| Orders (1M users, 2/day)    | 1M x 2 x 1 KB x 365                | 730 GB   | Database       |
| Chat messages (1M users)    | 1M x 20 msgs/day x 500 bytes x 365 | 3.6 TB   | DB + archive   |
| Photos (1M users, 2/week)   | 1M x 2 x 3 MB x 52                 | 312 TB   | Object storage |
| Video (100K creators, 1/wk) | 100K x 500 MB x 52                 | 2.6 PB   | Object storage |

Rule of thumb: anything over ~500 GB/year of blobs goes to S3 (~$23/TB/month).

### Bandwidth

```
Bandwidth = RPS x average_response_size
```

| Scenario                      | Calculation     | Bandwidth | Verdict                |
| ----------------------------- | --------------- | --------- | ---------------------- |
| Small API (300 RPS)           | 300 x 5 KB      | 12 Mbps   | One server, trivial    |
| Medium API (5,000 RPS)        | 5,000 x 10 KB   | 400 Mbps  | Fine on 1 Gbps link    |
| Large API (50,000 RPS)        | 50,000 x 50 KB  | 20 Gbps   | Multiple servers + CDN |
| Image serving (10,000 RPS)    | 10,000 x 500 KB | 40 Gbps   | CDN required           |
| Video streaming (1,000 users) | 1,000 x 5 Mbps  | 5 Gbps    | CDN required           |

Rule of thumb: 1 Gbps = 100,000 RPS of 1 KB responses, or 1,000 RPS of 1 MB responses.

### Cache sizing

```
Cache size = number_of_hot_items x item_size
```

| Scenario               | Calculation | Size   | Fits in                   |
| ---------------------- | ----------- | ------ | ------------------------- |
| User sessions (100K)   | 100K x 1 KB | 100 MB | Single Redis, trivial     |
| Product listings (1M)  | 1M x 5 KB   | 5 GB   | Single Redis, comfortable |
| Product listings (10M) | 10M x 5 KB  | 50 GB  | Redis Cluster             |

Redis single instance comfortably holds 10-25 GB. Above that, cluster or shard. Watch fork latency on snapshots above ~25 GB.

## Worked Example: E-Commerce Order System

This is how you'd walk through an estimation in an interview or design discussion. The goal isn't precise numbers — it's making decisions.

**Given:** 100,000 users, each placing ~2 orders/day, browsing ~50 pages/day.

**Traffic:**

```
Browsing:  100,000 x 50 = 5 million/day -> drop 5 zeros -> ~50 RPS avg -> ~150 peak
Ordering:  100,000 x 2  = 200,000/day   -> drop 5 zeros -> ~2 RPS avg  -> ~6 peak
Total:     ~156 peak RPS
```

Trivial for any modern app server.

**Database load:**

```
Browse: 150 peak RPS x 3 queries each = 450 DB QPS
Orders: 6 peak RPS x 5 queries each   = 30 DB QPS (but writes!)
Total:  ~480 DB QPS
```

A single Postgres instance handles this without breaking a sweat.

**Storage:**

```
Orders:   100,000 x 2/day x 365 x 1 KB = ~73 GB/year
Products: 500,000 x 5 KB                = 2.5 GB (static)
Users:    100,000 x 2 KB                = 200 MB
Total:    ~75 GB first year — fits on a single database
```

**Caching:**

```
Hot products (top 10K): 10,000 x 5 KB = 50 MB  -> Redis, trivial
User sessions: 100,000 x 1 KB         = 100 MB -> Redis, trivial
```

You might not even need a cache at this scale. But if product pages involve complex joins, caching the top 10,000 eliminates most reads — access follows a power law (top 20% gets 80% of views).

**Verdict:**

```
~156 peak RPS, ~480 DB QPS, 75 GB storage/year

One app server, one database, maybe Redis for sessions. Don't over-engineer. Ship it.
```

**What if it grows 100x?**

```
10,000,000 users:
  ~15,000 peak RPS -> multiple app servers behind a load balancer
  ~48,000 DB QPS   -> read replicas + caching + connection pooling
  7.3 TB/year      -> still one DB, but archive old orders
```

Now you need infrastructure. But you'd know because you'd watch the metrics grow — you don't build for this on day one.

## Rules of Thumb and Common Mistakes

Sometimes you don't need the formulas. If someone says "we have 50,000 users" you should immediately have a gut reaction:

1. **100 RPS** — one server, one database, go home early
2. **1,000 RPS** — still one server, indexes and connection pooling
3. **10,000 RPS** — multiple app servers, read replicas or caching, monitoring
4. **100,000 RPS** — distributed systems, CDN, sharding, dedicated team
5. **1,000,000 RPS** — FAANG scale, everything is custom

Most startups and internal tools never leave category 1 or 2.

The math is easy — what trips people up is the assumptions:

- **Peak != average.** Traffic follows patterns. Always multiply by 3-5x for peak.
- **Database queries multiply.** 100 RPS != 100 DB queries. Auth, main query, related data, logging — it's 3-5 queries per request.
- **Reads != writes.** A database handles 10x more reads than writes. 1,000 write QPS is much harder than 1,000 read QPS.
- **Fix the query, not the language.** If your DB query takes 50 ms and your app logic takes 1 ms, rewriting in Go saves you nothing. Add an index.
- **Payload size matters.** 10,000 RPS of 1 KB responses is 10 MB/sec (easy). 10,000 RPS of 1 MB responses is 10 GB/sec (CDN time).

## Key Takeaways

**1. Three speed worlds.** Nanoseconds (CPU/memory), microseconds (SSD/local), milliseconds (network/disk). The jumps are 1,000x. This tells you what's worth optimising — shaving 0.5 ms off app logic doesn't matter if the database takes 5 ms and the network takes 50 ms.

**2. The estimation chain.** Users -> actions/day -> drop 5 zeros -> RPS -> multiply by queries per request -> DB QPS. Compare to what your instance handles. If it fits, stop designing. If not: cache hot reads -> read replicas -> sharding.

**3. Most systems are smaller than you think.** 100,000 users is ~100-500 RPS. One server, one database. The instinct to build for scale before you have scale is the most expensive engineering mistake. Two minutes of estimation prevents weeks of over-engineering.

## What's Next

Now that you can estimate traffic, storage, and throughput, the next lesson traces a single request through every layer of the stack — DNS, load balancer, app server, database, response — and identifies where latency hides and what you can do about it at each stage.
