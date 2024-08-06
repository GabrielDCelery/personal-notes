### Private and Public services

From a networking perspective AWS separates `Private` and `Public` services.

`Private zones` are called `Virtual Private Cloud (VPC)`. These are isolated so nothing from the outside can communicate with services living insidethem unless you configure it, and these VPCs also can not communicate with each other unless you allow it. You can think of them as your home network.

Outside of these private networks there is the public internet and also the AWS `Public zone` which is a network that is connected to the public internet, but is still considered AWS's own network. This is where AWS's publicly accessible services (like S3) reside.

The reason why this distinction is important is because if you have a service in your VPC and you give it access to a public AWS service via an `Internet` or `NAT` gateway then than traffic will NOT traverse the public internet but stay in the AWS network even though it is a "public" service.
