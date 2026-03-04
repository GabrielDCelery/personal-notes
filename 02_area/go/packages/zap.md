# Zap

```sh
go get go.uber.org/zap
```

> High-performance structured logger from Uber. Two APIs: `zap.Logger` (fast, verbose) and `zap.SugaredLogger` (slower, convenient).

## Setup

```go
// Production (JSON output)
logger, err := zap.NewProduction()
if err != nil {
    log.Fatal(err)
}
defer logger.Sync()

// Development (human-readable output)
logger, _ := zap.NewDevelopment()
defer logger.Sync()
```

## Logger (typed fields — fastest)

```go
logger.Info("user created",
    zap.Int("id", 1),
    zap.String("name", "Alice"),
    zap.Duration("took", time.Since(start)),
)

logger.Error("failed to connect",
    zap.String("host", host),
    zap.Error(err),
)

logger.Warn("retry attempt",
    zap.Int("attempt", 3),
    zap.Int("max", 5),
)
```

## SugaredLogger (loosely typed — easier)

```go
sugar := logger.Sugar()

sugar.Infow("user created",
    "id", 1,
    "name", "Alice",
)

sugar.Infof("connected to %s", host)

// Convert back if needed
logger = sugar.Desugar()
```

## Global logger

```go
// Set global logger once at startup
zap.ReplaceGlobals(logger)

// Use anywhere
zap.L().Info("message", zap.String("key", "val"))
zap.S().Infow("message", "key", "val") // sugared global
```

## Add fields to all logs (child logger)

```go
requestLogger := logger.With(
    zap.String("request_id", requestID),
    zap.String("user_id", userID),
)

// All logs from requestLogger include request_id and user_id
requestLogger.Info("handler started")
requestLogger.Info("handler finished")
```

## Custom config

```go
cfg := zap.Config{
    Level:       zap.NewAtomicLevelAt(zap.InfoLevel),
    Development: false,
    Encoding:    "json",
    EncoderConfig: zapcore.EncoderConfig{
        TimeKey:        "time",
        LevelKey:       "level",
        MessageKey:     "msg",
        CallerKey:      "caller",
        EncodeLevel:    zapcore.LowercaseLevelEncoder,
        EncodeTime:     zapcore.ISO8601TimeEncoder,
        EncodeCaller:   zapcore.ShortCallerEncoder,
    },
    OutputPaths:      []string{"stdout"},
    ErrorOutputPaths: []string{"stderr"},
}

logger, err := cfg.Build()
```

## zap vs slog

|                  | zap                               | slog            |
| ---------------- | --------------------------------- | --------------- |
| Performance      | fastest                           | fast            |
| API              | typed fields                      | key-value pairs |
| Standard library | no                                | yes (Go 1.21+)  |
| Ecosystem        | mature                            | growing         |
| When to use      | perf-critical, existing codebases | new projects    |
