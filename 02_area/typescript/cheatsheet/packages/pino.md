# Pino

```sh
npm install pino
npm install -D pino-pretty    # human-readable output for dev
```

> High-performance structured JSON logger. Fastest Node.js logger — 5x+ faster than winston.

## Setup

```typescript
import pino from "pino";

// Production (JSON output)
const logger = pino({ level: "info" });

// Development (human-readable)
const logger = pino({
  transport: {
    target: "pino-pretty",
    options: { colorize: true },
  },
});
```

## Logging

```typescript
logger.info("server started");
logger.info({ port: 3000 }, "server started");
logger.error({ err }, "database connection failed");
logger.warn({ userId, attempt: 3 }, "rate limit approaching");
logger.debug({ query, params }, "executing query");

// Levels: trace, debug, info, warn, error, fatal
```

## Child loggers (add context)

```typescript
// Add fields to all subsequent logs
const requestLogger = logger.child({
  requestId: req.id,
  userId: req.user?.id,
});

requestLogger.info("handler started");
requestLogger.info({ orderId }, "order created");
// Both logs include requestId and userId automatically
```

## Configuration

```typescript
const logger = pino({
  level: process.env.LOG_LEVEL ?? "info",
  formatters: {
    level(label) {
      return { level: label }; // "info" instead of numeric 30
    },
  },
  timestamp: pino.stdTimeFunctions.isoTime,
  redact: ["password", "token", "authorization", "*.secret"],
});
```

## Express integration

```sh
npm install pino-http
```

```typescript
import pinoHttp from "pino-http";

app.use(pinoHttp({ logger }));

// Access in route handlers
app.get("/users", (req, res) => {
  req.log.info("fetching users"); // includes request context
  // ...
});
```

## Fastify integration

```typescript
// Fastify has pino built in
const app = Fastify({
  logger: {
    level: "info",
    transport:
      process.env.NODE_ENV === "development"
        ? { target: "pino-pretty" }
        : undefined,
  },
});

app.get("/", async (request) => {
  request.log.info("handling request"); // pino logger
});
```

## Log to file

```typescript
import { createWriteStream } from "node:fs";

const stream = createWriteStream("./app.log", { flags: "a" });
const logger = pino(stream);

// Multiple destinations
import pino from "pino";

const logger = pino(
  pino.transport({
    targets: [
      { target: "pino-pretty", level: "info", options: { destination: 1 } }, // stdout
      {
        target: "pino/file",
        level: "error",
        options: { destination: "./error.log" },
      },
    ],
  }),
);
```

## Error logging

```typescript
// Always pass errors as { err } — pino serializes stack traces
try {
  await doWork();
} catch (err) {
  logger.error({ err }, "operation failed"); // includes stack trace
}

// Custom error serializer is built in — no config needed
// Output: { "err": { "type": "Error", "message": "...", "stack": "..." } }
```

## Redaction (hide sensitive data)

```typescript
const logger = pino({
  redact: {
    paths: ["password", "*.token", "headers.authorization"],
    censor: "[REDACTED]",
  },
});

logger.info({ password: "secret123", user: "alice" }, "login");
// { "password": "[REDACTED]", "user": "alice", "msg": "login" }
```

## pino vs winston

|              | pino                              | winston                   |
| ------------ | --------------------------------- | ------------------------- |
| Performance  | fastest                           | slower                    |
| Output       | JSON (structured)                 | configurable              |
| Pretty print | pino-pretty (dev)                 | built-in                  |
| API          | simple                            | more features             |
| Fastify      | built-in                          | manual setup              |
| When to use  | perf-critical, structured logging | legacy, custom transports |
