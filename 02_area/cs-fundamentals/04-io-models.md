# I/O Models

In the first three lessons we stayed inside the CPU — instructions, caches, cores. But real programs spend most of their time waiting for things outside the CPU: disk reads, network responses, database queries. Understanding I/O at the OS level explains why Node.js is single-threaded, why Go doesn't need async/await, and why "async" exists in every modern language.

## The Syscall Boundary

Your program runs in **user space**. It cannot touch hardware directly — no reading from disk, no sending network packets, no allocating memory pages. To do any of these, it must ask the kernel through a **system call** (syscall).

```
┌─────────────────────────────────┐
│          User Space             │
│                                 │
│  Your Go / Node / Python code   │
│         │                       │
│         │ syscall               │
│─────────┼───────────────────────│
│         ▼                       │
│          Kernel Space           │
│                                 │
│  File systems, network stack,   │
│  device drivers, scheduler      │
│         │                       │
│         ▼                       │
│       Hardware                  │
│  Disk, NIC, GPU, etc.           │
└─────────────────────────────────┘
```

A syscall is expensive compared to a normal function call:

| Operation            | Cost        |
| -------------------- | ----------- |
| Normal function call | ~1-5 ns     |
| Syscall              | ~100-300 ns |

The cost comes from switching CPU modes (user → kernel → user), saving/restoring state, and security checks. This is why languages batch I/O operations and why buffered I/O exists — you want fewer syscalls, each doing more work.

### Common syscalls

| Syscall             | What it does                                         |
| ------------------- | ---------------------------------------------------- |
| `read(fd, buf, n)`  | Read n bytes from file descriptor into buffer        |
| `write(fd, buf, n)` | Write n bytes from buffer to file descriptor         |
| `open(path, flags)` | Open a file, returns a file descriptor (integer)     |
| `close(fd)`         | Close a file descriptor                              |
| `socket(...)`       | Create a network socket                              |
| `connect(fd, addr)` | Connect socket to a remote address                   |
| `accept(fd)`        | Accept an incoming connection                        |
| `epoll_wait(...)`   | Wait for events on multiple file descriptors (Linux) |
| `mmap(...)`         | Map a file into memory                               |

Everything is a **file descriptor** (fd) — an integer that represents an open resource. Files, sockets, pipes, stdin/stdout — all file descriptors. `read()` and `write()` work on all of them. This is Unix's "everything is a file" design.

### Buffered vs unbuffered I/O

Every `write()` syscall costs ~200 ns overhead regardless of size. Writing 1 byte at a time is catastrophically wasteful:

```
Unbuffered — 1 million writes of 1 byte:
  1,000,000 syscalls × ~200 ns = ~200 ms just in syscall overhead

Buffered — collect into 4 KB chunks, then write:
  ~244 syscalls × ~200 ns = ~0.05 ms in syscall overhead
```

This is why every language has buffered I/O by default:

| Language | Buffered writer                                                         |
| -------- | ----------------------------------------------------------------------- |
| Go       | `bufio.Writer` (also `fmt.Fprint` to an `os.File` is unbuffered — trap) |
| Java     | `BufferedOutputStream`, `BufferedWriter`                                |
| Node.js  | Streams are buffered by default                                         |
| Python   | `open()` returns buffered by default                                    |
| C        | `stdio` (`printf`, `fwrite`) is buffered, `write()` is not              |

```go
// Slow — one syscall per line
for _, line := range lines {
    f.WriteString(line + "\n")  // each call is a write() syscall
}

// Fast — batches writes, one syscall per ~4 KB
w := bufio.NewWriter(f)
for _, line := range lines {
    w.WriteString(line + "\n")  // writes to in-memory buffer
}
w.Flush()  // one final write() syscall for remaining data
```

## Blocking I/O — The Simple Model

The simplest I/O model: call `read()`, and your thread sleeps until data is available.

```
Thread:     read(socket, buf, 1024)
               │
               ▼
            ┌──────────┐
            │ BLOCKED  │  Thread is sleeping. OS moved it off the CPU.
            │ waiting  │  Another thread can use this core.
            │ for data │
            └──────────┘
               │
               │  ← data arrives from network
               ▼
            Thread wakes up, read() returns with data
```

This is what happens by default in Go, Java, C, and Python when you call `read()` or `recv()`.

### One thread per connection

The natural pattern with blocking I/O: spawn one thread per client connection.

```
Client A connects → Thread 1 handles it (blocks on read, processes, writes back)
Client B connects → Thread 2 handles it
Client C connects → Thread 3 handles it
...
Client 10000 connects → Thread 10000 handles it
```

This works, but scales poorly:

| Threads | Problem                                                                        |
| ------- | ------------------------------------------------------------------------------ |
| 100     | Fine — modern OS handles this easily                                           |
| 1,000   | Getting heavy — ~1 GB of stack memory (1 MB per thread)                        |
| 10,000  | Context switch overhead dominates — OS spends more time switching than working |
| 100,000 | Not feasible with OS threads                                                   |

Most of these threads are **idle** — waiting for network data. You're paying 1 MB of stack and context switch overhead for a thread that spends 99% of its time sleeping.

This is the problem that drove the invention of non-blocking I/O and event loops.

## Non-Blocking I/O — Don't Sleep, Check Later

With blocking I/O, a thread that calls `read()` with no data available just sleeps — it can't do anything else until data arrives. Non-blocking I/O changes that: `read()` returns immediately with `EWOULDBLOCK` instead of sleeping, so the thread stays free.

This isn't a mode the thread switches into — it's a **flag on the socket** itself (`O_NONBLOCK`, set via `fcntl(fd, F_SETFL, O_NONBLOCK)`). Once the flag is set on a socket, any thread calling `read()` on it will get the non-blocking behaviour.

```
Thread:     read(socket, buf, 1024)  ← socket has O_NONBLOCK flag set
               │
               ▼
            Returns immediately: EWOULDBLOCK (no data yet)
            Thread continues doing other work
            ...
            read(socket, buf, 1024)  ← try again
               │
               ▼
            Returns immediately: EWOULDBLOCK (still nothing)
            ...
            read(socket, buf, 1024)  ← try again
               │
               ▼
            Returns 47 bytes of data!
```

The thread gets control back immediately and can do anything. But the naive approach — calling `read()` in a tight loop until data arrives — is **busy polling**. The thread burns 100% CPU hammering the kernel with syscalls while doing no real work.

Non-blocking I/O alone doesn't solve the scalability problem. Its real purpose is to be used **in combination with epoll** (see next section). The pattern is:

1. Set the socket's `O_NONBLOCK` flag
2. Use `epoll_wait()` to block cheaply until the kernel signals which sockets have data
3. Call `read()` only on those sockets — since epoll confirmed they're ready, `read()` won't block

Non-blocking mode then acts as a **safety net**: if there's ever a race condition where you call `read()` on a socket that turns out not to be ready, you get `EWOULDBLOCK` instead of your thread hanging unexpectedly.

You need a way to say: "tell me when ANY of my 10,000 sockets have data."

## I/O Multiplexing — Watching Many Sockets at Once

The core idea: instead of one thread per connection, one thread sleeps until the kernel wakes it — and when it wakes, it only acts on connections that are ready.

The difference between `select` and `epoll` comes down to **who does the work of finding what's ready**.

### select — you ask, kernel scans

With `select`, you hand the kernel your full list of sockets on every call. The kernel scans the entire list, marks the ready ones, and hands it back. You then scan through it yourself to find them. 10,000 sockets means 10,000 checks — every single call, even if only one socket has data. It also has a hard limit of 1024 fds.

`poll()` removed the 1024 limit but kept the same scanning behaviour.

### epoll — kernel tells you

With epoll, you register your sockets once. The kernel attaches callbacks to each socket's receive buffer — when a packet arrives and the NIC fires its interrupt, the kernel's network stack processes it and the callback fires, adding that socket to a **ready list**. Your thread sleeps in `epoll_wait()` — genuinely sleeping, consuming zero CPU — and the kernel wakes it up as soon as the ready list is non-empty.

The result: `epoll_wait()` returns only the sockets that actually have data. The kernel never scans. Your thread never polls. It wakes up precisely when there is work to do.

```
select:  you ──ask──► kernel scans everything ──► you scan results
epoll:   kernel watches ──► interrupt fires ──► kernel wakes you ──► you act
```

|                   | `select`                    | `epoll`                           |
| ----------------- | --------------------------- | --------------------------------- |
| fd limit          | 1024                        | millions                          |
| kernel work       | O(n) scan on every call     | O(1) — callback-driven ready list |
| what you get back | full set, ready ones marked | only the ready fds                |

macOS/BSD uses `kqueue` (same idea as epoll). Windows uses `IOCP`.

## The Event Loop — How Node.js Works

Node.js wraps epoll/kqueue/IOCP into a library called **libuv**, which runs an event loop:

```
while (true) {
    // 1. Run any callbacks that are ready (your JavaScript)
    processCallbackQueue()

    // 2. Ask the OS: which I/O operations completed?
    events = epoll_wait(...)    // blocks if nothing is ready

    // 3. For each completed I/O, queue its callback
    for event in events {
        callbackQueue.push(event.callback)
    }

    // 4. Check timers (setTimeout, setInterval)
    processTimers()
}
```

Your JavaScript runs in step 1. It never directly waits for I/O — it registers callbacks. The event loop blocks in `epoll_wait()` only when there's nothing to do. When I/O completes, the callback runs.

```javascript
// What you write:
const data = await fetch("https://api.example.com/data");
console.log(data);

// What actually happens:
// 1. fetch() creates a socket, sends HTTP request
// 2. Registers the socket with epoll
// 3. Your function is suspended (it's a Promise)
// 4. Event loop continues — serves other requests
// 5. epoll_wait() eventually returns: "socket has data"
// 6. libuv queues the callback
// 7. Your function resumes with the data
```

### Why this works for I/O-heavy servers

A Node.js server handling 10,000 concurrent connections uses:

- 1 thread for JavaScript
- 1 thread sleeping in `epoll_wait()`, registered against 10,000 sockets
- ~0 idle threads

A traditional Java server handling 10,000 connections uses:

- 10,000 threads (most sleeping)
- ~10 GB of stack memory
- Context switch overhead when any of them wake up

The Node.js approach uses orders of magnitude less memory and zero context switches for I/O. The tradeoff: if your JavaScript callback takes 50 ms of CPU work, all 10,000 connections are stalled for 50 ms.

### What about file I/O and DNS?

For network I/O, the NIC fires a hardware interrupt when data arrives — epoll hooks into that. Disk reads have no equivalent: there's no interrupt the kernel can cleanly expose through epoll, it just blocks. DNS is the same problem — the standard resolver (`getaddrinfo`) is a synchronous blocking call with no async interface.

libuv works around this with a **thread pool** (default 4 threads). These worker threads do the blocking syscall while the event loop thread stays free. When the worker finishes, it queues the callback back onto the event loop.

```
Event loop thread:
  fs.readFile('big.csv', callback)
    │
    ▼
  Thread pool (OS threads):
    Thread 1: read('big.csv') ← blocking syscall on a worker thread
    Thread 2: (available)
    Thread 3: (available)
    Thread 4: (available)
    │
    │ ← read completes
    ▼
  Callback queued on event loop
  Event loop runs callback in JavaScript
```

This is why Node.js isn't truly single-threaded — it has worker threads for operations that can't be made async at the OS level. But your JavaScript code only runs on one thread.

The pool also handles CPU-bound crypto operations (`crypto.pbkdf2`, `crypto.scrypt`) since running those on the event loop thread would stall all other connections.

If you exhaust the 4 threads — many concurrent file reads, DNS lookups, or crypto calls queuing up — operations wait for a free worker. The event loop stays responsive but I/O latency climbs. You can increase the pool size via the `UV_THREADPOOL_SIZE` environment variable (max 1024, but beyond your CPU core count you're just adding context switch overhead):

```sh
UV_THREADPOOL_SIZE=16 node app.js
```

`io_uring` (Linux 5.1, 2019) finally gave the kernel a proper async disk I/O interface — submit an operation, get notified when done, no thread blocked. libuv has been gradually adopting it, which could eventually make the thread pool unnecessary for disk I/O.

## How Go Handles I/O — The Best of Both Worlds

Go gives you the simplicity of blocking I/O with the efficiency of non-blocking I/O. You write code that looks blocking, but the runtime uses epoll underneath.

```go
// This LOOKS like it blocks the thread:
data, err := conn.Read(buf)

// But here's what actually happens:
// 1. Go runtime calls read() on the socket (non-blocking mode)
// 2. If data is available: return immediately
// 3. If EWOULDBLOCK: runtime parks this goroutine
//    — goroutine is moved off the OS thread
//    — socket is registered with the runtime's epoll instance (netpoller)
//    — OS thread picks up another runnable goroutine
// 4. Later: epoll says "socket has data"
// 5. Runtime unparks the goroutine, schedules it on an available OS thread
// 6. read() returns with data
```

```
Goroutine A: conn.Read() → no data → parked ─────────────────→ data arrives → unparked
Goroutine B: conn.Read() → no data → parked ──→ data arrives → unparked
Goroutine C: conn.Read() → data ready → returns immediately

OS Thread 1: runs C, then picks up B when it's unparked
OS Thread 2: runs other goroutines while A and B are parked
```

The Go runtime has a component called the **netpoller** — it runs epoll in a dedicated thread, watching all sockets that goroutines are waiting on. When data arrives, it marks the corresponding goroutine as runnable.

### Why this is elegant

| Model                   | What you write  | What happens                   | Downsides                                       |
| ----------------------- | --------------- | ------------------------------ | ----------------------------------------------- |
| Blocking (Java classic) | `socket.read()` | Thread sleeps                  | 1 thread per connection, doesn't scale          |
| Event loop (Node.js)    | `await fetch()` | Callback queued                | Callback/promise complexity, no CPU parallelism |
| Goroutines (Go)         | `conn.Read()`   | Goroutine parked, thread freed | Need a runtime/scheduler (Go has one built in)  |

Go's approach means you can write straightforward sequential code — no callbacks, no promises, no async/await — and the runtime handles the multiplexing. One goroutine per connection, millions of goroutines, a handful of OS threads.

## How Java Caught Up — Virtual Threads (Project Loom)

Java's traditional model was one OS thread per connection. Java 21 introduced **virtual threads** — the same idea as goroutines:

```java
// Old Java: one OS thread per request — expensive
executor.submit(() -> {
    var data = socket.read();  // OS thread blocks
    process(data);
});

// Java 21+: one virtual thread per request — cheap
Thread.startVirtualThread(() -> {
    var data = socket.read();  // virtual thread parks, OS thread freed
    process(data);
});
```

Virtual threads are managed by the JVM, not the OS. When a virtual thread blocks on I/O, the JVM unmounts it from the OS thread (called a "carrier thread") and mounts a different virtual thread. Same idea as Go's goroutines — you write sequential blocking-style code, the runtime handles the multiplexing via epoll underneath.

|              | OS Thread        | Virtual Thread (Java 21) | Goroutine (Go)         |
| ------------ | ---------------- | ------------------------ | ---------------------- |
| Stack        | ~1 MB fixed      | ~few KB, grows           | ~4 KB, grows           |
| Scheduling   | OS kernel        | JVM runtime              | Go runtime             |
| I/O blocking | Blocks OS thread | Parks, frees OS thread   | Parks, frees OS thread |
| Count        | Thousands        | Millions                 | Millions               |

### Java vs Go — practical differences

The concurrency model is now essentially the same. The remaining gaps are operational:

- **Startup time** — Go starts in milliseconds. The JVM takes seconds to initialise. Matters a lot for Lambda and containers that scale to zero.
- **Memory footprint** — A Go service might use 20 MB. An equivalent Java service carries the JVM heap, metaspace, and GC overhead — often 200 MB+.
- **Deployment** — Go compiles to a single static binary. Java requires the JVM runtime.

**JIT vs native compilation** isn't as clear-cut as it sounds. HotSpot's JIT has decades of optimisation and uses runtime profiling data — warmed-up Java throughput can match or exceed Go for raw CPU work. The JVM's weakness is the warmup period and the memory cost to get there, not peak throughput.

**GraalVM Native Image** compiles Java to native (like Go), eliminating startup time and memory overhead. The tradeoff: you lose JIT's runtime profiling advantages, and some older libraries that rely on reflection or dynamic class loading break.

Virtual threads were also **retrofitted** onto an existing ecosystem built around OS threads. Most standard library code works transparently, but older libraries using `synchronized` blocks or `ThreadLocal` can behave unexpectedly — they were written assuming one OS thread per task, which virtual threads violate.

## Go vs Java 21 vs Node.js

With Java 21 and Go both using M:N scheduling over epoll, the more interesting comparison is how all three handle the combination of I/O concurrency and CPU work.

**Node.js is the odd one out.** Go and Java distribute work across all CPU cores transparently — goroutines and virtual threads run on a pool of OS threads equal to the core count. Node.js runs JavaScript on a **single thread**. No matter how efficiently epoll handles I/O, any CPU work blocks everything else. A 50 ms bcrypt call stalls all 10,000 connections.

Node.js has `worker_threads` for parallelism, but it's architecturally awkward:

- Each worker spins up a full V8 engine — separate heap, separate GC, several MB overhead just to exist
- You can't share JavaScript objects between threads — you must `postMessage(data)`, which serialises the object (deep copy), sends the bytes, and the receiver deserialises it back
- For large data this copying is expensive, often costing more than the parallelism saves
- You have to explicitly identify CPU-bound work and plumb the offloading yourself — the runtime never does it automatically

```
Go / Java:    goroutine/virtual thread ──► any available OS thread (shared heap, pointer passing)
Node.js:      main thread ──postMessage──► worker (full V8, serialised copy)
```

In practice, Node.js worker threads make sense for genuinely heavy isolated work — image processing, video encoding, complex cryptography. For typical web server tasks like JSON parsing, the serialisation round-trip costs more than the parallelism saves.

|                   | Node.js                       | Go                       | Java 21                   |
| ----------------- | ----------------------------- | ------------------------ | ------------------------- |
| Concurrency model | Event loop (single JS thread) | Goroutines (M:N)         | Virtual threads (M:N)     |
| CPU parallelism   | No — single JS thread         | Yes — GOMAXPROCS threads | Yes — carrier threads     |
| Cross-thread data | Serialised copy (postMessage) | Shared memory (pointer)  | Shared memory (pointer)   |
| Memory            | Moderate (V8)                 | Lean                     | Heavy (JVM)               |
| Startup           | Fast                          | Very fast                | Slow (GraalVM fixes this) |
| Code style        | async/await required          | Sequential               | Sequential                |

**Where Node.js still wins:** npm ecosystem size, sharing code between frontend and backend, and pure I/O-heavy services where CPU is never the bottleneck.

**Where Go wins over Java:** startup time, memory footprint, single static binary, and a runtime designed around goroutines from day one rather than retrofitted.

**For long-running I/O-bound services**, all three are now competitive on throughput. The choice comes down to ecosystem, operational characteristics, and whether CPU parallelism matters for your workload.

## Python's async/await — asyncio

Python's `asyncio` is an event loop, similar to Node.js but bolted onto a language that wasn't designed for it:

```python
import asyncio

async def handle_client(reader, writer):
    data = await reader.read(1024)     # coroutine suspends, event loop continues
    writer.write(data)
    await writer.drain()

async def main():
    server = await asyncio.start_server(handle_client, '0.0.0.0', 8080)
    await server.serve_forever()

asyncio.run(main())
```

Under the hood, `asyncio` uses epoll (Linux) or kqueue (macOS) — same as Node.js. The `await` keyword suspends the coroutine and returns control to the event loop.

### Python's I/O problem

Python has **three incompatible ways** to do I/O:

```python
# 1. Synchronous (blocks the thread, blocks the GIL)
data = socket.recv(1024)

# 2. Threading (blocks the thread, releases the GIL during I/O)
thread = threading.Thread(target=lambda: socket.recv(1024))

# 3. Async (doesn't block anything, but requires async all the way up)
data = await reader.read(1024)
```

The problem with asyncio: it's **infectious**. An async function can only be called from another async function. If you have a synchronous library (most Python libraries), you can't use it inside async code without wrapping it in a thread:

```python
# Can't do this — requests is synchronous
async def fetch():
    response = requests.get('https://...')  # blocks the event loop!

# Must do this instead
async def fetch():
    response = await asyncio.to_thread(requests.get, 'https://...')
```

This "function colouring" problem (sync functions and async functions are different colours that don't mix) is why async Python feels bolted-on. Go avoids this entirely — every function is the same colour because the runtime handles the async/sync boundary transparently.

## Disk I/O — Different from Network I/O

Network I/O and disk I/O look similar from your code (`read()`/`write()`) but behave very differently at the OS level.

### Network I/O

- Naturally asynchronous — data arrives whenever the remote end sends it
- epoll/kqueue work perfectly — "tell me when data arrives"
- Latency is unpredictable (1 ms to 1000 ms depending on network)

### Disk I/O

- Appears synchronous — you ask for a block, the disk fetches it
- Linux doesn't support true async disk I/O well through epoll
  - `io_uring` (Linux 5.1+, 2019) finally provides efficient async disk I/O
  - Before that: thread pool workaround (what libuv/Go/Java do)
- SSD latency: ~50-100 us. HDD latency: ~5-10 ms (seek time dominates)

### The page cache

The OS caches disk data in RAM — the **page cache**. When you `read()` a file:

```
First read:
  read(fd, buf, 4096)
    → kernel checks page cache: miss
    → kernel reads from disk: ~100 us (SSD) or ~10 ms (HDD)
    → data cached in page cache
    → data copied to your buffer

Second read (same data):
  read(fd, buf, 4096)
    → kernel checks page cache: hit
    → data copied to your buffer
    → cost: ~1-2 us (memcpy, no disk I/O)
```

This is why:

- The first run of a program that reads files is slow, subsequent runs are fast
- A database "warming up" is literally filling the page cache
- Free RAM on a Linux server isn't wasted — it's page cache

You can check page cache usage:

```sh
free -h
#               total   used    free    shared  buff/cache  available
# Mem:           16G    4.2G    2.1G      256M       9.7G      11.4G
#                                                    ^^^^
#                                         this is mostly page cache
```

### mmap — Mapping files directly into memory

`read()` copies data from the page cache into your userspace buffer — two copies: disk → page cache → your buffer. `mmap()` skips the second copy entirely by mapping the file directly into your process's virtual address space.

```go
data, _ := syscall.Mmap(int(f.Fd()), 0, int(fi.Size()), syscall.PROT_READ, syscall.MAP_SHARED)
// data is a []byte — the entire file
// no data has moved yet
```

When you call `mmap()`, the kernel carves out a region of your virtual address space and associates it with the file. No data is loaded. Page table entries for that region are marked **not present**.

When you first access `data[1000]`:

```
your code:  data[1000]  ← memory read
MMU:        looks up page table → "not present" → page fault
kernel:     checks page cache
            hit  → maps page cache page into your page table → resume
            miss → reads from disk → caches it → maps it → resume
your code:  gets the byte — reading directly from page cache, no copy
```

Subsequent accesses to the same page are just memory reads — no syscall, no copy, no fault.

```
read():   disk → page cache → copy to your buffer   (kernel copies to userspace)
mmap():   disk → page cache → page table maps to it  (you read directly from cache)
```

Sequential access triggers OS **readahead** — the kernel notices the pattern and pre-loads pages before you fault on them.

### mmap vs read() — when to use each

mmap is not always better:

| Situation | Prefer |
|---|---|
| Large read-only file, sequential scan | mmap |
| Need zero-copy slices as map keys | mmap |
| Small files | read() — page fault overhead not worth it |
| Streaming / network data | read() — mmap only works on files |
| Write-heavy with ordering requirements | read() — `msync()` complexity |
| Random access on very large files with low RAM | read() — explicit control over what's loaded |
| Need predictable error handling | read() — disk errors become SIGBUS with mmap, killing the process |

TLB pressure is also a real concern with mmap — each mapped region consumes TLB entries (the CPU's cache for page table lookups). Map enough large files and TLB misses start hurting performance. This is why some databases (PostgreSQL, newer SQLite) have moved away from heavy mmap use.

## Practical Summary

1. **Syscalls are expensive (~200 ns)** — batch I/O with buffered writers. One write of 4 KB beats 4096 writes of 1 byte.
2. **Blocking I/O is simple but doesn't scale** — one thread per connection works for hundreds of connections, not thousands.
3. **epoll/kqueue let one thread handle thousands of sockets** — the kernel tracks readiness, you process only what's ready.
4. **Event loops (Node.js, asyncio) build on epoll** — your code registers callbacks, the loop calls them when I/O completes. Great for I/O-heavy work, bad for CPU-heavy work.
5. **Go and Java 21 give you the best tradeoff** — write blocking-style code, the runtime handles the non-blocking I/O underneath. No callback complexity, no function colouring.
6. **Disk and network I/O are fundamentally different** — network is naturally async (epoll), disk is traditionally sync (thread pool or io_uring). The page cache makes repeated disk reads fast.
7. **The page cache is your friend** — free RAM on a server is used for caching file data. Don't assume disk I/O is always slow.

## What's Next

We've covered how the CPU runs code, how memory hierarchy affects performance, how cores coordinate, and how programs talk to the outside world. The final lesson covers memory management — stack vs heap allocation, garbage collection strategies across languages, reference counting, Rust's ownership model, and why these choices define a language's character.
