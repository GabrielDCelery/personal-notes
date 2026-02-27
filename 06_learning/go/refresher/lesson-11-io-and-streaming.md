# Lesson 11: I/O and Streaming

Go's I/O model is built around two interfaces — `io.Reader` and `io.Writer`. Everything streams through them: files, network connections, HTTP bodies, compression, encryption, JSON decoders, in-memory buffers. Understanding the contract of these interfaces, how to compose them, and where they silently go wrong separates developers who produce correct Go from those who produce subtly broken Go.

## The `io.Reader` and `io.Writer` Contracts

### `io.Reader`

```go
type Reader interface {
    Read(p []byte) error (n int, err error)
}
```

`Read` reads up to `len(p)` bytes into `p`. The contract:

- Returns the number of bytes read `n` and an error
- `n` may be **less than** `len(p)` even when more data is available — this is a partial read
- Returns `io.EOF` when no more data is available (may be returned alongside `n > 0`)
- Never returns `n > len(p)`
- A `nil` error does not mean all data was consumed

### The Partial Read Trap

```go
// ❌ Wrong — assumes Read fills the buffer
buf := make([]byte, 1024)
n, err := r.Read(buf)
data := buf[:n]   // might be 1 byte, might be 1024 bytes — unpredictable

// ✓ Correct — use io.ReadFull when you need exactly N bytes
n, err := io.ReadFull(r, buf)   // returns error if fewer than len(buf) bytes available

// ✓ Correct — use io.ReadAll when you want everything (but mind the size)
data, err := io.ReadAll(r)      // reads until EOF, returns all bytes

// ✓ Correct — use io.Copy to drain a reader to a writer
n, err := io.Copy(dst, src)     // handles partial reads internally in a loop
```

**Why `Read` returns partial data**: Underlying sources (network sockets, pipes) deliver data in chunks. The interface doesn't buffer — it returns whatever the OS gave it. Any code that assumes `Read` fills the buffer will work fine on files and bytes.Buffer, then silently break on network connections.

### `io.Writer`

```go
type Writer interface {
    Write(p []byte) (n int, err error)
}
```

Write writes `len(p)` bytes from `p`. The contract:

- Must return a non-nil error if it returns `n < len(p)` — partial writes always indicate an error
- If it returns `n == len(p)` and `err == nil`, all bytes were written

```go
// ✓ Standard library functions handle partial writes internally
fmt.Fprintf(w, "hello %s", name)   // wraps Write, handles errors

// ✓ Check errors on Write
n, err := w.Write(data)
if err != nil {
    return fmt.Errorf("write failed after %d bytes: %w", n, err)
}
```

### Key Interfaces Built On Reader/Writer

| Interface            | Adds                    | Example Implementations         |
| -------------------- | ----------------------- | ------------------------------- |
| `io.ReadWriter`      | Both `Read` and `Write` | `bytes.Buffer`, `net.Conn`      |
| `io.ReadCloser`      | `Read` + `Close`        | `http.Response.Body`, `os.File` |
| `io.WriteCloser`     | `Write` + `Close`       | `os.File`, `gzip.Writer`        |
| `io.ReadWriteCloser` | All three               | `net.Conn` (via `net.TCPConn`)  |
| `io.Seeker`          | `Seek(offset, whence)`  | `os.File`, `bytes.Reader`       |
| `io.ReaderAt`        | `ReadAt(p, off)`        | `os.File`, `bytes.Reader`       |

---

## The `io` Package Primitives

### `io.Copy`

The workhorse of I/O composition — copies from a `Reader` to a `Writer` until `EOF` or error:

```go
n, err := io.Copy(dst, src)       // returns bytes copied and first error
n, err := io.CopyN(dst, src, 512) // copy at most N bytes
n, err := io.CopyBuffer(dst, src, buf) // use a caller-supplied buffer (avoids allocation)
```

`io.Copy` handles partial reads internally. It calls `src.Read` in a loop until `io.EOF`.

**Optimization**: If `src` implements `io.WriterTo` or `dst` implements `io.ReaderFrom`, `io.Copy` delegates to those methods — allowing zero-copy transfers (e.g., `os.File` → `net.Conn` uses `sendfile` on Linux).

### `io.TeeReader`

Reads from a reader and simultaneously writes everything read to a writer — like a tee in a pipe:

```go
var buf bytes.Buffer
tee := io.TeeReader(r, &buf)

// Reading from tee reads from r AND writes to buf
data, err := io.ReadAll(tee)

// buf now contains everything that was read
fmt.Println(buf.String())   // same as string(data)
```

**Use case**: Log or capture a stream while also processing it.

```go
// ✓ Inspect HTTP response body without consuming it
func debugBody(resp *http.Response) {
    var buf bytes.Buffer
    tee := io.TeeReader(resp.Body, &buf)
    defer resp.Body.Close()

    var result MyStruct
    json.NewDecoder(tee).Decode(&result)

    log.Printf("raw response: %s", buf.String())
}
```

### `io.LimitReader`

Wraps a reader to return `io.EOF` after N bytes:

```go
limited := io.LimitReader(r, 1<<20)   // read at most 1 MB
data, err := io.ReadAll(limited)
```

**Critical for untrusted input**:

```go
// ❌ Unbounded — a malicious client sends infinite data, you run out of memory
body, err := io.ReadAll(resp.Body)

// ✓ Cap at a reasonable limit
body, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20)) // 10 MB max
```

### `io.MultiReader`

Concatenates multiple readers into one — they're read sequentially:

```go
r := io.MultiReader(
    strings.NewReader("header\n"),
    file,
    strings.NewReader("\nfooter\n"),
)
io.Copy(dst, r)   // reads header, then file contents, then footer
```

**Use case**: Prepend/append data to a stream without buffering it entirely.

### `io.MultiWriter`

Writes to multiple writers simultaneously — all must succeed:

```go
w := io.MultiWriter(os.Stdout, logFile, &buf)
fmt.Fprintf(w, "event: %s\n", event)   // writes to all three at once
```

**Use case**: Write to a log file, stdout, and a metrics buffer in one pass.

### `io.Pipe`

Creates a synchronous in-memory pipe — one goroutine writes, another reads:

```go
pr, pw := io.Pipe()

go func() {
    defer pw.Close()                      // ✓ always close the write end
    json.NewEncoder(pw).Encode(payload)   // blocks until reader consumes
}()

resp, err := http.Post(url, "application/json", pr)
```

**Why**: `http.Post` takes an `io.Reader`. If you have data that needs to be encoded on the fly, `io.Pipe` lets you stream the encoding directly into the HTTP request body without buffering the entire payload in memory first.

**Gotcha**: `io.Pipe` is synchronous — the writer blocks until the reader consumes the data. If neither goroutine makes progress, you deadlock. Always run the writer in a separate goroutine.

---

## `bufio`: When to Buffer

Raw `io.Reader`/`io.Writer` operations may result in many small syscalls. `bufio` wraps them with an in-memory buffer to amortize syscall overhead.

### `bufio.Reader`

```go
br := bufio.NewReader(r)          // default 4096-byte buffer
br := bufio.NewReaderSize(r, 64*1024)  // custom size

line, err := br.ReadString('\n')  // read until delimiter (includes delimiter)
line, isPrefix, err := br.ReadLine()  // read one line (without delimiter)
b, err := br.ReadByte()           // read single byte efficiently
br.UnreadByte()                   // push one byte back

// Peek without consuming
header, err := br.Peek(4)         // look at next 4 bytes without advancing
```

**When to use**: Wrap any reader where you'll be making many small reads (line-by-line parsing, protocol parsing). Don't wrap when you're doing one large `io.ReadAll`.

### `bufio.Writer`

```go
bw := bufio.NewWriter(w)
fmt.Fprintf(bw, "line %d\n", i)   // writes to internal buffer

// ❌ Forgetting to flush — last bytes never reach the underlying writer
// ✓ Always flush when done
if err := bw.Flush(); err != nil {
    return err
}
```

**When to use**: Wrap any writer where you'll make many small writes (record-by-record output, template rendering). The buffer is flushed automatically only when full — incomplete buffers are silently dropped if you don't call `Flush()`.

### `bufio.Scanner`

The idiomatic way to read line-by-line:

```go
scanner := bufio.NewScanner(r)
for scanner.Scan() {
    line := scanner.Text()   // current line, no newline
    process(line)
}
if err := scanner.Err(); err != nil {    // ✓ always check scanner errors
    return err
}
// Note: io.EOF is not returned by scanner.Err() — Scan() returning false covers it
```

**Default max token size**: 64KB per line. Larger lines cause an error.

```go
// ✓ Increase buffer for large lines
scanner := bufio.NewScanner(r)
scanner.Buffer(make([]byte, 1<<20), 1<<20)   // 1 MB max line
```

**Custom split functions**:

```go
scanner.Split(bufio.ScanWords)   // split on whitespace
scanner.Split(bufio.ScanBytes)   // one byte at a time
scanner.Split(bufio.ScanRunes)   // one UTF-8 rune at a time
scanner.Split(customSplitFunc)   // your own delimiter
```

### `bufio.ReadWriter`

```go
brw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
```

Useful for bidirectional buffering over a single connection (e.g., custom TCP protocols).

---

## In-Memory I/O

### `bytes.Buffer`

A `bytes.Buffer` implements both `io.Reader` and `io.Writer`. It grows automatically.

```go
var buf bytes.Buffer
fmt.Fprintf(&buf, "hello %s", name)   // write formatted string
buf.Write([]byte{0x01, 0x02})         // write raw bytes
buf.WriteByte(0xFF)                   // write single byte
buf.WriteString("suffix")             // write string

data := buf.Bytes()    // underlying slice (shared — don't modify after)
s    := buf.String()   // copy as string
n    := buf.Len()      // unread bytes remaining

// Read side
b, err := buf.ReadByte()
line, err := buf.ReadString('\n')
io.Copy(dst, &buf)     // drain buffer into writer
```

**Use case**: Build up output in memory, capture writes in tests, create readers from constructed data.

```go
// ✓ Use bytes.Buffer to create a reader from multiple pieces
var buf bytes.Buffer
buf.WriteString(header)
buf.Write(body)
http.Post(url, "application/octet-stream", &buf)
```

### `bytes.Reader`

A `bytes.Reader` wraps a `[]byte` as a read-only `io.Reader`. Unlike `bytes.Buffer`, it also implements `io.Seeker` and `io.ReaderAt`:

```go
r := bytes.NewReader(data)
r.Seek(0, io.SeekStart)   // rewind
r.ReadAt(buf, offset)     // read at specific offset without moving position
```

**Use it when**: You have a `[]byte` and need to pass it as an `io.Reader`. Prefer it over `bytes.Buffer` for read-only cases — it's cheaper.

### `strings.Reader`

Wraps a `string` as a read-only `io.Reader`:

```go
r := strings.NewReader("hello world")
io.Copy(os.Stdout, r)
```

**In tests**: Use `strings.NewReader` to provide controlled input to functions that take `io.Reader`.

### `strings.Builder`

Write-only builder optimised for string concatenation:

```go
var sb strings.Builder
fmt.Fprintf(&sb, "part %d", i)
sb.WriteString(", ")
sb.WriteByte('x')
result := sb.String()
```

Prefer `strings.Builder` over `bytes.Buffer` when you only need the final `string` — it avoids a copy on `String()`.

---

## Streaming JSON

### `json.Decoder` vs `json.Unmarshal`

```go
// json.Unmarshal — reads everything into memory first
data, err := io.ReadAll(r)
if err != nil { ... }
var result MyStruct
err = json.Unmarshal(data, &result)

// json.NewDecoder — streams directly from the reader
var result MyStruct
err := json.NewDecoder(r).Decode(&result)
```

|                            | `json.Unmarshal`          | `json.NewDecoder`                   |
| -------------------------- | ------------------------- | ----------------------------------- |
| Input                      | `[]byte`                  | `io.Reader`                         |
| Memory                     | Entire payload buffered   | Streams; only buffers what's needed |
| Multiple values            | Requires manual splitting | `Decode()` in a loop                |
| Performance (single value) | Slightly faster           | Negligible difference               |
| Preferred for              | Small known payloads      | HTTP bodies, files, pipes           |

### Decoding a Stream of JSON Values

```go
// NDJSON (newline-delimited JSON) — one object per line
dec := json.NewDecoder(r)
for {
    var event Event
    err := dec.Decode(&event)
    if err == io.EOF {
        break
    }
    if err != nil {
        return fmt.Errorf("decode error: %w", err)
    }
    process(event)
}
```

### Streaming JSON Encoding

```go
// ❌ Encodes entire slice into memory
data, err := json.Marshal(records)
w.Write(data)

// ✓ Streams each record directly to the writer
enc := json.NewEncoder(w)
for _, record := range records {
    if err := enc.Encode(record); err != nil {    // Encode appends \n automatically
        return err
    }
}
```

---

## HTTP Body Patterns

HTTP response bodies are `io.ReadCloser` — they must be read **and** closed correctly.

### Always Close the Body

```go
resp, err := http.Get(url)
if err != nil {
    return err
}
defer resp.Body.Close()   // ✓ always defer immediately after checking err
```

**Why**: The body is backed by an open TCP connection. Not closing it leaks the connection from the pool.

### Drain Before Closing for Connection Reuse

```go
// ✓ Drain body to allow connection reuse
defer func() {
    io.Copy(io.Discard, resp.Body)   // drain any remaining bytes
    resp.Body.Close()
}()
```

**Why**: HTTP/1.1 connection reuse requires the previous response body to be fully consumed. If you close without draining (e.g., after reading only the first few bytes), the connection is discarded rather than returned to the pool.

```go
// Common pattern: read JSON but discard on error
resp, err := http.Get(url)
if err != nil {
    return err
}
defer func() {
    io.Copy(io.Discard, resp.Body)
    resp.Body.Close()
}()

if resp.StatusCode != http.StatusOK {
    return fmt.Errorf("unexpected status: %d", resp.StatusCode)
}

var result MyStruct
return json.NewDecoder(resp.Body).Decode(&result)
```

### Limit Untrusted Response Bodies

```go
// ❌ Unbounded — a malicious server sends infinite data
data, _ := io.ReadAll(resp.Body)

// ✓ Cap at a reasonable size
data, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))   // 10 MB
```

### Request Bodies

```go
// ❌ Buffers entire payload
data, _ := json.Marshal(payload)
req, _ := http.NewRequest("POST", url, bytes.NewReader(data))

// ✓ Streams — no intermediate buffer
pr, pw := io.Pipe()
go func() {
    defer pw.Close()
    json.NewEncoder(pw).Encode(payload)
}()
req, _ := http.NewRequest("POST", url, pr)
```

---

## Writer Close and Flush Gotchas

Some writers buffer or finalise data on `Close()` or `Flush()`. Forgetting to call them silently truncates output.

### `bufio.Writer` — must `Flush()`

```go
bw := bufio.NewWriter(f)
bw.WriteString("data")
// ❌ file.Close() without Flush() — buffered bytes may not reach the file
f.Close()

// ✓
if err := bw.Flush(); err != nil { return err }
if err := f.Close(); err != nil { return err }
```

### `gzip.Writer` — must `Close()`

```go
gz := gzip.NewWriter(f)
gz.Write(data)
// ❌ Missing gz.Close() — gzip footer never written, file is corrupt
f.Close()

// ✓ Close in the right order: inner writer first, then outer
if err := gz.Close(); err != nil { return err }  // writes gzip footer
if err := f.Close(); err != nil { return err }   // flushes OS buffer
```

Same applies to `zlib.Writer`, `zip.Writer`, `tar.Writer`, `csv.Writer` (needs `Flush()`).

### Error on `Close()`

```go
// ❌ Defer close, ignore error — common but wrong for writers
defer f.Close()

// ✓ Check close error for writers — it may be the first indication of a write failure
defer func() {
    if cerr := f.Close(); cerr != nil && err == nil {
        err = cerr   // named return — propagate close error
    }
}()
```

Ignoring the error from `Close()` on a file writer can mask `fsync` failures — the OS accepted your writes into its buffer, but flushing to disk failed.

---

## Hands-On Exercise 1: Safe HTTP Response Processing

The following function has three I/O bugs. Identify and fix them.

```go
func fetchUser(url string) (*User, error) {
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }

    if resp.StatusCode != 200 {
        return nil, fmt.Errorf("status %d", resp.StatusCode)
    }

    data, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    var user User
    if err := json.Unmarshal(data, &user); err != nil {
        return nil, err
    }
    return &user, nil
}
```

<details>
<summary>Solution</summary>

**Bugs**:

1. ❌ `resp.Body` is never closed — leaks the TCP connection
2. ❌ When `StatusCode != 200`, the body is neither drained nor closed — connection not returned to pool
3. ❌ `io.ReadAll` with no limit — malicious/buggy server could exhaust memory

**Fixed**:

```go
func fetchUser(url string) (*User, error) {
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer func() {                              // ✓ always close body
        io.Copy(io.Discard, resp.Body)          // ✓ drain before closing
        resp.Body.Close()
    }()

    if resp.StatusCode != 200 {
        return nil, fmt.Errorf("status %d", resp.StatusCode)
        // defer handles drain+close even on early return
    }

    var user User
    // ✓ Stream directly with Decoder — no ReadAll needed
    // ✓ LimitReader caps untrusted input
    dec := json.NewDecoder(io.LimitReader(resp.Body, 1<<20))   // 1 MB
    if err := dec.Decode(&user); err != nil {
        return nil, fmt.Errorf("decode user: %w", err)
    }
    return &user, nil
}
```

</details>

## Hands-On Exercise 2: Composing Readers

Write a function `hashAndCopy(src io.Reader, dst io.Writer) (hash string, n int64, err error)` that:

1. Copies all bytes from `src` to `dst`
2. Simultaneously computes the SHA-256 hash of the data
3. Returns the hex-encoded hash, bytes copied, and any error

Do not buffer the entire input in memory.

<details>
<summary>Solution</summary>

```go
import (
    "crypto/sha256"
    "encoding/hex"
    "io"
)

func hashAndCopy(src io.Reader, dst io.Writer) (hash string, n int64, err error) {
    h := sha256.New()

    // TeeReader: every byte read from src is also written to h
    tee := io.TeeReader(src, h)

    // Copy from tee (which feeds h) to dst
    n, err = io.Copy(dst, tee)
    if err != nil {
        return "", n, err
    }

    hash = hex.EncodeToString(h.Sum(nil))
    return hash, n, nil
}
```

**Why it works**: `io.TeeReader(src, h)` returns a reader that, on each `Read`, reads from `src` and writes the same bytes to `h` (the SHA-256 hasher). `io.Copy` drives the reads in a loop. By the time `Copy` returns, `h` has seen all the bytes.

**Test**:

```go
func TestHashAndCopy(t *testing.T) {
    input := "hello world"
    src := strings.NewReader(input)
    var dst bytes.Buffer

    hash, n, err := hashAndCopy(src, &dst)

    if err != nil { t.Fatal(err) }
    if dst.String() != input { t.Errorf("dst = %q, want %q", dst.String(), input) }
    if int(n) != len(input) { t.Errorf("n = %d, want %d", n, len(input)) }

    // SHA-256 of "hello world"
    want := "b94d27b9934d3e08a52e52d7da7dabfac484efe04294e576f58e0b89f05ed3d"
    // (verify with: echo -n "hello world" | sha256sum)
    if hash != want { t.Errorf("hash = %s, want %s", hash, want) }
}
```

</details>

---

## Interview Questions

### Q1: What does `io.Reader.Read` guarantee, and what does it not guarantee?

Interviewers ask this to see if the candidate understands why code that works on files breaks on network connections. It's a litmus test for genuine systems experience.

<details>
<summary>Answer</summary>

**Guarantees**:

- `n` is always between 0 and `len(p)` inclusive
- If `n > 0` and `err != nil`, the `n` bytes in `p[:n]` are valid data — process them even on error
- `io.EOF` is returned when no more data is available; it may accompany `n > 0` on the final read
- A returned error is permanent — once an error is returned (other than `io.EOF`), subsequent calls also return errors

**Does not guarantee**:

- That `n == len(p)` — reads can be partial even when more data is available
- That `n > 0` — a zero-byte read with a `nil` error is valid (though unusual)
- Anything about timing — a blocking reader blocks indefinitely until data is available or an error occurs

**Why partial reads happen**: Network sockets, pipes, and OS I/O primitives return whatever data is currently available in the kernel buffer. If a TCP segment carries 512 bytes and you `Read` into a 4096-byte buffer, you get 512 bytes back — the rest hasn't arrived yet.

**The fix**: Use `io.ReadFull` when you need exactly N bytes, `io.ReadAll` when you want everything, and `io.Copy` when transferring between a reader and a writer. Never assume `Read` fills the buffer.

</details>

### Q2: When would you use `io.Pipe` and what deadlock risk does it carry?

A design question — tests whether the candidate knows when to stream vs buffer, and understands the synchronous nature of pipes.

<details>
<summary>Answer</summary>

**When to use `io.Pipe`**: When you need to connect a writer-based API to a reader-based API without buffering the entire payload in memory.

Classic example: encoding data and streaming it directly as an HTTP request body:

```go
pr, pw := io.Pipe()
go func() {
    defer pw.Close()
    json.NewEncoder(pw).Encode(largePayload)   // encoding streams to pipe
}()
http.Post(url, "application/json", pr)          // HTTP reads from pipe
```

Without `io.Pipe`, you'd encode to a `bytes.Buffer` first (full payload in memory), then pass that buffer as the body.

Other use cases: streaming compression, streaming encryption, chaining processing stages without intermediate allocations.

**The deadlock risk**: `io.Pipe` is fully synchronous — the writer blocks on `Write` until the reader calls `Read`, and vice versa. If both operations happen in the same goroutine, you deadlock immediately.

```go
// ❌ Deadlock — same goroutine writes and reads
pr, pw := io.Pipe()
pw.Write(data)          // blocks waiting for a reader
io.ReadAll(pr)          // never reached

// ✓ Writer in a separate goroutine
go func() {
    defer pw.Close()
    pw.Write(data)
}()
io.ReadAll(pr)          // reads what the goroutine writes
```

Also: if the reader returns an error (e.g., the HTTP request fails), `pw.Write` returns that error. Always check write errors in the goroutine and use `pw.CloseWithError(err)` to propagate failures to the reader.

</details>

### Q3: Why must you drain an HTTP response body before closing it, and what happens if you don't?

A practical question — directly tests production experience with HTTP clients. Getting this wrong causes connection pool exhaustion under load.

<details>
<summary>Answer</summary>

Go's `net/http` client maintains a pool of persistent TCP connections for connection reuse (HTTP keep-alive). When you finish reading a response, the client can return the underlying connection to the pool — but only if the response body has been fully consumed.

If you close the body without draining it:

- The client doesn't know whether there's remaining data on the wire
- It cannot safely reuse the connection (the next request would read stale data)
- The connection is discarded and closed instead of returned to the pool

Under load, this means every request opens a new TCP connection — eliminating the performance benefit of connection pooling and potentially exhausting the server's connection limit.

**The pattern**:

```go
defer func() {
    io.Copy(io.Discard, resp.Body)   // consume any remaining bytes
    resp.Body.Close()
}()
```

`io.Discard` is an `io.Writer` that discards all writes — it's the efficient way to drain without allocating.

**When draining is fine to skip**: If you're certain the response body is already fully consumed (e.g., after `io.ReadAll` or a complete `json.Decode` with no trailing data), the drain is a no-op. The defer pattern is safe either way — `io.Copy` from an exhausted reader returns immediately.

**The limit caveat**: If the server sends a very large error response, draining it takes time and bandwidth. In practice, limit the drain:

```go
io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
resp.Body.Close()
```

</details>

### Q4: What's the difference between `bufio.Scanner` and `bufio.Reader.ReadString`, and when do you choose each?

Tests practical knowledge of the bufio package and where each approach breaks down.

<details>
<summary>Answer</summary>

Both read line-by-line, but they have different contracts and failure modes.

**`bufio.Scanner`**:

```go
scanner := bufio.NewScanner(r)
for scanner.Scan() {
    line := scanner.Text()   // no newline
}
err := scanner.Err()         // check after loop
```

- Returns tokens (lines, words, or custom) without the delimiter
- Fails silently if a token exceeds the buffer size (default 64KB) — `scanner.Err()` returns `bufio.ErrTooLong`
- Error checking happens after the loop, not per-iteration
- Cannot distinguish between "short read" and EOF mid-loop

**`bufio.Reader.ReadString`**:

```go
br := bufio.NewReader(r)
for {
    line, err := br.ReadString('\n')   // includes the '\n'
    if len(line) > 0 {
        process(line)   // ✓ process even when err != nil (partial line on EOF)
    }
    if err == io.EOF {
        break
    }
    if err != nil {
        return err
    }
}
```

- Returns the delimiter as part of the string
- No line length limit (will allocate as needed)
- Error is returned per-call — you see errors immediately
- On the final line with no trailing newline: returns the partial line with `io.EOF`

**When to choose**:

|                          | `bufio.Scanner`                | `bufio.Reader.ReadString` |
| ------------------------ | ------------------------------ | ------------------------- |
| Lines under 64KB         | ✓ cleaner API                  | ✓ works                   |
| Lines over 64KB          | ❌ fails unless buffer resized | ✓ handles naturally       |
| Need delimiter in output | ❌ strips it                   | ✓ included                |
| Custom delimiters        | ✓ via `Split` function         | limited (single byte)     |
| Error handling           | After loop                     | Per call                  |

**Rule of thumb**: Use `bufio.Scanner` for well-formed line-oriented input where you control the source. Use `bufio.Reader` for protocol parsing, binary framing, or input where lines may exceed 64KB.

</details>

---

## Key Takeaways

1. **Partial reads are normal**: `Read` returns however many bytes the OS has ready — never assume it fills the buffer. Use `io.ReadFull`, `io.ReadAll`, or `io.Copy` instead of raw `Read`.
2. **`io.Copy` is your loop**: It handles partial reads, partial writes, and `io.EOF` correctly. Prefer it to manual read/write loops.
3. **`io.TeeReader`**: Duplicates a stream to a second writer as it's read — use it to log or hash data without buffering it.
4. **`io.LimitReader`**: Always wrap untrusted readers (`http.Response.Body`, uploaded files) with a size cap to prevent memory exhaustion.
5. **`io.Pipe`**: Connects a writer API to a reader API with no intermediate buffer — always put the writer in its own goroutine to avoid deadlock.
6. **`bufio.Writer.Flush()`**: Buffered writers must be flushed — unflushed buffers are silently dropped.
7. **`gzip.Writer.Close()`**: Compression writers write finalisation bytes on `Close()` — skipping it produces corrupt output.
8. **HTTP body: drain and close**: Always `defer resp.Body.Close()` and drain with `io.Copy(io.Discard, resp.Body)` to enable connection reuse.
9. **`json.NewDecoder` over `io.ReadAll` + `json.Unmarshal`**: For HTTP bodies and files, stream with a decoder rather than buffering the whole payload first.
10. **`strings.NewReader` and `bytes.Buffer` in tests**: Use these to inject controlled input to functions that accept `io.Reader`, and capture output from functions that accept `io.Writer`.

## Next Steps

This lesson completes the Go refresher series. For further depth:

- [Go Channels Deep Dive](../channels/) — 23-lesson series covering channel patterns exhaustively
- [Go Primitives](../primitives/) — foundational Go data types
- [`io` package documentation](https://pkg.go.dev/io) — full interface and function reference
- [`bufio` package documentation](https://pkg.go.dev/bufio) — buffered I/O reference
