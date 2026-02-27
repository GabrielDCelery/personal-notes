# Lesson 20: Or-Done Pattern and Channel Utilities

Concept: Create reusable channel utility functions that simplify common patterns and prevent goroutine leaks.

Your task:

Build two utility patterns:

1. orDone(ctx, ch) - Wraps any receive channel to automatically respect context cancellation
   - Prevents verbose select statements in every loop
   - Returns a new channel that closes when either input closes OR context is done

2. merge(ctx, channels...) - Combines multiple channels into one output channel
   - Takes variadic channels of same type
   - Returns single output channel with all values
   - Respects context cancellation
   - Closes output when all inputs are closed

3. In main():
   - Create 3 generator goroutines (each sends different numbers)
   - Use merge() to combine them
   - Use orDone() to read merged output
   - Cancel context after 2 seconds
   - Demonstrate clean shutdown

Why this matters: These utilities are real patterns used in production Go code. They eliminate boilerplate and prevent common mistakes like goroutine
leaks.

---

Would you like me to create this as lesson 20, or would you prefer a different topic? Some alternatives could be:

- Tee pattern (split one channel to multiple outputs)
- Bridge pattern (flatten channel of channels)
- Heartbeat pattern (monitor goroutine health)
- Circuit breaker with channels
