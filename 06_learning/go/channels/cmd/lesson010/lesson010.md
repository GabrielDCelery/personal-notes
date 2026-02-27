# Lesson 10: Context for Cancellation

## Concept

context.Context is Go's standard way to handle cancellation, deadlines, and request-scoped values. It's preferred over done channels in production code.

## Task

1. Create a generator that sends incrementing numbers (0, 1, 2...) every 100ms
2. Instead of a done channel, accept ctx context.Context as the first parameter
3. Use select with ctx.Done() to stop when cancelled
4. In main():
   - Create a context with context.WithCancel(context.Background())
   - Start the generator
   - Receive and print 5 numbers
   - Call cancel() to stop the generator
   - Print "generator stopped"

## Hints

- Import "context"
- ctx.Done() returns a channel that closes when cancelled
- Pattern inside generator:

```go
select {
  case <-ctx.Done():
return
  case out <- value:
}
```

## Key learning

Context is the idiomatic way to handle cancellation in Go APIs - you'll see it as the first parameter in most library functions.

## Questions I had during implementation

### why is context used over done channel

1. Deadlines and Timeouts Built-in

// Done channel: manual timeout setup

```go
done := make(chan struct{})
go func() {
  time.Sleep(5 * time.Second)
  close(done)
}()
```

// Context: one line

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
```

2. Propagation Through Call Stack

Context passes naturally through function calls:

```go
func HandleRequest(ctx context.Context) {
  result := QueryDatabase(ctx) // passes cancellation down
  SendEmail(ctx, result) // same context
}
```

With done channels, you'd pass them manually everywhere - messy.

3. Hierarchy (Parent-Child Cancellation)

```go
parentCtx, parentCancel := context.WithCancel(context.Background())
childCtx, childCancel := context.WithCancel(parentCtx)

parentCancel() // cancels BOTH parent and child
```

Done channels don't have this built-in.

4. Standard Library Integration

Most Go libraries expect context:

```go
http.NewRequestWithContext(ctx, "GET", url, nil)
db.QueryContext(ctx, "SELECT ...")
```

5. Carries Values (use sparingly)

```go
ctx = context.WithValue(ctx, "requestID", "abc123")
```

When done channels are still used

- Internal goroutine coordination in small scopes
- Simple one-off cancellation
- Legacy code
- When you explicitly want minimal overhead

Summary

| Feature               | Done Channel | Context  |
| --------------------- | ------------ | -------- |
| Simple cancellation   | Yes          | Yes      |
| Timeouts/deadlines    | Manual       | Built-in |
| Hierarchy             | Manual       | Built-in |
| Standard library      | No           | Yes      |
| Request-scoped values | No           | Yes      |

Context is the "batteries included" version. Done channels are fine for simple internal use, but context is the standard API choice.
