# Lesson 11: Context with Timeout and Deadline

## Concept

Building on context.WithCancel, Go provides context.WithTimeout and context.WithDeadline for automatic cancellation after a duration or at a specific time. This is essential for preventing operations from hanging indefinitely.

## Task

1. Create a function slowOperation(ctx context.Context) (string, error) that:
   - Simulates a slow operation taking 300ms
   - Returns "operation completed" on success
   - Returns ctx.Err() if cancelled before completion

2. In main():
   - First call: Create a context with 500ms timeout, call slowOperation - should succeed
   - Second call: Create a context with 100ms timeout, call slowOperation - should timeout
   - Print the result or error for each call

## Hints

- Create timeout context: ctx, cancel := context.WithTimeout(context.Background(), 500\*time.Millisecond)
- Always defer cancel() even if timeout fires (prevents resource leaks)
- Use select to race between work completion and context cancellation:

```go
select {
  case <-time.After(300 \* time.Millisecond):
    return "operation completed", nil
  case <-ctx.Done():
   return "", ctx.Err()
}
```

- ctx.Err() returns context.DeadlineExceeded for timeouts

Expected output:

```sh
Call 1: operation completed
Call 2: context deadline exceeded
```

## Key learning

Always set timeouts on external calls (HTTP, database, etc.) to prevent goroutine leaks and hung processes. The defer cancel() pattern ensures cleanup even if the operation completes before timeout.
