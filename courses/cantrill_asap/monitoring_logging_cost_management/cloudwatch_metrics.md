### Cloudwatch metrics 

Cloudwatch is the service of responsible for the `ingestion`, `storage` and `management` of metrics.

It is a `public` service with public space endpoints so it can be accessed from within `AWS` and `on premises` environments.

This also means services living in AWS need to have either access to the `Internet Gateway` or an `Interface Endpoint` withing a VPC.

#### Which services can publish metrics?

Several services have default integrations to Cloudwatch, but usually they are limited in scope so if you want to have access to detailed metrics then you will need to use the `Cloudwatch Agent` or the `Cloudwatch API`. Examples are:

- Install the agent on an EC2 instance you want to monitor
- Install the agent on an on premises computer
- Bake the agent directly into an application or use the API

#### What makes up a Cloudwatch metric

Cloudwatch logs are made up of `data points` that are of combination of `values` (like CPU usage), `timestamps` and optionally `dimensions` of the values.

These logs are stored in a `namespace` which you can think of as a container for logs. AWS by default organizes these based on services, so for example default logs go into namespaces like `AWS/Lambda` or `AWS/EC2`.

Each log also have a metric name like `CPUUtilization` to help identifying the nature of the metric.

If we wanted to have the ability to further gruop or filter our logs, we would attach extra dimensions to them (think of them as `tags`) which are simple name-value pairs (example -> Name:IntanceId,Value:i-111111111). Dependent on the name of the metric AWS might add dimensions of its own, so for example the `CPUUtilization` metric automatically gets attached the `ImageId`, `InstanceId`, `InstanceType`, `AutoscalingGroupName` dimensions., `AutoscalingGroupName` dimensions.

In summary every data point is made up of:

- a mandator `Namespace` (container that the log is stored in)
- a `MetricName` that tells us what the metric is about
- a `timestamp` and a `value`
- one or more Dimensions (tags)

Metrics that are collected can be used for monitoring or we can create actions that can either send a notification to `SNS`, `Event Bridge` or `Auto Scaling`.

#### Metric resolution and retention

When it comes to storing metrics there are two important aspects to consider. One is `resolution` and the other one is `retention`.

Metrics that are produced by AWS use the default `Standard`, `60 seconds` granularity. If you need more detailed breakdown you can enable `High` resolution which is `1 second` granularity. Data that was retrieved at a certain level of granularity can be retrieved at a "lower" rate.

Data `retention` is for how long AWS stores metrics. That is based on granularity. The less granular the data is, the longer time period AWS is retaining the data for.

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
