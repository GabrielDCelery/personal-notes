# High Availability

## Multi-AZ Deployments

### How It Works

- RDS maintains a synchronous standby replica in a different Availability Zone
- Replication is at the storage level (not streaming replication)
- Standby is not accessible for reads — it's purely for failover
- Automatic failover on: instance failure, AZ outage, storage failure, OS patching

### Failover Behavior

- DNS endpoint flips to standby (typically 60–120 seconds)
- Application must reconnect — connections are dropped during failover
- No data loss (synchronous replication)
- Failover events appear in RDS Events and CloudWatch

### Enable Multi-AZ

```sh
# On creation
aws rds create-db-instance \
  --db-instance-identifier mydb \
  --multi-az \
  --engine postgres \
  --db-instance-class db.r6g.xlarge \
  --allocated-storage 100

# Convert existing single-AZ to multi-AZ (causes brief I/O suspension)
aws rds modify-db-instance \
  --db-instance-identifier mydb \
  --multi-az \
  --apply-immediately
```

### Multi-AZ DB Cluster (PG 13.4+)

Different from standard Multi-AZ — uses two readable standby instances.

- One writer + two reader instances across 3 AZs
- Readers are accessible for read traffic (reader endpoint)
- Transaction log-based replication (not storage-level)
- Faster failover (~35 seconds)
- `mydb-cluster.cluster-xxxxx.region.rds.amazonaws.com` — writer endpoint
- `mydb-cluster.cluster-ro-xxxxx.region.rds.amazonaws.com` — reader endpoint

## Read Replicas

### Purpose

- Offload read traffic from the primary
- Cross-region disaster recovery
- Can be promoted to standalone instance

### Create and Manage

```sh
# Create read replica
aws rds create-db-instance-read-replica \
  --db-instance-identifier mydb-replica \
  --source-db-instance-identifier mydb \
  --db-instance-class db.r6g.xlarge

# Cross-region read replica
aws rds create-db-instance-read-replica \
  --db-instance-identifier mydb-replica-eu \
  --source-db-instance-identifier arn:aws:rds:us-east-1:111111111111:db:mydb \
  --region eu-west-1 \
  --db-instance-class db.r6g.xlarge

# Promote replica to standalone (irreversible)
aws rds promote-read-replica --db-instance-identifier mydb-replica

# Check replica lag
aws cloudwatch get-metric-statistics \
  --namespace AWS/RDS \
  --metric-name ReplicaLag \
  --dimensions Name=DBInstanceIdentifier,Value=mydb-replica \
  --start-time 2026-03-09T00:00:00Z \
  --end-time 2026-03-09T23:59:59Z \
  --period 300 \
  --statistics Average
```

### Replica Lag

- Uses asynchronous PostgreSQL streaming replication
- Typical lag: < 1 second under normal load
- Monitor: CloudWatch `ReplicaLag` metric (seconds)
- Causes of high lag: write-heavy primary, undersized replica, long-running queries on replica, network issues

### Read Replica Limits

- Up to 15 read replicas per primary (5 cross-region)
- Replicas can have their own read replicas (chaining, adds lag)
- Each replica is an independent instance (own parameter group, instance class)
- Replicas can have Multi-AZ enabled for their own HA

## Failover Strategies

### Application-Level Best Practices

```
# Connection string should use the RDS endpoint (DNS-based)
# Never hardcode IP addresses — they change on failover

# Connection pool settings for fast failover recovery
pool_max_conn_lifetime = 5m    # recycle connections frequently
pool_health_check_period = 30s # detect dead connections quickly
connect_timeout = 5            # fail fast on connection attempt

# Retry logic: retry failed queries that are safe to retry (idempotent reads)
# Don't blindly retry writes — they may have succeeded before the connection dropped
```

### DNS TTL

- RDS endpoint DNS TTL is 5 seconds
- Ensure your application / connection pool respects DNS TTL
- Java: set `networkaddress.cache.ttl=5` in JVM
- Go: uses OS resolver by default (respects TTL)
- Some connection pools cache resolved IPs — verify yours doesn't

### Testing Failover

```sh
# Force a failover (Multi-AZ only)
aws rds reboot-db-instance \
  --db-instance-identifier mydb \
  --force-failover

# Monitor failover events
aws rds describe-events \
  --source-identifier mydb \
  --source-type db-instance \
  --duration 60
```

## Blue-Green Deployments

Create a staged copy of the production environment for major changes (engine upgrades, schema changes). RDS handles replication between blue (current) and green (new).

```sh
# Create blue-green deployment
aws rds create-blue-green-deployment \
  --blue-green-deployment-name mydb-upgrade \
  --source arn:aws:rds:us-east-1:111111111111:db:mydb \
  --target-engine-version 16.4

# Monitor replication lag between blue and green
aws rds describe-blue-green-deployments \
  --blue-green-deployment-identifier mydb-upgrade

# Switchover (promotes green, demotes blue — typically < 1 min downtime)
aws rds switchover-blue-green-deployment \
  --blue-green-deployment-identifier bgd-xxxxx

# Delete blue-green deployment after validation
aws rds delete-blue-green-deployment \
  --blue-green-deployment-identifier bgd-xxxxx \
  --delete-target
```

### Blue-Green Use Cases

- Major version upgrades (e.g., PG 15 to PG 16)
- Parameter group changes that need restart
- Schema migrations you want to validate before cutover
- Instance class changes

## Monitoring HA

### Key CloudWatch Metrics

| Metric                   | Alert Threshold  | Notes                                 |
| ------------------------ | ---------------- | ------------------------------------- |
| `ReplicaLag`             | > 30 seconds     | Replica falling behind                |
| `FreeStorageSpace`       | < 20%            | Storage affecting replication         |
| `DatabaseConnections`    | > 80% of max     | Connection exhaustion during failover |
| `ReadIOPS` / `WriteIOPS` | Sustained spikes | I/O pressure causing lag              |

### RDS Event Subscriptions

```sh
# Subscribe to failover and recovery events
aws rds create-event-subscription \
  --subscription-name mydb-ha-alerts \
  --sns-topic-arn arn:aws:sns:us-east-1:111111111111:rds-alerts \
  --source-type db-instance \
  --event-categories "failover" "recovery" "maintenance"
```
