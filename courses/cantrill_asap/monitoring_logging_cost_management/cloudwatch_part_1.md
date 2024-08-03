### Cloudwatch Part 1

Cloudwatch is the service of responsible for the `ingestion`, `storage` and `management` of metrics.

It is a `public` service with public space endpoints so it can be accessed from within `AWS` and `on premises` environments.

This also means services living in AWS need to have either access to the `Internet Gateway` or an `Interface Endpoint` withing a VPC.

Several services have default integrations to Cloudwatch, but if you want to have access to detailed metrics then you will need to use the `Cloudwatch Agent` or the `Cloudwatch API`. Examples are:

- Install the agent on an EC2 instance you want to monitor
- Install the agent on an on premises computer
- Bake the agent directly into an application or use the API

Metrics that are collected can be used for monitoring or we can create actions that can either send a notification to `SNS`, `Event Bridge` or `Auto Scaling`.

Cloudwatch logs are made up of `data points` that are of `values` (like CPU usage), `timestamps` and optionally `dimensions` of the values.

These logs are stored in a `namespace` which you can think of as a container for logs. AWS by default organizes these based on services, so for example default logs go into namespaces like `AWS/Lambda` or `AWS/EC2`.

Each log also have a metric name like `CPUUtilization` to help identifying the nature of the metric.

If we wanted to have the ability to further gruop or filter our logs, we would attach extra dimensions to them (think of them as `tags`) which are simple name-value pairs (example -> Name:IntanceId,Value:i-111111111). Dependent on the name of the metric AWS might add dimensions of its own, so for example the `CPUUtilization` metric automatically gets attached the `ImageId`, `InstanceId`, `InstanceType`, `AutoscalingGroupName` dimensions., `AutoscalingGroupName` dimensions.

In summary every data point is made up of:

- a mandator `Namespace` (container that the log is stored in)
- a `MetricName` that tells us what the metric is about
- a `timestamp` and a `value`
- one or more Dimensions (tags)
