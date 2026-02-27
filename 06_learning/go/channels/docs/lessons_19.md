# Lesson 19: Pipelines with Error Handling

Concept: Build a multi-stage pipeline where each stage processes data and can fail. Errors propagate cleanly, and all stages respond to cancellation.

Your task:

Create a 3-stage pipeline: Generate → Transform → Save

1. Generator stage generator(ctx, nums []int) <-chan int
   - Sends numbers to output channel
   - Responds to context cancellation

2. Transform stage transform(ctx, in <-chan int) (<-chan int, <-chan error)
   - Doubles each number
   - Returns error if number is 6 (simulated failure)
   - Returns both output channel AND error channel

3. Save stage save(ctx, in <-chan int) <-chan error
   - Prints "saved: X" for each number
   - Returns error channel

4. In main():
   - Create context with cancel
   - Wire up: generator → transform → save
   - Monitor error channels from transform and save
   - Cancel pipeline on first error
   - Print which stage failed

Expected output:

```txt
saved: 2
saved: 4
saved: 8
saved: 10
transform error: number 6 is invalid
pipeline cancelled
```

Hints:

```go
Stage returning data + errors:
func transform(ctx context.Context, in <-chan int) (<-chan int, <-chan error) {
    out := make(chan int)
    errc := make(chan error, 1)  // buffered for one error

    go func() {
        defer close(out)
        defer close(errc)
        for n := range in {
            select {
            case <-ctx.Done():
                return
            default:
            }
            if n == 6 {
                errc <- fmt.Errorf("number %d is invalid", n)
                return
            }
            out <- n * 2
        }
    }()

    return out, errc
}

Merging error channels:
func mergeErrors(errs ...<-chan error) <-chan error {
    // combine multiple error channels into one
}
```

Key learning: Each pipeline stage manages its own goroutine and error reporting. The main function orchestrates and monitors for failures. This pattern scales to complex data processing systems.
