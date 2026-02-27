# Lesson 03: Goroutines & the Go Scheduler

Go's concurrency is built on goroutines, but the real interview depth is in the scheduler underneath them. Why are goroutines cheaper than threads? What happens when one goroutine blocks? Why does `GOMAXPROCS` matter? And why do goroutine leaks silently grow until they crash your production service?

> **Note**: Channels are covered in depth in [`go/channels/`](../channels/). This lesson focuses on the goroutine runtime, scheduler mechanics, and lifecycle management.

## Goroutines vs OS Threads

Most languages expose OS threads directly. Go abstracts them away behind goroutines.

|                     | OS Thread             | Goroutine                 |
| ------------------- | --------------------- | ------------------------- |
| **Stack size**      | ~1-8 MB (fixed)       | 2-8 KB (growable)         |
| **Scheduling**      | OS kernel             | Go runtime                |
| **Context switch**  | ~1-2 µs (kernel call) | ~200 ns (user space)      |
| **Practical limit** | ~thousands            | Millions                  |
| **Blocking I/O**    | Blocks the thread     | Scheduler parks goroutine |

A single Go program with 100,000 goroutines is normal. 100,000 OS threads would exhaust memory.

## The M:N Threading Model

Go uses **M:N scheduling**: M goroutines multiplexed onto N OS threads.

```
Goroutines (G):  G1 G2 G3 G4 G5 G6 ... G100,000
                  ↕  ↕  ↕
OS Threads (M):  T1  T2  T3   (GOMAXPROCS controls this)
                  ↕  ↕  ↕
CPU Cores:       C1  C2  C3
```

The Go scheduler manages this mapping invisibly. Your goroutines run on whatever thread is available.

### The Three Entities: G, M, P

The scheduler works with three concepts:

- **G** (Goroutine): The goroutine itself — its stack, state, and code to execute
- **M** (Machine): An OS thread
- **P** (Processor): A logical processor — holds a run queue of goroutines ready to execute

Each P has a local run queue. An M needs a P to run goroutines. The number of Ps is set by `GOMAXPROCS`.

```
P1 [G1, G2, G3] -- M1 -- CPU core 1
P2 [G4, G5, G6] -- M2 -- CPU core 2
P3 [G7, G8]     -- M3 -- CPU core 3
          global run queue: [G9, G10, ...]
```

## `GOMAXPROCS`

`GOMAXPROCS` sets the number of P (logical processors). By default it equals the number of CPU cores.

```go
import "runtime"

runtime.GOMAXPROCS(4)         // set to 4 logical processors
n := runtime.GOMAXPROCS(0)   // 0 = query current value without changing
fmt.Println(n)                // typically == runtime.NumCPU()
```

**Key insight**: `GOMAXPROCS` controls _parallelism_ (actual simultaneous execution). You can have thousands of concurrent goroutines with `GOMAXPROCS=1` — they just take turns on a single thread.

```go
// GOMAXPROCS=1: goroutines are concurrent but not parallel
// GOMAXPROCS=4: up to 4 goroutines run truly in parallel
```

**When to change it**:

- CPU-bound workloads: the default (= NumCPU) is usually optimal
- I/O-bound workloads: `GOMAXPROCS` matters less because goroutines spend most time blocked
- Container environments: set via `GOMAXPROCS` env var or use [automaxprocs](https://github.com/uber-go/automaxprocs) to auto-detect cgroup CPU limits

## What Happens When a Goroutine Blocks

This is the magic of Go's scheduler — blocking doesn't waste an OS thread.

**Blocking on channel/mutex** (cooperative scheduling):

```
G1 blocks waiting for channel → Scheduler moves G1 off M1 → M1 picks up G2
When channel receives data → G1 is placed back in P's run queue
```

**Blocking on system call** (preemptive):

```
G1 calls os.ReadFile() → Scheduler detects system call → Hands off P to M2
M1 keeps running with G1 waiting for kernel → When kernel returns, G1 re-enters run queue
```

This means slow I/O doesn't stall other goroutines — the P (and its run queue) keeps running.

**Blocking on CPU** (timer-based preemption):
Since Go 1.14, the scheduler preempts goroutines that have been running too long using async signals (SIGURG). Before 1.14, a tight loop without function calls could starve other goroutines.

## Goroutine Lifecycle

```
New goroutine (go func)
    │
    ▼
Runnable ──── scheduled ────► Running
    ▲                              │
    │           preempted          │
    └──────────────────────────────┘
    │
    │  blocking operation
    ▼
Waiting (blocked)
    │
    │  event occurs
    ▼
Runnable again
    │
    │  function returns
    ▼
Dead (stack reclaimed)
```

Goroutines never explicitly die — they exit when their function returns. The runtime reclaims the stack.

## Goroutine Leaks

A **goroutine leak** happens when a goroutine is started but never terminates. It stays in memory forever, consuming stack space and potentially holding resources (file handles, DB connections, channels).

### Common Leak Patterns

**Pattern 1: Blocked channel send, no receiver**

```go
// ❌ Leaks: if processResult never reads, sender goroutine blocks forever
func doWork() {
    result := make(chan int) // unbuffered
    go func() {
        result <- compute() // blocks here if nobody reads
    }()
    // forgot to read from result
    return
}
```

**Pattern 2: Goroutine waiting on context that never cancels**

```go
// ❌ Leaks: if ctx is never cancelled, goroutine runs forever
func startWorker(ctx context.Context) {
    go func() {
        for {
            select {
            case <-ctx.Done():
                return // this never happens
            default:
                doSomething()
            }
        }
    }()
}

// Called with context.Background() — never cancels
startWorker(context.Background())
```

**Pattern 3: Goroutine blocked on receive from empty channel**

```go
// ❌ Leaks: if producer dies before sending, consumer blocks forever
func process() {
    ch := make(chan []byte)
    go func() {
        data := <-ch       // blocks forever if nobody sends
        process(data)
    }()
    // forgot to send to ch
}
```

### Detecting Goroutine Leaks

**Method 1: `runtime.NumGoroutine()`**

```go
before := runtime.NumGoroutine()
doSomething()
after := runtime.NumGoroutine()
if after > before {
    fmt.Printf("goroutine leak: %d goroutines created\n", after-before)
}
```

**Method 2: `pprof` goroutine profile** (production-safe)

```go
import _ "net/http/pprof"

// In main:
go http.ListenAndServe(":6060", nil)

// In terminal:
// go tool pprof http://localhost:6060/debug/pprof/goroutine
```

**Method 3: [goleak](https://github.com/uber-go/goleak) library** (for tests)

```go
func TestMyFunction(t *testing.T) {
    defer goleak.VerifyNone(t) // fails test if goroutines leak

    myFunction() // test subject
}
```

### Fixing Leaks: Use Context or Done Channels

```go
// ✓ Goroutine with exit mechanism
func startWorker(ctx context.Context) {
    go func() {
        for {
            select {
            case <-ctx.Done():
                return // exits when context is cancelled
            default:
                doWork()
            }
        }
    }()
}

// Caller controls lifetime:
ctx, cancel := context.WithCancel(context.Background())
startWorker(ctx)
// ... later:
cancel() // goroutine will exit
```

## Work Stealing

When one P's local run queue is empty, it doesn't sit idle — it **steals goroutines** from other Ps' queues.

```
P1 [empty]   P2 [G1, G2, G3, G4]

P1 steals half of P2's queue:
P1 [G3, G4]  P2 [G1, G2]
```

Work stealing ensures all CPU cores stay busy as long as there are runnable goroutines. It's automatic and transparent — you don't configure it.

**Practical implication**: Goroutines don't "stick" to a particular OS thread or core. A goroutine started on CPU core 1 might finish on core 3. Don't rely on thread-local storage or CPU affinity.

## `runtime.Gosched()`

Yields the processor, allowing other goroutines to run. Rarely needed in modern Go (preemption handles this), but occasionally useful:

```go
// Occasionally useful in tight CPU loops to give other goroutines a turn
for {
    doExpensiveWork()
    runtime.Gosched() // yield voluntarily
}
```

## Hands-On Exercise 1: Goroutine Leak Detection

The following code leaks goroutines. Find and fix all leaks.

```go
func fetchAll(urls []string) []string {
    results := make(chan string)

    for _, url := range urls {
        go func(u string) {
            resp, err := http.Get(u)
            if err != nil {
                results <- ""
                return
            }
            defer resp.Body.Close()
            body, _ := io.ReadAll(resp.Body)
            results <- string(body)
        }(url)
    }

    // Collect only the first result
    return []string{<-results}
}
```

<details>
<summary>Solution</summary>

**Issues**:

1. ❌ `len(urls) - 1` goroutines are left blocked trying to send to `results` — nobody reads them after the first result
2. ❌ No timeout or context — goroutines can hang forever on slow URLs

**Fixed**:

```go
func fetchAll(ctx context.Context, urls []string) []string {
    // Buffered channel: all goroutines can send without blocking
    results := make(chan string, len(urls))

    for _, url := range urls {
        go func(u string) {
            req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
            if err != nil {
                results <- ""
                return
            }
            resp, err := http.DefaultClient.Do(req)
            if err != nil {
                results <- ""
                return
            }
            defer resp.Body.Close()
            body, _ := io.ReadAll(resp.Body)
            results <- string(body)
        }(url)
    }

    // Collect ALL results
    out := make([]string, 0, len(urls))
    for i := 0; i < len(urls); i++ {
        out = append(out, <-results)
    }
    return out
}

// Caller controls timeout:
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
results := fetchAll(ctx, urls)
```

**Key changes**:

- Buffered channel prevents goroutines from blocking on send
- Context propagation enables cancellation and timeout
- Collect all results, not just the first one

</details>

## Hands-On Exercise 2: Worker Pool

Implement a bounded worker pool that processes jobs with a fixed number of goroutines.

```go
// Requirements:
// 1. Create a worker pool with `numWorkers` goroutines
// 2. Process `jobs` ([]int representing work items)
// 3. Return []int results in any order
// 4. All goroutines must exit cleanly when work is done
// 5. Support context cancellation
```

<details>
<summary>Solution</summary>

```go
package main

import (
    "context"
    "fmt"
    "sync"
)

func workerPool(ctx context.Context, numWorkers int, jobs []int) []int {
    jobCh := make(chan int, len(jobs))
    resultCh := make(chan int, len(jobs))

    // Start workers
    var wg sync.WaitGroup
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for {
                select {
                case <-ctx.Done():
                    return // context cancelled
                case job, ok := <-jobCh:
                    if !ok {
                        return // channel closed, no more jobs
                    }
                    resultCh <- job * job // do work
                }
            }
        }()
    }

    // Send all jobs
    for _, j := range jobs {
        jobCh <- j
    }
    close(jobCh) // signal workers: no more jobs

    // Wait for all workers to finish, then close results
    go func() {
        wg.Wait()
        close(resultCh)
    }()

    // Collect results
    var results []int
    for r := range resultCh {
        results = append(results, r)
    }
    return results
}

func main() {
    ctx := context.Background()
    jobs := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
    results := workerPool(ctx, 3, jobs)
    fmt.Println("Results:", results) // squares, any order
}
```

</details>

## Interview Questions

### Q1: Explain Go's M:N scheduling model. Why is it better than 1:1?

A foundational question that separates Go developers who have thought about the runtime from those who just use goroutines.

<details>
<summary>Answer</summary>

**1:1 model** (Java threads, Python threads without GIL bypass): Each goroutine/thread maps to one OS thread. OS threads are expensive (~1-8MB stack, kernel context switch ~1-2µs). You can't have millions of them.

**M:N model**: M goroutines run on N OS threads (N = GOMAXPROCS, typically = CPU cores). The Go runtime scheduler handles the mapping in user space.

**Advantages**:

1. **Cheap goroutines**: 2-8KB initial stack (grows as needed). Millions are practical.
2. **Fast context switches**: User-space scheduling (~200ns) vs kernel context switch (~1-2µs).
3. **Blocking I/O doesn't waste threads**: When a goroutine blocks on I/O, the scheduler gives the P (and its other goroutines) to another OS thread. The blocked goroutine doesn't consume CPU.
4. **Work stealing**: Idle Ps steal goroutines from busy Ps — automatic load balancing.

**Trade-off**: The scheduler itself has overhead. For compute-heavy work with exactly `GOMAXPROCS` goroutines and no blocking, raw throughput may be similar to OS threads. The win is in concurrency at scale.

</details>

### Q2: What is a goroutine leak, how do you detect it, and how do you prevent it?

A practical question that tests production experience.

<details>
<summary>Answer</summary>

A **goroutine leak** occurs when a goroutine starts but never terminates, holding its stack and any referenced resources.

**Common causes**:

- Goroutine blocked on a channel send/receive with no counterpart
- Goroutine waiting on context that is never cancelled
- Goroutine in an infinite loop with no exit condition

**Detection**:

- `runtime.NumGoroutine()` — monitor over time; unexplained growth signals leaks
- `pprof` goroutine endpoint (`/debug/pprof/goroutine`) — shows all goroutine stacks in production
- `goleak` library in tests — fails the test if goroutines are left running after the test

**Prevention**:

- Every goroutine must have a clear exit path: function return, channel close, or `ctx.Done()`
- Use `context.Context` to propagate cancellation
- Use buffered channels (or fan-in) so goroutines don't block indefinitely on send
- Use `WaitGroup` to ensure you wait for goroutines to finish before returning from the function that started them

</details>

### Q3: What does `GOMAXPROCS` control, and when would you change it?

Tests understanding of the difference between concurrency and parallelism.

<details>
<summary>Answer</summary>

`GOMAXPROCS` controls the number of P (logical processors) — effectively the maximum number of goroutines running **in parallel** on OS threads simultaneously.

Default: `runtime.NumCPU()` (all available CPU cores).

**When to change**:

- **Container environments**: The Go runtime reads `/proc/cpuinfo` which shows physical host CPUs, not the container's CPU limit. In a container with `--cpus=2` running on a 32-core host, Go defaults to 32 Ps. Use `automaxprocs` or set `GOMAXPROCS=2` to match the actual allocation.
- **Benchmarking**: Temporarily set to 1 to understand sequential behaviour.
- **Rare tuning**: CPU-bound programs may benefit from `< NumCPU` to reduce scheduler overhead; usually not worth it.

**What it doesn't control**:

- The number of goroutines you can create (that's limited by available memory)
- I/O concurrency (goroutines blocked on I/O don't count against GOMAXPROCS)

</details>

### Q4: What changed in Go 1.14 regarding goroutine scheduling, and why does it matter?

Tests depth of runtime knowledge — a question interviewers ask when probing for deeper understanding.

<details>
<summary>Answer</summary>

Before Go 1.14, Go used **cooperative scheduling** for goroutines. A goroutine would only yield control at certain points: function calls, channel operations, system calls. A tight loop with no function calls (e.g., a numeric computation or a `for {}` spin-wait) would never yield and would starve other goroutines on the same P.

```go
// This goroutine could starve others pre-Go 1.14 on GOMAXPROCS=1:
go func() {
    for { } // tight loop, never yields
}()
```

Go 1.14 introduced **asynchronous preemption**: the runtime sends `SIGURG` to OS threads at ~10ms intervals. Signal handlers inspect the goroutine's current PC (program counter) and, at a safe point, preempt it. This means:

1. Long-running CPU loops no longer starve other goroutines
2. GC stop-the-world pauses are more bounded (can preempt tight loops)
3. Goroutine preemption latency is now O(10ms) worst case, not unbounded

**Practical implication**: You no longer need to add `runtime.Gosched()` calls inside tight loops to prevent starvation. The runtime handles it automatically.

</details>

## Key Takeaways

1. **M:N model**: M goroutines run on N OS threads — goroutines are cheap (2-8KB), threads are expensive (1-8MB).
2. **`GOMAXPROCS`**: Controls parallelism, not concurrency. Default = NumCPU; set correctly in containers.
3. **Blocking is transparent**: When a goroutine blocks on I/O, the scheduler moves other goroutines to available threads.
4. **Work stealing**: Idle Ps steal goroutines from busy Ps automatically — you don't configure this.
5. **Goroutine leaks**: Goroutines that never exit are a silent killer — use context cancellation or done channels for every goroutine.
6. **Detect leaks**: `runtime.NumGoroutine()` in code, `pprof` in production, `goleak` in tests.
7. **Preemption (1.14+)**: Tight loops no longer starve other goroutines — async preemption via SIGURG.
8. **Goroutine lifecycle**: Goroutines exit when their function returns. No explicit kill mechanism exists.

## Next Steps

In [Lesson 04: Sync Primitives & Patterns](lesson-04-sync-primitives-and-patterns.md), you'll learn:

- When to use a mutex vs a channel for synchronization
- How `sync.RWMutex` enables concurrent reads with exclusive writes
- `sync.WaitGroup`, `sync.Once`, and `sync/atomic` for common patterns
- `sync.Map` and when it outperforms a `map` with a mutex
- The most common race conditions and how to catch them
