# Monitoring

## Active Queries

```sql
-- Currently running queries
SELECT pid, usename, datname, state,
       now() - query_start AS duration,
       wait_event_type, wait_event,
       left(query, 100) AS query
FROM pg_stat_activity
WHERE state != 'idle'
ORDER BY duration DESC;

-- Long-running queries (> 1 minute)
SELECT pid, usename, datname,
       now() - query_start AS duration,
       query
FROM pg_stat_activity
WHERE state = 'active'
  AND now() - query_start > interval '1 minute'
ORDER BY duration DESC;

-- Queries waiting on locks
SELECT pid, usename,
       now() - query_start AS duration,
       wait_event_type, wait_event,
       query
FROM pg_stat_activity
WHERE wait_event_type IS NOT NULL
  AND state = 'active';
```

## Locks

```sql
-- All current locks with blocking info
SELECT blocked.pid AS blocked_pid,
       blocked.usename AS blocked_user,
       blocking.pid AS blocking_pid,
       blocking.usename AS blocking_user,
       blocked.query AS blocked_query,
       blocking.query AS blocking_query
FROM pg_stat_activity blocked
JOIN pg_locks bl ON bl.pid = blocked.pid
JOIN pg_locks kl ON kl.locktype = bl.locktype
  AND kl.database IS NOT DISTINCT FROM bl.database
  AND kl.relation IS NOT DISTINCT FROM bl.relation
  AND kl.page IS NOT DISTINCT FROM bl.page
  AND kl.tuple IS NOT DISTINCT FROM bl.tuple
  AND kl.virtualxid IS NOT DISTINCT FROM bl.virtualxid
  AND kl.transactionid IS NOT DISTINCT FROM bl.transactionid
  AND kl.pid != bl.pid
  AND kl.granted
JOIN pg_stat_activity blocking ON blocking.pid = kl.pid
WHERE NOT bl.granted;

-- Simpler view: who is blocking whom (PG 14+)
SELECT * FROM pg_blocking_pids(<blocked_pid>);

-- Lock types on a specific table
SELECT locktype, mode, granted, pid
FROM pg_locks
WHERE relation = 'orders'::regclass;

-- Kill a blocking query
SELECT pg_terminate_backend(<blocking_pid>);
```

### Lock Timeout

```sql
-- Set lock timeout to avoid waiting indefinitely
SET lock_timeout = '5s';

-- Per-transaction
BEGIN;
SET LOCAL lock_timeout = '5s';
ALTER TABLE orders ADD COLUMN new_col text;
COMMIT;
```

## Database Size

```sql
-- Database sizes
SELECT datname, pg_size_pretty(pg_database_size(datname)) AS size
FROM pg_database
ORDER BY pg_database_size(datname) DESC;

-- Table sizes (including indexes and TOAST)
SELECT relname,
       pg_size_pretty(pg_total_relation_size(relid)) AS total,
       pg_size_pretty(pg_table_size(relid)) AS table,
       pg_size_pretty(pg_indexes_size(relid)) AS indexes
FROM pg_stat_user_tables
ORDER BY pg_total_relation_size(relid) DESC
LIMIT 20;

-- Largest tables by growth (compare snapshots over time)
SELECT relname, n_tup_ins, n_tup_upd, n_tup_del,
       n_tup_ins - n_tup_del AS net_growth
FROM pg_stat_user_tables
ORDER BY n_tup_ins DESC
LIMIT 20;
```

## Table and Index Health

```sql
-- Table hit ratio (should be > 99% for hot tables)
SELECT relname,
       round(heap_blks_hit::numeric / GREATEST(heap_blks_hit + heap_blks_read, 1) * 100, 2) AS hit_pct,
       heap_blks_hit, heap_blks_read
FROM pg_statio_user_tables
ORDER BY heap_blks_read DESC
LIMIT 20;

-- Index hit ratio
SELECT relname, indexrelname,
       round(idx_blks_hit::numeric / GREATEST(idx_blks_hit + idx_blks_read, 1) * 100, 2) AS hit_pct,
       idx_blks_hit, idx_blks_read
FROM pg_statio_user_indexes
ORDER BY idx_blks_read DESC
LIMIT 20;

-- Tables with most sequential scans (missing indexes?)
SELECT relname, seq_scan, idx_scan, n_live_tup
FROM pg_stat_user_tables
WHERE seq_scan > 0
ORDER BY seq_scan DESC
LIMIT 20;

-- Dead tuples (need vacuum)
SELECT relname, n_dead_tup, n_live_tup,
       round(n_dead_tup::numeric / GREATEST(n_live_tup, 1) * 100, 1) AS dead_pct,
       last_autovacuum
FROM pg_stat_user_tables
ORDER BY n_dead_tup DESC
LIMIT 20;
```

## Replication Monitoring

```sql
-- On primary: check replication status
SELECT client_addr, state, sent_lsn, write_lsn, flush_lsn, replay_lsn,
       now() - pg_last_xact_replay_timestamp() AS replay_lag,
       pg_wal_lsn_diff(sent_lsn, replay_lsn) AS replay_lag_bytes
FROM pg_stat_replication;

-- On replica: check how far behind
SELECT now() - pg_last_xact_replay_timestamp() AS replay_lag;

-- WAL generation rate
SELECT pg_wal_lsn_diff(pg_current_wal_lsn(), '0/0') AS total_wal_bytes,
       pg_size_pretty(pg_wal_lsn_diff(pg_current_wal_lsn(), '0/0')) AS total_wal;
```

## Transaction ID Wraparound

```sql
-- Monitor XID age (alert if > 200M, panic at 2B)
SELECT datname, age(datfrozenxid) AS xid_age,
       round(age(datfrozenxid)::numeric / 2147483647 * 100, 2) AS pct_to_wraparound
FROM pg_database
ORDER BY xid_age DESC;

-- Per-table XID age
SELECT relname, age(relfrozenxid) AS xid_age
FROM pg_class
WHERE relkind = 'r'
ORDER BY xid_age DESC
LIMIT 10;
```

## System-Level Metrics

```sql
-- Checkpoint activity
SELECT checkpoints_timed, checkpoints_req,
       buffers_checkpoint, buffers_clean, buffers_backend,
       pg_size_pretty(buffers_checkpoint * 8192) AS data_written_checkpoint
FROM pg_stat_bgwriter;

-- WAL activity (PG 14+)
SELECT * FROM pg_stat_wal;

-- Connection usage
SELECT count(*) AS current,
       (SELECT setting::int FROM pg_settings WHERE name = 'max_connections') AS max,
       round(count(*)::numeric / (SELECT setting::int FROM pg_settings WHERE name = 'max_connections') * 100, 1) AS pct
FROM pg_stat_activity;
```

## RDS Monitoring

### Key CloudWatch Metrics

| Metric                         | What to Watch        | Alert Threshold                    |
| ------------------------------ | -------------------- | ---------------------------------- |
| `CPUUtilization`               | Sustained high CPU   | > 80% for 5 min                    |
| `FreeableMemory`               | Available RAM        | < 10% of instance memory           |
| `FreeStorageSpace`             | Disk space remaining | < 20%                              |
| `ReadIOPS` / `WriteIOPS`       | I/O pressure         | Approaching provisioned IOPS limit |
| `ReadLatency` / `WriteLatency` | Storage performance  | > 10ms sustained                   |
| `DatabaseConnections`          | Connection count     | > 80% of max_connections           |
| `ReplicaLag`                   | Replication delay    | > 30 seconds                       |
| `MaximumUsedTransactionIDs`    | XID wraparound risk  | > 1 billion                        |
| `SwapUsage`                    | Memory pressure      | Any swap usage                     |

### Performance Insights

```sh
# Enable Performance Insights
aws rds modify-db-instance \
  --db-instance-identifier mydb \
  --enable-performance-insights \
  --performance-insights-retention-period 731 \
  --apply-immediately
```

Performance Insights shows:

- **DB Load** — average active sessions broken down by wait events, SQL, users, hosts
- **Top SQL** — queries consuming the most database time
- **Wait events** — what the database is waiting on (CPU, I/O, locks, IPC)

### Enhanced Monitoring

```sh
# Enable Enhanced Monitoring (OS-level metrics at 1s granularity)
aws rds modify-db-instance \
  --db-instance-identifier mydb \
  --monitoring-interval 1 \
  --monitoring-role-arn arn:aws:iam::111111111111:role/rds-monitoring-role \
  --apply-immediately
```

Provides: CPU per process, memory breakdown, disk I/O, file system usage, network traffic.

### RDS Log Access

```sh
# List available logs
aws rds describe-db-log-files --db-instance-identifier mydb

# Download a log file
aws rds download-db-log-file-portion \
  --db-instance-identifier mydb \
  --log-file-name error/postgresql.log.2026-03-09-12 \
  --output text
```

### Useful Parameter Group Settings for Monitoring

```
log_min_duration_statement = 1000          # log queries > 1s
log_connections = 1                         # log connection attempts
log_disconnections = 1                      # log disconnections
log_lock_waits = 1                          # log lock waits > deadlock_timeout
log_checkpoints = 1                         # log checkpoint activity
log_temp_files = 0                          # log all temp file usage
rds.force_autovacuum_logging_level = log    # log autovacuum activity
```
