# Does the Client Send SYN and ACK Before HTTP Starts?

## Question
So before the HTTP stream starts flowing the client needs to send the first SYN and the ACK?

## Answer

Yes, exactly! The TCP connection must be fully established **before** any HTTP data can flow.

### The Complete Sequence

```
Client                          Server
  |                               |
  |  1. SYN (seq=x)              |
  |----------------------------->|
  |                               |
  |  2. SYN-ACK (seq=y, ack=x+1) |
  |<-----------------------------|
  |                               |
  |  3. ACK (ack=y+1)            |
  |----------------------------->|
  |                               |
  | ✓ TCP Connection Established |
  |                               |
  |  4. HTTP Request (GET /)     |
  |----------------------------->|
  |                               |
  |  5. HTTP Response            |
  |<-----------------------------|
```

### What the Client Sends

The client sends **two** packets to establish the connection:
1. **SYN** - "I want to connect"
2. **ACK** - "I received your SYN-ACK, we're good to go"

Only after these three packets (SYN, SYN-ACK, ACK) have been exchanged does the TCP connection exist, and only then can the HTTP request start flowing.

### Why This Matters for Performance

This is why:
- **Latency matters** - that's 1.5 round trips before any application data flows
- **Connection reuse** is important (HTTP/1.1 keep-alive, HTTP/2 multiplexing)
- **TLS adds even more** - TLS handshake happens *after* TCP, adding more round trips
- Technologies like **TCP Fast Open** try to reduce this overhead by sending data in the initial SYN

So yes, the client must initiate with SYN and complete with ACK before HTTP can begin!
