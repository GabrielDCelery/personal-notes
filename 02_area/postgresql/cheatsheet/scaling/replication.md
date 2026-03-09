# Replication

## Streaming Replication (Physical)

Byte-for-byte copy of the entire cluster. Replica is read-only.

### Setup on Primary

```
# postgresql.conf
wal_level = replica
max_wal_senders = 10               # max concurrent replication connections
wal_keep_size = '1GB'              # retain WAL for slow replicas (PG 13+)
max_replication_slots = 10
```

```
# pg_hba.conf — allow replication connections
host replication replicator 10.0.0.0/24 scram-sha-256
```

```sql
-- Create replication user
CREATE ROLE replicator WITH REPLICATION LOGIN PASSWORD 'secret';

-- Create replication slot (prevents WAL cleanup before replica consumes it)
SELECT pg_create_physical_replication_slot('replica1');
```

### Setup on Replica

```sh
# Base backup from primary
pg_basebackup -h primary-host -U replicator -D /var/lib/postgresql/16/main \
  -Fp -Xs -P -R --slot=replica1

# -R flag creates standby.signal and populates primary_conninfo in postgresql.auto.conf
```

```
# postgresql.auto.conf (created by -R flag)
primary_conninfo = 'host=primary-host user=replicator password=secret'
primary_slot_name = 'replica1'
```

```sh
# Start replica
pg_ctl -D /var/lib/postgresql/16/main start
```

### Verify Replication

```sql
-- On primary: check connected replicas
SELECT client_addr, state, sent_lsn, write_lsn, flush_lsn, replay_lsn,
       pg_wal_lsn_diff(sent_lsn, replay_lsn) AS replay_lag_bytes
FROM pg_stat_replication;

-- On replica: check replication status
SELECT status, received_lsn, latest_end_lsn,
       now() - pg_last_xact_replay_timestamp() AS replay_lag
FROM pg_stat_wal_receiver;

-- On replica: confirm it's in recovery mode
SELECT pg_is_in_recovery();
```

### Synchronous Replication

Guarantees data is written to replica before committing on primary. Adds latency.

```
# postgresql.conf on primary
synchronous_standby_names = 'FIRST 1 (replica1, replica2)'
synchronous_commit = on
```

| Mode           | Guarantee                             | Latency Impact |
| -------------- | ------------------------------------- | -------------- |
| `off`          | No durability guarantee               | None           |
| `local`        | Written to primary WAL                | None           |
| `remote_write` | Written to replica OS cache           | Low            |
| `on`           | Flushed to replica disk               | Moderate       |
| `remote_apply` | Applied on replica (visible to reads) | Highest        |

### Promote Replica to Primary

```sh
# Promote (irreversible — replica becomes standalone primary)
pg_ctl promote -D /var/lib/postgresql/16/main

# Or via SQL
SELECT pg_promote();
```

## Logical Replication

Replicates specific tables at the row level. Both sides are independent PostgreSQL instances — the subscriber can have its own tables, indexes, and even write to replicated tables.

### Use Cases

- Selective table replication
- Cross-version replication (PG 14 → PG 16)
- Data integration between databases
- Zero-downtime major version upgrades

### Setup

```
# postgresql.conf on publisher
wal_level = logical
max_replication_slots = 10
max_wal_senders = 10
```

```sql
-- On publisher: create publication
CREATE PUBLICATION my_pub FOR TABLE orders, customers;

-- Publish all tables
CREATE PUBLICATION my_pub FOR ALL TABLES;

-- Publish with row filter (PG 15+)
CREATE PUBLICATION my_pub FOR TABLE orders WHERE (region = 'eu');

-- Publish specific columns (PG 15+)
CREATE PUBLICATION my_pub FOR TABLE orders (id, customer_id, total, created_at);
```

```sql
-- On subscriber: create subscription
CREATE SUBSCRIPTION my_sub
  CONNECTION 'host=publisher-host dbname=mydb user=replicator password=secret'
  PUBLICATION my_pub;
```

### Manage Publications and Subscriptions

```sql
-- Add/remove tables from publication
ALTER PUBLICATION my_pub ADD TABLE products;
ALTER PUBLICATION my_pub DROP TABLE old_table;

-- Refresh subscription after publication changes
ALTER SUBSCRIPTION my_sub REFRESH PUBLICATION;

-- Pause/resume subscription
ALTER SUBSCRIPTION my_sub DISABLE;
ALTER SUBSCRIPTION my_sub ENABLE;

-- Drop subscription (stops replication)
DROP SUBSCRIPTION my_sub;

-- Monitor subscription status
SELECT subname, received_lsn, latest_end_lsn,
       latest_end_time
FROM pg_stat_subscription;

-- Check replication slots on publisher
SELECT slot_name, active, confirmed_flush_lsn
FROM pg_replication_slots;
```

### Limitations

- DDL is not replicated (schema changes must be applied manually on both sides)
- Sequences are not replicated
- Large objects are not replicated
- TRUNCATE is replicated (PG 11+)
- Tables must have a primary key or REPLICA IDENTITY set
- Conflicts (e.g., duplicate key) must be resolved manually

### Replica Identity

Logical replication needs to identify rows for UPDATE/DELETE.

```sql
-- Default: use primary key (works for most tables)
-- If no primary key, set replica identity to a unique index
ALTER TABLE orders REPLICA IDENTITY USING INDEX idx_orders_unique;

-- Or use full row (slower, compares all columns)
ALTER TABLE orders REPLICA IDENTITY FULL;

-- Check current setting
SELECT relname, relreplident
FROM pg_class
WHERE relname = 'orders';
-- d = default (PK), f = full, i = index, n = nothing
```

## Replication Slots

Prevent WAL from being removed before all consumers have received it.

```sql
-- Physical slot (streaming replication)
SELECT pg_create_physical_replication_slot('replica1');

-- Logical slot (logical replication)
SELECT pg_create_logical_replication_slot('my_slot', 'pgoutput');

-- List slots
SELECT slot_name, slot_type, active, restart_lsn,
       pg_wal_lsn_diff(pg_current_wal_lsn(), restart_lsn) AS lag_bytes
FROM pg_replication_slots;

-- Drop unused slot (CRITICAL — inactive slots prevent WAL cleanup and fill disk)
SELECT pg_drop_replication_slot('old_slot');
```

> **Warning**: An inactive replication slot will cause WAL to accumulate indefinitely, eventually filling the disk. Monitor `pg_replication_slots` for `active = false` slots with growing lag.

### Slot Safety on RDS

```
# RDS parameter to auto-drop stale slots (prevents disk full)
rds.logical_replication = 1
max_slot_wal_keep_size = '100GB'    # PG 13+, limits WAL retention per slot
```

## Failover with repmgr

repmgr automates streaming replication management and failover.

```sh
# Register primary
repmgr primary register

# Register standby
repmgr standby register

# Check cluster status
repmgr cluster show

# Manual switchover (planned)
repmgr standby switchover --siblings-follow

# Automatic failover (via repmgrd daemon)
repmgrd --daemonize
```

## Streaming vs Logical Comparison

| Feature              | Streaming (Physical) | Logical                      |
| -------------------- | -------------------- | ---------------------------- |
| Replication unit     | Entire cluster       | Per-table                    |
| Replica writable     | No                   | Yes (non-replicated tables)  |
| Cross-version        | No                   | Yes                          |
| DDL replicated       | Yes (automatic)      | No (manual)                  |
| Performance overhead | Low                  | Moderate                     |
| Failover             | Promote replica      | Not designed for failover    |
| Use case             | HA / read scaling    | Data integration / migration |
