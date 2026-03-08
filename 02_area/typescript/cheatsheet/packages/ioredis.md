# ioredis

```sh
npm install ioredis
```

> Full-featured Redis client for Node.js with built-in Cluster, Sentinel, and pipeline support.

## Setup

```typescript
import Redis from "ioredis";

const redis = new Redis({
  host: "localhost",
  port: 6379,
  password: undefined,
  db: 0,
  maxRetriesPerRequest: 3,
});

// Or from URL
const redis = new Redis("redis://:password@localhost:6379/0");

// Verify connection
redis.on("connect", () => console.log("Redis connected"));
redis.on("error", (err) => console.error("Redis error:", err));
```

## Get / Set

```typescript
// Set with expiry (seconds)
await redis.set("key", "value", "EX", 600); // 10 minutes

// Get
const val = await redis.get("key"); // string | null

// Set only if not exists (lock pattern)
const acquired = await redis.set("lock:resource", "1", "EX", 30, "NX");
// "OK" if set, null if already exists

// Set and get old value
const old = await redis.getset("key", "new-value");

// Multiple
await redis.mset("k1", "v1", "k2", "v2");
const [v1, v2] = await redis.mget("k1", "k2");
```

## Delete / Expire

```typescript
await redis.del("key1", "key2");

await redis.expire("key", 300); // set TTL (seconds)
const ttl = await redis.ttl("key"); // -1 if no expiry, -2 if not exists

await redis.exists("key"); // 1 if exists, 0 if not
```

## JSON (store objects)

```typescript
await redis.set("user:1", JSON.stringify(user), "EX", 3600);

const raw = await redis.get("user:1");
const user = raw ? JSON.parse(raw) : null;
```

## Hash

```typescript
// Set fields
await redis.hset("user:1", { name: "Alice", email: "alice@example.com" });

// Get field
const name = await redis.hget("user:1", "name");

// Get all fields
const fields = await redis.hgetall("user:1");
// { name: "Alice", email: "alice@example.com" }

// Delete field
await redis.hdel("user:1", "email");
```

## List

```typescript
await redis.lpush("queue", "item1", "item2"); // push left
await redis.rpush("queue", "item3"); // push right

const val = await redis.lpop("queue"); // pop left
const val = await redis.rpop("queue"); // pop right

// Blocking pop (wait up to 5 seconds)
const result = await redis.blpop("queue", 5); // [key, value] | null

// Range
const items = await redis.lrange("queue", 0, -1); // all items
```

## Set

```typescript
await redis.sadd("tags", "typescript", "node", "redis");
const members = await redis.smembers("tags");
const isMember = await redis.sismember("tags", "typescript"); // 1 or 0
await redis.srem("tags", "redis");
```

## Pub / Sub

```typescript
// Publisher (use main client)
await redis.publish(
  "channel",
  JSON.stringify({ event: "user:created", data: user }),
);

// Subscriber (must use separate connection)
const sub = new Redis();
await sub.subscribe("channel");

sub.on("message", (channel, message) => {
  const payload = JSON.parse(message);
  console.log(`${channel}:`, payload);
});
```

## Pipeline (batch commands)

```typescript
const pipeline = redis.pipeline();

pipeline.set("k1", "v1");
pipeline.set("k2", "v2");
pipeline.incr("counter");
pipeline.get("counter");

const results = await pipeline.exec();
// [[null, "OK"], [null, "OK"], [null, 1], [null, "1"]]
```

## Increment / Decrement

```typescript
await redis.incr("counter"); // +1
await redis.incrby("counter", 5); // +5
await redis.decr("counter"); // -1
await redis.incrbyfloat("score", 1.5); // +1.5
```

## Scan (iterate keys safely)

```typescript
const stream = redis.scanStream({ match: "user:*", count: 100 });

for await (const keys of stream) {
  for (const key of keys) {
    console.log(key);
  }
}
```

Never use `KEYS *` in production — it blocks. Use `scanStream` instead.

## Graceful shutdown

```typescript
process.on("SIGTERM", async () => {
  await redis.quit(); // waits for pending commands
});
```
