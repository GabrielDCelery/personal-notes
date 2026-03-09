# Vacuum

## Why Vacuum Exists

PostgreSQL uses MVCC — updates and deletes don't remove old row versions immediately. These "dead tuples" accumulate and waste space, slow down scans, and can cause transaction ID wraparound. VACUUM reclaims this space.

## Manual Vacuum

```sql
-- Standard vacuum (reclaims space for reuse, doesn't return to OS)
VACUUM orders;

-- Verbose output
VACUUM VERBOSE orders;

-- Vacuum + update planner statistics
VACUUM ANALYZE orders;

-- Full vacuum (rewrites entire table, reclaims space to OS, requires ACCESS EXCLUSIVE lock)
VACUUM FULL orders;

-- Vacuum entire database
VACUUM;
```

> **Warning**: `VACUUM FULL` locks the table for the entire duration. For large tables, use `pg_repack` instead (see below).

## Autovacuum

Autovacuum runs in the background and handles vacuuming automatically. The goal is to tune it, not disable it.

### Check Autovacuum Status

```sql
-- Tables most in need of vacuum
SELECT relname,
       n_dead_tup,
       n_live_tup,
       round(n_dead_tup::numeric / GREATEST(n_live_tup, 1) * 100, 1) AS dead_pct,
       last_vacuum,
       last_autovacuum
FROM pg_stat_user_tables
ORDER BY n_dead_tup DESC
LIMIT 20;

-- Check if autovacuum is currently running
SELECT pid, datname, relid::regclass, phase, heap_blks_total, heap_blks_scanned, heap_blks_vacuumed
FROM pg_stat_progress_vacuum;
```

### Global Configuration

```
# postgresql.conf
autovacuum = on                            # never turn this off
autovacuum_max_workers = 3                 # increase for many tables
autovacuum_naptime = 60                    # seconds between autovacuum runs

# Thresholds — vacuum triggers when: dead tuples > threshold + scale_factor * live tuples
autovacuum_vacuum_threshold = 50           # base dead tuple count
autovacuum_vacuum_scale_factor = 0.2       # 20% of live tuples
autovacuum_analyze_threshold = 50
autovacuum_analyze_scale_factor = 0.1      # 10% of live tuples

# Cost-based throttling (prevents autovacuum from using too much I/O)
autovacuum_vacuum_cost_delay = 2           # ms to sleep after hitting cost limit
autovacuum_vacuum_cost_limit = 200         # cost units before sleeping
```

### Per-Table Overrides

For high-churn tables, tighten the thresholds:

```sql
-- Aggressive autovacuum for a hot table
ALTER TABLE orders SET (
  autovacuum_vacuum_threshold = 100,
  autovacuum_vacuum_scale_factor = 0.05,     -- trigger at 5% dead tuples
  autovacuum_analyze_threshold = 100,
  autovacuum_analyze_scale_factor = 0.02,
  autovacuum_vacuum_cost_delay = 0           -- no throttling for this table
);

-- Check current per-table settings
SELECT relname, reloptions
FROM pg_class
WHERE reloptions IS NOT NULL AND relkind = 'r';

-- Reset to global defaults
ALTER TABLE orders RESET (
  autovacuum_vacuum_threshold,
  autovacuum_vacuum_scale_factor
);
```

### Autovacuum for Large Tables

Default `scale_factor = 0.2` means a 100M row table waits for 20M dead tuples before triggering. Lower it:

```sql
ALTER TABLE big_table SET (
  autovacuum_vacuum_scale_factor = 0.01,   -- trigger at 1%
  autovacuum_vacuum_threshold = 10000
);
```

## Bloat Detection

### Table Bloat

```sql
-- Estimate table bloat using pgstattuple extension
CREATE EXTENSION IF NOT EXISTS pgstattuple;

SELECT * FROM pgstattuple('orders');
-- Look at dead_tuple_percent and free_space

-- Quick estimate without extension
SELECT relname,
       pg_size_pretty(pg_total_relation_size(relid)) AS total_size,
       n_dead_tup,
       n_live_tup,
       round(n_dead_tup::numeric / GREATEST(n_live_tup + n_dead_tup, 1) * 100, 1) AS dead_pct
FROM pg_stat_user_tables
ORDER BY n_dead_tup DESC;
```

### Index Bloat

```sql
-- Index bloat estimate using pgstattuple
SELECT * FROM pgstatindex('orders_pkey');
-- Look at avg_leaf_density (< 50% means significant bloat)
```

## Reclaiming Space

### VACUUM (Standard)

Marks dead tuple space as reusable but does not shrink the file. Good enough for most cases.

### VACUUM FULL

Rewrites the entire table. Reclaims space to OS. Requires exclusive lock — downtime.

```sql
VACUUM FULL orders;
```

### pg_repack (Online Alternative to VACUUM FULL)

Rebuilds tables and indexes without heavy locking.

```sh
# Install
sudo apt install postgresql-16-repack   # Debian/Ubuntu

# Repack a table
pg_repack -d mydb -t orders

# Repack only indexes of a table
pg_repack -d mydb -t orders --only-indexes

# Repack entire database
pg_repack -d mydb
```

### REINDEX

Rebuilds bloated indexes. Less disruptive than VACUUM FULL.

```sql
-- Single index
REINDEX INDEX orders_customer_id_idx;

-- All indexes on a table
REINDEX TABLE orders;

-- Concurrent reindex (PG 12+, no lock)
REINDEX INDEX CONCURRENTLY orders_customer_id_idx;
```

## Transaction ID Wraparound Prevention

PostgreSQL uses 32-bit transaction IDs. Without vacuum, the database will shut down to prevent data loss at ~2 billion transactions.

### Monitor Wraparound Risk

```sql
-- Tables closest to wraparound
SELECT relname,
       age(relfrozenxid) AS xid_age,
       pg_size_pretty(pg_total_relation_size(oid)) AS size
FROM pg_class
WHERE relkind = 'r'
ORDER BY age(relfrozenxid) DESC
LIMIT 20;

-- Database-level check
SELECT datname, age(datfrozenxid) AS xid_age
FROM pg_database
ORDER BY xid_age DESC;
```

> **Alert threshold**: `age(relfrozenxid)` above 200 million warrants investigation. Autovacuum triggers aggressive freeze at `autovacuum_freeze_max_age` (default 200M).

### Emergency Wraparound Vacuum

If autovacuum isn't keeping up:

```sql
-- Force aggressive vacuum with freezing
VACUUM FREEZE orders;

-- For the whole database
VACUUM FREEZE;
```

## RDS Vacuum Considerations

- Autovacuum is enabled by default and cannot be disabled
- Tune via parameter groups: `rds.adaptive_autovacuum` is on by default (auto-tunes cost_delay and cost_limit)
- Monitor with CloudWatch: `MaximumUsedTransactionIDs` — alert if approaching 1 billion
- `rds.force_autovacuum_logging_level = log` to see autovacuum activity in RDS logs
- Enhanced Monitoring shows vacuum I/O impact
- Large RDS instances: increase `maintenance_work_mem` (up to 2GB) to speed up vacuum

### RDS Parameter Group Settings

```
# Recommended starting points for high-churn RDS databases
autovacuum_max_workers = 5
autovacuum_vacuum_cost_delay = 2
autovacuum_vacuum_cost_limit = 1000
maintenance_work_mem = '1GB'
```
