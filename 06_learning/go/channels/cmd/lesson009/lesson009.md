# Lesson 9: Fan-Out / Fan-In

## Concept

Fan-out distributes work to multiple goroutines. Fan-in collects results from multiple channels into one. This pattern maximizes throughput.

## Task

1. Create a generator function that returns a <-chan int and sends numbers 1-10 to it (then closes)
2. Create a square function that takes a <-chan int, returns a <-chan int, and sends the square of each received number
3. In main():
   - Start the generator
   - Fan-out: Start 3 square workers, each reading from the same generator channel
   - Fan-in: Merge all 3 output channels into one results channel
   - Print all results

## Hints

- Functions that return channels are a common Go pattern:

```go
  func generator() <-chan int {
    out := make(chan int)
    go func() {
      // send values
      close(out)
    }()
    return out
  }

```

- For fan-in, create a goroutine per input channel that forwards to the merged channel
- Use a WaitGroup to know when to close the merged channel

## Key learning

This pattern is powerful for parallel processing pipelines - each stage can have multiple workers.
