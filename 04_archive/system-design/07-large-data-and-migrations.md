# Large Data and Migrations

The previous lessons deal with request/response — a user asks for something, you return it in milliseconds. But some work doesn't fit that model. Migrating 500 million rows to a new schema. Backfilling a column across 200 GB of data. Streaming events from one system to another continuously. Exporting a year of transactions to a data warehouse nightly.

This is a different world. The operations take minutes, hours, or days instead of milliseconds. The failure modes are different — you don't retry a 4-hour job from scratch. The bottleneck isn't latency, it's throughput. And the hardest constraint is often "don't kill production while doing it."

## The Numbers

Before designing anything, know how long things actually take. Most people drastically underestimate how long it takes to move large amounts of data.

### Transfer speeds

| Operation                       | Speed         | Time for 1 GB | Time for 1 TB |
| ------------------------------- | ------------- | ------------- | ------------- |
| Sequential read from SSD        | ~500 MB/s     | 2 sec         | 34 min        |
| Sequential write to SSD         | ~400 MB/s     | 2.5 sec       | 43 min        |
| Network (same AZ, 10 Gbps)      | ~1 GB/s       | 1 sec         | 17 min        |
| Network (cross-region)          | ~100-300 MB/s | 3-10 sec      | 1-3 hours     |
| Network (public internet)       | ~10-50 MB/s   | 20-100 sec    | 6-28 hours    |
| S3 upload (single stream)       | ~50-100 MB/s  | 10-20 sec     | 3-6 hours     |
| S3 upload (multipart, parallel) | ~500 MB/s+    | 2 sec         | 34 min        |
| pg_dump (medium table)          | ~50-100 MB/s  | 10-20 sec     | 3-6 hours     |

```
                1 sec      10 sec     1 min      10 min     1 hour     10 hours
                │          │          │          │          │          │
  1 GB:
  SSD read      ██         ·          ·          ·          ·          ·
  Same AZ net   █          ·          ·          ·          ·          ·
  Cross-region  ···████    ·          ·          ·          ·          ·
  Public net    ·  ····████████████   ·          ·          ·          ·
  S3 single     ·    ██████████      ·          ·          ·          ·
                │          │          │          │          │          │
  1 TB:
  SSD read      ·          ·     █████████      ·          ·          ·
  Same AZ net   ·          ·       ███████      ·          ·          ·
  Cross-region  ·          ·          · ████████████████    ·          ·
  Public net    ·          ·          ·          ·   ████████████████████████
  S3 single     ·          ·          ·          ·████████████████     ·
  S3 parallel   ·          ·     █████████      ·          ·          ·
                │          │          │          │          │          │
                1 sec      10 sec     1 min      10 min     1 hour     10 hours
```

The mental model: **within a data centre, 1 TB moves in ~30 minutes. Over the internet, 1 TB takes hours.** If you're moving 10+ TB over the internet, seriously consider AWS Snowball or a direct connect line.

### Database scan speeds

Reading data from a database is much slower than reading from disk, because you're going through the query engine, deserialising rows, and (usually) going over the network.

| Operation                                | Speed                     | Time for 1M rows | Time for 100M rows |
| ---------------------------------------- | ------------------------- | ---------------- | ------------------ |
| `SELECT *` full table scan (Postgres)    | ~10,000-50,000 rows/sec   | 20-100 sec       | 30-160 min         |
| `SELECT *` with WHERE on index           | ~5,000-20,000 rows/sec    | 50-200 sec       | 80-330 min         |
| Cursor-based pagination (1000 per batch) | ~5,000-15,000 rows/sec    | 1-3 min          | 2-5 hours          |
| `COPY TO` (Postgres binary export)       | ~100,000-500,000 rows/sec | 2-10 sec         | 3-17 min           |
| DynamoDB Scan                            | ~5,000-15,000 items/sec   | 1-3 min          | 2-5 hours          |
| DynamoDB Scan (parallel, 10 segments)    | ~30,000-100,000 items/sec | 10-30 sec        | 17-55 min          |

```
                10 sec     1 min      10 min     1 hour     5 hours
                │          │          │          │          │
  1M rows:
  Full scan     ████████   ·          ·          ·          ·
  Cursor pages  ··█████████████       ·          ·          ·
  COPY TO       ██         ·          ·          ·          ·
  DynamoDB Scan ··█████████████       ·          ·          ·
                │          │          │          │          │
  100M rows:
  Full scan     ·          ·  ████████████████████████      ·
  Cursor pages  ·          ·          ·   ██████████████████████████
  COPY TO       ··████████████████    ·          ·          ·
  DynamoDB Scan ·          ·          ·   ██████████████████████████
                │          │          │          │          │
                10 sec     1 min      10 min     1 hour     5 hours
```

Key insight: **`COPY` is 10-50x faster than `SELECT *` for bulk reads.** If you're extracting data from Postgres for a migration or export, always prefer `COPY TO` over application-level row-by-row reads. For DynamoDB, parallel scan segments are essential — a single-threaded scan of a large table will take hours.

### Write speeds

| Operation                           | Speed                     | Time for 1M rows | Time for 100M rows |
| ----------------------------------- | ------------------------- | ---------------- | ------------------ |
| Postgres INSERT (single row, no tx) | ~1,000-3,000 rows/sec     | 5-17 min         | 9-28 hours         |
| Postgres INSERT (batches of 1000)   | ~20,000-50,000 rows/sec   | 20-50 sec        | 33-83 min          |
| Postgres `COPY FROM`                | ~100,000-500,000 rows/sec | 2-10 sec         | 3-17 min           |
| DynamoDB BatchWriteItem (25/batch)  | ~5,000-15,000 items/sec   | 1-3 min          | 2-5 hours          |
| S3 PutObject (small files)          | ~100-300 objects/sec      | 1-3 hours        | forever            |
| S3 (Parquet, large files)           | ~500 MB/s                 | depends on size  | depends on size    |

**The 50x rule: batching writes gives ~50x throughput over individual inserts.** Single-row INSERTs at 2,000/sec means 100M rows takes 14 hours. Batched INSERTs at 50,000/sec means 33 minutes. `COPY FROM` at 300,000/sec means 6 minutes. Always batch.

### The quick estimation formula

```
Time = rows / throughput

For a migration or backfill:
  Reading:  rows / 10,000 rows/sec  (cursor-based, safe estimate)
  Writing:  rows / 30,000 rows/sec  (batched inserts, safe estimate)
  Total:    rows / 5,000 rows/sec   (read + transform + write combined)

Quick table:
  1 million rows    / 5,000 = 200 sec      = ~3 minutes
  10 million rows   / 5,000 = 2,000 sec     = ~33 minutes
  100 million rows  / 5,000 = 20,000 sec    = ~5.5 hours
  1 billion rows    / 5,000 = 200,000 sec   = ~2.3 days
```

If your back-of-envelope says "days", you need parallelism, `COPY`, or a different approach entirely.

## Batch vs Stream

The fundamental question: do you process data in chunks on a schedule, or continuously as it arrives?

```
Batch:   Collect data → wait → process all at once → wait → repeat
         ████░░░░░░░░░░░░░░░████░░░░░░░░░░░░░░░░░░████░░░░░░░░

Stream:  Process each event as it arrives, continuously
         █·█··█·█·██··█·█··█·█··██·█·█··█·██·█·█··█·█·█··██·█·
```

|               | Batch                                   | Stream                                        |
| ------------- | --------------------------------------- | --------------------------------------------- |
| Latency       | Minutes to hours (scheduled)            | Seconds to minutes (near real-time)           |
| Complexity    | Lower — simpler failure handling        | Higher — ordering, exactly-once, backpressure |
| Throughput    | Very high (bulk operations)             | Lower per-event, but continuous               |
| Recovery      | Re-run the batch                        | Replay from offset / checkpoint               |
| Typical tools | Cron + scripts, Step Functions, Airflow | Kafka, Kinesis, DynamoDB Streams, SQS         |
| Good for      | Reports, ETL, nightly syncs, backfills  | Event propagation, CDC, real-time dashboards  |

### When to use batch

- Data is only needed periodically (nightly reports, weekly exports)
- Source system can't handle continuous reads
- You need complex transformations that benefit from bulk operations
- The data naturally arrives in batches (file uploads, partner data drops)

### When to use streaming

- Downstream systems need data within seconds/minutes
- You're propagating changes between services (CDC)
- The volume is continuous and steady (clickstream, IoT telemetry)
- You want to decouple producers from consumers

### The hybrid middle ground

Most real systems aren't purely batch or purely stream. The common pattern:

```
Source DB ──CDC──→ Kafka ──→ Consumer (real-time)
                     │
                     └──→ S3 (hourly dumps) ──→ Warehouse (batch)
```

Stream for operational use (keep services in sync), batch for analytical use (load into warehouse). Same data, two consumption patterns.

## Migration Patterns

Moving data from one place to another while the system is live. This is where most of the complexity lives — not in the moving itself, but in doing it without downtime or data loss.

### Pattern 1: Offline migration (simplest)

```
1. Take system down (maintenance window)
2. Export data from old system
3. Transform
4. Import into new system
5. Switch over
6. Bring system up
```

**When it works:** Small datasets (< 10M rows), acceptable downtime window, simple transformation.

**When it doesn't:** Large datasets where export + import takes hours, or zero-downtime requirement.

### Pattern 2: Dual-write

```
                    ┌──→ Old DB (reads still come from here)
App writes to ──────┤
                    └──→ New DB (building up data)

Phase 1: Dual-write. App writes to both. Reads from old.
Phase 2: Backfill. Copy historical data from old to new.
Phase 3: Verify. Compare old and new, fix discrepancies.
Phase 4: Switch reads. App reads from new. Still writes to both.
Phase 5: Cut over. Stop writing to old.
```

**Danger:** Dual writes without a transaction across both systems means they can diverge. If the write to the new DB fails but the old succeeds, you have inconsistency. Mitigate with:

- Reconciliation jobs that compare and fix
- Writing to old first (source of truth), then async replication to new
- Accepting temporary inconsistency and having a "fix-up" phase

### Pattern 3: CDC (Change Data Capture)

```
Old DB ──WAL/binlog──→ CDC tool ──→ Kafka ──→ Consumer ──→ New DB
         (Debezium)                              │
                                          transform here
```

CDC reads the database's write-ahead log (WAL in Postgres, binlog in MySQL) and streams every change as an event. The consumer applies changes to the new system.

**Why this is often the best approach:**

- Zero impact on the source database (reads the WAL, not the tables)
- Captures every change, including deletes
- Can transform data in flight
- Natural backfill: start from a snapshot, then stream changes from that point

**Tools:** Debezium (open source, Kafka Connect), AWS DMS (managed), RDS logical replication.

**The tricky part:** You need to handle the initial snapshot + ongoing stream without missing or duplicating data. The typical approach:

```
1. Take a consistent snapshot (pg_dump or Debezium snapshot mode)
2. Record the WAL position at snapshot time
3. Load the snapshot into the new system
4. Start streaming changes from the recorded WAL position
5. Consumer applies changes — snapshot fills history, stream fills the gap
```

### Pattern 4: Shadow reads

```
                         ┌──→ Old DB ──→ Response (returned to user)
App reads from both ─────┤
                         └──→ New DB ──→ Response (logged, compared, discarded)
```

Before cutting over reads, send read traffic to both systems and compare results. This catches data discrepancies before users see them. Log mismatches, fix the migration, repeat until clean.

## Long-Running Workers vs Event-Driven

A backfill that touches 100M rows will take hours. How you structure that work matters.

### The long-running worker

```
Worker starts
  └──→ Open cursor on source DB
       └──→ Read batch of 1000 rows
            └──→ Transform
                 └──→ Write batch to destination
                      └──→ Checkpoint progress
                           └──→ Read next batch... (repeat for hours)
```

**Problems with long-running workers:**

- If it crashes at hour 3 of a 5-hour job, what happens?
- If you deploy new code, do you kill the running job?
- If the source DB is slow, does the worker back off or hammer it?
- How do you monitor progress? Is it 40% done or stuck?

**Essential features for long-running workers:**

| Feature           | Why                                        | How                                                    |
| ----------------- | ------------------------------------------ | ------------------------------------------------------ |
| Checkpointing     | Resume from where you left off after crash | Store last processed ID/offset in DB or S3             |
| Idempotent writes | Re-processing a batch doesn't corrupt data | Use UPSERT / `ON CONFLICT DO UPDATE`                   |
| Progress tracking | Know if it's running, stuck, or done       | Log progress every N batches, expose metrics           |
| Throttling        | Don't kill the source DB                   | Sleep between batches, limit concurrent connections    |
| Graceful shutdown | Don't lose in-flight work on deploy        | Handle SIGTERM, finish current batch, checkpoint, exit |

### The event-driven alternative

Instead of one worker scanning the entire table, emit events and let consumers process them:

```
Step Function / Airflow
  └──→ Partition the work (e.g., ID ranges 0-1M, 1M-2M, ...)
       └──→ Fan out to N Lambda / ECS tasks in parallel
            └──→ Each processes its partition independently
                 └──→ Report completion
                      └──→ Coordinator waits for all, then marks done
```

**Advantages:** Parallelism (10 workers = ~10x faster), each unit of work is small and retryable, no long-running process to babysit.

**Disadvantages:** Coordination overhead, harder to maintain ordering, more moving parts.

### Decision framework

```
Use a long-running worker when:
  - Order matters (process rows in sequence)
  - The job runs regularly (nightly ETL)
  - Simple operational model (one process to monitor)
  - Data fits through a single pipeline

Use fan-out / event-driven when:
  - Speed matters (need to finish in minutes, not hours)
  - Work is naturally partitionable (by user ID, date, region)
  - Individual items are independent (no ordering requirement)
  - You need to scale beyond what one worker can do
```

### The numbers

```
Single worker, cursor-based:
  100M rows / 5,000 rows/sec = ~5.5 hours

10 parallel workers, partitioned by ID range:
  10M rows each / 5,000 rows/sec = ~33 minutes total

50 Lambda functions, partitioned:
  2M rows each / 5,000 rows/sec = ~7 minutes total
```

Parallelism helps linearly — until you hit the destination's write throughput or the source's read capacity. 50 workers each doing 5,000 reads/sec = 250,000 reads/sec on the source. Your Postgres instance probably can't handle that. Always calculate both the per-worker rate and the aggregate load on shared resources.

## Protecting Production

The hardest part of large data operations isn't the data — it's not breaking the live system while moving it.

### Throttling

Never run a migration at full speed against a production database. The migration will consume connections, I/O, and CPU that serve live traffic.

```
Unthrottled:
  Migration:  ████████████████████████████████████████
  Live traffic: ██░░██░░░░██░░░░██░░░░  (starved)

Throttled (sleep 100ms between batches):
  Migration:  ████░░████░░████░░████░░████░░████░░████
  Live traffic: ██████████████████████████████████████  (healthy)
```

Rules of thumb:

- Keep migration to < 20% of database capacity
- Monitor active connections, CPU, and replication lag during migration
- Add sleep between batches (start with 100ms, adjust based on load)
- Run during low-traffic hours if possible

### Read replicas for extraction

Don't read from the primary for large scans. Use a read replica:

```
Primary (serves live traffic)
  │
  └──replication──→ Replica (migration reads from here)
                      │
                      └──→ Migration worker
```

This isolates the migration's read load from production. Watch replication lag — if the migration puts too much load on the replica, it falls behind the primary.

### Backpressure

If your pipeline is: Source → Queue → Worker → Destination, and the destination is slow, the queue fills up. Without backpressure, you either run out of memory or start dropping data.

Backpressure strategies:

- **Fixed-size queue:** Producer blocks when queue is full (worker pulls, processes, then allows more)
- **Rate limiting:** Producer emits at a fixed rate regardless of consumer speed
- **Consumer-driven:** Consumer requests work when ready (pull model vs push model)

## Common ETL Patterns

### Database to database (schema migration, platform migration)

```
Source DB ──COPY TO──→ CSV/Parquet ──transform──→ COPY FROM──→ Destination DB
```

For simple schema changes (add column, rename, change type), prefer `ALTER TABLE` with backfill over export/import. For cross-platform (Postgres to DynamoDB), export + transform + load is usually the path.

### Database to warehouse (analytics ETL)

```
Nightly:
  Source DB ──COPY TO──→ S3 (Parquet) ──→ Redshift/BigQuery COPY

Streaming:
  Source DB ──CDC──→ Kafka ──→ S3 (micro-batches) ──→ Warehouse
```

Parquet is the format of choice for analytical workloads — columnar, compressed, splittable. A 10 GB CSV becomes ~2-3 GB Parquet.

### Service-to-service data sync

```
Service A ──events──→ Kafka ──→ Service B (maintains its own view)
```

Each service owns its data. When Service B needs data from Service A, it consumes events and builds a local projection. This is eventual consistency — Service B may be seconds behind. If that's unacceptable, use a synchronous API call instead.

## Worked Example: Migrating 200M Orders to a New Schema

**Given:** E-commerce platform, 200 million order rows (~100 GB), need to split the `orders` table into `orders` + `order_items` (normalisation). System must stay live.

**Estimation:**

```
200M rows at ~500 bytes each = ~100 GB

Reading (COPY TO):    200M / 200,000 rows/sec = ~17 minutes
Transform:            CPU-bound, ~100,000 rows/sec = ~33 minutes
Writing (COPY FROM):  200M / 200,000 rows/sec = ~17 minutes
Total (sequential):   ~67 minutes

With 10 parallel workers (partitioned by order_id range):
  20M rows each, ~7 minutes per worker
  Total: ~10 minutes (limited by destination write throughput)
```

**Plan:**

```
Phase 1: Prepare (no user impact)
  - Create new tables (orders_v2, order_items)
  - Deploy code that can read from old OR new schema (feature flag)
  - Build and test the migration script against a replica

Phase 2: Backfill historical data
  - Read from replica (not primary)
  - Partition by order_id range, fan out to 10 workers
  - Each worker: read batch → split order + items → write to new tables
  - Checkpoint after each batch (store last processed order_id)
  - Throttle to keep replica lag < 1 second
  - Duration: ~10-15 minutes

Phase 3: Catch up + dual write
  - Enable dual-write: new orders write to both old and new schema
  - Run catch-up migration for orders created during Phase 2
  - This gap is small (15 min of orders = maybe 50K rows = seconds)

Phase 4: Verify
  - Shadow reads: query both old and new, compare results
  - Run reconciliation: count rows, compare totals, spot-check
  - Fix any discrepancies

Phase 5: Switch reads
  - Flip feature flag: reads now come from new schema
  - Old schema still receives writes (safety net)
  - Monitor error rates, latencies

Phase 6: Clean up
  - Stop dual-writing to old schema
  - Keep old tables for N days as backup
  - Drop old tables
```

**What can go wrong:**

| Risk                              | Mitigation                                                                |
| --------------------------------- | ------------------------------------------------------------------------- |
| Migration worker crashes mid-way  | Checkpointing + idempotent writes (UPSERT). Restart from last checkpoint. |
| New orders arrive during backfill | Dual-write catches them. Catch-up phase handles the gap.                  |
| Data mismatch between old and new | Shadow reads + reconciliation job before switching.                       |
| Migration overloads production DB | Read from replica. Throttle writes. Monitor and pause if needed.          |
| Rollback needed after switch      | Old tables still exist. Flip feature flag back.                           |

## Key Takeaways

**1. Know the numbers.** Moving 1 TB within a data centre takes ~30 minutes. Scanning 100M database rows takes hours without `COPY` or parallelism. Single-row inserts are 50x slower than batched. Always estimate before starting.

**2. Batch for throughput, stream for freshness.** If the data can be hours old, batch it. If downstream needs it in seconds, stream it. Most systems use both — stream for operations, batch for analytics.

**3. Protect production.** Throttle, read from replicas, monitor lag. The migration itself is never the hard part — doing it without affecting live users is.

**4. Checkpoint everything.** Any job that runs for more than a few minutes must be resumable. Store progress externally. Make writes idempotent. Assume the process will crash and design for it.

**5. Parallelism scales linearly until it doesn't.** 10 workers = ~10x faster, but only until you saturate the source or destination. Calculate aggregate load on shared resources, not just per-worker throughput.
