# TypeScript EventEmitter

## Why

- **EventEmitter is Node's pub/sub** — Core pattern throughout Node.js. Streams, HTTP servers, child processes, and many libraries extend EventEmitter. Understanding it is essential.
- **Typed events prevent bugs** — Vanilla EventEmitter uses `string` event names and `any` args. Use a typed wrapper or declaration merging to get compile-time safety on event names and payloads.
- **removeListener to prevent leaks** — Every `.on()` registers a persistent listener. If you add listeners in a loop or per-request without removing them, you leak memory. Node warns at 10 listeners.
- **once for one-shot events** — `emitter.once()` auto-removes after the first call. Also available as `events.once(emitter, "event")` which returns a promise — great for startup/shutdown signals.
- **Error events are special** — If an `"error"` event is emitted with no listener, Node throws and crashes. Always add an error handler.

## Quick Reference

| Use case             | Method                                   |
| -------------------- | ---------------------------------------- |
| Listen               | `emitter.on("event", handler)`           |
| Listen once          | `emitter.once("event", handler)`         |
| Emit                 | `emitter.emit("event", ...args)`         |
| Remove listener      | `emitter.off("event", handler)`          |
| Remove all           | `emitter.removeAllListeners("event")`    |
| Await one event      | `events.once(emitter, "event")`          |
| Async iterate        | `events.on(emitter, "event")`            |
| Listener count       | `emitter.listenerCount("event")`         |

## Basics

### 1. Create and use an EventEmitter

```typescript
import { EventEmitter } from "node:events";

const emitter = new EventEmitter();

emitter.on("message", (text: string) => {
  console.log("Received:", text);
});

emitter.emit("message", "hello"); // "Received: hello"
```

### 2. once — auto-removes after first call

```typescript
emitter.once("ready", () => {
  console.log("Ready — this only fires once");
});

emitter.emit("ready"); // fires
emitter.emit("ready"); // nothing
```

### 3. Remove a listener

```typescript
function handler(data: string) {
  console.log(data);
}

emitter.on("data", handler);
emitter.off("data", handler); // remove specific listener
emitter.removeAllListeners("data"); // remove all for this event
```

## Typed EventEmitter

### 4. Type-safe events with interface

```typescript
import { EventEmitter } from "node:events";

interface AppEvents {
  "user:created": [user: { id: string; name: string }];
  "user:deleted": [userId: string];
  "error": [error: Error];
}

class AppEmitter extends EventEmitter<AppEvents> {}

const app = new AppEmitter();

app.on("user:created", (user) => {
  console.log(user.name); // fully typed
});

app.emit("user:created", { id: "1", name: "Alice" }); // type-checked
// app.emit("user:created", "wrong"); // compile error
```

### 5. Typed emitter as class member

```typescript
import { EventEmitter } from "node:events";

interface OrderEvents {
  "placed": [order: Order];
  "shipped": [orderId: string, trackingNumber: string];
  "error": [error: Error];
}

class OrderService extends EventEmitter<OrderEvents> {
  async placeOrder(data: CreateOrderData): Promise<Order> {
    const order = await this.repository.create(data);
    this.emit("placed", order);
    return order;
  }
}

const orders = new OrderService();
orders.on("placed", (order) => {
  // send confirmation email
});
```

## Promise Integration

### 6. Await a single event

```typescript
import { once } from "node:events";

const server = createServer();
server.listen(3000);
await once(server, "listening");
console.log("Server is ready");
```

### 7. Async iteration over events

```typescript
import { on } from "node:events";

const ac = new AbortController();

// Process events as an async stream
setTimeout(() => ac.abort(), 10_000); // stop after 10s

for await (const [message] of on(emitter, "message", { signal: ac.signal })) {
  console.log("Got:", message);
}
```

## Error Handling

### 8. Always handle error events

```typescript
const emitter = new EventEmitter();

// Without this, an emitted "error" crashes the process
emitter.on("error", (err) => {
  console.error("Emitter error:", err.message);
});

emitter.emit("error", new Error("something broke"));
```

### 9. Max listeners warning

```typescript
// Node warns if you add more than 10 listeners for one event
// Increase if you genuinely need more
emitter.setMaxListeners(20);

// Check current count
emitter.listenerCount("data"); // number
```

## Patterns

### 10. Event-driven service communication

```typescript
import { EventEmitter } from "node:events";

interface DomainEvents {
  "order:placed": [order: Order];
  "payment:received": [payment: Payment];
  "notification:send": [to: string, message: string];
}

// Shared event bus
const bus = new EventEmitter<DomainEvents>();

// Order service emits
bus.emit("order:placed", order);

// Notification service listens
bus.on("order:placed", (order) => {
  bus.emit("notification:send", order.email, `Order ${order.id} confirmed`);
});

// Payment service listens
bus.on("payment:received", (payment) => {
  // update order status
});
```

### 11. Cleanup pattern

```typescript
function setupHandlers(emitter: EventEmitter) {
  const handlers = {
    data: (chunk: Buffer) => process(chunk),
    error: (err: Error) => console.error(err),
    end: () => console.log("done"),
  };

  emitter.on("data", handlers.data);
  emitter.on("error", handlers.error);
  emitter.once("end", handlers.end);

  // Return cleanup function
  return () => {
    emitter.off("data", handlers.data);
    emitter.off("error", handlers.error);
    emitter.off("end", handlers.end);
  };
}

const cleanup = setupHandlers(stream);
// later...
cleanup();
```
