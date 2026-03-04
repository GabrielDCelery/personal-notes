# Go File Operations

## Why

- **os.WriteFile vs os.Create** — WriteFile is a one-shot convenience for small data. Create gives you a file handle for multiple writes, buffered I/O, or when you need more control.
- **defer f.Close()** — File descriptors are a limited OS resource. Leaking them causes "too many open files" errors under load.
- **bufio.NewWriter** — Each f.Write is a syscall. Buffered writer batches them, which is significantly faster for many small writes. Don't forget to Flush.
- **O_APPEND flag** — Without it, os.Create truncates the file. Use OpenFile with O_APPEND when you want to add to existing content (log files, etc).

## Quick Reference

| Use case              | Method                        |
| --------------------- | ----------------------------- |
| Simple one-shot write | `os.WriteFile`                |
| Multiple writes       | `os.Create` + `Write`         |
| Appending             | `os.OpenFile` with `O_APPEND` |
| High performance      | `bufio.NewWriter`             |
| Formatted data        | `fmt.Fprintf`                 |

## Writing to a File

### 1. Write all at once (simplest)

```go
err := os.WriteFile("file.txt", []byte("hello world"), 0644)
```

### 2. Create/open then write

```go
f, err := os.Create("file.txt")  // truncates if exists
if err != nil {
    return err
}
defer f.Close()

f.WriteString("hello world\n")
// or
f.Write([]byte("hello world\n"))
```

### 3. Append to existing file

```go
f, err := os.OpenFile("file.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
if err != nil {
    return err
}
defer f.Close()

f.WriteString("new line\n")
```

### 4. Buffered writer (best for many/large writes)

```go
import "bufio"

f, err := os.Create("file.txt")
if err != nil {
    return err
}
defer f.Close()

w := bufio.NewWriter(f)
w.WriteString("hello world\n")
w.Flush() // don't forget this
```

### 5. fmt.Fprintf (formatted output)

```go
f, _ := os.Create("file.txt")
defer f.Close()

fmt.Fprintf(f, "name: %s, age: %d\n", "Alice", 30)
```
