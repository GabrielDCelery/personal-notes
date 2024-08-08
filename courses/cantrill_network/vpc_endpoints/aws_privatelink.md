### AWS PrivateLink

AWS PrivteLink is a `technology stack` that supports `VPC interface endpoints`. But it goes beyond that. It allows you to securely connect to resources that are hosted on other AWS accounts. These are called `endpoint services`. This also applies to external `marketplace services`.

When using interface endpoints we have a `VPC service provider` and a `VPC service consumer`. The former is the service that wants to share its resources with the other VPC. In order to do that the provider creates an `endpoint service`. This service gets created alongside a `network load balancer` which gets provisioned with `network load balancer nodes` in each AZ. It is important to note that the provider is responsible for correctly configuring the load balancer with `cross-zone load balancing` to ensure even traffic distribution and making sure the services themselves are `highly available` across AZs.

After setting up the endpoint access can be granted to `accounts`, `IAM users` and even `Roles`.

On the other hand the consumer sets up `interface endpoints` that can be used to access services that are shared by the `endpoint service`. These can also get `security groups` and `network access lists` applied to them. These `interface endpoints` are delivered as `network interfaces` that get their own IP addresses allocated and can be used to access the provider service as if it was living withing the consumer's VPC.

#### Some key considerations when using PrivateLink

When setting up an `endpoint service` it is worth setting it up in a `highly available` manner with multiple endpoints. Currently endpoints only support `IPv4 & TCP`, there is no IPv6 support.

Private DNS is also supported, which means you can configure your endpoint to have `private dns hostnames` that the consumer can call. Endpoints are also accessible via `Direct Connect`, `Site-to-site VPN` and `VPC peering` as well.
