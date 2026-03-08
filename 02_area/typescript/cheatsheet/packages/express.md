# Express

```sh
npm install express
npm install -D @types/express
```

> Minimal, unopinionated web framework. The most widely used Node.js HTTP framework.

## Setup

```typescript
import express from "express";

const app = express();

app.use(express.json()); // parse JSON bodies

app.get("/health", (req, res) => {
  res.json({ status: "ok" });
});

app.listen(3000, () => {
  console.log("Listening on :3000");
});
```

## Routes

```typescript
app.get("/items", listItems);
app.post("/items", createItem);
app.put("/items/:id", updateItem);
app.delete("/items/:id", deleteItem);

// Route param
app.get("/items/:id", (req, res) => {
  const { id } = req.params;
});

// Query param
app.get("/items", (req, res) => {
  const page = req.query.page ?? "1"; // string | undefined
  const limit = req.query.limit;
});
```

## JSON response

```typescript
app.get("/items/:id", (req, res) => {
  const item = { id: 1, name: "Alice" };
  res.json(item); // sets Content-Type and serializes

  // With status code
  res.status(201).json(item);

  // Error response
  res.status(404).json({ error: "not found" });
});
```

## Request body

```typescript
app.post("/items", (req, res) => {
  const { name, email } = req.body; // any — validate before using
  // ...
});
```

## Middleware

```typescript
// Global
app.use(express.json());
app.use(express.urlencoded({ extended: true }));

// Custom middleware
function authMiddleware(req: Request, res: Response, next: NextFunction) {
  const token = req.headers.authorization;
  if (!token) {
    res.status(401).json({ error: "unauthorized" });
    return;
  }
  // attach user to request
  (req as any).user = verifyToken(token);
  next();
}

// Apply to specific routes
app.use("/api", authMiddleware);
```

## Router (group routes)

```typescript
import { Router } from "express";

const router = Router();

router.get("/", listUsers);
router.post("/", createUser);
router.get("/:id", getUser);

// Mount on app
app.use("/api/v1/users", router);

// With middleware on group
app.use("/api/v1", authMiddleware, router);
```

## Error handling

```typescript
// Async wrapper — errors go to error handler
function asyncHandler(
  fn: (req: Request, res: Response, next: NextFunction) => Promise<void>,
) {
  return (req: Request, res: Response, next: NextFunction) => {
    fn(req, res, next).catch(next);
  };
}

app.get(
  "/users/:id",
  asyncHandler(async (req, res) => {
    const user = await getUser(req.params.id);
    if (!user) throw new NotFoundError("User", req.params.id);
    res.json(user);
  }),
);

// Error handler middleware (must have 4 params)
app.use((err: Error, req: Request, res: Response, next: NextFunction) => {
  if (err instanceof AppError) {
    res.status(err.statusCode).json({ error: err.message, code: err.code });
    return;
  }
  console.error(err);
  res.status(500).json({ error: "Internal server error" });
});
```

## Request logging

```typescript
import morgan from "morgan";

app.use(morgan("combined")); // production
app.use(morgan("dev")); // development
```

## CORS

```typescript
import cors from "cors";

app.use(cors()); // allow all
app.use(cors({ origin: "https://example.com", credentials: true }));
```

## Graceful shutdown

```typescript
const server = app.listen(3000);

process.on("SIGTERM", () => {
  server.close(() => {
    console.log("Server closed");
    process.exit(0);
  });
});
```

## express vs fastify

|                | express             | fastify                      |
| -------------- | ------------------- | ---------------------------- |
| API style      | callback-based      | async-first                  |
| Validation     | manual / middleware | built-in (JSON Schema / zod) |
| Performance    | good                | faster                       |
| Ecosystem      | massive             | growing                      |
| TypeScript     | @types/express      | built-in types               |
| Learning curve | low                 | low                          |
