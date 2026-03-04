# chi

```sh
go get github.com/go-chi/chi/v5
```

> Lightweight, idiomatic HTTP router. Compatible with `net/http` middleware.

## Setup

```go
import "github.com/go-chi/chi/v5"

r := chi.NewRouter()

r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
})

http.ListenAndServe(":8080", r)
```

## Routes

```go
r.Get("/items", listItems)
r.Post("/items", createItem)
r.Put("/items/{id}", updateItem)
r.Delete("/items/{id}", deleteItem)

// Route param
func getItem(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    // ...
}
```

## Middleware

```go
import "github.com/go-chi/chi/v5/middleware"

r := chi.NewRouter()

r.Use(middleware.Logger)
r.Use(middleware.Recoverer)
r.Use(middleware.RequestID)
r.Use(middleware.RealIP)
r.Use(middleware.Timeout(60 * time.Second))
```

## Custom middleware

```go
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
}

r.Use(authMiddleware)
```

## Route groups

```go
r.Route("/api/v1", func(r chi.Router) {
    r.Use(authMiddleware)

    r.Route("/users", func(r chi.Router) {
        r.Get("/", listUsers)
        r.Post("/", createUser)
        r.Get("/{id}", getUser)
    })

    r.Route("/items", func(r chi.Router) {
        r.Get("/", listItems)
        r.Post("/", createItem)
    })
})
```

## Apply middleware to subset of routes

```go
r.Group(func(r chi.Router) {
    r.Use(authMiddleware)
    r.Get("/profile", getProfile)
    r.Put("/profile", updateProfile)
})

// Public routes — no auth
r.Get("/login", login)
r.Post("/register", register)
```

## Context values

```go
// Set in middleware
ctx := context.WithValue(r.Context(), userKey, user)
next.ServeHTTP(w, r.WithContext(ctx))

// Read in handler
user := r.Context().Value(userKey).(User)
```

## URL pattern matching

```go
r.Get("/files/*", serveFiles)           // wildcard
r.Get("/users/{id:[0-9]+}", getUser)    // regex constraint
```
