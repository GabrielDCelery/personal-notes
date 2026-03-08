# TypeScript Error Handling

## Why

- **throw is untyped** — TypeScript cannot track what a function throws. Unlike Go's `(val, error)` return, you have no compiler help. This makes consistent error handling patterns even more important.
- **catch gives you unknown** — Since TS 4.4, catch clause variables are `unknown` by default. You must narrow before accessing `.message` or any properties.
- **Custom error classes** — Extending Error gives you instanceof checks, proper stack traces, and structured data. Much better than throwing strings or plain objects.
- **Operational vs programmer errors** — Operational errors (network timeout, file not found) should be caught and handled. Programmer errors (TypeError, null reference) should crash — they indicate bugs.
- **Error codes over message matching** — Never match on `error.message` — it's fragile and changes between versions. Use a `code` property or custom error class for programmatic handling.

## Quick Reference

| Use case              | Method                                 |
| --------------------- | -------------------------------------- |
| Throw an error        | `throw new Error("message")`           |
| Catch and narrow      | `catch (err) { if (err instanceof...)` |
| Custom error class    | `class AppError extends Error`         |
| Type guard for errors | `err instanceof Error`                 |
| Wrap / add context    | `new Error("context", { cause: err })` |
| Access original error | `error.cause`                          |

## Basics

### 1. Catching and narrowing unknown errors

```typescript
try {
  await fetchData();
} catch (err: unknown) {
  if (err instanceof Error) {
    console.error(err.message);
    console.error(err.stack);
  } else {
    console.error("Unknown error:", err);
  }
}
```

### 2. Error cause (wrapping with context)

```typescript
async function getUser(id: string): Promise<User> {
  try {
    const res = await fetch(`/api/users/${id}`);
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    return await res.json();
  } catch (err) {
    throw new Error(`Failed to get user ${id}`, { cause: err });
  }
}

// Access the chain
try {
  await getUser("123");
} catch (err) {
  console.error(err.message); // "Failed to get user 123"
  console.error((err as Error).cause); // original error
}
```

## Custom Error Classes

### 3. Base application error

```typescript
class AppError extends Error {
  constructor(
    message: string,
    public readonly code: string,
    public readonly statusCode: number = 500,
    options?: ErrorOptions,
  ) {
    super(message, options);
    this.name = "AppError";
  }
}
```

### 4. Specific error types

```typescript
class NotFoundError extends AppError {
  constructor(resource: string, id: string) {
    super(`${resource} ${id} not found`, "NOT_FOUND", 404);
    this.name = "NotFoundError";
  }
}

class ValidationError extends AppError {
  constructor(
    message: string,
    public readonly field: string,
  ) {
    super(message, "VALIDATION_ERROR", 400);
    this.name = "ValidationError";
  }
}

class ConflictError extends AppError {
  constructor(message: string) {
    super(message, "CONFLICT", 409);
    this.name = "ConflictError";
  }
}
```

### 5. Using instanceof to handle errors

```typescript
try {
  await createUser(data);
} catch (err) {
  if (err instanceof ValidationError) {
    return res.status(400).json({ error: err.message, field: err.field });
  }
  if (err instanceof ConflictError) {
    return res.status(409).json({ error: err.message });
  }
  if (err instanceof AppError) {
    return res
      .status(err.statusCode)
      .json({ error: err.message, code: err.code });
  }
  throw err; // rethrow unknown errors
}
```

## Type-Safe Error Patterns

### 6. Result type (Go-style returns)

```typescript
type Result<T, E = Error> = { ok: true; value: T } | { ok: false; error: E };

async function getUser(id: string): Promise<Result<User>> {
  try {
    const user = await db.users.findUnique({ where: { id } });
    if (!user) return { ok: false, error: new NotFoundError("User", id) };
    return { ok: true, value: user };
  } catch (err) {
    return {
      ok: false,
      error: err instanceof Error ? err : new Error(String(err)),
    };
  }
}

const result = await getUser("123");
if (!result.ok) {
  console.error(result.error);
  return;
}
console.log(result.value.name); // fully typed
```

### 7. Type guard for error narrowing

```typescript
function isAppError(err: unknown): err is AppError {
  return err instanceof AppError;
}

function isErrorWithCode(err: unknown, code: string): err is AppError {
  return err instanceof AppError && err.code === code;
}

try {
  await doSomething();
} catch (err) {
  if (isErrorWithCode(err, "NOT_FOUND")) {
    // err is typed as AppError here
  }
}
```

## Patterns

### 8. Express/Fastify error handler middleware

```typescript
function errorHandler(
  err: Error,
  req: Request,
  res: Response,
  next: NextFunction,
) {
  if (err instanceof AppError) {
    res.status(err.statusCode).json({
      error: { code: err.code, message: err.message },
    });
    return;
  }

  // Unexpected error — log full details, return generic message
  console.error(err);
  res.status(500).json({
    error: { code: "INTERNAL_ERROR", message: "Internal server error" },
  });
}
```

### 9. Async wrapper (avoid try/catch in every route)

```typescript
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
    const user = await getUser(req.params.id); // errors go to error handler
    res.json(user);
  }),
);
```
