# Lesson 5: Timeouts with select

## Concept

Using select with time.After() lets you set deadlines on channel operations, preventing your program from waiting forever.

## Task

1. Create a channel ch of type string
2. Launch a goroutine that waits 500ms before sending "operation completed"
3. Use select to either:
   - Receive from ch and print the message, OR
   - Timeout after 200ms and print "timeout: operation took too long"

## Hints

- time.After(duration) returns a <-chan time.Time that receives a value after the duration
- Use it directly in a select case:
  case <-time.After(200 \* time.Millisecond):
  // timed out

Expected output:

```sh
timeout: operation took too long
```

Bonus: After you get it working, try changing the timeout to 600ms - you should see the success message instead.

## Key learning

This pattern is essential for network calls, database queries, or any operation that might hang.
