# Lesson 6: Done Channel (Cancellation Pattern)

## Concept

A "done" channel signals goroutines to stop working. This is how you gracefully shut down concurrent operations.

## Task

1. Create a done channel of type struct{} (empty struct - uses zero memory)
2. Create a results channel of type int
3. Launch a goroutine that runs a loop, sending incrementing numbers (0, 1, 2...) to results every 100ms - but it should exit when done is closed
4. In main(), receive and print 5 numbers, then close done to stop the worker
5. Add a small sleep at the end to let the goroutine exit, then print "worker stopped"

## Hints

- Use select inside the goroutine to check both sending and done:

```go
  select {
  case <-done:
    return
  case results <- value:
  // sent successfully
  }
```

- Close signals all receivers: close(done)

Expected output:

```sh
0
1
2
3
4
```

worker stopped

## Key learning

This pattern is the foundation of Go's context.Context cancellation.

## Questions I had during implementation

### how does select work with both sends and receives

A select case can be:

- Receive: case msg := <-ch: - ready when someone is waiting to receive
- Send: case ch <- value: - ready when someone is waiting to send

How your code flows

case results <- count: // This line attempts to SEND
count += 1 // This runs AFTER send succeeds

1. select evaluates which cases are ready
2. results <- count is ready when main() is blocked on <-results
3. When selected, the send completes (value transferred)
4. Then the case body runs (count += 1)

Why this is useful

Without select, the send would block with no way to check done:

// BAD: Can't cancel while blocked on send

```go
results <- count // stuck here until someone receives
```

With select, you can attempt to send OR respond to cancellation:

// GOOD: Can respond to done even while waiting to send

```go
select {
case <-done:
  return // cancelled!
case results <- count:
  count += 1 // send succeeded
}
```

### if the select is random how come done is executed at the correct time

The case body (count += 1) only executes after the channel operation completes successfully. The send itself is part of the case condition, not the body.

Timeline of your code

| Time              | done status | main() doing         | results <- count ready? | <-done ready? |
| ----------------- | ----------- | -------------------- | ----------------------- | ------------- |
| 0-500ms           | open        | waiting on <-results | Yes (receiver waiting)  | No            |
| after 5th receive | open→closed | close(done)          | No (no receiver)        | Yes           |

Why done fires at the right time

1. While main is receiving: done is open, so <-done blocks (not ready). Only results <- count is ready because main is waiting to receive. No random choice needed - only one option.
2. After main closes done: Main stops receiving from results. Now:
   - results <- count blocks (no receiver waiting) - not ready
   - <-done returns immediately (closed channel) - ready

Again, only one option is ready.

When would randomness matter?

If both were ready at the exact same moment:

// Hypothetical: both ready simultaneously

```go
select {
case <-done: // ready
case results <- count: // also ready
}
```

// Go randomly picks one

In your code, this race window is tiny/nonexistent because main stops receiving before closing done.

Key insight: Random selection is for fairness when multiple channels compete, not for causing unpredictable behavior.
