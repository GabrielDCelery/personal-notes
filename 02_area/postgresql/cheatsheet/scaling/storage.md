# Storage

## Disk Usage

### Database Size

```sql
SELECT datname, pg_size_pretty(pg_database_size(datname)) AS size
FROM pg_database
ORDER BY pg_database_size(datname) DESC;
```

### Table Size Breakdown

```sql
-- Total size (table + indexes + TOAST)
SELECT relname,
       pg_size_pretty(pg_total_relation_size(relid)) AS total,
       pg_size_pretty(pg_table_size(relid)) AS table,
       pg_size_pretty(pg_indexes_size(relid)) AS indexes,
       pg_size_pretty(pg_total_relation_size(relid) - pg_table_size(relid) - pg_indexes_size(relid)) AS toast
FROM pg_stat_user_tables
ORDER BY pg_total_relation_size(relid) DESC
LIMIT 20;

-- Size of a specific table with details
SELECT pg_size_pretty(pg_relation_size('orders')) AS main,
       pg_size_pretty(pg_table_size('orders')) AS table_with_toast,
       pg_size_pretty(pg_indexes_size('orders')) AS indexes,
       pg_size_pretty(pg_total_relation_size('orders')) AS total;
```

### Schema Size

```sql
SELECT schemaname,
       pg_size_pretty(sum(pg_total_relation_size(relid))) AS total
FROM pg_stat_user_tables
GROUP BY schemaname
ORDER BY sum(pg_total_relation_size(relid)) DESC;
```

### Largest Indexes

```sql
SELECT tablename, indexname,
       pg_size_pretty(pg_relation_size(indexname::regclass)) AS size
FROM pg_indexes
WHERE schemaname = 'public'
ORDER BY pg_relation_size(indexname::regclass) DESC
LIMIT 20;
```

## TOAST (The Oversized-Attribute Storage Technique)

PostgreSQL stores large field values (> ~2KB) in a separate TOAST table, compressed.

### Check TOAST Usage

```sql
-- TOAST table size per table
SELECT c.relname AS table,
       pg_size_pretty(pg_relation_size(c.reltoastrelid)) AS toast_size
FROM pg_class c
WHERE c.reltoastrelid != 0
  AND c.relkind = 'r'
ORDER BY pg_relation_size(c.reltoastrelid) DESC
LIMIT 20;
```

### TOAST Strategies

```sql
-- Check current strategy for each column
SELECT attname, attstorage
FROM pg_attribute
WHERE attrelid = 'orders'::regclass AND attnum > 0;
-- p = plain (no TOAST), e = external (no compression), m = main (compress, avoid TOAST), x = extended (compress + TOAST)

-- Change strategy for a column
ALTER TABLE orders ALTER COLUMN payload SET STORAGE EXTERNAL;  -- don't compress JSONB (faster reads if already small)
ALTER TABLE orders ALTER COLUMN description SET STORAGE MAIN;  -- prefer inline, compress
```

### TOAST Compression (PG 14+)

```sql
-- Use LZ4 compression instead of default pglz (faster)
ALTER TABLE orders ALTER COLUMN payload SET COMPRESSION lz4;

-- Set default for new columns
SET default_toast_compression = 'lz4';
```

## Tablespaces

Store tables/indexes on different filesystems or disks.

```sql
-- Create tablespace on a different disk
CREATE TABLESPACE fast_storage LOCATION '/mnt/nvme/pg_data';

-- Create table in specific tablespace
CREATE TABLE hot_data (...) TABLESPACE fast_storage;

-- Move existing table
ALTER TABLE orders SET TABLESPACE fast_storage;

-- Move an index
ALTER INDEX idx_orders_customer_id SET TABLESPACE fast_storage;

-- Set default tablespace for a database
ALTER DATABASE mydb SET TABLESPACE fast_storage;

-- List tablespaces
SELECT spcname, pg_size_pretty(pg_tablespace_size(spcname)) AS size
FROM pg_tablespace;
```

> On RDS, you cannot create custom tablespaces — storage is managed by AWS.

## Bloat Management

### Detect Table Bloat

```sql
-- Using pgstattuple (most accurate)
CREATE EXTENSION IF NOT EXISTS pgstattuple;

SELECT * FROM pgstattuple('orders');
-- dead_tuple_percent: percentage of dead space
-- free_percent: percentage of reusable free space

-- Estimate without extension
SELECT relname,
       n_dead_tup,
       n_live_tup,
       round(n_dead_tup::numeric / GREATEST(n_live_tup + n_dead_tup, 1) * 100, 1) AS dead_pct,
       pg_size_pretty(pg_total_relation_size(relid)) AS total_size
FROM pg_stat_user_tables
WHERE n_dead_tup > 10000
ORDER BY n_dead_tup DESC;
```

### Detect Index Bloat

```sql
SELECT indexrelid::regclass AS index,
       pg_size_pretty(pg_relation_size(indexrelid)) AS size,
       avg_leaf_density,
       leaf_fragmentation
FROM pgstatindex(indexrelid::regclass::text), pg_index
WHERE indrelid = 'orders'::regclass;
-- avg_leaf_density < 50% indicates significant bloat
```

### Fix Bloat

```sql
-- Standard VACUUM: marks dead space as reusable (doesn't shrink file)
VACUUM orders;

-- VACUUM FULL: rewrites table, reclaims space to OS (ACCESS EXCLUSIVE lock)
VACUUM FULL orders;

-- REINDEX: rebuild bloated indexes
REINDEX INDEX CONCURRENTLY idx_orders_customer_id;
REINDEX TABLE CONCURRENTLY orders;
```

### pg_repack (Online Rebuild)

Rebuilds tables and indexes without heavy locking — preferred over VACUUM FULL.

```sh
# Install
sudo apt install postgresql-16-repack

# Repack a single table
pg_repack -d mydb -t orders

# Repack only indexes
pg_repack -d mydb -t orders --only-indexes

# Repack entire database
pg_repack -d mydb

# Dry run
pg_repack -d mydb -t orders --dry-run
```

Requirements:

- Table must have a primary key or unique index with NOT NULL
- Needs enough free disk space for a copy of the table
- Not available on RDS (use `VACUUM FULL` during maintenance window or consider Aurora)

## File System Layout

```
$PGDATA/
├── base/              # database files (one subdirectory per database OID)
│   ├── 1/             # template1
│   ├── 16384/         # user database
│   │   ├── 16385      # table or index file (relfilenode)
│   │   ├── 16385.1    # overflow segment (if > 1GB)
│   │   └── 16385_fsm  # free space map
│   │   └── 16385_vm   # visibility map
├── global/            # shared system catalogs
├── pg_wal/            # WAL files (critical, never delete manually)
├── pg_tblspc/         # symlinks to tablespace locations
├── pg_stat_tmp/       # temporary stats files
└── postgresql.conf
```

### WAL Disk Usage

```sql
-- Current WAL size
SELECT pg_size_pretty(sum(size))
FROM pg_ls_waldir();

-- WAL retention settings
SHOW max_wal_size;
SHOW min_wal_size;
SHOW wal_keep_size;

-- Check if replication slots are holding WAL
SELECT slot_name, active,
       pg_size_pretty(pg_wal_lsn_diff(pg_current_wal_lsn(), restart_lsn)) AS retained_wal
FROM pg_replication_slots;
```

## Data Compression

### Column Compression (PG 14+)

```sql
-- Per-column compression
ALTER TABLE events ALTER COLUMN payload SET COMPRESSION lz4;
ALTER TABLE events ALTER COLUMN description SET COMPRESSION pglz;    -- default, better ratio

-- Check compression used
SELECT attname, attcompression
FROM pg_attribute
WHERE attrelid = 'events'::regclass AND attnum > 0;
-- empty = default (pglz), l = lz4, p = pglz
```

### When to Use LZ4 vs pglz

| Aspect              | pglz (default)      | lz4                         |
| ------------------- | ------------------- | --------------------------- |
| Compression ratio   | Better              | Good                        |
| Compression speed   | Slow                | Fast                        |
| Decompression speed | Slow                | Fast                        |
| Use when            | Storage-constrained | Read-heavy, CPU-constrained |

## Monitoring Storage Growth

```sql
-- Table growth over time (requires periodic snapshots)
-- Compare pg_stat_user_tables.n_tup_ins, n_tup_upd, n_tup_del between snapshots

-- Current write activity
SELECT relname, n_tup_ins, n_tup_upd, n_tup_del,
       n_tup_ins + n_tup_upd + n_tup_del AS total_writes
FROM pg_stat_user_tables
ORDER BY total_writes DESC
LIMIT 20;

-- Estimate row size
SELECT avg(pg_column_size(t.*)) AS avg_row_bytes
FROM orders t
LIMIT 10000;
```
