# Cost and Storage Lifecycle

The previous lessons cover what to use — Postgres, Redis, S3, queues, app servers. But once a system is running, the biggest ongoing problem isn't "does it work?" — it's "why is the bill $8,000 this month?" Cost isn't a one-time decision. Data grows, instances run 24/7, and nobody deletes anything. Without a strategy, cloud bills grow linearly with time even when traffic doesn't.

This lesson covers: what things cost, when to move data between storage tiers, how to right-size instances, and the common mistakes that quietly double your bill.

## The Cost of Storage

Storage is the sneaky one. Compute is obvious — you see the instance running. Storage accumulates silently. 9 TB/year of journey data doesn't hurt in month 1 (750 GB, ~$17 on S3). By year 3, it's 27 TB ($621/month). By year 5, it's 45 TB ($1,035/month) — and that's the cheap option.

### Storage tier comparison

Think of four tiers, each roughly 3-5x cheaper than the one above:

| Tier                               | Cost/TB/month | Access                  | Jump        | When to use                         |
| ---------------------------------- | ------------- | ----------------------- | ----------- | ----------------------------------- |
| **Database** (RDS)                 | ~$100         | Milliseconds, queryable | —           | Data you query with SQL constantly  |
| **Object store** (S3 Standard)     | ~$25          | 50-100 ms, by key       | ~4x cheaper | Files, exports, data accessed by ID |
| **Infrequent access** (S3 IA)      | ~$12          | 50-100 ms, by key       | ~2x cheaper | Data accessed < 1x/month            |
| **Archive** (Glacier/Deep Archive) | ~$1-4         | Minutes to hours        | ~4x cheaper | Compliance, backups, "just in case" |

```
  Tier              $/TB/mo    Access speed       Step down
  ─────────────────────────────────────────────────────────
  Database (RDS)     $100      ms (SQL queries)
                                                   ~4x cheaper
  Object (S3)        $25       50-100 ms (by key)
                                                   ~2x cheaper
  Infrequent (S3 IA) $12       50-100 ms (by key)
                                                   ~4x cheaper
  Archive (Glacier)  $1-4      minutes to hours
  ─────────────────────────────────────────────────────────
  Total spread: ~100x between database and deep archive
```

The mental model: **each tier down is roughly 2-4x cheaper.** The question is always "how fast do I need this data, and how often?"

The Glacier sub-tiers (Instant at $4, Flexible at $3.60, Deep Archive at $1) are variations within the archive tier — pick based on how quickly you need retrieval when you do access it. For estimation purposes, treat archive as "$1-4/TB/month" and move on.

### What goes where

| Data                                          | Storage tier             | Why                                                         |
| --------------------------------------------- | ------------------------ | ----------------------------------------------------------- |
| Active database tables (recent orders, users) | RDS                      | Need millisecond queries, joins, transactions               |
| Application logs (last 30 days)               | S3 Standard              | Frequently searched, Athena queries                         |
| Application logs (older than 30 days)         | S3 IA or Glacier Instant | Rarely searched, but need access for debugging              |
| Uploaded files, images, documents             | S3 Standard              | Served to users, CDN origin                                 |
| Database backups                              | S3 IA                    | Only accessed during disaster recovery                      |
| Audit logs, compliance data                   | Glacier Deep Archive     | Legal requirement, never read unless audited                |
| Old database rows (orders > 2 years)          | S3 as Parquet            | Queryable with Athena when needed, not worth keeping in RDS |

### S3 lifecycle policies

Don't manually move data between tiers. Set lifecycle rules and forget about it:

```json
{
  "Rules": [
    {
      "ID": "age-off-logs",
      "Filter": { "Prefix": "logs/" },
      "Transitions": [
        { "Days": 30, "StorageClass": "STANDARD_IA" },
        { "Days": 90, "StorageClass": "GLACIER_IR" },
        { "Days": 365, "StorageClass": "DEEP_ARCHIVE" }
      ],
      "Expiration": { "Days": 2555 }
    }
  ]
}
```

```
Day 0          Day 30         Day 90         Day 365        Day 2555 (7 years)
│              │              │              │              │
S3 Standard ──→ S3 IA ───────→ Glacier ─────→ Deep Archive → Deleted
$23/TB/mo      $12.50/TB/mo   $4/TB/mo       $1/TB/mo       $0
```

This is how you keep 7 years of logs without the cost growing linearly. Year 1 logs cost $1/TB/month by the time they're in Deep Archive. Without lifecycle policies, they'd still cost $23/TB/month sitting in S3 Standard.

### The cost of keeping everything in Postgres

Worked example — your journey pipeline after 5 years (180 GB metadata/year, 9 TB journey data/year):

**Scenario A: Everything in Postgres**

| Component    | Data                                            | Cost/month  |
| ------------ | ----------------------------------------------- | ----------- |
| RDS instance | db.r6g.4xlarge (most data is cold, wasting RAM) | ~$1,600     |
| RDS storage  | 46 TB (900 GB metadata + 45 TB journeys)        | ~$5,290     |
| **Total**    |                                                 | **~$6,890** |

**Scenario B: Hot/cold split**

| Component       | Data                                | Tier        | Cost/month |
| --------------- | ----------------------------------- | ----------- | ---------- |
| RDS instance    | db.r6g.large (only hot data, small) | —           | ~$200      |
| RDS storage     | 180 GB metadata (last 12 months)    | Database    | ~$21       |
| Old metadata    | 720 GB as Parquet                   | S3 Standard | ~$17       |
| Recent journeys | 9 TB (last 12 months)               | S3 Standard | ~$207      |
| Older journeys  | 27 TB (years 2-4)                   | S3 IA       | ~$337      |
| Oldest journeys | 9 TB (year 5)                       | Glacier     | ~$36       |
| **Total**       |                                     |             | **~$818**  |

**$6,890 vs $818/month.** Same data, same queryability for recent data, **8x cheaper.**

## The Cost of Compute

Storage accumulates. Compute just runs — and the biggest waste is instances running 24/7 when they don't need to.

### Instance pricing models

| Model             | Discount      | Commitment                     | Use case                                        |
| ----------------- | ------------- | ------------------------------ | ----------------------------------------------- |
| On-demand         | 0% (baseline) | None                           | Unpredictable workloads, short-lived            |
| Reserved (1 year) | ~30-40% off   | 1 year, specific instance type | Steady-state databases, always-on services      |
| Reserved (3 year) | ~50-60% off   | 3 years                        | You're sure about the workload                  |
| Savings Plans     | ~30-40% off   | 1 year, $ commitment           | Flexible — any instance family in a region      |
| Spot              | ~60-90% off   | None, can be interrupted       | Batch processing, fault-tolerant workers, CI/CD |

```
On-demand:    $1.00  ████████████████████████████████████████
Reserved 1yr: $0.65  ██████████████████████████
Reserved 3yr: $0.42  █████████████████
Spot:         $0.20  ████████
```

### What should run on what

| Workload                          | Pricing model                        | Why                                          |
| --------------------------------- | ------------------------------------ | -------------------------------------------- |
| Production database (RDS)         | Reserved 1yr or 3yr                  | Always on, predictable, can't be interrupted |
| Production app servers (baseline) | Reserved or Savings Plan             | Minimum capacity always needed               |
| Production app servers (burst)    | On-demand with auto-scaling          | Scales up for peak, scales down after        |
| Redis/ElastiCache                 | Reserved                             | Always on                                    |
| Batch processing / ETL workers    | Spot                                 | Interruptible, can retry, massive savings    |
| CI/CD runners                     | Spot                                 | Short-lived, retryable                       |
| Dev/staging environments          | On-demand, shut down nights/weekends | Only used during working hours               |
| Lambda                            | Pay per invocation                   | Naturally matches demand, no idle cost       |

### The dev/staging waste

A common blind spot. Your production setup has a staging environment that mirrors it:

```
Staging (running 24/7):
  db.r6g.large:    $200/month
  2x ECS tasks:    $120/month
  Redis:           $150/month
  ALB:             $20/month
  Total:           $490/month

Staging (running 10 hours/day, weekdays only):
  Same instances but only 220 hours/month instead of 730
  Total:           ~$148/month
```

That's $342/month saved on one environment. Multiply by dev, staging, QA, demo — it adds up.

### Lambda vs always-on: when each wins

```
Lambda:  Pay per request. 1 million requests x 200ms x 256MB = ~$3.50/month
ECS:     Pay per hour.    1 task running 24/7 = ~$35/month minimum

Lambda wins when:   < ~5 million requests/month (crossover depends on duration)
ECS wins when:      Steady, high traffic (always utilised)
```

```
Monthly requests:  10K     100K     1M       10M      100M
                   │       │        │        │        │
Lambda cost:       $0.01   $0.10    $3.50    $35      $350
ECS (1 task 24/7): $35     $35      $35      $35      $35 (need more tasks)
                   │       │        │        │        │
                   Lambda wins ────────────→ ECS wins
```

The crossover is roughly where Lambda's per-request cost equals the fixed cost of an always-on container. For low-traffic services (internal tools, webhooks, cron jobs), Lambda is almost free. For high-traffic APIs, it's more expensive than containers.

## Database Archival Strategies

The pattern from the Postgres comfort zone discussion: keep hot data in the database, move cold data out.

### Partition by time, archive by partition

Postgres table partitioning lets you split a table by date range. Each partition is physically separate, so you can archive old partitions without touching recent data:

```sql
-- Create partitioned table
CREATE TABLE orders (
    id BIGINT,
    created_at TIMESTAMP,
    customer_id BIGINT,
    total NUMERIC
) PARTITION BY RANGE (created_at);

-- Create partitions per quarter
CREATE TABLE orders_2025_q1 PARTITION OF orders
    FOR VALUES FROM ('2025-01-01') TO ('2025-04-01');
CREATE TABLE orders_2025_q2 PARTITION OF orders
    FOR VALUES FROM ('2025-04-01') TO ('2025-07-01');
```

Archival process:

```
1. Export old partition:     COPY orders_2024_q1 TO 's3://archive/orders/2024-q1.parquet'
2. Verify export:           Compare row counts and checksums
3. Detach partition:        ALTER TABLE orders DETACH PARTITION orders_2024_q1
4. Drop partition:          DROP TABLE orders_2024_q1
5. Database shrinks:        RDS storage freed, queries faster (less data to scan)
```

Queries on recent data stay fast because Postgres only scans relevant partitions. Queries on archived data go through Athena against S3.

### The Parquet + Athena pattern

Parquet is a columnar format — compressed, splittable, and fast for analytical queries. When you archive Postgres data to S3 as Parquet, you can still query it:

```
                     Hot path (recent data)
User query ─────────→ Postgres ─────→ Response (fast, milliseconds)

                     Cold path (old data)
Analyst query ──────→ Athena ──→ S3 Parquet ──→ Response (seconds to minutes)
```

Athena costs $5 per TB scanned. A well-partitioned, columnar dataset means you scan very little data per query. Querying 1 GB of Parquet costs $0.005.

### Retention policy framework

| Data type                        | Hot (Postgres) | Warm (S3 Standard) | Cold (Glacier) | Delete                     |
| -------------------------------- | -------------- | ------------------ | -------------- | -------------------------- |
| Transactional (orders, payments) | 12 months      | 12-36 months       | 36-84 months   | 7 years (compliance)       |
| User profiles                    | While active   | —                  | —              | Account deletion + 30 days |
| Application logs                 | 7-30 days      | 30-90 days         | 90-365 days    | 1-2 years                  |
| Analytics events                 | 30 days        | 12 months          | —              | 2 years                    |
| Audit logs                       | 90 days        | 12 months          | 12-84 months   | 7 years (compliance)       |

The exact numbers depend on compliance requirements (GDPR, PCI, SOX), but the pattern is universal: recent data is hot and expensive, old data is cold and cheap, very old data is either archived or deleted.

## Common Cost Mistakes

### Running everything on-demand

**The mistake:** Never buying reserved instances because "we might change things."

**The cost:** 30-60% premium on your entire infrastructure. A db.r6g.xlarge on-demand costs ~$400/month. Reserved 1-year: ~$260/month. That's $1,680/year saved on one instance.

**The fix:** Anything that's been running for 3+ months and will keep running → reserved or savings plan. The break-even point on a 1-year reserved instance is ~7-8 months.

### Never deleting anything

**The mistake:** "We might need it someday" applied to every log, every backup, every temporary file.

**The cost:** Linear growth forever. 10 TB of logs nobody reads at $23/TB/month = $230/month = $2,760/year for nothing.

**The fix:** Lifecycle policies from day one. Define retention per data type. Automate deletion. If someone needs 3-year-old logs, they can wait 12 hours for a Glacier restore.

### Over-provisioning for peak

**The mistake:** Sizing everything for worst-case traffic 24/7.

**The cost:** An r6g.4xlarge running at 10% utilisation. You're paying for 16 cores and using 1.6.

**The fix:** Auto-scaling for app servers. Right-size databases based on actual metrics (CloudWatch), not fear. Review instance utilisation monthly.

### Ignoring data transfer costs

**The mistake:** Designing a system where services in different regions or AZs chat constantly.

**The cost:** AWS charges for data transfer between AZs (~$0.01/GB each way) and between regions (~$0.02/GB). At high throughput:

| Traffic      | Cross-AZ cost | Cross-region cost |
| ------------ | ------------- | ----------------- |
| 1 TB/month   | $20           | $40               |
| 10 TB/month  | $200          | $400              |
| 100 TB/month | $2,000        | $4,000            |

**The fix:** Keep chatty services in the same AZ when possible. Use VPC endpoints for S3/DynamoDB (free within region). Cache aggressively to reduce cross-service calls.

### Lambda for everything

**The mistake:** Using Lambda for high-traffic, long-running workloads because "serverless is cheaper."

**The cost:** Lambda at 50 million requests/month x 500ms x 512MB = ~$420/month. Two ECS tasks handling the same load: ~$70/month.

**The fix:** Lambda for low-traffic, bursty, or event-driven workloads. Containers for steady, high-throughput workloads. Check the crossover math.

## Cost Estimation Quick Reference

For back-of-envelope cost estimates:

```
Compute:
  Small ECS task (0.5 vCPU, 1 GB):    ~$15-20/month
  Medium ECS task (1 vCPU, 2 GB):      ~$35-40/month
  Lambda (1M invocations, 200ms):      ~$3.50/month

Database:
  RDS db.t3.medium (dev/small prod):   ~$50/month
  RDS db.r6g.large (production):       ~$200/month
  RDS db.r6g.xlarge (medium prod):     ~$400/month
  DynamoDB (on-demand, 1M reads/day):  ~$7.50/month

Cache:
  ElastiCache cache.t3.medium:         ~$45/month
  ElastiCache cache.r6g.large:         ~$150/month

Storage:
  S3 Standard (per TB):                ~$23/month
  S3 IA (per TB):                      ~$12.50/month
  RDS storage (per TB):                ~$115/month

Network:
  ALB:                                 ~$20/month + traffic
  NAT Gateway:                         ~$32/month + $0.045/GB
  CloudFront (1 TB transfer):          ~$85/month
```

## Key Takeaways

**1. Storage is the silent killer.** It grows every day and nobody notices until the bill arrives. Set lifecycle policies and retention rules from day one, not when it's already 50 TB.

**2. Keep hot data close, move cold data away.** Recent data in Postgres (fast, expensive). Old data in S3 as Parquet (queryable with Athena, cheap). Ancient data in Glacier (almost free). The 115x cost difference between RDS and Glacier Deep Archive is real.

**3. Match pricing to workload pattern.** Reserved for steady-state. On-demand for burst. Spot for batch. Lambda for low-traffic. The wrong pricing model on the right architecture still wastes money.

**4. Right-size, don't over-provision.** Check actual utilisation before sizing up. A database at 15% CPU doesn't need a bigger instance — it might need a smaller one. Review monthly.

**5. Design for cost from the start.** Where data lands (RDS vs S3), how long it stays, and what pricing model instances use — these decisions compound over years. A $200/month difference today is $12,000 over 5 years.
