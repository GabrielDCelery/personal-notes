# Here's the core concept for Lesson 13: errgroup:

## Task

Import

```go
import "golang.org/x/sync/errgroup"
```

(You'll need go get golang.org/x/sync/errgroup)

Basic usage:

```go
g, ctx := errgroup.WithContext(context.Background())

g.Go(func() error {
// do work
  return nil // or return err
})

g.Go(func() error {
// do other work
  return nil
})

if err := g.Wait(); err != nil {
// handle first error
}
```

## Key features

- g.Wait() blocks until all goroutines finish
- Returns the first error encountered (others are discarded)
- When using WithContext, the context cancels immediately when any goroutine returns an error
- Other goroutines should check ctx.Done() to exit early

Why it's useful:

- Replaces manual WaitGroup + error channel boilerplate
- Built-in cancellation propagation on failure
- Clean "fail fast" semantics

## Exercise idea

Create 3 workers that "fetch" data (simulated). One should fail after 200ms. Others should detect the cancellation via context and stop early.
