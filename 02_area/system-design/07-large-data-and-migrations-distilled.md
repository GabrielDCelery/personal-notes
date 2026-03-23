# Large Data and Migrations

The challenge with large data isn't the moving — it's estimating how long it takes, then doing it without killing production.

## The Core Mental Model: The Throughput Ladder

Everything in this domain comes down to one formula: **Time = rows / throughput**. The throughput you get depends entirely on how you read and write. Know the ladder, pick your rung, estimate first.

```
Read throughput:
  SELECT * cursor (1000/batch)  ~10,000 rows/sec   ████
  COPY TO (Postgres binary)    ~200,000 rows/sec   ████████████████████████████████████████

Write throughput:
  INSERT single-row            ~2,000 rows/sec     █
  INSERT batched (1000/batch)  ~35,000 rows/sec    ████████████████
  COPY FROM                   ~200,000 rows/sec    ████████████████████████████████████████
```

COPY is 10–50x faster than SELECT for reads, and batching is 50x faster than single-row INSERTs for writes. Use `COPY` whenever you can.

### Quick estimation

```
Conservative safe estimates (read + transform + write):
  rows / 5,000  = total seconds

  1M rows   → 200 sec   (~3 min)
  10M rows  → 2,000 sec (~33 min)
  100M rows → 20,000 sec (~5.5 hours)
  1B rows   → 200,000 sec (~2.3 days)
```

If the estimate says "days", you need parallelism, `COPY`, or a fundamentally different approach.

## Transfer Speeds

Network location matters more than most people expect. Moving data across regions is 10x slower than within a datacentre; moving over the public internet is 100x slower.

```
                1 sec    10 sec   1 min    10 min   1 hr     10 hr
                │        │        │        │        │        │
  1 GB:
  SSD read      ██       ·        ·        ·        ·        ·
  Same AZ net   █        ·        ·        ·        ·        ·
  Cross-region  ···████  ·        ·        ·        ·        ·
  Public net    ·  ·····██████████·        ·        ·        ·
                │        │        │        │        │        │
  1 TB:
  SSD read      ·        ·   █████████     ·        ·        ·
  Same AZ net   ·        ·     ███████     ·        ·        ·
  Cross-region  ·        ·        · ███████████████·        ·
  Public net    ·        ·        ·        ·  ██████████████████████
  S3 parallel   ·        ·   █████████     ·        ·        ·
                │        │        │        │        │        │
```

**Within a datacentre, 1 TB moves in ~30 minutes. Over the public internet, 1 TB takes hours.** For 10+ TB cross-region, consider AWS Snowball or Direct Connect rather than fighting network throughput.

| Operation                 | Speed         | 1 GB     | 1 TB      |
| ------------------------- | ------------- | -------- | --------- |
| SSD sequential read       | ~500 MB/s     | 2 sec    | 34 min    |
| Same AZ network (10 Gbps) | ~1 GB/s       | 1 sec    | 17 min    |
| Cross-region network      | ~100–300 MB/s | 3–10 sec | 1–3 hours |
| Public internet           | ~10–50 MB/s   | 20+ sec  | 6–28 hrs  |
| S3 multipart + parallel   | ~500 MB/s     | 2 sec    | 34 min    |

## Database Read and Write Speeds

Reading through the query engine is far slower than raw disk I/O — you pay for serialisation, the query planner, and the network hop. This is why `COPY` is the right tool for bulk extraction.

```
  100M rows:
                10 sec  1 min   10 min  1 hr    5 hr
                │       │       │       │       │
  COPY TO       ·██████████████ ·       ·       ·
  Full scan     ·       · ████████████████████  ·
  Cursor pages  ·       ·       ·  █████████████████████
                │       │       │       │       │
```

For writes, the gap is even larger — single-row INSERTs are almost always a mistake at scale:

```
  100M rows:
                1 min   1 hr    10 hr   1 day
                │       │       │       │
  COPY FROM     ████    ·       ·       ·
  Batch INSERT  ·  ████████████ ·       ·
  Single INSERT ·       ·    ██████████████████
                │       │       │       │
```

| Operation                       | Speed             | 1M rows | 100M rows |
| ------------------------------- | ----------------- | ------- | --------- |
| `SELECT *` full scan (Postgres) | ~30,000 rows/sec  | ~30 sec | ~1 hr     |
| Cursor pagination (1000/batch)  | ~10,000 rows/sec  | ~2 min  | ~3 hrs    |
| `COPY TO` binary export         | ~200,000 rows/sec | ~5 sec  | ~8 min    |
| INSERT single-row               | ~2,000 rows/sec   | ~8 min  | ~14 hrs   |
| INSERT batched (1000/batch)     | ~35,000 rows/sec  | ~30 sec | ~48 min   |
| `COPY FROM`                     | ~200,000 rows/sec | ~5 sec  | ~8 min    |

**Always batch writes. Always prefer `COPY` for bulk reads and loads.**

## Batch vs Stream

Batch collects data and processes it in scheduled chunks; streaming processes each event as it arrives. The choice turns on how fresh the data needs to be.

```
Batch:   ████░░░░░░░░░░░░░░░░████░░░░░░░░░░░░░░░░████░░░░░░░░
Stream:  █·█··█·█·██··█·█··█·█··██·█·█··█·██·█·█··█·

Batch: simple to build, high throughput, minutes-to-hours latency
Stream: complex to build, continuous, seconds-to-minutes latency
```

| Dimension  | Batch                             | Stream                                       |
| ---------- | --------------------------------- | -------------------------------------------- |
| Latency    | Minutes to hours                  | Seconds to minutes                           |
| Complexity | Low — re-run the batch on failure | High — ordering, exactly-once, backpressure  |
| Throughput | Very high (bulk ops)              | Lower per-event, continuous                  |
| Recovery   | Re-run the batch                  | Replay from offset / checkpoint              |
| Use for    | Reports, nightly ETL, backfills   | CDC, event propagation, real-time dashboards |

Most real systems use both: **stream for operational use, batch for analytical**. The same data, two consumption paths.

```
Source DB ──CDC──→ Kafka ──→ Consumer (operational, real-time)
                     │
                     └──→ S3 (hourly) ──→ Warehouse (analytical, batch)
```

## Migration Patterns

Moving data while the system is live is the hard part. Here are the four patterns, ordered from simplest to most capable.

### Offline migration (simplest)

Take the system down, export, transform, import, switch, bring back up. Works for small datasets (< 10M rows) or acceptable downtime windows. Non-starter for large data or zero-downtime requirements.

### Dual-write

```
App writes ──→ Old DB (reads come from here)
           └──→ New DB (building up data)

Phase 1: Dual-write. Reads from old.
Phase 2: Backfill historical data from old → new.
Phase 3: Verify. Compare, reconcile.
Phase 4: Switch reads to new. Still dual-writing.
Phase 5: Stop writing to old.
```

**Danger:** No atomic transaction across both systems — they can diverge on partial failure. Mitigate by writing old first (source of truth), then async sync to new, with a reconciliation job that patches gaps.

### CDC (Change Data Capture)

The cleanest approach for live migrations. Reads the database's WAL (Debezium + Kafka, or AWS DMS) instead of querying tables — zero impact on production reads.

```
Old DB ──WAL──→ Debezium ──→ Kafka ──→ Consumer ──→ New DB
                                           │
                                     transform here
```

The bootstrap sequence: snapshot at a known WAL position → load snapshot → stream changes from that WAL position forward. No gaps, no duplicates.

### Shadow reads

Before cutting over, send reads to both systems and compare. Log mismatches, fix the migration, repeat until clean. The last safety net before switching.

## Long-Running Workers

A worker that runs for hours needs different engineering than one that handles requests.

**Essential properties:**

| Property          | Why it matters                        | How                                                   |
| ----------------- | ------------------------------------- | ----------------------------------------------------- |
| Checkpointing     | Crash at hour 3 → resume, not restart | Store last processed ID/offset in DB or S3            |
| Idempotent writes | Safe to re-process a batch            | UPSERT / `ON CONFLICT DO UPDATE`                      |
| Progress tracking | Know if it's running, stuck, or done  | Log every N batches, expose metrics                   |
| Throttling        | Don't starve live traffic             | Sleep between batches, cap concurrent connections     |
| Graceful shutdown | Don't lose in-flight work on deploy   | Catch SIGTERM, finish current batch, checkpoint, exit |

### Fan-out for speed

When you need hours cut to minutes, partition the work and run it in parallel:

```
Coordinator
  └──→ Partition by ID range (0–10M, 10M–20M, ...)
       └──→ N workers process their slice independently
            └──→ Each small and retryable; coordinator waits for all
```

```
Single worker, cursor:  100M / 5,000/sec = ~5.5 hours
10 parallel workers:    10M each / 5,000/sec = ~33 minutes
50 Lambda functions:    2M each / 5,000/sec = ~7 minutes
```

Parallelism scales linearly — until you saturate the source or destination. 50 workers each doing 5,000 reads/sec = 250,000 reads/sec on your Postgres primary. **Always calculate aggregate load, not just per-worker rate.**

## Protecting Production

### Throttle aggressively

An unthrottled migration consumes connections, I/O, and CPU that live traffic needs.

```
Unthrottled:
  Migration:     ████████████████████████████████████████
  Live traffic:  ██░░██░░░░██░░  (starved, latency spikes)

Throttled (100ms sleep between batches):
  Migration:     ████░░████░░████░░████░░████░░████░░████
  Live traffic:  ████████████████████████████████████████
```

Keep migration load below 20% of database capacity. Monitor CPU, active connections, and replication lag throughout.

### Read from replicas, not the primary

```
Primary (serves live traffic)
  └──replication──→ Replica ──→ Migration worker
```

Isolates migration read load entirely. Watch replica lag — if the migration overloads the replica, it falls behind.

### Backpressure

If the destination is slow, the queue fills up. Prevent data loss by: blocking producers when the queue is full (push), limiting producer rate regardless of consumer speed (rate-limit), or having the consumer request work when ready (pull). Pull models are generally safer.

---

## Key Mental Models

1. **Time = rows / throughput.** Estimate before starting. 100M rows at 5,000/sec = 5.5 hours.
2. **COPY is 10–50x faster than SELECT.** For bulk reads and loads, bypass the query engine.
3. **Batching is 50x faster than single-row INSERTs.** Never insert rows one at a time at scale.
4. **1 TB within a datacentre = ~30 minutes. Over the internet = hours.** Location determines transfer time.
5. **Batch for throughput, stream for freshness.** Most systems use both for the same data.
6. **CDC is the cleanest live migration tool.** WAL reads don't touch production tables; snapshot + stream covers history and ongoing changes.
7. **Dual-write requires reconciliation.** No atomic guarantee across two systems means divergence is possible — plan for it.
8. **Checkpoint everything that runs longer than a few minutes.** Assume crashes. Design for restart, not retry from zero.
9. **Parallelism scales linearly until the shared resource saturates.** Calculate aggregate load, not per-worker load.
10. **Throttle migrations to < 20% of DB capacity.** The migration is never the point — protecting live traffic is.
