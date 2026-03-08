# TypeScript JSON

## Why

- **JSON.parse returns any** — No type safety at all. You must either cast, assert, or validate at runtime. For external data (API responses, config files, message queues), always validate.
- **JSON.stringify drops undefined** — Properties with `undefined` values are silently removed. Functions and Symbols are also dropped. This can cause subtle bugs when serializing objects.
- **No Date, Map, Set support** — JSON only knows strings, numbers, booleans, null, arrays, and objects. Dates become strings, Maps/Sets become `{}`. You need custom serialization for these.
- **BigInt throws** — `JSON.stringify` throws a TypeError on BigInt values. You must provide a replacer or convert to string/number first.
- **Reviver and replacer** — Built-in hooks for custom serialization/deserialization. Reviver transforms values during parse, replacer during stringify. Use these for dates, BigInt, or any custom type.

## Quick Reference

| Use case             | Method                           |
| -------------------- | -------------------------------- |
| Parse JSON string    | `JSON.parse(str)`                |
| Stringify            | `JSON.stringify(obj)`            |
| Pretty print         | `JSON.stringify(obj, null, 2)`   |
| Custom serialization | `JSON.stringify(obj, replacer)`  |
| Custom parsing       | `JSON.parse(str, reviver)`       |
| Type-safe parse      | validate with zod / manual check |

## Basics

### 1. Parse and stringify

```typescript
const json = '{"id":1,"name":"Alice"}';
const user = JSON.parse(json); // any — no type safety

const str = JSON.stringify({ id: 1, name: "Alice" });
// '{"id":1,"name":"Alice"}'
```

### 2. Type assertion after parse (simple cases)

```typescript
interface User {
  id: number;
  name: string;
  email: string;
}

const user = JSON.parse(raw) as User; // trust the source
```

### 3. Runtime validation (external data)

```typescript
function parseUser(raw: string): User {
  const data: unknown = JSON.parse(raw);

  if (
    typeof data !== "object" ||
    data === null ||
    !("id" in data) ||
    typeof (data as any).id !== "number" ||
    !("name" in data) ||
    typeof (data as any).name !== "string" ||
    !("email" in data) ||
    typeof (data as any).email !== "string"
  ) {
    throw new Error("Invalid user JSON");
  }

  return data as User;
}
```

## Gotchas

### 4. undefined is dropped

```typescript
const obj = { name: "Alice", age: undefined, role: "admin" };
JSON.stringify(obj);
// '{"name":"Alice","role":"admin"}' — age is gone
```

### 5. Dates become strings

```typescript
const obj = { createdAt: new Date("2024-01-15") };
const json = JSON.stringify(obj);
// '{"createdAt":"2024-01-15T00:00:00.000Z"}'

const parsed = JSON.parse(json);
typeof parsed.createdAt; // "string", not Date
```

### 6. BigInt throws

```typescript
JSON.stringify({ id: 123n }); // TypeError: Do not know how to serialize a BigInt

// Fix: use a replacer
JSON.stringify({ id: 123n }, (_, v) =>
  typeof v === "bigint" ? v.toString() : v,
);
```

## Reviver and Replacer

### 7. Reviver — restore dates during parse

```typescript
const ISO_DATE_RE = /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}/;

function dateReviver(_key: string, value: unknown): unknown {
  if (typeof value === "string" && ISO_DATE_RE.test(value)) {
    return new Date(value);
  }
  return value;
}

const data = JSON.parse(raw, dateReviver);
// data.createdAt is now a Date object
```

### 8. Replacer — control what gets serialized

```typescript
// Only include specific fields
JSON.stringify(user, ["id", "name"]);
// '{"id":1,"name":"Alice"}'

// Transform values
JSON.stringify(data, (key, value) => {
  if (key === "password") return undefined; // strip sensitive fields
  if (value instanceof Date) return value.toISOString();
  return value;
});
```

## Patterns

### 9. Safe JSON parse wrapper

```typescript
function safeJsonParse<T>(raw: string): T | null {
  try {
    return JSON.parse(raw) as T;
  } catch {
    return null;
  }
}

const user = safeJsonParse<User>(raw);
if (!user) {
  console.error("Invalid JSON");
}
```

### 10. Read/write JSON files

```typescript
import { readFile, writeFile } from "node:fs/promises";

async function readJson<T>(path: string): Promise<T> {
  const raw = await readFile(path, "utf-8");
  return JSON.parse(raw) as T;
}

async function writeJson(path: string, data: unknown): Promise<void> {
  await writeFile(path, JSON.stringify(data, null, 2) + "\n");
}

const config = await readJson<Config>("./config.json");
await writeJson("./output.json", results);
```
