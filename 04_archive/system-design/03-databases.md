# Databases

The database is almost always the bottleneck. Lesson 01 showed that 50-80% of a typical request's latency is spent in the database. Lesson 02 showed that most performance problems trace back to missing indexes, N+1 queries, or lack of connection pooling. This lesson goes deeper — how databases actually work under the hood, why writes are slower than reads, when to use what, and how to scale when a single node isn't enough.

## The Two Fundamental Paths

Every database operation is either a read or a write. These paths are fundamentally different, and understanding them explains almost every performance behaviour you'll encounter.

### The read path

When you run `SELECT * FROM orders WHERE user_id = 42`:

```
1. Parse the SQL                         → 0.01-0.1 ms
2. Plan the query (which index to use?)  → 0.01-0.5 ms
3. Execute:
   a. Look up index for user_id = 42     → find the row location
   b. Fetch the row from disk/cache      → if in memory: ~0.01 ms
                                          → if on SSD: ~0.1 ms
                                          → if on HDD: ~5 ms
4. Return result over network            → 0.5-1 ms
```

The critical question is step 3b: **is the data in memory or on disk?** If your working set (the data you access frequently) fits in RAM, reads are fast — the database is essentially an in-memory lookup with durability. If it doesn't fit, every read might hit disk, and performance falls off a cliff.

This is why RAM matters more than CPU for databases. A `db.r6g.large` (16 GB RAM) is faster than a `db.m6g.xlarge` (16 GB RAM but more CPU) for most workloads — because the bottleneck is whether the data is cached in memory, not how fast the CPU processes it.

### The write path

When you run `INSERT INTO orders (user_id, total) VALUES (42, 99.00)`:

```
1. Parse the SQL                          → 0.01-0.1 ms
2. Start a transaction (acquire locks)    → 0.01-1 ms (depends on contention)
3. Write to the WAL (Write-Ahead Log)     → 0.1-1 ms (sequential disk write)
4. Update the table data                  → 0.01-0.1 ms (in memory)
5. Update EVERY index on this table       → 0.1-2 ms per index
6. If FK constraints: validate them       → 1-5 ms (reads the referenced table)
7. If replication: send to replicas       → 0-50 ms (async: 0, sync: network RTT)
8. Commit (fsync the WAL to disk)         → 0.1-2 ms
9. Release locks                          → 0.01 ms
10. Return confirmation over network      → 0.5-1 ms
```

Writes are slower because they do more work. The WAL must hit disk for durability (if the server crashes, uncommitted WAL entries are replayed). Every index must be updated — a table with 5 indexes does 5x the index work per write. Locks prevent concurrent writes from corrupting data, but they serialise access to the same rows.

### Why the read:write ratio isn't fixed

The common claim "reads are 3-5x faster than writes" is a rough average, but the real ratio depends on your schema:

| Factor                  | Effect on write speed               | Example                                      |
| ----------------------- | ----------------------------------- | -------------------------------------------- |
| Number of indexes       | Each index adds ~0.1-2 ms per write | 1 index ≈ 2x slower, 10 indexes ≈ 10x slower |
| Foreign key constraints | Each FK requires a read to validate | FK to a large table adds 1-5 ms              |
| Replication mode        | Sync replication waits for replica  | Async ≈ 0 ms, sync ≈ 1-50 ms                 |
| Row size                | Larger rows = more data to WAL      | 100 bytes vs 10 KB row                       |
| Lock contention         | Hot rows = threads waiting          | Updating a counter vs inserting new rows     |
| Trigger/hooks           | Postgres triggers run per row       | Audit triggers, denormalization triggers     |

A table with 1 index, no FKs, and async replication: reads are maybe 2x faster than writes. A table with 10 indexes, 3 FKs, and sync replication: reads might be 20x faster. This is why you can't use a single ratio — you need to think about what your specific writes are doing.

## How Indexes Work

Indexes are the single most important performance lever in any database. A missing index turns a 2 ms query into a 500 ms query. An unnecessary index slows every write on that table. Understanding how they work lets you make the right trade-off.

### The B-tree — how almost all indexes work

Whether it's Postgres, MySQL, MongoDB, or DynamoDB, the default index is a B-tree (or a variant). It works like a sorted phonebook:

```
Looking up user_id = 42 WITHOUT an index:
  Scan every row in the table, check each one.
  1 million rows → 1 million comparisons → 100-1000 ms

Looking up user_id = 42 WITH a B-tree index:
  Tree has ~20 levels for 1 billion rows (log base ~100).
  Start at root → pick the right child → repeat 3-4 times → found.
  3-4 page reads → 0.01-0.1 ms (if in memory) or 0.3-2 ms (if on disk)
```

The difference is O(n) vs O(log n). For a table with 1 million rows, that's 1,000,000 comparisons vs ~4 tree lookups. For 1 billion rows, it's 1,000,000,000 vs ~5 lookups. The bigger the table, the more indexes matter.

### What indexes cost on writes

Every index is a separate data structure that must be maintained. When you insert a row into a table with 5 indexes:

```
1. Insert the row into the table          → 1 operation
2. Insert into index_user_id              → 1 B-tree insertion
3. Insert into index_email                → 1 B-tree insertion
4. Insert into index_created_at           → 1 B-tree insertion
5. Insert into index_status               → 1 B-tree insertion
6. Insert into index_user_id_status       → 1 B-tree insertion
                                          ─────────────────────
                                          6 write operations total
```

Each B-tree insertion means finding the right leaf node, inserting the entry, and potentially splitting the node if it's full. It's not that each one is expensive (~0.1-0.5 ms), but they add up. A table with 10 indexes makes every INSERT/UPDATE/DELETE do 10x the write work.

### The index trade-off

```
Indexes:     0        2        5        10       15
             │        │        │        │        │
Read speed:  terrible good     good     good     good (diminishing returns)
Write speed: fast     good     slower   slow     very slow
Storage:     baseline +20%     +50%     +100%    +150%
```

The rule: **index columns you filter, join, or sort by. Don't index everything.** A read-heavy table (95% reads) benefits from many indexes. A write-heavy table (50%+ writes) should have as few as possible.

### Types of indexes

| Index type       | How it works                 | Good for                                              | Bad for                                   |
| ---------------- | ---------------------------- | ----------------------------------------------------- | ----------------------------------------- |
| B-tree (default) | Sorted tree, O(log n) lookup | Equality, range queries, sorting                      | Full-text search, high cardinality writes |
| Hash             | Direct lookup, O(1)          | Exact equality only                                   | Range queries, sorting                    |
| GIN (Postgres)   | Inverted index               | Full-text search, JSONB, arrays                       | Simple equality, write-heavy              |
| Partial index    | B-tree on a subset of rows   | Filtering rare conditions (`WHERE status = 'active'`) | Queries that don't match the condition    |
| Composite index  | B-tree on multiple columns   | Queries filtering on all or a prefix of those columns | Queries on non-prefix columns             |

### Composite index ordering matters

A composite index on `(user_id, created_at)` is a sorted structure:

```
Index: (user_id, created_at)

Sorted like a phonebook — last name first, then first name:
  (1, 2024-01-01)
  (1, 2024-01-15)
  (1, 2024-02-03)
  (2, 2024-01-05)
  (2, 2024-01-20)
  (42, 2024-03-01)  ← easy to find all of user 42's entries
  (42, 2024-03-15)
  (42, 2024-04-01)
```

This index efficiently supports:

- `WHERE user_id = 42` — jump to user 42's section
- `WHERE user_id = 42 AND created_at > '2024-03-01'` — jump to user 42, then scan forward
- `WHERE user_id = 42 ORDER BY created_at` — already sorted

This index does NOT efficiently support:

- `WHERE created_at > '2024-03-01'` — can't skip to a date without knowing the user_id first (same as how a phonebook sorted by last name, first name doesn't help you find all people named "Alice")

The leftmost prefix rule: a composite index on `(A, B, C)` supports queries on `(A)`, `(A, B)`, and `(A, B, C)`, but not `(B)`, `(C)`, or `(B, C)` alone.

## Connection Pooling

Opening a new database connection is expensive:

```
New connection every time:
  1. TCP handshake              → 0.5-1 ms
  2. TLS handshake (if enabled) → 0.5-2 ms
  3. Postgres auth handshake    → 1-5 ms
  4. Postgres forks a process   → 1-5 ms (Postgres creates a new OS process per connection)
  ─────────────────────────────────────
  Total: 3-13 ms before your query even runs

  × 100 requests/sec = 100 new connections/sec = constant overhead

Connection pool:
  Startup: open 20 connections (pay the cost once)
  Per request: grab an idle connection from pool → 0.05 ms
  After request: return connection to pool

  Total: 0.05 ms per request. The connections stay open for hours.
```

Postgres is particularly sensitive because it creates a new OS process per connection (unlike MySQL which uses threads). 500 Postgres connections = 500 OS processes, each with its own memory. At ~10 MB per connection, 500 connections = 5 GB of overhead just for connection management.

### Connection pool sizing

The right pool size isn't "as many as possible." Too many connections cause contention inside the database:

```
Pool too small (5 connections, 100 concurrent requests):
  95 requests waiting for a connection → queuing delay
  5 requests running queries → database is barely loaded

Pool too large (500 connections, 100 concurrent requests):
  100 requests all running queries simultaneously
  Database thrashes: lock contention, context switching, shared buffer contention
  Paradoxically slower than fewer connections

Sweet spot (20-50 connections for most workloads):
  All connections busy, minimal queuing
  Database has manageable concurrent load
```

The rule of thumb from the Postgres community: **connections ≈ (2 × CPU cores) + number of disks**. For a 4-core instance with SSD, that's roughly 10-20 connections. If you need more concurrency, use a pooler like PgBouncer in front of Postgres — it multiplexes hundreds of application connections onto a smaller number of real Postgres connections.

| App connections        | PgBouncer connections to Postgres | Actual Postgres processes |
| ---------------------- | --------------------------------- | ------------------------- |
| 500 (from app servers) | 20 (managed by PgBouncer)         | 20                        |

PgBouncer adds ~0.1 ms of latency but lets you handle 500 concurrent application connections with only 20 Postgres processes.

## Database Comparison

Each database makes different trade-offs. The choice isn't "which is best" — it's "which trade-offs match my workload."

### PostgreSQL

Postgres is the default choice for most applications. It's the most versatile — relational, JSONB support, full-text search, geospatial, strong consistency, and a massive ecosystem.

**How it works internally:**

- Processes, not threads — each connection gets its own OS process
- MVCC (Multi-Version Concurrency Control) — readers never block writers, writers never block readers
- WAL-based durability — writes go to the WAL first, then to data files
- Shared buffer cache — hot data lives in RAM, managed by Postgres (not just the OS page cache)

**Where Postgres excels:**

- Complex queries with JOINs, aggregations, subqueries — the query planner is excellent
- Mixed read/write workloads — MVCC handles concurrency well
- Data integrity — constraints, transactions, foreign keys
- JSONB — when you need some schema flexibility without giving up SQL

**Where Postgres struggles:**

- Very high write throughput (>50,000 writes/sec on a single node) — MVCC creates dead tuples that need vacuuming
- Horizontal scaling — sharding is manual or requires extensions (Citus). No built-in auto-sharding.
- Connection overhead — process-per-connection model means PgBouncer is almost mandatory at scale

**Typical numbers (db.r6g.xlarge — 4 vCPU, 32 GB):**

| Metric                      | Value                               |
| --------------------------- | ----------------------------------- |
| Simple reads (indexed)      | 5,000-15,000/sec                    |
| Simple writes               | 1,000-8,000/sec                     |
| Max connections (practical) | 200-300 (use PgBouncer for more)    |
| Max table size before pain  | ~500 GB (then partition or archive) |
| Vacuum overhead             | 5-20% of write throughput           |

### MySQL

MySQL is similar to Postgres for most workloads. The InnoDB engine (default since 5.5) provides ACID transactions, MVCC, and row-level locking. MySQL is typically slightly faster for simple reads and slightly worse for complex queries.

**Key differences from Postgres:**

- Threads, not processes — lighter per-connection overhead. 1,000 connections is fine without a pooler.
- Simpler query planner — handles simple queries well but struggles with complex JOINs and subqueries compared to Postgres
- Replication is easier — built-in async and semi-sync replication, widely used
- Less extensible — no equivalent to Postgres extensions, custom types, or GIN indexes

**When to pick MySQL over Postgres:**

- You need a very large number of connections (>500) without a pooler
- Simple read-heavy workload (CRUD API, no complex analytics)
- Team has strong MySQL expertise
- You're using AWS Aurora (Aurora MySQL has better read scaling than Aurora Postgres in some benchmarks)

For most new projects, Postgres is the better default. Pick MySQL if you have a specific reason.

### MongoDB

MongoDB is a document database. Instead of tables and rows, you have collections and documents (JSON-like objects). The fundamental difference: MongoDB stores each document as a self-contained unit, while a relational database normalises data across tables.

**How it works internally:**

- Documents are stored in BSON (binary JSON) — each document can have a different structure
- WiredTiger storage engine — B-tree indexes, document-level locking, compression
- Journal (equivalent to WAL) — writes go to journal first for durability
- Working set concept — performance depends on whether your active data fits in RAM

**Where MongoDB excels:**

- Schema flexibility — documents can have different fields, nested objects, arrays
- Embedded data — instead of JOIN (order + order_items), embed items inside the order document. One read fetches everything.
- Horizontal scaling — built-in auto-sharding. You pick a shard key, MongoDB handles the rest.
- High write throughput — document-level locking means less contention than row-level locking for unrelated writes

**Where MongoDB struggles:**

- JOINs — `$lookup` exists but is much slower than SQL JOINs. If you need many JOINs, use a relational database.
- Transactions across documents — supported since 4.0, but slower and more limited than Postgres. If your data is highly relational, Mongo is the wrong tool.
- Data duplication — embedding data means updating the same data in multiple documents. "The user changed their address" might mean updating thousands of order documents.
- Aggregation pipeline — powerful but harder to read and optimise than SQL

**The embedding vs referencing decision:**

```
Relational (Postgres):
  orders table:      { id: 1, user_id: 42, total: 99 }
  order_items table: { id: 1, order_id: 1, product: "Widget", qty: 2 }
                     { id: 2, order_id: 1, product: "Gadget", qty: 1 }

  To get an order with items: JOIN orders and order_items (2 table reads)

Document (MongoDB):
  orders collection: {
    _id: 1,
    user_id: 42,
    total: 99,
    items: [
      { product: "Widget", qty: 2 },
      { product: "Gadget", qty: 1 }
    ]
  }

  To get an order with items: one document read (everything is embedded)
```

Embed when: the child data always appears with the parent, doesn't change independently, and doesn't grow unbounded. Reference when: the child is shared across parents, changes independently, or could grow very large.

**Typical numbers (M30 Atlas / equivalent — 8 GB RAM):**

| Metric                              | Value                |
| ----------------------------------- | -------------------- |
| Simple reads (by \_id)              | 10,000-30,000/sec    |
| Simple writes                       | 5,000-15,000/sec     |
| Max connections                     | 1,500+ (lightweight) |
| Max document size                   | 16 MB (hard limit)   |
| Max collection size before sharding | ~100 GB (then shard) |

### DynamoDB

DynamoDB is fundamentally different from the others. It's not a general-purpose database — it's a managed key-value/document store designed for specific access patterns at any scale. You trade query flexibility for guaranteed performance and zero operational overhead.

**How it works internally:**

- Data is partitioned by partition key (hash) — DynamoDB decides which server stores each partition
- No server to manage — AWS handles scaling, replication, patching, backups
- Provisioned or on-demand capacity — you pay per read/write unit, not per server
- Single-digit millisecond reads and writes at any scale — the latency promise

**Where DynamoDB excels:**

- Known access patterns — if you know exactly how you'll query your data, DynamoDB is unbeatable
- Massive scale with zero ops — billions of rows, millions of requests/sec, no server management
- Consistent performance — latency doesn't degrade as data grows (because partitioning is built in)
- Serverless architectures — pairs naturally with Lambda, API Gateway

**Where DynamoDB struggles:**

- Ad-hoc queries — no SQL, no JOINs, no aggregations. You can only query by partition key + sort key, or by secondary indexes you defined upfront.
- Changing access patterns — if your query patterns change, you may need to redesign your table. Adding a new "get all orders by status" query might require a new Global Secondary Index.
- Analytics — you can't run "SELECT COUNT(\*) WHERE created_at > last_week" efficiently. Use DynamoDB Streams → S3 → Athena for analytics.
- Cost at medium scale — for a Postgres-sized workload (5,000 QPS), DynamoDB can be more expensive than an RDS instance. It wins on cost at very large scale or very small scale (on-demand).

**The single-table design pattern:**

DynamoDB works best when you model your data around your access patterns, not around entities:

```
Relational thinking (doesn't work well in DynamoDB):
  Users table, Orders table, Products table → JOIN them

DynamoDB thinking:
  One table, multiple entity types, access-pattern-driven keys:

  PK              SK                  Data
  USER#42         PROFILE             { name: "Alice", email: "..." }
  USER#42         ORDER#001           { total: 99, status: "shipped" }
  USER#42         ORDER#002           { total: 45, status: "pending" }
  ORDER#001       ITEM#1              { product: "Widget", qty: 2 }
  ORDER#001       ITEM#2              { product: "Gadget", qty: 1 }

  Get user profile: Query PK=USER#42, SK=PROFILE  → 1 read
  Get user's orders: Query PK=USER#42, SK begins_with ORDER  → 1 query
  Get order items: Query PK=ORDER#001, SK begins_with ITEM  → 1 query
```

This feels unnatural coming from relational databases, but it means every query is a single partition key lookup — O(1) regardless of table size.

**Typical numbers:**

| Metric               | Value                              |
| -------------------- | ---------------------------------- |
| GetItem latency      | 1-5 ms (single-digit ms guarantee) |
| PutItem latency      | 2-10 ms                            |
| On-demand read cost  | $0.25 per million reads            |
| On-demand write cost | $1.25 per million writes           |
| Max item size        | 400 KB                             |
| Max throughput       | Unlimited (auto-scales)            |

## When to Use What

The decision isn't about which database is "best." It's about matching the trade-offs to your workload.

```
                           Query flexibility ──────────────────►
                           │
  Operational overhead     │  DynamoDB         MongoDB          Postgres
  ◄──────────────────      │  (zero ops,       (flexible        (most flexible,
                           │   limited          schema,          most operational
                           │   queries)         good scaling)    overhead)
                           │
                           │  ● Key-value       ● Documents      ● Relational
                           │    access only       with embedded   ● Complex JOINs
                           │  ● Known access      data            ● Aggregations
                           │    patterns        ● Some JOINs     ● Transactions
                           │  ● Any scale       ● Horizontal     ● Strong integrity
                           │                      sharding
```

| If your workload is...                    | Start with          | Why                                                                |
| ----------------------------------------- | ------------------- | ------------------------------------------------------------------ |
| Generic web app (CRUD, some JOINs)        | Postgres            | Most flexible, handles everything adequately                       |
| Simple key-value lookups at massive scale | DynamoDB            | Zero ops, guaranteed latency at any scale                          |
| Schema varies per record, embedded data   | MongoDB             | Document model is natural, no migrations needed                    |
| High write throughput, horizontal scaling | MongoDB or DynamoDB | Both shard natively, Postgres doesn't                              |
| Complex analytics and reporting           | Postgres            | Best query planner, CTEs, window functions                         |
| Serverless (Lambda-based)                 | DynamoDB            | No connection limits, no pool management, scales to zero           |
| You don't know yet                        | Postgres            | It can do everything. Optimise when you know your access patterns. |

## Scaling a Database

Scaling follows a predictable progression. Each step is harder and more expensive than the last. Don't jump ahead — exhaust the simpler options first.

### Step 1: Optimise what you have (0-5,000 QPS)

Before scaling the database, make sure you're not wasting the capacity you already have.

**Check indexes:** Run `EXPLAIN ANALYZE` on your slowest queries. A missing index is the most common performance problem and the cheapest fix. A single index addition can turn a 500 ms query into a 2 ms query.

**Check for N+1 queries:** If your app makes 50 queries to render one page, the fix isn't a bigger database — it's batching those queries. Use JOINs, IN clauses, or dataloader patterns.

**Check connection pooling:** If you're opening and closing connections per request, add PgBouncer (Postgres) or use a built-in pool. This alone can 2-3x your effective throughput.

**Check query patterns:** Are you `SELECT *` when you need 3 columns? Are you fetching 10,000 rows when you display 50? Paginate, select specific columns, use LIMIT.

### Step 2: Vertical scaling (5,000-30,000 QPS)

Buy a bigger instance. This is the simplest and most underrated scaling strategy.

```
db.r6g.large  (2 vCPU, 16 GB)  →  5,000 reads/sec     $200/month
db.r6g.xlarge (4 vCPU, 32 GB)  →  15,000 reads/sec    $400/month
db.r6g.4xlarge (16 vCPU, 128 GB) → 50,000 reads/sec   $1,600/month
```

Going from 5,000 to 50,000 QPS by changing an instance class takes 10 minutes and costs $1,400/month more. The engineering effort to shard your database or add a caching layer costs far more in developer time. Always consider vertical scaling first.

The limit: the biggest RDS instance (`db.r6g.16xlarge`, 64 vCPU, 512 GB) handles roughly 50,000-100,000 reads/sec. If you need more, you need horizontal scaling.

### Step 3: Read replicas (30,000-100,000 read QPS)

If your workload is read-heavy (most web apps are — 80-95% reads), you can create read replicas that handle read traffic while the primary handles writes.

```
                    Writes
                      │
                      ▼
                ┌───────────┐
                │  Primary  │ ← handles all writes
                │  (writer) │
                └─────┬─────┘
                      │ replication (async, ~10-100 ms lag)
              ┌───────┼───────┐
              ▼       ▼       ▼
         ┌─────────┐ ┌─────────┐ ┌─────────┐
         │Replica 1│ │Replica 2│ │Replica 3│  ← handle reads
         └─────────┘ └─────────┘ └─────────┘

  Write capacity: unchanged (still one primary)
  Read capacity: 3-4x (one primary + three replicas)
```

**What replication gives you:**

- More read throughput (each replica handles ~same QPS as primary)
- Better availability (if primary fails, promote a replica)
- Geographic distribution (replica in EU for EU users)

**What replication does NOT give you:**

- More write throughput — all writes still go to one primary
- Strong consistency — replicas are slightly behind (10-100 ms for async replication). A user writes then immediately reads → might not see their own write.

**Handling replication lag:**

| Strategy                    | How it works                                                                  | Trade-off                                   |
| --------------------------- | ----------------------------------------------------------------------------- | ------------------------------------------- |
| Read-your-writes            | After a write, read from primary for that user for a few seconds              | Slightly more load on primary               |
| Causal consistency          | Track which writes a user has done, route reads to a replica that's caught up | More complex routing logic                  |
| Accept eventual consistency | User might see stale data briefly                                             | Simplest, works for most non-critical reads |

Most applications can tolerate a few hundred milliseconds of stale data for most reads. "Show me my orders" can be slightly stale. "Show me my account balance after a transfer" should read from primary.

### Step 4: Caching hot reads (reducing DB load by 50-90%)

Covered in depth in lesson 04, but the key insight: if 20% of your data gets 80% of reads (it almost always does), caching that 20% in Redis removes 80% of your database read load.

```
Before cache:
  All reads → Database (10,000 QPS)

After cache (90% hit rate):
  90% of reads → Redis (9,000 QPS, ~0.5 ms each)
  10% of reads → Database (1,000 QPS)

  Database load dropped from 10,000 to 1,000 QPS — an order of magnitude.
```

This is often more effective than read replicas because it reduces latency (Redis: 0.5 ms, database: 5 ms) AND reduces database load.

### Step 5: Partitioning / table partitioning (large tables, 100M+ rows)

Before sharding the entire database, you can partition large tables. Partitioning splits one logical table into multiple physical tables, usually by a time range or ID range.

```
orders table (500M rows, 200 GB)  →  partitioned by month:
  orders_2024_01 (40M rows, 15 GB)
  orders_2024_02 (42M rows, 16 GB)
  ...
  orders_2024_12 (45M rows, 18 GB)

Query: WHERE created_at > '2024-11-01'
  → scans only orders_2024_11 and orders_2024_12 (2 partitions)
  → instead of scanning 500M rows
```

Partitioning also makes maintenance easier — dropping old data is `DROP PARTITION` (instant) instead of `DELETE FROM` (slow, generates dead tuples).

This stays on a single database instance. It's a data organisation strategy, not a scaling strategy — but it enables more efficient queries and easier data lifecycle management.

### Step 6: Sharding (100,000+ QPS, or data too large for one node)

Sharding splits your data across multiple independent database instances. Each shard holds a portion of the data, determined by a shard key.

```
Shard by user_id:

  user_id % 4 = 0  →  Shard 0  (users 0, 4, 8, 12...)
  user_id % 4 = 1  →  Shard 1  (users 1, 5, 9, 13...)
  user_id % 4 = 2  →  Shard 2  (users 2, 6, 10, 14...)
  user_id % 4 = 3  →  Shard 3  (users 3, 7, 11, 15...)

  Each shard: independent Postgres/MySQL instance
  Total capacity: 4x a single instance
```

**Sharding is hard because:**

- Cross-shard queries are expensive — "get all orders from last week" must query all shards and merge results
- Joins across shards are impossible (or extremely slow)
- Rebalancing (adding/removing shards) requires moving data
- Application code must know about sharding (which shard has user 42?)
- Transactions across shards require distributed transactions (2PC) which are slow and complex

**Choosing a shard key:**

| Shard key         | Good for                            | Bad for                                          |
| ----------------- | ----------------------------------- | ------------------------------------------------ |
| user_id           | User-scoped queries (most web apps) | "All orders this week" (cross-shard)             |
| tenant_id         | Multi-tenant SaaS                   | Uneven tenant sizes (one big tenant = hot shard) |
| geographic region | Region-scoped queries               | Cross-region queries                             |
| random hash       | Even distribution                   | Range queries, locality                          |

The best shard key: one that means **most of your queries only hit one shard**. For most web apps, that's `user_id` or `tenant_id` — because most operations are "get this user's stuff."

**When to shard vs. when to use a natively distributed database:**

| Approach             | Example                                                  | Effort | When                                                  |
| -------------------- | -------------------------------------------------------- | ------ | ----------------------------------------------------- |
| Manual sharding      | Application-level routing to separate Postgres instances | High   | Full control, Postgres-specific features needed       |
| Managed sharding     | Citus (Postgres), Vitess (MySQL)                         | Medium | Want Postgres/MySQL with built-in sharding            |
| Natively distributed | CockroachDB, TiDB, YugabyteDB                            | Low    | Willing to trade some features for automatic sharding |
| DynamoDB / MongoDB   | Built-in auto-sharding                                   | Low    | Access patterns fit the model                         |

## The Scaling Decision Flowchart

```
Is the database slow?
  │
  ├─ No → Don't touch it. Ship features.
  │
  └─ Yes → Where is the time going?
       │
       ├─ Slow queries → Add indexes, fix N+1, use EXPLAIN ANALYZE
       │                  (free, takes hours)
       │
       ├─ Too many connections → Add connection pooling (PgBouncer)
       │                        (free, takes hours)
       │
       ├─ CPU/memory maxed → Vertical scale: bigger instance
       │                     (easy, takes minutes, costs $$)
       │
       ├─ Read QPS too high → Add read replicas
       │   │                  (medium effort, takes a day)
       │   │
       │   └─ Still too high → Add caching layer (Redis)
       │                       (medium effort, takes days)
       │
       ├─ Write QPS too high → Reduce indexes on write-heavy tables
       │   │                   Queue + batch writes
       │   │
       │   └─ Still too high → Shard the database
       │                       (hard, takes weeks-months)
       │
       └─ Data too large → Partition large tables
                           Archive old data to cold storage
                           (medium effort, takes days)
```

Most applications never get past "bigger instance" or "add a read replica." Sharding is a last resort, not a starting point.

## Key Takeaways

**1. Reads vs writes are different worlds.** Reads find data and return it. Writes must update the WAL, modify every index, validate constraints, and replicate. The speed gap depends on your schema — more indexes and constraints mean a wider gap.

**2. Indexes are the most important lever.** A missing index is the #1 performance problem. But each index slows every write. Index what you query; don't index what you don't.

**3. Connection pooling is mandatory for Postgres.** Postgres creates a process per connection. Use PgBouncer. MySQL is more forgiving with threads, but pooling still helps.

**4. Choose the right database for your access patterns.** Postgres for flexibility and complex queries. MongoDB for documents and horizontal scaling. DynamoDB for known access patterns at any scale. "I don't know yet" means Postgres.

**5. Scale in order: optimise → vertical → replicas → cache → shard.** Each step is 10x harder than the last. Exhaust the easy options before reaching for the hard ones. Most systems never need sharding.

## What's Next

Caching is the most effective way to reduce database load — a well-configured cache eliminates 80-90% of reads. The next lesson covers Redis, Memcached, and CDNs: when to cache, what to cache, invalidation strategies, and the trade-offs you're accepting.
