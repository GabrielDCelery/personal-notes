# Backups

## Logical Backups (pg_dump / pg_restore)

### pg_dump

```sh
# Single database — custom format (compressed, supports parallel restore)
pg_dump -Fc -d mydb -f mydb.dump

# Single database — plain SQL
pg_dump -Fp -d mydb -f mydb.sql

# Schema only (no data)
pg_dump -Fc --schema-only -d mydb -f mydb_schema.dump

# Data only
pg_dump -Fc --data-only -d mydb -f mydb_data.dump

# Specific tables
pg_dump -Fc -d mydb -t orders -t customers -f partial.dump

# Specific schema
pg_dump -Fc -d mydb -n public -f public_schema.dump

# Exclude large tables
pg_dump -Fc -d mydb -T audit_log -T event_stream -f mydb_slim.dump

# Parallel dump (custom format only, uses 4 workers)
pg_dump -Fd -j 4 -d mydb -f mydb_dir/
```

### pg_restore

```sh
# Restore custom format dump
pg_restore -d mydb mydb.dump

# Restore into a new database
createdb mydb_restored
pg_restore -d mydb_restored mydb.dump

# Parallel restore (directory format)
pg_restore -d mydb -j 4 mydb_dir/

# Schema only
pg_restore --schema-only -d mydb mydb.dump

# Data only
pg_restore --data-only -d mydb mydb.dump

# Specific table
pg_restore -d mydb -t orders mydb.dump

# List contents of a dump (useful for selective restore)
pg_restore -l mydb.dump

# Clean (drop) objects before recreating
pg_restore --clean --if-exists -d mydb mydb.dump
```

### pg_dumpall

```sh
# All databases + globals (roles, tablespaces)
pg_dumpall -f full_cluster.sql

# Globals only (roles and tablespaces — use with per-db pg_dump)
pg_dumpall --globals-only -f globals.sql
```

## Physical Backups (pg_basebackup)

Full filesystem-level copy of the data directory. Required for point-in-time recovery.

```sh
# Basic backup
pg_basebackup -D /backup/base -Ft -z -P

# With WAL streaming (self-contained backup)
pg_basebackup -D /backup/base -Ft -z -P -Xs

# Backup to a specific host
pg_basebackup -h pg-primary -U replication -D /backup/base -Ft -z -P -Xs
```

### Replication Slot for Backup

```sql
-- Prevent WAL segments from being recycled during long backups
SELECT pg_create_physical_replication_slot('backup_slot');
```

```sh
pg_basebackup -D /backup/base -Ft -z -P -Xs --slot=backup_slot
```

## Point-in-Time Recovery (PITR)

Restore to any point in time using a base backup + WAL archive.

### 1. Set Up WAL Archiving

```
# postgresql.conf
wal_level = replica
archive_mode = on
archive_command = 'cp %p /archive/wal/%f'
```

### 2. Take a Base Backup

```sh
pg_basebackup -D /backup/base -Ft -z -P -Xs
```

### 3. Restore to a Point in Time

```sh
# Extract base backup to new data directory
mkdir /var/lib/postgresql/16/recovery
tar xzf /backup/base/base.tar.gz -C /var/lib/postgresql/16/recovery
```

```
# postgresql.conf (in recovery data directory)
restore_command = 'cp /archive/wal/%f %p'
recovery_target_time = '2026-03-09 14:30:00'
recovery_target_action = 'promote'
```

```sh
# Create signal file to start recovery
touch /var/lib/postgresql/16/recovery/recovery.signal

# Start PostgreSQL pointing to recovery directory
pg_ctl -D /var/lib/postgresql/16/recovery start
```

### Recovery Target Options

```
recovery_target_time = '2026-03-09 14:30:00'   # specific timestamp
recovery_target_xid = '12345'                   # specific transaction ID
recovery_target_lsn = '0/1A2B3C4D'              # specific WAL position
recovery_target = 'immediate'                    # first consistent point
recovery_target_action = 'promote'               # promote to primary after recovery
recovery_target_action = 'pause'                 # pause for inspection
```

## Backup Validation

```sh
# Verify a custom format dump is readable
pg_restore -l mydb.dump > /dev/null

# Test restore to a throwaway database
createdb mydb_test_restore
pg_restore -d mydb_test_restore mydb.dump
# Run sanity checks
psql -d mydb_test_restore -c "SELECT count(*) FROM orders;"
dropdb mydb_test_restore

# Verify physical backup checksums (PG 13+)
pg_verifybackup /backup/base
```

## RDS Backups

### Automated Backups

- Enabled by default, retention 1–35 days (default 7)
- Daily snapshot + continuous WAL archiving to S3
- Set backup window to low-traffic period
- Backups happen on the standby in Multi-AZ (no performance impact)

### Manual Snapshots

```sh
# Create snapshot via CLI
aws rds create-db-snapshot \
  --db-instance-identifier mydb \
  --db-snapshot-identifier mydb-pre-migration-2026-03-09

# List snapshots
aws rds describe-db-snapshots --db-instance-identifier mydb

# Restore from snapshot (creates a NEW instance)
aws rds restore-db-instance-from-db-snapshot \
  --db-instance-identifier mydb-restored \
  --db-snapshot-identifier mydb-pre-migration-2026-03-09

# Delete old snapshot
aws rds delete-db-snapshot --db-snapshot-identifier mydb-old-snapshot
```

### Point-in-Time Recovery on RDS

```sh
# Restore to a specific time (creates a NEW instance)
aws rds restore-db-instance-to-point-in-time \
  --source-db-instance-identifier mydb \
  --target-db-instance-identifier mydb-pitr \
  --restore-time "2026-03-09T14:30:00Z"

# Restore to latest restorable time
aws rds restore-db-instance-to-point-in-time \
  --source-db-instance-identifier mydb \
  --target-db-instance-identifier mydb-pitr \
  --use-latest-restorable-time
```

### Cross-Region / Cross-Account

```sh
# Copy snapshot to another region
aws rds copy-db-snapshot \
  --source-db-snapshot-identifier arn:aws:rds:us-east-1:111111111111:snapshot:mydb-snap \
  --target-db-snapshot-identifier mydb-snap-copy \
  --region eu-west-1

# Share snapshot with another account
aws rds modify-db-snapshot-attribute \
  --db-snapshot-identifier mydb-snap \
  --attribute-name restore \
  --values-to-add "222222222222"
```

## Backup Strategy Summary

| Method               | Type     | Granularity                   | Use Case                                        |
| -------------------- | -------- | ----------------------------- | ----------------------------------------------- |
| pg_dump              | Logical  | Per database/table            | Dev environments, migrations, selective restore |
| pg_basebackup + PITR | Physical | Any point in time             | Production disaster recovery                    |
| RDS automated        | Physical | Any point in retention window | Managed production workloads                    |
| RDS snapshot         | Physical | Point-in-time snapshot        | Pre-migration safety net, cross-region DR       |
