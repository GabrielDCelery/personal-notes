# Lesson 13: HTTP Server Patterns

Go's `net/http` package is deliberately minimal — it gives you a handler interface and a mux, and largely gets out of the way. That minimalism means the patterns for middleware, timeouts, and graceful shutdown aren't enforced by the framework; they're conventions you either know or get wrong. These conventions are exactly what interviewers probe when they ask about production-ready Go HTTP servers.

## `http.Handler` and `http.HandlerFunc`

Everything in Go's HTTP server revolves around one interface:

```go
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}
```

Any type implementing `ServeHTTP` is a handler. `http.HandlerFunc` is a function type that implements `Handler`, letting you use plain functions as handlers:

```go
// Function type that satisfies http.Handler
type HandlerFunc func(ResponseWriter, *Request)

func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) {
    f(w, r)
}
```

```go
// ✓ Adapt a plain function to http.Handler
func helloHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "hello")
}

mux.Handle("/hello", http.HandlerFunc(helloHandler))
mux.HandleFunc("/hello", helloHandler)   // ✓ shorthand — does the conversion for you
```

The interface exists for composability: anything that wraps, decorates, or chains handlers works with any handler implementation.

---

## Middleware Pattern

Middleware wraps a handler, executing code before and/or after the inner handler. The idiomatic signature:

```go
func middlewareName(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // before: runs before the inner handler
        next.ServeHTTP(w, r)
        // after: runs after the inner handler
    })
}
```

### Example: Logging Middleware

```go
func logging(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
    })
}
```

### Example: Recovery Middleware

```go
func recovery(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("panic: %v\n%s", err, debug.Stack())
                http.Error(w, "internal server error", http.StatusInternalServerError)
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```

### Chaining Middleware

```go
// ✓ Wrap outermost to innermost — execution order is outside-in
handler := logging(recovery(authMiddleware(mux)))
// Request flow: logging → recovery → auth → mux → handler
// Response flow (after): auth ← recovery ← logging

// ✓ Helper to apply a list of middleware
func chain(h http.Handler, middleware ...func(http.Handler) http.Handler) http.Handler {
    for i := len(middleware) - 1; i >= 0; i-- {
        h = middleware[i](h)
    }
    return h
}

handler := chain(mux, logging, recovery, requestID)
```

### Passing Values Through Middleware

```go
type contextKey string

const requestIDKey contextKey = "requestID"

func requestID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        id := uuid.New().String()
        ctx := context.WithValue(r.Context(), requestIDKey, id)
        w.Header().Set("X-Request-ID", id)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Downstream handler retrieves it
func handler(w http.ResponseWriter, r *http.Request) {
    id, _ := r.Context().Value(requestIDKey).(string)
    log.Printf("handling request %s", id)
}
```

**Gotcha**: Always use an unexported type for context keys — never a plain `string`. Two packages using `"requestID"` as a string key would collide. An unexported type is unique to your package.

---

## `http.ServeMux`

### Pattern Matching Rules

`http.ServeMux` matches on path patterns:

```go
mux := http.NewServeMux()
mux.HandleFunc("/",        rootHandler)    // matches everything not matched elsewhere
mux.HandleFunc("/api/",    apiHandler)     // trailing slash = subtree: matches /api/, /api/users, etc.
mux.HandleFunc("/api/v1",  v1Handler)      // exact match only (no trailing slash)
```

- **Longest prefix wins**: `/api/users` is matched by `/api/` and `/`, but `/api/` wins
- **Trailing slash = subtree**: `/api/` matches `/api/anything`
- **No trailing slash = exact**: `/api` matches only `/api`
- **Host-specific patterns**: `example.com/api` matches only requests to that host

**Gotcha**: The default `http.DefaultServeMux` is a package-level variable. Third-party packages that call `http.Handle` register on it silently. Never use `http.ListenAndServe("addr", nil)` in production — pass your own mux.

```go
// ❌ Uses DefaultServeMux — any imported package can register handlers on it
http.ListenAndServe(":8080", nil)

// ✓ Explicit mux
mux := http.NewServeMux()
http.ListenAndServe(":8080", mux)
```

### Go 1.22 Enhanced Routing

Go 1.22 added method and wildcard support to `ServeMux`:

```go
mux.HandleFunc("GET /users/{id}", getUserHandler)      // method + path parameter
mux.HandleFunc("POST /users", createUserHandler)
mux.HandleFunc("DELETE /users/{id}", deleteUserHandler)

// Retrieve path parameter
func getUserHandler(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")   // "42" from /users/42
    // ...
}

// {id...} matches the rest of the path including slashes
mux.HandleFunc("GET /files/{path...}", fileHandler)
```

---

## Server Timeouts

An `http.Server` without timeouts is vulnerable to slow clients and resource exhaustion. Never use `http.ListenAndServe` directly — it creates a server with all timeouts at zero (disabled).

```go
// ❌ No timeouts — a slow client can hold a goroutine and connection forever
http.ListenAndServe(":8080", mux)

// ✓ Always configure an explicit server
srv := &http.Server{
    Addr:              ":8080",
    Handler:           mux,
    ReadTimeout:       5 * time.Second,
    ReadHeaderTimeout: 2 * time.Second,
    WriteTimeout:      10 * time.Second,
    IdleTimeout:       120 * time.Second,
}
srv.ListenAndServe()
```

### The Four Timeouts

```
Client connects
    │
    ├── ReadHeaderTimeout: time to read request headers
    │
    ├── ReadTimeout: time to read entire request (headers + body)
    │         starts from accept, ends when body is fully read
    │
    ├── [handler executes]
    │
    ├── WriteTimeout: time from request accepted to response fully written
    │         Note: includes ReadTimeout duration on the server side
    │
Client disconnects / next request
    │
    └── IdleTimeout: how long a keep-alive connection can sit idle
                     waiting for the next request
```

| Timeout             | Protects Against                              | Default      |
| ------------------- | --------------------------------------------- | ------------ |
| `ReadHeaderTimeout` | Slowloris attack (headers sent 1 byte/minute) | 0 (disabled) |
| `ReadTimeout`       | Slow request body uploads                     | 0 (disabled) |
| `WriteTimeout`      | Slow response reads by client                 | 0 (disabled) |
| `IdleTimeout`       | Idle keep-alive connections accumulating      | 0 (disabled) |

**Slowloris**: An attacker opens many connections and sends HTTP headers extremely slowly, never completing them. Without `ReadHeaderTimeout`, each connection holds a goroutine indefinitely.

**`WriteTimeout` includes handler time**: If your handler takes 8 seconds and `WriteTimeout` is 10 seconds, you have 2 seconds to write the response. Size `WriteTimeout` accordingly for long-running handlers, or use per-request context deadlines instead.

---

## Graceful Shutdown

`srv.Shutdown(ctx)` stops accepting new connections, waits for active requests to complete, then returns. `srv.Close()` terminates immediately.

```go
srv := &http.Server{Addr: ":8080", Handler: mux}

// Run server in background goroutine
go func() {
    if err := srv.ListenAndServe(); err != http.ErrServerClosed {
        log.Fatalf("server error: %v", err)   // ErrServerClosed is expected on Shutdown
    }
}()

// Wait for interrupt signal
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

log.Println("shutting down...")

// Give in-flight requests up to 30 seconds to complete
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

if err := srv.Shutdown(ctx); err != nil {
    log.Fatalf("forced shutdown: %v", err)   // context deadline exceeded = forced
}

log.Println("shutdown complete")
```

**Key points**:

- `signal.Notify` requires a **buffered channel** (size ≥ 1) — signals sent to an unbuffered channel are dropped if nothing is receiving at that exact moment
- `http.ErrServerClosed` is the expected error after `Shutdown` — don't treat it as fatal
- The shutdown context timeout is your deadline for in-flight requests; if exceeded, `Shutdown` returns `context.DeadlineExceeded` and remaining connections are closed forcibly

---

## Request Context

Every `*http.Request` carries a context that is cancelled when the client disconnects:

```go
func slowHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    result, err := db.QueryContext(ctx, "SELECT ...")   // ✓ cancelled if client leaves
    if err != nil {
        if ctx.Err() != nil {
            // client disconnected — no point writing a response
            return
        }
        http.Error(w, "db error", http.StatusInternalServerError)
        return
    }
    // ...
}
```

**Always propagate `r.Context()`** to database queries, HTTP client calls, and any blocking operations. This enables request cancellation to propagate through the entire call stack automatically.

---

## Response Writing Gotchas

### `WriteHeader` Must Come Before `Write`

```go
// ❌ Wrong order — header is ignored, default 200 is sent with the body
w.Write([]byte("error"))
w.WriteHeader(http.StatusBadRequest)

// ✓ Set status before writing body
w.WriteHeader(http.StatusBadRequest)
w.Write([]byte("error"))

// ✓ http.Error sets status and writes body correctly
http.Error(w, "bad request", http.StatusBadRequest)
```

### Set Headers Before `WriteHeader`

```go
// ❌ Header set after WriteHeader is ignored — headers are already sent
w.WriteHeader(http.StatusOK)
w.Header().Set("Content-Type", "application/json")

// ✓ Set headers before WriteHeader or before the first Write
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(result)   // first Write implicitly calls WriteHeader(200)
```

### Double Write

Once `WriteHeader` is called (explicitly or by the first `Write`), calling it again logs a warning and is ignored:

```go
// ❌ Common mistake in error handling — WriteHeader called twice
func handler(w http.ResponseWriter, r *http.Request) {
    result, err := doWork()
    if err != nil {
        http.Error(w, "error", 500)   // calls WriteHeader(500)
        // missing return:
    }
    json.NewEncoder(w).Encode(result)  // calls WriteHeader(200) — superfluous, logs warning
}

// ✓ Always return after writing an error response
func handler(w http.ResponseWriter, r *http.Request) {
    result, err := doWork()
    if err != nil {
        http.Error(w, "error", 500)
        return   // ✓
    }
    json.NewEncoder(w).Encode(result)
}
```

### Capturing the Status Code in Middleware

`http.ResponseWriter` doesn't expose the status code after writing. Wrap it:

```go
type statusRecorder struct {
    http.ResponseWriter
    status int
}

func (r *statusRecorder) WriteHeader(status int) {
    r.status = status
    r.ResponseWriter.WriteHeader(status)
}

func logging(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
        next.ServeHTTP(rec, r)
        log.Printf("%s %s %d", r.Method, r.URL.Path, rec.status)
    })
}
```

---

## Hands-On Exercise 1: Middleware Chain

Build a middleware stack for a JSON API that:

1. Recovers from panics and returns `500`
2. Injects a unique `X-Request-ID` header into both the request context and the response
3. Logs method, path, status code, and duration after each request

Wire them together so logging is outermost (sees the final status code).

<details>
<summary>Solution</summary>

```go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "runtime/debug"
    "time"

    "github.com/google/uuid"
)

type contextKey string
const requestIDKey contextKey = "requestID"

// statusRecorder captures the written status code
type statusRecorder struct {
    http.ResponseWriter
    status int
}
func (r *statusRecorder) WriteHeader(status int) {
    r.status = status
    r.ResponseWriter.WriteHeader(status)
}

// recovery catches panics
func recovery(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("panic: %v\n%s", err, debug.Stack())
                http.Error(w, "internal server error", http.StatusInternalServerError)
            }
        }()
        next.ServeHTTP(w, r)
    })
}

// requestID injects a unique ID
func requestID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        id := uuid.New().String()
        w.Header().Set("X-Request-ID", id)
        ctx := context.WithValue(r.Context(), requestIDKey, id)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// logging is outermost — wraps a statusRecorder to capture status
func logging(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
        next.ServeHTTP(rec, r)
        log.Printf("%s %s %d %v", r.Method, r.URL.Path, rec.status, time.Since(start))
    })
}

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("GET /hello", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "hello")
    })

    // ✓ logging outermost so it sees final status; recovery inside logging
    handler := logging(requestID(recovery(mux)))

    srv := &http.Server{
        Addr:              ":8080",
        Handler:           handler,
        ReadHeaderTimeout: 2 * time.Second,
        ReadTimeout:       5 * time.Second,
        WriteTimeout:      10 * time.Second,
        IdleTimeout:       120 * time.Second,
    }
    log.Fatal(srv.ListenAndServe())
}
```

**Why logging is outermost**: It wraps the `statusRecorder` around the entire chain, so it can read `rec.status` only after the inner handlers have finished — including after recovery has potentially written a 500.

</details>

## Hands-On Exercise 2: Graceful Shutdown

The following server shuts down immediately on SIGINT, killing in-flight requests. Fix it to give requests up to 10 seconds to complete.

```go
func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
        time.Sleep(5 * time.Second)
        fmt.Fprintln(w, "done")
    })

    srv := &http.Server{Addr: ":8080", Handler: mux}

    c := make(chan os.Signal)
    signal.Notify(c, os.Interrupt)
    go func() {
        <-c
        srv.Close()
    }()

    srv.ListenAndServe()
}
```

<details>
<summary>Solution</summary>

**Issues**:

1. ❌ `signal.Notify` requires a buffered channel — signals sent to an unbuffered channel are dropped if not immediately received
2. ❌ `srv.Close()` terminates connections immediately — in-flight requests are killed
3. ❌ `ListenAndServe` error not handled — `ErrServerClosed` after shutdown would go unlogged

**Fixed**:

```go
func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
        time.Sleep(5 * time.Second)
        fmt.Fprintln(w, "done")
    })

    srv := &http.Server{
        Addr:              ":8080",
        Handler:           mux,
        ReadHeaderTimeout: 2 * time.Second,
        WriteTimeout:      15 * time.Second,
        IdleTimeout:       120 * time.Second,
    }

    go func() {
        if err := srv.ListenAndServe(); err != http.ErrServerClosed {
            log.Fatalf("server: %v", err)   // ✓ ErrServerClosed is expected
        }
    }()

    quit := make(chan os.Signal, 1)           // ✓ buffered
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("shutting down...")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {  // ✓ graceful
        log.Fatalf("shutdown: %v", err)
    }
    log.Println("done")
}
```

</details>

---

## Interview Questions

### Q1: What are the four HTTP server timeouts and what does each protect against?

Interviewers ask this to distinguish developers who have run Go servers in production from those who've only written them locally. Misconfigured timeouts are one of the most common sources of production incidents in Go services.

<details>
<summary>Answer</summary>

- **`ReadHeaderTimeout`**: Time allowed to read request headers. Protects against Slowloris attacks — clients that open connections and send headers one byte at a time to hold goroutines indefinitely. This should always be set; it's the most critical timeout for public-facing servers.

- **`ReadTimeout`**: Time allowed to read the entire request, including the body. Starts when the connection is accepted, ends when the request body is fully read. Protects against clients that send large bodies slowly.

- **`WriteTimeout`**: Time from connection accept to response fully written, including the handler's execution time. Protects against handlers that run too long and clients that receive responses slowly. Note: this includes `ReadTimeout` — if `ReadTimeout` is 5s and `WriteTimeout` is 10s, handlers effectively get 5s.

- **`IdleTimeout`**: How long a keep-alive connection can sit idle between requests. Prevents idle connections from accumulating and consuming file descriptors. Falls back to `ReadTimeout` if not set.

All four default to zero (disabled). The minimum safe configuration for a public server:

```go
srv := &http.Server{
    ReadHeaderTimeout: 2 * time.Second,
    ReadTimeout:       5 * time.Second,
    WriteTimeout:      10 * time.Second,
    IdleTimeout:       120 * time.Second,
}
```

For endpoints with long-running handlers (reports, exports), use per-request context deadlines rather than inflating `WriteTimeout` for all endpoints.

</details>

### Q2: Why does the middleware pattern use `func(http.Handler) http.Handler` rather than a method or struct?

A design question that tests understanding of Go's composition model and the `http.Handler` interface.

<details>
<summary>Answer</summary>

The `func(http.Handler) http.Handler` signature is idiomatic because:

**It composes cleanly**: Functions that take and return the same type can be chained indefinitely. Any function with this signature works with any handler — there's no inheritance or embedding required.

**It fits Go's interface model**: `http.Handler` is satisfied by any type with `ServeHTTP`. A closure returned by a middleware function satisfies it via `http.HandlerFunc`. No struct definition needed.

**It separates concerns**: Each middleware knows nothing about other middleware. It only knows about `next`. You can reorder, add, or remove middleware without changing any middleware implementation.

**Comparison to alternatives**:

- A `Middleware` struct with a chain method would couple middleware to a framework-specific type
- A variadic `ServeHTTP` with pre/post hooks would require middleware to know about the framework
- The function approach works with any mux, any framework, any deployment — it's just Go

The pattern is also trivially testable: pass a handler that records what it receives, wrap it with the middleware under test, fire a test request, inspect the result.

</details>

### Q3: What happens when `http.Server.Shutdown` is called, and how does it differ from `Close`?

Tests understanding of graceful shutdown — a real production concern for zero-downtime deployments.

<details>
<summary>Answer</summary>

**`Shutdown(ctx)`**:

1. Stops accepting new connections immediately (closes the listener)
2. Closes idle keep-alive connections (those not actively serving a request)
3. Waits for active connections to finish their current request, then closes them
4. Returns when all active connections are done, or when `ctx` is cancelled (whichever comes first)
5. Returns `context.DeadlineExceeded` or `context.Canceled` if the context expires before all connections finish

**`Close()`**:

- Immediately closes all connections regardless of their state — active requests are killed mid-flight
- Returns immediately without waiting

**The correct pattern**:

```go
// ctx gives in-flight requests a deadline to complete
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
err := srv.Shutdown(ctx)
```

**After calling `Shutdown`**, `ListenAndServe` returns `http.ErrServerClosed`. This must not be treated as an error:

```go
if err := srv.ListenAndServe(); err != http.ErrServerClosed {
    log.Fatal(err)
}
```

**Important**: `Shutdown` does not shut down WebSocket connections or hijacked connections (connections taken over via `http.Hijacker`). Those must be tracked and closed separately.

</details>

### Q4: Why should you never use a plain `string` as a context key, and what's the correct approach?

A code-quality question that appears in Go code reviews constantly. It reveals whether the candidate understands package-level namespacing in Go.

<details>
<summary>Answer</summary>

The `context.WithValue` / `context.Value` API uses `interface{}` for keys. If two packages both store values under the string key `"userID"`, they collide — one package's value overwrites the other's, causing silent bugs.

```go
// ❌ String keys collide across packages
ctx = context.WithValue(ctx, "userID", "alice")       // package auth
ctx = context.WithValue(ctx, "userID", "session-123") // package session — overwrites!
```

The fix is to use an **unexported type** as the key type. An unexported type is unique to its package — no other package can construct a value of that type, so collisions are impossible:

```go
// ✓ Unexported type — unique to this package
type contextKey string

const (
    userIDKey    contextKey = "userID"
    requestIDKey contextKey = "requestID"
)

ctx = context.WithValue(ctx, userIDKey, "alice")

// Retrieve with type assertion
userID, ok := ctx.Value(userIDKey).(string)
```

Even if another package defines `type contextKey string` and uses `contextKey("userID")`, its type is distinct from yours — they don't collide.

**Why not use `int` or `iota`?** You can — any unexported type works. The convention in the standard library is often an unexported struct type:

```go
type contextKey struct{}   // zero-size struct, no values needed
var userIDKey = contextKey{}
```

Using a struct avoids allocating string data and makes it impossible to accidentally construct the key from an external string. Either approach (unexported string type or unexported struct) is acceptable; the critical point is that the type is unexported.

</details>

---

## Key Takeaways

1. **`http.Handler` is the unit of composition**: any type with `ServeHTTP` is a handler — middleware, muxes, static file servers all compose through this interface.
2. **Middleware signature**: `func(http.Handler) http.Handler` — take a handler, return a handler. Order matters: outermost middleware runs first on the way in, last on the way out.
3. **Never use `http.ListenAndServe` directly**: it uses `DefaultServeMux` (global, pollutable) and zero timeouts. Always construct `http.Server` explicitly.
4. **Set all four timeouts**: `ReadHeaderTimeout` (Slowloris), `ReadTimeout` (slow bodies), `WriteTimeout` (slow handlers/clients), `IdleTimeout` (idle keep-alive).
5. **Graceful shutdown**: `Shutdown(ctx)` drains in-flight requests; `Close()` kills them. Use a buffered channel with `signal.Notify`.
6. **`WriteHeader` before `Write`**: headers must be set before the first `Write` call (which implicitly sends a 200). Always `return` after error responses.
7. **Context keys must be unexported types**: plain string keys collide across packages, causing silent value overwrites.
8. **Propagate `r.Context()`**: pass it to every blocking call (DB, HTTP client, file I/O) so client disconnects cancel work automatically.
9. **Go 1.22 routing**: built-in method and path parameter support reduces the need for third-party routers for simple APIs.
10. **Capture status in middleware**: wrap `ResponseWriter` with a recorder struct to read the status code after the handler writes it.

## Next Steps

In [Lesson 14: Reflection & Struct Tags](lesson-14-reflection-and-struct-tags.md), you'll learn:

- How `reflect.Type` and `reflect.Value` work and what each gives you
- Traversing struct fields and reading struct tags at runtime
- Why settability requires a pointer and how to use `Value.Elem()`
- When to use reflection vs generics
- The performance cost of reflection and how libraries mitigate it
