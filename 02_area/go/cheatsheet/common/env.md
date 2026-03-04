# Go Environment Variables

## Why

- **LookupEnv vs Getenv** — Getenv returns "" for both "not set" and "set to empty". LookupEnv distinguishes the two with an ok bool. Use LookupEnv when the distinction matters.
- **Config struct pattern** — Centralizes all env var reads into one place at startup. Fails fast if required vars are missing. The rest of the app works with typed fields, not raw strings.
- **No .env in production** — dotenv files are for local dev. In production, env vars come from the platform (ECS task def, K8s configmap, systemd service file). Don't ship .env files.
- **Parse early, use typed values** — Convert strings to int/bool/duration at startup. The rest of the code shouldn't call strconv or deal with parse errors.

## Quick Reference

| Use case       | Method                    |
| -------------- | ------------------------- |
| Get var        | `os.Getenv("KEY")`        |
| Get with check | `os.LookupEnv("KEY")`     |
| Set var        | `os.Setenv("KEY", "val")` |
| Unset var      | `os.Unsetenv("KEY")`      |
| All vars       | `os.Environ()`            |

## Basics

### 1. Get environment variable

```go
port := os.Getenv("PORT")  // returns "" if not set
```

### 2. Check if set vs empty

```go
val, ok := os.LookupEnv("PORT")
if !ok {
    // not set at all
} else if val == "" {
    // set but empty
}
```

### 3. Get with default

```go
func getEnv(key, fallback string) string {
    if val, ok := os.LookupEnv(key); ok {
        return val
    }
    return fallback
}

port := getEnv("PORT", "8080")
```

## Config Pattern

### 4. Load config from env

```go
type Config struct {
    Port        string
    DatabaseURL string
    Debug       bool
}

func LoadConfig() (*Config, error) {
    port, ok := os.LookupEnv("PORT")
    if !ok {
        port = "8080"
    }

    dbURL, ok := os.LookupEnv("DATABASE_URL")
    if !ok {
        return nil, fmt.Errorf("DATABASE_URL is required")
    }

    return &Config{
        Port:        port,
        DatabaseURL: dbURL,
        Debug:       os.Getenv("DEBUG") == "true",
    }, nil
}
```

### 5. Required env helper

```go
func mustGetEnv(key string) string {
    val, ok := os.LookupEnv(key)
    if !ok {
        log.Fatalf("required env var %s is not set", key)
    }
    return val
}
```

### 6. Parse non-string types

```go
timeout, err := strconv.Atoi(os.Getenv("TIMEOUT_SECONDS"))
if err != nil {
    timeout = 30
}

maxRetries, err := strconv.ParseInt(os.Getenv("MAX_RETRIES"), 10, 64)
enabled, err := strconv.ParseBool(os.Getenv("FEATURE_ENABLED"))
rate, err := strconv.ParseFloat(os.Getenv("RATE_LIMIT"), 64)
```
