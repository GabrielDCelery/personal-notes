# Lesson 8: sync.WaitGroup

## Concept

A WaitGroup waits for a collection of goroutines to finish. This solves the "how do I know when all workers are done?" problem properly.

## Task

Refactor your worker pool to use WaitGroup:

1. Create a sync.WaitGroup
2. Before launching each worker, call wg.Add(1)
3. Each worker calls defer wg.Done() at the start of its function
4. Launch a separate goroutine that:
   - Waits for all workers with wg.Wait()
   - Then closes results

5. Now you can use for result := range results safely

Hints:

- Import "sync"
- Pass WaitGroup by pointer: wg \*sync.WaitGroup
- Structure:

```go
var wg sync.WaitGroup

for i := range 3 {
  wg.Add(1)
  go worker(&wg, ...)
}

go func() {
  wg.Wait()
  close(results)
}()

```

Expected output: Same as before, but now using range results.

## Key learning

WaitGroup is essential for coordinating goroutine completion, especially when you don't know how many items will be processed.

## Questions I had during implementation

### does job := range jobs exit when all tasks have been completed

● Yes, exactly. for job := range jobs exits when both:

1. The channel is closed, AND
2. All remaining buffered values have been received

Here's the behavior:

```go
jobs := make(chan int, 3)
jobs <- 1
jobs <- 2
jobs <- 3
close(jobs)

for job := range jobs {
  fmt.Println(job)
}
// Prints 1, 2, 3, then exits loop
```

What happens under the hood

for range is roughly equivalent to:

```go
for {
  job, ok := <-jobs
  if !ok {
    break // channel closed AND empty
  }
  // use job
}

```

Key points

| Channel state               | range behavior            |
| --------------------------- | ------------------------- |
| Open, has value             | receives value, continues |
| Open, empty                 | blocks waiting            |
| Closed, has buffered values | receives remaining values |
| Closed, empty               | exits loop                |

That's why in your worker pool, workers exit cleanly after processing all jobs - they drain the channel, then range sees it's closed and empty, and exits.
