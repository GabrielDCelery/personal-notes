# Vitest

```sh
npm install -D vitest
```

> Fast unit test framework powered by Vite. Compatible with Jest API. Native TypeScript and ESM support.

## Setup

```json
// package.json
{
  "scripts": {
    "test": "vitest run",
    "test:watch": "vitest",
    "test:coverage": "vitest run --coverage"
  }
}
```

```typescript
// vitest.config.ts (optional — works without config)
import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    globals: true, // no need to import describe/it/expect
    coverage: {
      provider: "v8",
      reporter: ["text", "lcov"],
    },
  },
});
```

## Basic Tests

```typescript
import { describe, it, expect } from "vitest";

describe("add", () => {
  it("adds two numbers", () => {
    expect(add(2, 3)).toBe(5);
  });

  it("handles negative numbers", () => {
    expect(add(-1, 1)).toBe(0);
  });
});
```

## Common Assertions

```typescript
// Equality
expect(value).toBe(5); // strict ===
expect(obj).toEqual({ a: 1, b: 2 }); // deep equality
expect(obj).toStrictEqual({ a: 1 }); // deep + no extra properties

// Truthiness
expect(value).toBeTruthy();
expect(value).toBeFalsy();
expect(value).toBeNull();
expect(value).toBeUndefined();
expect(value).toBeDefined();

// Numbers
expect(value).toBeGreaterThan(3);
expect(value).toBeLessThanOrEqual(10);
expect(0.1 + 0.2).toBeCloseTo(0.3);

// Strings
expect(str).toMatch(/pattern/);
expect(str).toContain("substring");

// Arrays / Objects
expect(arr).toContain(item);
expect(arr).toHaveLength(3);
expect(obj).toHaveProperty("key");
expect(obj).toMatchObject({ a: 1 }); // partial match

// Errors
expect(() => fn()).toThrow();
expect(() => fn()).toThrow("message");
expect(() => fn()).toThrow(CustomError);
```

## Async Tests

```typescript
it("fetches user", async () => {
  const user = await getUser("123");
  expect(user.name).toBe("Alice");
});

it("rejects on not found", async () => {
  await expect(getUser("bad-id")).rejects.toThrow("not found");
});

it("resolves with user", async () => {
  await expect(getUser("123")).resolves.toEqual({ id: "123", name: "Alice" });
});
```

## Setup / Teardown

```typescript
describe("UserService", () => {
  let db: Database;

  beforeAll(async () => {
    db = await connectTestDb();
  });

  afterAll(async () => {
    await db.close();
  });

  beforeEach(async () => {
    await db.clear();
  });

  it("creates a user", async () => {
    // ...
  });
});
```

## Mocking

```typescript
import { vi } from "vitest";

// Mock a function
const fn = vi.fn();
fn("arg");
expect(fn).toHaveBeenCalledWith("arg");
expect(fn).toHaveBeenCalledTimes(1);

// Mock return value
const fn = vi.fn().mockReturnValue(42);
const fn = vi.fn().mockResolvedValue({ id: "1" });
const fn = vi.fn().mockRejectedValue(new Error("fail"));

// Mock implementation
const fn = vi.fn((x: number) => x * 2);
```

## Mocking Modules

```typescript
import { vi } from "vitest";

// Mock an entire module
vi.mock("./db", () => ({
  getUser: vi.fn().mockResolvedValue({ id: "1", name: "Alice" }),
}));

// Partial mock — keep real implementations for other exports
vi.mock("./utils", async (importOriginal) => {
  const actual = await importOriginal<typeof import("./utils")>();
  return {
    ...actual,
    sendEmail: vi.fn(), // only mock sendEmail
  };
});

// Reset mocks between tests
afterEach(() => {
  vi.restoreAllMocks();
});
```

## Spying

```typescript
import { vi } from "vitest";

const spy = vi.spyOn(console, "log");
doSomething();
expect(spy).toHaveBeenCalledWith("expected output");
spy.mockRestore();

// Spy on object method
const spy = vi.spyOn(userService, "create");
```

## Test Filtering

```sh
vitest run --reporter=verbose       # detailed output
vitest run src/utils.test.ts        # specific file
vitest run -t "creates a user"      # by test name
```

```typescript
it.only("runs only this test", () => {
  /* ... */
});
it.skip("skipped", () => {
  /* ... */
});
it.todo("not implemented yet");
```

## Coverage

```sh
npm install -D @vitest/coverage-v8

vitest run --coverage
```

```typescript
// vitest.config.ts
export default defineConfig({
  test: {
    coverage: {
      provider: "v8",
      reporter: ["text", "lcov", "html"],
      include: ["src/**"],
      exclude: ["src/**/*.test.ts", "src/types/**"],
      thresholds: {
        lines: 80,
        branches: 80,
      },
    },
  },
});
```

## vitest vs jest

|            | vitest          | jest            |
| ---------- | --------------- | --------------- |
| TypeScript | native          | needs transform |
| ESM        | native          | experimental    |
| Speed      | fast (Vite)     | slower          |
| API        | Jest-compatible | original        |
| Config     | vite.config     | jest.config     |
| Watch mode | instant (HMR)   | file-based      |
