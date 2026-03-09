# Connections

## Check Active Connections

```sql
-- Current connections by state
SELECT state, count(*)
FROM pg_stat_activity
GROUP BY state;

-- Detailed view of active connections
SELECT pid, usename, application_name, client_addr, state, query_start, query
FROM pg_stat_activity
WHERE state != 'idle'
ORDER BY query_start;

-- Connections per database
SELECT datname, count(*)
FROM pg_stat_activity
GROUP BY datname;
```

## Connection Limits

```sql
-- Check max connections
SHOW max_connections;

-- Check remaining connection slots
SELECT max_conn, used, max_conn - used AS available
FROM (SELECT setting::int AS max_conn FROM pg_settings WHERE name = 'max_connections') max,
     (SELECT count(*) AS used FROM pg_stat_activity) current;

-- Check reserved superuser connections
SHOW superuser_reserved_connections;
```

## Terminate Connections

```sql
-- Cancel a running query (graceful)
SELECT pg_cancel_backend(pid);

-- Terminate a connection (forceful)
SELECT pg_terminate_backend(pid);

-- Terminate all connections to a specific database
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE datname = 'mydb' AND pid != pg_backend_pid();
```

## Idle Connection Management

```sql
-- Find long-running idle connections
SELECT pid, usename, state, state_change, now() - state_change AS idle_duration
FROM pg_stat_activity
WHERE state = 'idle'
ORDER BY idle_duration DESC;

-- Kill idle connections older than 10 minutes
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE state = 'idle'
  AND state_change < now() - interval '10 minutes';
```

### Idle Timeout Configuration

```
# postgresql.conf
idle_in_transaction_session_timeout = '5min'   # kill idle-in-transaction after 5 min
idle_session_timeout = '30min'                  # kill idle sessions after 30 min (PG 14+)
```

## PgBouncer

PgBouncer sits between the application and PostgreSQL, pooling connections to avoid exhausting `max_connections`.

### Pool Modes

| Mode          | Description                                | Use Case                             |
| ------------- | ------------------------------------------ | ------------------------------------ |
| `session`     | Connection held for entire client session  | Full feature compatibility           |
| `transaction` | Connection returned after each transaction | Most common, good balance            |
| `statement`   | Connection returned after each statement   | Simple queries only, no transactions |

### Minimal Configuration

```ini
# pgbouncer.ini
[databases]
mydb = host=127.0.0.1 port=5432 dbname=mydb

[pgbouncer]
listen_addr = 0.0.0.0
listen_port = 6432
auth_type = md5
auth_file = /etc/pgbouncer/userlist.txt
pool_mode = transaction
max_client_conn = 1000
default_pool_size = 25
min_pool_size = 5
reserve_pool_size = 5
reserve_pool_timeout = 3
```

### Key Sizing Rules

- `default_pool_size` — connections per user/database pair, start with `(core_count * 2) + effective_spindle_count` on the PG server
- `max_client_conn` — max app-side connections PgBouncer accepts, can be much higher than PG `max_connections`
- Total PG connections used = `number_of_pools * default_pool_size + reserve_pool_size`

### Admin Commands

```sql
-- Connect to PgBouncer admin console (port 6432, database pgbouncer)
-- Show pool status
SHOW POOLS;

-- Show active client/server connections
SHOW CLIENTS;
SHOW SERVERS;

-- Show stats
SHOW STATS;

-- Reload config without restart
RELOAD;
```

## RDS Connection Considerations

- `max_connections` is derived from instance memory: `LEAST({DBInstanceClassMemory/9531392}, 5000)`
- Use RDS Proxy for managed connection pooling (alternative to self-hosted PgBouncer)
- RDS Proxy supports IAM auth and auto-failover to standby
- Monitor with CloudWatch: `DatabaseConnections`, `DBLoadCPU`, `DBLoadNonCPU`

### RDS Proxy Basics

```
# Application connects to proxy endpoint instead of RDS endpoint
# Proxy handles pooling, failover, and IAM auth transparently

Endpoint: myproxy.proxy-xxxxx.region.rds.amazonaws.com
Port: 5432
```

## Connection String Tuning

```
# Application-side pool settings (example: pgx in Go, or any connection pool)
# These are general principles, not PgBouncer settings

pool_max_conns = 25          # match PgBouncer default_pool_size or PG capacity
pool_min_conns = 5           # keep warm connections
pool_max_conn_lifetime = 1h  # recycle connections to handle DNS changes / failover
pool_max_conn_idle_time = 30m
pool_health_check_period = 1m
```
