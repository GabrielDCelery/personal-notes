# What Does HTTP/2 Do With Plain Text?

## Question
So if I were sending plain text what would HTTP/2 do with it, what gets encoded? Just the header but not the body?

## Answer

Exactly! With plain text (or any body content), HTTP/2 only compresses the **headers**, not the body.

### Plain Text Example

Let's say you're sending a simple HTTP request:

```
POST /api/users HTTP/2
Host: example.com
Content-Type: text/plain
Content-Length: 26

Hello, this is plain text!
```

### What HTTP/2 Does

**Headers (compressed with HPACK):**
```
:method: POST
:path: /api/users
:scheme: https
:authority: example.com
content-type: text/plain
content-length: 26
```
- These get compressed using HPACK
- Might reduce from ~100 bytes to ~20 bytes (depending on header reuse)

**Body (NOT compressed):**
```
Hello, this is plain text!
```
- Sent as-is in HTTP/2 DATA frames
- Just framed (split into chunks with length prefixes)
- **No compression applied**

### The Complete Flow

```
Client                          Server
  |                               |
  | HEADERS frame                 |
  | (HPACK compressed)            |
  | :method: POST                 |
  | :path: /api/users             |
  | content-type: text/plain      |
  |----------------------------->|
  |                               |
  | DATA frame                    |
  | (NOT compressed)              |
  | "Hello, this is plain text!"  |
  |----------------------------->|
```

### If You Want Body Compression

You have options at the **application layer**:

**Option 1: Content-Encoding header**
```
POST /api/users HTTP/2
Content-Type: text/plain
Content-Encoding: gzip
Content-Length: 15

[gzipped data]
```
- You manually gzip the body before sending
- Server sees `Content-Encoding: gzip` and decompresses

**Option 2: Accept-Encoding (for responses)**
```
GET /data HTTP/2
Accept-Encoding: gzip, deflate, br

Response:
Content-Encoding: br
[brotli compressed data]
```

### Why Doesn't HTTP/2 Compress Bodies?

1. **Security**: CRIME/BREACH attacks exploit compression + encryption
2. **Flexibility**: Applications can choose the right compression for their data
3. **Already compressed**: Images (JPEG, PNG), videos (MP4), already-compressed files don't benefit
4. **Performance**: Compressing everything wastes CPU on incompressible data

### Real-World Usage

**Headers benefit most from HPACK** because:
- Many headers repeat across requests (`:method: GET`, `user-agent:`, etc.)
- HPACK maintains a dictionary of seen headers
- Subsequent requests can reference headers by index

**Example:**
```
Request 1:
  :method: GET          [62 = literal]
  :path: /api/users     [literal]
  user-agent: Mozilla/5.0...  [literal]

Request 2:
  :method: GET          [reference index 62]
  :path: /api/posts     [literal]
  user-agent: ...       [reference previous]
```

### Summary

- **HTTP/2 HPACK**: Compresses headers only
- **HTTP/2 DATA frames**: Body sent as-is, just framed
- **Application layer**: Can use `Content-Encoding` if you want body compression
- **Plain text**: Would go through uncompressed in DATA frames

So yes, your plain text would be framed but not compressed by HTTP/2 itself!
