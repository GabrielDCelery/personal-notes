# Go Channels

## Quick Reference

| Use case               | Method                     |
| ---------------------- | -------------------------- |
| Unbuffered channel     | `ch := make(chan T)`       |
| Buffered channel       | `ch := make(chan T, size)` |
| Send                   | `ch <- val`                |
| Receive                | `val := <-ch`              |
| Close                  | `close(ch)`                |
| Range over             | `for val := range ch`      |
| Select                 | `select { case ... }`      |
| Send-only direction    | `chan<- T`                 |
| Receive-only direction | `<-chan T`                 |

## Basics

### 1. Unbuffered — sender blocks until receiver is ready

```go
ch := make(chan string)

go func() {
    ch <- "hello"  // blocks until someone reads
}()

msg := <-ch  // "hello"
```

### 2. Buffered — sender blocks only when buffer is full

```go
ch := make(chan int, 3)

ch <- 1  // doesn't block
ch <- 2  // doesn't block
ch <- 3  // doesn't block
// ch <- 4  // would block — buffer full

fmt.Println(<-ch)  // 1
```

### 3. Close and range

```go
ch := make(chan int)

go func() {
    for i := 0; i < 5; i++ {
        ch <- i
    }
    close(ch)  // signal no more values
}()

for val := range ch {
    fmt.Println(val)  // 0, 1, 2, 3, 4
}
```

### 4. Check if channel is closed

```go
val, ok := <-ch
if !ok {
    // channel is closed and drained
}
```

## Channel Direction

### 5. Restrict direction in function signatures

```go
func producer(out chan<- int) {   // send-only
    out <- 42
}

func consumer(in <-chan int) {    // receive-only
    fmt.Println(<-in)
}

ch := make(chan int)
go producer(ch)
consumer(ch)
```

## Select

### 6. Wait on multiple channels

```go
select {
case msg := <-ch1:
    fmt.Println("from ch1:", msg)
case msg := <-ch2:
    fmt.Println("from ch2:", msg)
case <-ctx.Done():
    return
}
// If both ready, one is picked at random
```

### 7. Non-blocking with default

```go
select {
case msg := <-ch:
    fmt.Println(msg)
default:
    fmt.Println("no message ready")
}
```

### 8. Send with timeout

```go
select {
case ch <- val:
    // sent
case <-time.After(5 * time.Second):
    // timed out
}
```

## Patterns

### 9. Done / signal channel

```go
done := make(chan struct{})

go func() {
    doWork()
    close(done)  // signal completion
}()

<-done  // wait for completion
```

### 10. Fan-out — multiple goroutines reading from one channel

```go
jobs := make(chan int, 100)

for i := 0; i < 5; i++ {
    go func() {
        for job := range jobs {
            process(job)
        }
    }()
}

for j := 0; j < 50; j++ {
    jobs <- j
}
close(jobs)
```

### 11. Fan-in — merge multiple channels into one

```go
func fanIn(channels ...<-chan string) <-chan string {
    out := make(chan string)
    var wg sync.WaitGroup

    for _, ch := range channels {
        wg.Add(1)
        go func(c <-chan string) {
            defer wg.Done()
            for val := range c {
                out <- val
            }
        }(ch)
    }

    go func() {
        wg.Wait()
        close(out)
    }()

    return out
}
```

### 12. Pipeline — chain stages

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

func square(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        for n := range in {
            out <- n * n
        }
        close(out)
    }()
    return out
}

ch := square(generate(1, 2, 3))
for val := range ch {
    fmt.Println(val)  // 1, 4, 9
}
```

### 13. Semaphore — limit concurrency

```go
sem := make(chan struct{}, 3)  // max 3 concurrent

for _, item := range items {
    sem <- struct{}{}  // acquire
    go func(it Item) {
        defer func() { <-sem }()  // release
        process(it)
    }(item)
}
```
