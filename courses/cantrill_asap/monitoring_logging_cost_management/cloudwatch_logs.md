### Cloudwatch logs

When it comes to Cloudwatch logs there are two sides to the product. The `ingestion` side and the `subscription` side.

#### From what kind of sources can you ingest data?

Cloudwatch provides the capability to ingest logs from services that live in `AWS`, `on premises`, `IoT devices` or any `application` regardless whether they are running inside or outside of AWS.

In order to have the capability to consume `application` or `system logs` you have to install the `CloudWatch Agent` which can be configured to consume and forward `logging outputs`.

When it comes to the AWS integrartions Cloudwatch works pretty much with every service withing the ecosystem. Some examples include:

- VPC flow logs
- CloudTrail logs of who accessed what (account events and API calls)
- Elastic Beanstalk, ECS container logs, API Gateway execution logs, Lambda execution logs
- Route53 DNS request logs

#### How do logs get stored in Cloudwatch

In terms of locality Cloudwatch by default is a `regional` service, so if a service lives in `eu-west-2` then it sends its logging data to `eu-west-2`. There are certain exceptions, global services like Route 53 send their data to `us-east-1`.

Log events that are sent with their `timestamp` and `value` are grouped together into `log streams` that are a collection of logs from within the same source that occurred during a short period of time. If there are multiple sources there are multiple log streams. Those log streams then are collected to `log groups` (like a container) which generally represent the thing that is being monitored.

Log groups are retained indefinitely by default, but that can be modified for cost efficiency. `IAM policies` can also be applied to control who can have access to log groups and logs can be `encrypted at rest` in case there is sensitive data there that we want to protect.

#### How to convert logs into metrics

Log groups can get `Metric filters` attached to them which filter out logs that match a certain criteria. Those logs can be then converted to metrics. For example we can have a log that details an application crash and we can turn that into a metric that tracks the number of crashes within a time period. These metrics then can be used to trigger alarms.

#### How to extract logs from Cloudwatch

There are a few ways to get logs out of Cloudwatch.

The first method is to create an `S3` export task that will automatically export logs. This is not immediate, there is an `up to 12 hours` delay. An other caveat that this data cannot be custom encrypted, it has to use the `default AWS KMS`.

If we want to deliver our logs at `real time` to other systems then we have to create a `Subscription filter` on a `log group` and that specifies which logs get forwarded and where they are being sent.

If you need `near real time` delivery then `Kinesis Firehose` is a good choice because that can send data either every `60 seconds` or when it's `buffer gets full`. That data from the Firehose can be forwarded to `S3`. This is also a very `cost effective` solution.

If we need `truly real time` delivery then we can have a `Lambda` subscribe to the filter. That lambda then can forward the logs to services like `Elastic Search`, or even external providers like `DataDog`. 

For real time delivery we can also use `Kinesis data stream`, but that is more costly.

An other more common use of `Kinesis data stream` is having it used as a data aggregator that listens to `multiple subscription filters` even accross multiple environments (like accounts or regions). That data can be forwarded to other consumers, like for example `Kinesis firehose + S3` for data retention.
