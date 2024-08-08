### Interface endpoints

An interface endpoint is very similar to a Gateway endpoint in the sense that its role is to give private access to public services. Historically `S3` and `DynamoDB` was only accessible through Gateway endpoints so for everything else (like API Gateway) you had to use an `Interface Endpoint`.

One important difference is that unlike the Gateway endpoint these are `not highly available`. They are added at the `subnet level` in the form of a `network interface`. So if the subnet or the AZ goes then the interface endpoint goes.

Unlike Gateway endpoints though since these are network interfaces within the VPC you can apply `security groups` to them. You still have the option to add `endpoint policies` to them to restrict what exactly they have access to. Same as Gateway endpoints these only support `IPv4` and `TCP`.

It uses `PrivateLink` which is a technology to inject `private services` or `AWS services` to inject to your VPC.

When setting up an interface endpoint it gets its own `IP address` and an endpoint specific `DNS name`. On top of that the endpoint also gets associated with a `Regional DNS` and a `Zonal DNS` to give quick access to any other services either at a `regional` or `availability zone` level.

There is also a feature called `Private DNS Override` which if enabled allows your services to communicate with AWS services `using their public endpoint`, but via the `interface endpoint` without any re-configuration. In that case what AWS does under the hood is they create a `hidden` managed `private hosted zone` that contains an override of the `default DNS name` of the public service to your interface endpoint.

This is important, because either you have to enable this to correctly reach the public service's endpoint (rather than still trying to go through the internet gateway), or you have to reconfigure your service to call the interface endpoint when trying to reach the public service (aka. change the DNS hostname you are trying to reach at the application level.

Always remember `interface endpoints use DNS to route the traffic`.

