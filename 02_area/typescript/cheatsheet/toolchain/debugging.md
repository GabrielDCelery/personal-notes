# Node.js Debugging & Profiling

## Quick Reference

| Use case               | Tool / Method                           |
| ---------------------- | --------------------------------------- |
| Debugger (breakpoints) | `--inspect` + Chrome DevTools / VS Code |
| CPU profile            | `--prof` or Chrome DevTools             |
| Memory leak            | `--inspect` + heap snapshot             |
| Flame graph            | `clinic flame` or `0x`                  |
| Log-based              | `console.time` / `performance.now()`    |
| Heap stats             | `process.memoryUsage()`                 |

## Node Inspector

### 1. Start with debugger

```sh
# Start and pause on first line
node --inspect-brk dist/server.js

# Start without pausing
node --inspect dist/server.js

# Custom port
node --inspect=0.0.0.0:9229 dist/server.js
```

Open `chrome://inspect` in Chrome, or attach VS Code.

### 2. VS Code launch.json

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug Server",
      "type": "node",
      "request": "launch",
      "program": "${workspaceFolder}/src/server.ts",
      "runtimeExecutable": "tsx",
      "console": "integratedTerminal",
      "env": {
        "NODE_ENV": "development"
      }
    },
    {
      "name": "Attach to Running",
      "type": "node",
      "request": "attach",
      "port": 9229
    }
  ]
}
```

### 3. debugger statement

```typescript
function processOrder(order: Order) {
  debugger; // execution pauses here when inspector is attached
  // step through from here
}
```

## Console Timing

### 4. Measure execution time

```typescript
// Simple
const start = performance.now();
await doWork();
console.log(`Took ${(performance.now() - start).toFixed(2)}ms`);

// console.time (groups nicely)
console.time("db-query");
const result = await db.query("SELECT ...");
console.timeEnd("db-query"); // "db-query: 42.123ms"

// Multiple steps
console.time("total");
console.time("step-1");
await step1();
console.timeEnd("step-1");
console.time("step-2");
await step2();
console.timeEnd("step-2");
console.timeEnd("total");
```

## Memory

### 5. Check memory usage

```typescript
const mem = process.memoryUsage();
console.log({
  rss: `${(mem.rss / 1024 / 1024).toFixed(1)} MB`, // total allocated
  heapUsed: `${(mem.heapUsed / 1024 / 1024).toFixed(1)} MB`, // JS objects
  heapTotal: `${(mem.heapTotal / 1024 / 1024).toFixed(1)} MB`,
  external: `${(mem.external / 1024 / 1024).toFixed(1)} MB`, // C++ objects (Buffers)
});
```

### 6. Heap snapshot (find memory leaks)

```sh
# Start with inspector
node --inspect dist/server.js

# In Chrome DevTools:
# 1. Open chrome://inspect
# 2. Click "inspect" on your process
# 3. Go to Memory tab
# 4. Take heap snapshot
# 5. Use the app, take another snapshot
# 6. Compare snapshots to find leaked objects
```

### 7. Programmatic heap dump

```typescript
import { writeHeapSnapshot } from "node:v8";

// Trigger manually (e.g. via admin endpoint)
app.post("/admin/heapdump", (req, res) => {
  const filename = writeHeapSnapshot();
  res.json({ filename });
});
```

Load the `.heapsnapshot` file in Chrome DevTools Memory tab.

## CPU Profiling

### 8. V8 CPU profile from tests

```sh
node --prof dist/server.js
# Generate load, then stop the process
# This creates isolate-*.log

node --prof-process isolate-*.log > profile.txt
```

### 9. Chrome DevTools CPU profiling

```sh
node --inspect dist/server.js

# In Chrome DevTools:
# 1. Go to Performance tab (or Profiler)
# 2. Click Record
# 3. Generate load
# 4. Stop recording
# 5. Analyze flame chart
```

### 10. clinic.js (automated diagnostics)

```sh
npm install -g clinic

# Flame graph — identify slow functions
clinic flame -- node dist/server.js
# Run load test, then Ctrl+C → opens flame graph in browser

# Doctor — detect common issues (event loop, GC, I/O)
clinic doctor -- node dist/server.js

# Bubbleprof — async bottlenecks
clinic bubbleprof -- node dist/server.js
```

## Event Loop

### 11. Detect event loop lag

```typescript
let lastCheck = performance.now();

setInterval(() => {
  const now = performance.now();
  const lag = now - lastCheck - 1000; // expected 1000ms
  if (lag > 100) {
    console.warn(`Event loop lag: ${lag.toFixed(0)}ms`);
  }
  lastCheck = now;
}, 1000);
```

### 12. Monitor with diagnostics_channel (Node 19+)

```typescript
import diagnostics_channel from "node:diagnostics_channel";

const channel = diagnostics_channel.channel("http.server.request.start");
channel.subscribe((message) => {
  console.log("HTTP request started:", message);
});
```

## Useful Flags

### 13. Node.js diagnostic flags

```sh
# Trace warnings with stack traces
node --trace-warnings dist/server.js

# Trace deprecations
node --trace-deprecation dist/server.js

# Expose GC for manual triggering
node --expose-gc dist/server.js

# Max old space size (increase heap limit)
node --max-old-space-size=4096 dist/server.js

# Report on crash
node --report-on-fatalerror dist/server.js
```
