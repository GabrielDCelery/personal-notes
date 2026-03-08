# TypeScript Types

## Why

- **Interfaces vs type aliases** — Both can describe object shapes. Interfaces support declaration merging and are better for public APIs. Type aliases handle unions, intersections, and mapped types. For most backend code, either works — just be consistent.
- **unknown vs any** — `any` disables type checking. `unknown` forces you to narrow before use. Always prefer `unknown` for values with uncertain types (API responses, catch clauses, user input).
- **Discriminated unions** — A union with a shared literal field (`type`, `kind`, `status`) that TypeScript can narrow on in switch/if blocks. The compiler ensures you handle every variant.
- **Utility types** — Built-in types like `Partial`, `Pick`, `Omit` derive new types from existing ones. Avoids duplicating type definitions and keeps them in sync.
- **as const** — Narrows a value to its literal type. Turns `"admin"` from `string` to the literal `"admin"`. Essential for discriminated unions and config objects.

## Quick Reference

| Use case                | Method                             |
| ----------------------- | ---------------------------------- |
| Object shape            | `interface` or `type`              |
| Union of types          | `type A = B \| C`                  |
| Intersection            | `type A = B & C`                   |
| Narrow unknown          | `typeof`, `instanceof`, type guard |
| Make fields optional    | `Partial<T>`                       |
| Make fields required    | `Required<T>`                      |
| Pick fields             | `Pick<T, "a" \| "b">`              |
| Remove fields           | `Omit<T, "a" \| "b">`              |
| Record type             | `Record<string, T>`                |
| Extract from union      | `Extract<Union, Condition>`        |
| Exclude from union      | `Exclude<Union, Condition>`        |
| Return type of function | `ReturnType<typeof fn>`            |

## Basics

### 1. Interfaces and type aliases

```typescript
// Interface — use for object shapes, especially public APIs
interface User {
  id: string;
  name: string;
  email: string;
}

// Type alias — use when you need unions, intersections, or mapped types
type Status = "active" | "inactive" | "suspended";

type UserWithStatus = User & { status: Status };
```

### 2. unknown vs any

```typescript
// BAD — any disables all checking
function parse(input: any) {
  return input.data.items; // no error, blows up at runtime
}

// GOOD — unknown forces narrowing
function parse(input: unknown) {
  if (typeof input === "object" && input !== null && "data" in input) {
    // safely narrowed
  }
}
```

### 3. Type narrowing

```typescript
function handle(value: string | number) {
  if (typeof value === "string") {
    console.log(value.toUpperCase()); // string methods available
  } else {
    console.log(value.toFixed(2)); // number methods available
  }
}
```

## Discriminated Unions

### 4. Tagged union pattern

```typescript
type Event =
  | { type: "USER_CREATED"; payload: { userId: string; name: string } }
  | { type: "ORDER_PLACED"; payload: { orderId: string; total: number } }
  | { type: "PAYMENT_FAILED"; payload: { orderId: string; reason: string } };

function handleEvent(event: Event) {
  switch (event.type) {
    case "USER_CREATED":
      console.log(event.payload.name); // typed as { userId, name }
      break;
    case "ORDER_PLACED":
      console.log(event.payload.total); // typed as { orderId, total }
      break;
    case "PAYMENT_FAILED":
      console.log(event.payload.reason); // typed as { orderId, reason }
      break;
  }
}
```

### 5. Exhaustive checking

```typescript
function assertNever(value: never): never {
  throw new Error(`Unhandled value: ${value}`);
}

function handleEvent(event: Event) {
  switch (event.type) {
    case "USER_CREATED":
      // ...
      break;
    case "ORDER_PLACED":
      // ...
      break;
    case "PAYMENT_FAILED":
      // ...
      break;
    default:
      assertNever(event); // compile error if a case is missing
  }
}
```

## Utility Types

### 6. Partial, Required, Pick, Omit

```typescript
interface User {
  id: string;
  name: string;
  email: string;
  role: string;
}

type UpdateUser = Partial<Omit<User, "id">>; // all fields optional except id is removed
type CreateUser = Omit<User, "id">; // everything except id
type UserSummary = Pick<User, "id" | "name">; // only id and name
```

### 7. Record

```typescript
type Role = "admin" | "user" | "guest";

const permissions: Record<Role, string[]> = {
  admin: ["read", "write", "delete"],
  user: ["read", "write"],
  guest: ["read"],
};
```

### 8. Extract and Exclude

```typescript
type Status = "active" | "inactive" | "suspended" | "deleted";

type ActiveStatuses = Extract<Status, "active" | "inactive">; // "active" | "inactive"
type ArchivedStatuses = Exclude<Status, "active" | "inactive">; // "suspended" | "deleted"
```

## Custom Type Guards

### 9. Type predicates

```typescript
interface ApiResponse {
  data: unknown;
}

function isUser(value: unknown): value is User {
  return (
    typeof value === "object" &&
    value !== null &&
    "id" in value &&
    "name" in value &&
    "email" in value
  );
}

const response: ApiResponse = await fetchData();
if (isUser(response.data)) {
  console.log(response.data.email); // fully typed as User
}
```

## Generics

### 10. Generic function and constraints

```typescript
// Basic generic
function first<T>(items: T[]): T | undefined {
  return items[0];
}

// Constrained generic
function getProperty<T, K extends keyof T>(obj: T, key: K): T[K] {
  return obj[key];
}

const user: User = { id: "1", name: "Alice", email: "a@b.com", role: "admin" };
const name = getProperty(user, "name"); // typed as string

// Generic interface
interface Repository<T> {
  findById(id: string): Promise<T | null>;
  findAll(): Promise<T[]>;
  create(data: Omit<T, "id">): Promise<T>;
  update(id: string, data: Partial<T>): Promise<T>;
  delete(id: string): Promise<void>;
}
```
