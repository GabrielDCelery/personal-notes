---
title: "Learn Wireshark with Chris Geer"
date: 2025-11-07
tags: ["networking", "wireshark"]
---

# Part 1 - Basic layout and settings

[Learn Wireshark! Tutorial for BEGINNERS](https://www.youtube.com/watch?v=OU-A2EmVrKQ&list=PLW8bTPfXNGdC5Co0VnBK1yVzAwSSphzpJ&index=1)

- Bottom right to create, switch profiles
- Preferences -> Layout to change the layout of the window
- Pane can have `Packet layout` view which is really cool
- Preference -> Columns to edit or add new columns

> [!TIP]
> Worth adding a delta column because it is immensely useful

- The time column can be configured using the View -> Time Display format

- View -> Coloring rules allow for editing or introducing new colors (e.g. TCP SYN, `tcp.flags.syn == 1`)

- Upper right corner allows for adding quick filters

> [!TIP]
> When editing the data you can add a column by right clicking on one of the values and apply as column OR filter (which speeds up setting up columns and filters)

# Part 2

[Wireshark for BEGINNERS // Capture Network Traffic](https://www.youtube.com/watch?v=nWvscuxqais&list=PLW8bTPfXNGdC5Co0VnBK1yVzAwSSphzpJ&index=2)

- Wireshark needs npcap and admin access to be able to access the machine for it wants to capture packets for
- On the main menu there is a little black circle (`Capture options`)
  - INPUT TAB
    - allows for selecting which interfaces we want to capture the packets for
    - there is a setting called `snaplen` which specifies how much data we want to capture for each capture (e.g. we are in a secure environment and only want to capture the first 64 bytes of each item, though you can undercapture)
    - enabling promiscous mode on interfaces allows for wireshark capturing traffic between machines that are unicasting to each other
  - OUTPUT TAB
  - you can specify the file you want to save the packet capture into
  - you can enable a setting to create a new file after for example exceeding a certain size of packet capture (e.g. 500 MegaByte)
  - there is also a ring buffer option to continuously override old captures if we don't want to just save data forever

# Part 3

[Intro to Wireshark Tutorial // Lesson 3 // Capturing Packets with Dumpcap](https://www.youtube.com/watch?v=DAtyzE1TUlI&list=PLW8bTPfXNGdC5Co0VnBK1yVzAwSSphzpJ&index=3)

How to capture packets using `tcpdump` or `dumpcap`

`dumpcap -D` - list interfaces
`dumpcap -i 1` - start capturing on specific interface
`dumpcap -b filesize:500000 -b files:10` - ring buffer storage

# Part 4

[Intro to Wireshark Tutorial // Lesson 4 // Where do we capture network traffic? How?](https://www.youtube.com/watch?v=Atde35_9AAc&list=PLW8bTPfXNGdC5Co0VnBK1yVzAwSSphzpJ&index=4)

- `TAP` - network tap - hardware device that allows for creating a physical copy of the data passing through the network connection
- `SPAN` - port mirroring - sends duplicate of data passing through port to an other monitoring port (span port)

[Everyone Should Have One of These - EASY Packet Capture!](https://www.youtube.com/watch?v=LPz3qKjUVvY) - pretty cool video on how to use a Netgear GS105E to mirror traffic

# Part 5

[How to Filter Traffic // Intro to Wireshark Tutorial // Lesson 5](https://www.youtube.com/watch?v=-HDpYR_QSFw&list=PLW8bTPfXNGdC5Co0VnBK1yVzAwSSphzpJ&index=5)

- capture filter - only capture specific type of traffic - has to be specified `BEFORE` starting the capture
- display filter - filter down the already captured traffic - has to be specified `AFTER` starting the capture

> [!TIP]
> Right click on something you want to filter on and use "apply as filter" or "prepare as filter"

`not (arp or ipv6 or ssdp)`
`tcp.port in {80,443,8080}`

# Part 6

[Wireshark Tutorial // Lesson 6 // Name Resolution](https://www.youtube.com/watch?v=gfxxCBCKvMU&list=PLW8bTPfXNGdC5Co0VnBK1yVzAwSSphzpJ&index=6)

- In the preferences settings one can go to `Name resolution` and set which values we want to resolve
- `Resolve network (IP) addresses` by default is unchecked but by setting it to true rather than just seeing IP addresses we can see actual DNS names

> [!IMPORTANT]
> When the name resolution feature is enabled the rest of the checkboxes below that setting determine how wireshark will attempt the resolution

- Right clicking on an IP adress offers the `Edit Resolved Name` option which allows for manually setting the name

- In the main menu `Statistics` -> `Resolved addressess` we can view the IP addresses that were resolved by wireshark

# Part 7

[Wireshark Tutorial - Lesson 7 // Using the Time Column](https://www.youtube.com/watch?v=SllJu5MdkAg&list=PLW8bTPfXNGdC5Co0VnBK1yVzAwSSphzpJ&index=7)

- The `View` -> `Time Display format` allows for changing the format of the time column
- Right clicking on a specific time allows for `Set/Unset Time Reference` which allows for using the selected time as a reference point
- `Edit` -> `Unset all time references` allows for resetting the view
- When looking at a specific packet there is a `Timestamps` section that shows relevant time information `in relation with the chain of packets that belong to the same stream`

> [!TIP]
> The timestamps can also be applied as columns which allows for a great view of delta times in related streams (this can also be sorted to look for large delays)

# Part 8

[Reading PCAPs with Wireshark Statistics // Lesson 8 // Wireshark Tutorial](https://www.youtube.com/watch?v=ZNS115MPsO0&list=PLW8bTPfXNGdC5Co0VnBK1yVzAwSSphzpJ&index=8)

- Looking at a packet capture packet by packet is hard, this is where statistics is useful (to give us a high level view of what is happenning)
- You can use the `Conversations` -> `Statistics` view to get a high level view
  - example: go to the TCP section and order the view by packet count to find IPs that are potentially scanning the network (low packet count)
  - example: sort traffic by transferred bytes
- There is a neat trick to right click on the information that catches my eye then apply that traffic as a filter if I just want to see that specific traffic
- The nice thing about filtering traffic at the IP level is that I can see all the types of traffic that was involved there (ICMP, TCP etc...)

# Part 9 - How to pull files out from a capture

[Extracting Files from PCAPs with Wireshark // Lesson 9 // Wireshark Tutorial](https://www.youtube.com/watch?v=Fn__yRYW6Wo&list=PLW8bTPfXNGdC5Co0VnBK1yVzAwSSphzpJ&index=9)

- Don't forget to pick a TPC packet, go to `Transmission Control Protocol` and right click on it, select `Protocol Preferences` -> `Allow Subdissector to reassemble TCP streamms`
- In the top menu there is `Files` -> `Export Objects` (and then for example choose http)
- This way we can extract images, executables, binaries (don't execute them!) from the TCP streams
- There is also a neat trick to actually right click on the packet in the logs and select `Follow` -> `TCP Stream`

# Part 10

[Map IP Address Locations with Wireshark (Using GeoIP)](https://www.youtube.com/watch?v=IlVppluWTHw&list=PLW8bTPfXNGdC5Co0VnBK1yVzAwSSphzpJ&index=10)

- We need to download the following databases [GeoLite Databases and Web Services](https://dev.maxmind.com/geoip/geolite2-free-geolocation-data/)
  - GeoLite2 Country
  - GeoLite2 City
  - ASN
- In order to import it to wireshark we can use `Preferences` -> `Name Resolution` -> `MaxMind database directories` which allows us to point a directory which stores the above mentioned `.mmdb` files
- Once enabled that will allow for looking at IP packets and not just seeing the IP address, but also the `GeoIP`
- The information can also be seen in `Statistics` -> `Endpoints` and on the bottom left there is also a `Map` feature to view the traffic on a map
