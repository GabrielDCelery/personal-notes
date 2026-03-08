# Fastify

```sh
npm install fastify
```

> High-performance async-first web framework with built-in validation, serialization, and TypeScript support.

## Setup

```typescript
import Fastify from "fastify";

const app = Fastify({ logger: true });

app.get("/health", async () => {
  return { status: "ok" };
});

await app.listen({ port: 3000, host: "0.0.0.0" });
```

## Routes

```typescript
// Return value is auto-serialized as JSON
app.get("/items", async (request, reply) => {
  return db.items.findAll();
});

// Route params
app.get<{ Params: { id: string } }>("/items/:id", async (request) => {
  const { id } = request.params;
  return db.items.find(id);
});

// Query params
app.get<{ Querystring: { page?: string; limit?: string } }>(
  "/items",
  async (request) => {
    const page = parseInt(request.query.page ?? "1", 10);
    return db.items.findPage(page);
  },
);

// POST with body
app.post<{ Body: { name: string; email: string } }>(
  "/items",
  async (request, reply) => {
    const item = await db.items.create(request.body);
    reply.status(201);
    return item;
  },
);
```

## Validation with JSON Schema

```typescript
app.post("/users", {
  schema: {
    body: {
      type: "object",
      required: ["name", "email"],
      properties: {
        name: { type: "string", minLength: 1 },
        email: { type: "string", format: "email" },
      },
    },
    response: {
      201: {
        type: "object",
        properties: {
          id: { type: "string" },
          name: { type: "string" },
          email: { type: "string" },
        },
      },
    },
  },
  handler: async (request, reply) => {
    const user = await createUser(request.body);
    reply.status(201);
    return user;
  },
});
```

## Validation with zod (via fastify-type-provider-zod)

```sh
npm install fastify-type-provider-zod zod
```

```typescript
import {
  serializerCompiler,
  validatorCompiler,
  ZodTypeProvider,
} from "fastify-type-provider-zod";
import { z } from "zod";

app.setValidatorCompiler(validatorCompiler);
app.setSerializerCompiler(serializerCompiler);

const appWithZod = app.withTypeProvider<ZodTypeProvider>();

appWithZod.post("/users", {
  schema: {
    body: z.object({
      name: z.string().min(1),
      email: z.string().email(),
    }),
  },
  handler: async (request, reply) => {
    // request.body is typed as { name: string; email: string }
    const user = await createUser(request.body);
    reply.status(201);
    return user;
  },
});
```

## Hooks (middleware)

```typescript
// Global — runs before every request
app.addHook("onRequest", async (request, reply) => {
  const token = request.headers.authorization;
  if (!token) {
    reply.status(401).send({ error: "unauthorized" });
    return;
  }
  request.user = verifyToken(token);
});

// Route-specific
app.get("/admin", { preHandler: [authHook, adminHook] }, async (request) => {
  return { admin: true };
});
```

## Plugins (encapsulation)

```typescript
// plugins/users.ts
import { FastifyPluginAsync } from "fastify";

const usersPlugin: FastifyPluginAsync = async (app) => {
  app.get("/", async () => db.users.findAll());
  app.post("/", async (request) => db.users.create(request.body));
  app.get("/:id", async (request) => db.users.find(request.params.id));
};

export default usersPlugin;

// app.ts
app.register(usersPlugin, { prefix: "/api/v1/users" });
```

## Decorators (extend request/reply)

```typescript
// Add typed property to request
declare module "fastify" {
  interface FastifyRequest {
    user: { id: string; role: string };
  }
}

app.decorateRequest("user", null);

app.addHook("onRequest", async (request) => {
  request.user = await authenticate(request);
});
```

## Error handling

```typescript
app.setErrorHandler(async (error, request, reply) => {
  if (error.validation) {
    reply
      .status(400)
      .send({ error: "Validation failed", details: error.validation });
    return;
  }

  request.log.error(error);
  reply.status(500).send({ error: "Internal server error" });
});
```

## Graceful shutdown

```typescript
const signals: NodeJS.Signals[] = ["SIGINT", "SIGTERM"];

for (const signal of signals) {
  process.on(signal, async () => {
    await app.close();
    process.exit(0);
  });
}
```
