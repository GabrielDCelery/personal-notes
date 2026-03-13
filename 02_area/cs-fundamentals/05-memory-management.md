# Memory Management

In the previous lessons we kept saying "this goes on the stack" or "this escapes to the heap" without fully explaining what that means. This lesson covers where data actually lives, how each region works, and why every language makes different tradeoffs around memory management.

## The Two Regions: Stack and Heap

Every program has two main areas of memory for storing data. They serve different purposes and have very different performance characteristics.

### The Stack

The stack is a contiguous region of memory (~1 MB on most systems) allocated per thread by the OS. It stores **local variables, function arguments, and return addresses**. It's called a "stack" because it works like a stack of plates — you can only add or remove from the top.

Every time you call a function, a new **frame** gets pushed onto the stack. When the function returns, the frame is popped off.

```go
func main() {
    result := add(3, 4)
    fmt.Println(result)
}

func add(a, b int) int {
    sum := a + b
    return sum
}
```

```
Stack (grows downward in memory):

When main() is running:
┌──────────────────┐
│ main's frame     │
│   result = ???   │
└──────────────────┘ ← RSP (stack pointer)

When main() calls add(3, 4):
┌──────────────────┐
│ main's frame     │
│   result = ???   │
├──────────────────┤
│ add's frame      │
│   a = 3          │
│   b = 4          │
│   sum = 7        │
│   return to: main│
└──────────────────┘ ← RSP moves down

When add() returns:
┌──────────────────┐
│ main's frame     │
│   result = 7     │
└──────────────────┘ ← RSP moves back up, add's frame is gone
```

Stack allocation is essentially free — it's just moving the stack pointer (RSP) down by a few bytes. Deallocation is just moving it back up. No bookkeeping, no garbage collector, no searching for free space.

The top of the stack is accessed so frequently (every function call, every local variable) that it's almost always in L1 cache. This makes stack access extremely fast — ~1 ns instead of the ~100 ns you'd pay for a cache miss to RAM.

### The Heap

The heap is a large region of memory (gigabytes) shared by all threads in a process. It stores data that needs to **outlive the function that created it** — anything that can't just disappear when the function returns.

```go
func newUser() *User {
    u := User{Name: "Alice"}
    return &u  // pointer escapes — caller needs this data after newUser() returns
}
// If u were on the stack, it would be gone when newUser() returns.
// The pointer would point to dead memory. So the compiler puts u on the heap.
```

Heap allocation is expensive compared to the stack:

```
Stack allocation:
  SUB RSP, 8        ← move stack pointer down 8 bytes. Done. (~0.3 ns)

Heap allocation:
  1. Ask the allocator for 8 bytes
  2. Allocator searches its free lists for a suitable block
  3. May need to request more memory from the OS (mmap syscall)
  4. Returns a pointer to somewhere in the heap
  5. Later: garbage collector must find and free this memory
  Cost: ~25-100 ns per allocation, plus GC overhead later
```

### Stack vs Heap Summary

|                 | Stack                                      | Heap                                           |
| --------------- | ------------------------------------------ | ---------------------------------------------- |
| What goes there | Local variables, function args             | Data that outlives the function                |
| Allocation cost | ~0.3 ns (move stack pointer)               | ~25-100 ns (find free space)                   |
| Deallocation    | Automatic — frame disappears on return     | GC must find and free it (or manual free)      |
| Size            | ~1 MB per thread (fixed on most languages) | Gigabytes, grows as needed                     |
| Cache behaviour | Top always in L1 — contiguous, hot         | Unpredictable — scattered across memory        |
| Thread safety   | Each thread has its own stack              | Shared between threads — needs synchronisation |

## Stack Overflow

The stack has a fixed size. If your function calls go too deep, you run out of space. The OS places a **guard page** (protected memory) right after the stack. Write past it and the hardware triggers a fault — your program crashes.

```
func factorial(n int) int {
    if n <= 1 { return 1 }
    return n * factorial(n-1)    // each call adds a frame
}

factorial(5)       // 5 frames, ~40 bytes each — fine
factorial(500000)  // 500,000 frames — stack overflow
```

```
┌──────────────────┐
│ frame 50,000     │
├──────────────────┤
│ ...              │
├──────────────────┤
│ frame 2          │
├──────────────────┤
│ frame 1          │ ← 1 MB limit
├──────────────────┤
│ GUARD PAGE       │ ← write here = crash
└──────────────────┘
```

**Why doesn't the compiler prevent this?** Because the depth usually depends on runtime data. The compiler can't know what `n` will be when you call `factorial(n)`.

### Stack sizes per language

| Runtime               | Default stack size   | Can it grow?                                         |
| --------------------- | -------------------- | ---------------------------------------------------- |
| Linux (C, Rust, Java) | ~1-8 MB per thread   | No — fixed at creation, overflow = crash             |
| Go goroutines         | ~4 KB initial        | Yes — runtime grows it automatically (up to 1 GB)    |
| Node.js               | ~1 MB                | No — overflow = crash                                |
| Python                | ~8 MB (configurable) | No — has a recursion limit (~1000) before hitting it |

### Go's growable stack

Go takes a unique approach. Goroutines start with a tiny ~4 KB stack. When the runtime detects it's about to overflow, it:

1. Allocates a new, larger stack (2x the size)
2. Copies everything from the old stack to the new one
3. Updates all pointers that referenced the old stack
4. Frees the old stack

This is why goroutines are so cheap to create — 4 KB vs 1 MB for an OS thread. You can have millions of goroutines because most of them have shallow call stacks and never need to grow.

The tradeoff: occasionally a goroutine pays the cost of copying its entire stack. But this happens rarely and the copy is to contiguous memory, which is fast (sequential write, prefetcher-friendly — lesson 2).

## Escape Analysis — Who Decides Stack vs Heap?

The compiler decides. It looks at each variable and asks: "does this value's reference leave the function?" If yes, it **escapes** to the heap. If no, it stays on the stack.

```go
func noEscape() int {
    x := 42          // x is used locally and returned by value
    return x          // x stays on the stack — cheap
}

func escapes() *User {
    u := User{Name: "Alice"}
    return &u         // u's pointer leaves the function
}                     // u MUST go on the heap — caller needs it
```

You can see the compiler's escape decisions:

```sh
go build -gcflags="-m" ./...
# output:
# ./main.go:10:2: moved to heap: u
```

### Common reasons things escape to the heap

| Pattern                  | Why it escapes                       |
| ------------------------ | ------------------------------------ |
| `return &x`              | Pointer leaves the function          |
| Storing in an interface  | Interface holds a pointer internally |
| Captured by a closure    | The closure outlives the function    |
| Too large for the stack  | Very large arrays/structs            |
| `append()` grows a slice | New backing array allocated on heap  |

### Why this matters

Every heap allocation is work for the garbage collector later. Code that keeps things on the stack:

- Allocates faster (~0.3 ns vs ~25-100 ns)
- Frees automatically (no GC)
- Has better cache locality (stack is contiguous and hot)

This doesn't mean you should avoid pointers — it means understanding the cost when you use them.

## Garbage Collection — Cleaning Up the Heap

When data goes on the heap, something needs to free it eventually. Languages take different approaches.

### Manual memory management (C, C++)

You allocate and free memory yourself:

```c
int *data = malloc(1024);    // ask the OS for 1024 bytes
// use data...
free(data);                   // give it back
```

If you forget `free()` → memory leak (your program slowly eats all RAM).
If you `free()` too early → use-after-free (reading garbage, security vulnerability).
If you `free()` twice → double free (corrupts the allocator, crash or security vulnerability).

Manual management is the fastest (zero overhead at runtime) and the most dangerous. Most security vulnerabilities in C/C++ programs are memory bugs.

### Reference counting (Python, Swift, Objective-C)

Each heap object has a counter: "how many things point to me?" When the counter reaches zero, the object is freed immediately.

```python
a = [1, 2, 3]     # list created, refcount = 1
b = a              # refcount = 2 (a and b both point to it)
a = None           # refcount = 1
b = None           # refcount = 0 → freed immediately
```

```
Object on the heap:
┌─────────────────┐
│ refcount: 2     │ ← a and b both point here
│ type: list      │
│ data: [1, 2, 3] │
└─────────────────┘

b = None → refcount drops to 0 → freed
```

**Pros**: deterministic — objects are freed the instant nothing references them. No pauses.

**Cons**:

- Every pointer assignment updates the refcount (overhead on every operation)
- **Cycles** can't be collected — if A points to B and B points to A, both have refcount 1 forever, even if nothing else references them. Python solves this with a separate cycle detector that runs periodically.
- Refcount updates on shared objects need to be atomic in multi-threaded code (this is why Python has the GIL — making every refcount update atomic would be too slow)

### Tracing garbage collection (Go, Java, JavaScript)

The GC periodically scans the heap to find objects that are no longer reachable from any live variable. It doesn't count references — it **traces** from roots (stack variables, globals) and marks everything it can reach. Anything unmarked is garbage.

```
Roots (stack, globals):
  main.users ──→ User{"Alice"} ──→ Address{"123 Main St"}
  main.cache ──→ Map{...}

  (nothing points to User{"Bob"} anymore)

GC traces from roots:
  ✓ User{"Alice"}     — reachable from main.users
  ✓ Address{"123..."}  — reachable through Alice
  ✓ Map{...}          — reachable from main.cache
  ✗ User{"Bob"}       — unreachable → garbage → freed
```

**Pros**: handles cycles naturally (unreachable cycles get collected), no per-assignment overhead.

**Cons**: GC pauses — the collector needs to stop or slow your program while it scans. Modern GCs minimise this but can't eliminate it entirely.

### Go's garbage collector

Go uses a **concurrent, tri-color mark-and-sweep** collector. The key design goal: keep GC pauses under 1 ms, even with gigabytes of heap.

How it works, simplified:

```
1. Stop the world briefly (~10-50 us)
   — enable a write barrier (tracks pointer changes during GC)

2. Mark phase (concurrent — runs alongside your goroutines)
   — trace from roots, mark all reachable objects
   — your code keeps running, write barrier catches new pointer changes

3. Stop the world briefly again (~10-50 us)
   — finish marking, disable write barrier

4. Sweep phase (concurrent)
   — walk the heap, free unmarked objects
   — your code keeps running
```

The pauses are the two "stop the world" moments — typically 10-50 microseconds. The actual scanning runs concurrently with your code. This is why Go can handle low-latency network services — sub-millisecond GC pauses.

The tradeoff: the concurrent GC uses CPU time. While the GC is marking, some of your cores are scanning the heap instead of running your code. Go typically uses about 25% of CPU for GC during collection.

### Java's garbage collectors

Java has multiple GC implementations you can choose from:

| GC           | Optimised for                     | Pause times                      |
| ------------ | --------------------------------- | -------------------------------- |
| G1 (default) | Balance of throughput and latency | ~10-200 ms                       |
| ZGC          | Low latency                       | < 1 ms (similar to Go)           |
| Shenandoah   | Low latency                       | < 1 ms                           |
| Parallel GC  | Maximum throughput                | Longer pauses, higher throughput |

Java gives you the choice because different applications have different needs. A batch processing job doesn't care about pauses — it wants maximum throughput. A trading system needs sub-millisecond pauses.

### JavaScript's garbage collector (V8)

V8 uses a **generational** collector based on the observation that most objects die young:

```
Nursery (young generation) — small, collected frequently:
  Most objects are created and become garbage within milliseconds.
  Scanning a small space is fast.

Old generation — large, collected rarely:
  Objects that survive several nursery collections get promoted here.
  These are likely long-lived (caches, connection pools, etc.)
```

This is why creating lots of short-lived objects in JavaScript isn't as expensive as you'd think — the nursery is small and fast to scan. But if you keep accumulating objects in the old generation, full GC pauses get longer.

### Ownership (Rust)

Rust takes a completely different approach: **no runtime garbage collection at all**. Instead, the compiler tracks ownership at compile time.

```rust
fn main() {
    let s = String::from("hello");   // s owns the string
    let t = s;                        // ownership moves to t
    // println!("{}", s);             // compile error — s no longer owns anything
    println!("{}", t);                // fine — t owns it
}                                     // t goes out of scope → string is freed
```

Rules enforced at compile time:

1. Every value has exactly one owner
2. When the owner goes out of scope, the value is freed
3. You can have either one mutable reference OR any number of immutable references — never both

```rust
fn main() {
    let mut data = vec![1, 2, 3];

    let r1 = &data;        // immutable borrow — fine
    let r2 = &data;        // another immutable borrow — fine
    // let r3 = &mut data; // compile error — can't mutate while borrowed immutably

    println!("{:?} {:?}", r1, r2);
    // r1 and r2 are no longer used after this point

    let r3 = &mut data;    // now this is fine — no other borrows active
    r3.push(4);
}
```

**Pros**: zero runtime overhead — no GC pauses, no refcount updates, memory is freed at exactly the right moment. As fast as manual C, but the compiler prevents all the bugs (use-after-free, double-free, data races).

**Cons**: the compiler is strict. Code that would be fine in Go or Java gets rejected. You spend time satisfying the borrow checker. The learning curve is steep.

## Comparison Across Languages

| Language   | Strategy                            | Runtime cost                             | Safety             | Ease of use                            |
| ---------- | ----------------------------------- | ---------------------------------------- | ------------------ | -------------------------------------- |
| C          | Manual malloc/free                  | Zero                                     | You're on your own | Easy to write, hard to write correctly |
| C++        | Manual + smart pointers (RAII)      | Near zero                                | Better than C      | Complex — many ways to manage memory   |
| Rust       | Ownership + borrow checker          | Zero                                     | Compiler-enforced  | Steep learning curve                   |
| Go         | Tracing GC (concurrent)             | ~25% CPU during collection, <1 ms pauses | Safe               | Simple — just allocate, GC handles it  |
| Java       | Tracing GC (multiple options)       | Varies by GC choice                      | Safe               | Simple — just allocate, GC handles it  |
| JavaScript | Generational GC (V8)                | Moderate                                 | Safe               | Simple — no control over allocation    |
| Python     | Reference counting + cycle detector | Overhead per assignment                  | Safe               | Simple — no control over allocation    |
| Swift      | Reference counting (ARC)            | Overhead per assignment                  | Safe               | Moderate — weak references for cycles  |

## Practical Summary

1. **Stack is fast, heap is flexible** — stack allocation is essentially free (~0.3 ns) and automatically cleaned up. Heap allocation costs ~25-100 ns and requires garbage collection. Keep data on the stack when you can.
2. **Escape analysis decides** — in Go, Java, and other GC languages the compiler decides stack vs heap. Return a pointer and it escapes to the heap. Return a value and it stays on the stack.
3. **GC is not free** — modern GCs are impressive (Go: <1 ms pauses, Java ZGC: <1 ms) but they use CPU time. High-allocation code creates GC pressure. Reducing allocations is often the biggest performance win.
4. **Rust trades convenience for zero cost** — no GC, no runtime overhead, but the borrow checker demands more from you at compile time.
5. **Python's refcounting is why the GIL exists** — every pointer assignment updates a refcount. Making that atomic for multi-threading would slow down all single-threaded code. The GIL is the compromise.
6. **Know your language's model** — when you understand why your language chose its memory strategy, you stop fighting it and start writing code that works with it.
