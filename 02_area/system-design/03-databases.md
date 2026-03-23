# Databases

The database is almost always the bottleneck — knowing _why_ tells you where to look before you reach for a bigger instance or a new tool.

## Core Mental Model: Two Paths, One Question

Every database operation is either a read or a write. These are fundamentally different paths, and nearly every performance behaviour you'll encounter traces back to which path you're on and what's happening along it.

**The one question that determines read speed:** is the data in RAM or on disk?

```
Data location   Latency         Scale
────────────────────────────────────────────────────────
RAM             ~0.01 ms   ·
SSD             ~0.1 ms    ··
HDD             ~5 ms      ··············
```

If your working set fits in RAM, reads are fast. If it doesn't, every read might hit disk and performance falls off a cliff. This is why RAM matters more than CPU for databases — a memory-optimised instance (`r` family on AWS) beats a compute-optimised one (`m` family) for most DB workloads.

**Writes are slower because they do more work.** A read finds data and returns it. A write must:

1. Write to the WAL (durability — survives a crash)
2. Update **every index** on the table
3. Validate FK constraints (which requires reads)
4. Acquire and release locks
5. Replicate to replicas (if sync: wait for network RTT)

The read/write speed gap is not a fixed ratio — it depends entirely on your schema. A table with 1 index and no FKs might be 2x faster to read. A table with 10 indexes and sync replication might be 20x faster to read. More indexes = wider gap.

## Indexes

A missing index is the single most common performance problem. It turns a 2 ms query into a 500 ms query. The fix is cheap; the oversight is expensive.

The default index type is a B-tree: a sorted structure that turns O(n) full-table scans into O(log n) lookups. The size of the table is nearly irrelevant once indexed.

```
Without index (1M rows):  scan all rows
  comparisons: ████████████████████████████████████  ~1,000,000

With B-tree index (1B rows):  tree traversal
  comparisons: ·  ~30
```

But every index is a separate structure maintained on every write. A table with 5 indexes performs 6 write operations per insert. The more indexes, the slower the writes.

```
Indexes:       0        2        5        10       15
               │        │        │        │        │
Read speed:    terrible good     good     good     good
Write speed:   fast     good     slower   slow     very slow
Storage:       1x       1.2x     1.5x     2x       2.5x
```

**Index columns you filter, join, or sort by. Don't index everything.** Read-heavy tables can afford many indexes. Write-heavy tables should have as few as possible.

**Composite index column order matters.** A composite index on `(user_id, created_at)` supports queries on `user_id` alone or `user_id + created_at`, but not `created_at` alone — the same way a phonebook sorted by last name then first name doesn't help you find everyone named "Alice."

Leftmost prefix rule: `(A, B, C)` supports `(A)`, `(A, B)`, `(A, B, C)` — not `(B)` or `(C)` alone.

| Index type | Good for                      | Bad for                 |
| ---------- | ----------------------------- | ----------------------- |
| B-tree     | Equality, range, sorting      | Full-text search        |
| Hash       | Exact equality only           | Range queries, sorting  |
| GIN        | Full-text, JSONB, arrays      | Simple equality, writes |
| Partial    | Filtering rare conditions     | General queries         |
| Composite  | Queries on all/prefix of cols | Non-prefix columns      |

## Connection Pooling

Opening a new database connection costs 3–13 ms: TCP handshake, TLS, auth, plus Postgres forking an OS process. At 100 req/s, that's 100 new connections per second — pure overhead before a single query runs.

**Postgres is especially sensitive** because it creates one OS process per connection (~10 MB each). 500 connections = 5 GB of memory consumed before the database does any real work.

```
New connection per request:   3–13 ms setup + query time
Connection from pool:         ~0.05 ms grab + query time
```

A pool pays the connection cost once at startup and reuses connections. **This is mandatory for Postgres at any meaningful scale.**

Pool sizing: `connections ≈ (2 × CPU cores) + number of disks`. For a 4-core instance, that's ~10–20 connections. If you need more concurrency, put PgBouncer in front — it multiplexes hundreds of application connections onto 20 real Postgres connections, adding only ~0.1 ms latency.

## Which Database

The choice is not about which is "best" — it's about which trade-offs match your access patterns.

```
                 Query flexibility ──────────────────►
                 ▲
Ops overhead     DynamoDB        MongoDB          Postgres
◄──────────────  zero ops,       flexible schema  most flexible
                 limited         good scaling     most ops overhead
                 queries
```

| Workload                                | Use          | Why                                                  |
| --------------------------------------- | ------------ | ---------------------------------------------------- |
| Generic web app (CRUD, some JOINs)      | **Postgres** | Most flexible, handles everything adequately         |
| Key-value lookups at massive scale      | DynamoDB     | Zero ops, guaranteed single-digit ms at any scale    |
| Schema varies per record, embedded data | MongoDB      | Document model fits naturally, no migrations         |
| High write throughput, need to shard    | Mongo/Dynamo | Both shard natively; Postgres doesn't                |
| Complex analytics, reporting            | **Postgres** | Best query planner, CTEs, window functions           |
| Serverless (Lambda-based)               | DynamoDB     | No connection limits, scales to zero                 |
| You don't know yet                      | **Postgres** | It can do everything. Optimise when patterns emerge. |

**Postgres** uses MVCC (readers never block writers), a process-per-connection model, and the most capable query planner. Struggles above ~50k writes/sec on one node due to dead tuple accumulation (MVCC vacuum overhead). No built-in horizontal sharding.

**MongoDB** embeds child data inside the parent — one read fetches everything, no JOIN needed. Built-in auto-sharding. But `$lookup` (the JOIN equivalent) is slow, and embedding means data duplication: "user changed their address" may require updating thousands of order documents. Embed when child data always appears with the parent and doesn't grow unbounded. Reference when child data is shared or changes independently.

**DynamoDB** has no SQL, no JOINs, no ad-hoc queries. You can only query by partition key + sort key, or via Global Secondary Indexes you defined upfront. Single-table design models all entities in one table with keys that encode access patterns — every query is a O(1) partition key lookup regardless of table size. Changing access patterns means redesigning the table. Cost: favourable at very large or very small scale; can be expensive at medium scale compared to RDS.

## Scaling — In Order

The progression is predictable. Each step is harder and more expensive than the last. Don't skip.

```
Is the database slow?
  │
  ├─ Slow queries       → EXPLAIN ANALYZE, add indexes, fix N+1     (hours, free)
  ├─ Too many conns     → PgBouncer / connection pool                (hours, free)
  ├─ CPU/memory maxed   → Vertical scale: bigger instance            (minutes, $$)
  ├─ Read QPS too high  → Read replicas → then caching (Redis)       (days, $$$)
  ├─ Write QPS too high → Reduce indexes, batch writes → then shard  (weeks, hard)
  └─ Data too large     → Partition tables, archive cold data        (days)
```

**Vertical scaling is underrated.** Going from 5,000 to 50,000 reads/sec by bumping an RDS instance class takes 10 minutes. The engineering effort to shard or cache costs far more in developer time. Always consider vertical scaling before architectural changes.

**Read replicas** add read throughput without changing write capacity. Replicas lag 10–100 ms behind the primary (async replication). A user might not see their own write immediately. For most reads, this is acceptable. For "show account balance after a transfer," read from the primary.

**Caching** (covered separately) is often more effective than replicas — Redis at 0.5 ms vs DB at 5 ms, and it removes load entirely rather than distributing it. A 90% cache hit rate reduces database read load by 90%.

**Sharding** splits data across independent database instances by a shard key. It's hard because cross-shard queries must hit all shards and merge results, JOINs across shards are impossible, rebalancing requires moving data, and distributed transactions (2PC) are slow and complex. Best shard key: one where most queries hit only one shard — for most web apps, `user_id` or `tenant_id`.

Most systems never need sharding. Exhaust every other option first.

## Key Mental Models

1. **Reads = find + return. Writes = WAL + all indexes + locks + replicate.** More indexes means a wider gap.
2. **"Is it in RAM?" is the most important read question.** Disk is 10–100x slower. RAM matters more than CPU.
3. **Missing index is the #1 performance problem.** Run `EXPLAIN ANALYZE` before anything else.
4. **Postgres needs PgBouncer.** Process-per-connection is expensive at scale. MySQL is more forgiving (threads).
5. **Match the database to your access patterns.** Postgres if unsure. DynamoDB if you know exactly how you'll query at scale. MongoDB if data is naturally document-shaped.
6. **Scale in order: optimise → vertical → replicas → cache → shard.** Each step is ~10x harder. Sharding is a last resort, not a starting point.
