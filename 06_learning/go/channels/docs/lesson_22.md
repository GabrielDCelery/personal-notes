# Lesson 22: Bridge Pattern - Flatten Channel of Channels

Concept: Take a channel that receives channels and flatten all values into a single output channel. Useful when workers dynamically create result
channels.

Your task:

1. bridge(ctx, chanOfChans) <-chan T - Flattens channel of channels
   - Input: <-chan (<-chan T) - a channel that receives channels
   - Output: <-chan T - single channel with all values
   - For each channel received, read all its values
   - Send all values to single output
   - Respects context cancellation
   - Closes output when input closes AND all sub-channels are drained

2. worker(id, ctx) <-chan int - Simulates dynamic worker
   - Returns a channel
   - Sends 3-5 values with delays
   - Prints "worker X: value Y"

3. In main():
   - Create channel of channels: chanOfChans := make(chan (<-chan int))
   - Spawn goroutine that creates 3 workers dynamically
   - Send each worker's channel into chanOfChans
   - Close chanOfChans after all workers sent
   - Use bridge(ctx, chanOfChans) to flatten
   - Print all received values
   - Context with 5 second timeout

Expected output:
worker 1: 0
worker 2: 0
worker 3: 0
received: 0
received: 0
worker 1: 1
received: 0
worker 2: 1
...

Key challenge: You need to spawn a goroutine for each sub-channel received to drain it concurrently. Use a WaitGroup to know when all are done.

## When to use the pattern

Example scenario:

```go
// Worker creates a channel of results
func worker() <-chan int { ... }

// You spawn 5 workers dynamically
workerChannels := make(chan (<-chan int))
go workerChannels <- worker() // Send channels into the channel!

// Bridge flattens all worker outputs into one
output := bridge(ctx, workerChannels)
```

Why Bridge Pattern Exists

Problem: Dynamic Workers with Unknown Lifecycle

### Scenario 1: Processing files in a directory

```go
// BAD: Pre-create one shared channel
results := make(chan Result)

// Problem: How many workers? When do they finish?
for _, file := range files {
    go processFile(file, results)  // All share same channel
}
// When do you close `results`? Need WaitGroup + goroutine!

GOOD: Each worker returns its own channel
workerChans := make(chan (<-chan Result))

go func() {
    for _, file := range files {
        // Each worker manages its own channel lifecycle
        workerChans <- processFile(file)  // Send the channel itself
    }
    close(workerChans)
}()

output := bridge(ctx, workerChans)  // Bridge handles the complexity
```

### Scenario 2: Recursive/Tree Processing

```go
// Crawl a website - each page spawns more crawlers
func crawl(url string) <-chan Page {
    out := make(chan Page)
    go func() {
        defer close(out)
        page := fetch(url)
        out <- page

        // Each link creates a NEW worker dynamically
        for _, link := range page.Links {
            subPages := crawl(link)  // Returns a channel
            // How do you merge these into one stream?
            // Bridge pattern!
        }
    }()
    return out
}
```

### Scenario 3: Real-World - API Rate Limiting

```go
// Process 1000 users, but API limits to 10 concurrent
users := []User{...1000 users...}

workerChans := make(chan (<-chan Result))

// Semaphore pattern - only 10 workers at a time
sem := make(chan struct{}, 10)

go func() {
    for _, user := range users {
        sem <- struct{}{}  // Acquire
        worker := processUser(user)  // Returns channel
        workerChans <- worker  // Send channel to bridge
        go func() {
            // Release after worker finishes
            <-worker
            <-sem
        }()
    }
    close(workerChans)
}()
```

```go
func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    users := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10} // 10 users to process

    workerChans := make(chan (<-chan Result))
    sem := make(chan struct{}, 3)  // Only 3 concurrent workers

    // Spawn workers with rate limiting
    go func() {
        defer close(workerChans)
        for _, userID := range users {
            sem <- struct{}{}  // Acquire semaphore (blocks if 3 workers running)

            workerChan := processUser(ctx, userID)  // Returns <-chan Result
            workerChans <- workerChan  // Send channel to bridge

            // Release semaphore when this worker's channel closes
            go func(ch <-chan Result) {
                for range ch {
                    // Drain the channel
                }
                <-sem  // Release
            }(workerChan)
        }
    }()

    // Bridge flattens all worker channels into one
    results := bridge(ctx, workerChans)

    // Consume all results
    for result := range results {
        fmt.Printf("Got result: %v\n", result)
    }
}

func processUser(ctx context.Context, id int) <-chan Result {
    out := make(chan Result)
    go func() {
        defer close(out)
        time.Sleep(500 * time.Millisecond)  // Simulate API call
        out <- Result{UserID: id, Data: "processed"}
    }()
    return out
}
```
