# Does gRPC Use HTTP?

## Question
Does gRPC use HTTP?

## Answer

Yes! gRPC uses **HTTP/2** as its transport protocol.

### gRPC Architecture

```
Application Layer:    gRPC (Protocol Buffers serialization)
                         |
Application Protocol: HTTP/2 (framing, multiplexing, flow control)
                         |
Security Layer:       TLS (encryption - optional but recommended)
                         |
Transport Layer:      TCP
```

### How gRPC Uses HTTP/2

**gRPC maps its concepts to HTTP/2:**
- **gRPC calls** → HTTP/2 streams
- **Request/response messages** → HTTP/2 frames
- **Metadata** → HTTP/2 headers
- **Multiple simultaneous calls** → HTTP/2 multiplexing over a single TCP connection

**HTTP/2 Headers in gRPC:**
```
:method = POST
:scheme = https
:path = /service.name/MethodName
:authority = example.com
content-type = application/grpc
grpc-encoding = gzip
grpc-timeout = 1S
```

### Why HTTP/2?

gRPC chose HTTP/2 because it provides:

1. **Multiplexing** - Multiple RPC calls over one connection without head-of-line blocking
2. **Streaming** - Bidirectional streaming (client stream, server stream, both)
3. **Flow control** - Built-in backpressure mechanisms
4. **Header compression** - Efficient metadata transmission
5. **Binary protocol** - Matches gRPC's binary nature (Protocol Buffers)
6. **Existing infrastructure** - Works with HTTP/2 proxies, load balancers, etc.

### gRPC Call Types Mapping to HTTP/2

**Unary (single request/response):**
- One HTTP/2 stream
- Client sends request frames
- Server responds with response frames

**Server streaming:**
- One HTTP/2 stream
- Client sends one request
- Server sends multiple response frames

**Client streaming:**
- One HTTP/2 stream
- Client sends multiple request frames
- Server sends one response

**Bidirectional streaming:**
- One HTTP/2 stream
- Both sides send frames concurrently

### gRPC vs REST

| Feature | gRPC | REST |
|---------|------|------|
| Protocol | HTTP/2 | HTTP/1.1 or HTTP/2 |
| Payload | Protocol Buffers (binary) | JSON (text) |
| Streaming | Native support | Limited/awkward |
| Browser support | Limited (needs grpc-web) | Full |

### Connection Establishment for gRPC

```
Client                          Server
  |  TCP Handshake               |
  |----------------------------->|
  |  TLS Handshake (usually)     |
  |----------------------------->|
  |  HTTP/2 Connection Preface   |
  |  "PRI * HTTP/2.0\r\n..."     |
  |----------------------------->|
  |  HTTP/2 SETTINGS frame       |
  |<---------------------------->|
  |  gRPC calls can start        |
```

So yes, gRPC is essentially "RPC over HTTP/2" with Protocol Buffers for serialization. It leverages all of HTTP/2's features while adding its own conventions on top!
