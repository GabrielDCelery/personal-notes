# Rate limiting

Concept: Use time.Ticker to enforce a maximum rate of operations (e.g., 3 requests per second). Unlike semaphores which limit concurrency, rate limiting controls throughput over time.

Your task:

1. Create a function makeRequest(id int) that:
   - Prints "request X sent at [time]"
   - Use time.Now().Format("15:04:05.000") for readable timestamps

2. In main():
   - Create a ticker that ticks every 200ms (5 requests/second max)
   - Process 10 requests, but only send one per tick
   - Stop the ticker when done
   - Print "all requests completed"

3. Observe the output - requests should be evenly spaced ~200ms apart

```txt
Expected output pattern:
request 0 sent at 14:30:01.000
request 1 sent at 14:30:01.200
request 2 sent at 14:30:01.400
request 3 sent at 14:30:01.600
...
all requests completed

```

Hints:

- Create ticker: ticker := time.NewTicker(200 \* time.Millisecond)
- Wait for tick: <-ticker.C
- Clean up: ticker.Stop()
- Ticker pattern:

```go
for id := range 10 {
    <-ticker.C  // wait for next tick
    makeRequest(id)
}
```

Key learning: Rate limiting protects external services from being overwhelmed and helps stay within API quotas. time.Ticker provides consistent intervals vs time.Sleep which can drift.
