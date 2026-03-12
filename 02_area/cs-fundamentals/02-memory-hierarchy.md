# Memory Hierarchy

In lesson 01 we saw that the CPU works on registers — but registers only hold 128 bytes. Your program's data lives in RAM, which is huge but slow. The memory hierarchy exists to bridge this gap. Understanding it explains why two programs with identical algorithmic complexity can differ by 10-100x in real-world performance.

## The Problem: CPU Is Fast, RAM Is Slow

A modern CPU can execute an instruction every ~0.3 ns. A RAM access takes ~100 ns. That's a 300x mismatch. If the CPU had to wait for RAM on every load, it would sit idle 99% of the time.

```
CPU speed:   ~3 GHz = one instruction every 0.3 ns
RAM latency: ~100 ns = ~300 CPU cycles

If the CPU waited for RAM every time:
  Useful work: 1 cycle
  Waiting:     300 cycles
  Efficiency:  0.3%
```

This is the fundamental problem. Caches exist to solve it.

## The Hierarchy

```
                  ┌──────────┐
                  │ Registers│  0.3 ns    128 bytes     per core
                  └────┬─────┘
                  ┌────┴─────┐
                  │ L1 Cache │  ~1 ns     32-64 KB      per core
                  └────┬─────┘
                  ┌────┴─────┐
                  │ L2 Cache │  ~4 ns     256 KB-1 MB   per core
                  └────┬─────┘
                  ┌────┴─────┐
                  │ L3 Cache │  ~10 ns    8-64 MB       shared across cores
                  └────┬─────┘
                  ┌────┴─────┐
                  │   RAM    │  ~100 ns   8-128 GB      shared
                  └────┬─────┘
                  ┌────┴─────┐
                  │   Disk   │  ~10,000-  TB+            shared
                  │ (SSD/HDD)│  100,000 ns
                  └──────────┘

Each level: ~3-10x slower, ~10-1000x bigger
```

Every level acts as a cache for the level below it. L1 caches hot data from L2. L2 caches from L3. L3 caches from RAM. RAM caches from disk (virtual memory / page cache). The hardware manages all of this transparently — your code just reads and writes memory addresses.

## Cache Lines — The Unit of Transfer

The CPU never fetches a single byte from RAM. It fetches a **cache line** — a contiguous block of 64 bytes on virtually all modern hardware.

When you read `array[0]` (an 8-byte int64), the hardware fetches bytes 0-63 into the cache. Now `array[1]` through `array[7]` are already cached — they came for free.

```
Memory:  [  0-63  ] [ 64-127 ] [128-191 ] [192-255 ] ...
              ^
         You asked for byte 4 (array[0])
         CPU fetched the entire 64-byte cache line
         array[0] through array[7] are now in L1 cache
```

This is why **sequential access is fast** and **random access is slow** — it's not about the algorithm, it's about how the hardware fetches data.

### Sequential vs random access — the real cost

```go
// Sequential: each access is a cache hit (after the first)
// because the next element is already in the cache line
func sumSequential(data []int64) int64 {
    var total int64
    for i := 0; i < len(data); i++ {
        total += data[i]  // next element is 8 bytes away, same cache line
    }
    return total
}

// Random: every access is likely a cache miss
// because we jump to an unpredictable location
func sumRandom(data []int64, indices []int) int64 {
    var total int64
    for _, idx := range indices {
        total += data[idx]  // random jump, probably a different cache line
    }
    return total
}
```

On a large array (say, 100 million elements), the random version can be **10-50x slower** than the sequential version. Same number of additions. Same algorithm. The difference is entirely cache behaviour.

You can measure this in Go:

```go
func BenchmarkSequential(b *testing.B) {
    data := make([]int64, 10_000_000)
    for i := range data { data[i] = int64(i) }
    b.ResetTimer()
    for n := 0; n < b.N; n++ {
        sumSequential(data)
    }
}

func BenchmarkRandom(b *testing.B) {
    data := make([]int64, 10_000_000)
    indices := rand.Perm(len(data))
    for i := range data { data[i] = int64(i) }
    b.ResetTimer()
    for n := 0; n < b.N; n++ {
        sumRandom(data, indices)
    }
}
```

### Struct layout and cache lines

This has direct consequences for how you design data structures:

```go
// Bad for iteration: each Player is 80+ bytes
// Iterating over health touches one field per cache line, wasting the rest
type Player struct {
    Name      string    // 16 bytes (pointer + length)
    Inventory [32]Item  // lots of bytes
    Health    int       // 8 bytes — this is what we want
    Score     int       // 8 bytes
    X, Y, Z  float64   // 24 bytes
}

func totalHealth(players []Player) int {
    total := 0
    for _, p := range players {
        total += p.Health  // each access loads a full cache line,
                           // but only uses 8 bytes of it
    }
    return total
}
```

```go
// Better for iteration over health: separate the hot field
type PlayerHealth struct {
    Health []int        // packed together, 8 ints per cache line
}
type PlayerData struct {
    Name      []string
    Inventory [][32]Item
    Score     []int
    X, Y, Z  []float64
}

func totalHealth(h PlayerHealth) int {
    total := 0
    for _, hp := range h.Health {
        total += hp  // sequential, 8 values per cache line
    }
    return total
}
```

This pattern is called **Struct of Arrays (SoA)** vs **Array of Structs (AoS)**. Game engines and high-performance systems use SoA because it packs the data the CPU actually needs into contiguous cache lines. Most business software uses AoS because it's simpler and the performance difference doesn't matter when you're waiting on network I/O anyway.

**When to care:** if you're iterating over thousands+ of items and only touching a few fields, SoA can make a real difference. If you're doing CRUD with database calls, don't bother.

## Hardware Prefetcher — The CPU Predicts Your Access Pattern

The CPU doesn't just cache what you asked for — it tries to predict what you'll ask for next. The **hardware prefetcher** detects sequential and strided access patterns and fetches cache lines ahead of time.

```
Your code reads:  array[0], array[1], array[2], array[3] ...

Prefetcher sees:  "sequential pattern detected"
Prefetcher does:  fetches array[8..15] while you're still processing array[4]

Result: by the time you reach array[8], it's already in cache.
        Effective latency: near zero.
```

This works for:

- **Sequential access** — arrays, slices iterated in order
- **Constant stride** — accessing every 2nd, 4th, etc. element
- **Multiple streams** — up to ~8-16 independent sequential patterns simultaneously

This does NOT work for:

- **Pointer chasing** — linked lists, trees, hash maps with chaining
- **Random access** — hash map lookups, random array indices
- **Large strides** — jumping by more than ~2KB between accesses

```go
// Prefetcher loves this — predictable stride
for i := 0; i < len(matrix); i += stride {
    sum += matrix[i]
}

// Prefetcher is useless here — next address depends on current value
node := head
for node != nil {
    sum += node.Value
    node = node.Next  // address of Next is unknowable until this loads
}
```

This is why **slices beat linked lists** in virtually every real-world benchmark, even for operations where linked lists have better algorithmic complexity (like insertion in the middle). The cache and prefetcher advantages of contiguous memory overwhelm the O(n) vs O(1) difference for any list that fits in cache.

## Cache Misses — What Actually Makes Code Slow

A **cache hit** is when the data you need is already in cache. A **cache miss** is when it's not, and the CPU has to go to a slower level (or all the way to RAM).

Types of cache misses:

| Type           | Cause                                    | Example                                             |
| -------------- | ---------------------------------------- | --------------------------------------------------- |
| **Compulsory** | First time accessing this data           | First iteration of a loop over new data             |
| **Capacity**   | Working set exceeds cache size           | Iterating over a 100 MB array — can't fit in L3     |
| **Conflict**   | Two addresses map to the same cache slot | Rare in practice with modern set-associative caches |

### The working set concept

Your program's **working set** is the amount of data it actively touches in a given time window. Performance cliffs happen when your working set exceeds a cache level:

```
Working set size    Where it lives       Effective access time
< 32 KB             L1 cache             ~1 ns
< 256 KB            L2 cache             ~4 ns
< 8 MB              L3 cache             ~10 ns
> L3 size           RAM                  ~100 ns
```

This is why you sometimes see sudden performance drops when data grows past a threshold — it's not your algorithm getting slower, it's your working set falling out of a cache level.

You can observe this by benchmarking the same operation on arrays of increasing size:

```go
func BenchmarkSum(b *testing.B) {
    for _, size := range []int{
        1 << 10,   // 8 KB  — fits in L1
        1 << 13,   // 64 KB — exceeds L1, fits in L2
        1 << 17,   // 1 MB  — exceeds L2, fits in L3
        1 << 23,   // 64 MB — exceeds L3, hits RAM
    } {
        data := make([]int64, size)
        for i := range data { data[i] = int64(i) }
        b.Run(fmt.Sprintf("size=%d", size*8), func(b *testing.B) {
            for n := 0; n < b.N; n++ {
                sumSequential(data)
            }
        })
    }
}
```

You'll see throughput drop at each cache boundary.

## How Different Languages Interact With the Memory Hierarchy

The memory hierarchy is hardware — it's the same for every language. But languages differ in how well they let you exploit it.

### Go — You're close to the metal

Go gives you direct control over memory layout:

```go
// Contiguous in memory — cache-friendly
points := make([]Point, 1000000)

// Each Point is laid out sequentially:
// [x0,y0,z0] [x1,y1,z1] [x2,y2,z2] ...
// Iterating over this is fast — sequential access, prefetcher works
```

Go slices are backed by contiguous arrays. When you iterate a `[]Point`, you're walking through memory sequentially. The prefetcher kicks in after a few elements, and the rest of the iteration runs at nearly memory-bandwidth speed.

**Pointers break this.** If your slice contains pointers (`[]*Point`), each element is an 8-byte pointer to a `Point` allocated somewhere on the heap. Iterating means chasing pointers to random locations:

```go
// Contiguous — fast
points := make([]Point, 1000000)      // all Points packed together

// Pointer chasing — slow
points := make([]*Point, 1000000)     // pointers are packed, but the
for i := range points {                // Points they reference are
    points[i] = &Point{...}           // scattered across the heap
}
```

The pointer version can be 3-10x slower for iteration-heavy code, purely because of cache behaviour.

### Java — A layer of indirection

In Java, almost everything is a reference (pointer). `ArrayList<Point>` is an array of references to Point objects scattered across the heap:

```
Java ArrayList<Point>:
  [ref0] [ref1] [ref2] [ref3] ...    ← contiguous array of references
    |      |      |      |
    v      v      v      v
  Point  Point  Point  Point          ← objects scattered on the heap
  @0x1a  @0x7f  @0x3c  @0x92

Every access: load reference → follow pointer → load object
Two cache misses instead of zero.
```

Java's value types (Project Valhalla, still in progress) aim to fix this by allowing objects to be stored inline in arrays, like Go structs. Until then, Java's cache performance for data-heavy iteration is fundamentally limited by this indirection.

Primitives (`int[]`, `double[]`) don't have this problem — they're stored contiguously, like Go slices.

### TypeScript/JavaScript — V8 is smarter than you'd expect

JavaScript arrays are complicated. V8 uses different internal representations depending on what you put in them:

```javascript
// V8 internally stores this as a contiguous C array of doubles
// — same layout as Go's []float64
const nums = [1.1, 2.2, 3.3, 4.4, 5.5];

// V8 internally stores this as a contiguous C array of SMIs (Small Integers)
// — no heap allocation per element
const ids = [1, 2, 3, 4, 5];

// This forces V8 to use a "dictionary mode" or mixed array
// — much slower, like Java's ArrayList<Object>
const mixed = [1, "hello", { x: 1 }, null];
```

V8's hidden classes and inline caches do remarkable work to make object access fast. But you have no direct control over memory layout. You can't decide "pack these objects contiguously." V8 might do it, or might not. You can write cache-friendly JavaScript by using typed arrays:

```javascript
// Guaranteed contiguous, cache-friendly — same as Go's []float64
const positions = new Float64Array(1000000);

// Regular array — V8 tries its best but no guarantees
const positions = new Array(1000000);
```

### Python — Worst case for cache performance

Python combines all the problems:

```python
# A Python list of integers
nums = [1, 2, 3, 4, 5]
```

```
Python list:
  [PyObject*] [PyObject*] [PyObject*] ...   ← array of pointers
       |           |           |
       v           v           v
   PyObject     PyObject     PyObject        ← full objects on heap
   refcount=1   refcount=1   refcount=1      ← 28 bytes each for a
   type=int     type=int     type=int           simple integer
   value=1      value=2      value=3
```

Every integer in Python is a heap-allocated object with a reference count, a type pointer, and the actual value. A Python `int` takes ~28 bytes. A Go `int64` takes 8 bytes.

Iterating over a Python list of integers means: load pointer, follow to heap object, read type tag, confirm it's an int, extract value. Two levels of indirection, type checking on every access, and the objects are scattered in memory.

This is why NumPy exists — it stores numbers in contiguous C arrays (like Go slices), bypassing Python's object model entirely:

```python
import numpy as np

# Contiguous C array of 64-bit floats — cache-friendly
arr = np.array([1.0, 2.0, 3.0, 4.0, 5.0])

# Summing this runs a tight C loop over packed data
# 100x faster than sum() on a Python list
total = np.sum(arr)
```

NumPy's performance advantage isn't clever algorithms — it's that the data is laid out in cache-friendly contiguous memory and processed by compiled C code that the CPU can pipeline and prefetch.

## Practical Summary — What Data Layout Decisions Actually Matter

| Situation                           | Fast                             | Slow                                 | Why                                      |
| ----------------------------------- | -------------------------------- | ------------------------------------ | ---------------------------------------- |
| Iterating a collection              | `[]Point` (value slice)          | `[]*Point` (pointer slice)           | Contiguous vs scattered memory           |
| Processing numbers in Python        | `numpy.array`                    | Python `list`                        | C array vs heap-allocated objects        |
| JS numeric arrays                   | `Float64Array`                   | Regular `Array` with mixed types     | Guaranteed layout vs V8 guessing         |
| Accessing one field of many objects | SoA (separate arrays per field)  | AoS (array of full structs)          | Cache lines carry useful data vs waste   |
| Map lookup pattern                  | Small maps, or arrays with index | Large maps with pointer-heavy values | Sequential scan beats random for small N |

### Rules of thumb

1. **Sequential beats random** — always. Iterating a slice is faster than chasing pointers, even if the pointer structure has better big-O complexity.
2. **Values beat pointers** — `[]Thing` is faster to iterate than `[]*Thing`. Use pointers when you need shared ownership or nil semantics, not by default.
3. **Pack hot data together** — if you always access three fields together, put them in the same struct. If you iterate millions of items but only read one field, consider splitting it out.
4. **Know your working set** — if your hot data fits in L1/L2, you're in good shape. If it spills to RAM, look for ways to shrink it or process it in cache-sized chunks.
5. **Measure, don't guess** — use `go test -bench`, `perf stat` (Linux), or language-specific profilers. Cache effects are hard to predict from code alone.

### Profiling cache performance

On Linux, you can directly measure cache misses:

```sh
# Count cache misses for a Go benchmark
perf stat -e cache-references,cache-misses,L1-dcache-load-misses \
    go test -bench=BenchmarkSequential -count=1 ./...
```

In Go, you can also use the built-in profiler:

```sh
go test -bench=. -cpuprofile=cpu.prof ./...
go tool pprof cpu.prof
# In pprof: "top" shows where time is spent
# Memory-bound code will show time in runtime.memmove, pointer derefs
```

## What's Next

The CPU is fast and caches help feed it data. But modern machines have multiple cores, and when multiple cores access the same data, things get complicated. The next lesson covers concurrency at the hardware level — how cores share (or don't share) memory, what happens when two goroutines write to the same variable, and why Go's memory model exists.
