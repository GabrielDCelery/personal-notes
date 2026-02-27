# Lesson 04: Utility Types & Type Manipulation

Master TypeScript's built-in utility types and advanced type manipulation techniques.

## Built-in Utility Types

You find yourself writing the same type transformations over and over: "take this interface but make all properties optional," or "create a type with these exact keys." TypeScript's utility types are the standard library for type-level operations - pre-built, battle-tested, and universally understood.

Learning these isn't just about saving keystrokes. When you use `Partial<User>` instead of manually marking each property optional, you're writing self-documenting code that other developers instantly recognize. These utilities are the building blocks for more complex type manipulations - master them and you'll understand how TypeScript's type system really works.

### Partial<T>

You have a User interface with 10 required fields. Your update function should accept ANY subset of those fields - maybe just the email, maybe email and age, whatever. Without `Partial`, you'd duplicate the interface and mark everything optional. With `Partial`, one line does it all.

Makes all properties optional.

```typescript
interface User {
  name: string;
  email: string;
  age: number;
}

type PartialUser = Partial<User>;
// { name?: string; email?: string; age?: number; }

// Common use: update functions
function updateUser(id: number, updates: Partial<User>) {
  // Can pass any subset of properties
}

updateUser(1, { age: 30 }); // ✓ Only updating age
```

**Implementation**:

```typescript
type Partial<T> = {
  [P in keyof T]?: T[P];
};
```

### Required<T>

The flip side of Partial. You have configuration with optional fields and defaults, but deep in your validation logic you need to ensure ALL fields are set. `Required` transforms optional properties to required, giving you compile-time proof everything is present.

Makes all properties required (opposite of Partial).

```typescript
interface Config {
  apiUrl?: string;
  timeout?: number;
}

type RequiredConfig = Required<Config>;
// { apiUrl: string; timeout: number; }

// Ensure all fields are set
function validateConfig(config: Required<Config>) {
  // config.apiUrl and config.timeout are guaranteed
}
```

**Implementation**:

```typescript
type Required<T> = {
  [P in keyof T]-?: T[P]; // -? removes optionality
};
```

### Readonly<T>

Immutability isn't just a best practice - it prevents bugs. Pass a config object to a function and you don't want it mutating your data. `Readonly` makes all properties read-only at compile time. The runtime object can still be mutated (JavaScript limitation), but TypeScript prevents accidental mutations in your code.

Makes all properties readonly.

```typescript
interface Todo {
  title: string;
  completed: boolean;
}

const todo: Readonly<Todo> = {
  title: "Learn TypeScript",
  completed: false,
};

todo.completed = true; // ❌ Error: readonly property
```

**Implementation**:

```typescript
type Readonly<T> = {
  readonly [P in keyof T]: T[P];
};
```

### Pick<T, K>

Your User type has password, security questions, and other sensitive fields. Your API returns public user data - just id, name, and email. Instead of maintaining a duplicate PublicUser interface that drifts out of sync, `Pick` selects exactly the fields you need from the source of truth.

Selects subset of properties.

```typescript
interface User {
  id: number;
  name: string;
  email: string;
  password: string;
}

type PublicUser = Pick<User, "id" | "name" | "email">;
// { id: number; name: string; email: string; }

// Return safe user data
function getPublicUser(user: User): PublicUser {
  return {
    id: user.id,
    name: user.name,
    email: user.email,
  };
}
```

**Implementation**:

```typescript
type Pick<T, K extends keyof T> = {
  [P in K]: T[P];
};
```

### Omit<T, K>

Sometimes it's easier to say what you DON'T want. "Give me the User type except for the password." `Omit` is the inverse of `Pick` - exclude specific properties instead of including them. Particularly useful when you have many fields to keep and few to remove.

Removes properties (opposite of Pick).

```typescript
interface User {
  id: number;
  name: string;
  email: string;
  password: string;
}

type UserWithoutPassword = Omit<User, "password">;
// { id: number; name: string; email: string; }

type UserIdOnly = Omit<User, "name" | "email" | "password">;
// { id: number; }
```

**Implementation**:

```typescript
type Omit<T, K extends keyof any> = Pick<T, Exclude<keyof T, K>>;
```

### Record<K, T>

You need an object mapping user roles to permissions. TypeScript's `{ [key: string]: Permissions }` allows ANY string key, so typos silently create new entries. `Record<Role, Permissions>` locks down the exact keys allowed, catching `permissions.superadmin` at compile time when `superadmin` isn't in your Role union.

Creates object type with keys K and values T.

```typescript
type Role = "admin" | "user" | "guest";

const permissions: Record<Role, string[]> = {
  admin: ["read", "write", "delete"],
  user: ["read", "write"],
  guest: ["read"],
};

// Ensures all roles are defined
// permissions.superadmin = [...];  // ❌ Error
```

**Common pattern**: Dictionary/map type

```typescript
const userCache: Record<string, User> = {};
userCache["123"] = { id: 123, name: "Alice" };
```

**Implementation**:

```typescript
type Record<K extends keyof any, T> = {
  [P in K]: T;
};
```

### Exclude<T, U>

Your API can return `'success' | 'error' | 'pending' | null`. But in a specific function, you've already handled null and pending - you only care about success/error. `Exclude` removes unwanted members from a union type, giving you a narrower type to work with.

Removes types from union.

```typescript
type AllTypes = "a" | "b" | "c" | "d";
type Excluded = Exclude<AllTypes, "a" | "c">;
// 'b' | 'd'

// Real-world: Remove null/undefined
type NonNullable<T> = Exclude<T, null | undefined>;

type Result = string | null | undefined;
type SafeResult = NonNullable<Result>; // string
```

**Implementation**:

```typescript
type Exclude<T, U> = T extends U ? never : T;
// Distributes over union
```

### Extract<T, U>

The opposite of Exclude - keep only the union members you want. Your type is `string | number | (() => void)` but you need just the functions. `Extract<Mixed, Function>` filters to only the function type. Think of it as "intersection" for union types.

Keeps only types assignable to U.

```typescript
type AllTypes = "a" | "b" | "c" | "d";
type Extracted = Extract<AllTypes, "a" | "c" | "e">;
// 'a' | 'c'

// Real-world: Extract specific types
type Mixed = string | number | (() => void);
type Functions = Extract<Mixed, Function>; // () => void
```

**Implementation**:

```typescript
type Extract<T, U> = T extends U ? T : never;
```

### ReturnType<T>

Your colleague wrote a function that returns a complex object. You need to work with that same type in your code. Instead of duplicating the interface, `ReturnType` extracts it automatically. Change the function's return value and your types update automatically - single source of truth.

Extracts function return type.

```typescript
function getUser() {
  return { id: 1, name: "Alice" };
}

type User = ReturnType<typeof getUser>;
// { id: number; name: string; }

// Works with generic functions
function createPair<T>(value: T) {
  return { value, timestamp: Date.now() };
}

type Pair = ReturnType<typeof createPair<string>>;
// { value: string; timestamp: number; }
```

**Implementation**:

```typescript
type ReturnType<T extends (...args: any) => any> = T extends (
  ...args: any
) => infer R
  ? R
  : any;
```

### Parameters<T>

You're wrapping a function - calling it with the same arguments plus some logging. TypeScript's rest parameters (`...args: any[]`) lose all type safety. `Parameters` extracts the exact parameter types as a tuple, letting you forward arguments with full type checking.

Extracts function parameter types as tuple.

```typescript
function createUser(name: string, age: number, active: boolean) {
  // ...
}

type CreateUserParams = Parameters<typeof createUser>;
// [name: string, age: number, active: boolean]

// Use to match function signature
function logAndCreate(...args: CreateUserParams) {
  console.log("Creating user with:", args);
  return createUser(...args);
}
```

**Implementation**:

```typescript
type Parameters<T extends (...args: any) => any> = T extends (
  ...args: infer P
) => any
  ? P
  : never;
```

### Awaited<T>

Async/await made Promises easier, but the types got tricky. `ReturnType<typeof fetchUser>` gives you `Promise<User>`, not `User`. You need to unwrap the Promise. `Awaited` does this recursively, even handling `Promise<Promise<T>>` from badly typed code.

Unwraps Promise type (TypeScript 4.5+).

```typescript
type AsyncValue = Promise<string>;
type Value = Awaited<AsyncValue>; // string

// Nested Promises
type Nested = Promise<Promise<number>>;
type Unwrapped = Awaited<Nested>; // number

// Real-world: Infer async function return
async function fetchUser() {
  return { id: 1, name: "Alice" };
}

type User = Awaited<ReturnType<typeof fetchUser>>;
// { id: number; name: string; }
```

## Mapped Types

Utility types are pre-built transformations. Mapped types let you BUILD your own. Need to make every property nullable? Turn every value into an array? Add prefixes to all keys? Mapped types iterate over properties and transform them according to your rules. This is how TypeScript's utility types are implemented - once you master mapped types, you can create custom utilities for your specific needs.

Create new types by transforming properties.

### Basic Mapped Type

```typescript
type Flags<T> = {
  [K in keyof T]: boolean;
};

interface Features {
  darkMode: string;
  notifications: number;
}

type FeatureFlags = Flags<Features>;
// { darkMode: boolean; notifications: boolean; }
```

### Adding Modifiers

```typescript
// Make everything optional and readonly
type Frozen<T> = {
  readonly [K in keyof T]?: T[T];
};

// Remove readonly (using -)
type Mutable<T> = {
  -readonly [K in keyof T]: T[K];
};

// Remove optional (using -)
type Concrete<T> = {
  [K in keyof T]-?: T[K];
};
```

### Key Remapping (TypeScript 4.1+)

The `as` clause in mapped types is a superpower. You can not only transform values but also the KEYS themselves. Generate getter methods? Add prefixes? Filter out certain keys? Key remapping turns property manipulation into key manipulation. This unlocks patterns like auto-generating DTOs, API types, and ORM models from a single source.

```typescript
// Prefix all keys with 'get'
type Getters<T> = {
  [K in keyof T as `get${Capitalize<string & K>}`]: () => T[K];
};

interface Person {
  name: string;
  age: number;
}

type PersonGetters = Getters<Person>;
// {
//   getName: () => string;
//   getAge: () => number;
// }
```

### Filtering Keys

```typescript
// Remove specific keys
type OmitType<T, V> = {
  [K in keyof T as T[K] extends V ? never : K]: T[K];
};

interface Data {
  id: number;
  name: string;
  active: boolean;
  count: number;
}

type NonNumbers = OmitType<Data, number>;
// { name: string; active: boolean; }
```

## Conditional Types

Types can make decisions. "If T is a string, return X, otherwise return Y." This enables type-level logic - the same kind of branching you do at runtime, but at compile time. Conditional types power TypeScript's most advanced features: ReturnType, mapped type filtering, and recursive type definitions.

The real power comes from combining conditionals with `infer` to extract and manipulate types. Once you grasp this, you can write types that feel like functions - taking types as input and producing new types as output.

Types that depend on conditions.

### Basic Conditional

```typescript
type IsString<T> = T extends string ? true : false;

type A = IsString<string>; // true
type B = IsString<number>; // false
```

### With Unions (Distributive)

```typescript
type ToArray<T> = T extends any ? T[] : never;

type Result = ToArray<string | number>;
// string[] | number[] (distributes over union)
```

### Infer Keyword

Here's where type manipulation gets wild. `infer` lets you declare type variables WITHIN a conditional check. "If T is an array, extract and capture the element type." You're pattern matching on types and pulling out the pieces you need. It's how ReturnType works, how you unwrap Promises, and how you can parse complex types at compile time.

Extract types from within other types.

```typescript
// Extract array element type
type ElementType<T> = T extends (infer E)[] ? E : never;

type Str = ElementType<string[]>; // string
type Num = ElementType<number[]>; // number

// Extract function return type (ReturnType implementation)
type GetReturn<T> = T extends (...args: any[]) => infer R ? R : never;

// Extract Promise value
type UnwrapPromise<T> = T extends Promise<infer U> ? U : T;

type A = UnwrapPromise<Promise<string>>; // string
type B = UnwrapPromise<number>; // number
```

### Advanced: DeepPartial

```typescript
type DeepPartial<T> = {
  [K in keyof T]?: T[K] extends object ? DeepPartial<T[K]> : T[K];
};

interface Nested {
  user: {
    profile: {
      name: string;
      age: number;
    };
  };
}

type PartialNested = DeepPartial<Nested>;
// {
//   user?: {
//     profile?: {
//       name?: string;
//       age?: number;
//     };
//   };
// }
```

## Template Literal Types

Types can now manipulate strings like runtime code manipulates values. Build API route types from method + path. Generate event handler names from event types. Validate CSS units. Template literal types bring string manipulation to the type level, enabling incredibly precise typing for string-heavy domains like routing, CSS-in-JS, and configuration.

Combined with unions, they explode into cartesian products - `${Method} ${Route}` generates every possible combination. This creates type-safe DSLs that catch typos and enforce patterns without runtime overhead.

String manipulation at type level (TypeScript 4.1+).

### Basic Template Literals

```typescript
type Greeting = `Hello ${string}`;

const g1: Greeting = "Hello World"; // ✓
const g2: Greeting = "Hi World"; // ❌ Error

// With unions
type Color = "red" | "blue" | "green";
type HexColor = `#${string}`;
type ColorPalette = Color | HexColor;
```

### String Manipulation Utilities

```typescript
type Uppercase<S extends string> = ...;  // Built-in
type Lowercase<S extends string> = ...;  // Built-in
type Capitalize<S extends string> = ...; // Built-in
type Uncapitalize<S extends string> = ...;  // Built-in

type Loud = Uppercase<'hello'>;      // 'HELLO'
type Quiet = Lowercase<'WORLD'>;     // 'world'
type Title = Capitalize<'typescript'>;  // 'Typescript'
```

### API Route Builder

```typescript
type HTTPMethod = "GET" | "POST" | "PUT" | "DELETE";
type Route = "/users" | "/posts" | "/comments";
type APIRoute = `${HTTPMethod} ${Route}`;

// 'GET /users' | 'GET /posts' | ... | 'DELETE /comments' (12 combinations)

function handleRoute(route: APIRoute) {
  // Type-safe route handling
}

handleRoute("GET /users"); // ✓
handleRoute("PATCH /users"); // ❌ Error
```

### Event Names

```typescript
type EventName<T extends string> = `on${Capitalize<T>}`;

type Events = "click" | "focus" | "blur";
type EventHandlers = EventName<Events>;
// 'onClick' | 'onFocus' | 'onBlur'

type Handlers = {
  [E in Events as EventName<E>]: (event: Event) => void;
};
// {
//   onClick: (event: Event) => void;
//   onFocus: (event: Event) => void;
//   onBlur: (event: Event) => void;
// }
```

## Hands-On Exercise 1: Build Custom Utilities

Create these utility types:

```typescript
// 1. PickByType<T, V> - Pick properties of specific type
interface Data {
  id: number;
  name: string;
  active: boolean;
  count: number;
}

type Numbers = PickByType<Data, number>;
// { id: number; count: number; }

// 2. FunctionKeys<T> - Extract function property names
interface API {
  baseUrl: string;
  timeout: number;
  get: () => void;
  post: () => void;
}

type Methods = FunctionKeys<API>;
// 'get' | 'post'
```

<details>
<summary>Solution</summary>

```typescript
// 1. PickByType
type PickByType<T, V> = {
  [K in keyof T as T[K] extends V ? K : never]: T[K];
};

// 2. FunctionKeys
type FunctionKeys<T> = {
  [K in keyof T]: T[K] extends Function ? K : never;
}[keyof T];

// Alternative:
type FunctionKeys<T> = keyof {
  [K in keyof T as T[K] extends Function ? K : never]: T[K];
};
```

</details>

## Hands-On Exercise 2: Event Emitter Types

Build type-safe event emitter:

```typescript
interface Events {
  click: { x: number; y: number };
  submit: { value: string };
  error: { message: string };
}

// Create:
// - EventMap that converts to { 'click': (data: { x, y }) => void, ... }
// - EventNames type union
```

<details>
<summary>Solution</summary>

```typescript
type EventMap<T> = {
  [K in keyof T]: (data: T[K]) => void;
};

type EventNames<T> = keyof T;

// Usage:
class EventEmitter<T> {
  private handlers: Partial<EventMap<T>> = {};

  on<K extends EventNames<T>>(event: K, handler: EventMap<T>[K]) {
    this.handlers[event] = handler;
  }

  emit<K extends EventNames<T>>(event: K, data: T[K]) {
    this.handlers[event]?.(data);
  }
}

const emitter = new EventEmitter<Events>();

emitter.on("click", (data) => {
  console.log(data.x, data.y); // Type-safe: { x, y }
});

emitter.emit("click", { x: 10, y: 20 }); // ✓
emitter.emit("click", { x: 10 }); // ❌ Error: missing y
```

</details>

## Interview Questions

### Q1: Difference between Pick and Omit?

A fundamental question that tests whether you understand type manipulation strategies. It also reveals whether you think about API design - when to be explicit about inclusion versus exclusion. The follow-up often is "which would you use for this specific case?"

<details>
<summary>Answer</summary>

- **Pick<T, K>**: Include only specified keys
- **Omit<T, K>**: Exclude specified keys

```typescript
interface User {
  id: number;
  name: string;
  email: string;
}

type A = Pick<User, "id" | "name">; // { id, name }
type B = Omit<User, "email">; // { id, name }
// Same result, different approach
```

**When to use**:

- Pick: Few properties to include
- Omit: Few properties to exclude

</details>

### Q2: How does `infer` work?

This separates developers who memorize syntax from those who understand the type system. `infer` is advanced - it's how you build types that extract and manipulate other types. Shows you can work with conditional types and understand pattern matching at the type level.

<details>
<summary>Answer</summary>

`infer` declares a type variable within a conditional type.

```typescript
// Extract return type
type ReturnType<T> = T extends (...args: any[]) => infer R // Declare R
  ? R // Use R
  : never;

// Extract array element
type ElementType<T> = T extends (infer E)[] ? E : never;
```

Only valid in `extends` clause of conditional types.

</details>

### Q3: What's distributive conditional type?

Advanced type system knowledge. Tests whether you understand how TypeScript processes unions in conditional types. Most developers don't know this explicitly but run into the behavior. Understanding distribution helps you predict type behavior and avoid surprises when working with unions.

<details>
<summary>Answer</summary>

When conditional type is applied to union, it distributes over each member.

```typescript
type ToArray<T> = T extends any ? T[] : never;

type Result = ToArray<string | number>;
// = ToArray<string> | ToArray<number>
// = string[] | number[]

// Non-distributive (using tuple):
type ToArrayNonDist<T> = [T] extends [any] ? T[] : never;

type Result2 = ToArrayNonDist<string | number>;
// = (string | number)[]
```

</details>

### Q4: How to make all properties of nested objects optional?

This tests recursion understanding - can you write a type that calls itself? Also shows whether you've dealt with complex object shapes in real code. The naive answer is wrong (only makes top level optional). The correct answer handles arbitrary nesting.

<details>
<summary>Answer</summary>

```typescript
type DeepPartial<T> = {
  [K in keyof T]?: T[K] extends object ? DeepPartial<T[K]> : T[K];
};

// Better version (handles arrays):
type DeepPartial<T> = T extends object
  ? { [K in keyof T]?: DeepPartial<T[K]> }
  : T;
```

</details>

## Key Takeaways

1. **Utility Types**: Master Partial, Pick, Omit, Record, ReturnType, Parameters
2. **Mapped Types**: Transform properties with `[K in keyof T]`
3. **Modifiers**: Use `?` `-?` `readonly` `-readonly`
4. **Conditional Types**: `T extends U ? X : Y`
5. **Infer**: Extract types with `infer` in conditional types
6. **Template Literals**: Build string types, combine with unions
7. **Key Remapping**: Filter/transform keys with `as`

## Next Steps

In [Lesson 05: Generics Deep Dive](lesson-05-generics-deep-dive.md), you'll learn:

- Generic constraints
- Default type parameters
- Variance in generics
- Generic performance considerations
