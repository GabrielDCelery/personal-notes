### Gateway endpoints

Gateway services are designed to be able to access `S3` or `DynamoDB` without entering AWS's public network zone. Or if you don't want to worry about routing your traffic through `IGW` or a `NAT Gateway`.

When creating a Gateway endpoint it `does not go into the VPC subnet`. Instead what happens is that a `prefix list` gets added to the `route table` with the endpoint as a target. Which means any traffic targetting `S3` or `DynamoDB` rather than using the `IGW` to reach those services will be routed through the `Gateway Endpoint`.

It is important to note that since `route tables are attached to the subbnets` even though the Gateway endpoint is declared at the VPC level you have to configure which `route tables` and implicitly which `subnets` will use the endpoint.

It is a `highly available` service across all regions by default.

The endpoint can also be configured with `IAM policies` to limit for example which `S3 buckets` it can access.

It is very important that the service is `regional` which can present problems if the `S3 bucket` has been set up in a different region than the VPC that is trying toaccess it (even though it is a "global" service).

An other neat trick that can be used with Gateway endpoints is setting a `bucket policy` on the bucket to only allow access from the Gateway endpoint to `Prevent leaky buckets`.


