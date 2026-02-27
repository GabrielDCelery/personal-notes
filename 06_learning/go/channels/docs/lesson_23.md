# Lesson 23: Broadcast Pattern - Non-Blocking Fan-Out

Concept: Send values to multiple consumers without being blocked by slow consumers. Unlike tee (lesson 21), slow consumers don't block fast ones.

Your task:

1. broadcaster(ctx) (send func(T), subscribe func() <-chan T)
   - Returns two functions:
     - send(val) - broadcasts value to all subscribers (non-blocking)
     - subscribe() - returns a new channel for a consumer
   - Each subscriber gets their own buffered channel
   - Dropped messages if subscriber's buffer is full (or use alternative strategy)

2. In main():
   - Create broadcaster
   - Subscribe 3 consumers with different speeds:
     - Fast: no delay
     - Medium: 50ms delay
     - Slow: 200ms delay
   - Producer sends values every 100ms
   - Run for 2 seconds
   - Print what each consumer received

Expected behavior: Fast consumer gets all messages, slow consumer misses some (or lags behind).
