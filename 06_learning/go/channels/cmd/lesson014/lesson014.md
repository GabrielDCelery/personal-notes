# Lesson 14

## Concept

Use a buffered channel as a counting semaphore to limit concurrent operations. This prevents overwhelming resources (APIs, databases, file system).

## Your task

1. Create a function process(id int) that:
   - Prints "starting job X"
   - Simulates work with 1 second sleep
   - Prints "finished job X"

2. In main():
   - Create 10 jobs to process
   - Use a semaphore to limit to max 3 concurrent workers
   - Use a WaitGroup to wait for all jobs to complete
   - Print "all jobs completed" at the end

3. Observe the output - you should see jobs starting in batches of 3

Expected output pattern:

```txt
starting job 0
starting job 1
starting job 2
finished job 1
starting job 3
finished job 0
starting job 4
...
all jobs completed
```

## Hints

- Semaphore: sem := make(chan struct{}, 3)
- Acquire before work: sem <- struct{}{}
- Release after work: <-sem (use defer)
- Remember to pass loop variable correctly to goroutine

Key learning: Buffered channels as semaphores is idiomatic Go - no need for external libraries. This pattern is essential for rate-limiting concurrent access to limited resources.

## Why semaphores

Why:

- Clever use of buffered channels you haven't seen yet
- Very practical - limiting concurrent DB connections, API calls, file handles
- Simple concept but powerful
- Builds directly on your existing knowledge
- Natural stepping stone to worker pools

The core pattern:

```go
sem := make(chan struct{}, 3) // max 3 concurrent

for _, job := range jobs {
    sem <- struct{}{} // acquire (blocks if 3 already running)
    go func(j Job) {
        defer func() { <-sem }() // release
        // do work
    }(job)
}
```

It's essentially a "counting lock" using a buffered channel - elegant and idiomatic Go.

Want to go with semaphores?
