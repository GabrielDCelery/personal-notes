# TypeScript Async/Await

## Why

- **Promises are the foundation** — Every async operation in Node.js returns a Promise. async/await is syntactic sugar over Promises — it doesn't replace them, it makes them readable.
- **await only pauses the current function** — Other code keeps running. This is fundamentally different from Go goroutines where each has its own stack. Node.js has one thread and an event loop.
- **Always handle rejections** — An unhandled promise rejection crashes the process in Node 15+. Every await needs a try/catch or a .catch() somewhere up the chain.
- **Promise.all vs sequential awaits** — Awaiting in a loop runs requests one at a time. Promise.all runs them concurrently. For independent operations, always use Promise.all.
- **Promise.allSettled vs Promise.all** — all fails fast on the first rejection. allSettled waits for everything and tells you which succeeded and which failed. Use allSettled when partial results are acceptable.

## Quick Reference

| Use case                 | Method                                   |
| ------------------------ | ---------------------------------------- |
| Wait for one             | `await promise`                          |
| Wait for all (fail fast) | `await Promise.all([...])`               |
| Wait for all (no fail)   | `await Promise.allSettled([...])`        |
| First to resolve         | `await Promise.race([...])`              |
| First to succeed         | `await Promise.any([...])`               |
| Create resolved          | `Promise.resolve(value)`                 |
| Create rejected          | `Promise.reject(error)`                  |
| Delay / sleep            | `await setTimeout(ms)` (timers/promises) |

## Basics

### 1. Basic async/await

```typescript
async function getUser(id: string): Promise<User> {
  const response = await fetch(`/api/users/${id}`);
  if (!response.ok) {
    throw new Error(`Failed to fetch user: ${response.status}`);
  }
  return response.json();
}
```

### 2. Error handling with try/catch

```typescript
async function getUser(id: string): Promise<User | null> {
  try {
    const response = await fetch(`/api/users/${id}`);
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    return await response.json();
  } catch (err) {
    console.error("Failed to fetch user:", err);
    return null;
  }
}
```

### 3. Sequential vs concurrent

```typescript
// BAD — sequential, takes sum of all request times
const user = await getUser(id);
const orders = await getOrders(id);
const prefs = await getPreferences(id);

// GOOD — concurrent, takes time of the slowest request
const [user, orders, prefs] = await Promise.all([
  getUser(id),
  getOrders(id),
  getPreferences(id),
]);
```

## Promise Combinators

### 4. Promise.all — fail fast

```typescript
try {
  const [users, products] = await Promise.all([fetchUsers(), fetchProducts()]);
} catch (err) {
  // first rejection cancels everything
}
```

### 5. Promise.allSettled — partial results

```typescript
const results = await Promise.allSettled([
  fetchUsers(),
  fetchProducts(),
  fetchOrders(),
]);

for (const result of results) {
  if (result.status === "fulfilled") {
    console.log(result.value);
  } else {
    console.error(result.reason);
  }
}
```

### 6. Promise.race — timeout pattern

```typescript
async function withTimeout<T>(promise: Promise<T>, ms: number): Promise<T> {
  const timeout = new Promise<never>((_, reject) =>
    setTimeout(() => reject(new Error("Timeout")), ms),
  );
  return Promise.race([promise, timeout]);
}

const user = await withTimeout(getUser(id), 5000);
```

### 7. Promise.any — first success

```typescript
// Try multiple mirrors, use whichever responds first
const data = await Promise.any([
  fetchFrom("https://primary.example.com/data"),
  fetchFrom("https://mirror1.example.com/data"),
  fetchFrom("https://mirror2.example.com/data"),
]);
```

## Patterns

### 8. Async iteration

```typescript
async function processItems(ids: string[]) {
  // Sequential — when order matters or you need backpressure
  for (const id of ids) {
    await processItem(id);
  }
}
```

### 9. Controlled concurrency (batching)

```typescript
async function processInBatches<T>(
  items: T[],
  batchSize: number,
  fn: (item: T) => Promise<void>,
) {
  for (let i = 0; i < items.length; i += batchSize) {
    const batch = items.slice(i, i + batchSize);
    await Promise.all(batch.map(fn));
  }
}

await processInBatches(userIds, 10, processUser);
```

### 10. Sleep / delay

```typescript
import { setTimeout } from "node:timers/promises";

await setTimeout(1000); // sleep 1 second

// Retry with delay
async function retry<T>(
  fn: () => Promise<T>,
  attempts: number,
  delayMs: number,
): Promise<T> {
  for (let i = 0; i < attempts; i++) {
    try {
      return await fn();
    } catch (err) {
      if (i === attempts - 1) throw err;
      await setTimeout(delayMs);
    }
  }
  throw new Error("unreachable");
}
```
