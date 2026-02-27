# Lesson 2: Buffered Channels

## Concept

Buffered channels have capacity. Sends only block when the buffer is full, receives only block when empty.

## Task

1. Create a buffered channel of type int with capacity 3
2. Send three values (1, 2, 3) to the channel without using a goroutine
3. Receive and print all three values

## Hints

- Buffered channel: ch := make(chan int, 3)
- You can send up to 3 values before blocking

## Expected output

```sh
1
2
3
```

## Key learning

With unbuffered channels, you couldn't do this without a goroutine (deadlock). Buffered channels decouple send/receive timing up to the buffer size.
