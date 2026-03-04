# Go Sync Primitives

## Why

- **WaitGroup** — The simplest way to wait for goroutines to finish. Add before launching, Done inside the goroutine, Wait in the caller. Without it, main exits before goroutines complete.
- **Mutex vs channels** — Use a mutex to protect shared state (a map, a counter). Use channels to communicate between goroutines. "Share memory by communicating" doesn't mean never use mutexes.
- **RWMutex** — When reads vastly outnumber writes, RWMutex allows concurrent reads while still ensuring exclusive writes. Regular Mutex blocks all access, even concurrent reads.
- **sync.Once** — Guarantees a function runs exactly once, even under concurrent access. Perfect for lazy initialization of singletons (DB connections, config loading).
- **sync.Map** — Optimized for two patterns: keys written once and read many times, or disjoint key sets per goroutine. For everything else, a regular map with a Mutex is simpler and faster.
- **Atomic vs Mutex** — Atomics are faster for simple operations (counters, flags) because they avoid lock contention. Use Mutex when you need to protect multiple fields or complex logic.

## Quick Reference

| Use case                      | Method                        |
| ----------------------------- | ----------------------------- |
| Wait for goroutines to finish | `sync.WaitGroup`              |
| Protect shared data           | `sync.Mutex` / `sync.RWMutex` |
| One-time initialization       | `sync.Once`                   |
| Safe concurrent map           | `sync.Map`                    |
| Atomic counter                | `atomic.AddInt64`             |

## WaitGroup

### 1. Wait for multiple goroutines

```go
var wg sync.WaitGroup

for _, item := range items {
    wg.Add(1)
    go func(v Item) {
        defer wg.Done()
        process(v)
    }(item)
}

wg.Wait() // blocks until all Done()
```

### 2. Collect results with WaitGroup

```go
var (
    wg      sync.WaitGroup
    mu      sync.Mutex
    results []string
)

for _, item := range items {
    wg.Add(1)
    go func(v Item) {
        defer wg.Done()
        result := process(v)
        mu.Lock()
        results = append(results, result)
        mu.Unlock()
    }(item)
}

wg.Wait()
```

## Mutex

### 3. Protect shared state

```go
type SafeCounter struct {
    mu    sync.Mutex
    count int
}

func (c *SafeCounter) Inc() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count++
}

func (c *SafeCounter) Value() int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.count
}
```

### 4. RWMutex (multiple readers, one writer)

```go
type SafeCache struct {
    mu    sync.RWMutex
    store map[string]string
}

func (c *SafeCache) Get(key string) (string, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    v, ok := c.store[key]
    return v, ok
}

func (c *SafeCache) Set(key, val string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.store[key] = val
}
```

## Once

### 5. One-time initialization (singleton)

```go
var (
    instance *DB
    once     sync.Once
)

func GetDB() *DB {
    once.Do(func() {
        instance = connectDB()
    })
    return instance
}
```

## sync.Map

### 6. Concurrent map (high read, low write)

```go
var m sync.Map

// Store
m.Store("key", "value")

// Load
val, ok := m.Load("key")

// Delete
m.Delete("key")

// Iterate
m.Range(func(k, v any) bool {
    fmt.Println(k, v)
    return true // return false to stop
})
```

## Atomic

### 7. Atomic counter (no mutex needed)

```go
import "sync/atomic"

var count int64

atomic.AddInt64(&count, 1)
atomic.AddInt64(&count, -1)

val := atomic.LoadInt64(&count)
```

### 8. Atomic flag (e.g. running state)

```go
var running atomic.Bool

running.Store(true)

if running.Load() {
    // ...
}

running.Store(false)
```
