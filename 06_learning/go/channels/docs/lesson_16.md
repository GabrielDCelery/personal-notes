# Lesson 16: Worker Pool

Concept: Instead of spawning a goroutine per job, create a fixed pool of workers that pull jobs from a shared channel. More efficient for high job volumes - avoids goroutine creation overhead.

Your task:

1. Create a worker(id int, jobs <-chan int, results chan<- int, wg \*sync.WaitGroup) function that:
   - Loops over jobs channel (for job := range jobs)
   - Prints "worker X processing job Y"
   - Simulates work (100ms sleep)
   - Sends job \* 2 to results
   - Calls wg.Done() when the jobs channel closes

2. In main():
   - Create a jobs channel (buffered, size 10)
   - Create a results channel (buffered, size 10)
   - Start 3 workers (fixed pool)
   - Send 10 jobs (numbers 1-10) to the jobs channel
   - Close jobs channel to signal no more work
   - Wait for workers to finish, then close results
   - Collect and print all results

Expected output pattern:

```txt
worker 1 processing job 1
worker 2 processing job 2
worker 3 processing job 3
worker 1 processing job 4
...
results: [2, 4, 6, 8, ...]
```

Hints:

- Workers exit when for job := range jobs ends (channel closed + empty)
- Use WaitGroup to know when all workers done
- Close results in a goroutine that waits for workers:

```go
go func() {
  wg.Wait()
  close(results)
}()
```

Key learning: Worker pools are ideal when you have many small jobs - reusing goroutines is cheaper than constantly creating new ones. This pattern is used in web servers, task queues, etc.
