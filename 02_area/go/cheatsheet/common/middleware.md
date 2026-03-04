# Go Middleware & Context

## Why

- **Middleware is just function composition** — A function that takes a handler and returns a handler. No framework needed — it's the same pattern in stdlib, chi, and gin.
- **Order matters** — Middleware wraps right-to-left but executes left-to-right. `logging(auth(mux))` means logging runs first (before), then auth, then handler.
- **Unexported context key types** — Prevents other packages from accidentally overwriting your context values. A `type contextKey string` in your package is invisible to everyone else.
- **r.WithContext** — HTTP request is the carrier for context in web apps. Middleware enriches the context and passes it forward via `r.WithContext(ctx)`.

## Quick Reference

| Use case                | Method                        |
| ----------------------- | ----------------------------- |
| Basic middleware        | wrap `http.Handler`           |
| Chain middleware        | nest wrappers or use a helper |
| Store value in context  | `context.WithValue`           |
| Read value from context | `ctx.Value(key)`              |
| Pass context in request | `r.WithContext(ctx)`          |
| Typed context keys      | custom unexported type        |

## Middleware

### 1. Basic middleware pattern

```go
func myMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // before
        next.ServeHTTP(w, r)
        // after
    })
}
```

### 2. Chaining middleware

```go
// Apply right to left: logging runs first, then auth
handler := loggingMiddleware(authMiddleware(mux))

http.ListenAndServe(":8080", handler)
```

### 3. Chain helper (cleaner for many middleware)

```go
func chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
    for i := len(middlewares) - 1; i >= 0; i-- {
        h = middlewares[i](h)
    }
    return h
}

// Usage
handler := chain(mux, loggingMiddleware, authMiddleware, corsMiddleware)
```

### 4. Logging middleware

```go
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
    })
}
```

### 5. Auth middleware (store user in context)

```go
type contextKey string

const userKey contextKey = "user"

func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }

        user := getUserFromToken(token) // your logic

        ctx := context.WithValue(r.Context(), userKey, user)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

## Context

### 6. Store and retrieve a value

```go
type contextKey string

const requestIDKey contextKey = "requestID"

// Store
ctx := context.WithValue(r.Context(), requestIDKey, "abc-123")
r = r.WithContext(ctx)

// Retrieve
id, ok := r.Context().Value(requestIDKey).(string)
if !ok {
    // key not set or wrong type
}
```

### 7. Read context value in a handler

```go
func getItem(w http.ResponseWriter, r *http.Request) {
    user, ok := r.Context().Value(userKey).(User)
    if !ok {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }
    // use user
}
```

### 8. Context with timeout / cancellation

```go
func myHandler(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()

    result, err := doSlowThing(ctx)
    if err != nil {
        http.Error(w, "timeout or error", http.StatusInternalServerError)
        return
    }
    // use result
}
```

### 9. Pass context to downstream calls

```go
func doSlowThing(ctx context.Context) (string, error) {
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://example.com/api", nil)
    if err != nil {
        return "", err
    }
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return "", err // returns if ctx cancelled
    }
    defer resp.Body.Close()
    // ...
}
```
