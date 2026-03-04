# Go Logging (log/slog)

> Standard library structured logger since Go 1.21. Use `log/slog` for new projects, zap for perf-critical paths.

## Quick Reference

| Use case            | Method                          |
| ------------------- | ------------------------------- |
| Basic log           | `slog.Info("msg")`              |
| With fields         | `slog.Info("msg", "key", val)`  |
| JSON output         | `slog.NewJSONHandler`           |
| Text output         | `slog.NewTextHandler`           |
| Child logger        | `logger.With("key", val)`       |
| Group fields        | `logger.WithGroup("name")`      |
| Set global logger   | `slog.SetDefault(logger)`       |
| Logger from context | `slog.Default()` (no ctx-based) |

## Basics

### 1. Default logger

```go
slog.Debug("starting up")              // hidden by default (level Info)
slog.Info("server started", "port", 8080)
slog.Warn("slow query", "duration", "2.3s")
slog.Error("failed to connect", "err", err)
```

### 2. JSON handler (production)

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
slog.SetDefault(logger)

slog.Info("user created", "id", 42, "name", "Alice")
// {"time":"...","level":"INFO","msg":"user created","id":42,"name":"Alice"}
```

### 3. Text handler (development)

```go
logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
slog.SetDefault(logger)

slog.Info("user created", "id", 42, "name", "Alice")
// time=... level=INFO msg="user created" id=42 name=Alice
```

## Configuration

### 4. Set minimum log level

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
}))
```

### 5. Add source location (file:line)

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    AddSource: true,
}))
```

## Child Loggers

### 6. Add fields to all logs

```go
logger := slog.Default().With("service", "auth", "version", "1.2.0")

logger.Info("started")  // includes service=auth version=1.2.0
logger.Info("stopped")  // includes service=auth version=1.2.0
```

### 7. Group related fields

```go
logger := slog.Default().WithGroup("request")
logger.Info("handled", "method", "GET", "path", "/users")
// JSON: {"msg":"handled","request":{"method":"GET","path":"/users"}}
```

## Typed Attributes (faster)

### 8. Use slog.Attr for performance

```go
slog.LogAttrs(context.Background(), slog.LevelInfo, "user created",
    slog.Int("id", 42),
    slog.String("name", "Alice"),
    slog.Duration("took", time.Since(start)),
    slog.Any("tags", []string{"admin", "active"}),
)
```

## Context Pattern

### 9. Pass logger through context (manual)

```go
type ctxKey struct{}

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
    return context.WithValue(ctx, ctxKey{}, logger)
}

func FromContext(ctx context.Context) *slog.Logger {
    if logger, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok {
        return logger
    }
    return slog.Default()
}
```

### 10. Use in middleware

```go
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        logger := slog.Default().With(
            "request_id", r.Header.Get("X-Request-ID"),
            "method", r.Method,
            "path", r.URL.Path,
        )
        ctx := WithLogger(r.Context(), logger)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

## slog vs log vs zap

|                  | log           | slog           | zap           |
| ---------------- | ------------- | -------------- | ------------- |
| Structured       | no            | yes            | yes           |
| Standard library | yes           | yes (Go 1.21+) | no            |
| Performance      | basic         | fast           | fastest       |
| JSON output      | no            | yes            | yes           |
| When to use      | quick scripts | new projects   | perf-critical |
