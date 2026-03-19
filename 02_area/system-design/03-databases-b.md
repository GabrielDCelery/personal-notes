# Databases

The database is almost always the bottleneck. Everything here is about knowing why — so you can make fast decisions.

## Reads vs Writes

**Reads are fast when data is in RAM. If it hits disk, they're 10-100x slower.**
This is why RAM matters more than CPU for databases. A memory-optimised instance (`r` family on AWS) beats a compute-optimised one (`m` family) for most DB workloads.

**Writes are slower than reads because they do more work:**

- Write to the WAL (durability)
- Update every index on the table
- Validate FK constraints (requires reads)
- Acquire and release locks
- Replicate to replicas (if sync)

The more indexes and constraints you have, the wider the read/write gap. A table with 10 indexes might be 20x faster to read than to write.

## Indexes

**The most important performance lever.** A missing index turns a 2 ms query into a 500 ms query.

- Default is a B-tree: O(log n) lookup vs O(n) full scan. For 1M rows, that's ~4 lookups vs 1,000,000 comparisons.
- Every index is a separate structure that must be updated on every write.
- More indexes = faster reads, slower writes, more storage.

**Rule:** Index columns you filter, join, or sort by. Don't index everything. Write-heavy tables should have as few indexes as possible.

**Composite indexes — order matters:**
`(user_id, created_at)` supports queries on `user_id` alone or `user_id + created_at`, but NOT `created_at` alone.
Leftmost prefix rule: `(A, B, C)` supports `(A)`, `(A, B)`, `(A, B, C)` — not `(B)` or `(C)` alone.

| Index type | Good for                            | Bad for                 |
| ---------- | ----------------------------------- | ----------------------- |
| B-tree     | Equality, range, sorting            | Full-text               |
| Hash       | Exact equality only                 | Range, sorting          |
| GIN        | Full-text search, JSONB, arrays     | Simple equality, writes |
| Partial    | Filtering rare conditions           | General queries         |
| Composite  | Queries on all/prefix of those cols | Non-prefix columns      |

## Connection Pooling

Opening a new DB connection costs 3-13 ms (TCP + TLS + auth handshake + process fork). At 100 req/s that's constant overhead.

**Postgres is especially sensitive** — it forks a new OS process per connection (~10 MB each). 500 connections = 5 GB of overhead before a single query runs.

Use a pool: pay the cost once at startup, reuse connections per request (~0.05 ms overhead).

**Pool sizing:** connections ≈ (2 × CPU cores) + disks. For a 4-core instance, ~10-20 connections. If you need more concurrency, put PgBouncer in front — it multiplexes hundreds of app connections onto 20 real Postgres connections, adding ~0.1 ms latency.

## Which Database

| If your workload is...                  | Use          | Why                                                     |
| --------------------------------------- | ------------ | ------------------------------------------------------- |
| Generic web app (CRUD, some JOINs)      | Postgres     | Most flexible, handles everything adequately            |
| Key-value lookups at massive scale      | DynamoDB     | Zero ops, guaranteed single-digit ms at any scale       |
| Schema varies per record, embedded data | MongoDB      | Document model fits, no migrations                      |
| High write throughput, need to shard    | Mongo/Dynamo | Both shard natively, Postgres doesn't                   |
| Complex analytics, reporting            | Postgres     | Best query planner, CTEs, window functions              |
| Serverless (Lambda)                     | DynamoDB     | No connection limits, scales to zero                    |
| You don't know yet                      | Postgres     | It can do everything. Optimise when patterns are clear. |

### Postgres

- Process per connection → use PgBouncer at scale
- MVCC: readers never block writers, writers never block readers
- Struggles above ~50k writes/sec on one node (MVCC dead tuples + vacuum overhead)
- No built-in horizontal sharding

### MongoDB

- Document model: embed child data inside the parent — one read fetches everything (vs JOIN)
- Built-in auto-sharding
- `$lookup` (JOIN equivalent) is slow — if you need many JOINs, use Postgres
- Data duplication is a real cost: "user changed address" may update thousands of documents
- Embed when child always appears with parent and doesn't grow unbounded. Reference when child is shared or changes independently.

### DynamoDB

- Key-value / document store. No SQL, no JOINs, no ad-hoc queries.
- You can only query by partition key + sort key, or by GSIs you defined upfront.
- Single-table design: one table, multiple entity types, keys encode access patterns. Every query is a single partition key lookup — O(1) regardless of table size.
- Changing access patterns = table redesign. Plan upfront.
- Cost: cheap at very large or very small scale. Can be expensive at medium scale vs RDS.

## Scaling — Do This In Order

Most systems never get past step 2 or 3. Don't skip ahead.

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

**Read replicas** copy data from the primary. They add read throughput but not write throughput. Replicas lag 10-100 ms (async) — a user might not see their own write immediately. For most reads this is fine; for "show balance after transfer" read from primary.

**Caching** (covered in lesson 04) is often more effective than replicas — Redis at 0.5 ms vs DB at 5 ms, and it removes load rather than just distributing it.

**Sharding** splits data across independent DB instances by a shard key. Hard because:

- Cross-shard queries must hit all shards and merge
- JOINs across shards are impossible
- Rebalancing requires moving data
- Distributed transactions (2PC) are slow and complex

Best shard key: one where most queries hit only one shard. For web apps, usually `user_id` or `tenant_id`.

## Key Mental Models

1. **Reads = find + return. Writes = WAL + all indexes + locks + replicate.** More indexes = wider gap.
2. **Missing index is the #1 performance problem.** Check `EXPLAIN ANALYZE` before anything else.
3. **Postgres needs PgBouncer.** Process-per-connection is expensive. MySQL is more forgiving.
4. **Match the database to your access patterns.** Postgres if unsure. DynamoDB if you know exactly how you'll query at scale. MongoDB if your data is naturally document-shaped.
5. **Scale in order: optimise → vertical → replicas → cache → shard.** Each step is 10x harder. Sharding is a last resort.
