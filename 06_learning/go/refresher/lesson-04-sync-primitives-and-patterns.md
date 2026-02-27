# Lesson 04: Sync Primitives & Patterns

Go's mantra is "share memory by communicating" — but channels aren't always the right tool. When you have shared state that multiple goroutines need to read and write, sync primitives are often clearer, faster, and less error-prone. Knowing when to reach for a mutex vs a channel, and understanding the subtle rules of `sync.Mutex` and friends, separates experienced Go developers from those who cargo-cult channel-everything.

## Mutex: Mutual Exclusion

A `sync.Mutex` ensures that only one goroutine can execute a section of code at a time.

```go
type SafeCounter struct {
    mu    sync.Mutex
    count int
}

func (c *SafeCounter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock() // always defer unlock
    c.count++
}

func (c *SafeCounter) Value() int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.count
}
```

### Rules for Mutex Usage

1. **Always defer `Unlock`**: It runs even if the function panics, preventing deadlocks.
2. **Never copy a mutex**: `sync.Mutex` has internal state. Copying creates two independent locks.
3. **Use pointer receivers**: Methods that lock a mutex must use pointer receivers, otherwise a copy is made.

```go
// ❌ Don't copy
m := sync.Mutex{}
m2 := m            // bug: m2 is a separate mutex, not protecting the same resource

// ❌ Don't pass by value
func bad(m sync.Mutex) { // copies the mutex
    m.Lock()
    defer m.Unlock()
}

// ✓ Pass by pointer
func good(m *sync.Mutex) {
    m.Lock()
    defer m.Unlock()
}
```

**`go vet`** catches mutex copies — run it in your CI pipeline.

## RWMutex: Concurrent Reads, Exclusive Writes

`sync.RWMutex` allows many goroutines to **read** simultaneously, but only one to **write** (and writes block all reads).

```go
type Cache struct {
    mu    sync.RWMutex
    items map[string]string
}

func (c *Cache) Get(key string) (string, bool) {
    c.mu.RLock()         // shared read lock - multiple goroutines can hold this
    defer c.mu.RUnlock()
    v, ok := c.items[key]
    return v, ok
}

func (c *Cache) Set(key, value string) {
    c.mu.Lock()          // exclusive write lock - blocks all readers and writers
    defer c.mu.Unlock()
    c.items[key] = value
}
```

**When `RWMutex` wins**: Read-heavy workloads (e.g., in-memory caches). If reads dominate (99% reads, 1% writes), `RWMutex` allows all those reads to proceed in parallel.

**When `RWMutex` loses**: Write-heavy workloads. Each write must wait for all reads to finish, and vice versa — the overhead exceeds a plain mutex. Also: `RWMutex` is more expensive than `Mutex` per operation.

```
Read-heavy cache: use RWMutex
Counters, queues, write-heavy: use Mutex
```

## WaitGroup: Waiting for Goroutines

`sync.WaitGroup` lets you wait for a group of goroutines to finish.

```go
var wg sync.WaitGroup

for i := 0; i < 5; i++ {
    wg.Add(1)           // ✓ Add BEFORE starting goroutine
    go func(n int) {
        defer wg.Done() // signals completion
        process(n)
    }(i)
}

wg.Wait() // blocks until all Done() calls match Add() calls
```

### Common WaitGroup Mistakes

```go
// ❌ Add inside goroutine - race condition
for i := 0; i < 5; i++ {
    go func(n int) {
        wg.Add(1)      // might happen after wg.Wait() in some schedules
        defer wg.Done()
        process(n)
    }(i)
}

// ❌ Forgetting to call Done - deadlock
for i := 0; i < 5; i++ {
    wg.Add(1)
    go func(n int) {
        process(n)
        // forgot wg.Done()
    }(i)
}
wg.Wait() // hangs forever
```

**Rule**: Call `wg.Add(1)` in the launching goroutine, immediately before `go`. Call `wg.Done()` with `defer` as the first line of the launched goroutine.

## Once: One-Time Initialization

`sync.Once` ensures a function runs exactly once, even if called from multiple goroutines simultaneously.

```go
type Singleton struct {
    db *sql.DB
}

var (
    instance *Singleton
    once     sync.Once
)

func GetInstance() *Singleton {
    once.Do(func() {
        db, _ := sql.Open("postgres", dsn)
        instance = &Singleton{db: db}
    })
    return instance
}
```

**Properties**:

- The function passed to `Do` runs exactly once across all goroutines
- Subsequent calls to `Do` are no-ops but **wait** if the first is still running
- If the first call panics, `Once` considers the function to have completed — subsequent calls are still no-ops

```go
// ❌ Gotcha: panic in once.Do means the initialization "succeeds" as far as Once is concerned
once.Do(func() {
    panic("init failed")  // future calls to once.Do will NOT retry
})
```

For retry-on-failure, `sync.Once` is not the right tool — use explicit synchronization.

## Atomic Operations

`sync/atomic` provides lock-free operations on integer and pointer types. Faster than a mutex for simple counters.

```go
import "sync/atomic"

var count int64

atomic.AddInt64(&count, 1)          // increment
atomic.AddInt64(&count, -1)         // decrement
n := atomic.LoadInt64(&count)       // read
atomic.StoreInt64(&count, 42)       // write
ok := atomic.CompareAndSwapInt64(&count, old, new) // CAS
```

Since Go 1.19, you can use the typed `atomic.Int64`, `atomic.Bool` etc.:

```go
var count atomic.Int64

count.Add(1)
n := count.Load()
count.Store(42)
count.Swap(100)            // returns old value
count.CompareAndSwap(42, 0) // CAS
```

**When to use atomics vs mutexes**:

|                             | Atomic                        | Mutex    |
| --------------------------- | ----------------------------- | -------- |
| **Simple counter**          | ✓ faster                      | Overkill |
| **Read/write single value** | ✓                             | Works    |
| **Compound operations**     | ❌ not atomic across multiple | ✓        |
| **Complex data structures** | ❌                            | ✓        |

```go
// ❌ This is NOT atomic even with atomic load/store:
n := atomic.LoadInt64(&count)
atomic.StoreInt64(&count, n+1) // another goroutine may have changed count between load and store
// Use atomic.AddInt64 for atomic increment
```

## sync.Map

`sync.Map` is a concurrent map built into the standard library. It's optimized for specific access patterns.

```go
var m sync.Map

m.Store("key", "value")

v, ok := m.Load("key")
if ok {
    fmt.Println(v.(string))
}

m.LoadOrStore("key", "default") // returns existing or stores default

m.Delete("key")

m.Range(func(k, v any) bool {
    fmt.Println(k, v)
    return true // return false to stop iteration
})
```

**When `sync.Map` outperforms `map + RWMutex`**:

- Keys are written once and read many times (no updates after initial write)
- Different goroutines operate on disjoint sets of keys

**When to prefer `map + sync.Mutex/RWMutex`**:

- You need range with mutation (sync.Map's Range doesn't guarantee consistency with concurrent writes)
- Write-heavy or update-heavy workloads
- You need `len()` (sync.Map has no Len method)

## Channels vs Mutexes: Decision Framework

The canonical Go question. Both achieve synchronization — the choice is about expressing intent.

| Use Channels When                         | Use Mutexes When                              |
| ----------------------------------------- | --------------------------------------------- |
| Passing data ownership between goroutines | Protecting shared state                       |
| Signalling events (done, cancel, tick)    | Simple counters or caches                     |
| Pipeline / fan-out / fan-in patterns      | Struct fields accessed by multiple goroutines |
| Work queues                               | Read-heavy data structures (use RWMutex)      |
| Coordinating goroutine lifecycle          | Avoiding the overhead of channel allocation   |

```go
// ✓ Channel: transferring ownership of data
func producer() <-chan Item {
    ch := make(chan Item)
    go func() {
        for _, item := range source {
            ch <- item  // ownership transferred to consumer
        }
        close(ch)
    }()
    return ch
}

// ✓ Mutex: protecting a shared cache
type Cache struct {
    mu    sync.RWMutex
    items map[string]Item
}
```

**The real rule**: Use whichever makes the code clearer. If you're protecting a struct field, reach for a mutex. If you're coordinating goroutine lifecycle (start/stop/done), use channels or `context`.

## Common Race Conditions

### Closure Capture Bug

```go
// ❌ All goroutines capture the same `i`
for i := 0; i < 5; i++ {
    go func() {
        fmt.Println(i) // i is captured by reference, may be 5 for all
    }()
}

// ✓ Pass i as argument
for i := 0; i < 5; i++ {
    go func(n int) {
        fmt.Println(n) // n is a copy
    }(i)
}

// ✓ Also works: create new variable in loop body (Go 1.22+ fixes for range vars)
for i := 0; i < 5; i++ {
    i := i // shadows loop variable with a new copy
    go func() {
        fmt.Println(i)
    }()
}
```

### Map Concurrent Access

```go
// ❌ Concurrent map read/write - runtime panic (in Go 1.6+)
m := make(map[string]int)
go func() { m["a"] = 1 }()
go func() { _ = m["a"] }()

// ✓ Protect with mutex or use sync.Map
```

Go detects concurrent map access at runtime and panics with `concurrent map read and map write`. This is not a race condition that silently corrupts — it crashes loudly.

### Detecting Races

```sh
go test -race ./...
go run -race main.go
go build -race -o myapp
```

The race detector inserts instrumentation to detect conflicting concurrent accesses. It has ~5-10x performance overhead — use in CI, not production.

## Hands-On Exercise 1: Thread-Safe Cache

Implement a TTL (time-to-live) cache where entries expire after a duration.

```go
// Requirements:
// 1. Get(key string) (value interface{}, ok bool) - returns false if expired
// 2. Set(key string, value interface{}, ttl time.Duration)
// 3. Delete(key string)
// 4. Thread-safe for concurrent use
// 5. Expired entries should not be returned but don't need to be eagerly deleted
```

<details>
<summary>Solution</summary>

```go
package main

import (
    "sync"
    "time"
)

type entry struct {
    value     interface{}
    expiresAt time.Time
}

type TTLCache struct {
    mu    sync.RWMutex
    items map[string]entry
}

func NewTTLCache() *TTLCache {
    return &TTLCache{
        items: make(map[string]entry),
    }
}

func (c *TTLCache) Set(key string, value interface{}, ttl time.Duration) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.items[key] = entry{
        value:     value,
        expiresAt: time.Now().Add(ttl),
    }
}

func (c *TTLCache) Get(key string) (interface{}, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    e, ok := c.items[key]
    if !ok || time.Now().After(e.expiresAt) {
        return nil, false
    }
    return e.value, true
}

func (c *TTLCache) Delete(key string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    delete(c.items, key)
}
```

</details>

## Hands-On Exercise 2: Rate Limiter

Implement a token bucket rate limiter using atomic operations and time.

```go
// Requirements:
// 1. Allow(n int) bool - returns true if n tokens are available and consumes them
// 2. Tokens replenish at `rate` per second up to `capacity`
// 3. Thread-safe
// 4. Use atomic operations where appropriate
```

<details>
<summary>Solution</summary>

```go
package main

import (
    "sync"
    "time"
)

type RateLimiter struct {
    mu       sync.Mutex
    tokens   float64
    capacity float64
    rate     float64 // tokens per second
    lastTime time.Time
}

func NewRateLimiter(capacity, rate float64) *RateLimiter {
    return &RateLimiter{
        tokens:   capacity,
        capacity: capacity,
        rate:     rate,
        lastTime: time.Now(),
    }
}

func (r *RateLimiter) Allow(n float64) bool {
    r.mu.Lock()
    defer r.mu.Unlock()

    now := time.Now()
    elapsed := now.Sub(r.lastTime).Seconds()
    r.lastTime = now

    // Replenish tokens
    r.tokens += elapsed * r.rate
    if r.tokens > r.capacity {
        r.tokens = r.capacity
    }

    if r.tokens < n {
        return false // not enough tokens
    }
    r.tokens -= n
    return true
}
```

</details>

## Interview Questions

### Q1: What's the difference between `sync.Mutex` and `sync.RWMutex`, and when do you choose each?

A classic question that tests understanding of read/write access patterns.

<details>
<summary>Answer</summary>

- `sync.Mutex`: Exclusive lock — one goroutine holds it at a time, for both reads and writes.
- `sync.RWMutex`: Shared/exclusive lock — any number of goroutines can hold a read lock simultaneously; a write lock is exclusive and blocks all readers.

**Choose `RWMutex`** when:

- Your workload is read-dominated (e.g., 90%+ reads)
- Concurrent reads are safe (reads don't modify state)
- Reads are frequent enough that contention on a plain mutex is measurable

**Choose `Mutex`** when:

- Write-heavy workload (RWMutex overhead isn't worth it)
- Operations are short (mutex overhead is negligible)
- Simplicity matters (Mutex is simpler to reason about)

**Gotcha**: `RWMutex` is more complex and more expensive per operation. In benchmarks, a `Mutex` often beats `RWMutex` unless reads are significantly more frequent than writes. Always benchmark before assuming `RWMutex` is faster.

</details>

### Q2: How do you safely initialize a singleton in Go?

Tests knowledge of `sync.Once` and the alternatives.

<details>
<summary>Answer</summary>

Three approaches:

**1. `sync.Once`** (idiomatic for lazy initialization):

```go
var (
    instance *DB
    once     sync.Once
)
func GetDB() *DB {
    once.Do(func() { instance = newDB() })
    return instance
}
```

**2. Package-level `init()`** (eager initialization, simpler):

```go
var instance = newDB() // initialized at program start
```

**3. `sync/atomic` pointer swap** (advanced, for updateable singletons):

```go
var instance atomic.Pointer[DB]
func GetDB() *DB {
    if db := instance.Load(); db != nil {
        return db
    }
    db := newDB()
    instance.CompareAndSwap(nil, db)
    return instance.Load()
}
```

**Don't use**: double-checked locking with a plain bool and mutex — it's not thread-safe without `sync.Once` or atomics because of the Go memory model.

`sync.Once` is the standard answer for lazy initialization. If startup cost is trivial, package-level initialization is simpler.

</details>

### Q3: When should you use channels vs mutexes for synchronization?

A philosophical question that reveals whether you've internalized Go's design philosophy.

<details>
<summary>Answer</summary>

The Go team's advice: "Use whichever is more natural." The practical breakdown:

**Use channels when**:

- You're transferring ownership of data from one goroutine to another
- You need to signal events (done, cancel, timer ticks)
- You're building pipelines or fan-out/fan-in patterns
- You want backpressure (bounded work queues)

**Use mutexes when**:

- You're protecting a shared struct's fields from concurrent access
- You need read/write access to a cache or registry
- Performance matters and the lock is held briefly (mutex is cheaper than channel operations)
- The critical section involves multiple related variables that must change atomically

**Real-world heuristic**: If you find yourself passing values between goroutines, use a channel. If you find yourself protecting access to a struct's internal state, use a mutex. Mixing both is fine — a struct can have a mutex protecting its fields AND return channels for event notification.

</details>

### Q4: Describe a real race condition you might write accidentally, and how the race detector catches it.

Tests practical experience with concurrent bugs.

<details>
<summary>Answer</summary>

**Classic race: map concurrent read/write**

```go
func handler(m map[string]int, key string) {
    go func() { m[key]++ }()  // goroutine 1 writes
    go func() { _ = m[key] }() // goroutine 2 reads
}
```

This is a data race: two goroutines access the same map location concurrently with at least one write. In Go, this panics at runtime with `concurrent map read and map write`.

**Classic race: closure captures loop variable**

```go
results := make([]int, 5)
for i := 0; i < 5; i++ {
    go func() {
        results[i] = i * i // two goroutines may write to same index
    }()
}
```

**The race detector** (`go test -race`, `go run -race`):

- Instruments every memory access with metadata (goroutine ID, logical clock)
- At each read/write, checks if a conflicting access happened from another goroutine without synchronization
- Reports the two conflicting accesses with their goroutine stacks

Output looks like:

```
DATA RACE
Write at 0x... by goroutine 7:
  main.main.func1()  main.go:15
Previous read at 0x... by goroutine 6:
  main.main.func2()  main.go:20
```

Use it in CI; the ~5-10x slowdown is acceptable for tests but not production.

</details>

## Key Takeaways

1. **Mutex for state, channels for communication**: The clearest mental model for choosing between them.
2. **Always `defer Unlock()`**: Prevents lock leaks on early returns and panics.
3. **Never copy a mutex**: Pass by pointer; `go vet` will catch copies.
4. **`RWMutex` for read-heavy caches**: Allows concurrent readers, but benchmark before assuming it's faster.
5. **`sync.Once` for lazy singletons**: Exactly one goroutine runs the initializer; subsequent calls are no-ops.
6. **Atomics for counters**: `sync/atomic` is faster than a mutex for simple integer operations.
7. **`sync.Map`**: Use for write-once/read-many patterns; use `map + mutex` for general concurrent maps.
8. **Race detector**: Run with `-race` in CI; it catches data races that cause subtle, intermittent bugs.
9. **Closure capture in goroutines**: Always pass loop variables as arguments, not captured references.

## Next Steps

In [Lesson 05: Memory Model & Escape Analysis](lesson-05-memory-model-and-escape-analysis.md), you'll learn:

- What "happens-before" means in the Go memory model and why it matters for synchronization
- How the compiler decides whether a variable lives on the heap or stack
- How to use `go build -gcflags="-m"` to see escape analysis decisions
- How to tune the garbage collector with `GOGC` and `GOMEMLIMIT`
