# Concurrency and the OS

In lessons 1 and 2 we looked at a single core executing instructions and accessing memory. But modern CPUs have multiple cores, and real programs need to do multiple things at once. This lesson covers how the OS manages execution, how cores share (or fight over) memory, and why every language has a different concurrency model built on the same hardware.

## What Is a Thread, Really?

A thread is not a physical thing on the CPU. It's a **bookkeeping structure in the OS** — a data structure that tracks everything needed to pause and resume a sequence of instructions.

A thread consists of:

1. **A set of saved register values** — the program counter (where in the code it is), stack pointer, and the ~16 general-purpose registers from lesson 1. When the thread is running, these values are in the CPU's physical registers. When the thread is paused, they're saved to a struct in RAM.
2. **A stack** — a region of memory (~1 MB) allocated by the OS for this thread's local variables, function call frames, and return addresses. Each thread gets its own stack so function calls in one thread don't overwrite another's local variables.
3. **Metadata** — thread ID, priority, state (running/sleeping/ready), which process it belongs to.

```
Thread (as an OS data structure in RAM):
┌──────────────────────────────┐
│ Thread ID: 4821              │
│ State: RUNNING               │
│ Priority: normal             │
│                              │
│ Saved registers:             │
│   RIP = 0x4013  (next instr) │
│   RSP = 0x7ffd3a00  (stack)  │
│   RAX = 42                   │
│   RBX = 0                    │
│   ... (all 16)               │
│                              │
│ Stack: 0x7ffd3a00 - 0x7ffe3a00  (1 MB region in RAM)  │
│ Process: belongs to PID 1234 │
└──────────────────────────────┘
```

Think of a CPU core as a **desk with 16 slots** (registers). Only one person can sit at the desk at a time. The desk doesn't know or care who's sitting there — it just has slots with values in them.

A thread is a **notebook in a filing cabinet**. It records: "when I was last at the desk, the slots contained these values, and I was on page 47 of my work." The OS is the **office manager** — also a person who uses the same desk. Every few milliseconds the office manager taps the current worker on the shoulder, writes down what's in all the desk slots into that worker's notebook, files it away, pulls out a different notebook, loads those saved values back into the desk slots, and lets the new worker continue from where they left off.

**Why save the registers at all?** Because there's only one desk. If Thread A has `total = 42` in register RAX and the OS lets Thread B run without saving first, Thread B will overwrite RAX with its own values. When Thread A comes back, its `total` is gone. The saved registers are how a thread remembers where it was and what it was working on.

```
Thread A at the desk:        Thread A's notebook (in RAM):
  RAX = 42  (total)            "RAX was 42"
  RBX = 7   (counter)   →     "RBX was 7"
  RIP = 0x4013 (next line)     "I was at line 0x4013"

OS loads Thread B's notebook into the desk:
  RAX = 99  (connectionCount)
  RBX = 0
  RIP = 0x8020

Thread A's values are safe in its notebook.
When it's Thread A's turn again, the OS loads them back.
```

The OS itself runs on the same CPU — it's not watching from outside. It's just another piece of code that gets control when a hardware timer fires, does the notebook swap, and hands control back.

So when we say "each thread has its own registers" — we don't mean physical registers. We mean each thread has a saved copy of what the registers should contain when it's that thread's turn to run. There are only ~16 physical registers per core. If you have 100 threads, 99 of them have their register values sitting in RAM, and 1 has its values loaded into the actual CPU registers.

## Processes vs Threads

The OS gives you two abstractions for running code concurrently:

**A process** is an isolated program. It gets its own virtual address space — its own view of memory. Two processes cannot see each other's memory (unless they explicitly set up shared memory or communicate through pipes/sockets).

**A thread** is a unit of execution within a process. Threads in the same process share the same address space — they can all read and write the same variables. Each thread has its own stack and saved register values, but the heap is shared.

```
Process A                          Process B
┌─────────────────────┐           ┌─────────────────────┐
│ Virtual memory      │           │ Virtual memory      │
│ (isolated)          │           │ (isolated)          │
│                     │           │                     │
│ ┌───────┐ ┌───────┐ │           │ ┌───────┐ ┌───────┐ │
│ │Thread1│ │Thread2│ │           │ │Thread1│ │Thread2│ │
│ │stack  │ │stack  │ │           │ │stack  │ │stack  │ │
│ │regs   │ │regs   │ │           │ │regs   │ │regs   │ │
│ └───────┘ └───────┘ │           │ └───────┘ └───────┘ │
│                     │           │                     │
│ Shared heap         │           │ Shared heap         │
│ Shared code         │           │ Shared code         │
└─────────────────────┘           └─────────────────────┘
     Cannot see each other's memory
```

|                     | Process                                                    | Thread                                          |
| ------------------- | ---------------------------------------------------------- | ----------------------------------------------- |
| Memory              | Own address space                                          | Shares with other threads in same process       |
| Creation cost       | Expensive (~ms) — OS sets up page tables, file descriptors | Cheap (~us) — just a new stack and register set |
| Communication       | IPC (pipes, sockets, shared memory)                        | Direct — read/write shared variables            |
| Isolation           | Full — one process crashing doesn't kill another           | None — one thread corrupting memory affects all |
| Context switch cost | Expensive — must flush TLB, switch page tables             | Cheaper — same address space, no TLB flush      |

### What "virtual address space" means

Every process thinks it has the entire address space to itself. When process A writes to address `0x1000`, it's writing to **its** `0x1000`, which is a completely different physical memory location than process B's `0x1000`. The OS and CPU's MMU (Memory Management Unit) translate virtual addresses to physical addresses using page tables.

This is why processes are isolated — they literally cannot address each other's memory. The hardware enforces it.

Threads **within the same process** don't get this isolation because they share the same page table. If Thread 1 and Thread 2 both belong to Process A, then Thread 1 writing to `0x1000` and Thread 2 reading from `0x1000` are accessing the **same physical memory**. This is both the power and the danger of threads — they can share data easily, but one thread can corrupt another's data. Threads in different processes can't see each other's memory at all, just like the processes themselves.

## Context Switches — What Actually Happens

From lesson 1, you know the CPU has ~16 registers per core. A context switch is the OS swapping one thread's register state for another's.

### Thread context switch (same process)

```
1. Timer interrupt fires (hardware, typically every 1-10 ms)
2. CPU traps into kernel mode
3. Kernel saves Thread A's registers → Thread A's context block in RAM
   (RIP, RSP, RAX, RBX, ... all 16 general-purpose registers + flags)
4. Kernel picks next thread to run (scheduler decision)
5. Kernel loads Thread B's registers from Thread B's context block → CPU
6. Kernel returns to user mode
7. Thread B resumes exactly where it left off

Cost: ~1-10 us (microseconds)
```

The registers are saved and restored as an opaque blob — the OS doesn't know what your compiler put in RAX. It just saves and restores all of them.

### Process context switch

Same as above, plus:

```
8. Switch page tables (CR3 register on x86)
9. TLB (Translation Lookaside Buffer) is flushed
   — TLB caches virtual→physical address translations
   — new process has different translations, old ones are useless
   — next ~100-1000 memory accesses will be TLB misses (slow)

Cost: ~10-50 us (more expensive due to TLB flush)
```

**Why does a process switch need a TLB flush but a thread switch doesn't?** The TLB caches "virtual address X maps to physical address Y." Threads in the same process share the same page table — the same mappings. So when you swap Thread 1 for Thread 2 in the same process, address `0x1000` still maps to the same physical location. The TLB entries are still correct.

When you switch to a different process, the mappings are completely different. Process A's `0x1000` and Process B's `0x1000` point to different physical memory. If the TLB still had Process A's mappings, Process B would read Process A's data. So the CPU flushes the entire TLB, and Process B starts cold — paying ~100 ns per memory access until the TLB warms up again.

### Voluntary vs involuntary

Not all context switches are caused by the timer interrupt:

| Type            | Cause                             | Example                                |
| --------------- | --------------------------------- | -------------------------------------- |
| **Involuntary** | Timer interrupt — OS preempts you | Thread used its time slice             |
| **Voluntary**   | Thread blocks on something        | Waiting for I/O, mutex, channel, sleep |

Voluntary switches are more efficient because the thread is already in a known state (it called a syscall). Involuntary switches happen at arbitrary points in your code.

### Why bother with multiple threads if swapping is expensive?

Two reasons:

**1. Multiple cores mean no swapping.** If you have 4 cores and 4 threads, each thread gets its own core — its own desk. No swapping needed. A single-threaded program on a 4-core machine leaves 3 cores sitting idle.

```
Single thread on a 4-core machine:
  Core 0: [your program]
  Core 1: idle
  Core 2: idle
  Core 3: idle
  → 75% of your CPU is wasted

Four threads:
  Core 0: [thread 1]    no swapping between them —
  Core 1: [thread 2]    each thread has its own desk
  Core 2: [thread 3]
  Core 3: [thread 4]
```

**2. I/O freezes the thread.** When a thread calls `read()` on a socket and data hasn't arrived, the OS puts it to sleep. If that's your only thread, your entire program is frozen until the data arrives.

```
Single thread:
  handle request → read from database → BLOCKED (5 ms)
                                         entire program frozen

Two threads:
  Thread 1: handle request A → read from database → BLOCKED
  Thread 2: handle request B → still running, serves response
```

Swapping only becomes a problem when you create **way more threads than cores**. 4 threads on 4 cores = no swapping. 10,000 threads on 4 cores = constant swapping. That's exactly the problem goroutines solve.

### Who decides how many threads a program gets?

The OS doesn't decide — **your program does**. Every process starts with one thread (the main thread). After that, your code creates more.

```go
// Go: the runtime creates OS threads automatically (~1 per core)
// You create goroutines, not OS threads
go handleRequest(conn)  // does NOT create an OS thread
```

```java
// Java: you create OS threads explicitly
new Thread(() -> doWork()).start();  // creates 1 OS thread
```

```javascript
// Node.js: 1 thread for your code + 4 worker threads for disk I/O
new Worker("./heavy-task.js"); // explicitly creates 1 more
```

There's no OS limit saying "this process gets 4 threads." A process can create as many as it wants until it runs out of memory (~1 MB stack per thread). The OS just schedules whatever threads exist across the available cores.

A typical system has thousands of threads total, but most are **sleeping** — waiting for I/O, timers, or user input, using zero CPU. The OS only swaps between threads that are actually runnable. 2,000 threads sounds like a lot, but if 1,950 are sleeping, the OS is only juggling ~50 across your cores.

### How does the OS know when a sleeping thread should wake up?

It doesn't poll. It doesn't loop through threads checking "are you ready yet?" It's **event-driven** — the hardware tells the OS.

When a thread calls `read(socket)` and there's no data, the kernel puts the thread on a **wait queue** attached to that socket. The thread is completely ignored until:

```
1. Network card receives data
2. Network card sends a hardware interrupt to the CPU
   (an electrical signal — the CPU can't ignore it)
3. CPU stops whatever it's doing, jumps to the kernel's interrupt handler
4. Kernel: "data arrived on socket 47"
5. Kernel checks socket 47's wait queue: "Thread 812 is waiting on this"
6. Kernel moves Thread 812 from sleeping → runnable
7. Next time the scheduler runs, Thread 812 can be picked for a core
```

Same for timers — `sleep(5 seconds)` doesn't mean the OS checks every millisecond. The kernel sets a hardware timer. When it fires, the CPU gets an interrupt and the kernel wakes the thread.

### Don't interrupts disrupt whatever the CPU is running?

Yes, but it's a quick tap on the shoulder, not a full context switch. The interrupt handler borrows the CPU for a few microseconds and gives it back. Your thread never notices.

```
1. CPU saves a handful of registers (RIP, RSP, flags) — just enough to get back
2. Jumps to the kernel's interrupt handler
3. Handler needs to use a couple of registers for its work, so it
   saves their current values to the stack first (PUSH), does its work,
   then restores them (POP)
4. CPU jumps back to exactly where it was — your thread's registers are intact
```

The stack (the top of it) is accessed so frequently that it's almost always sitting in L1 cache — so saving and restoring a few registers costs ~1 ns per register, not the ~100 ns you'd pay for a RAM access.

A quiet system gets a few hundred interrupts per second. A busy server might get thousands. But each takes microseconds, so even 1,000 interrupts/sec is ~0.5% of CPU time.

### Does the OS itself have threads?

Yes. The kernel runs on the same CPU as everything else, but in a privileged **kernel mode** where it can access all physical memory and hardware directly.

```
Your programs (user mode):
  Process A → virtual memory → page tables → physical RAM
  Process B → virtual memory → page tables → physical RAM

OS kernel (kernel mode):
  Can see all physical RAM directly
  Can modify any process's page tables
  Has its own threads for background work
```

The kernel has threads for housekeeping — managing memory, handling interrupts, balancing load across cores. On Linux you can see them:

```sh
ps aux | grep '\[.*\]'
# [kworker/0:1]   — general kernel work on core 0
# [kswapd0]       — swaps memory pages to disk when RAM is full
# [migration/2]   — moves threads between cores for balance
```

When your program makes a syscall (like `read()`), your thread doesn't get swapped out. Instead, the CPU switches from user mode to kernel mode — your thread temporarily runs kernel code, does the privileged work, then switches back to user mode. Same thread, same registers, no swap. This costs ~200 ns (mode switch) vs ~10 us (full context switch).

## Multiple Cores — True Parallelism

A 4-core CPU has 4 independent pipelines, each with their own registers. Four threads can execute truly simultaneously — not time-sliced, actually running at the same instant on different hardware.

```
Core 0              Core 1              Core 2              Core 3
┌────────────────┐  ┌────────────────┐  ┌────────────────┐  ┌────────────────┐
│ Thread A       │  │ Thread B       │  │ Thread C       │  │ Thread D       │
│ own regs       │  │ own regs       │  │ own regs       │  │ own regs       │
│ L1i 32K(instr) │  │ L1i 32K(instr) │  │ L1i 32K(instr) │  │ L1i 32K(instr) │
│ L1d 32K(data)  │  │ L1d 32K(data)  │  │ L1d 32K(data)  │  │ L1d 32K(data)  │
│ L2  256K       │  │ L2  256K       │  │ L2  256K       │  │ L2  256K       │
└───────┬────────┘  └───────┬────────┘  └───────┬────────┘  └───────┬────────┘
        └───────────────────┴───────────────────┴───────────────────┘
                              Shared L3 cache
                              Shared RAM
```

### Cache structure per core

L1 is not a single cache — it's two separate 32 KB caches, so 64 KB total per core. L2 and L3 are unified (instructions and data mixed together).

```
Per core:
  L1i: 32 KB   ← instructions only  ┐ separate hardware, can be
  L1d: 32 KB   ← data only          ┘ accessed simultaneously

  L2:  256 KB  ← both instructions and data, mixed together

Shared across all cores:
  L3:  16 MB   ← both instructions and data, mixed together
```

The split only exists at L1 because the CPU needs to fetch instructions and data simultaneously on every cycle — two separate caches let it do both without contention. By L2 and L3 the access frequency is lower (only hit on L1 misses), so a single unified cache is simpler and adapts naturally — a program with lots of code but little data uses more of the cache for instructions, and vice versa.

### Cache lines — the unit of caching

The cache doesn't load individual bytes or variables. It loads fixed-size **64-byte blocks** called cache lines. This is the smallest unit the cache can load, store, or evict. Think of it like buying eggs — you can't buy one, you get the whole carton.

```
You want 8 bytes at address 0x1000

Cache fetches 64 bytes: addresses 0x0FC0 through 0x0FFF
├── your 8 bytes ──┤── 56 bytes of neighbours you didn't ask for ──┤
```

This works well because programs almost always access nearby memory right after (arrays, struct fields, sequential code) — **spatial locality**. Loading 64 bytes at once means the next few accesses are likely already cached.

**Why 64 bytes specifically?** It's a tradeoff. Too small (8 bytes) and you'd miss on every adjacent access, paying full RAM latency each time. Too large (4 KB) and you waste cache space on data you never touch, evictions happen more often, and cache coherence gets worse — any write to any byte in that block invalidates the whole thing on other cores. 64 bytes has been the sweet spot on x86 since ~2000. Some architectures differ — Apple M-series uses 128 bytes for certain cache levels.

**How many cache lines fit in each level:**

```
L1d:  32 KB / 64 bytes =    512 cache lines   (per core)
L2:  256 KB / 64 bytes =  4,096 cache lines   (per core)
L3:   16 MB / 64 bytes = 262,144 cache lines  (shared)
```

A single core can hold about 512 data "blocks" in its fastest memory. If your working data fits in those 512 lines, everything runs at ~1 ns. If it doesn't, you spill to L2 (~5 ns), L3 (~20 ns), or RAM (~100 ns).

**Does small data guarantee L1 hits?** Mostly, but not perfectly. Even if your data fits in 32 KB, a few things displace cache lines: OS timer interrupts run handler code that evicts some of your lines, syscalls transition to kernel code that displaces L1 contents, and set associativity means a given address can only map to ~8 specific slots, so collisions can force evictions even when the cache isn't full. But a tight loop over a small array is the ideal case — almost entirely L1 hits.

Each core has private L1 and L2 caches. L3 and RAM are shared. This creates a problem: what happens when Core 0 and Core 1 both cache the same memory address, and Core 0 writes to it?

This only happens with **threads in the same process**. Remember — threads share the same memory (same page table, same heap). If Thread A on Core 0 and Thread B on Core 1 both belong to the same process, they can both access `var counter` — and both cores will cache it in their private L1. Two threads from **different processes** can't see each other's memory, so their caches hold completely unrelated data and there's no conflict.

```
Same process, two threads on different cores:
  Core 0 (Thread 1):  counter = 42   ← cached in Core 0's L1
  Core 1 (Thread 2):  reads counter  ← cached in Core 1's L1
  → Two copies of the same data. What if Core 0 changes it?

Different processes on different cores:
  Core 0 (Process A): address 0x1000 → physical 0xABC   ← cached
  Core 1 (Process B): address 0x1000 → physical 0xF00   ← cached
  → Different physical memory. No conflict. Caches don't interfere.
```

## Cache Coherence — Keeping Cores In Sync

When two cores cache the same memory address (because they're running threads from the same process), the hardware must keep them consistent. This is the **cache coherence protocol** (MESI on Intel, MOESI on AMD).

### The MESI protocol

Each cache line is in one of four states:

| State             | Meaning                                           | What happens on write              |
| ----------------- | ------------------------------------------------- | ---------------------------------- |
| **M** (Modified)  | This core has the only copy and it's been changed | Write freely — you own it          |
| **E** (Exclusive) | This core has the only copy, unchanged from RAM   | Write freely — transition to M     |
| **S** (Shared)    | Multiple cores have this cache line               | Must invalidate other copies first |
| **I** (Invalid)   | Stale — another core modified it                  | Must re-fetch from L3/RAM          |

### What happens when two cores write to the same variable

```
Initial state: counter = 0, cached on both Core 0 and Core 1 (Shared)

Core 0: counter++
  1. Core 0 wants to write → cache line is Shared, can't write yet
  2. Core 0 sends "invalidate" message to Core 1
  3. Core 1 marks its copy as Invalid
  4. Core 1 sends acknowledgement
  5. Core 0 transitions to Modified, performs the write
  Cost: ~40-100 ns for the invalidation round-trip

Core 1: counter++
  1. Core 1 wants to read counter → its cache line is Invalid
  2. Core 1 requests data → Core 0 has it in Modified state
  3. Core 0 writes back to L3, transitions to Shared
  4. Core 1 gets the data, transitions to Shared
  5. Core 1 now wants to write → sends invalidate to Core 0
  6. Core 0 transitions to Invalid
  7. Core 1 transitions to Modified, performs the write
  Cost: another ~40-100 ns
```

This "cache line bouncing" between cores is the real cost of shared mutable state. It's not the lock itself that's slow — it's the cache coherence traffic underneath.

## False Sharing — The Hidden Performance Killer

False sharing happens when two threads access **different variables** that happen to sit on the **same cache line**. The hardware doesn't track individual variables — it tracks cache lines (64 bytes). If any byte in the line changes, the entire line gets invalidated on other cores.

```go
// These two counters are adjacent in memory — likely on the same cache line
type Counters struct {
    CounterA int64  // 8 bytes  ┐
    CounterB int64  // 8 bytes  ┘ both fit in one 64-byte cache line
}

var c Counters

// Goroutine 1: only touches CounterA
go func() {
    for i := 0; i < 1_000_000; i++ {
        c.CounterA++
    }
}()

// Goroutine 2: only touches CounterB
go func() {
    for i := 0; i < 1_000_000; i++ {
        c.CounterB++
    }
}()
```

These goroutines don't share any data logically. But physically, `CounterA` and `CounterB` are on the same cache line. Every increment by one goroutine invalidates the other's cache line, forcing a re-fetch. This can be **10-50x slower** than if they were on separate cache lines.

### The fix: padding

```go
type Counters struct {
    CounterA int64
    _pad     [56]byte  // push CounterB to the next cache line
    CounterB int64
}
```

Now each counter is on its own cache line. The goroutines no longer interfere with each other's caches. This is why you sometimes see `_ [64]byte` padding in performance-critical concurrent code.

You'll see this in Go's standard library — `sync.Pool` and the runtime scheduler use cache line padding internally.

## Cache-Friendly Access Patterns

Cache performance isn't only a concurrency concern. Even single-threaded code benefits hugely from accessing memory in patterns the cache is designed for.

### Hot paths

Not all code matters equally for performance. Most programs follow the 90/10 rule: 90% of the time is spent in 10% of the code. It doesn't matter if your program is 50 MB total if the inner loop that runs millions of times only touches 16 KB of data.

The hot path is usually obvious from the code structure:

- **Loops** — anything inside a loop that runs many times is hot. The tighter and more frequent, the hotter.
- **Nested loops** — the innermost loop is usually the hottest. A doubly nested loop over 1000 elements runs the inner body 1,000,000 times.
- **Request handlers in servers** — the code that runs on every incoming request is hot, the startup/config code is not.

```go
func main() {
    config := loadConfig()          // runs once — cold
    db := connectToDB()             // runs once — cold

    http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
        user := parseRequest(r)     // runs per request — hot
        result := queryDB(db, user) // runs per request — hot
        json.Encode(w, result)      // runs per request — hot
    })
}
```

For anything non-obvious, you profile. In Go:

```sh
go test -cpuprofile cpu.prof -bench .
go tool pprof cpu.prof
```

The profiler tells you exactly which functions consumed the most CPU time. The bottleneck is rarely where you expect — profile first, optimise second.

### Matrix traversal — loop order matters

In most languages (C, Go, Java), 2D arrays are stored **row-major** — each row is contiguous in memory:

```
matrix[0][0], matrix[0][1], matrix[0][2], ..., matrix[0][999]    ← row 0, contiguous
matrix[1][0], matrix[1][1], matrix[1][2], ..., matrix[1][999]    ← row 1, contiguous
```

Walking within a row is fast — the inner loop moves sequentially through memory:

```go
// FAST — sequential memory access, cache-friendly
for i := 0; i < 1000; i++ {        // each row
    for j := 0; j < 1000; j++ {    // walk columns sequentially
        matrix[i][j] += 1          // 8 bytes apart in memory
    }
}
```

The first access loads a cache line with 8 consecutive `int64` values. The next 7 accesses are free hits.

Swapping the loop order is slow — each access jumps a whole row:

```go
// SLOW — strided access, cache-unfriendly
for j := 0; j < 1000; j++ {        // each column
    for i := 0; i < 1000; i++ {    // walk rows — jumping 8KB each time
        matrix[i][j] += 1          // 8000 bytes apart in memory
    }
}
```

Each access lands on a different cache line. You load 64 bytes but only use 8, then jump to a completely different part of memory. Same computation, same result, just the loop order swapped — easily a **5-10x** performance difference on large matrices.

### JavaScript doesn't have real 2D arrays

In JavaScript, a "2D array" is an array of pointers to separate row objects scattered across the heap:

```javascript
const matrix = [];
for (let i = 0; i < 1000; i++) {
    matrix[i] = new Array(1000);
}

// matrix[i][j] = two lookups:
//   1. follow pointer to find the row object
//   2. index into the row
```

Walking within a row is still better than jumping between rows (each row's data is contiguous within itself), but you're paying a pointer dereference per row either way.

For cache-friendly matrix operations in JavaScript, use a flat typed array:

```javascript
const matrix = new Float64Array(1000 * 1000);

// Access [i][j] as:
matrix[i * 1000 + j]  // one contiguous block, no pointers
```

`Float64Array` is a view over an `ArrayBuffer` — a fixed-size contiguous block of memory. This gives you the same layout as a C array. The tradeoff is it can't grow — if you need more space, you allocate a new bigger one and copy:

```javascript
const old = new Float64Array(1000);
const bigger = new Float64Array(2000);
bigger.set(old);  // copies old data into the beginning of bigger
```

This is the fundamental tradeoff: contiguous memory is fast but can't grow in place. A regular JavaScript array can grow because it uses pointers and indirection, which is exactly what makes it slower for sequential access.

## Atomics — Hardware-Level Synchronisation

When two cores need to safely modify the same variable, the CPU provides **atomic instructions**. These are single instructions that the hardware guarantees will complete without interruption.

```go
// NOT safe — two cores can read the same value, both increment, both write back
// Result: counter goes up by 1 instead of 2
counter++

// Safe — the CPU locks the cache line for the duration of the operation
atomic.AddInt64(&counter, 1)
```

What `atomic.AddInt64` compiles to on x86:

```asm
LOCK XADD [address], 1
```

The `LOCK` prefix tells the CPU: "take exclusive ownership of this cache line before executing, and don't let anyone else touch it until I'm done." Under the hood, it's the same MESI protocol — transition to Modified, but with a guarantee that no other core can intervene.

### Cost of atomics

| Operation                                               | Approximate cost |
| ------------------------------------------------------- | ---------------- |
| Normal read/write (no contention)                       | ~1 ns            |
| Atomic on uncontended cache line                        | ~5-10 ns         |
| Atomic on contended cache line (bouncing between cores) | ~40-100 ns       |

Atomics are fast when there's no contention. The cost comes from cache coherence traffic when multiple cores fight over the same cache line.

## Memory Ordering — Why Code Runs Out of Order

Both the compiler and the CPU can reorder operations for performance. This is invisible to single-threaded code but matters for multi-threaded:

```go
// You write:
x = 42
ready = true

// The compiler might reorder to:
ready = true
x = 42

// Or the CPU might execute stores out of order
// Another core could see ready=true but x=0
```

This happens because:

1. **Compiler reordering** — the compiler is free to reorder independent operations if it produces the same single-threaded result
2. **CPU store buffers** — writes go to a buffer first, then drain to cache in whatever order is efficient. Another core reading from RAM/cache might see stale data

### Memory barriers (fences)

A memory barrier is an instruction that prevents reordering across it:

```
x = 42
MFENCE          ← all stores before this are visible before any stores after
ready = true
```

You rarely write barriers directly. Atomic operations and mutexes include the appropriate barriers automatically.

### Memory ordering in practice

Every language provides guarantees about when writes become visible to other threads:

| Mechanism                      | Ordering guarantee                                                               |
| ------------------------------ | -------------------------------------------------------------------------------- |
| Regular variable               | None — can be reordered, may never be visible to other threads                   |
| `atomic.Store` / `atomic.Load` | Happens-before relationship — if you see the store, you see everything before it |
| `sync.Mutex` Lock/Unlock       | Everything inside the lock is visible to the next thread that locks it           |
| Channel send/receive           | Send happens-before the corresponding receive                                    |

This is Go's **memory model**. Every language has one. Java's is defined in the JLS. C++ has `std::memory_order`. JavaScript doesn't need one for regular code because it's single-threaded (but SharedArrayBuffer has Atomics).

## Mutexes — What They Actually Do

A mutex (mutual exclusion lock) combines atomics and OS scheduling:

```go
var mu sync.Mutex
var balance int

func deposit(amount int) {
    mu.Lock()       // acquire exclusive access
    balance += amount
    mu.Unlock()     // release
}
```

### What Lock() does under the hood

```
1. Try to atomically set the lock state from 0 (unlocked) to 1 (locked)
   — Uses an atomic compare-and-swap (CAS) instruction
   — If successful: you own the lock, proceed

2. If the CAS fails (someone else holds the lock):
   a. Spin briefly — loop trying the CAS a few more times
      (the holder might release it in nanoseconds, cheaper than a syscall)
   b. If still locked after spinning: make a syscall to the OS
      — OS puts this goroutine/thread to sleep
      — OS wakes it up when the lock is released
      — This is the expensive path: context switch + syscall overhead
```

This is called a **hybrid mutex** or **adaptive mutex**. Go's `sync.Mutex` does this. Java's `synchronized` does this. The spin phase avoids expensive syscalls for short critical sections.

### Why locks are "slow"

It's not the lock instruction itself — it's what it causes:

| Cost                     | Source                                                                    |
| ------------------------ | ------------------------------------------------------------------------- |
| The CAS instruction      | ~5-10 ns (atomic on uncontended line)                                     |
| Cache line bouncing      | ~40-100 ns if another core held the lock recently                         |
| Spinning (wasted cycles) | Variable — burns CPU waiting                                              |
| Syscall + context switch | ~1-10 us if the thread has to sleep                                       |
| Serialisation            | While one thread holds the lock, all others wait — parallelism drops to 1 |

The biggest cost is usually **serialisation** — a lock turns parallel code into sequential code for the duration of the critical section.

## How Languages Map to These Primitives

Every language's concurrency model is built on OS threads, with different abstractions on top.

### OS threads (Java, C, C++, Rust)

One language-level thread = one OS thread. You get true parallelism but each thread costs ~1 MB of stack and context switches are expensive (~1-10 us).

```
Java Thread  →  OS Thread  →  runs on a CPU core
Java Thread  →  OS Thread  →  runs on a CPU core
Java Thread  →  OS Thread  →  runs on a CPU core
```

This works well for dozens or hundreds of threads. At thousands, the context switch overhead dominates.

### Green threads / goroutines (Go)

Many lightweight goroutines multiplexed onto a small number of OS threads. The Go runtime acts as a userspace scheduler.

```
Goroutine 1 ─┐
Goroutine 2 ─┤→ OS Thread (P0) → Core 0
Goroutine 3 ─┘
Goroutine 4 ─┐
Goroutine 5 ─┤→ OS Thread (P1) → Core 1
Goroutine 6 ─┘
... millions of goroutines, handful of OS threads
```

|                | OS thread         | Goroutine                           |
| -------------- | ----------------- | ----------------------------------- |
| Stack size     | ~1 MB fixed       | ~4 KB initial, grows as needed      |
| Creation cost  | ~1 ms (syscall)   | ~1 us (userspace allocation)        |
| Context switch | ~1-10 us (kernel) | ~100-200 ns (userspace, no syscall) |
| Count limit    | Thousands         | Millions                            |

Goroutine context switches are cheap because Go's runtime does them in userspace — it saves and restores registers without entering the kernel. The runtime knows when a goroutine blocks (channel, I/O, mutex) and swaps in another goroutine on the same OS thread.

This is why Go is good at network services — you can have one goroutine per connection (millions) without worrying about thread overhead.

### Event loop (Node.js / JavaScript)

One OS thread runs all JavaScript. Asynchronous I/O is handled by the OS, and callbacks/promises execute when I/O completes.

```
Single OS Thread:
  ┌──────────────────────────────────┐
  │         Event Loop               │
  │                                  │
  │  1. Run current callback/handler │
  │  2. Check for completed I/O      │
  │  3. Run next callback            │
  │  4. Repeat                       │
  └──────────────────────────────────┘
           │
           │  I/O operations delegated to OS
           ├──→ file read (OS thread pool)
           ├──→ network recv (epoll/kqueue)
           └──→ DNS lookup (OS thread pool)
```

No parallelism for JavaScript code — only one function runs at a time. This means no data races, no need for locks, no cache coherence issues. The tradeoff: CPU-bound work blocks everything.

```javascript
// This blocks the entire event loop — no other request can be served
app.get("/slow", (req, res) => {
  const result = heavyComputation(); // blocks for 2 seconds
  res.send(result); // nothing else ran during those 2 seconds
});

// This is fine — I/O is non-blocking
app.get("/fast", async (req, res) => {
  const data = await db.query("SELECT ..."); // thread is free while waiting
  res.send(data);
});
```

This is why Node.js is good for I/O-heavy services (APIs, proxies) but bad for CPU-heavy work (image processing, compression). The single thread can juggle thousands of I/O-bound connections but can't parallelise computation.

Worker threads (`worker_threads` module) give you actual OS threads for CPU work, but they don't share memory freely — they communicate by message passing, similar to processes.

### Python — The GIL problem

Python has OS threads, but the Global Interpreter Lock (GIL) means only one thread can execute Python bytecode at a time.

```
Python Thread 1 ─┐                    ┌→ OS Thread → Core 0
Python Thread 2 ─┤→ GIL (one at a    │
Python Thread 3 ─┤   time only)  ─────┘
Python Thread 4 ─┘                      Cores 1, 2, 3: idle
```

The GIL exists because CPython's memory management (reference counting) is not thread-safe. Making every reference count update atomic would slow down single-threaded code — and most Python code is single-threaded.

**Python threads DO help for I/O** — when a thread is waiting for I/O, it releases the GIL and another thread can run. But for CPU-bound work, threads give zero parallelism.

Python's workarounds:

- `multiprocessing` — separate processes, each with their own GIL. True parallelism but expensive communication (serialisation/IPC).
- `asyncio` — event loop like Node.js. Good for I/O, doesn't help CPU.
- C extensions (NumPy) — release the GIL and run parallel C code. This is why NumPy operations parallelise fine.
- Python 3.13+ — experimental free-threaded mode (no GIL), still maturing.

## Concurrency Patterns and Their Costs

| Pattern                    | When it works                                 | Cost                                                 | Used by                            |
| -------------------------- | --------------------------------------------- | ---------------------------------------------------- | ---------------------------------- |
| Shared memory + locks      | Small critical sections, low contention       | Cache bouncing, serialisation                        | Go, Java, Rust, C                  |
| Message passing (channels) | Independent workers, pipeline processing      | Copying data between goroutines/threads              | Go (channels), Erlang, Rust (mpsc) |
| Event loop + callbacks     | I/O-heavy, many connections                   | No CPU parallelism                                   | Node.js, Python asyncio            |
| Fork/join (thread pools)   | CPU-bound parallelism, batch processing       | Thread creation overhead (amortised by pool)         | Java (ForkJoinPool), Rust (rayon)  |
| Lock-free data structures  | Ultra-high throughput, reader-heavy workloads | Complex, hard to get right, still has cache bouncing | All languages via atomics          |

### Choosing the right model

```
Is your bottleneck CPU or I/O?

CPU-bound:
  → Use real parallelism (multiple cores doing work)
  → Go: goroutines (automatic)
  → Java: thread pool / parallel streams
  → Python: multiprocessing or C extension
  → Node.js: worker_threads

I/O-bound:
  → Use async / non-blocking I/O
  → Go: goroutines (automatic — runtime handles it)
  → Java: virtual threads (Project Loom) or async frameworks
  → Python: asyncio or threads (GIL released during I/O)
  → Node.js: default model — everything is already async
```

Go's advantage is that goroutines handle both cases transparently. You write the same code whether you're doing CPU work or waiting for I/O — the runtime figures out the scheduling.

## Optimising I/O-Bound Code — Batching and Durability

For HTTP handlers and real-time data pipelines, the bottleneck is almost always **I/O, not CPU**. A typical request spends ~90% of its time waiting for the database, not parsing JSON or transforming data.

```
Handle request: 50ms total
  ├── Parse body:         0.1ms   ← CPU, almost never the bottleneck
  ├── Validate input:     0.01ms  ← CPU, irrelevant
  ├── Query database:    45ms     ← I/O, network round-trip + query execution
  ├── Transform result:   0.5ms   ← CPU, rarely matters
  └── Serialize response: 0.2ms   ← CPU, rarely matters
```

Optimise I/O first: add database indexes, eliminate N+1 queries (one JOIN instead of 100 queries), cache results that don't change often, and make independent external calls in parallel instead of sequentially.

### Batching writes

When receiving high-frequency data (e.g. financial updates over a WebSocket), each database call costs a network round-trip (~1-5 ms). Buffering records in memory and flushing in batches reduces round-trips:

```
25 individual inserts:  25 × ~2ms = ~50ms
1 batch insert:          1 × ~3ms = ~3ms
```

The pattern: buffer records, flush when the buffer hits a size threshold **or** a time limit — whichever comes first. The time limit prevents records from sitting in memory forever during slow periods.

```
on_message(record):
    buffer.append(record)
    if buffer.length >= 25:
        flush(buffer)

every 500ms:
    if buffer is not empty:
        flush(buffer)
```

**Go vs TypeScript difference:** In Go, if multiple goroutines push to the buffer concurrently, you need a mutex to protect it — or use a channel to funnel records to a single flushing goroutine. In TypeScript/Node.js, the event loop is single-threaded, so no lock is needed — `buffer.push()` and `buffer = []` can never interleave.

### Crash resilience — WAL (Write-Ahead Log)

Buffered records in memory are lost if the process crashes. A WAL is a simple file on local disk where you write each record **before** adding it to the in-memory buffer. If the process dies, the WAL file survives and you replay it on restart.

```
write-ahead logging:
    on_message(record):
        append record to WAL file on disk    ← survives crash
        buffer.append(record)
        if buffer.length >= 25:
            batch_insert(buffer)
            clear WAL file                   ← only after successful DB write

    on_startup:
        if WAL file has records:
            batch_insert(records from WAL)   ← recover what was lost
            clear WAL file
```

The key ordering: **disk first, then memory, then DB, then clear WAL**. At any crash point, the WAL has everything that didn't make it to the database.

This is exactly what databases do internally — PostgreSQL writes to its own WAL before updating tables, and replays it on crash recovery.

**Tradeoff:** WAL writes are synchronous (must block until data is on disk), which slows down the hot path. Fine for hundreds of messages per second, but becomes a bottleneck at thousands.

### How expensive is a WAL write?

A file write has two parts: the `write()` syscall and the `fsync()` to force data to physical disk.

```
write() syscall:           ~0.001ms  ← just copies to OS page cache in RAM
fsync() to force to disk:  ~0.1-2ms (SSD), ~5-15ms (HDD)
```

Without `fsync()`, the data sits in an OS memory buffer. An application crash is fine (the OS buffer survives), but an OS crash or power loss loses the buffer. For a WAL to be truly durable, you need `fsync()`.

In context of a typical request:

```
Handle request: 50ms total
  ├── Parse body:         0.1ms
  ├── Validate input:     0.01ms
  ├── WAL append + fsync: 0.5-2ms   ← noticeable but small vs database
  ├── Query database:    45ms       ← still dominates by 20x
  ├── Transform result:   0.5ms
  └── Serialize response: 0.2ms
```

File reads have similar variability depending on the page cache (see lesson 04 for details):

```
File read (page cache hit):        ~0.01-0.1ms   ← data already in RAM
File read (page cache miss, SSD):  ~0.05-0.5ms   ← actual disk access
File read (page cache miss, HDD):  ~5-10ms       ← physical seek
```

A WAL adds a couple of milliseconds at most on SSD — a reasonable tradeoff when the alternative is losing data on crash.

### WAL doesn't work everywhere

A local WAL file needs **persistent local disk**, which not all environments have:

| Environment         | Local disk              | WAL viable?                                  |
| ------------------- | ----------------------- | -------------------------------------------- |
| Bare metal / VM     | Persistent              | Yes                                          |
| ECS on EC2          | Survives task restarts  | Mostly — lost if EC2 instance dies           |
| ECS on Fargate      | Ephemeral               | No — lost when task stops                    |
| AWS Lambda          | `/tmp` is ephemeral     | No — lost when container is recycled         |

When local disk isn't reliable, use a **managed durable buffer** — SQS, Kinesis, or Kafka. These are essentially managed WALs with consumer APIs:

```
Producer (ECS/Lambda):
    on_message(record):
        send to SQS/Kinesis                  ← durable, replicated, survives crashes

Consumer (separate Lambda/ECS task):
    triggered by batch of 10-25 messages
    batch_insert(records)
```

SQS/Kinesis handle the batching for you — you configure a batch size and a max wait time, and the consumer is triggered when either threshold is reached. Same "size or time, whichever first" pattern, but AWS manages it.

**Kafka vs database writes:** Kafka is fast because it's an append-only log — it just writes to the end of a file sequentially. No query parsing, no index updates, no B-tree lookups. A database INSERT does all of those. But Kafka is still a network round-trip to a separate service, so for simplicity, SQS/Kinesis is often good enough unless you need Kafka's ordering and throughput guarantees.

## Practical Summary

1. **Threads share memory, processes don't** — shared memory is faster but requires synchronisation. Processes are safer but communication is expensive.
2. **Context switches cost 1-50 us** — thread switches are cheaper than process switches. Goroutine switches are cheapest (~200 ns) because they stay in userspace.
3. **Cache coherence is the real cost of shared state** — it's not the lock instruction, it's the cache line bouncing between cores at ~40-100 ns per bounce.
4. **False sharing is invisible** — two variables on the same cache line cause the same coherence traffic as two threads writing to the same variable. Pad your hot concurrent data.
5. **Memory ordering is not guaranteed** — compilers and CPUs reorder operations. Use atomics, mutexes, or channels to establish happens-before relationships.
6. **Every concurrency model is a tradeoff** — OS threads (simple, expensive), green threads (cheap, needs a runtime), event loops (no locks needed, no CPU parallelism), processes (fully isolated, expensive communication).
7. **Cache lines are 64 bytes** — the smallest unit the cache loads or evicts. L1 holds ~512 lines, L2 ~4,096, L3 ~262,144. If your hot data fits in L1, everything runs at ~1 ns.
8. **L1 is split into L1i (instructions) and L1d (data)** — 32 KB each, separate hardware. L2 and L3 are unified.
9. **Loop order matters for cache performance** — walk memory sequentially (row-major) not strided. Swapping loop order on a large matrix can be a 5-10x difference.
10. **I/O is almost always the bottleneck** — optimise database queries (indexes, eliminate N+1, cache) before optimising CPU code. Profile first.
11. **Batch writes to reduce round-trips** — buffer records in memory, flush on size or time threshold. Each network round-trip costs ~1-5 ms regardless of payload size.
12. **Use a WAL for crash resilience** — write to disk before memory, clear after successful DB write. When local disk isn't available (Lambda, Fargate), use a managed durable buffer (SQS, Kinesis, Kafka) instead.

## What's Next

We've covered how the CPU executes code, how memory hierarchy affects performance, and how multiple cores coordinate. The next lesson covers I/O — how your program talks to the outside world (disk, network), the syscall boundary, blocking vs non-blocking I/O, and why every language's async model exists.
