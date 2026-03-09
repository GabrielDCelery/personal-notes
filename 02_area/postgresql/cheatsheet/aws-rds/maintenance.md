# Maintenance

## Maintenance Windows

### Configure

```sh
# Set preferred maintenance window (UTC)
aws rds modify-db-instance \
  --db-instance-identifier mydb \
  --preferred-maintenance-window "sun:03:00-sun:04:00" \
  --apply-immediately
```

- 30-minute minimum window
- Schedule during lowest-traffic period
- Multi-AZ: patching happens on standby first, then failover, then old primary — brief downtime (~60s)
- Single-AZ: instance is unavailable during patching

### Pending Maintenance

```sh
# Check pending maintenance actions
aws rds describe-pending-maintenance-actions \
  --resource-identifier arn:aws:rds:us-east-1:111111111111:db:mydb

# Apply immediately instead of waiting for window
aws rds apply-pending-maintenance-action \
  --resource-identifier arn:aws:rds:us-east-1:111111111111:db:mydb \
  --apply-action system-update \
  --opt-in-type immediate
```

### Maintenance Types

| Type                   | Auto-Applied                  | Downtime                    |
| ---------------------- | ----------------------------- | --------------------------- |
| OS patches             | Yes, during window            | Brief (Multi-AZ: failover)  |
| Minor version upgrades | If auto-minor-upgrade enabled | Brief reboot                |
| Major version upgrades | Never automatic               | Extended (minutes to hours) |
| Hardware maintenance   | Yes, during window            | Brief                       |

## Minor Version Upgrades

```sh
# Enable auto minor version upgrade
aws rds modify-db-instance \
  --db-instance-identifier mydb \
  --auto-minor-version-upgrade \
  --apply-immediately

# Manual minor upgrade
aws rds modify-db-instance \
  --db-instance-identifier mydb \
  --engine-version 16.6 \
  --apply-immediately
```

- Auto minor upgrades happen during the maintenance window
- Minor upgrades are generally safe (bug fixes, security patches)
- Replicas are upgraded automatically after the primary

## Major Version Upgrades

### Pre-Upgrade Checklist

1. **Check compatibility**

```sh
# Identify unsupported features or extensions
aws rds describe-db-engine-versions \
  --engine postgres \
  --engine-version 16.4 \
  --query 'DBEngineVersions[0].ValidUpgradeTarget[*].EngineVersion'
```

2. **Test on a snapshot restore**

```sh
# Create snapshot
aws rds create-db-snapshot \
  --db-instance-identifier mydb \
  --db-snapshot-identifier mydb-pre-upgrade

# Restore to test instance
aws rds restore-db-instance-from-db-snapshot \
  --db-instance-identifier mydb-upgrade-test \
  --db-snapshot-identifier mydb-pre-upgrade

# Upgrade test instance
aws rds modify-db-instance \
  --db-instance-identifier mydb-upgrade-test \
  --engine-version 16.4 \
  --allow-major-version-upgrade \
  --apply-immediately
```

3. **Check pg_upgrade logs** — available in RDS logs after upgrade attempt

4. **Update parameter group** — create a new one for the target version family

### In-Place Upgrade

```sh
aws rds modify-db-instance \
  --db-instance-identifier mydb \
  --engine-version 16.4 \
  --db-parameter-group-name mydb-params-pg16 \
  --allow-major-version-upgrade \
  --apply-immediately
```

- Downtime: typically 10-30 minutes depending on database size
- Runs `pg_upgrade` internally
- Extensions are upgraded automatically where possible
- `ANALYZE` runs automatically after upgrade to rebuild statistics

### Blue-Green Upgrade (Preferred)

Minimizes downtime for major upgrades.

```sh
# Create blue-green deployment with target version
aws rds create-blue-green-deployment \
  --blue-green-deployment-name mydb-pg16-upgrade \
  --source arn:aws:rds:us-east-1:111111111111:db:mydb \
  --target-engine-version 16.4 \
  --target-db-parameter-group-name mydb-params-pg16

# Wait for green environment to be ready and in sync
aws rds describe-blue-green-deployments \
  --blue-green-deployment-identifier mydb-pg16-upgrade

# Validate green environment
# - run test queries
# - check application compatibility
# - verify extensions

# Switchover (< 1 min downtime)
aws rds switchover-blue-green-deployment \
  --blue-green-deployment-identifier bgd-xxxxx \
  --switchover-timeout 300

# Cleanup after validation
aws rds delete-blue-green-deployment \
  --blue-green-deployment-identifier bgd-xxxxx \
  --delete-target
```

Advantages over in-place:

- Test the upgraded database before switching traffic
- Rollback = don't switch over (no data risk)
- < 1 minute downtime vs 10-30 minutes

## Post-Upgrade Tasks

```sql
-- Rebuild optimizer statistics (usually runs automatically, but verify)
ANALYZE;

-- Check for invalid indexes (can happen after major upgrades)
SELECT indexrelid::regclass, indisvalid
FROM pg_index
WHERE NOT indisvalid;

-- Reindex invalid indexes
REINDEX INDEX CONCURRENTLY <index_name>;

-- Verify extensions are working
SELECT extname, extversion FROM pg_extension;
```

## Instance Class Changes

```sh
# Change instance class (causes reboot)
aws rds modify-db-instance \
  --db-instance-identifier mydb \
  --db-instance-class db.r7g.2xlarge \
  --apply-immediately
```

- Multi-AZ: changes standby first, failover, then old primary — brief downtime
- Single-AZ: full reboot, longer downtime
- Memory-derived parameters (shared_buffers, max_connections) adjust automatically if using formulas like `{DBInstanceClassMemory/4}`

## Storage Changes

```sh
# Increase storage (no downtime, but I/O may be affected)
aws rds modify-db-instance \
  --db-instance-identifier mydb \
  --allocated-storage 1000 \
  --apply-immediately

# Change storage type
aws rds modify-db-instance \
  --db-instance-identifier mydb \
  --storage-type gp3 \
  --iops 6000 \
  --storage-throughput 500 \
  --apply-immediately
```

- Storage can only be increased, never decreased
- 6-hour cooldown between storage modifications
- Storage optimization phase can last hours after modification

## Reboot

```sh
# Regular reboot (applies pending parameter changes)
aws rds reboot-db-instance \
  --db-instance-identifier mydb

# Reboot with failover (Multi-AZ only, tests failover)
aws rds reboot-db-instance \
  --db-instance-identifier mydb \
  --force-failover
```

## Monitoring Maintenance Impact

```sh
# Subscribe to maintenance events
aws rds create-event-subscription \
  --subscription-name mydb-maintenance \
  --sns-topic-arn arn:aws:sns:us-east-1:111111111111:rds-alerts \
  --source-type db-instance \
  --event-categories "maintenance" "notification" "recovery"

# View recent events
aws rds describe-events \
  --source-identifier mydb \
  --source-type db-instance \
  --duration 1440
```
