# Go Goroutines & Channels

## Why

- **Goroutines are cheap** — They start with ~2KB of stack (vs ~1MB for OS threads). You can spawn thousands without worry. The Go scheduler multiplexes them onto OS threads.
- **Unbuffered vs buffered** — Unbuffered channels synchronize sender and receiver (handoff). Buffered channels decouple them up to the buffer size. Use unbuffered by default; buffer only when you have a reason.
- **Always close from the sender side** — The sender knows when there's no more data. Closing from the receiver can cause a panic if the sender writes to a closed channel.
- **chan struct{} for signals** — Uses zero memory per send. It's idiomatic for done/cancellation channels where the value doesn't matter, only the event.
- **select picks randomly** — When multiple cases are ready, Go picks one at random. This prevents starvation and means you can't rely on case ordering.

## Quick Reference

| Use case           | Method                              |
| ------------------ | ----------------------------------- |
| Spawn goroutine    | `go func()`                         |
| Send/receive       | `ch <- val` / `val := <-ch`         |
| Buffered channel   | `make(chan T, n)`                   |
| Close channel      | `close(ch)`                         |
| Select on multiple | `select { case ... }`               |
| Done/cancel signal | `chan struct{}`                     |
| Fan-out / fan-in   | multiple goroutines, single channel |

## Goroutines

### 1. Basic goroutine

```go
go func() {
    fmt.Println("running concurrently")
}()
```

### 2. Goroutine with argument (avoid closure capture bug)

```go
for _, v := range items {
    v := v // shadow to capture correctly
    go func() {
        process(v)
    }()
}
```

## Channels

### 3. Unbuffered channel (synchronous handoff)

```go
ch := make(chan int)

go func() {
    ch <- 42
}()

val := <-ch
```

### 4. Buffered channel (async up to capacity)

```go
ch := make(chan int, 3)

ch <- 1
ch <- 2
ch <- 3
// does not block until full
```

### 5. Range over channel

```go
ch := make(chan int)

go func() {
    for _, v := range []int{1, 2, 3} {
        ch <- v
    }
    close(ch) // range exits when channel is closed
}()

for v := range ch {
    fmt.Println(v)
}
```

### 6. Done channel (cancellation signal)

```go
done := make(chan struct{})

go func() {
    for {
        select {
        case <-done:
            return
        default:
            doWork()
        }
    }
}()

close(done) // signal all listeners to stop
```

### 7. Select (multiplex channels)

```go
select {
case msg := <-ch1:
    fmt.Println("ch1:", msg)
case msg := <-ch2:
    fmt.Println("ch2:", msg)
case <-time.After(1 * time.Second):
    fmt.Println("timeout")
}
```

### 8. Non-blocking send/receive

```go
select {
case ch <- val:
    // sent
default:
    // channel full or no receiver, skip
}
```

## Patterns

### 9. Fan-out (distribute work)

```go
jobs := make(chan int, 10)

for w := 0; w < 3; w++ {
    go func() {
        for j := range jobs {
            process(j)
        }
    }()
}

for _, j := range work {
    jobs <- j
}
close(jobs)
```

### 10. Fan-in (merge results)

```go
func merge(cs ...<-chan int) <-chan int {
    out := make(chan int)
    var wg sync.WaitGroup

    for _, c := range cs {
        wg.Add(1)
        go func(ch <-chan int) {
            defer wg.Done()
            for v := range ch {
                out <- v
            }
        }(c)
    }

    go func() {
        wg.Wait()
        close(out)
    }()

    return out
}
```

### 11. Pipeline

```go
func generate(nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        for _, n := range nums {
            out <- n
        }
        close(out)
    }()
    return out
}

func double(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        for n := range in {
            out <- n * 2
        }
        close(out)
    }()
    return out
}

// Usage
for v := range double(generate(1, 2, 3)) {
    fmt.Println(v) // 2, 4, 6
}
```
