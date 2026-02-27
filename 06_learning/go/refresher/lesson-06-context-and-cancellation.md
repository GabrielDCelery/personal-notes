# Lesson 06: Context & Cancellation

Every non-trivial Go service uses `context.Context`. It's the standard mechanism for propagating deadlines, cancellation signals, and request-scoped values through your call stack. But it's also one of the most commonly misused packages — especially `context.WithValue`. Understanding the contract, the patterns, and the anti-patterns is table stakes for senior Go interviews.

## The `context.Context` Contract

`context.Context` is an interface:

```go
type Context interface {
    Deadline() (deadline time.Time, ok bool)
    Done() <-chan struct{}
    Err() error
    Value(key any) any
}
```

| Method       | Returns                                                                     |
| ------------ | --------------------------------------------------------------------------- |
| `Deadline()` | The time when context will be cancelled, and whether a deadline is set      |
| `Done()`     | A channel that's closed when context is cancelled or deadline exceeded      |
| `Err()`      | `nil` if not done; `context.Canceled` or `context.DeadlineExceeded` if done |
| `Value(key)` | The value associated with key, or nil                                       |

**The core contract**: A `Context` is immutable. You don't modify it — you derive a new one with additional constraints.

## Root Contexts

Every context tree starts with a root:

```go
// Use in main or tests - never cancels, no deadline
ctx := context.Background()

// Use as a placeholder when context is required but not yet available
// Functionally identical to Background - semantic difference only
ctx := context.TODO()
```

**Never pass `nil`** where a `Context` is expected — use `context.Background()` or `context.TODO()`.

## Deriving Contexts

All derived contexts inherit from their parent. Cancelling a parent cancels all children.

```
Background ──► WithCancel ──► WithTimeout ──► WithValue
   (root)        (cancel)      (deadline)       (kv)
                    │
                    └──► WithValue (sibling)
```

### `WithCancel`

Creates a child context with a cancel function:

```go
ctx, cancel := context.WithCancel(parent)
defer cancel() // always call cancel to release resources

go func() {
    // Do work, check ctx.Done()
    select {
    case <-ctx.Done():
        return // cancelled
    case result <- compute():
        // done
    }
}()

// When you're done with the work:
cancel() // propagates cancellation to all children
```

**Always call `cancel()`**: Even if the context will be cancelled through a parent. Calling cancel releases resources held by the context (goroutines watching it). Use `defer cancel()` immediately after creating the context.

### `WithTimeout` and `WithDeadline`

```go
// WithTimeout: cancel after duration
ctx, cancel := context.WithTimeout(parent, 5*time.Second)
defer cancel()

// WithDeadline: cancel at absolute time
deadline := time.Now().Add(5 * time.Second)
ctx, cancel := context.WithDeadline(parent, deadline)
defer cancel()
```

Both return a cancel function that you must call. The context auto-cancels at the deadline, but `cancel()` must still be called to release resources if you finish early.

```go
func fetchWithTimeout(url string) ([]byte, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return nil, err
    }
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err // err wraps context.DeadlineExceeded if timed out
    }
    defer resp.Body.Close()
    return io.ReadAll(resp.Body)
}
```

### Checking Cancellation

```go
// Pattern 1: select for non-blocking check
select {
case <-ctx.Done():
    return ctx.Err() // context.Canceled or context.DeadlineExceeded
default:
    // continue working
}

// Pattern 2: blocking wait for cancellation or result
select {
case result := <-resultCh:
    return result, nil
case <-ctx.Done():
    return nil, ctx.Err()
}
```

## Context Propagation Patterns

Context flows **down** the call stack, never up. Every function that does I/O or could block should accept a `context.Context` as its first parameter.

```go
// ✓ Context as first parameter, named ctx by convention
func (s *Server) GetUser(ctx context.Context, id int) (*User, error) {
    return s.db.QueryUser(ctx, id) // propagate down
}

func (db *DB) QueryUser(ctx context.Context, id int) (*User, error) {
    row := db.pool.QueryRowContext(ctx, "SELECT * FROM users WHERE id=$1", id)
    // ...
}
```

### HTTP Server: Using the Request Context

```go
func (s *Server) handler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context() // provided by net/http; cancelled when client disconnects

    user, err := s.db.GetUser(ctx, userID)
    if err != nil {
        if errors.Is(err, context.Canceled) {
            // Client disconnected - no need to respond
            return
        }
        http.Error(w, "internal error", 500)
        return
    }
    json.NewEncoder(w).Encode(user)
}
```

The request context is automatically cancelled when:

- The client closes the connection
- The HTTP server shuts down
- The request handler returns

### Fan-out with Context

```go
func fetchMultiple(ctx context.Context, ids []int) ([]*User, error) {
    type result struct {
        user *User
        err  error
    }

    results := make(chan result, len(ids))

    for _, id := range ids {
        go func(id int) {
            user, err := fetchUser(ctx, id) // context propagated to each goroutine
            results <- result{user, err}
        }(id)
    }

    users := make([]*User, 0, len(ids))
    for range ids {
        r := <-results
        if r.err != nil {
            return nil, r.err
        }
        users = append(users, r.user)
    }
    return users, nil
}
```

## `context.WithValue`: Patterns and Anti-Patterns

`WithValue` stores a key-value pair in the context, retrievable by any function in the call chain.

```go
ctx = context.WithValue(ctx, keyType("requestID"), "abc-123")
// Later in the call chain:
id := ctx.Value(keyType("requestID")).(string)
```

### The Key Type Rule

**Always use unexported, package-specific types for keys** — never strings or built-in types. This prevents key collisions between packages.

```go
// ❌ String key - collides with any package using the same string
ctx = context.WithValue(ctx, "userID", 42)

// ❌ Built-in type key - same collision problem
ctx = context.WithValue(ctx, "user", user)

// ✓ Unexported type key - scoped to this package only
type contextKey int
const (
    userIDKey contextKey = iota
    requestIDKey
)
ctx = context.WithValue(ctx, userIDKey, 42)
```

### What Belongs in Context

```go
// ✓ Request-scoped data that crosses API boundaries:
// - Request/trace IDs (for logging correlation)
// - Auth tokens / user identity (propagated through middleware)
// - Locale / language preference

// ❌ NOT for:
// - Function parameters (use explicit args - they're clear and type-safe)
// - Optional configuration
// - Database handles, connections, loggers (use dependency injection)
// - Mutable state
```

### The Classic Middleware Pattern

```go
// Middleware sets request ID in context
func requestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        id := generateRequestID()
        ctx := context.WithValue(r.Context(), requestIDKey, id)
        w.Header().Set("X-Request-ID", id)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Handler retrieves it for logging
func (s *Server) handler(w http.ResponseWriter, r *http.Request) {
    requestID, _ := r.Context().Value(requestIDKey).(string)
    s.logger.Info("handling request", "requestID", requestID)
}
```

## Context Anti-Patterns

### Anti-Pattern 1: Storing context in a struct

```go
// ❌ Don't store context in a struct
type Server struct {
    ctx context.Context // wrong
}

// ✓ Pass context as a parameter to methods
func (s *Server) Handle(ctx context.Context, req *Request) (*Response, error)
```

The Go team explicitly states: "Do not store Contexts inside a struct type; instead, pass a Context explicitly to each function that needs it."

### Anti-Pattern 2: Using `context.Background()` in a handler

```go
// ❌ Loses cancellation - DB query runs even after client disconnects
func handler(w http.ResponseWriter, r *http.Request) {
    ctx := context.Background() // creates new root, disconnected from request
    user, _ := db.QueryUser(ctx, id)
}

// ✓ Propagate the request context
func handler(w http.ResponseWriter, r *http.Request) {
    user, _ := db.QueryUser(r.Context(), id)
}
```

### Anti-Pattern 3: Using `WithValue` for function arguments

```go
// ❌ Passing function parameters through context is opaque and untyped
ctx = context.WithValue(ctx, pageSizeKey, 50)
results := db.List(ctx) // What parameters does this take? Unknown without reading source.

// ✓ Explicit function parameters
results := db.List(ctx, ListOptions{PageSize: 50})
```

### Anti-Pattern 4: Ignoring `ctx.Err()`

```go
// ❌ You process the result even if the context was cancelled
result, _ := someOperation(ctx)
return result

// ✓ Check if context is still valid before doing expensive post-processing
result, err := someOperation(ctx)
if err != nil {
    return nil, err
}
if ctx.Err() != nil { // check even if operation succeeded
    return nil, ctx.Err()
}
return result, nil
```

## Hands-On Exercise 1: Context-Aware Retry

Implement a retry function that respects context cancellation.

```go
// Requirements:
// 1. Retry fn up to maxAttempts times
// 2. Stop immediately if ctx is cancelled (don't start the next attempt)
// 3. Wait `delay` between attempts
// 4. Return the last error if all attempts fail
// 5. Return context error if cancelled during wait
```

<details>
<summary>Solution</summary>

```go
package main

import (
    "context"
    "fmt"
    "time"
)

func RetryWithContext(ctx context.Context, maxAttempts int, delay time.Duration, fn func(ctx context.Context) error) error {
    var lastErr error

    for attempt := 0; attempt < maxAttempts; attempt++ {
        // Check context before starting attempt
        if ctx.Err() != nil {
            return ctx.Err()
        }

        lastErr = fn(ctx)
        if lastErr == nil {
            return nil
        }

        if attempt == maxAttempts-1 {
            break // no wait after last attempt
        }

        // Wait with context awareness
        select {
        case <-time.After(delay):
            // continue to next attempt
        case <-ctx.Done():
            return ctx.Err()
        }
    }

    return fmt.Errorf("after %d attempts: %w", maxAttempts, lastErr)
}
```

</details>

## Hands-On Exercise 2: Pipeline with Cancellation

Implement a three-stage pipeline that cancels all stages when any stage fails.

```go
// Pipeline: generate -> transform -> collect
// Requirements:
// 1. generate(ctx) produces integers 1..n on a channel
// 2. transform(ctx, in) squares each number
// 3. collect(ctx, in) returns the slice of results
// 4. If ctx is cancelled at any stage, all stages stop cleanly
// 5. No goroutine leaks
```

<details>
<summary>Solution</summary>

```go
package main

import (
    "context"
    "fmt"
)

func generate(ctx context.Context, n int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for i := 1; i <= n; i++ {
            select {
            case out <- i:
            case <-ctx.Done():
                return // stop if context cancelled
            }
        }
    }()
    return out
}

func transform(ctx context.Context, in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for v := range in {
            select {
            case out <- v * v:
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}

func collect(ctx context.Context, in <-chan int) ([]int, error) {
    var results []int
    for {
        select {
        case v, ok := <-in:
            if !ok {
                return results, nil // channel closed, done
            }
            results = append(results, v)
        case <-ctx.Done():
            return nil, ctx.Err()
        }
    }
}

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    nums := generate(ctx, 10)
    squares := transform(ctx, nums)
    results, err := collect(ctx, squares)
    if err != nil {
        fmt.Println("error:", err)
        return
    }
    fmt.Println(results) // [1 4 9 16 25 36 49 64 81 100]
}
```

</details>

## Interview Questions

### Q1: What is `context.Context` and why does it exist?

A foundational question — the answer reveals how deeply you understand Go's concurrency model.

<details>
<summary>Answer</summary>

`context.Context` is an interface that carries three things across API and goroutine boundaries:

1. **Cancellation signal**: When the context is cancelled (by a parent or timeout), all goroutines watching `ctx.Done()` can stop work and clean up.
2. **Deadline**: An absolute time after which the context auto-cancels.
3. **Request-scoped values**: A small amount of metadata that crosses API boundaries (request IDs, auth tokens).

**Why it exists**: Go's concurrency model has no built-in way to stop goroutines from the outside — you can't kill a goroutine. Context provides the standard pattern: goroutines check `ctx.Done()` voluntarily and exit when cancelled. This enables:

- Client disconnection detection in HTTP servers (the request context is cancelled)
- Service-level timeouts that propagate through all downstream calls
- Graceful shutdown — a root context cancellation propagates to every goroutine in the chain

Before `context`, developers invented their own cancellation channels and "stop" signals. Context standardized this so every library (database drivers, HTTP clients, gRPC) uses the same interface.

</details>

### Q2: What are the rules for using `context.WithValue`?

A practical question that reveals whether you know the anti-patterns.

<details>
<summary>Answer</summary>

`context.WithValue` attaches a key-value pair to a context. The rules:

**What to store**: Only request-scoped data that crosses API/process boundaries — request IDs, trace IDs, authentication tokens, locale. Anything that's the same for the entire lifetime of a request.

**What NOT to store**:

- Function parameters (use explicit args)
- Database handles or connections (use dependency injection)
- Mutable state
- Application configuration
- Loggers (controversial — some apps do this, but it hides dependencies)

**Key type rule**: Always use unexported, package-local types for keys to prevent collisions:

```go
type ctxKey int
const userKey ctxKey = 0
```

Never use string keys — any package using `ctx.Value("user")` would conflict.

**Type assertion on retrieval**: `ctx.Value(key)` returns `any` — always do a safe type assertion:

```go
if v, ok := ctx.Value(userKey).(*User); ok {
    // use v
}
```

**The Go team's rule**: "Use context Values only for request-scoped data that transits processes and APIs, not for passing optional parameters to functions."

</details>

### Q3: Why must you always call the cancel function returned by `WithCancel`, `WithTimeout`, and `WithDeadline`?

Tests attention to resource management details.

<details>
<summary>Answer</summary>

The cancel function releases the goroutines and timers associated with the context. Specifically:

- `WithCancel`: registers the child context with the parent. The parent holds a reference to the child until either the parent is cancelled or the child's cancel is called. Without calling cancel, the child is never unregistered — if the parent lives a long time, these registrations accumulate.

- `WithTimeout`/`WithDeadline`: starts an internal timer goroutine. Without calling cancel, the timer goroutine runs until the deadline passes, even if your work finished long before.

In high-request-rate services, leaked contexts accumulate goroutines and timers, consuming memory and eventually causing performance degradation.

The Go idiom:

```go
ctx, cancel := context.WithTimeout(parent, 5*time.Second)
defer cancel() // always, immediately after creation
```

`defer cancel()` ensures cleanup happens regardless of how the function exits (normal return, early return, panic). It's idempotent — calling it multiple times is safe.

</details>

### Q4: How does the HTTP server's request context interact with client disconnection?

A practical question testing knowledge of how context works in the real world.

<details>
<summary>Answer</summary>

`net/http` creates a context for each incoming request. This context is cancelled when:

1. The client closes the connection (browser navigates away, mobile app backgrounded)
2. The server calls `http.Server.Shutdown()` or `Close()`
3. The request handler returns

Handlers access it via `r.Context()`. By propagating this context to all downstream calls (DB queries, HTTP calls to other services), those operations automatically cancel when the client disconnects:

```go
func handler(w http.ResponseWriter, r *http.Request) {
    user, err := db.GetUser(r.Context(), userID) // cancels if client disconnects
    if err != nil {
        if errors.Is(err, context.Canceled) {
            return // client gone, no need to respond
        }
        http.Error(w, err.Error(), 500)
        return
    }
    json.NewEncoder(w).Encode(user)
}
```

**Why this matters**: Without context propagation, a slow database query keeps running even after the client has gone. In high-traffic services, this wastes DB connections and CPU. With proper context propagation, the DB driver cancels the query (if the driver supports it — `database/sql` does via `QueryContext`), freeing resources immediately.

`errors.Is(err, context.Canceled)` distinguishes "client left" from a real error, so you don't log a client disconnect as a server error.

</details>

## Key Takeaways

1. **Context is immutable**: Derive new contexts (`WithCancel`, `WithTimeout`) rather than modifying.
2. **Always `defer cancel()`**: Immediately after creating a derived context — prevents goroutine and timer leaks.
3. **Context as first parameter**: The convention is `ctx context.Context` as the first parameter of any function that does I/O.
4. **`r.Context()` in HTTP handlers**: Propagate the request context to all downstream calls.
5. **`WithValue` for request metadata only**: Request IDs, auth tokens — not function parameters or configuration.
6. **Unexported key types**: Prevent key collisions between packages.
7. **Check `ctx.Err()`**: Distinguish `context.Canceled` from `context.DeadlineExceeded` in error handling.
8. **Never store context in structs**: Pass it explicitly to each method that needs it.
9. **Context cancels cascade**: Cancelling a parent cancels all children automatically.

## Next Steps

In [Lesson 07: Generics](lesson-07-generics.md), you'll learn:

- How type parameters and constraints work in Go 1.18+
- The `~T` underlying type constraint and when you need it
- When generics improve your code and when interfaces are still the better choice
- Real-world use cases: collections, functional helpers, ordered types
