# Lesson 05: Memory Model & Escape Analysis

Most Go developers understand that "the GC handles memory." Fewer understand _when_ a variable lives on the stack vs the heap, _what_ the Go memory model guarantees about concurrent access, or _how_ to prevent GC pressure from slowing down high-throughput services. This is where experienced Go developers distinguish themselves.

## The Go Memory Model

The **Go memory model** specifies when one goroutine can be guaranteed to _observe_ a write made by another goroutine. Without these guarantees, concurrent code is undefined behavior — even if it looks correct.

### The Happens-Before Relationship

An operation A **happens-before** B if A's effects are guaranteed to be visible to B. Concurrent Go code is only safe if you can establish a happens-before relationship between your writes and reads.

**What establishes happens-before?**

| Synchronization          | Guarantee                                                 |
| ------------------------ | --------------------------------------------------------- |
| `sync.Mutex` Lock/Unlock | Unlock happens-before the next Lock                       |
| Channel send             | Send happens-before the corresponding receive completes   |
| Channel close            | Close happens-before a receive of the zero value          |
| `sync.WaitGroup` Done    | All `Done()` calls happen-before `Wait()` returns         |
| `sync/atomic` operations | Sequentially consistent within atomic operations          |
| `sync.Once`              | `Do` function return happens-before any `Do` call returns |

### What Is NOT Safe Without Synchronization

```go
var x int
var ready bool

// Goroutine 1
x = 42
ready = true  // ❌ No synchronization

// Goroutine 2
if ready {
    fmt.Println(x) // ❌ May see x=0 even if ready=true
                    // CPU and compiler may reorder writes
}
```

Without synchronization, the compiler and CPU are free to reorder writes. Goroutine 2 might see `ready=true` before seeing `x=42`.

**Fix: synchronize with a channel or mutex**

```go
var x int
ch := make(chan struct{})

go func() {
    x = 42
    ch <- struct{}{} // send happens-after x = 42
}()

<-ch            // receive happens-after send
fmt.Println(x)  // ✓ guaranteed to see x=42
```

### The Common Pattern That Is NOT Thread-Safe

```go
// ❌ This is a data race even though it "looks atomic"
type Config struct {
    mu  sync.Mutex
    val int
}

func (c *Config) Update(v int) {
    c.mu.Lock()
    c.val = v
    c.mu.Unlock()
}

// In another goroutine, without the mutex:
fmt.Println(c.val) // ❌ DATA RACE: no happens-before with Update
```

Always access shared state through the same synchronization primitive.

## Heap vs Stack Allocation

Every variable in Go lives either on the **stack** or the **heap**:

|                  | Stack                             | Heap                                  |
| ---------------- | --------------------------------- | ------------------------------------- |
| **Allocation**   | Near-instant (pointer bump)       | Runtime allocator                     |
| **Deallocation** | Automatic when function returns   | GC                                    |
| **GC pressure**  | None                              | Yes — more allocations = more GC work |
| **Access**       | Slightly faster (cache locality)  | Slightly slower                       |
| **Limit**        | Grows dynamically (starts ~2-8KB) | Available RAM                         |

The Go compiler decides stack vs heap through **escape analysis** at compile time.

## Escape Analysis

A variable **escapes** to the heap when the compiler determines it must outlive the stack frame that created it.

### Common Escape Scenarios

**1. Returning a pointer to a local variable**

```go
func newUser() *User {
    u := User{Name: "Alice"} // u escapes to heap
    return &u                // pointer outlives the stack frame
}
```

This is fine and idiomatic. The compiler handles it automatically.

**2. Storing in an interface**

```go
func store(i interface{}) {
    // i's underlying value escapes because we don't know the concrete type at compile time
}

u := User{}
store(u) // u escapes to heap
```

This is why hot paths with many interface conversions can increase GC pressure.

**3. Closures capturing variables**

```go
func counter() func() int {
    n := 0         // n escapes: the closure outlives counter()
    return func() int {
        n++
        return n
    }
}
```

**4. Slice/map elements whose backing array is on the heap**

```go
s := make([]int, 1000) // backing array escapes to heap (too large for stack)
```

### Viewing Escape Analysis

```sh
go build -gcflags="-m" ./...
# or more verbose:
go build -gcflags="-m -m" ./...
```

Output:

```
./main.go:10:6: moved to heap: u
./main.go:15:14: &u escapes to heap
./main.go:20:13: make([]int, 1000) escapes to heap
```

```sh
# Escape analysis for a specific function:
go tool compile -m main.go
```

### Reducing Escapes for Performance

```go
// ❌ Allocates on heap (escapes due to interface)
func sum(nums []interface{}) int {
    total := 0
    for _, n := range nums {
        total += n.(int)
    }
    return total
}

// ✓ Stack allocation (concrete types, no escape)
func sum(nums []int) int {
    total := 0
    for _, n := range nums {
        total += n
    }
    return total
}
```

```go
// ❌ Heap allocation: struct returned as interface escapes
func newWriter() io.Writer {
    return &bytes.Buffer{} // escapes
}

// ✓ Return concrete type when possible; let caller decide
func newBuffer() *bytes.Buffer {
    return &bytes.Buffer{} // still escapes, but fewer interface conversions in hot path
}
```

## Garbage Collector Basics

Go uses a **concurrent, tri-color, mark-and-sweep** GC. Key properties:

- Runs concurrently with your program (most of the time)
- Causes brief **stop-the-world** pauses (typically <1ms in modern Go)
- Triggered when heap size doubles since last GC (controlled by `GOGC`)

### GC Tuning: `GOGC`

`GOGC` controls when GC runs. Default: `100` (means "GC when live heap doubles").

```sh
GOGC=100   # default: GC when heap grows 100% above baseline
GOGC=200   # GC less frequently: heap can grow 200% before GC triggers
GOGC=off   # disable GC (dangerous: memory grows unbounded)
GOGC=50    # GC more frequently: lower latency, higher CPU usage
```

**Trade-off**: Higher `GOGC` = less GC CPU overhead = higher peak memory. Lower `GOGC` = more GC = lower peak memory.

```go
import "runtime/debug"

// Set programmatically:
debug.SetGCPercent(200) // equivalent to GOGC=200
```

### GC Tuning: `GOMEMLIMIT` (Go 1.19+)

`GOMEMLIMIT` sets an absolute memory ceiling. When the heap approaches the limit, GC runs more aggressively.

```sh
GOMEMLIMIT=1GiB   # GC will keep heap below 1GB
GOMEMLIMIT=512MiB # for memory-constrained containers
```

```go
import "runtime/debug"

debug.SetMemoryLimit(1 << 30) // 1 GiB
```

**`GOMEMLIMIT` vs `GOGC`**: Use `GOMEMLIMIT` in containerized environments to prevent OOM kills. It gives the GC a hard ceiling rather than a ratio. Often, setting `GOGC=off` with `GOMEMLIMIT` is effective — the GC only runs when needed to stay under the memory limit.

```sh
# Common production pattern for containers:
GOGC=off GOMEMLIMIT=1800MiB  # container limit is 2GiB, headroom for non-heap memory
```

### `runtime.GC()` Gotchas

```go
runtime.GC() // Forces an immediate GC cycle
```

**When NOT to call it**:

- In hot paths — you're adding synchronous GC pauses
- As a "fix" for high memory — address the allocation root cause instead
- In benchmarks without understanding the impact

**When it's acceptable**:

- In tests to get a stable memory baseline before checking for leaks
- In CLI tools at exit, to ensure finalizers run

## Object Pooling with `sync.Pool`

`sync.Pool` reduces GC pressure by reusing objects instead of allocating new ones.

```go
var bufPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func process(data []byte) string {
    buf := bufPool.Get().(*bytes.Buffer) // get or create
    defer func() {
        buf.Reset()
        bufPool.Put(buf) // return for reuse
    }()

    buf.Write(data)
    return buf.String()
}
```

**Rules for `sync.Pool`**:

- Objects in the pool can be **garbage collected at any time** (particularly during GC)
- Never use a Pool to manage resources that need explicit cleanup (file handles, connections)
- Always `Reset()` pooled objects before returning them — next caller gets a clean object
- Use for expensive-to-allocate, frequently-used temporary objects

The `fmt` package and `encoding/json` use pools internally for their scratch buffers.

## Hands-On Exercise 1: Finding Escape Points

Run escape analysis on this code and explain why each allocation escapes.

```go
package main

import "fmt"

type Point struct {
    X, Y float64
}

func newPoint(x, y float64) *Point {
    p := Point{X: x, Y: y}
    return &p
}

func describe(v interface{}) string {
    return fmt.Sprintf("%v", v)
}

func makePoints(n int) []*Point {
    points := make([]*Point, n)
    for i := 0; i < n; i++ {
        points[i] = newPoint(float64(i), float64(i))
    }
    return points
}

func main() {
    p := newPoint(1, 2)
    fmt.Println(describe(p))
    pts := makePoints(10)
    fmt.Println(len(pts))
}
```

<details>
<summary>Solution</summary>

Run: `go build -gcflags="-m" .`

Expected output and explanations:

1. **`p := Point{X: x, Y: y}` in `newPoint`** → escapes to heap
   - `&p` is returned to the caller, outliving the stack frame of `newPoint`

2. **`v interface{}` in `describe`** → the argument passed to `describe` escapes
   - Storing a concrete type in an `interface{}` parameter causes the value to escape (interface boxing)

3. **`make([]*Point, n)`** → escapes to heap
   - The slice backing array escapes because `n` is not a compile-time constant; the size is unknown at compile time so it can't be stack-allocated

4. **`fmt.Sprintf("%v", v)`** → internally allocates; the result string is on heap
   - `fmt.Sprintf` allocates the output string on the heap

**Reduction strategies**:

- Return `Point` (value) instead of `*Point` in `newPoint` — avoids heap allocation if the caller doesn't need the pointer to outlive its scope
- Use `fmt.Println` with concrete types instead of `interface{}` in hot paths

</details>

## Hands-On Exercise 2: Object Pool for JSON Encoding

Implement a JSON encoder pool to reduce allocations in a high-throughput handler.

```go
// Requirements:
// 1. Create a pool of *bytes.Buffer for use in JSON encoding
// 2. Implement EncodeJSON(v interface{}) ([]byte, error) that uses the pool
// 3. Make sure the buffer is properly reset before returning to pool
// 4. Handle encoding errors without leaking the buffer
```

<details>
<summary>Solution</summary>

```go
package main

import (
    "bytes"
    "encoding/json"
    "sync"
)

var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func EncodeJSON(v interface{}) ([]byte, error) {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()      // clear contents
        bufferPool.Put(buf) // return to pool
    }()

    if err := json.NewEncoder(buf).Encode(v); err != nil {
        return nil, err // defer still returns buf to pool
    }

    // Make a copy: buf will be reset when returned to pool
    result := make([]byte, buf.Len())
    copy(result, buf.Bytes())
    return result, nil
}
```

**Note**: We copy the bytes before returning the buffer to the pool. If we returned `buf.Bytes()` directly, the caller would hold a slice pointing into a buffer that gets reset the moment the deferred function runs.

</details>

## Interview Questions

### Q1: What is escape analysis and how do you use it to reduce GC pressure?

Tests whether candidates understand the runtime implications of their code, not just correctness.

<details>
<summary>Answer</summary>

**Escape analysis** is a compiler optimization that determines at compile time whether a variable can live on the stack or must live on the heap. Stack allocations are essentially free (pointer bump) and don't require GC. Heap allocations are managed by the GC.

A variable **escapes to the heap** when:

- Its address is returned from the function (pointer outlives the stack frame)
- It's stored in an interface (concrete type must be heap-allocated for boxing)
- The compiler can't determine its size at compile time
- A closure captures it and the closure outlives the function

**To view escape analysis**:

```sh
go build -gcflags="-m" ./...
```

**Reducing escapes** in hot paths:

- Return values (not pointers) for small structs — let the caller take the address if needed
- Use concrete types instead of interfaces in inner loops
- Use `sync.Pool` for frequently-allocated temporary objects (buffers, scratch slices)
- Pre-allocate slices with `make([]T, 0, capacity)` when size is known

**Caveat**: Premature optimization. Measure with `go test -bench` and `go tool pprof` before optimizing. Most programs are not allocation-bound.

</details>

### Q2: What does the Go memory model guarantee, and what are the implications for concurrent programming?

A deeper question that tests understanding of the formal memory model.

<details>
<summary>Answer</summary>

The Go memory model specifies **happens-before** relationships: the conditions under which one goroutine's writes are guaranteed to be visible to another goroutine's reads.

**Guarantees**:

- Within a single goroutine: statements execute in the order they appear (no reordering visible to that goroutine)
- Channel sends happen-before the corresponding receive
- A mutex unlock happens-before the next lock of the same mutex
- Package `init` functions complete before `main.main` starts
- `sync.WaitGroup.Done` happens-before `Wait` returns

**What is NOT guaranteed without synchronization**:

- One goroutine's write is visible to another goroutine's read
- Two goroutines see writes in the same order

**Practical implication**: Any shared memory accessed from multiple goroutines (where at least one access is a write) must be protected by a synchronization primitive. Even "trivial" writes like `flag = true` are not safe without synchronization, because the compiler or CPU may reorder operations.

The race detector (`-race`) catches violations at runtime, but some races may not trigger on every run. The memory model is the formal specification; the race detector is the practical tool.

</details>

### Q3: How does `GOGC` differ from `GOMEMLIMIT`, and when do you use each?

A production-focused question about GC tuning.

<details>
<summary>Answer</summary>

**`GOGC`** (default: `100`) controls the **ratio** at which GC triggers. At `GOGC=100`, GC runs when the live heap has grown 100% since the last GC. It's a relative measure — the threshold grows with your live data.

- Higher `GOGC`: less frequent GC, more memory, less CPU overhead
- Lower `GOGC`: more frequent GC, less memory, more CPU overhead
- `GOGC=off`: GC disabled (memory grows unbounded)

**`GOMEMLIMIT`** (default: unlimited) sets an **absolute ceiling** on the Go heap. When the heap approaches the limit, GC runs more aggressively regardless of `GOGC`.

**When to use each**:

- **Containerized services** (most common today): Set `GOMEMLIMIT` to ~90% of the container memory limit to prevent OOM kills. The GC will manage below that ceiling. Pair with `GOGC=off` if you want "run GC only when needed to stay under the limit."

- **Latency-sensitive services**: Lower `GOGC` (e.g., 50) reduces the size of each GC cycle, which can lower tail latency at the cost of higher average CPU.

- **Throughput-focused batch jobs**: Higher `GOGC` (e.g., 400-800) amortizes GC cost, but requires more memory headroom.

```sh
# Production container pattern:
GOGC=off GOMEMLIMIT=1800MiB  # container limit is 2GiB
```

</details>

### Q4: What is `sync.Pool` and what are its limitations?

Tests knowledge of pooling patterns and the subtleties of Go's GC interaction.

<details>
<summary>Answer</summary>

`sync.Pool` is a thread-safe pool of temporary objects that can be reused across goroutines. It reduces GC pressure by amortizing allocation cost — instead of allocating a new `bytes.Buffer` for every request, you get one from the pool and return it when done.

**Key properties**:

- `Get()`: returns a pooled object, or calls `New` if the pool is empty
- `Put()`: returns an object to the pool for reuse
- Objects may be **garbage collected at any GC cycle** — the pool gives no persistence guarantee

**Limitations**:

1. **No guarantee of persistence**: The GC can drain the pool at any time. After a GC, `Get()` may call `New` again. Don't use it to "cache" expensive state across GC cycles.
2. **Not for resources with finalizers**: File handles, database connections — use a proper resource pool with explicit lifecycle management.
3. **Must reset before Put**: Return the object in a clean state. If you don't, the next caller gets dirty data.
4. **Not sized**: `sync.Pool` has no capacity limit. In theory, every goroutine can `Get` and never `Put`, causing unbounded `New` calls.

**Ideal use cases**: Temporary scratch buffers (`bytes.Buffer`), encoder/decoder instances, large arrays for intermediate computation. The standard library uses it in `fmt`, `encoding/json`, `compress/gzip`.

</details>

## Key Takeaways

1. **Memory model = happens-before**: Shared state requires synchronization to be visible across goroutines — the compiler and CPU can reorder writes.
2. **Stack vs heap**: Stack allocations are free; heap allocations cause GC work. The compiler decides via escape analysis.
3. **Escape triggers**: Returning pointers, storing in interfaces, closures, unknown-size allocations all cause heap escape.
4. **View escapes**: `go build -gcflags="-m"` shows what the compiler allocates on the heap and why.
5. **`GOGC`**: Ratio-based GC trigger (default 100% growth). Tune for latency vs throughput vs memory.
6. **`GOMEMLIMIT`**: Absolute memory ceiling (Go 1.19+). Essential for containerized workloads to prevent OOM kills.
7. **`sync.Pool`**: Reduces allocation pressure for frequently-used temporary objects — but the GC can drain it at any time.
8. **Profile before optimizing**: Use `go test -bench` and `pprof` to find actual allocation hotspots, not guesses.

## Next Steps

In [Lesson 06: Context & Cancellation](lesson-06-context-and-cancellation.md), you'll learn:

- The `context.Context` interface contract and what it represents
- How to propagate cancellation through call stacks with `WithCancel`, `WithTimeout`, `WithDeadline`
- Why `context.WithValue` is often misused and the right patterns for it
- Practical context use in HTTP servers and database calls
