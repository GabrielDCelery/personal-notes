# Numbers and Quick Math

System design estimation isn't about getting exact numbers. It's about getting within an order of magnitude so you can make decisions: do I need a cache? Will one database handle this? Should this be async? You need a small set of memorised numbers and a few formulas to get there in two minutes.

## Latency Numbers Every Engineer Should Know

Everything in computing lives in one of three speed worlds: **nanoseconds** (CPU and memory), **microseconds** (SSD, local operations), and **milliseconds** (network, disk). The jump between each world is roughly 1,000x. Once you internalise this, you can reason about any system: if the operation touches the network, it's milliseconds. If it's in memory, it's nanoseconds. If it's on disk, somewhere in between.

These are approximate, but stable across years of hardware. Memorise the order of magnitude, not the exact value.

| Operation                             | Latency    | Notes                             |
| ------------------------------------- | ---------- | --------------------------------- |
| L1 cache reference                    | 1 ns       | CPU-local, fastest memory         |
| L2 cache reference                    | 4 ns       |                                   |
| L3 cache reference                    | 10 ns      | Shared across cores               |
| RAM access                            | 100 ns     | Main memory                       |
| SSD random read                       | 100 us     | 1,000x slower than RAM            |
| SSD sequential read (1 MB)            | 1 ms       |                                   |
| HDD random read                       | 5-10 ms    | Seek time dominates               |
| HDD sequential read (1 MB)            | 5-10 ms    |                                   |
| Network round trip (same data centre) | 0.5 ms     | Same region                       |
| Network round trip (cross-region)     | 50-100 ms  | EU to US, depends on distance     |
| Network round trip (cross-ocean)      | 150-300 ms | EU to Asia, worst case            |
| Mutex lock/unlock                     | 25 ns      | Uncontended                       |
| Syscall                               | 100-300 ns | User space to kernel and back     |
| Context switch (OS thread)            | 1-10 us    | Save/restore registers, flush TLB |
| Process fork                          | 1-10 ms    | Depends on memory size            |

### At a glance — hardware latency scale

Each `█` is roughly 10x. This is a logarithmic scale — the jumps are enormous.

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

The gap between "in-memory" and "over the network" is 4-5 orders of magnitude. This is why caching works — the gap between memory and everything else is enormous. If RAM is 1 second, an SSD read is 16 minutes, a same-DC network hop is 1.4 hours, and a cross-region call is 5.7 days.

## Common Operation Latencies

The hardware numbers above are useful for understanding why things are fast or slow, but you rarely think in terms of "L1 cache hit" when designing a system. What you actually care about is: how long does a Redis call take? A database query? An API call to another service?

The answer is almost always dominated by the network. A Redis GET is 0.5-1 ms not because Redis is slow (the lookup itself takes microseconds) but because the request has to travel over the network and back. A Postgres query is 1-5 ms because it's network + query planning + disk/cache access. An external API call is 50-500 ms because the internet is slow. The operation itself is rarely the bottleneck — the journey to and from the operation is.

| Operation                              | Latency     | Notes                                           |
| -------------------------------------- | ----------- | ----------------------------------------------- |
| In-process function call               | 1-50 ns     | Depends on work done                            |
| **Reads**                              |             |                                                 |
| Redis GET (same data centre)           | 0.5-1 ms    | Network + lookup, almost all network            |
| Memcached GET (same data centre)       | 0.5-1 ms    | Similar to Redis for simple ops                 |
| PostgreSQL simple query (indexed)      | 1-5 ms      | Network + parse + plan + execute + return       |
| PostgreSQL complex join                | 10-100 ms   | Depends on data size, indexes, joins            |
| PostgreSQL full table scan (1M rows)   | 100-1000 ms | Avoid in production — missing index             |
| MongoDB find by \_id                   | 1-5 ms      | Similar to indexed Postgres query               |
| MongoDB query (indexed)                | 2-20 ms     | Depends on index selectivity and doc size       |
| DynamoDB GetItem                       | 5-10 ms     | Single-digit ms is the promise, network adds    |
| DynamoDB Query (index, 100 items)      | 10-30 ms    | Pagination helps, filter adds cost              |
| Elasticsearch simple query             | 5-20 ms     | Depends on index size and query complexity      |
| **Writes**                             |             |                                                 |
| Redis SET (same data centre)           | 0.5-1 ms    | Same as GET, network-dominated                  |
| Memcached SET (same data centre)       | 0.5-1 ms    | Same as GET                                     |
| PostgreSQL single INSERT               | 2-10 ms     | Parse + WAL write + index update + fsync        |
| PostgreSQL INSERT (batched, 1000 rows) | 20-100 ms   | ~10x faster per row than individual inserts     |
| PostgreSQL UPDATE (indexed)            | 2-10 ms     | Finds row by index, updates, writes WAL         |
| PostgreSQL DELETE (indexed)            | 2-10 ms     | Similar to UPDATE, marks dead tuple             |
| PostgreSQL transaction (3-5 stmts)     | 10-30 ms    | Latency adds up, lock contention can make worse |
| MongoDB insertOne                      | 2-10 ms     | Journal write + index update                    |
| MongoDB updateOne (indexed)            | 2-10 ms     | Find by index + update + journal                |
| MongoDB bulkWrite (1000 docs)          | 20-100 ms   | Same batching benefit as Postgres               |
| DynamoDB PutItem                       | 5-15 ms     | Slightly higher than GetItem, replication cost  |
| DynamoDB BatchWriteItem (25 items)     | 15-50 ms    | Max 25 items per batch, parallel internally     |
| Elasticsearch index (single doc)       | 5-20 ms     | Indexed on next refresh (default 1 sec)         |
| Elasticsearch bulk (1000 docs)         | 50-200 ms   | Preferred way to index, amortises overhead      |
| Kafka produce (acks=1)                 | 2-5 ms      | Broker writes to local log                      |
| Kafka produce (acks=all)               | 5-30 ms     | Waits for all replicas, safest                  |
| SQS SendMessage                        | 10-30 ms    | HTTP API call to AWS                            |
| **Network & processing**               |             |                                                 |
| HTTP API call (same data centre)       | 5-50 ms     | Serialisation + network + processing + response |
| HTTP API call (external)               | 50-500 ms   | Internet latency + processing                   |
| DNS lookup (uncached)                  | 20-100 ms   | Usually cached after first lookup               |
| TLS handshake                          | 2-50 ms     | 1-2 round trips, CPU for crypto                 |
| Compress 1 KB with gzip                | 3-10 us     | CPU-bound                                       |
| JSON serialise/deserialise 1 KB        | 5-50 us     | Language-dependent                              |
| JSON serialise/deserialise 1 MB        | 5-50 ms     | Can dominate if you're not careful              |
| Bcrypt hash (cost factor 10)           | 100 ms      | Intentionally slow                              |
| Send 1 MB over 1 Gbps network          | 10 ms       | Theoretical, actual is higher                   |

### At a glance — operation latency scale

Where your application time actually goes. Reads on top, writes below, processing at the bottom.

```
                1 us       10 us      100 us      1 ms       10 ms      100 ms      1 s
                │          │          │           │          │          │           │
  READS
  Redis/Memcached GET      ·          ·           ├──█──┤    ·          ·           ·
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
  Redis/Memcached SET      ·          ·           ├──█──┤    ·          ·           ·
  Postgres INSERT          ·          ·           · · ├─██──┤·          ·           ·
  Postgres UPDATE/DELETE   ·          ·           · · ├─██──┤·          ·           ·
  Postgres batch (1K rows) ·          ·           · ·       ├████████──┤·           ·
  Postgres transaction     ·          ·           · ·     ├──██████──┤  ·           ·
  MongoDB insertOne/update ·          ·           · · ├─██──┤           ·           ·
  MongoDB bulk (1K docs)   ·          ·           · ·       ├████████──┤·           ·
  DynamoDB PutItem         ·          ·           · ·  ├████──┤         ·           ·
  DynamoDB BatchWrite (25) ·          ·           · ·      ├──████████──┤           ·
  Elasticsearch index      ·          ·           · · ├─███──┤          ·           ·
  Elasticsearch bulk (1K)  ·          ·           · ·          ├──██████████──┤     ·
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

The takeaway: everything between your cache and your database is 2-10 ms. Everything involving the network outside your DC is 50+ ms. Design to minimise the slow stuff, not speed up the fast stuff.

### Where time goes in a typical API request

A simple "get user by ID" API call:

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

| Data type                  | Size          | Notes                              |
| -------------------------- | ------------- | ---------------------------------- |
| UUID                       | 16 bytes      | 128 bits, 36 chars as string       |
| Timestamp (Unix)           | 8 bytes       | int64                              |
| IPv4 address               | 4 bytes       | 32 bits                            |
| IPv6 address               | 16 bytes      | 128 bits                           |
| Average tweet/short text   | 200-500 bytes | UTF-8 text                         |
| Average JSON API response  | 1-10 KB       | Typical REST endpoint              |
| Average web page           | 2-5 MB        | HTML + CSS + JS + images           |
| Average photo (compressed) | 2-5 MB        | JPEG, reasonable quality           |
| Average 1-min video (720p) | 10-30 MB      | Compressed, depends on codec       |
| Average log line           | 200-500 bytes | Structured JSON log                |
| Average DB row             | 100-500 bytes | Depends on schema, obviously       |
| 1 million DB rows          | 100-500 MB    | At 100-500 bytes each              |
| 1 billion DB rows          | 100-500 GB    | Now you're thinking about sharding |

### Quick size conversions

```
1 KB  = 1,000 bytes       (10^3)
1 MB  = 1,000 KB          (10^6)
1 GB  = 1,000 MB          (10^9)
1 TB  = 1,000 GB          (10^12)

Handy shortcuts:
  1 million rows × 500 bytes = 500 MB
  1 billion rows × 500 bytes = 500 GB
  1 million users × 1 KB profile = 1 GB
  1 million × 1 MB photos = 1 TB
```

## Throughput Numbers

Latency tells you how fast one operation is. Throughput tells you how many operations a system can handle per second. They're related but not the same — a database query might take 5 ms (latency), but the database can handle 10,000 of them per second (throughput) because it processes many queries concurrently.

The mental model here is a hierarchy with three tiers. **Databases** sit at the bottom: they're the slowest because every operation potentially touches disk, acquires locks, and maintains consistency guarantees. Writes are slower than reads because they must update the WAL/journal, modify every index, and (in replicated setups) wait for confirmation from replicas. **App servers** sit in the middle: they're fast because they mostly shuffle data between clients and databases. **Caches and queues** sit at the top: they're fastest because they operate in-memory (Redis, Memcached) or do sequential I/O (Kafka appending to a log).

This hierarchy is why the database is almost always the bottleneck. Your app server can handle 30,000 req/sec, but if each request makes 3 database queries, your Postgres instance only needs to handle 10,000 QPS before it's the ceiling.

| System                         | Throughput                 | Notes                                        |
| ------------------------------ | -------------------------- | -------------------------------------------- |
| **Databases (reads)**          |                            |                                              |
| PostgreSQL reads               | 5,000-20,000 queries/sec   | Simple queries, indexed, connection pooled   |
| MySQL reads                    | 5,000-30,000 queries/sec   | Similar to Postgres, slightly faster simple  |
| MongoDB reads                  | 10,000-50,000 ops/sec      | Simple finds, depends on working set in RAM  |
| DynamoDB reads                 | Unlimited (provisioned)    | Pay per RCU, scales horizontally             |
| Elasticsearch reads            | 5,000-20,000 queries/sec   | Depends on index size and query complexity   |
| **Databases (writes)**         |                            |                                              |
| PostgreSQL writes              | 1,000-10,000 inserts/sec   | Depends on indexes, WAL, fsync               |
| MySQL writes                   | 1,000-10,000 inserts/sec   | Similar to Postgres                          |
| MongoDB writes                 | 5,000-25,000 inserts/sec   | Journal + index update, faster without fsync |
| DynamoDB writes                | Unlimited (provisioned)    | Pay per WCU, scales horizontally             |
| Elasticsearch writes           | 5,000-20,000 index ops/sec | Depends on mapping, refresh interval         |
| **Caches (reads)**             |                            |                                              |
| Redis GET                      | 100,000+ ops/sec           | In-memory, single-threaded                   |
| Memcached GET                  | 100,000-500,000 ops/sec    | In-memory, multi-threaded                    |
| **Caches (writes)**            |                            |                                              |
| Redis SET                      | 80,000-100,000 ops/sec     | Slightly lower than GET due to persistence   |
| Memcached SET                  | 100,000-400,000 ops/sec    | No persistence, pure memory                  |
| **Queues (produce/write)**     |                            |                                              |
| Kafka produce (single broker)  | 100,000-500,000 msgs/sec   | Depends on message size, replication         |
| SQS SendMessage (standard)     | ~unlimited                 | FIFO queues: 300/sec, 3,000/sec with batching |
| **Queues (consume/read)**      |                            |                                              |
| Kafka consume (per partition)  | 10,000-50,000 msgs/sec     | Single consumer throughput                   |
| Kafka consume (consumer group) | 100,000-500,000 msgs/sec   | Scales with partitions                       |
| SQS ReceiveMessage (standard)  | ~unlimited                 | FIFO queues: 300/sec; scale with consumers   |
| **App servers**                |                            |                                              |
| Nginx (static/proxy)           | 50,000-100,000 req/sec     | Simple proxy, no heavy processing            |
| Node.js (HTTP, I/O-bound)      | 10,000-30,000 req/sec      | Simple JSON API, no heavy computation        |
| Go (HTTP, I/O-bound)           | 30,000-100,000 req/sec     | Simple JSON API                              |
| Java (HTTP, I/O-bound)         | 20,000-80,000 req/sec      | Spring Boot / Netty, warmed up               |

These are ballpark numbers for a single node with decent hardware (4-8 cores, 16-32 GB RAM, SSD). Real numbers depend on your query complexity, payload size, and whether your data fits in memory.

### At a glance — throughput per single node

```
                    1K         5K        10K        50K       100K       500K
                    │          │          │          │          │          │
  DATABASES (reads)
  Postgres          · ├────████████████──┤           ·          ·          ·
  MySQL             · ├────██████████████████──┤     ·          ·          ·
  MongoDB           ·          · ├───████████████████████──┤    ·          ·
  DynamoDB          ·          ·          ·     ← unlimited (provisioned) →
  Elasticsearch     · ├────████████████──┤           ·          ·          ·
                    │          │          │          │          │          │
  DATABASES (writes)
  Postgres          ├─████████████──┤     ·          ·          ·          ·
  MySQL             ├─████████████──┤     ·          ·          ·          ·
  MongoDB           · ├────████████████████──┤       ·          ·          ·
  DynamoDB          ·          ·          ·     ← unlimited (provisioned) →
  Elasticsearch     · ├────████████████──┤           ·          ·          ·
                    │          │          │          │          │          │
  CACHES (reads)
  Redis GET         ·          ·          ·          ·     ├────█████████████──→
  Memcached GET     ·          ·          ·          ·     ├────██████████████████──→
                    │          │          │          │          │          │
  CACHES (writes)
  Redis SET         ·          ·          ·          ·  ├──█████████████──┤    ·
  Memcached SET     ·          ·          ·          ·     ├────████████████████──→
                    │          │          │          │          │          │
  QUEUES (produce)
  Kafka (broker)    ·          ·          ·          ·     ├────██████████████████──→
  SQS (standard)    ·          ·          ·     ← nearly unlimited (standard) →
                    │          │          │          │          │          │
  QUEUES (consume)
  Kafka (partition) ·          · ├───████████████████████──┤    ·          ·
  Kafka (group)     ·          ·          ·          ·     ├────██████████████████──→
  SQS (standard)    ·          ·          ·     ← nearly unlimited (standard) →
                    │          │          │          │          │          │
  APP SERVERS
  Node.js           ·          ·├──████████████████──┤          ·          ·
  Java              ·          · ├────████████████████████████████──┤      ·
  Go                ·          ·          · ├───████████████████████████████──┤
  Nginx (proxy)     ·          ·          ·          ├─████████████████──┤
                    │          │          │          │          │          │
                    1K         5K        10K        50K       100K       500K
                                     ops/sec per single node

  Pattern: database (1K-50K) → app server (10K-100K) → cache/queue (100K-500K+)
           ◄── usually the bottleneck                    almost never the bottleneck ──►
```

The database is almost always the ceiling. Caches and queues handle 10-100x more throughput than databases. App servers sit in between — you'll need to scale the database long before you need more app servers.

### How instance size changes these numbers

The throughput numbers above assume a mid-range instance (~4-8 vCPU, 16-32 GB RAM). But when you go to AWS and pick an instance, you're choosing from a range where the smallest is 100x weaker than the largest. This matters because a back-of-the-envelope estimate that says "Postgres handles 10,000 QPS" is wrong if you're running on a `db.t3.micro` — that instance tops out around 200-500 QPS.

The mental model: **know which instance class your numbers assume, then adjust up or down**. Doubling cores doesn't double throughput (you hit I/O and lock contention), so the rough scaling rule is **2x cores ≈ 1.5-1.8x throughput** for reads, and less for writes.

**Databases (PostgreSQL RDS — reads, simple indexed queries):**

| Instance class  | vCPU | RAM    | Approx reads/sec | Approx writes/sec | Typical use                     |
| --------------- | ---- | ------ | ---------------- | ----------------- | ------------------------------- |
| db.t3.micro     | 2    | 1 GB   | 200-500          | 50-200            | Dev/test only                   |
| db.t3.medium    | 2    | 4 GB   | 500-2,000        | 200-1,000         | Light production, side projects |
| db.r6g.large    | 2    | 16 GB  | 2,000-5,000      | 500-3,000         | Small production workload       |
| db.r6g.xlarge   | 4    | 32 GB  | 5,000-15,000     | 1,000-8,000       | Medium production               |
| db.r6g.4xlarge  | 16   | 128 GB | 15,000-50,000    | 5,000-20,000      | Large production                |
| db.r6g.16xlarge | 64   | 512 GB | 50,000-100,000+  | 15,000-50,000     | Heavy workload, before sharding |

The numbers in the main throughput table map to roughly `db.r6g.xlarge` to `db.r6g.4xlarge` — a typical production instance. A `db.t3.micro` handles 10-50x less. A `db.r6g.16xlarge` handles 5-10x more. MongoDB (Atlas/DocumentDB) and MySQL (RDS) follow similar scaling curves at similar instance sizes.

**Caches (ElastiCache Redis):**

| Instance class    | vCPU | RAM    | Approx ops/sec  | Typical use                |
| ----------------- | ---- | ------ | --------------- | -------------------------- |
| cache.t3.micro    | 2    | 0.5 GB | 20,000-40,000   | Dev/test, tiny working set |
| cache.r6g.large   | 2    | 13 GB  | 80,000-120,000  | Small-medium production    |
| cache.r6g.xlarge  | 4    | 26 GB  | 100,000-200,000 | Medium production          |
| cache.r6g.4xlarge | 16   | 105 GB | 200,000-500,000 | Large working set          |

Redis is single-threaded for commands, so scaling vCPU helps less than for databases. You scale Redis with sharding (Redis Cluster) rather than bigger instances. The main reason to pick a bigger instance is more RAM for a larger working set, not more throughput.

**App servers (ECS tasks / containers):**

| ECS task size     | vCPU | RAM    | Go req/sec     | Node.js req/sec | Typical use              |
| ----------------- | ---- | ------ | -------------- | --------------- | ------------------------ |
| 0.25 vCPU, 512 MB | 0.25 | 512 MB | 1,000-3,000    | 500-1,500       | Light background workers |
| 1 vCPU, 2 GB      | 1    | 2 GB   | 8,000-20,000   | 3,000-8,000     | Small API service        |
| 4 vCPU, 8 GB      | 4    | 8 GB   | 30,000-80,000  | 10,000-25,000   | Medium API service       |
| 16 vCPU, 32 GB    | 16   | 32 GB  | 80,000-200,000 | 25,000-60,000   | High-throughput service  |

The main throughput table numbers map to roughly 4 vCPU ECS tasks.

**Lambda:**

| Memory (→ vCPU)        | Approx vCPU | Cold start | Warm req/sec (per instance) | Notes                   |
| ---------------------- | ----------- | ---------- | --------------------------- | ----------------------- |
| 128 MB                 | ~0.08       | 200-500 ms | 5-20                        | Minimal, not for APIs   |
| 512 MB                 | ~0.3        | 100-300 ms | 20-100                      | Light event processing  |
| 1,024 MB               | ~0.6        | 50-200 ms  | 50-200                      | Small API handlers      |
| 1,769 MB (1 full vCPU) | 1           | 50-150 ms  | 100-500                     | General purpose         |
| 3,008 MB               | ~2          | 50-150 ms  | 200-1,000                   | Heavier processing      |
| 10,240 MB              | ~6          | 50-200 ms  | 500-2,000                   | CPU-intensive workloads |

Lambda is fundamentally different — throughput scales by running more instances in parallel (up to 1,000-3,000 concurrent by default), not by making one instance faster. A single Lambda handles one request at a time. 1,000 RPS means 1,000 concurrent Lambdas if each takes 1 second, or 100 if each takes 100 ms.

**The sizing mental model:**

```
                  db.t3.micro    db.r6g.large   db.r6g.xlarge   db.r6g.4xlarge   db.r6g.16xlarge
                  (dev/test)     (small prod)   (medium prod)   (large prod)     (before sharding)
                  │              │              │               │                │
  DB reads/sec    200            2,000          5,000-15,000    15,000-50,000    50,000-100,000+
  DB writes/sec   50             500            1,000-8,000     5,000-20,000     15,000-50,000
                  │              │              │               │                │
                  ◄── costs      │     ◄── the "main table"     │                ── costs $$$$ ──►
                  ~$15/mo ──►    │         numbers live here    │
                                 │                              │
                           ◄── most startups ──►    ◄── scale-up before sharding ──►
```

Most startups run on `db.r6g.large` to `db.r6g.xlarge` ($200-400/month) and never need anything bigger. If your estimation says you need 5,000 DB QPS and you're on a `db.r6g.xlarge`, you're fine. If you're on a `db.t3.micro`, you're not — and the fix is sizing up the instance, not redesigning the architecture.

### The important insight

Most web applications don't need to worry about app server throughput. A single Go or Node.js server handles 10,000-50,000 req/sec. If you have 100,000 users making 100 requests/day, that's:

```
100,000 users × 100 requests/day = 10 million/day = ~100 requests/sec
```

That's nothing. A single app server handles this with 99% of its capacity idle. The bottleneck is almost always the database, not the app server.

### The simple throughput model

The same 10x jumps from the latency tiers appear in throughput — just in reverse:

| Tier       | Throughput    | Anchor      |
| ---------- | ------------- | ----------- |
| DB writes  | ~1,000/sec    | 10^3        |
| DB reads   | ~10,000/sec   | 10x writes  |
| Cache ops  | ~100,000/sec  | 10x reads   |

**DB writes ~1K, DB reads ~10K, cache ~100K. Each tier is 10x the previous.**

This connects directly to the latency numbers. A DB read takes ~1-5ms. With ~10 concurrent connections per core on a 4-core instance, that's ~40-80 connections × ~200 reads/sec each ≈ 10,000 reads/sec. Throughput falls out of latency and concurrency combined — the faster the operation, the more of them you can do per second.

App servers (~10K-100K req/sec) sit between DB reads and cache. They're faster than what they call into, which is why they're rarely the bottleneck.

The quick mental check for any design:

| Threshold             | What it signals                      | First move                          |
| --------------------- | ------------------------------------ | ----------------------------------- |
| Approaching 1K writes | WAL pressure, lock contention        | Connection pooling (PgBouncer)      |
| Approaching 10K reads | Database working hard                | Read replicas or cache hot rows     |
| Beyond 100K ops       | Past single-node cache territory     | Redis Cluster or shard the cache    |

## The Quick Math Formulas

The numbers above are your building blocks. Now you need a way to go from a product requirement ("100,000 users placing orders") to a system requirement ("how many database queries per second?"). The chain is always the same:

```
Users → actions per day → requests per second → DB queries per second
                                              → storage per year
                                              → bandwidth
                                              → cache size
```

Every estimation starts with "how many requests per second?" and everything else follows from that single number.

### Time unit conversions for mental math

Forget 86,400. These round numbers are close enough and easy to work with:

| Time period | Quick estimate     | Actual         | Error | Trick             |
| ----------- | ------------------ | -------------- | ----- | ----------------- |
| 1 day       | 100,000 sec (10^5) | 86,400 sec     | ~15%  | Drop 5 zeros      |
| 1 hour      | 4,000 sec          | 3,600 sec      | ~10%  | Drop 3 zeros, ×4  |
| 1 month     | 2.5 million sec    | 2,592,000 sec  | ~3%   | day × 25          |
| 1 year      | 30 million sec     | 31,536,000 sec | ~5%   | day × 300         |

The one rule you need: **1 day ≈ 10^5 seconds. To go from "per day" to "per second", drop 5 zeros.**

| Per day     | Quick (÷10^5) | Exact (÷86,400) | Error |
| ----------- | ------------- | --------------- | ----- |
| 1 million   | 10/sec        | 11.6/sec        | -14%  |
| 10 million  | 100/sec       | 116/sec         | -14%  |
| 100 million | 1,000/sec     | 1,157/sec       | -14%  |
| 1 billion   | 10,000/sec    | 11,574/sec      | -14%  |

The quick number is always ~14% lower. This never changes a design decision — you'd pick the same architecture at 100 RPS or 116 RPS.

### Users to requests per second

```
RPS = (users × actions_per_day) / 10^5       ← just drop 5 zeros
```

| Users      | Actions/day | Total/day | Quick RPS | Exact RPS | Peak (×3) |
| ---------- | ----------- | --------- | --------- | --------- | --------- |
| 100,000    | 10          | 1M        | 10        | 12        | 30        |
| 100,000    | 100         | 10M       | 100       | 116       | 300       |
| 1,000,000  | 50          | 50M       | 500       | 579       | 1,500     |
| 10,000,000 | 20          | 200M      | 2,000     | 2,315     | 6,000     |
| 10,000,000 | 100         | 1B        | 10,000    | 11,574    | 30,000    |

In practice, traffic isn't evenly distributed — peak is usually 2-5x the average. The peak column uses ×3 as a safe default. For spiky traffic (flash sales, viral events), peak can be 10-100x average — that's when you need auto-scaling or rate limiting.

### Requests per second to database queries

This is where most people underestimate. Not every API request is one database query. A "simple" endpoint that shows an order typically involves: check auth (1 query), fetch the order (1 query), fetch order items (1 query), fetch shipping status (1 query), maybe log the access (1 write). That's 5 queries for what feels like one request.

```
DB QPS = RPS × queries_per_request

300 peak RPS × 3 queries = 900 DB queries/sec
```

A single PostgreSQL instance handles 5,000-20,000 queries/sec. So 900 is comfortable. You're at 5-18% capacity. No read replicas needed yet.

### When to worry about database scaling

The progression is always the same: first you optimise queries (indexes, batching), then you add connection pooling, then you add a cache for hot reads, then read replicas, and only then do you consider sharding. Most systems never get past step two.

| DB queries/sec | Single Postgres | Action needed                              |
| -------------- | --------------- | ------------------------------------------ |
| < 1,000        | Relaxed         | Focus on indexes and query optimisation    |
| 1,000-5,000    | Comfortable     | Connection pooling (PgBouncer), monitoring |
| 5,000-15,000   | Working hard    | Read replicas for read-heavy workloads     |
| 15,000-50,000  | Needs help      | Sharding or move reads to cache            |
| > 50,000       | Won't cut it    | Distributed database or aggressive caching |

### Storage estimation

```
Storage = users × data_per_user × retention_period
```

| Scenario                    | Calculation                | Per day | Per year | Store where?       |
| --------------------------- | -------------------------- | ------- | -------- | ------------------ |
| User profiles (1M users)    | 1M × 2 KB                  | —       | 2 GB     | Database           |
| Activity logs (1M users)    | 1M × 50 events × 500 bytes | 25 MB   | 9 GB     | Database           |
| Orders (1M users, 2/day)    | 1M × 2 × 1 KB              | 2 GB    | 730 GB   | Database           |
| Chat messages (1M users)    | 1M × 20 msgs × 500 bytes   | 10 GB   | 3.6 TB   | Database + archive |
| Photos (1M users, 2/week)   | 1M × 2 × 3 MB / 7 days     | 860 GB  | 312 TB   | Object storage     |
| Video (100K creators, 1/wk) | 100K × 500 MB / 7 days     | 7 TB    | 2.6 PB   | Object storage     |

Rule of thumb: anything over ~500 GB/year of blobs goes to object storage (S3). S3 costs ~$23/TB/month.

### Bandwidth estimation

```
Bandwidth = RPS × average_response_size
```

| Scenario                      | Calculation     | Bandwidth | Verdict                |
| ----------------------------- | --------------- | --------- | ---------------------- |
| Small API (300 RPS)           | 300 × 5 KB      | 12 Mbps   | One server, trivial    |
| Medium API (5,000 RPS)        | 5,000 × 10 KB   | 400 Mbps  | Fine on 1 Gbps link    |
| Large API (50,000 RPS)        | 50,000 × 50 KB  | 20 Gbps   | Multiple servers + CDN |
| Image serving (10,000 RPS)    | 10,000 × 500 KB | 40 Gbps   | CDN required           |
| Video streaming (1,000 users) | 1,000 × 5 Mbps  | 5 Gbps    | CDN required           |

Rule of thumb: 1 Gbps ≈ 100,000 RPS of 1 KB responses, or 1,000 RPS of 1 MB responses.

### Memory estimation for caching

```
Cache size = number_of_hot_items × item_size
```

| Scenario                        | Calculation | Cache size | Fits in                         |
| ------------------------------- | ----------- | ---------- | ------------------------------- |
| User sessions (100K)            | 100K × 1 KB | 100 MB     | Single Redis, trivial           |
| User profiles (100K hot)        | 100K × 2 KB | 200 MB     | Single Redis, trivial           |
| Product listings (1M)           | 1M × 5 KB   | 5 GB       | Single Redis, comfortable       |
| Product listings (10M)          | 10M × 5 KB  | 50 GB      | Redis cluster or multiple nodes |
| API responses (5M cached pages) | 5M × 10 KB  | 50 GB      | Redis cluster or multiple nodes |
| Full-text search results (50M)  | 50M × 2 KB  | 100 GB     | Dedicated cache layer           |

Redis single instance: comfortably holds 10-25 GB. Above that, use a cluster or shard. Watch fork latency on snapshots above ~25 GB.

## Worked Example: E-Commerce Order System

This is how you'd actually walk through an estimation in an interview or design discussion. The goal isn't precise numbers — it's making decisions: do I need more than one server? Do I need a cache? Will my database handle this? Each step takes 15-30 seconds.

**Given:** 100,000 users, each placing ~2 orders/day, browsing ~50 pages/day.

### Step 1 — Traffic estimation

```
Browsing:  100,000 × 50 = 5 million/day → drop 5 zeros → ~50 RPS avg → ~150 peak RPS
Ordering:  100,000 × 2  = 200,000/day   → drop 5 zeros → ~2 RPS avg  → ~6 peak RPS
Total:     ~156 peak RPS
```

~156 peak RPS is trivial for any modern app server.

### Step 2 — Database load

```
Browse requests: 150 peak RPS × 3 queries each = 450 DB QPS
Order requests:  6 peak RPS × 5 queries each = 30 DB QPS (but writes!)
Total: ~480 DB QPS
```

A single Postgres instance handles this without breaking a sweat. No read replicas needed.

### Step 3 — Storage

```
Orders: 100,000 users × 2 orders/day × 365 days × 1 KB/order = ~73 GB/year
Products: 500,000 products × 5 KB each = 2.5 GB (static, barely grows)
User data: 100,000 users × 2 KB = 200 MB
```

Total: ~75 GB first year. Fits on a single database.

### Step 4 — Caching

```
Hot products (top 10,000): 10,000 × 5 KB = 50 MB → Redis, trivial
User sessions: 100,000 × 1 KB = 100 MB → Redis, trivial
```

You might not even need a cache at this scale. But if product pages are slow (complex joins), caching the top 10,000 products eliminates most database reads since access patterns follow a power law — top 20% of products get 80% of views.

### Step 5 — What actually matters at this scale

```
~156 peak RPS, ~480 DB QPS, 75 GB storage/year

Answer: one app server, one database, maybe Redis for sessions.
Don't over-engineer. Ship it.
```

### Step 6 — What if it grows 100x?

```
10,000,000 users, same behaviour:
  Browsing: 10M × 50 = 500M/day → 5,000 RPS avg → ~15,000 peak RPS
  Ordering: 10M × 2  = 20M/day  → 200 RPS avg   → ~600 peak RPS
  ~15,600 peak RPS → need multiple app servers behind a load balancer
  ~48,000 DB QPS → need read replicas + caching layer + connection pooling
  7.3 TB storage/year → still fits on one DB, but think about archiving old orders
```

Now you need infrastructure. But you'd know because you'd watch the metrics grow — you don't need to build for this on day one.

## The 5-Second Rules of Thumb

Sometimes you don't even need the formulas. If someone says "we have 50,000 users" you should immediately have a gut reaction about scale. These rules give you that instinct:

1. **100 RPS** — one server, one database, go home early
2. **1,000 RPS** — still one server, make sure you have indexes and connection pooling
3. **10,000 RPS** — multiple app servers, read replicas or caching, you need monitoring
4. **100,000 RPS** — distributed systems, CDN, sharding, dedicated team
5. **1,000,000 RPS** — you're at FAANG scale, everything is custom

Most startups and internal tools never leave category 1 or 2.

## Common Estimation Mistakes

The math itself is easy. What trips people up is the assumptions they feed into it. These are the mistakes that lead to designing a distributed system for a problem that one server could handle, or the opposite — assuming one server is fine when the numbers clearly say it isn't.

**Mistake 1: Assuming peak = average.** Traffic follows patterns — lunch hour, timezone peaks, marketing campaigns. Always multiply by 3-5x for peak.

**Mistake 2: Forgetting that database queries multiply.** 100 RPS doesn't mean 100 DB queries. Auth, main query, related data, logging — it's usually 3-5 queries per request.

**Mistake 3: Treating reads and writes equally.** A database handles 10x more reads than writes. 1,000 write QPS is much harder than 1,000 read QPS because writes need locking, WAL, index updates, and replication.

**Mistake 4: Optimising the app server when the database is the bottleneck.** If your DB query takes 50 ms and your app logic takes 1 ms, rewriting the app in a faster language saves you nothing. Fix the query.

**Mistake 5: Forgetting payload size.** 10,000 RPS of 1 KB responses is 10 MB/sec (easy). 10,000 RPS of 1 MB responses is 10 GB/sec (you need a CDN and a conversation with your team).

## Key Takeaways

The mental model boils down to three ideas:

**1. Three speed worlds.** Nanoseconds (CPU/memory), microseconds (SSD/local), milliseconds (network/disk). Every operation lives in one of these worlds. The jumps between them are 1,000x. This tells you what's worth optimising and what isn't — shaving 0.5 ms off your app logic doesn't matter if the database takes 5 ms and the network takes 50 ms.

**2. The estimation chain.** Users → actions/day → drop 5 zeros → RPS → multiply by queries per request → DB QPS. Then compare that number to what your database instance can handle. If it fits, stop designing. If it doesn't, cache the hot reads first, then consider read replicas, then consider sharding.

**3. Most systems are smaller than you think.** 100,000 users is ~100-500 RPS. That's one server and one database. The instinct to build for scale before you have scale is the most expensive engineering mistake. Do the math first — 2 minutes of estimation prevents weeks of over-engineering.

## What's Next

Now that you can estimate traffic, storage, and throughput, the next lesson traces a single request through every layer of the stack — DNS, load balancer, app server, database, response — and identifies where latency hides and what you can do about it at each stage.
