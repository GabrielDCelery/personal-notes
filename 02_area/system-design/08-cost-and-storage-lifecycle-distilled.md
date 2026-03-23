# Cost and Storage Lifecycle

Cloud bills don't grow with traffic — they grow with time, because storage accumulates silently and idle compute keeps running.

## The Core Mental Model: The Four-Tier Ladder

Every storage decision is a trade-off between access speed and cost. There are four tiers, each roughly 2–4x cheaper than the one above — a 100x spread from top to bottom.

| Tier                          | $/TB/mo | Access            | Jump        | When to use                              |
| ----------------------------- | ------- | ----------------- | ----------- | ---------------------------------------- |
| **Database** (RDS)            | ~$100   | ms, SQL queries   | —           | Data you query constantly with SQL       |
| **Object store** (S3 Std)     | ~$25    | 50–100 ms, by key | ~4x cheaper | Files, exports, data accessed by ID      |
| **Infrequent access** (S3 IA) | ~$12    | 50–100 ms, by key | ~2x cheaper | Data accessed less than once a month     |
| **Archive** (Glacier)         | ~$1–4   | minutes–hours     | ~4x cheaper | Compliance, backups, "just in case" data |

The question for every dataset: **how fast do I need this, and how often?** Access speed and access frequency determine which tier the data belongs in.

## Storage Lifecycle: Let Policies Move the Data

Storage is the silent accumulator. 750 GB at $17/month in month 1 becomes 45 TB at $1,035/month by year 5 — just from data sitting in S3 Standard forever. The fix is lifecycle policies configured once and forgotten.

```
Day 0          Day 30         Day 90         Day 365        Day 2555 (7 yr)
│              │              │              │              │
S3 Standard ──→ S3 IA ───────→ Glacier ─────→ Deep Archive → Deleted
$25/TB/mo      $12/TB/mo      $4/TB/mo       $1/TB/mo       $0
```

A log file written today costs $25/TB/month. In year 3 it costs $1/TB/month in Deep Archive — 25x cheaper for data nobody's reading. **Set lifecycle policies from day one, not when the bill is already painful.**

```
                  1m   3m   6m    1yr     2yr  3yr      7yr
                   │    │    │      │       │    │       │
Tx/orders     [──────── HOT ───────][── WARM ───][─COLD─]×  (PCI)

App logs      [HOT][WRM][───────COLD───────]×

Audit logs    [─ HOT ──][── WARM ──][──────── COLD ─────]×  (SOX)

Analytics     [HOT][─────────WARM──────────]×

User profiles [─────────── active ───────────────────────]×  (on deletion)

HOT = Postgres / S3 Std   WRM/WARM = S3 IA   COLD = Glacier   × = deleted
```

## The Hot/Cold Split: Don't Pay for Cold Data in Postgres

RDS charges 4x more per TB than S3 Standard — and most production databases contain years of data that gets queried once a month by an analyst. Keeping cold data in Postgres means paying for fast, transactional storage to serve analytical queries that can tolerate seconds of latency.

The worked example makes the magnitude concrete. A platform with 5 years of accumulated data:

```
Scenario A — Everything in Postgres:
  db.r6g.4xlarge (most RAM wasted on cold data)  $1,600/mo
  46 TB RDS storage                               $5,290/mo
  Total:                                          $6,890/mo

Scenario B — Hot/cold split:
  db.r6g.large (only hot data, small)              $200/mo
  180 GB RDS (last 12 months metadata)              $21/mo
  9 TB S3 Standard (recent journeys)               $207/mo
  27 TB S3 IA (older journeys)                     $337/mo
  9 TB Glacier (oldest journeys)                    $36/mo
  Total:                                            $818/mo
```

**Same data. Same queryability for recent records. 8x cheaper.** The old data is still queryable — through Athena against S3 Parquet at $5/TB scanned, rather than Postgres at $100/TB/month in storage.

### The Parquet + Athena pattern

Archive Postgres data to S3 as Parquet: columnar, compressed (~3x smaller than CSV), and splittable by partition. Then route queries by age:

```
Recent data  ──→ Postgres ──→ milliseconds
Old data     ──→ Athena → S3 Parquet ──→ seconds to minutes, $0.005/GB scanned
```

For archival, Postgres time-partitioning makes this mechanical: export old partitions via `COPY TO` Parquet, verify row counts, detach and drop the partition. The live table shrinks; Athena covers the history.

## Compute: Match the Pricing Model to the Workload Pattern

Unlike storage, compute cost is visible — but the mistake is using the same pricing model everywhere. There are four models spanning 5x in price:

```
On-demand:    $1.00  ████████████████████████████████████████
Reserved 1yr: $0.65  ██████████████████████████
Reserved 3yr: $0.42  █████████████████
Spot:         $0.20  ████████
```

| Workload               | Pricing model           | Why                                         |
| ---------------------- | ----------------------- | ------------------------------------------- |
| Production database    | Reserved 1yr or 3yr     | Always on, can't be interrupted             |
| App servers (baseline) | Reserved / Savings Plan | Minimum capacity always needed              |
| App servers (burst)    | On-demand + auto-scale  | Scales up for peak, down after              |
| Batch / ETL workers    | Spot                    | Interruptible, retryable, 60–90% savings    |
| CI/CD runners          | Spot                    | Short-lived, retryable                      |
| Dev/staging            | On-demand, off at night | Only used during working hours              |
| Lambda                 | Per invocation          | No idle cost; natural for bursty/event work |

**The dev/staging trap:** a staging environment mirroring prod runs 730 hours/month. Shut it down nights and weekends (220 hours/month) and the bill drops by 70%.

### Lambda vs containers: know the crossover

Lambda charges per invocation; containers charge per hour. Lambda wins at low traffic; containers win at high, steady throughput.

```
Monthly requests: 10K    100K   1M     10M    100M
                  │      │      │      │      │
Lambda cost:      $0.01  $0.10  $3.50  $35    $350
ECS (1 task 24/7):$35    $35    $35    $35    $35+
                  │      │      │      │      │
                  ← Lambda wins ──────→ ECS wins
```

The crossover is around 5–10 million requests/month depending on duration and memory. For internal tools, webhooks, and cron jobs, Lambda is almost free. For high-traffic APIs serving millions of requests, containers are cheaper.

## Common Mistakes

**Running everything on-demand.** A db.r6g.xlarge on-demand costs ~$400/month; reserved 1-year is ~$260/month. The break-even on a 1-year reservation is ~8 months. Anything that's been running 3+ months and will keep running → reserve it.

**Never deleting anything.** "We might need it" compounded over years. 10 TB of unread logs at $25/TB/month = $3,000/year. Set a retention policy, automate deletion, and let Glacier hold the edge cases.

**Over-provisioning for peak.** An r6g.4xlarge running at 10% CPU is paying for 14 idle cores. Right-size from CloudWatch metrics, not fear. Auto-scale app servers; right-size databases.

**Ignoring data transfer costs.** Cross-AZ traffic costs ~$0.01/GB each way; cross-region ~$0.02/GB. At 100 TB/month inter-region that's $4,000/month you didn't budget for. Keep chatty services in the same AZ. Use VPC endpoints for S3/DynamoDB (free within region).

## Quick Cost Reference

```
Compute:
  ECS task (0.5 vCPU, 1 GB):      ~$18/month
  ECS task (1 vCPU, 2 GB):        ~$37/month
  Lambda (1M req, 200ms, 256MB):  ~$3.50/month

Database:
  RDS db.t3.medium (dev):         ~$50/month
  RDS db.r6g.large (prod):        ~$200/month
  RDS db.r6g.xlarge (med prod):   ~$400/month
  DynamoDB (1M reads/day, on-demand): ~$7.50/month

Cache:
  ElastiCache t3.medium:          ~$45/month
  ElastiCache r6g.large:          ~$150/month

Storage (per TB/month):
  RDS storage:                    ~$100
  S3 Standard:                    ~$25
  S3 Infrequent Access:           ~$12
  Glacier Deep Archive:           ~$1

Network:
  ALB:                            ~$20/month + traffic
  NAT Gateway:                    ~$32/month + $0.045/GB
  CloudFront (1 TB):              ~$85/month
```

---

## Key Mental Models

1. **Storage grows silently; compute runs idle.** These are the two sources of cloud bill creep — set lifecycle policies and shut down non-prod environments.
2. **The four-tier ladder spans 100x.** RDS at $100/TB → S3 Standard at $25 → S3 IA at $12 → Glacier at $1. Each step down costs access speed.
3. **Hot data in Postgres, cold data in S3 Parquet, ancient data in Glacier.** The 8x cost difference on a real 5-year dataset is not hypothetical.
4. **Athena makes cold data queryable without keeping it hot.** $5/TB scanned vs $100/TB/month in storage. Route analytical queries to Athena.
5. **Set lifecycle policies from day one.** Data that ages into Glacier is 25x cheaper than data sitting in S3 Standard.
6. **Match pricing to pattern.** Reserved for steady state. Spot for batch. On-demand for burst. Lambda for bursty/event-driven. Wrong model on right architecture still wastes money.
7. **Lambda wins below ~5M requests/month; containers win above it.** Check the crossover math — "serverless is cheaper" is only true at low traffic.
8. **Reserve anything that's been running 3+ months.** Break-even on a 1-year reservation is ~8 months. The savings compound.
9. **Data transfer is invisible until it isn't.** Cross-AZ at $0.01/GB, cross-region at $0.02/GB — at high throughput this becomes a major line item.
10. **Right-size from metrics, not assumptions.** CPU at 15% doesn't need a bigger instance. Review utilisation monthly.
