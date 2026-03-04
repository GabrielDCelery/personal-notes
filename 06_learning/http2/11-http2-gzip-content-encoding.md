# HTTP/2 with Content-Encoding: gzip

## Question
So for example for normal HTTP/2 requests if the Content-Encoding is gzip does this mean the payload is compressed and split into chunks and sent over the same TCP connection?

## Answer

Yes, exactly! Let me walk through what happens:

### HTTP/2 with Content-Encoding: gzip

**The flow:**

```
Client side:
1. Take your payload (e.g., JSON text)
2. Compress it with gzip
3. Add header: Content-Encoding: gzip
4. Send via HTTP/2

What gets sent:
- HTTP/2 HEADERS frame (HPACK compressed):
  :method: POST
  :path: /api/data
  content-type: application/json
  content-encoding: gzip
  content-length: 150

- HTTP/2 DATA frames (gzipped payload, possibly split):
  Frame 1: [gzipped bytes 0-16384]
  Frame 2: [gzipped bytes 16384-remaining]
```

### Key Points

**1. Compression happens BEFORE framing:**
```
Original JSON: 10KB
     ↓
gzip compress: 2KB
     ↓
HTTP/2 splits into DATA frames: maybe 1 frame (2KB < 16KB default)
     ↓
Sent over TCP connection
```

**2. Chunking is HTTP/2's job:**
- HTTP/2 can split the gzipped data into multiple DATA frames
- Frame size is negotiable (default max: 16KB)
- This is different from HTTP/1.1 "chunked transfer encoding"

**3. Same TCP connection:**
- Yes, all frames go over the same TCP connection
- Multiple HTTP/2 streams can be multiplexed on this one connection
- Each stream has its own DATA frames

### Visual Example

```
Client                          Server
  |                               |
  | HEADERS frame (stream 1)      |
  | content-encoding: gzip        |
  |----------------------------->|
  |                               |
  | DATA frame (stream 1)         |
  | [gzipped payload part 1]      |
  |----------------------------->|
  |                               |
  | DATA frame (stream 1)         |
  | [gzipped payload part 2]      |
  |----------------------------->|
  |                               |
  | HEADERS frame (stream 3)      | ← Different request
  | content-type: text/plain      |   on same connection!
  |----------------------------->|
  |                               |
  | DATA frame (stream 1)         | ← Can interleave!
  | [gzipped payload part 3]      |
  |----------------------------->|
```

### HTTP/1.1 vs HTTP/2 Comparison

**HTTP/1.1 with gzip:**
```
POST /api HTTP/1.1
Content-Encoding: gzip
Transfer-Encoding: chunked

1A4\r\n
[gzipped chunk 1]\r\n
F3\r\n
[gzipped chunk 2]\r\n
0\r\n
\r\n
```
- Text-based protocol
- Chunked encoding is optional
- One request at a time per connection

**HTTP/2 with gzip:**
```
HEADERS frame: content-encoding: gzip
DATA frame 1: [gzipped bytes]
DATA frame 2: [gzipped bytes]
```
- Binary protocol
- Always framed (not optional)
- Multiple concurrent streams

### Server Side

```
Server receives:
1. HEADERS frame → sees "content-encoding: gzip"
2. DATA frames → reassembles into complete gzipped payload
3. gunzip the payload
4. Parse the decompressed data (e.g., JSON)
```

### Summary

- **Yes**: Payload is compressed with gzip first
- **Yes**: HTTP/2 splits it into DATA frames (chunks)
- **Yes**: All sent over the same TCP connection
- **Bonus**: Multiple different requests can share the same TCP connection simultaneously (multiplexing)

The compression happens at the application layer (you or your HTTP library does it), and the framing/chunking happens at the HTTP/2 protocol layer automatically!
