# Lesson 18: Worker Pool with Context and errgroup

Concept: Combine worker pool with graceful shutdown and error handling. This is the production-ready pattern - workers respond to cancellation and errors propagate properly.

Your task:

1. Create a worker(ctx context.Context, id int, jobs <-chan int) error that:
   - Processes jobs in a loop
   - Returns an error if job value is 7 (simulated failure)
   - Responds to context cancellation
   - Prints "worker X: processing job Y" and "worker X: shutting down" on exit

2. In main():
   - Create a signal context (SIGINT, SIGTERM)
   - Use errgroup.WithContext derived from signal context
   - Start 3 workers using g.Go()
   - Send jobs 0-9 to the jobs channel
   - Handle three scenarios:
     - Ctrl+C → all workers stop gracefully
     - Worker hits job 7 → error propagates, others stop
     - All jobs complete → clean exit

3. Print which scenario triggered shutdown

Expected behavior:

```txt
worker 0: processing job 0
worker 1: processing job 1
worker 2: processing job 2
...
worker 1: processing job 7
worker 1: encountered error
worker 0: shutting down
worker 2: shutting down
shutdown: worker error: job 7 failed
```

Hints:

Worker structure:

```go
  func worker(ctx context.Context, id int, jobs <-chan int) error {
      for {
          select {
          case <-ctx.Done():
              fmt.Printf("worker %d: shutting down\n", id)
              return nil
          case job, ok := <-jobs:
              if !ok {
                  return nil  // jobs channel closed
              }
              if job == 7 {
                  return fmt.Errorf("job %d failed", job)
              }
              fmt.Printf("worker %d: processing job %d\n", id, job)
          }
      }
  }
```

Main structure:

```go
signalCtx, stop := signal.NotifyContext(...)
defer stop()

g, ctx := errgroup.WithContext(signalCtx)

jobs := make(chan int)

// Start workers with g.Go()

// Send jobs in separate goroutine (so it can be cancelled too)

err := g.Wait()
// Check signalCtx.Err() vs err to determine cause
```

Key learning: This pattern handles all shutdown scenarios cleanly - signals, errors, and normal completion. It's the foundation for production task processors, job queues, and server workers.
