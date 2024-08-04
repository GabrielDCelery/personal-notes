### CloudTrail 

The job of CloudTrail is to log activities within an AWS account. Every activity or API call gets logged there as a `CloudTrail Event`. These includes actions by `users`, `roles` or `services`.

AWS by default stores `90 days` of history for no cost in the form of an `Event History`. If you want to customize this in any way you have to create a `Trail`.

There are three types of events that you can work with within CloudTrail, `Management`, `Data` and `Insight` events.

`Management` events are control plane actions like for example creating or deleting resources (e.g. EC2) or logging into the console, while `Data` events are operations are performed on a `resource`. Like uploading objects to `S3`, or invoking a `Lambda` function. By default AWS only logs Management events, because Data events can be much higher in volume therefore would be much more expensive.

#### How to configure CloudTrail

Within the service a `CloudTrail Trail` is the smallest unit of configuration. CloudTrail is a `regional` service which means it has to be set up within a region. When configuring it can be set to be a `One region` or `All regions` trail.

If it uses `One region` configuration then every service that it is trailing has to be in the same region. If it is an `All regions` trail then it will collect logs from all regions. This has the added benefit that `if AWS introduces new regions` then the trail with this setting will automatically pick up the new region. 

Global services like `IAM`, `STS` or `Cloudfront` send their data to `us-east-1`. These types of events are called `Global service events` and apart from having a CloudTrail being correctly set up in the `us-east-1` region they also have to have the feature to log these events enabled to make it work.

Cloudtrail logs can also have `log verification` enabled which means if someone tampered with the logs you will be able to determine that.

#### How do CloudTrail logs get stored?

When setting up CloudTrail you have to specify an `S3 bucket` where the logs get stored as `compressed JSON` files to minimize storage. The logs can also be encrypted, but for that you need to have a custom `KMS key`.

In addition to that CloudTrail can be configured to forward the collected logs to `CloudWatch Logs`.

#### CloudTrail pricing

In terms of pricing CloudTrail charges for the logs that get analyzed and stored. In terms of analysis `one trail in every region on every account for Management events come for free`, but Data events or additional trails cost money. In terms of storage the `S3` storage rates apply (so it is a smart move to configure the bucket accordingly).

#### Managing logs accross an organization

One of the newer features of CloudTrails is that if you set up the trail within the `management account` of the orgainization then you get the ability to track logs accross every account withing that organization. The way thay works is that when you are logged into the management account and set up a new trail you will be presented with a checkbox to enable the trail for all accounts withing the organization.

#### Additional considerations

CloudTrail is `not a real time service`, typically the logs appear with a `15 minute delay` and logs multiple times an hour.
