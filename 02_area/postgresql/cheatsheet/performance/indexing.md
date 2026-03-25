# Indexing

## Index Types

| Type    | Use Case                                                  | Example                       |
| ------- | --------------------------------------------------------- | ----------------------------- |
| B-tree  | Default. Equality, range, sorting, LIKE 'prefix%'         | Most columns                  |
| Hash    | Equality only (rarely better than B-tree)                 | Exact lookups on large values |
| GIN     | Full-text search, JSONB, arrays, tsvector                 | `WHERE tags @> '{urgent}'`    |
| GiST    | Geometric, range types, full-text (less precise than GIN) | PostGIS, range overlaps       |
| BRIN    | Very large tables with naturally ordered data             | Time-series `created_at`      |
| SP-GiST | Non-balanced tree structures (quad-trees, radix trees)    | IP addresses, phone numbers   |

## Creating Indexes

```sql
-- Basic B-tree
CREATE INDEX idx_orders_customer_id ON orders (customer_id);

-- Unique index
CREATE UNIQUE INDEX idx_users_email ON users (email);

-- Concurrent (no table lock, safe for production)
CREATE INDEX CONCURRENTLY idx_orders_status ON orders (status);

-- Multi-column (order matters — leftmost columns are used first)
CREATE INDEX idx_orders_customer_status ON orders (customer_id, status);

-- GIN on JSONB
CREATE INDEX idx_events_payload ON events USING GIN (payload);

-- GIN on JSONB with path operators only (smaller index)
CREATE INDEX idx_events_payload_path ON events USING GIN (payload jsonb_path_ops);

-- GIN on array column
CREATE INDEX idx_posts_tags ON posts USING GIN (tags);

-- GiST on range type
CREATE INDEX idx_reservations_period ON reservations USING GiST (period);

-- BRIN on time-series data (very small index, works when data is physically ordered)
CREATE INDEX idx_logs_created_at ON logs USING BRIN (created_at);

-- Full-text search with GIN
CREATE INDEX idx_articles_search ON articles USING GIN (to_tsvector('english', title || ' ' || body));
```

> **Always use CONCURRENTLY in production.** Regular `CREATE INDEX` locks the table for writes.

## Partial Indexes

Index only the rows that matter. Smaller, faster, cheaper.

```sql
-- Only index active orders
CREATE INDEX idx_orders_active ON orders (customer_id)
WHERE status = 'active';

-- Only index non-null values
CREATE INDEX idx_users_phone ON users (phone)
WHERE phone IS NOT NULL;

-- Only index recent data
CREATE INDEX idx_orders_recent ON orders (created_at)
WHERE created_at > '2025-01-01';
```

The query must include the WHERE clause (or a subset of it) for the partial index to be used.

## Covering Indexes (INCLUDE)

Add non-key columns to enable Index Only Scans without adding them to the index key.

```sql
-- Index on customer_id, but also store status and total so the query
-- doesn't need to visit the table at all
CREATE INDEX idx_orders_customer_covering ON orders (customer_id)
INCLUDE (status, total);

-- This query can now be an Index Only Scan:
SELECT status, total FROM orders WHERE customer_id = 42;
```

## Expression Indexes

Index on a computed value when queries filter by that expression.

```sql
-- Index on lowercased email
CREATE INDEX idx_users_email_lower ON users (lower(email));
-- Query must match: WHERE lower(email) = 'user@example.com'

-- Index on date extracted from timestamp
CREATE INDEX idx_orders_date ON orders (date(created_at));
-- Query must match: WHERE date(created_at) = '2026-03-09'

-- Index on JSONB field
CREATE INDEX idx_events_type ON events ((payload->>'type'));
-- Query must match: WHERE payload->>'type' = 'click'
```

## Multi-Column Index Strategy

Multi-column B-tree indexes support queries using leftmost prefix columns.

```sql
CREATE INDEX idx_orders_multi ON orders (customer_id, status, created_at);
```

| Query Filter                                                        | Uses Index?                     |
| ------------------------------------------------------------------- | ------------------------------- |
| `WHERE customer_id = 42`                                            | Yes                             |
| `WHERE customer_id = 42 AND status = 'active'`                      | Yes                             |
| `WHERE customer_id = 42 AND status = 'active' AND created_at > ...` | Yes                             |
| `WHERE status = 'active'`                                           | No (missing leftmost column)    |
| `WHERE customer_id = 42 AND created_at > ...`                       | Partial (uses customer_id only) |

**Column order rule of thumb**: equality columns first, then range/sort columns.

## Index Maintenance

### Find Unused Indexes

```sql
SELECT relname AS table,
       indexrelname AS index,
       pg_size_pretty(pg_relation_size(i.indexrelid)) AS size,
       idx_scan
FROM pg_stat_user_indexes i
JOIN pg_index USING (indexrelid)
WHERE idx_scan = 0
  AND NOT indisunique
  AND NOT indisprimary
ORDER BY pg_relation_size(i.indexrelid) DESC;
```

> Reset stats with `SELECT pg_stat_reset();` then wait days/weeks before trusting `idx_scan = 0`.

### Find Duplicate Indexes

```sql
SELECT array_agg(indexrelid::regclass) AS indexes,
       indrelid::regclass AS table,
       indkey AS columns
FROM pg_index
GROUP BY indrelid, indkey
HAVING count(*) > 1;
```

### Index Size

```sql
-- All indexes with sizes
SELECT tablename, indexname,
       pg_size_pretty(pg_relation_size(indexname::regclass)) AS size
FROM pg_indexes
WHERE schemaname = 'public'
ORDER BY pg_relation_size(indexname::regclass) DESC;

-- Total index size vs table size
SELECT relname,
       pg_size_pretty(pg_table_size(relid)) AS table_size,
       pg_size_pretty(pg_indexes_size(relid)) AS index_size
FROM pg_stat_user_tables
ORDER BY pg_indexes_size(relid) DESC;
```

### Rebuild Bloated Indexes

```sql
-- Concurrent reindex (PG 12+, no lock)
REINDEX INDEX CONCURRENTLY idx_orders_customer_id;

-- All indexes on a table
REINDEX TABLE CONCURRENTLY orders;
```

## Index-Only Scans

The fastest scan type — reads only the index, never touches the table. Requirements:

1. All selected columns are in the index (key or INCLUDE)
2. The visibility map is up-to-date (run `VACUUM` if you see "Heap Fetches" in EXPLAIN)

```sql
-- Check if Index Only Scan is happening
EXPLAIN SELECT customer_id FROM orders WHERE customer_id = 42;
-- Look for: "Index Only Scan" and "Heap Fetches: 0"
```

If you see high "Heap Fetches", run `VACUUM orders;` to update the visibility map.

## GIN Deep Dive

### How GIN works internally

A regular B-tree index maps one value per row to a sorted lookup. GIN (Generalised Inverted Index) handles columns where a single row contains **multiple values** (arrays, JSONB, tsvector). It inverts the relationship:

Instead of `row → values`, GIN builds `value → list of rows`:

```
-- Given:
station 1: ['market', 'cloning', 'repair']
station 2: ['market', 'reprocessing-plant']
station 3: ['cloning', 'insurance']

-- GIN index stores:
'market'             → [station 1, station 2]
'cloning'            → [station 1, station 3]
'repair'             → [station 1]
'reprocessing-plant' → [station 2]
'insurance'          → [station 3]
```

Query `WHERE services @> ARRAY['market']` goes straight to the 'market' entry — no full scan.

### `@>` containment vs path extraction

For JSONB, there are two ways to query a field — they look similar but behave differently with indexes:

```sql
-- Containment (@>) — uses GIN index
WHERE data @> '{"race": "caldari"}'
WHERE data @> '{"station": {"race": "caldari"}}'  -- works at any nesting depth

-- Path extraction — does NOT use GIN index, requires a B-tree expression index
WHERE data->>'race' = 'caldari'
WHERE data->'station'->>'race' = 'caldari'
```

Use `@>` for equality/existence checks. Use path extraction for ranges, NULLs, sorting, or LIKE — and add a B-tree expression index if the query runs frequently:

```sql
CREATE INDEX idx_race ON my_table ((data->>'race'));
```

### When GIN doesn't help (use B-tree instead)

Path extraction queries need B-tree expression indexes:

```sql
-- Range comparison — can't use @>
WHERE data->>'temperature' > '5000'

-- NULL check
WHERE data->>'deleted_at' IS NULL

-- Sorting
ORDER BY data->>'name'

-- LIKE pattern
WHERE data->>'name' LIKE '%jita%'
```

### Array columns: intentional denormalisation

Storing a fixed set of values as `TEXT[]` with a GIN index is a valid alternative to a junction table when:

- The data is read-only or rarely updated (no update anomaly risk)
- The values are a fixed known enum (no inconsistency risk)
- The query pattern is simple containment ("has this station got service X?")
- Scale is small enough that a junction table adds complexity without benefit

```sql
-- Store
services TEXT[]

-- Index
CREATE INDEX idx_station_services ON stations USING GIN (services);

-- Query
WHERE services @> ARRAY['reprocessing-plant']
```

GIN on `TEXT[]` supports exact element matching only — not partial text search. For partial search on array elements, create a TSVECTOR from the array and GIN index that instead.

### `jsonb_path_ops` variant

A smaller, faster GIN index that only supports `@>` containment:

```sql
CREATE INDEX idx_data_path ON my_table USING GIN (data jsonb_path_ops);
```

Use when you only need containment queries and want a leaner index. Does not support `?` (key existence) or `?|`/`?&` operators.

---

## When NOT to Index

- Small tables (< few thousand rows) — Seq Scan is faster
- Columns with very low cardinality (e.g., boolean) — unless used in a partial index
- Write-heavy tables with rarely queried columns — indexes slow down INSERT/UPDATE/DELETE
- Columns only used with functions that don't match an expression index
