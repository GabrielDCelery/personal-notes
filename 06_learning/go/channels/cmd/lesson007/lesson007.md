# Lesson 7: Worker Pool

## Concept

A worker pool distributes tasks across multiple goroutines, limiting concurrency and efficiently processing work. This is one of the most practical channel patterns.

## Task

1. Create a jobs channel of type int (buffered, capacity 10)
2. Create a results channel of type int (buffered, capacity 10)
3. Start 3 worker goroutines, each running a loop that:
   - Receives a job from jobs
   - "Processes" it by doubling the number (simulating work)
   - Sends the result to results
   - Exits when jobs is closed

4. Send 5 jobs (numbers 1-5) to the jobs channel, then close it
5. Collect and print all 5 results

## Hints

- Workers should use for job := range jobs to automatically exit when channel closes
- Give each worker an ID and print which worker processed which job:
  worker 2 processing job 3
- Send all jobs before collecting results (buffered channels allow this)

Expected output (order may vary):

```sh
worker 1 processing job 1
worker 2 processing job 2
worker 3 processing job 3
worker 1 processing job 4
worker 2 processing job 5
result: 2
result: 4
result: 6
result: 8
result: 10
```

## Key learning

Worker pools let you control parallelism - e.g., limit to 10 concurrent database connections regardless of how many tasks exist.
