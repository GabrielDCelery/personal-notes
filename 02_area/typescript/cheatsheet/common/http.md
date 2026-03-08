# TypeScript HTTP Client

## Why

- **fetch is built-in since Node 18** — No need for axios or node-fetch anymore. The global `fetch` uses undici under the hood and is the standard way to make HTTP requests.
- **fetch doesn't throw on 4xx/5xx** — Just like Go's http.Client, a 404 or 500 response is not an error. You must check `response.ok` or `response.status` yourself.
- **Always set timeouts** — fetch has no default timeout. Use `AbortSignal.timeout()` (Node 18+) to avoid hanging requests in production.
- **response.json() returns any** — The return type is `Promise<any>`. You get zero type safety without explicit typing or runtime validation (e.g. zod).
- **response.body is a ReadableStream** — You can only consume it once. Calling both `.text()` and `.json()` on the same response will fail.

## Quick Reference

| Use case           | Method                                            |
| ------------------ | ------------------------------------------------- |
| Simple GET         | `fetch(url)`                                      |
| POST JSON          | `fetch(url, { method: "POST", body })`            |
| Set headers        | `fetch(url, { headers: {...} })`                  |
| Timeout            | `fetch(url, { signal: AbortSignal.timeout(ms) })` |
| Read JSON response | `await response.json()`                           |
| Read text response | `await response.text()`                           |
| Check success      | `response.ok` or `response.status`                |
| Stream response    | `response.body` (ReadableStream)                  |

## Making HTTP Requests

### 1. Simple GET

```typescript
const response = await fetch("https://example.com/api/items");
if (!response.ok) {
  throw new Error(`HTTP ${response.status}: ${response.statusText}`);
}
const items: Item[] = await response.json();
```

### 2. POST JSON

```typescript
const response = await fetch("https://example.com/api/items", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ name: "Alice" }),
});
if (!response.ok) {
  throw new Error(`HTTP ${response.status}`);
}
const created: Item = await response.json();
```

### 3. Custom headers and auth

```typescript
const response = await fetch("https://example.com/api/items", {
  headers: {
    Authorization: `Bearer ${token}`,
    Accept: "application/json",
  },
});
```

### 4. With timeout

```typescript
const response = await fetch("https://example.com/api/items", {
  signal: AbortSignal.timeout(5000), // 5 second timeout
});
```

### 5. Abort with custom signal

```typescript
const controller = new AbortController();

// Cancel after some condition
someEvent.on("cancel", () => controller.abort());

const response = await fetch("https://example.com/api/items", {
  signal: controller.signal,
});
```

## Response Handling

### 6. Check status and parse

```typescript
const response = await fetch("https://example.com/api/items/1");

switch (response.status) {
  case 200:
    return await response.json();
  case 404:
    return null;
  default:
    throw new Error(`Unexpected status: ${response.status}`);
}
```

### 7. Type-safe response parsing

```typescript
async function fetchJson<T>(url: string, options?: RequestInit): Promise<T> {
  const response = await fetch(url, {
    ...options,
    signal: options?.signal ?? AbortSignal.timeout(10_000),
  });
  if (!response.ok) {
    throw new Error(`HTTP ${response.status}: ${await response.text()}`);
  }
  return response.json() as Promise<T>;
}

const user = await fetchJson<User>("/api/users/1");
```

## Patterns

### 8. Retry with backoff

```typescript
async function fetchWithRetry(
  url: string,
  options?: RequestInit,
  retries = 3,
): Promise<Response> {
  for (let i = 0; i < retries; i++) {
    const response = await fetch(url, options);
    if (response.ok || response.status < 500) return response;

    if (i < retries - 1) {
      await new Promise((r) => setTimeout(r, 2 ** i * 1000));
    }
  }
  throw new Error(`Failed after ${retries} retries: ${url}`);
}
```

### 9. Parallel requests

```typescript
const [users, products] = await Promise.all([
  fetchJson<User[]>("/api/users"),
  fetchJson<Product[]>("/api/products"),
]);
```

### 10. Streaming large responses

```typescript
const response = await fetch("https://example.com/api/large-dataset");
if (!response.ok || !response.body) throw new Error("Failed");

const reader = response.body.getReader();
const decoder = new TextDecoder();

while (true) {
  const { done, value } = await reader.read();
  if (done) break;
  process(decoder.decode(value, { stream: true }));
}
```
