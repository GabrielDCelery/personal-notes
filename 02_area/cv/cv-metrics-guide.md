# Metrics Guide

Key metrics to research and add to CV bullets for experience.

## By Task

| Bullet                                   | Metrics to find                                                                                                   |
| ---------------------------------------- | ----------------------------------------------------------------------------------------------------------------- |
| Claims processing system                 | Claims processed per day/month? Latency from ingestion to Databricks? Number of third-party providers integrated? |
| Billing calculator (pay-per-mile)        | Number of active policies it tracks? Transaction volume? How often it syncs (real-time = X per second)?           |
| New products / integrations / migrations | How many products launched? How many integrations? Any revenue impact or customer count?                          |
| White-labeling project                   | Number of partners using it? Time to onboard a new partner before vs after?                                       |
| Journey processing pipeline              | Volume of journeys processed (per day/month)? Cost reduction from the redesign? Storage savings from archival?    |
| AWS infrastructure / cloud costs         | Percentage or £ amount reduced? Number of services managed? Deployment frequency improvement?                     |
| Local dev tooling                        | Time saved per developer? Number of engineers using it? Setup time before vs after?                               |
| Internal tools                           | Hours saved per week/month? Number of users? Manual process time before vs automated time after?                  |

## What "Good" Looks Like

**Weak:**

> Maintained and improved our AWS infrastructure. Worked on cutting cloud costs.

**Strong:**

> Managed AWS infrastructure (15+ ECS services, 30+ Lambdas). Reduced monthly spend by 25% through rightsizing and reserved capacity.

**Weak:**

> Built the billing calculator and tracker for our pay-per-mile product

**Strong:**

> Built real-time billing system tracking 50k active policies, processing 2M journey events/month and syncing with Stripe

## If You Don't Have Exact Numbers

Estimates are fine - use qualifiers:

- "~50k policies"
- "reduced costs by roughly 30%"
- "serving 10+ internal users daily"

Ballpark numbers beat no numbers. Just don't invent them entirely - think back to dashboards, Datadog, CloudWatch, or Jira tickets that might have this data.

## Easiest Wins

If you only have time to find 3-4 metrics, prioritize:

1. **Cloud cost reduction** - everyone understands money saved
2. **Volume numbers** - claims/day, policies tracked, journeys processed
3. **Team impact** - engineers mentored, internal tool users
4. **Speed improvements** - deployment time, pipeline latency, dev setup time

These translate across companies and don't require domain knowledge to appreciate.
