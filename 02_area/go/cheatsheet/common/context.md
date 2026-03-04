# Go Context

## Quick Reference

| Use case            | Method                               |
| ------------------- | ------------------------------------ |
| Create root context | `context.Background()`               |
| With timeout        | `context.WithTimeout(ctx, duration)` |
| With deadline       | `context.WithDeadline(ctx, time)`    |
| With cancel         | `context.WithCancel(ctx)`            |
| With value          | `context.WithValue(ctx, key, val)`   |
| Check cancellation  | `ctx.Done()` channel / `ctx.Err()`   |

## Basics

### 1. Create and pass context

```go
func main() {
    ctx := context.Background()
    result, err := fetchData(ctx)
}

// Always pass context as first parameter
func fetchData(ctx context.Context) (string, error) {
    // ...
}
```

### 2. WithTimeout — cancel after duration

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel() // always defer cancel to release resources

result, err := db.QueryContext(ctx, "SELECT ...")
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        // query took longer than 5s
    }
    return err
}
```

### 3. WithCancel — manual cancellation

```go
ctx, cancel := context.WithCancel(context.Background())

go func() {
    // cancel when signal received
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    <-sigCh
    cancel()
}()

if err := server.Serve(ctx); err != nil {
    log.Println("server stopped:", err)
}
```

### 4. Check if context is done

```go
func longOperation(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err() // context.Canceled or context.DeadlineExceeded
        default:
            // do work
        }
    }
}
```

## Context Values

### 5. Store and retrieve values (type-safe pattern)

```go
// Define unexported key type to avoid collisions
type contextKey string

const requestIDKey contextKey = "requestID"

// Set value
func WithRequestID(ctx context.Context, id string) context.Context {
    return context.WithValue(ctx, requestIDKey, id)
}

// Get value
func RequestID(ctx context.Context) (string, bool) {
    id, ok := ctx.Value(requestIDKey).(string)
    return id, ok
}
```

### 6. Use in middleware

```go
func RequestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        id := r.Header.Get("X-Request-ID")
        if id == "" {
            id = uuid.New().String()
        }
        ctx := WithRequestID(r.Context(), id)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func handler(w http.ResponseWriter, r *http.Request) {
    id, _ := RequestID(r.Context())
    log.Printf("[%s] handling request", id)
}
```

## Patterns

### 7. HTTP client with context

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
if err != nil {
    return err
}

resp, err := http.DefaultClient.Do(req)
```

### 8. Database queries with context

```go
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()

row := db.QueryRowContext(ctx, "SELECT name FROM users WHERE id = $1", id)

var name string
if err := row.Scan(&name); err != nil {
    return err
}
```

### 9. Propagate context through goroutines

```go
func processItems(ctx context.Context, items []Item) error {
    g, ctx := errgroup.WithContext(ctx)

    for _, item := range items {
        g.Go(func() error {
            // Each goroutine respects the parent context
            return processItem(ctx, item)
        })
    }

    return g.Wait()
}
```

### 10. Graceful shutdown pattern

```go
func main() {
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    srv := &http.Server{Addr: ":8080", Handler: router}

    go func() {
        if err := srv.ListenAndServe(); err != http.ErrServerClosed {
            log.Fatal(err)
        }
    }()

    <-ctx.Done()
    log.Println("shutting down...")

    shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    srv.Shutdown(shutdownCtx)
}
```

## Rules of Thumb

- Always pass `context.Context` as the first parameter
- Never store context in a struct
- Always `defer cancel()` when using `WithTimeout` or `WithCancel`
- Use `context.TODO()` as a placeholder when unsure which context to use
- Keep context values to request-scoped data only (request ID, auth token) — not config or dependencies
