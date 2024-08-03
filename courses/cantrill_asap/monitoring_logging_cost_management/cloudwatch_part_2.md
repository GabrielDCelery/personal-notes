### Cloudwatch Part 2

When it comes to storing metrics there are two important aspects to consider. One is `resolution` and the other one is `retention`.

Metrics that are produced by AWS use the default `Standard`, `60 seconds` granularity. If you need more detailed breakdown you can enable `High` resolution which is `1 second` granularity. Data that was retrieved at a certain level of granularity can be retrieved at a "lower" rate.

Data retention is for how long AWS stores metrics. That is based on granularity. The less granular the data is, the longer time period AWS is retaining the data for.

Some numbers are:

- Sub 60 seconds retained for less than 3 hours
- 60 second retainer for 15 days
- 600 seconds are retained for 63 days
- 3600 seconds are retained for 455 days

One important thing to remember is that as data ages, it gets aggregated into less resolution to get it stored longer.

When it comes to analyzing logs one way to do that is via `Statisctics`. That means we take a series of data points (metrics) and aggregate them in a certain way (for example `minimum`, `maximum`, `average` etc...).

We can also filter data based on percentiles, for example we ask for the data points that are within p95 (95% percentile) which means we can remove 5% of the outliers.

If our intentions are to drive actions based of metrics then we have to use alarms. Alarms watch metrics over a peroid of time and can be either in an `ok` or an `alarm` state. Once they breach a threshold we can use them to drive actions, for example sendind `SNS` notifications or triggering `Autoscaling Groups`. 

The resolution of the metrics directly affects the resolution of the alarm. The resolution of the alarm follows the same logic of what we have when viewing a metric. If the metrics are `Standard` resolution then the alarms can be set to the multiplies of 60 seconds, if it is `High` resolution then we can have alarms that are set to 1, 5, 10, 30 or multiplies of 60 seconds.

When configuring alarms you can set:

- Period - the length of time used to evaluate the metric
- Evaluation perioids - the number of periods used to determine whether we should transition to alarm state or not
- Datapoints to alarm - the number of data points that need to breach the threshold withing the evaluation period to trigger the alarm
- Condition - the threshold itself that we need to breach for the alarm to trigger
