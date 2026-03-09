# Query Analysis

## EXPLAIN Basics

```sql
-- Query plan without executing
EXPLAIN SELECT * FROM orders WHERE customer_id = 42;

-- Query plan WITH actual execution stats (runs the query)
EXPLAIN ANALYZE SELECT * FROM orders WHERE customer_id = 42;

-- Full detail output
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT) SELECT * FROM orders WHERE customer_id = 42;

-- JSON format (useful for tools like explain.dalibo.com)
EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) SELECT * FROM orders WHERE customer_id = 42;
```

> **Warning**: `EXPLAIN ANALYZE` executes the query. Wrap mutating queries in a transaction and roll back:
>
> ```sql
> BEGIN;
> EXPLAIN ANALYZE DELETE FROM orders WHERE created_at < '2020-01-01';
> ROLLBACK;
> ```

## Reading EXPLAIN Output

### Key Fields

| Field                      | Meaning                                    |
| -------------------------- | ------------------------------------------ |
| `cost=0.00..35.50`         | Startup cost..total cost (arbitrary units) |
| `rows=1000`                | Estimated rows returned                    |
| `width=72`                 | Average row size in bytes                  |
| `actual time=0.015..0.420` | Actual startup..total time in ms           |
| `actual rows=950`          | Actual rows returned                       |
| `loops=1`                  | Times this node was executed               |
| `Buffers: shared hit=128`  | Pages read from cache                      |
| `Buffers: shared read=42`  | Pages read from disk                       |

### Scan Types (Best to Worst, Generally)

| Scan              | When Used                                         | Notes                                                 |
| ----------------- | ------------------------------------------------- | ----------------------------------------------------- |
| Index Only Scan   | All columns in index                              | Best case, no table access needed                     |
| Index Scan        | Selective filter with index                       | Good for low-cardinality result sets                  |
| Bitmap Index Scan | Multiple index conditions or moderate selectivity | Combines indexes, then fetches                        |
| Seq Scan          | No useful index or large % of table               | Fine for small tables, bad for large filtered queries |

### Join Types

| Join        | Description                                                           |
| ----------- | --------------------------------------------------------------------- |
| Nested Loop | For each outer row, scan inner; good for small result sets            |
| Hash Join   | Builds hash table of one side; good for large unsorted joins          |
| Merge Join  | Merges two sorted inputs; good when both sides are pre-sorted/indexed |

## Identifying Problems

### Large Row Estimate Mismatches

```
Seq Scan on orders (cost=0.00..25000.00 rows=5 ...) (actual rows=450000 ...)
```

Estimated 5 rows, got 450k. Fix: run `ANALYZE orders;` to update statistics.

### Sequential Scans on Large Tables

```sql
-- Find tables with high sequential scan ratios
SELECT relname, seq_scan, idx_scan,
       round(seq_scan::numeric / GREATEST(seq_scan + idx_scan, 1) * 100, 1) AS seq_pct
FROM pg_stat_user_tables
WHERE seq_scan + idx_scan > 0
ORDER BY seq_pct DESC;
```

### Slow Sorts / Disk Sorts

```
Sort Method: external merge  Disk: 102400kB
```

`work_mem` too low for this query. Either increase `work_mem` globally or per-session:

```sql
SET work_mem = '256MB';  -- per session
```

### Lossy Bitmap Heap Scans

```
Bitmap Heap Scan ... Recheck Cond ...  Lossy: true
```

`work_mem` too low to hold the full bitmap. Same fix as above.

## pg_stat_statements

Essential extension for finding slow queries across the entire workload.

```sql
-- Enable (requires restart on self-managed, pre-enabled on RDS)
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- Top queries by total execution time
SELECT round(total_exec_time::numeric, 2) AS total_ms,
       calls,
       round(mean_exec_time::numeric, 2) AS avg_ms,
       round((100 * total_exec_time / sum(total_exec_time) OVER ())::numeric, 2) AS pct,
       query
FROM pg_stat_statements
ORDER BY total_exec_time DESC
LIMIT 20;

-- Top queries by average time (find individually slow queries)
SELECT round(mean_exec_time::numeric, 2) AS avg_ms,
       calls,
       query
FROM pg_stat_statements
WHERE calls > 10
ORDER BY mean_exec_time DESC
LIMIT 20;

-- Queries with the worst hit ratio (most disk reads)
SELECT query, calls,
       shared_blks_hit,
       shared_blks_read,
       round(shared_blks_hit::numeric / GREATEST(shared_blks_hit + shared_blks_read, 1) * 100, 2) AS hit_pct
FROM pg_stat_statements
ORDER BY hit_pct ASC
LIMIT 20;

-- Reset stats (do periodically to get fresh data)
SELECT pg_stat_statements_reset();
```

## Slow Query Logging

```
# postgresql.conf (or RDS parameter group)
log_min_duration_statement = 1000   # log queries taking > 1s (in ms)
log_statement = 'none'             # don't log all statements, just slow ones
```

On RDS: set `log_min_duration_statement` in the parameter group.

## auto_explain

Automatically logs EXPLAIN output for slow queries. Useful for catching plans without manually running EXPLAIN.

```
# postgresql.conf
shared_preload_libraries = 'auto_explain'
auto_explain.log_min_duration = '3s'
auto_explain.log_analyze = true
auto_explain.log_buffers = true
auto_explain.log_format = 'json'
```

On RDS: add `auto_explain` to `shared_preload_libraries` in the parameter group.

## Common Antipatterns

### SELECT \*

Prevents Index Only Scans. Always select only the columns you need.

### NOT IN with NULLs

```sql
-- Bad: if subquery returns any NULL, entire result is empty
SELECT * FROM a WHERE id NOT IN (SELECT id FROM b);

-- Good: use NOT EXISTS
SELECT * FROM a WHERE NOT EXISTS (SELECT 1 FROM b WHERE b.id = a.id);
```

### Functions on Indexed Columns

```sql
-- Bad: index on created_at won't be used
SELECT * FROM orders WHERE date(created_at) = '2024-01-01';

-- Good: use range
SELECT * FROM orders WHERE created_at >= '2024-01-01' AND created_at < '2024-01-02';
```

### OR Across Different Columns

```sql
-- Bad: often results in Seq Scan
SELECT * FROM orders WHERE customer_id = 42 OR product_id = 99;

-- Good: use UNION
SELECT * FROM orders WHERE customer_id = 42
UNION
SELECT * FROM orders WHERE product_id = 99;
```

### Implicit Casts

```sql
-- Bad: if user_id is integer, this prevents index use
SELECT * FROM users WHERE user_id = '42';

-- Good: match the type
SELECT * FROM users WHERE user_id = 42;
```
