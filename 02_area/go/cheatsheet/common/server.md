# Go HTTP Server Setup

## Quick Reference

| Use case            | Method                           |
| ------------------- | -------------------------------- |
| Basic server        | `http.ListenAndServe`            |
| Custom mux          | `http.NewServeMux`               |
| With middleware     | wrap handler func                |
| Graceful shutdown   | `server.Shutdown` with context   |
| HTTPS               | `http.ListenAndServeTLS`         |
| Read request body   | `json.NewDecoder(r.Body).Decode` |
| Write JSON response | `json.NewEncoder(w).Encode`      |

## Server Setups

### 1. Basic server (simplest)

```go
// uses the DefaultServeMux multiplexer (router)
http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
})

http.ListenAndServe(":8080", nil)
```

### 2. Custom mux

```go
// mux is a multiplexer (router) that watches incoming requests and calls handlers
mux := http.NewServeMux()

mux.HandleFunc("GET /items", listItems)
mux.HandleFunc("POST /items", createItem)
mux.HandleFunc("GET /items/{id}", getItem)

http.ListenAndServe(":8080", mux)
```

### 3. Custom server with timeouts (recommended for production)

```go
srv := &http.Server{
    Addr:         ":8080",
    Handler:      mux,
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  60 * time.Second,
}

if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
    log.Fatal(err)
}
```

### 4. Graceful shutdown

```go
srv := &http.Server{Addr: ":8080", Handler: mux}

go func() {
    if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatal(err)
    }
}()

quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

if err := srv.Shutdown(ctx); err != nil {
    log.Fatal(err)
}
```

### 5. Middleware

```go
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("%s %s", r.Method, r.URL.Path)
        next.ServeHTTP(w, r)
    })
}

mux := http.NewServeMux()
mux.HandleFunc("GET /items", listItems)

http.ListenAndServe(":8080", loggingMiddleware(mux))
```

### 6. Read JSON request body

```go
type CreateItemRequest struct {
    Name string `json:"name"`
}

func createItem(w http.ResponseWriter, r *http.Request) {
    var req CreateItemRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid body", http.StatusBadRequest)
        return
    }
    // use req.Name
}
```

### 7. Write JSON response

```go
func listItems(w http.ResponseWriter, r *http.Request) {
    items := []Item{{ID: 1, Name: "Alice"}}

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(items)
}
```

### 8. HTTPS

```go
http.ListenAndServeTLS(":443", "cert.pem", "key.pem", mux)
```
