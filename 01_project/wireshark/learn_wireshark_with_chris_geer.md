---
title: "Learn Wireshark with Chris Geer"
date: 2025-11-07
tags: ["networking", "wireshark"]
---

# Part 1

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
