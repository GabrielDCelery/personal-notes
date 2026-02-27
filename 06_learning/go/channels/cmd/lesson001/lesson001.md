# Lesson 1: Basic Channel Creation and Communication

## Concept

Channels are typed conduits for communication between goroutines. They follow the principle "Don't communicate by sharing memory; share memory by communicating."

## Task

1. Create an unbuffered channel of type string
2. Launch a goroutine that sends the message "hello from goroutine" to the channel
3. In main(), receive from the channel and print the result

## Hints

- Create a channel: ch := make(chan string)
- Send to channel: ch <- value
- Receive from channel: value := <-ch
- Launch goroutine: go func() { ... }()

## Expected output

```sh
hello from goroutine
```

## Key learning

Unbuffered channels block on send until someone receives, and block on receive until someone sends. This provides synchronization.
