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

Instead of blocking, you can tell the OS: "if there's no data, return immediately and tell me there's nothing."

```
Thread:     read(socket, buf, 1024)  ← socket set to non-blocking
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

1. Set sockets to non-blocking mode
2. Use `epoll_wait()` to block cheaply until the kernel signals which sockets have data
3. Call `read()` only on those sockets — since epoll confirmed they're ready, `read()` won't block

Non-blocking mode then acts as a **safety net**: if there's ever a race condition where you call `read()` on a socket that turns out not to be ready, you get `EWOULDBLOCK` instead of your thread hanging unexpectedly.

You need a way to say: "tell me when ANY of my 10,000 sockets have data."

## I/O Multiplexing — Watching Many Sockets at Once

The core idea: instead of one thread per connection, one thread watches all connections and only acts when something is ready.

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
- 1 `epoll_wait()` monitoring 10,000 sockets
- ~0 idle threads

A traditional Java server handling 10,000 connections uses:

- 10,000 threads (most sleeping)
- ~10 GB of stack memory
- Context switch overhead when any of them wake up

The Node.js approach uses orders of magnitude less memory and zero context switches for I/O. The tradeoff: if your JavaScript callback takes 50 ms of CPU work, all 10,000 connections are stalled for 50 ms.

### What about file I/O and DNS?

Disk reads and DNS lookups don't support epoll well on Linux. libuv uses a **thread pool** (default 4 threads) for these:

```
Event loop thread:
  fs.readFile('big.csv', callback)
    │
    ▼
  Thread pool (OS threads):
    Thread 1: read('big.csv') ← blocking read, but on a worker thread
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

Virtual threads are managed by the JVM, not the OS. When a virtual thread blocks on I/O, the JVM unmounts it from the OS thread (called a "carrier thread") and mounts a different virtual thread. Same idea as Go's goroutines.

|              | OS Thread        | Virtual Thread (Java 21) | Goroutine (Go)         |
| ------------ | ---------------- | ------------------------ | ---------------------- |
| Stack        | ~1 MB fixed      | ~few KB, grows           | ~4 KB, grows           |
| Scheduling   | OS kernel        | JVM runtime              | Go runtime             |
| I/O blocking | Blocks OS thread | Parks, frees OS thread   | Parks, frees OS thread |
| Count        | Thousands        | Millions                 | Millions               |

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

Instead of read() (kernel copies data to your buffer), you can map the file into your address space:

```go
data, _ := syscall.Mmap(fd, 0, fileSize, syscall.PROT_READ, syscall.MAP_PRIVATE)
// data is now a []byte that IS the file
// accessing data[i] triggers a page fault if not in page cache
// kernel loads the page from disk, maps it into your address space
// no copy — you're reading directly from the page cache
```

This avoids the kernel → userspace copy. It's useful for large files you access randomly (databases use this extensively). The downside: page faults are handled by the kernel and can block unpredictably.

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
