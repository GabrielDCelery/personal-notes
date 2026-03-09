# Partitioning

## When to Partition

- Tables with hundreds of millions of rows or hundreds of GB
- Queries consistently filter on one column (date, tenant_id, region)
- Need to efficiently drop old data (detach partition vs DELETE)
- Maintenance operations (VACUUM, REINDEX) are too slow on the full table

Don't partition tables under ~10M rows — the overhead isn't worth it.

## Partition Types

### Range Partitioning

Best for time-series data.

```sql
CREATE TABLE events (
    id bigserial,
    created_at timestamptz NOT NULL,
    event_type text,
    payload jsonb
) PARTITION BY RANGE (created_at);

-- Create partitions
CREATE TABLE events_2026_01 PARTITION OF events
    FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');

CREATE TABLE events_2026_02 PARTITION OF events
    FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');

CREATE TABLE events_2026_03 PARTITION OF events
    FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');

-- Default partition (catches rows that don't match any partition)
CREATE TABLE events_default PARTITION OF events DEFAULT;
```

### List Partitioning

Best for categorical data.

```sql
CREATE TABLE orders (
    id bigserial,
    region text NOT NULL,
    customer_id bigint,
    total numeric
) PARTITION BY LIST (region);

CREATE TABLE orders_us PARTITION OF orders FOR VALUES IN ('us-east', 'us-west');
CREATE TABLE orders_eu PARTITION OF orders FOR VALUES IN ('eu-west', 'eu-central');
CREATE TABLE orders_ap PARTITION OF orders FOR VALUES IN ('ap-southeast', 'ap-northeast');
CREATE TABLE orders_default PARTITION OF orders DEFAULT;
```

### Hash Partitioning

Distributes rows evenly. Useful for tenant-based sharding.

```sql
CREATE TABLE sessions (
    id bigserial,
    user_id bigint NOT NULL,
    data jsonb
) PARTITION BY HASH (user_id);

CREATE TABLE sessions_0 PARTITION OF sessions FOR VALUES WITH (MODULUS 4, REMAINDER 0);
CREATE TABLE sessions_1 PARTITION OF sessions FOR VALUES WITH (MODULUS 4, REMAINDER 1);
CREATE TABLE sessions_2 PARTITION OF sessions FOR VALUES WITH (MODULUS 4, REMAINDER 2);
CREATE TABLE sessions_3 PARTITION OF sessions FOR VALUES WITH (MODULUS 4, REMAINDER 3);
```

### Multi-Level Partitioning

```sql
CREATE TABLE events (
    id bigserial,
    created_at timestamptz NOT NULL,
    region text NOT NULL,
    payload jsonb
) PARTITION BY RANGE (created_at);

CREATE TABLE events_2026_01 PARTITION OF events
    FOR VALUES FROM ('2026-01-01') TO ('2026-02-01')
    PARTITION BY LIST (region);

CREATE TABLE events_2026_01_us PARTITION OF events_2026_01
    FOR VALUES IN ('us-east', 'us-west');

CREATE TABLE events_2026_01_eu PARTITION OF events_2026_01
    FOR VALUES IN ('eu-west', 'eu-central');
```

## Partition Pruning

PostgreSQL skips partitions that can't contain matching rows. Must be enabled (default is on).

```sql
-- Verify pruning is enabled
SHOW enable_partition_pruning;

-- Check pruning in action
EXPLAIN SELECT * FROM events WHERE created_at = '2026-02-15';
-- Should show only events_2026_02 being scanned
```

For pruning to work, queries must include the partition key in the WHERE clause.

## Indexes on Partitioned Tables

```sql
-- Index on the parent creates matching indexes on all partitions
CREATE INDEX idx_events_type ON events (event_type);

-- Unique indexes MUST include the partition key
CREATE UNIQUE INDEX idx_events_id ON events (id, created_at);
-- Cannot create unique on (id) alone across partitions

-- Primary key must include partition key
ALTER TABLE events ADD PRIMARY KEY (id, created_at);
```

## Managing Partitions

### Add New Partition

```sql
CREATE TABLE events_2026_04 PARTITION OF events
    FOR VALUES FROM ('2026-04-01') TO ('2026-05-01');
```

### Detach Partition (Remove Old Data)

```sql
-- Detach without blocking queries (PG 14+)
ALTER TABLE events DETACH PARTITION events_2025_01 CONCURRENTLY;

-- Then drop or archive at leisure
DROP TABLE events_2025_01;

-- Or keep it around as a standalone table
-- SELECT * FROM events_2025_01 still works
```

### Attach Existing Table as Partition

```sql
-- Table must have matching schema and a constraint matching the partition bounds
ALTER TABLE events_archive
    ADD CONSTRAINT check_dates CHECK (created_at >= '2025-06-01' AND created_at < '2025-07-01');

ALTER TABLE events ATTACH PARTITION events_archive
    FOR VALUES FROM ('2025-06-01') TO ('2025-07-01');
```

> Adding the CHECK constraint before attaching avoids a full table scan during ATTACH.

### Move Data from Default Partition

When creating a new partition that overlaps with data in the default partition:

```sql
-- Detach default
ALTER TABLE events DETACH PARTITION events_default;

-- Create the new partition
CREATE TABLE events_2026_04 PARTITION OF events
    FOR VALUES FROM ('2026-04-01') TO ('2026-05-01');

-- Move matching rows from detached default to new partition
INSERT INTO events SELECT * FROM events_default
    WHERE created_at >= '2026-04-01' AND created_at < '2026-05-01';

DELETE FROM events_default
    WHERE created_at >= '2026-04-01' AND created_at < '2026-05-01';

-- Re-attach default
ALTER TABLE events ATTACH PARTITION events_default DEFAULT;
```

## Automating Partition Creation

### pg_partman

Extension for automatic partition management.

```sql
CREATE EXTENSION pg_partman;

-- Configure automatic monthly partitions
SELECT partman.create_parent(
    p_parent_table := 'public.events',
    p_control := 'created_at',
    p_type := 'native',
    p_interval := '1 month',
    p_premake := 3              -- create 3 future partitions
);

-- Run maintenance (creates new partitions, drops old ones)
-- Schedule this via pg_cron or external cron
SELECT partman.run_maintenance();
```

```
# Retention — automatically drop partitions older than 12 months
UPDATE partman.part_config
SET retention = '12 months',
    retention_keep_table = false
WHERE parent_table = 'public.events';
```

### pg_cron for Scheduling

```sql
CREATE EXTENSION pg_cron;

-- Run partition maintenance daily at 3 AM
SELECT cron.schedule('partition-maintenance', '0 3 * * *',
    $$SELECT partman.run_maintenance()$$);
```

## Querying Across Partitions

```sql
-- Queries on the parent table automatically include all partitions
SELECT count(*) FROM events WHERE created_at >= '2026-01-01';

-- Aggregate across partitions
SELECT date_trunc('month', created_at) AS month, count(*)
FROM events
GROUP BY 1
ORDER BY 1;

-- Direct query on a specific partition (bypasses pruning overhead)
SELECT * FROM events_2026_03 WHERE event_type = 'click';
```

## Limitations

- Foreign keys referencing partitioned tables require PG 12+
- Foreign keys FROM partitioned tables work since PG 11
- Unique constraints must include partition key
- Cannot create exclusion constraints across partitions
- INSERT triggers on the parent don't fire on child tables (use per-partition triggers)
- `pg_dump` dumps each partition separately — large backups
- Row movement between partitions (UPDATE that changes partition key) must be enabled:

```sql
-- Allow rows to move between partitions on UPDATE
ALTER TABLE events ENABLE ROW MOVEMENT;
```
