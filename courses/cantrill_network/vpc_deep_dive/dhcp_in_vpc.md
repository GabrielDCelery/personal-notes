### DHCP in VPC

DHCP allows for the automatic configuration of network resources. Each device that appears on a network only starts with a `MAC address` and starts an `L2 broadcast` to find a `DHCP server`. Once they find each other they start communicating using `layer 2` to acquire `layer 3` capabilities. Once they do the new device acquires an `IP address`, a `subnet mask` and a `default gateway`.
