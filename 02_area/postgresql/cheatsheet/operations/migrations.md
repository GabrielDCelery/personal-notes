# Migrations

## Zero-Downtime Principles

1. Never hold long locks on tables that serve traffic
2. Make changes backwards-compatible (old code must work with new schema)
3. Deploy in multiple phases: schema change first, then code, then cleanup
4. Always test migrations against production-sized data

## Safe vs Unsafe Operations

### Safe (No/Minimal Locking)

| Operation                           | Notes                                               |
| ----------------------------------- | --------------------------------------------------- |
| `CREATE INDEX CONCURRENTLY`         | No write lock, but slower                           |
| `ADD COLUMN` (nullable, no default) | Instant metadata change                             |
| `ADD COLUMN ... DEFAULT x` (PG 11+) | Instant — default stored in catalog, not rewritten  |
| `DROP COLUMN`                       | Instant — marks column as dropped, doesn't rewrite  |
| `CREATE TABLE`                      | No lock on existing tables                          |
| `ADD CONSTRAINT ... NOT VALID`      | Doesn't scan existing rows                          |
| `VALIDATE CONSTRAINT`               | ShareUpdateExclusive lock — allows reads and writes |

### Unsafe (Table Rewrite or Heavy Lock)

| Operation                                | Risk                                    | Safe Alternative                                                     |
| ---------------------------------------- | --------------------------------------- | -------------------------------------------------------------------- |
| `ADD COLUMN ... DEFAULT x` (PG < 11)     | Rewrites entire table                   | Add column, set default, backfill in batches                         |
| `ALTER COLUMN TYPE`                      | Rewrites table                          | Add new column, backfill, swap                                       |
| `ADD CONSTRAINT ... (without NOT VALID)` | Scans entire table with AccessExclusive | Use NOT VALID + VALIDATE                                             |
| `SET NOT NULL`                           | Full table scan with AccessExclusive    | Add CHECK constraint NOT VALID, validate, then SET NOT NULL (PG 12+) |
| `VACUUM FULL`                            | AccessExclusive for entire duration     | Use pg_repack                                                        |
| `CREATE INDEX` (without CONCURRENTLY)    | Blocks writes                           | Always use CONCURRENTLY                                              |

## Common Patterns

### Add a NOT NULL Column

```sql
-- Phase 1: Add nullable column with default (PG 11+, instant)
ALTER TABLE orders ADD COLUMN priority int DEFAULT 0;

-- Phase 2: Add NOT NULL constraint safely
ALTER TABLE orders ADD CONSTRAINT orders_priority_not_null
  CHECK (priority IS NOT NULL) NOT VALID;
ALTER TABLE orders VALIDATE CONSTRAINT orders_priority_not_null;

-- Phase 3 (PG 12+): SET NOT NULL uses existing valid check constraint, no scan
ALTER TABLE orders ALTER COLUMN priority SET NOT NULL;
ALTER TABLE orders DROP CONSTRAINT orders_priority_not_null;
```

### Change Column Type

```sql
-- Never do: ALTER TABLE orders ALTER COLUMN amount TYPE numeric(12,2);
-- This rewrites the entire table under AccessExclusive lock.

-- Safe approach: add, backfill, swap
-- Phase 1: Add new column
ALTER TABLE orders ADD COLUMN amount_new numeric(12,2);

-- Phase 2: Backfill in batches
UPDATE orders SET amount_new = amount WHERE id BETWEEN 1 AND 100000;
UPDATE orders SET amount_new = amount WHERE id BETWEEN 100001 AND 200000;
-- ... continue in batches

-- Phase 3: Add trigger to keep in sync during backfill
CREATE OR REPLACE FUNCTION sync_amount() RETURNS trigger AS $$
BEGIN
  NEW.amount_new := NEW.amount;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER sync_amount_trigger
  BEFORE INSERT OR UPDATE ON orders
  FOR EACH ROW EXECUTE FUNCTION sync_amount();

-- Phase 4: Swap columns (brief lock)
BEGIN;
ALTER TABLE orders RENAME COLUMN amount TO amount_old;
ALTER TABLE orders RENAME COLUMN amount_new TO amount;
COMMIT;

-- Phase 5: Deploy code pointing to new column, then cleanup
DROP TRIGGER sync_amount_trigger ON orders;
DROP FUNCTION sync_amount;
ALTER TABLE orders DROP COLUMN amount_old;
```

### Add a Foreign Key

```sql
-- Bad: scans entire table under lock
ALTER TABLE orders ADD CONSTRAINT fk_customer
  FOREIGN KEY (customer_id) REFERENCES customers(id);

-- Good: two-step
ALTER TABLE orders ADD CONSTRAINT fk_customer
  FOREIGN KEY (customer_id) REFERENCES customers(id) NOT VALID;

ALTER TABLE orders VALIDATE CONSTRAINT fk_customer;
```

### Add a Unique Constraint

```sql
-- Bad: creates index non-concurrently under the hood
ALTER TABLE users ADD CONSTRAINT uq_email UNIQUE (email);

-- Good: create index first, then attach
CREATE UNIQUE INDEX CONCURRENTLY idx_users_email ON users (email);
ALTER TABLE users ADD CONSTRAINT uq_email UNIQUE USING INDEX idx_users_email;
```

### Rename a Column

```sql
-- ALTER TABLE orders RENAME COLUMN status TO order_status;
-- This is instant but breaks old code. Use a phased approach:

-- Phase 1: Add new column, keep both in sync via trigger
-- Phase 2: Deploy code using new column name
-- Phase 3: Drop old column and trigger
-- Or just update all code at once if downtime is acceptable
```

### Drop a Column

```sql
-- Instant in PostgreSQL — doesn't rewrite the table
-- But make sure no code references it first
ALTER TABLE orders DROP COLUMN IF EXISTS legacy_field;
```

### Create an Enum Type

```sql
-- Creating is safe
CREATE TYPE order_status AS ENUM ('pending', 'active', 'completed');

-- Adding a value is safe (PG 11+, no lock)
ALTER TYPE order_status ADD VALUE 'cancelled';

-- Removing/renaming values is not supported — create new type and migrate
```

## Backfilling Data

### Batch Updates

```sql
-- Never do a single UPDATE on millions of rows — it creates a huge transaction,
-- holds locks, and generates massive WAL

-- Batch by primary key
DO $$
DECLARE
  batch_size int := 10000;
  max_id bigint;
  current_id bigint := 0;
BEGIN
  SELECT max(id) INTO max_id FROM orders;
  WHILE current_id < max_id LOOP
    UPDATE orders
    SET new_column = compute_value(old_column)
    WHERE id > current_id AND id <= current_id + batch_size
      AND new_column IS NULL;
    current_id := current_id + batch_size;
    COMMIT;
    PERFORM pg_sleep(0.1);  -- brief pause to reduce load
  END LOOP;
END $$;
```

## Lock Timeout Safety

Always set a lock timeout when running migrations against live databases:

```sql
-- Fail fast if lock can't be acquired in 5 seconds
SET lock_timeout = '5s';
SET statement_timeout = '30s';

-- Then retry the migration if it times out rather than blocking traffic
```

## Migration Tools

| Tool           | Language | Notes                                   |
| -------------- | -------- | --------------------------------------- |
| golang-migrate | Go       | SQL-based, simple, widely used          |
| goose          | Go       | SQL or Go-based migrations              |
| Atlas          | Go       | Declarative + versioned, schema diffing |
| Flyway         | Java     | Enterprise-grade, SQL-based             |
| Alembic        | Python   | SQLAlchemy integration                  |
| sqitch         | Perl     | Change-based, dependency-aware          |
| Tern           | Go       | Simple, PostgreSQL-specific             |

## Rollback Strategies

### Forward-Only Migrations

Preferred approach — every migration is additive:

1. Add new column/table
2. Deploy code that writes to both old and new
3. Backfill new from old
4. Deploy code that reads from new
5. Remove old column/table

Rolling back = deploying a new forward migration that reverses the change.

### Reversible Migrations

If your tool supports down migrations:

```sql
-- up.sql
ALTER TABLE orders ADD COLUMN priority int DEFAULT 0;

-- down.sql
ALTER TABLE orders DROP COLUMN priority;
```

> Test down migrations. Untested rollbacks fail when you need them most.

### Pre-Migration Snapshots

```sh
# RDS: take a snapshot before risky migrations
aws rds create-db-snapshot \
  --db-instance-identifier mydb \
  --db-snapshot-identifier mydb-pre-migration-$(date +%Y%m%d)

# Self-managed: pg_dump the affected tables
pg_dump -Fc -d mydb -t orders -f orders_pre_migration.dump
```

## RDS Migration Considerations

- Use `lock_timeout` — RDS has no way to kill stuck DDL from outside
- Parameter group changes that require restart: plan for maintenance window
- Major version upgrades: use blue-green deployments to minimize downtime
- Large backfills: increase instance class temporarily, or use a read replica + promote approach
- Monitor `ReplicaLag` during migrations — heavy writes can cause replicas to fall behind
