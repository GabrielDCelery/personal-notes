# Configuration

## Instance Sizing

### Instance Families

| Family              | Use Case                 | Notes                            |
| ------------------- | ------------------------ | -------------------------------- |
| `db.t4g`            | Dev/test, low-traffic    | Burstable, CPU credits, cheapest |
| `db.r6g` / `db.r7g` | Production, memory-heavy | Memory-optimized, Graviton       |
| `db.m6g` / `db.m7g` | Balanced workloads       | General purpose, Graviton        |
| `db.r6i` / `db.m6i` | Intel-specific workloads | x86, slightly more expensive     |
| `db.x2g`            | Very large memory needs  | Up to 1TB RAM                    |

### Choosing Instance Size

- **CPU**: match to concurrent active queries, not connections
- **Memory**: `shared_buffers` + working set should fit in RAM
- **Baseline**: start with `db.r6g.xlarge` (4 vCPU, 32GB) for production, scale from there
- `max_connections` is derived from memory: `LEAST({DBInstanceClassMemory/9531392}, 5000)`

### Key Instance Limits

| Instance         | vCPU | Memory | max_connections (approx) | Network       |
| ---------------- | ---- | ------ | ------------------------ | ------------- |
| `db.t4g.micro`   | 2    | 1GB    | 112                      | Low           |
| `db.t4g.medium`  | 2    | 4GB    | 420                      | Moderate      |
| `db.r6g.large`   | 2    | 16GB   | 1680                     | Up to 10 Gbps |
| `db.r6g.xlarge`  | 4    | 32GB   | 3360                     | Up to 10 Gbps |
| `db.r6g.2xlarge` | 8    | 64GB   | 5000                     | Up to 10 Gbps |
| `db.r6g.4xlarge` | 16   | 128GB  | 5000                     | Up to 10 Gbps |

## Storage Types

| Type     | IOPS                            | Throughput                         | Use Case               |
| -------- | ------------------------------- | ---------------------------------- | ---------------------- |
| gp3      | 3000 baseline, up to 16000      | 125 MBps baseline, up to 1000 MBps | Most workloads         |
| io1      | Up to 64000                     | Up to 1000 MBps                    | High IOPS requirements |
| io2      | Up to 64000, 99.999% durability | Up to 1000 MBps                    | Critical production    |
| magnetic | Low                             | Low                                | Legacy, avoid          |

### gp3 Recommendations

```sh
# Default: 3000 IOPS, 125 MBps (included in storage cost)
# Scale IOPS and throughput independently of storage size

aws rds modify-db-instance \
  --db-instance-identifier mydb \
  --storage-type gp3 \
  --allocated-storage 500 \
  --iops 6000 \
  --storage-throughput 500 \
  --apply-immediately
```

- gp3 is almost always the right choice
- Provision IOPS above baseline only if CloudWatch shows `ReadIOPS`/`WriteIOPS` consistently hitting 3000
- Monitor `ReadLatency`/`WriteLatency` — if > 5ms, consider increasing IOPS

### Storage Autoscaling

```sh
aws rds modify-db-instance \
  --db-instance-identifier mydb \
  --max-allocated-storage 1000 \
  --apply-immediately
```

- Scales up automatically when free space < 10% and storage stays low for 5 min
- Cannot scale down — only up
- Scaling events cause brief I/O suspension

## Parameter Groups

RDS doesn't allow editing `postgresql.conf` directly. Use parameter groups.

### Create and Apply

```sh
# Create custom parameter group
aws rds create-db-parameter-group \
  --db-parameter-group-name mydb-params \
  --db-parameter-group-family postgres16 \
  --description "Custom params for mydb"

# Modify parameters
aws rds modify-db-parameter-group \
  --db-parameter-group-name mydb-params \
  --parameters \
    "ParameterName=shared_buffers,ParameterValue={DBInstanceClassMemory/4},ApplyMethod=pending-reboot" \
    "ParameterName=work_mem,ParameterValue=65536,ApplyMethod=immediate" \
    "ParameterName=random_page_cost,ParameterValue=1.1,ApplyMethod=immediate" \
    "ParameterName=effective_cache_size,ParameterValue={DBInstanceClassMemory*3/4},ApplyMethod=immediate"

# Attach to instance
aws rds modify-db-instance \
  --db-instance-identifier mydb \
  --db-parameter-group-name mydb-params \
  --apply-immediately

# Check pending-reboot parameters
aws rds describe-db-instances \
  --db-instance-identifier mydb \
  --query 'DBInstances[0].DBParameterGroups'
```

### Recommended Parameters

```
# Memory
shared_buffers = {DBInstanceClassMemory/4}
effective_cache_size = {DBInstanceClassMemory*3/4}
work_mem = 65536                        # 64MB, adjust per workload
maintenance_work_mem = 1048576          # 1GB

# Planner
random_page_cost = 1.1                 # SSD storage
effective_io_concurrency = 200
default_statistics_target = 100

# WAL
max_wal_size = 4096                    # 4GB in MB
checkpoint_completion_target = 0.9

# Logging
log_min_duration_statement = 1000      # log queries > 1s
log_connections = 1
log_disconnections = 1
log_lock_waits = 1
log_temp_files = 0
log_checkpoints = 1

# Autovacuum
autovacuum_max_workers = 5
autovacuum_vacuum_cost_delay = 2
autovacuum_vacuum_cost_limit = 1000
rds.force_autovacuum_logging_level = log

# Monitoring
shared_preload_libraries = pg_stat_statements,auto_explain
pg_stat_statements.track = all
auto_explain.log_min_duration = 3000
```

> RDS uses `{DBInstanceClassMemory}` as a variable in parameter values — it resolves to the instance's memory in bytes.

### Dynamic vs Static Parameters

```sh
# Check if parameter needs reboot
aws rds describe-db-parameters \
  --db-parameter-group-name mydb-params \
  --query "Parameters[?ParameterName=='shared_buffers'].[ParameterName,ApplyType]"
```

| Apply Type | Meaning                               |
| ---------- | ------------------------------------- |
| `dynamic`  | Applied immediately or on next reload |
| `static`   | Requires instance reboot              |

## Extensions

```sh
# List available extensions
aws rds describe-db-engine-versions \
  --engine postgres \
  --engine-version 16.4 \
  --query 'DBEngineVersions[0].SupportedFeatureNames'
```

```sql
-- Install extension (must be in shared_preload_libraries for some)
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS pgstattuple;
CREATE EXTENSION IF NOT EXISTS pg_hint_plan;

-- List installed extensions
SELECT * FROM pg_available_extensions WHERE installed_version IS NOT NULL;
```

### Extensions Requiring shared_preload_libraries

Must be added to the parameter group and requires reboot:

- `pg_stat_statements`
- `auto_explain`
- `pg_hint_plan`
- `pg_cron`

## Encryption

### At Rest

- Enabled at creation time, cannot be changed after
- Uses AWS KMS (default or custom key)
- Encrypts storage, backups, snapshots, and replicas

```sh
aws rds create-db-instance \
  --db-instance-identifier mydb \
  --storage-encrypted \
  --kms-key-id arn:aws:kms:us-east-1:111111111111:key/xxxxx \
  ...
```

### In Transit

```
# Force SSL in parameter group
rds.force_ssl = 1
```

```sh
# Connect with SSL
psql "host=mydb.xxxxx.us-east-1.rds.amazonaws.com dbname=mydb sslmode=verify-full sslrootcert=us-east-1-bundle.pem"
```
