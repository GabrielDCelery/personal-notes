---
title: "TCP vs QUIC"
date: 2025-11.10
tags: ["tcp", "quic"]
---

# QUIC Protocol Overview

QUIC (Quick UDP Internet Connections) is a modern transport protocol developed by Google, now standardized by IETF. It's designed to replace TCP+TLS for web traffic.

Core Concept

QUIC runs on top of UDP but adds the reliability and features that UDP lacks, while avoiding TCP's limitations.

Key Features:

1. Built-in Encryption (TLS 1.3)

- Encryption is mandatory and integrated into the protocol
- Handshake combines connection establishment + encryption setup
- Faster than TCP+TLS which does them separately

2. Multiplexing Without Head-of-Line Blocking

- Multiple streams in one connection (like HTTP/2)
- But unlike TCP: if one stream loses a packet, other streams aren't blocked
- Each stream is independent

3. Connection Migration

- Connections identified by Connection ID (not IP+Port like TCP)
- Can survive network changes (WiFi → cellular, IP address changes)
- Your connection persists even if your IP changes

4. 0-RTT Resumption

- Can send data on the very first packet for resumed connections
- TCP requires at least 1 RTT for handshake, TLS adds more
- QUIC can be instant for returning clients

5. Improved Congestion Control

- Better packet loss recovery than TCP
- More precise timing information
- Faster adaptation to network conditions

Why UDP?

QUIC uses UDP as a "dumb pipe" because:

- TCP is baked into OS kernels - hard to update/innovate
- Middleboxes (firewalls, NAT) interfere with new TCP features
- UDP lets QUIC implement everything in userspace, easier to evolve

Used By:

- HTTP/3 (runs on QUIC instead of TCP)
- Google services
- Facebook/Meta
- Cloudflare
- ~25% of internet traffic

TL;DR: QUIC = UDP + reliability + encryption + modern features, designed to be faster and more robust than TCP+TLS.

# HTTP/3 Data Integrity over UDP

HTTP/3 ensures data integrity through QUIC, which implements its own reliability layer on top of UDP. Here's how:

1. Packet Numbers & Acknowledgments

- Every QUIC packet has a unique, monotonically increasing packet number
- Receiver sends ACK frames back to confirm receipt
- Unlike TCP sequence numbers, packet numbers never reuse (simplifies loss detection)

Sender: Packet #1, #2, #3, #4, #5
Receiver: ACK: received 1,2,4,5 (missing 3)
Sender: Retransmits packet #3 with NEW packet number #6

2. Retransmission on Loss

- If ACK isn't received within timeout, packet is retransmitted
- Uses sophisticated algorithms to detect loss quickly
- Retransmitted data gets a new packet number (not same as TCP)

3. Checksums

- Every QUIC packet has cryptographic integrity protection
- Built into the TLS 1.3 encryption layer
- Detects any corruption or tampering
- Stronger than UDP's optional checksum

4. Stream-Level Flow Control

- Each QUIC stream tracks:
  - What data was sent
  - What was acknowledged
  - What needs retransmission
- Guarantees data arrives in order per stream
- Different streams don't block each other

5. Connection-Level Ordering

HTTP/3 Request → QUIC Stream
┌─────────────────────────────┐
│ Stream 1: /index.html │ ← Independent reliability
│ Stream 2: /style.css │ ← Independent reliability
│ Stream 3: /script.js │ ← Independent reliability
└─────────────────────────────┘
↓
UDP packets (unreliable)
↓
QUIC adds: ACKs, retransmits, checksums

Key Difference from TCP:

TCP: One lost packet blocks everything (head-of-line blocking)
Lost packet in stream 1 → streams 2 & 3 must wait

QUIC/HTTP/3: Loss only affects that specific stream
Lost packet in stream 1 → streams 2 & 3 continue normally

In Practice:

UDP: "Fire and forget" ❌ No guarantees
QUIC: "Fire, track, verify, retransmit if needed" ✅ Reliable
HTTP/3: Uses QUIC's reliability, gets all TCP benefits without drawbacks

Bottom line: UDP is just the transport mechanism. QUIC implements all the reliability, ordering, and integrity checks in userspace, giving it the benefits of TCP while being more flexible and efficient.

# TCP Head-of-Line Blocking (the problem):

TCP has a single byte stream with sequence numbers. Even if you multiplex multiple HTTP requests over one TCP connection (like HTTP/2 does):

TCP connection (single ordered byte stream):
[Stream1: byte 0-100][Stream2: byte 101-200][Stream1: byte 201-300]

If bytes 50-60 are lost:
❌ Bytes 101-300 are already received by kernel BUT
❌ TCP can't deliver them to application until 50-60 arrive
❌ Everything blocks waiting for retransmit

Why? TCP guarantees in-order delivery of the entire byte stream. Even though Stream 2's data (bytes 101-200) arrived fine, TCP won't deliver it until the earlier missing bytes arrive.

QUIC/HTTP/3 Solution (no blocking):

QUIC has independent streams, each with their own ordering:

Stream 1: packets 1, 2, 3, 4
Stream 2: packets 1, 2, 3, 4
Stream 3: packets 1, 2, 3, 4

If Stream 1, packet 2 is lost:
✅ Stream 1 waits for retransmit (blocked)
✅ Stream 2 continues delivering data (NOT blocked)
✅ Stream 3 continues delivering data (NOT blocked)

Why? Each QUIC stream tracks its own sequence independently. Loss in one stream doesn't affect others.

Visual Comparison:

HTTP/2 over TCP:
Time →
Request A: [▓▓▓▓▓] ← Complete, but waiting...
Request B: [▓▓▓▓▓] ← Complete, but waiting...
Request C: [▓▓X▓▓] ← Missing packet X blocks A & B!
↑
All blocked until X arrives

HTTP/3 over QUIC:
Time →
Request A: [▓▓▓▓▓] ✅ Delivered immediately
Request B: [▓▓▓▓▓] ✅ Delivered immediately
Request C: [▓▓X▓▓] ❌ Waits for X, but doesn't block A & B

Thanks for catching my confusing explanation! QUIC's whole point is to avoid the blocking problem that TCP has.

# HTTP/2 over TCP:

Client requests:
GET /index.html
GET /style.css
GET /script.js
GET /image.png

All to: example.com:443

These all go over:

- ✅ Multiple HTTP/2 streams (logical multiplexing)
- ❌ Single TCP connection (one physical byte stream)

TCP Connection (single ordered byte stream):
┌────────────────────────────────────────────────────┐
│ [stream1][stream2][stream3][stream1][stream2]... │
│ All interleaved in ONE byte stream │
└────────────────────────────────────────────────────┘
↑
If ANY byte is lost, everything after it blocks

The Problem:

HTTP/2 multiplexing is at the application layer, but it still runs over TCP which only knows about bytes, not streams.

What HTTP/2 sees: 4 independent streams
What TCP sees: One sequence of bytes (byte 0 → byte 50,000)

TCP's job: Deliver bytes IN ORDER
Result: One lost packet blocks ALL streams

Example Timeline:

TCP byte stream:
Bytes 0-100: /index.html data
Bytes 101-200: /style.css data
Bytes 201-300: /script.js data ← LOST PACKET
Bytes 301-400: /image.png data ← Already received!

TCP behavior:

- Bytes 0-100: ✅ Delivered to HTTP/2
- Bytes 101-200: ✅ Delivered to HTTP/2
- Bytes 201-300: ❌ Lost, waiting for retransmit
- Bytes 301-400: ❌ Received BUT blocked in TCP buffer
  Can't deliver until 201-300 arrives

Even though /image.png data arrived fine, TCP won't hand it to HTTP/2 until the /script.js packet is retransmitted.

HTTP/3 over QUIC fixes this:

QUIC Connection:
┌─────────────────┐
│ Stream 1 (independent ordering)
│ Stream 2 (independent ordering)
│ Stream 3 (independent ordering)
│ Stream 4 (independent ordering)
└─────────────────┘
↓
All sent via UDP packets

Loss in Stream 3 only blocks Stream 3. Streams 1, 2, 4 continue normally.

---

TL;DR: HTTP/2 has logical streams, but TCP underneath only sees one ordered byte stream. HTTP/3 + QUIC has truly independent streams at the transport layer.

# Packet retransmisson in QUIC when the retransmitted packet has a different id

Two Separate Numbering Systems

QUIC uses two different concepts:

1. Packet Numbers - for transport/loss detection (never reused)
2. Stream Offsets - for ordering data within a stream

How It Works:

Packet Number: Identifies the packet itself (for ACKs/loss detection)
Stream Offset: Identifies where data belongs in the stream (for ordering)

Packet #1: Stream 4, offset 0-100, data: "Hello"
Packet #2: Stream 4, offset 101-200, data: "World"
Packet #3: Stream 4, offset 201-300, data: "!" ← LOST
Packet #4: Stream 5, offset 0-50, data: "Foo"

Retransmit:
Packet #5: Stream 4, offset 201-300, data: "!" ← NEW packet number
↑ ↑
Different packet SAME stream offset

Receiver's Perspective:

Received packets by packet number:
Packet #1: Stream 4, bytes at offset 0-100
Packet #2: Stream 4, bytes at offset 101-200
Packet #4: Stream 5, bytes at offset 0-50
Packet #5: Stream 4, bytes at offset 201-300 ← Fills the gap!

Reconstructed Stream 4:
Offset 0-100: "Hello" (from packet #1)
Offset 101-200: "World" (from packet #2)
Offset 201-300: "!" (from packet #5)

Result: "HelloWorld!" in correct order

Why This Design?

TCP's problem:
TCP Sequence #100: "Data" ← Lost
TCP Sequence #100: "Data" ← Retransmit (SAME sequence number)

Ambiguity: Is the ACK for original or retransmit?
Hard to measure RTT accurately

QUIC's solution:
Packet #3, Stream offset 200: "Data" ← Lost
Packet #7, Stream offset 200: "Data" ← Retransmit (NEW packet number)

Clear: ACK for packet #7 means retransmit was received
Can measure RTT precisely
Stream offset 200 ensures correct ordering

In Practice:

Each QUIC STREAM frame contains:

- Stream ID: Which stream this belongs to
- Offset: Where in that stream's byte sequence
- Length: How many bytes
- Data: The actual payload

Frame in Packet #5:
{
type: STREAM,
stream_id: 4,
offset: 201, ← This determines order!
length: 100,
data: "..."
}

The receiver assembles each stream independently using the offsets, regardless of packet numbers.

TL;DR: Packet numbers track packets for ACKs. Stream offsets track data position for ordering. Separate concerns = clearer protocol design.

# TCP and HTTP Relationship

HTTP runs ON TOP of TCP - they're different layers of the network stack.

The Layer Stack:

┌─────────────────────────────────────┐
│ Application Layer: HTTP │ "GET /index.html HTTP/1.1"
├─────────────────────────────────────┤
│ Transport Layer: TCP │ Handles reliable delivery
├─────────────────────────────────────┤
│ Network Layer: IP │ Handles routing between hosts
├─────────────────────────────────────┤
│ Link Layer: Ethernet/WiFi │ Physical transmission
└─────────────────────────────────────┘

HTTP is the application protocol (what you're saying)
TCP is the transport protocol (how it gets delivered reliably)

Packet Structure:

When you send an HTTP request, it gets wrapped in multiple layers:

┌────────────────────────────────────────────────────────────┐
│ ETHERNET FRAME │
│ ┌──────────────────────────────────────────────────────┐ │
│ │ IP PACKET │ │
│ │ ┌────────────────────────────────────────────────┐ │ │
│ │ │ TCP SEGMENT │ │ │
│ │ │ ┌──────────────────────────────────────────┐ │ │ │
│ │ │ │ HTTP MESSAGE │ │ │ │
│ │ │ │ │ │ │ │
│ │ │ │ GET /index.html HTTP/1.1 │ │ │ │
│ │ │ │ Host: example.com │ │ │ │
│ │ │ │ User-Agent: curl/7.0 │ │ │ │
│ │ │ │ │ │ │ │
│ │ │ └──────────────────────────────────────────┘ │ │ │
│ │ └────────────────────────────────────────────────┘ │ │
│ └──────────────────────────────────────────────────────┘ │
└────────────────────────────────────────────────────────────┘

Detailed Structure:

1. Ethernet Frame (Link Layer)

┌─────────────┬─────────────┬──────┬─────────┬─────┐
│ Destination │ Source │ Type │ Payload │ CRC │
│ MAC (6B) │ MAC (6B) │ (2B) │ │(4B) │
└─────────────┴─────────────┴──────┴─────────┴─────┘

2. IP Packet (Network Layer)

┌─────────┬────────┬─────┬────────┬─────────────┬─────────────┬─────────┐
│ Version │ Header │ ... │Protocol│ Source │ Destination │ Payload │
│ (4b) │Len(4b) │ │ (1B) │ IP (4B) │ IP (4B) │ │
└─────────┴────────┴─────┴────────┴─────────────┴─────────────┴─────────┘
TCP=6
UDP=17

3. TCP Segment (Transport Layer)

┌──────────┬─────────────┬─────────────┬───────────┬─────────────┐
│ Source │Destination │ Sequence │ ACK │ Flags │
│ Port(2B) │ Port (2B) │ Number(4B) │Number(4B) │ (SYN/ACK/.) │
├──────────┴─────────────┴─────────────┴───────────┴─────────────┤
│ Window Size, Checksum, Options... │
├─────────────────────────────────────────────────────────────────┤
│ DATA (HTTP) │
└─────────────────────────────────────────────────────────────────┘

4. HTTP Message (Application Layer)

GET /index.html HTTP/1.1\r\n
Host: example.com\r\n
User-Agent: Mozilla/5.0\r\n
Accept: text/html\r\n
\r\n
[optional body]

Real Example:

You type in browser: http://example.com/page.html

What actually happens:

1. DNS lookup: example.com → 93.184.216.34

2. TCP Handshake (3 packets):
   [SYN] Your IP:54321 → 93.184.216.34:80
   [SYN-ACK] 93.184.216.34:80 → Your IP:54321
   [ACK] Your IP:54321 → 93.184.216.34:80

3. HTTP Request wrapped in TCP:
   ┌─────────────────────────────────┐
   │ TCP Header: │
   │ Source Port: 54321 │
   │ Dest Port: 80 │
   │ Seq: 1000 │
   │ ─────────────────────────────── │
   │ HTTP Data: │
   │ GET /page.html HTTP/1.1 │
   │ Host: example.com │
   └─────────────────────────────────┘

4. Server sends HTTP Response in TCP segments:
   ┌─────────────────────────────────┐
   │ TCP Header: │
   │ Source Port: 80 │
   │ Dest Port: 54321 │
   │ Seq: 5000 │
   │ ─────────────────────────────── │
   │ HTTP Data: │
   │ HTTP/1.1 200 OK │
   │ Content-Type: text/html │
   │ <html>...</html> │
   └─────────────────────────────────┘

Key Points:

- HTTP = the message format (requests/responses)
- TCP = the delivery mechanism (reliable, ordered)
- HTTP doesn't care about packet loss/ordering - TCP handles that
- One HTTP request might be split across multiple TCP segments
- TCP adds ~20 bytes of header overhead per segment
- IP adds another ~20 bytes
- Ethernet adds ~18 bytes

Analogy:

- HTTP = the letter you write
- TCP = the postal service ensuring it arrives
- IP = the routing/addressing system
- Ethernet = the delivery truck

# TCP Connection Handling: Software & Hardware

TCP is handled at multiple layers by different components. Here's the breakdown:

1. Operating System Kernel (Primary Handler)

The OS kernel does the heavy lifting:

┌─────────────────────────────────────────────────────┐
│ YOUR APPLICATION │
│ (Browser, curl, your Go program) │
└──────────────────┬──────────────────────────────────┘
│ System calls: socket(), connect(),
│ send(), recv(), close()
↓
┌─────────────────────────────────────────────────────┐
│ OPERATING SYSTEM KERNEL │
│ ┌───────────────────────────────────────────────┐ │
│ │ TCP/IP Stack │ │
│ │ - TCP state machine (SYN, ACK, FIN, etc) │ │
│ │ - Sequence number tracking │ │
│ │ - Retransmission timers │ │
│ │ - Congestion control │ │
│ │ - Checksum validation │ │
│ │ - Socket buffers (send/receive queues) │ │
│ │ - Connection table │ │
│ └───────────────────────────────────────────────┘ │
└──────────────────┬──────────────────────────────────┘
│ Raw packets
↓
┌─────────────────────────────────────────────────────┐
│ NETWORK INTERFACE CARD (NIC) │
│ - Ethernet framing │
│ - Physical layer transmission │
└─────────────────────────────────────────────────────┘

2. Specific Components by OS:

Linux:
/net/ipv4/tcp.c - Core TCP implementation
/net/ipv4/tcp_input.c - Incoming packet processing
/net/ipv4/tcp_output.c - Outgoing packet processing
/net/ipv4/tcp_timer.c - Retransmission timers

Windows:
tcpip.sys - TCP/IP driver
afd.sys - Ancillary Function Driver (sockets)

macOS:
/System/Library/Extensions/Network.kext - Network stack

3. What Happens at Each Layer:

Your Application (Userspace)

// Your Go code
conn, err := net.Dial("tcp", "example.com:80")
conn.Write([]byte("GET / HTTP/1.1\r\n..."))

What actually happens:

1. Application calls connect() system call
2. Switches to kernel mode
3. Kernel takes over...

Kernel (TCP Stack)

Application → Kernel handoff
↓
┌──────────────────────────────────────┐
│ Kernel TCP Stack: │
│ │
│ 1. Create socket structure │
│ 2. Assign local port (if not set) │
│ 3. Look up destination IP (DNS) │
│ 4. Send SYN packet: │
│ - Generate sequence number │
│ - Set TCP flags (SYN=1) │
│ - Calculate checksum │
│ 5. Start retransmission timer │
│ 6. Wait for SYN-ACK │
│ 7. Send ACK │
│ 8. Mark connection ESTABLISHED │
│ 9. Return control to application │
└──────────────────────────────────────┘
↓
Pass packet to Network Interface Card

Network Card (Hardware)

┌────────────────────────────────────┐
│ NIC (Network Interface Card) │
│ │
│ - Add Ethernet header │
│ - Convert to electrical signals │
│ - Put on physical wire/WiFi │
└────────────────────────────────────┘

4. Real Example - Go Program:

// Your code
conn, \_ := net.Dial("tcp", "example.com:80")

Behind the scenes:

1. Go runtime calls: syscall.Socket()
   → Kernel creates socket structure in memory

2. Go runtime calls: syscall.Connect()
   → Kernel's TCP stack:
   a. Allocates source port (e.g., 54321)
   b. Creates TCP control block (TCB)
   c. Sends SYN packet
   d. Waits for SYN-ACK (blocking your goroutine)
   e. Sends ACK
   f. Updates TCB state to ESTABLISHED

3. Returns to your Go program
   → You can now write/read data

4. Who Does What:

| Task                     | Handler                   |
| ------------------------ | ------------------------- |
| TCP handshake (SYN/ACK)  | OS Kernel                 |
| Sequence number tracking | OS Kernel                 |
| Retransmissions          | OS Kernel                 |
| Checksums                | OS Kernel (sometimes NIC) |
| Congestion control       | OS Kernel                 |
| Socket API               | OS Kernel                 |
| Physical transmission    | NIC Hardware              |
| Application logic        | Your Program              |

6. View Active TCP Connections:

Linux/Mac:

# See kernel's TCP connection table

netstat -an | grep ESTABLISHED
ss -tan # Newer tool

# See TCP stats

cat /proc/net/tcp

Your Go program just sees:
conn.Write(data) // "Send this please"

Kernel actually does:

1. Buffer data in socket send queue
2. Break into MSS-sized segments
3. Add TCP headers (seq, ack, flags)
4. Add IP headers
5. Calculate checksums
6. Queue to NIC
7. Wait for ACK from receiver
8. Retransmit if timeout
9. Update congestion window
10. Return to your program when buffered

11. Hardware Offloading (Modern NICs):

Some modern NICs can offload TCP work:

┌────────────────────────────────────┐
│ TCP Offload Engine (TOE) NIC │
│ - Checksum calculation │
│ - TCP segmentation (TSO) │
│ - Large receive offload (LRO) │
└────────────────────────────────────┘

But the core TCP logic (state machine, retransmits, congestion control) is still in the kernel.

Summary:

Software:

- OS Kernel = 95% of TCP work
- Your application = Just uses socket API
- Network drivers = Bridge kernel ↔ hardware

Hardware:

- NIC = Physical transmission + optional offloading
- CPU = Runs kernel code that does TCP

In your Go UDP sender example, you're bypassing most of this because UDP has no connection state - the kernel just wraps your data in a UDP header and sends it. No handshake, no retransmits, no state tracking.

# Where QUIC reordering is handled

● Yes! QUIC handles packet reordering at the application layer (userspace), not in the kernel like TCP.

The Key Difference:

TCP (Kernel Space)

┌─────────────────────────────┐
│ Your Application │ ← Receives ordered data
├─────────────────────────────┤
│ Kernel (TCP Stack) │ ← Handles reordering here
│ - Reorder packets │
│ - Retransmit lost ones │
│ - Manage state │
├─────────────────────────────┤
│ NIC │
└─────────────────────────────┘

QUIC (User Space)

┌─────────────────────────────┐
│ Your Application │
├─────────────────────────────┤
│ QUIC Library (userspace) │ ← Handles reordering HERE
│ - Reorder packets │ (in your process)
│ - Retransmit lost ones │
│ - Manage streams │
│ - ACKs, congestion ctrl │
├─────────────────────────────┤
│ Kernel (UDP only) │ ← Just passes raw packets
│ - No reordering │
│ - No state │
├─────────────────────────────┤
│ NIC │
└─────────────────────────────┘

Why This Matters:

1. Easier to Update
   TCP: Need OS kernel update → slow, risky
   QUIC: Just update library → fast, safe

2. Visible to Your Application
   // With TCP - invisible, kernel handles it
   conn, \_ := net.Dial("tcp", "example.com:443")
   conn.Read(buf) // Already ordered by kernel

// With QUIC - library handles it in your process
session, _ := quic.Dial("example.com:443")
stream, _ := session.OpenStream()
stream.Read(buf) // Ordered by QUIC library (userspace)

3. Per-Stream Independence

QUIC can reorder packets independently per stream because it's in application code:

QUIC Library in your process:
┌──────────────────────────────┐
│ Stream 1 reorder buffer: │
│ [pkt1][pkt2][\_\_\_][pkt4] │ ← Missing pkt3, stream blocked
├──────────────────────────────┤
│ Stream 2 reorder buffer: │
│ [pkt1][pkt2][pkt3][pkt4] │ ← Complete, deliver to app!
├──────────────────────────────┤
│ Stream 3 reorder buffer: │
│ [pkt1][pkt2][pkt3] │ ← Complete, deliver to app!
└──────────────────────────────┘

Kernel UDP socket just sees:
Packets coming in any order → pass to QUIC library

What Kernel Does with UDP:

Kernel UDP handling:

1. Receive packet from NIC
2. Check UDP checksum
3. Look up destination port
4. Put in socket receive queue
5. Wake up application
6. Done ✅ (no reordering!)

The QUIC library then:
QUIC library (in your app):

1. Read UDP packet from kernel
2. Decrypt packet
3. Parse QUIC headers (packet #, stream ID, offset)
4. Put data in correct stream's reorder buffer
5. Check if stream has contiguous data
6. If yes: deliver to application
7. If no: wait for missing packets
8. Send ACK frames back

Example Flow:

Packets arrive out of order:
UDP socket receives: Pkt#4, Pkt#2, Pkt#1, Pkt#3

Kernel passes to QUIC library: "Here are 4 UDP datagrams"

QUIC library reorders based on packet numbers:
┌─────────────────────────────────┐
│ Receive Buffer: │
│ Pkt#1 (stream 5, offset 0) │
│ Pkt#2 (stream 5, offset 100) │
│ Pkt#3 (stream 7, offset 0) │
│ Pkt#4 (stream 5, offset 200) │
└─────────────────────────────────┘
↓
QUIC reassembles per stream:
Stream 5: [offset 0][offset 100][offset 200] ✅ Complete
Stream 7: [offset 0] ✅ Complete

Application reads Stream 5: Gets ordered data
Application reads Stream 7: Gets ordered data

Advantages:

1. No kernel bottleneck

- TCP reordering happens in kernel (shared resource)
- QUIC reordering happens in your process (parallel)

2. Custom logic possible
   // QUIC libraries can add features without OS changes
   quic.SetReorderingThreshold(custom_value)
   quic.EnableCustomCongestionControl()

3. Better debugging
   // Can inspect QUIC state from your app
   stats := stream.GetStats()
   fmt.Println("Packets reordered:", stats.ReorderedPackets)

Summary:

| Layer            | TCP        | QUIC                            |
| ---------------- | ---------- | ------------------------------- |
| Reordering       | Kernel     | Application (userspace library) |
| State management | Kernel     | Application (userspace library) |
| Retransmissions  | Kernel     | Application (userspace library) |
| What kernel does | Everything | Just UDP packet delivery        |

QUIC = TCP-like features implemented in userspace on top of dumb UDP pipes.

This is why QUIC can innovate faster - no need to update every OS kernel in the world!
