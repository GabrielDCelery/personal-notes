---
title: "Networking round 1, question 1"
date: 2025-11-03
tags: ["networking"]
---

# Problem

Send an ICMP request to `18.153.44.182` with payload `OITM_2025`, you should receive two responses, the second one contains the answer to the challenge.

# Solution

1. Start capture in background, save to file

```sh
sudo tcpdump -i any icmp and host 18.153.44.182 -w /tmp/icmp_capture.pcap -v &
```

2. Wait a moment, then send packet

```sh
sudo nping --icmp --data-string "OITM_2025" -c 1 18.153.44.182
```

3. Stop capture

```sh
sudo pkill tcpdump
```

4. Read the capture with payload display

```sh
sudo tcpdump -r /tmp/icmp_capture.pcap -X
```

Which provided the following output:

```sh
reading from file /tmp/icmp_capture.pcap, link-type LINUX_SLL2 (Linux cooked v2), snapshot length 262144
Warning: interface names might be incorrect
11:13:05.768477 eth0  Out IP 172.24.21.132 > ec2-18-153-44-182.eu-central-1.compute.amazonaws.com: ICMP echo request, id 32419, seq 1, length 17
        0x0000:  4500 0025 16d6 0000 4001 6317 ac18 1584  E..%....@.c.....
        0x0010:  1299 2cb6 0800 1160 7ea3 0001 4f49 544d  ..,....`~...OITM
        0x0020:  5f32 3032 35                             _2025
11:13:05.806761 eth0  In  IP ec2-18-153-44-182.eu-central-1.compute.amazonaws.com > 172.24.21.132: ICMP echo reply, id 32419, seq 1, length 17
        0x0000:  4500 0025 4996 0000 7601 fa56 1299 2cb6  E..%I...v..V..,.
        0x0010:  ac18 1584 0000 1960 7ea3 0001 4f49 544d  .......`~...OITM
        0x0020:  5f32 3032 35                             _2025
11:13:05.806837 eth0  In  IP ec2-18-153-44-182.eu-central-1.compute.amazonaws.com > 172.24.21.132: ICMP echo reply, id 32419, seq 1, length 15
        0x0000:  4500 0023 4997 4000 7601 ba57 1299 2cb6  E..#I.@.v..W..,.
        0x0010:  ac18 1584 0000 2d5d 7ea3 0001 594f 555f  ......-]~...YOU_
        0x0020:  574f 4e                                  WON
```

5. Using nping with verbose output

> [!TIP]
> Use the verbose mode of nping to get the data immediately

```sh
sudo nping --icmp --data-string "OITM_2025" 18.153.44.182 -c 1 -v4
```
