# Go I/O — Readers, Writers & Composition

## Why

- **io.Reader / io.Writer** — The two interfaces everything plugs into: files, network connections, HTTP bodies, compression, encryption, buffers. Learn to compose them and you can build streaming pipelines with zero intermediate allocations.
- **Partial reads are normal** — `Read` returns however many bytes the OS has ready. Code that assumes `Read` fills the buffer works on files, then silently breaks on network connections. Use `io.ReadFull`, `io.ReadAll`, or `io.Copy` instead.
- **io.Copy is your loop** — It handles partial reads, partial writes, and EOF correctly. Prefer it over manual read/write loops.
- **io.LimitReader on untrusted input** — Without a size cap, a malicious client can send infinite data and exhaust your memory. Always limit HTTP bodies, uploads, and external streams.
- **bufio.Writer.Flush()** — Buffered writers must be flushed. Unflushed buffers are silently dropped. Same for `gzip.Writer.Close()` — skipping it produces corrupt output.

## Quick Reference — Core

| Use case             | Method                        |
| -------------------- | ----------------------------- |
| Read everything      | `io.ReadAll(r)`               |
| Read exactly N bytes | `io.ReadFull(r, buf)`         |
| Copy reader → writer | `io.Copy(dst, src)`           |
| Copy at most N bytes | `io.CopyN(dst, src, n)`       |
| Limit reader size    | `io.LimitReader(r, maxBytes)` |
| Discard all data     | `io.Copy(io.Discard, r)`      |
| Concat readers       | `io.MultiReader(r1, r2, r3)`  |
| Write to multiple    | `io.MultiWriter(w1, w2)`      |
| Tee (read + copy)    | `io.TeeReader(r, w)`          |
| In-memory pipe       | `io.Pipe()`                   |

## Quick Reference — bufio

| Use case               | Method                             |
| ---------------------- | ---------------------------------- |
| Read line-by-line      | `bufio.NewScanner(r)`              |
| Read until delimiter   | `bufio.NewReader(r).ReadString(d)` |
| Buffered writes        | `bufio.NewWriter(w)` + `Flush()`   |
| Peek without consuming | `bufio.NewReader(r).Peek(n)`       |

## Quick Reference — In-Memory

| Use case                | Method                 |
| ----------------------- | ---------------------- |
| Read+write buffer       | `bytes.Buffer`         |
| Read-only from `[]byte` | `bytes.NewReader(b)`   |
| Read-only from `string` | `strings.NewReader(s)` |
| Build a string          | `strings.Builder`      |

## io.Copy — Stream Reader to Writer

```go
n, err := io.Copy(dst, src)              // copy until EOF
n, err := io.CopyN(dst, src, 1024)       // copy at most 1024 bytes
n, err := io.CopyBuffer(dst, src, buf)   // use caller-supplied buffer (avoids alloc)
```

## io.ReadAll / io.ReadFull

```go
// Read everything into memory (careful with large/untrusted input)
data, err := io.ReadAll(r)

// Read exactly len(buf) bytes — errors if fewer available
n, err := io.ReadFull(r, buf)
```

## io.LimitReader — Cap Untrusted Input

```go
// ❌ Unbounded — malicious client sends infinite data
data, _ := io.ReadAll(resp.Body)

// ✓ Cap at a reasonable limit
data, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20)) // 10 MB max
```

## io.TeeReader — Read + Copy Simultaneously

```go
var buf bytes.Buffer
tee := io.TeeReader(r, &buf)

data, err := io.ReadAll(tee)
// buf now contains the same data — useful for logging/hashing while processing
```

Hash while copying:

```go
h := sha256.New()
tee := io.TeeReader(src, h)
io.Copy(dst, tee)
hash := hex.EncodeToString(h.Sum(nil))
```

## io.MultiReader — Concat Readers

```go
r := io.MultiReader(
    strings.NewReader("header\n"),
    file,
    strings.NewReader("\nfooter\n"),
)
io.Copy(dst, r) // reads header, then file, then footer
```

## io.MultiWriter — Write to Multiple

```go
w := io.MultiWriter(os.Stdout, logFile, &buf)
fmt.Fprintf(w, "event: %s\n", event) // writes to all three
```

## io.Pipe — Connect Writer API to Reader API

```go
pr, pw := io.Pipe()

go func() {
    defer pw.Close()
    json.NewEncoder(pw).Encode(payload) // blocks until reader consumes
}()

resp, err := http.Post(url, "application/json", pr)
```

Gotcha: fully synchronous — writer blocks until reader reads. Always put the writer in a separate goroutine or you deadlock.

## bufio.Scanner — Line-by-Line Reading

```go
scanner := bufio.NewScanner(r)
for scanner.Scan() {
    line := scanner.Text() // no trailing newline
}
if err := scanner.Err(); err != nil {
    return err
}
```

Default max line: 64KB. Increase for large lines:

```go
scanner.Buffer(make([]byte, 1<<20), 1<<20) // 1 MB max line
```

Custom split functions:

```go
scanner.Split(bufio.ScanWords) // split on whitespace
scanner.Split(bufio.ScanBytes) // one byte at a time
scanner.Split(bufio.ScanRunes) // one UTF-8 rune at a time
```

## bufio.Reader — More Control Than Scanner

```go
br := bufio.NewReader(r)
line, err := br.ReadString('\n') // includes the '\n'
header, err := br.Peek(4)       // look ahead without consuming
b, err := br.ReadByte()         // single byte
br.UnreadByte()                 // push back one byte
```

Use `bufio.Reader` over Scanner when: lines may exceed 64KB, you need per-call error handling, or you're parsing binary protocols.

## bufio.Writer — Batch Small Writes

```go
bw := bufio.NewWriter(w)
for _, line := range lines {
    bw.WriteString(line)
    bw.WriteByte('\n')
}
// ❌ Forgetting Flush — last bytes silently lost
// ✓ Always flush when done
if err := bw.Flush(); err != nil {
    return err
}
```

## bytes.Buffer — In-Memory Read/Write

```go
var buf bytes.Buffer
buf.WriteString("hello ")
buf.Write([]byte{0x01, 0x02})
fmt.Fprintf(&buf, "count: %d", n)

data := buf.Bytes()  // underlying slice (don't modify after further writes)
s := buf.String()    // copy as string
io.Copy(dst, &buf)   // drain buffer into writer
```

## Close/Flush Order — Common Gotcha

```go
// ✓ Close inner writer first, then outer
gz := gzip.NewWriter(f)
gz.Write(data)

if err := gz.Close(); err != nil { return err } // writes gzip footer
if err := f.Close(); err != nil { return err }  // flushes OS buffer

// ❌ Skipping gz.Close() — corrupt gzip file
// ❌ Closing f before gz — footer never written
```

Same applies to `zlib.Writer`, `zip.Writer`, `tar.Writer`, `csv.Writer` (needs `Flush()`).
