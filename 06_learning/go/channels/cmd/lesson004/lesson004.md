# Lesson 4: The select Statement

## Concept

select lets you wait on multiple channel operations simultaneously. It's like a switch for channels - whichever case is ready first executes.

## Task

1. Create two channels: ch1 and ch2 (both chan string)
2. Launch two goroutines:
   - First sends "from channel 1" after 100ms delay
   - Second sends "from channel 2" after 200ms delay

3. Use a select inside a loop to receive from whichever channel is ready first
4. Print both messages as they arrive, then exit

## Hints

- time.Sleep(100 \* time.Millisecond) for delays
- Basic select structure:

```go
  select {
  case msg := <-ch1:
  // handle ch1
  case msg := <-ch2:
  // handle ch2
  }
```

- You need to receive exactly 2 messages total

Expected output:

```sh
from channel 1
from channel 2
```

## Key learning

select is fundamental for handling multiple concurrent operations, timeouts, and cancellation patterns.

## Questions I had during implementation

### Why is a closed channel "always ready"?

This is a design decision in Go. A closed channel can always be read from - it returns the zero value instantly. This allows consumers to drain remaining buffered values and detect closure.

```go
ch := make(chan int, 2)
ch <- 1
ch <- 2
close(ch)

fmt.Println(<-ch) // 1 (buffered value)
fmt.Println(<-ch) // 2 (buffered value)
fmt.Println(<-ch) // 0 (zero value, channel closed)
fmt.Println(<-ch) // 0 (zero value, forever)
```

Closing doesn't "lock" the channel - it signals "no more sends." Reads always succeed.

### Why doesn't default get picked randomly in a select?

default is not part of the random selection. The rules are:

1. If one or more channel cases are ready → pick randomly among only those ready cases
2. If zero channel cases are ready → run default
3. If zero channel cases are ready AND no default → block

> [!WARNING]
> default is the fallback for "nothing ready," not an equal participant.

### How does ch2 ever get selected if ch1 is always ready?

Both closed channels are always ready simultaneously. So select randomly picks between ch1 and ch2 each iteration. Both get hit frequently, but default never runs because at least one case is always ready.

```sh
Iteration 1: ch1 ready, ch2 ready → random pick → ch1
Iteration 2: ch1 ready, ch2 ready → random pick → ch2
Iteration 3: ch1 ready, ch2 ready → random pick → ch1
... (default never runs)
```
