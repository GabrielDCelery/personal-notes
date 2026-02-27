# Lesson 3: Directional Channels (Channel Types)

## Concept

You can restrict a channel to send-only or receive-only. This enforces correct usage at compile time and makes function signatures self-documenting.

## Task

1. Create a function producer(ch chan<- int) that sends numbers 1-5 to the channel, then closes it
2. Create a function consumer(ch <-chan int) that receives all values and prints them
3. In main(), create a bidirectional channel, pass it to both functions (producer in a goroutine), and let consumer run

## Hints

- chan<- int = send-only (arrow points INTO channel)
- <-chan int = receive-only (arrow points OUT OF channel)
- A regular chan int can be passed where directional channels are expected (Go converts automatically)

## Expected output

```sh
1
2
3
4
5
```

## Key learning

Directional channels prevent bugs. If consumer accidentally tries to close or send to the channel, the compiler catches it. This is especially valuable in larger codebases.
