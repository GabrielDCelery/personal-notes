### VPC Subnets

VPC subnets are sub-sections of a VPC. A subnet always lives within an AZ and is resilient withing that data center. One important thing to remember is that a subnet can never be in multiple AZs. Subnets withing a VPC can communicate with each other.

A subnet always gets allocated with an IPv4 CIDR range that is a subset of the VPCs CIDR range and can not overlap with an other subnet's CIDR range. Subnets can also get optionally allocated with an `IPv6 address` but only if IPv6 was enabled on the VPC. The CIDR range uses a `/64` mask which means we can have 256 subnets in a `/56` VPC CIDR range.

#### Reserved IP addresses

Every subnet has `5 reserved` IP addresses that can not be used by any of the instances withing that subnet. In an hypothetic subnet `10.16.16.0/20` the following addresses are unusable:

- `10.16.16.0` - network address - this is not an AWS specific thing, the network address is reserved on every network
- `10.16.16.1` - vpc router - every subnet has a `network interface` and this logical device is resopoinsible routing traffic in and out of the subnet
- `10.16.16.2` - dns resolver - this is a bit tricky because technically AWS uses the VPC+2 address for DNS resolution, but they reserve the address in every subnet regardless
- `10.16.16.3` - reserved - not used for anything, reserved for the future in case it is needed
- `10.16.31.255` - broadcast - even though broadcasting is not supported within a VPC, AWS does reserve this regardless

#### DHCP Option set


