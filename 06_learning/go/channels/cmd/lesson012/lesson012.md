# Lesson 12: Graceful Shutdown with OS Signals

## Concept

Production applications need to shut down cleanly when receiving termination signals (Ctrl+C, container stop, etc.). Go's `os/signal` package combined with context provides an elegant solution.

## Task

1.  Create a worker function `worker(ctx context.Context, id int)` that:
    - Simulates processing work in a loop (every 500ms prints "worker X: processing...")
    - Stops gracefully when context is cancelled
    - Prints "worker X: shutting down" before returning
2.  In main():
    - Use `signal.NotifyContext` to create a context that cancels on SIGINT or SIGTERM
    - Start 3 workers
    - Use a WaitGroup to track when all workers have stopped
    - Print "all workers stopped, exiting" after cleanup
3.  Test by running the program and pressing Ctrl+C Hints:
    - Import `os/signal` and `syscall`
    - Create signal-aware context:

```go
ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
defer stop()
```

- Worker loop patterns

```go
for {
    select {
    case <-ctx.Done():
        fmt.Printf("worker %d: shutting down\n", id)
        return
    case <-time.After(500 * time.Millisecond):
        fmt.Printf("worker %d: processing...\n", id)
    }
}
```

- Use `wg.Wait()` in main to ensure all workers finish before exiting

```txt
worker 1: processing...
worker 2: processing...
worker 3: processing...
worker 1: processing...
worker 1: shutting down
worker 2: shutting down
worker 3: shutting down
all workers stopped, exiting
```

## Key learning

`signal.NotifyContext` (Go 1.16+) is the modern way to handle shutdown. It combines signal handling with context cancellation, allowing your entire application to respond to termination signals through the same context propagation you use for timeouts and cancellation.

## Why graceful shutdown matters

1. **Data integrity** - Allow in-flight database transactions to complete
2. **Connection cleanup** - Close network connections properly
3. **Resource release** - Release file handles, flush buffers
4. **User experience** - Don't drop active HTTP requests mid-response

## Common signals

| Signal  | Trigger                        | Default behavior  |
| ------- | ------------------------------ | ----------------- |
| SIGINT  | Ctrl+C                         | Terminate         |
| SIGTERM | `kill` command, container stop | Terminate         |
| SIGKILL | `kill -9`                      | Cannot be caught! |

## Production pattern

In real applications, you often give a grace period:

```go
ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
defer stop()

// Start server/workers with ctx...

<-ctx.Done()
fmt.Println("shutdown signal received")

// Give workers time to finish (e.g., 10 seconds)
shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

// Shutdown server with grace period
server.Shutdown(shutdownCtx)
```
