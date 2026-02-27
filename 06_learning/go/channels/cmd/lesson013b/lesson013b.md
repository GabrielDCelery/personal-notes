# Lesson 13b - combine error groups and signal interrupts

## Task

1. Two cancellation sources - either a worker fails OR user hits Ctrl+C
2. Context chaining - signal context as parent, errgroup derives from it
3. Real-world applicability - this is how production services actually work

Here's the pattern:

```go
// Signal context as the root
ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
defer stop()

// errgroup derives from signal context
g, ctx := errgroup.WithContext(ctx)

g.Go(func() error {
// ... workers check ctx.Done()
// cancelled by EITHER signal OR another worker's error
})
```

Exercise idea for Lesson 13b:

- 3 workers doing "long work" (e.g., 2 seconds each)
- If you Ctrl+C, all workers stop gracefully
- If one worker fails (simulate after 500ms), others stop
- Print which scenario triggered shutdown
  - Signal context from lesson 12: signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

- Pass that context to errgroup.WithContext(ctx)
- Workers check ctx.Done() - triggers from either source

Test both scenarios:

1. Let a worker fail → others should stop
2. Press Ctrl+C before failure → all should stop

## How to tell whether the error group or the signal interrupt stopped the workers

You can check the signal context separately. Keep a reference to it:

```go
signalCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
defer stop()

g, ctx := errgroup.WithContext(signalCtx)
// ... start workers with ctx ...
err := g.Wait()

// After Wait(), check which caused the shutdown:
if signalCtx.Err() != nil {
  fmt.Println("shutdown: received signal")
} else if err != nil {
  fmt.Println("shutdown: worker error:", err)
}
```

Why this works:

| Scenario       | signalCtx.Err() | g.Wait() returns           |
| -------------- | --------------- | -------------------------- |
| Ctrl+C pressed | != nil          | nil (workers returned nil) |
| Worker failed  | nil             | the error                  |

The errgroup's derived context gets cancelled in both cases, but only the signal context "knows" if a signal was the root cause.

Inside workers, you can't easily tell - ctx.Done() fires either way. But typically you don't need to know inside the worker; you just exit cleanly and
let main report the cause.
