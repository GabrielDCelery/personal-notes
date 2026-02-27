# Lesson 09: Exec Form, Shell Form, and PID 1

How Docker actually launches your process — and why getting it wrong silently breaks graceful shutdown.

---

## The Problem You Don't Notice Until Production

Your app works fine locally. `docker stop` takes 10 seconds instead of 1. You dismiss it as slow startup. Then in production, during a rolling deploy, requests start dropping. In-flight database transactions get cut mid-write. Connections are not drained.

The root cause is almost always the same: your process is not PID 1, or it is PID 1 but ignores `SIGTERM`. Both lead to `SIGKILL` after the timeout — no cleanup, no graceful shutdown, no warning.

Understanding exec form vs shell form is not a Dockerfile trivia question. It is the difference between a process that shuts down cleanly and one that gets killed.

---

## Shell Form vs Exec Form

`CMD`, `ENTRYPOINT`, and `RUN` each accept two syntaxes.

**Shell form** (string):

```dockerfile
CMD node server.js
```

**Exec form** (JSON array):

```dockerfile
CMD ["node", "server.js"]
```

| Property                | Shell form               | Exec form              |
| ----------------------- | ------------------------ | ---------------------- |
| Syntax                  | Plain string             | JSON array             |
| Shell invoked           | Yes — `/bin/sh -c "..."` | No — direct `execve()` |
| PID 1                   | `sh`                     | Your process           |
| Signal forwarding       | Unreliable               | Direct                 |
| Variable expansion      | Yes                      | No                     |
| Requires shell in image | Yes                      | No                     |

---

## How the Shell Actually Works

Shell form is not just a convenience. It runs your string through a full shell pipeline before calling `execve()`. The shell does this in order:

### 1. Tokenization

Splits on `IFS` characters (space, tab, newline by default):

```sh
node server.js --port 3000
# → ["node", "server.js", "--port", "3000"]
```

Quotes suppress splitting:

```sh
node -e "console.log('hello world')"
# → ["node", "-e", "console.log('hello world')"]
#                   ^ space inside quotes, not split
```

### 2. Variable Expansion

```sh
PORT=3000
node server.js --port $PORT
# → node server.js --port 3000
```

### 3. Command Substitution

```sh
echo $(date)
# runs `date`, captures stdout, substitutes inline
```

### 4. Glob Expansion

```sh
cat *.log
# → cat app.log error.log access.log  (whatever files exist)
```

### 5. Redirect Handling

```sh
node server.js > /var/log/app.log 2>&1
```

`>` and `2>&1` are **not** passed as arguments to `node`. The shell intercepts them, wires up file descriptors, then calls `execve()` without those tokens in `argv[]`. The process never sees them.

### 6. execve()

After all of the above, the shell calls:

```c
execve("/usr/bin/node", ["node", "server.js", "--port", "3000"], envp)
```

---

## Why Exec Form Is an Array

`execve()` is the Linux syscall that replaces the current process with a new one. Its signature:

```c
execve(const char *pathname, char *const argv[], char *const envp[])
```

There is no "string" at the OS level. `argv[]` is an array — each element is one argument, passed verbatim to the new process. A shell's entire job during command parsing is to build this array from a human-readable string.

Exec form skips the shell and passes `argv[]` directly to `execve()`. **You** are building the array. Each element becomes one argument.

```dockerfile
# These are NOT equivalent:
CMD ["node", "--max-old-space-size=512 server.js --port 3000"]
#                ^ one blob string passed as a single argument to node

CMD ["node", "--max-old-space-size=512", "server.js", "--port", "3000"]
#            ^ four separate arguments, node receives them correctly
```

The rule: each space-separated token the shell would produce becomes one array element.

---

## What execve() Actually Does at the Kernel Level

When `execve()` is called, the kernel:

1. Opens the binary, checks permissions
2. Reads the ELF header — finds the entry point and required dynamic linker (`ld.so`)
3. **Tears down the current address space** — all memory mappings are freed
4. Maps the new binary's segments into memory (code, data, BSS)
5. Sets up a fresh stack with `argv[]`, `envp[]`, and the auxiliary vector
6. Hands control to `ld.so`, which loads shared libraries
7. `ld.so` jumps to `main()`

The old process's code is gone. The kernel never returns to it. Crucially: **the PID does not change**. Same process slot in the kernel's process table, completely different code running.

Every process is represented by a `task_struct` in the kernel:

```c
struct task_struct {
    pid_t pid;              // unchanged across execve()
    sigset_t pending;       // bitmask of pending signals
    sigset_t blocked;       // bitmask of blocked signals
    struct sigaction *sighand; // table of handlers per signal
    // memory maps, file descriptors, etc.
};
```

`execve()` replaces the memory maps and resets signal handlers to defaults. PID, file descriptors (unless `FD_CLOEXEC`), and environment survive if passed explicitly.

---

## Shell Form and PID 1

When Docker runs shell form:

```dockerfile
CMD node server.js
```

Docker calls:

```c
execve("/bin/sh", ["/bin/sh", "-c", "node server.js"], envp)
```

`sh` becomes PID 1. It then forks and execs `node`:

```
fork()     → new process (PID 2)
execve()   → PID 2 becomes node
```

Process tree:

```
PID 1: /bin/sh         ← receives signals from docker stop
PID 2: node            ← never directly receives signals
```

When `docker stop` sends `SIGTERM` to PID 1, `sh` receives it. Most shells do not forward signals to children. `node` never gets `SIGTERM`.

---

## How docker stop Works at the Kernel Level

### 1. SIGTERM to PID 1

The Docker daemon calls:

```c
kill(1, SIGTERM)
```

This places `SIGTERM` in the `pending` bitmask of PID 1's `task_struct`. The signal is delivered the next time that process returns from kernel space to user space.

### 2. Signal Delivery

Three possible outcomes:

| Situation                 | What happens                             |
| ------------------------- | ---------------------------------------- |
| Signal handler registered | Kernel redirects to the handler function |
| Default disposition       | Kernel terminates the process            |
| PID 1, no handler         | Signal is **silently dropped**           |

The PID 1 exception is intentional — it is the init process. The kernel trusts it to manage its own lifecycle. Unlike every other process, unhandled signals are not acted on with their default disposition for PID 1. If `node` is PID 1 and has no `SIGTERM` handler, `docker stop` does nothing until the timeout.

### 3. The Timeout and SIGKILL

After 10 seconds (configurable with `--time`), Docker sends `SIGKILL`:

```c
kill(1, SIGKILL)
```

`SIGKILL` cannot be caught, blocked, or ignored — not even by PID 1. The kernel terminates the process without entering user space. Your shutdown handlers never run.

```
docker stop
    │
    ├─► SIGTERM → PID 1
    │       ├─► [handled] graceful shutdown → exit_group() → container stops cleanly
    │       └─► [ignored] process keeps running
    │
    ├─► [10s timeout]
    │
    └─► SIGKILL → PID 1 → kernel forcibly terminates, no cleanup
```

---

## The exec Shell Builtin

Entrypoint scripts often need to do setup before launching the main process. The `exec` shell builtin lets you do both cleanly:

```sh
#!/bin/sh
# do setup
export DATABASE_URL="$(fetch_secret db/url)"
migrate_database

exec node server.js   # ← exec
```

`exec` calls `execve()` on the current process — no fork. The shell replaces itself with `node`:

```
Before: PID 1 = sh
After:  PID 1 = node    ← same PID, shell is gone
```

Without `exec`:

```sh
#!/bin/sh
node server.js   # ← sh forks, node is PID 2
```

```
PID 1 = sh (running, waiting for child)
PID 2 = node
```

With `exec`, you get the signal handling of exec form while retaining the ability to run shell setup logic. This is the standard pattern for Docker entrypoints.

---

## Variable Expansion in Exec Form

Exec form bypasses the shell, so environment variables are not expanded:

```dockerfile
ENV PORT=3000
CMD ["node", "server.js", "--port", "$PORT"]
#                                   ^ literal string "$PORT", not 3000
```

If you need expansion in exec form, invoke the shell explicitly:

```dockerfile
CMD ["/bin/sh", "-c", "node server.js --port $PORT"]
```

This gives you expansion while still allowing you to control exactly what shell features are active.

---

## Redirects in Exec Form

```dockerfile
# ❌ Broken — > and /var/log/app.log are passed as literal args to node
CMD ["node", "server.js", ">", "/var/log/app.log"]

# ✓ Correct — shell handles the redirect
CMD ["/bin/sh", "-c", "node server.js > /var/log/app.log 2>&1"]
```

---

## ENTRYPOINT vs CMD

These two interact in ways that trip people up:

|                     | ENTRYPOINT exec form                | ENTRYPOINT shell form |
| ------------------- | ----------------------------------- | --------------------- |
| **CMD exec form**   | CMD appended as args to ENTRYPOINT  | CMD ignored           |
| **CMD shell form**  | CMD ignored                         | CMD ignored           |
| **docker run args** | Appended to ENTRYPOINT, replace CMD | Ignored               |

The standard production pattern:

```dockerfile
ENTRYPOINT ["node"]   # exec form — becomes PID 1, receives signals
CMD ["server.js"]     # default arg — overridable at runtime
```

```sh
docker run myapp                  # runs: node server.js
docker run myapp worker.js        # runs: node worker.js
```

If `ENTRYPOINT` is shell form, `CMD` and runtime arguments are silently ignored regardless of their form. This is a common source of confusion.

---

## Hands-On Exercise 1: Diagnose the Signal Problem

This Dockerfile has a signal handling problem. Identify what it is and fix it.

```dockerfile
FROM node:20-alpine

WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .

CMD node server.js
```

```js
// server.js
const http = require("http");
const server = http.createServer((req, res) => res.end("ok"));

process.on("SIGTERM", () => {
  console.log("Shutting down gracefully...");
  server.close(() => process.exit(0));
});

server.listen(3000);
```

The developer added a `SIGTERM` handler but graceful shutdown still does not work. Why? Fix it.

<details>
<summary>Solution</summary>

**Problem:**

Shell form is used — `CMD node server.js` — which means Docker runs:

```
/bin/sh -c "node server.js"
```

Process tree:

```
PID 1: sh
PID 2: node    ← SIGTERM handler is here, but SIGTERM goes to sh
```

`SIGTERM` is sent to `sh` (PID 1). `sh` does not forward it. `node` never receives it. The handler never fires. After 10 seconds, `SIGKILL` terminates everything.

**Fix:**

```dockerfile
CMD ["node", "server.js"]
```

Now `node` is PID 1 directly and receives `SIGTERM`.

**Verify:**

```sh
docker run -d --name test myapp
docker stop test  # should stop in < 1s, not 10s
docker logs test  # should show "Shutting down gracefully..."
```

</details>

---

## Hands-On Exercise 2: Fix the Entrypoint Script

This entrypoint script runs setup before starting the app. Graceful shutdown does not work. Fix it without removing the setup logic.

```dockerfile
FROM node:20-alpine

WORKDIR /app
COPY . .
RUN npm ci

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
```

```sh
#!/bin/sh
# entrypoint.sh

echo "Running migrations..."
node migrate.js

echo "Starting server..."
node server.js
```

<details>
<summary>Solution</summary>

**Problem:**

The ENTRYPOINT is exec form ✓ — `/entrypoint.sh` is PID 1.

But `entrypoint.sh` runs `node server.js` as a child process:

```
PID 1: /entrypoint.sh (sh)
PID 2: node server.js    ← SIGTERM goes to PID 1, not here
```

`sh` receives `SIGTERM` and may or may not forward it. Even if it does, many shells do not cleanly propagate signals to child processes.

**Fix:**

```sh
#!/bin/sh
# entrypoint.sh

echo "Running migrations..."
node migrate.js

echo "Starting server..."
exec node server.js   # ← exec replaces sh with node
```

After `exec`, process tree becomes:

```
PID 1: node server.js   ← sh is gone, node receives SIGTERM directly
```

The migration still runs. The setup logic is intact. But now `node` is PID 1 and receives signals directly.

</details>

---

## Interview Questions

### Q1: What is the difference between shell form and exec form in a Dockerfile?

Interviewers ask this to filter candidates who have only used Docker superficially from those who understand what it actually does. The answer reveals whether you know what PID 1 is and why it matters.

<details>
<summary>Answer</summary>

Shell form (`CMD node server.js`) runs your command through `/bin/sh -c`, making `sh` PID 1 and your process a child. Exec form (`CMD ["node", "server.js"]`) calls `execve()` directly, making your process PID 1.

The practical difference is signal handling: `docker stop` sends `SIGTERM` to PID 1. With shell form, `sh` receives it and typically does not forward it to children, so your process gets `SIGKILL` after the timeout instead of a chance to shut down gracefully.

Exec form also does not require a shell in the image, which matters for distroless or scratch-based images.

</details>

---

### Q2: Why does exec form use a JSON array instead of a string?

Interviewers ask this to probe whether you understand how the OS actually launches processes. It distinguishes developers who know Linux process fundamentals from those who just memorize Dockerfile syntax.

<details>
<summary>Answer</summary>

At the kernel level, `execve()` takes an `argv[]` array — not a string. There is no string-based process launch at the OS level. A shell's job is to parse a human-readable string and build that array.

Exec form bypasses the shell and passes `argv[]` directly to `execve()`. The JSON array maps 1:1 to `argv[]`. Each element is one argument, passed verbatim to the process with no parsing, expansion, or splitting.

A single string like `"node server.js"` cannot be used because `execve()` has no mechanism to split it — that splitting is exactly what a shell does.

</details>

---

### Q3: Your container takes 10 seconds to stop every time. What is likely wrong and how do you diagnose it?

Interviewers ask this because it is a real production symptom. The question tests whether you can reason from observable behavior back to root cause without guessing.

<details>
<summary>Answer</summary>

10 seconds is Docker's default `SIGKILL` timeout. The container is not exiting on `SIGTERM`, so Docker waits the full timeout before force-killing it.

Diagnosis:

```sh
# Check what is PID 1
docker inspect mycontainer --format '{{ .Config.Cmd }}'
docker inspect mycontainer --format '{{ .Config.Entrypoint }}'

# Check if shell form is being used (sh will appear as PID 1)
docker exec mycontainer ps aux
# If you see: sh -c "node server.js" → shell form is the problem

# Check signal handlers in the running process
docker exec mycontainer cat /proc/1/status | grep Sig
# SigCgt bitmask shows which signals are caught
```

Common causes:

1. ❌ Shell form — `sh` is PID 1, does not forward `SIGTERM`
2. ❌ Exec form, but no `SIGTERM` handler and PID 1 special behavior drops the signal
3. ❌ Entrypoint script without `exec` — script is PID 1, app is child

Fix: use exec form for `CMD`/`ENTRYPOINT`, or use `exec` in entrypoint scripts. Add a `SIGTERM` handler in the application if it needs to drain connections before exiting.

</details>

---

### Q4: What does the exec shell builtin do and why is it used in Docker entrypoint scripts?

Interviewers ask this to test depth of knowledge around process management. It also tests whether you know how to write entrypoint scripts correctly rather than just copying them from StackOverflow.

<details>
<summary>Answer</summary>

The `exec` shell builtin calls `execve()` on the current process without forking. The shell replaces itself with the target process — same PID, completely different code. The shell process ceases to exist.

Without `exec` in an entrypoint script:

```
PID 1: sh (running entrypoint.sh, waiting for child)
PID 2: node server.js
```

`SIGTERM` goes to `sh`. `sh` may not forward it. `node` may not shut down cleanly.

With `exec node server.js` as the last line of the script:

```
PID 1: node server.js   ← sh replaced itself
```

`SIGTERM` goes directly to `node`. The shell is completely gone.

This is the standard pattern: use a shell script for setup (environment variables, secrets fetching, database migrations), then `exec` into the main process. You get both the flexibility of shell scripting and the correct PID 1 behavior of exec form.

</details>

---

## Key Takeaways

1. **Shell form prepends `/bin/sh -c`** — your process runs as a child of `sh`, not as PID 1
2. **Exec form calls `execve()` directly** — your process is PID 1, receives signals directly
3. **execve() takes an array, not a string** — the JSON array maps 1:1 to `argv[]` at the kernel level
4. **Each array element is one argument** — split where the shell would tokenize on spaces
5. **PID 1 is special** — unhandled signals are silently dropped, not acted on with default disposition
6. **docker stop sends SIGTERM, waits, then SIGKILL** — if your process does not handle SIGTERM, it gets force-killed after the timeout
7. **SIGKILL cannot be caught** — no cleanup, no connection draining, no graceful shutdown
8. **Use `exec` in entrypoint scripts** — it replaces the shell with your process, preserving PID 1 and signal delivery
9. **Variable expansion and redirects require a shell** — use `["/bin/sh", "-c", "..."]` in exec form when you need them
10. **Shell form ENTRYPOINT ignores CMD entirely** — always use exec form for ENTRYPOINT in production images
