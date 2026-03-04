# How Do Two Computers Establish a TCP Connection?

## Question
How do two computers establish a TCP connection to begin with? Considering routers/servers have to find each other?

## Answer

Great question! Establishing a TCP connection involves multiple layers working together. Let me break it down:

### 1. DNS Resolution (if using hostnames)

Before connecting, you need an IP address:
- Your computer asks a DNS server "What's the IP address of example.com?"
- DNS server responds with the IP address (e.g., 93.184.216.34)

### 2. Routing (finding the path)

Your computer doesn't directly know how to reach the destination:
- Your computer knows its **default gateway** (usually your router)
- It sends packets to the gateway with the destination IP
- Each router along the path uses **routing tables** to determine the next hop
- Routers use protocols like BGP, OSPF to share routing information and build these tables
- The packet "hops" from router to router until it reaches the destination network

Think of it like mailing a letter - you give it to your local post office, they don't know the exact route, but each post office knows which direction to send it next.

### 3. The TCP Three-Way Handshake

Once packets can reach the destination, TCP establishes the connection:

```
Client                          Server
  |                               |
  |  SYN (seq=x)                 |
  |----------------------------->|
  |                               |
  |         SYN-ACK (seq=y, ack=x+1)
  |<-----------------------------|
  |                               |
  |  ACK (ack=y+1)               |
  |----------------------------->|
  |                               |
  | Connection Established        |
```

**Step 1 - SYN (Synchronize):**
- Client sends a packet with SYN flag set and an initial sequence number

**Step 2 - SYN-ACK:**
- Server responds with both SYN and ACK flags
- Acknowledges the client's sequence number
- Sends its own initial sequence number

**Step 3 - ACK:**
- Client acknowledges the server's sequence number
- Connection is now established

### 4. What Gets Exchanged

During the handshake, both sides agree on:
- Initial sequence numbers (for tracking bytes)
- Window sizes (for flow control)
- Maximum segment size
- Optional features (like TCP timestamps, selective acknowledgment)

### Key Points

- **IP addresses** identify the computers
- **Routing protocols** handle finding the path (your computer doesn't need to know the full route)
- **TCP ports** identify the specific application (e.g., port 80 for HTTP, 443 for HTTPS)
- The three-way handshake ensures both sides are ready to communicate before any data is sent

This is why there's a slight delay when you first connect to a website - DNS lookup, routing, and the TCP handshake all need to happen before any HTTP data flows!
