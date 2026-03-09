# Configuration

## Memory

### shared_buffers

PostgreSQL's main cache for table and index data. Pages are read from disk into shared buffers.

```
# Rule of thumb: 25% of total RAM
shared_buffers = '4GB'    # on a 16GB server
```

- Too low: excessive disk reads
- Too high: leaves insufficient memory for OS page cache and work_mem
- Check effectiveness with cache hit ratio (should be > 99%):

```sql
SELECT round(
  sum(heap_blks_hit)::numeric / GREATEST(sum(heap_blks_hit + heap_blks_read), 1) * 100, 2
) AS hit_pct
FROM pg_statio_user_tables;
```

### work_mem

Memory per sort/hash operation per query. A single complex query can use multiple work_mem allocations.

```
# Default is 4MB — often too low for analytical queries
work_mem = '64MB'
```

- Too low: sorts and hash joins spill to disk (look for "Sort Method: external merge" in EXPLAIN)
- Too high: risk of OOM with many concurrent queries
- Formula: `available_ram / max_connections / 2` as a starting ceiling
- Can set per-session for heavy queries:

```sql
SET work_mem = '256MB';
-- run expensive query
RESET work_mem;
```

### maintenance_work_mem

Memory for maintenance operations: VACUUM, CREATE INDEX, ALTER TABLE.

```
# 5-10% of RAM, max ~2GB
maintenance_work_mem = '1GB'
```

Larger values speed up vacuum and index creation significantly.

### effective_cache_size

Not an allocation — tells the planner how much memory is available for caching (shared_buffers + OS page cache). Affects planner cost estimates for index scans.

```
# 50-75% of total RAM
effective_cache_size = '12GB'    # on a 16GB server
```

Too low: planner avoids index scans in favor of sequential scans.

### huge_pages

Reduces TLB misses for large shared_buffers allocations.

```
huge_pages = try    # use if available, fall back to regular pages
```

Requires OS-level configuration:

```sh
# Calculate needed huge pages (2MB each)
# shared_buffers = 4GB → 4096 / 2 = 2048 huge pages + some headroom
sudo sysctl -w vm.nr_hugepages=2200
```

## WAL (Write-Ahead Log)

### wal_buffers

Buffer for WAL data before writing to disk.

```
# Default auto-sizes to 1/32 of shared_buffers, capped at 64MB
# Usually fine at default, but set explicitly for large shared_buffers
wal_buffers = '64MB'
```

### max_wal_size / min_wal_size

Controls checkpoint frequency. Larger values = fewer checkpoints but longer recovery.

```
max_wal_size = '4GB'     # trigger checkpoint when WAL reaches this
min_wal_size = '1GB'     # recycle WAL files down to this
```

### checkpoint_completion_target

Spreads checkpoint I/O over time to avoid spikes.

```
checkpoint_completion_target = 0.9    # spread over 90% of checkpoint interval
```

### wal_compression

Compress WAL records to reduce I/O and WAL volume.

```
wal_compression = on    # PG 15+: lz4, zstd also available
```

## Connections

```
max_connections = 200              # keep low if using connection pooler
superuser_reserved_connections = 3 # emergency access slots
```

> With PgBouncer or RDS Proxy, keep `max_connections` moderate and let the pooler handle concurrency.

## Query Planner

### random_page_cost

Cost of a non-sequential disk read. Lowering it makes the planner prefer index scans.

```
random_page_cost = 1.1    # SSDs (default is 4.0, designed for spinning disks)
```

### effective_io_concurrency

Number of concurrent I/O operations for bitmap heap scans.

```
effective_io_concurrency = 200    # SSDs (default is 1)
```

### default_statistics_target

Controls how much data ANALYZE collects for planner statistics. Higher = more accurate estimates but slower ANALYZE.

```
default_statistics_target = 100    # default, increase to 500 for columns with skewed data
```

Per-column override:

```sql
ALTER TABLE orders ALTER COLUMN status SET STATISTICS 500;
ANALYZE orders;
```

## Parallel Query

```
max_parallel_workers_per_gather = 4     # workers per query node
max_parallel_workers = 8                # total parallel workers system-wide
max_parallel_maintenance_workers = 4    # for CREATE INDEX, VACUUM
parallel_tuple_cost = 0.01
parallel_setup_cost = 1000
min_parallel_table_scan_size = '8MB'
min_parallel_index_scan_size = '512kB'
```

## Logging (Performance-Related)

```
log_min_duration_statement = 1000       # log queries > 1s
log_checkpoints = on                    # log checkpoint activity
log_lock_waits = on                     # log lock waits > deadlock_timeout
log_temp_files = 0                      # log all temp file usage (work_mem spills)
log_autovacuum_min_duration = 0         # log all autovacuum activity
```

## Applying Changes

### Reload vs Restart

```sql
-- Check if a setting requires restart
SELECT name, setting, context
FROM pg_settings
WHERE name IN ('shared_buffers', 'work_mem', 'max_connections');
-- context = 'postmaster' → requires restart
-- context = 'user' or 'superuser' → reload or SET
```

```sh
# Reload configuration (no downtime)
sudo systemctl reload postgresql
# or
SELECT pg_reload_conf();

# Restart (required for postmaster-level settings)
sudo systemctl restart postgresql
```

### Settings Requiring Restart

| Setting                | Context    |
| ---------------------- | ---------- |
| `shared_buffers`       | postmaster |
| `max_connections`      | postmaster |
| `huge_pages`           | postmaster |
| `wal_buffers`          | postmaster |
| `max_worker_processes` | postmaster |
| `max_parallel_workers` | postmaster |

### Settings That Can Be Reloaded

| Setting                      | Context   |
| ---------------------------- | --------- |
| `work_mem`                   | user      |
| `maintenance_work_mem`       | user      |
| `effective_cache_size`       | user      |
| `random_page_cost`           | user      |
| `log_min_duration_statement` | superuser |
| `autovacuum_*`               | sighup    |
| `checkpoint_*`               | sighup    |

## Quick Reference by Server Size

| Setting                    | 4GB RAM   | 16GB RAM  | 64GB RAM  |
| -------------------------- | --------- | --------- | --------- |
| `shared_buffers`           | 1GB       | 4GB       | 16GB      |
| `effective_cache_size`     | 3GB       | 12GB      | 48GB      |
| `work_mem`                 | 16MB      | 64MB      | 256MB     |
| `maintenance_work_mem`     | 256MB     | 1GB       | 2GB       |
| `max_connections`          | 100       | 200       | 300       |
| `random_page_cost`         | 1.1 (SSD) | 1.1 (SSD) | 1.1 (SSD) |
| `effective_io_concurrency` | 200 (SSD) | 200 (SSD) | 200 (SSD) |
| `max_wal_size`             | 2GB       | 4GB       | 8GB       |
