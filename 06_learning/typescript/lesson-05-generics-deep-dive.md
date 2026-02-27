# Lesson 05: Generics Deep Dive

Master generic types - constraints, defaults, variance, and performance considerations.

## Generic Basics Refresher

Without generics, you'd write `identity(value: any): any` and lose all type safety, or write separate `identityString`, `identityNumber`, `identityUser` functions for every type. Generics let you write one function that works with any type while preserving type information. It's abstraction without sacrifice.

This isn't just about code reuse - generics power TypeScript's most important APIs. Array methods, Promise handling, React component props - all rely on generics. Understanding them deeply means understanding how to build reusable, type-safe abstractions.

```typescript
// Generic function
function identity<T>(value: T): T {
  return value;
}

const num = identity(42); // T = number
const str = identity("hello"); // T = string

// Generic interface
interface Box<T> {
  value: T;
}

const stringBox: Box<string> = { value: "hello" };
const numberBox: Box<number> = { value: 42 };

// Generic class
class Container<T> {
  constructor(private value: T) {}

  getValue(): T {
    return this.value;
  }
}
```

## Generic Constraints

Unconstrained generics (`<T>`) accept anything, which means you can't safely do ANYTHING with T. Can't call methods, access properties, or make assumptions. You need a middle ground between "any type" and "specific type" - that's constraints. They say "T can be anything, as long as it has these properties."

Constraints turn generics from abstract placeholders into useful, type-safe tools. You get the flexibility of generics with the safety of knowing what operations are valid.

Restrict what types can be used as type parameters.

### extends Constraint

The classic mistake: write a generic function that tries to use a property that might not exist. TypeScript correctly errors because unconstrained T could be ANYTHING - numbers, booleans, null. The `extends` constraint solves this: "T can be any type, but it must have a length property."

```typescript
// Without constraint
function getLength<T>(arg: T): number {
  return arg.length; // ❌ Error: Property 'length' does not exist on type 'T'
}

// With constraint
function getLength<T extends { length: number }>(arg: T): number {
  return arg.length; // ✓ TypeScript knows T has length
}

getLength("hello"); // ✓ string has length
getLength([1, 2, 3]); // ✓ array has length
getLength({ length: 5 }); // ✓ matches constraint
getLength(42); // ❌ Error: number has no length
```

### Multiple Constraints

Sometimes one constraint isn't enough. Your function needs objects with BOTH an id and a name. Instead of creating a combined interface just for the constraint, use intersection (`&`) to compose multiple constraints. This keeps your type definitions minimal and focused.

```typescript
interface HasId {
  id: number;
}

interface HasName {
  name: string;
}

// T must satisfy both constraints
function process<T extends HasId & HasName>(item: T): void {
  console.log(item.id, item.name);
}

process({ id: 1, name: "Alice" }); // ✓
process({ id: 1, name: "Bob", age: 30 }); // ✓ Extra properties OK
process({ id: 1 }); // ❌ Missing name
```

### keyof Constraint

The type-safe property accessor pattern. You want to get a property from an object, but `obj[key]` with a string key loses type safety - key might not exist, return type is unknown. `K extends keyof T` constrains the key to ONLY valid property names, and TypeScript knows the exact return type.

This pattern is everywhere in type-safe libraries - lodash's `get`, React's prop access, ORM query builders. Master this and you'll write APIs that catch typos at compile time.

```typescript
// Get property value by key
function getProperty<T, K extends keyof T>(obj: T, key: K): T[K] {
  return obj[key];
}

const person = { name: "Alice", age: 30 };

const name = getProperty(person, "name"); // Type: string
const age = getProperty(person, "age"); // Type: number
getProperty(person, "email"); // ❌ Error: 'email' not in person
```

### Constraint Inference

Sometimes you don't need explicit constraints - TypeScript infers them from how you USE the generic. The spread operator `{...obj1, ...obj2}` works with any object, so TypeScript infers T and U extend object. You get the constraint benefits without writing the `extends object` explicitly.

```typescript
// TypeScript infers the constraint
function merge<T, U>(obj1: T, obj2: U): T & U {
  return { ...obj1, ...obj2 };
}

const merged = merge({ a: 1 }, { b: 2 });
// Type: { a: number } & { b: number }
// = { a: number; b: number }

console.log(merged.a, merged.b); // ✓ Both properties available
```

## Default Type Parameters

Not every generic needs to be specified. API responses are usually `unknown` data. React component state defaults to empty object. These are sensible defaults that users rarely override. Default type parameters make the common case convenient while keeping the door open for customization.

It's the difference between `ApiResponse<User>` when you know the type and just `ApiResponse` when you don't yet. Same type, appropriate defaults for each situation.

Provide fallback types when not specified.

```typescript
interface ApiResponse<T = unknown> {
  data: T;
  status: number;
}

// With explicit type
const response1: ApiResponse<User> = {
  data: { id: 1, name: "Alice" },
  status: 200,
};

// Uses default (unknown)
const response2: ApiResponse = {
  data: { anything: true },
  status: 200,
};
// Type: ApiResponse<unknown>
```

### Cascading Defaults

Defaults can depend on previous type parameters. If you specify T, E defaults to Error. Specify both for full control. This creates a progressive disclosure of complexity - simple case is simple, complex case is possible.

```typescript
type Result<T = string, E = Error> =
  | { success: true; value: T }
  | { success: false; error: E };

type StringResult = Result; // Result<string, Error>
type NumberResult = Result<number>; // Result<number, Error>
type CustomResult = Result<User, ApiError>; // Result<User, ApiError>
```

### Conditional Defaults

Defaults can be conditional on the first parameter. If T is a string, default U to string array. If T is a number, default to number array. The default adapts based on what you provide. This is advanced but powerful for creating intuitive APIs.

```typescript
type Container<T, U = T extends string ? string[] : number[]> = {
  value: T;
  items: U;
};

type StrContainer = Container<string>;
// { value: string; items: string[] }

type NumContainer = Container<number>;
// { value: number; items: number[] }

type CustomContainer = Container<string, boolean[]>;
// { value: string; items: boolean[] }
```

## Generic Type Inference

You rarely write `identity<number>(42)` - you just write `identity(42)` and TypeScript figures out T = number. This inference is what makes generics feel natural instead of verbose. Understanding HOW inference works helps you design APIs that infer correctly and know when to provide explicit types.

TypeScript infers generics from usage.

### Function Argument Inference

TypeScript looks at your arguments and works backwards to figure out the type parameters. Pass an array of numbers and a function that returns strings? T = number, U = string. The whole generic machinery becomes invisible - it just works.

```typescript
function map<T, U>(arr: T[], fn: (item: T) => U): U[] {
  return arr.map(fn);
}

// TypeScript infers T = number, U = string
const result = map([1, 2, 3], (n) => n.toString());
// Type: string[]
```

### Multiple Inference Sources

When multiple arguments inform the same type parameter, TypeScript finds the "best common type" - the type that satisfies all uses. Pass a number and a string to a function expecting the same type? T becomes `string | number`, the union that accommodates both.

```typescript
function combine<T>(a: T, b: T): T[] {
  return [a, b];
}

combine(1, 2); // T = number
combine("a", "b"); // T = string
combine(1, "b"); // T = string | number (best common type)
```

### Inference Priority

Critical gotcha: TypeScript infers from ARGUMENT POSITION, not usage. Your defaults parameter is typed `Partial<T>`, but T is inferred from value (the first argument). The second argument doesn't influence inference - it just has to be compatible with the inferred T. This surprises many developers.

```typescript
function create<T>(value: T, defaults?: Partial<T>): T {
  return { ...defaults, ...value } as T;
}

// T inferred from first argument
const user = create({ name: "Alice" }, { age: 30 });
// Type: { name: string }
// ⚠️ defaults type is ignored for inference
```

## Advanced Generic Patterns

Generics in isolation are useful. Generic PATTERNS solve real problems. These are the blueprints for type-safe factories, builders, and utilities that appear in production codebases everywhere.

### Generic Factory

You want a function that instantiates classes. But `new ClassName()` hardcodes the class. A generic factory accepts ANY class constructor and returns an instance of that class. The `Constructor<T>` pattern captures "something that news up a T." This is how dependency injection, testing frameworks, and plugin systems work.

```typescript
interface Constructor<T> {
  new (...args: any[]): T;
}

function createInstance<T>(ctor: Constructor<T>, ...args: any[]): T {
  return new ctor(...args);
}

class User {
  constructor(public name: string) {}
}

const user = createInstance(User, "Alice");
// Type: User
```

### Generic Builder

The builder pattern with type safety. Each `where` call returns `this`, enabling chaining while preserving the generic type. The filters array holds predicates that all operate on the same type T. This ensures you can't accidentally mix query conditions for different types.

```typescript
class QueryBuilder<T> {
  private filters: Array<(item: T) => boolean> = [];

  where(predicate: (item: T) => boolean): this {
    this.filters.push(predicate);
    return this; // Return 'this' for chaining
  }

  execute(data: T[]): T[] {
    return data.filter((item) => this.filters.every((fn) => fn(item)));
  }
}

interface User {
  name: string;
  age: number;
}

const users: User[] = [
  { name: "Alice", age: 30 },
  { name: "Bob", age: 25 },
];

const result = new QueryBuilder<User>()
  .where((u) => u.age > 20)
  .where((u) => u.name.startsWith("A"))
  .execute(users);
// [{ name: 'Alice', age: 30 }]
```

### Generic Constraints with Conditional Types

Advanced type manipulation: extract only the keys whose values match a specific type. `KeysOfType<User, string>` gives you only the string-valued keys. Then use THAT as a constraint for the key parameter. The result? A function that only accepts keys you KNOW are strings, with no runtime checking needed.

```typescript
// Extract keys of specific type
type KeysOfType<T, V> = {
  [K in keyof T]: T[K] extends V ? K : never;
}[keyof T];

interface User {
  id: number;
  name: string;
  age: number;
  active: boolean;
}

type StringKeys = KeysOfType<User, string>; // 'name'
type NumberKeys = KeysOfType<User, number>; // 'id' | 'age'

// Use in generic function
function getStringValue<T, K extends KeysOfType<T, string>>(
  obj: T,
  key: K,
): string {
  return obj[key] as string;
}

const user: User = { id: 1, name: "Alice", age: 30, active: true };
getStringValue(user, "name"); // ✓
getStringValue(user, "age"); // ❌ Error: 'age' is not string key
```

## Variance in Generics

We covered variance in the type system lesson, but it's particularly important for generics. When you have `Container<Dog>`, is it assignable to `Container<Animal>`? The answer depends on how Container uses its type parameter - read-only (covariant), write-only (contravariant), or both (invariant).

Getting variance wrong leads to unsound types - code that compiles but breaks at runtime. Understanding variance helps you design generic types that are both flexible and safe.

How generic types relate to each other based on their type parameters.

### Covariance (Arrays, Promises)

Arrays let you READ elements. If you have `Dog[]` and need `Animal[]`, it's safe - every element you read is a Dog, which IS an Animal. Promises are the same: if you await a `Promise<Dog>`, treating it as `Promise<Animal>` is safe because you're just reading the value.

This is covariance: the container relationship follows the type relationship (Dog → Animal means Dog[] → Animal[]).

```typescript
interface Animal {
  name: string;
}

interface Dog extends Animal {
  breed: string;
}

// Arrays are covariant
let animals: Animal[] = [];
let dogs: Dog[] = [];

animals = dogs; // ✓ Dog[] is subtype of Animal[]

// Promises are covariant
let animalPromise: Promise<Animal>;
let dogPromise: Promise<Dog>;

animalPromise = dogPromise; // ✓ Promise<Dog> is subtype of Promise<Animal>
```

### Contravariance (Function Parameters)

Functions WRITE to their parameters (conceptually - they receive and consume them). If you have a function expecting `Dog`, can you use a function expecting `Animal`? YES - the Animal handler can handle ANY animal, including Dogs. But the reverse fails - a Dog handler expects `.breed`, which Cats don't have.

This is contravariance: the parameter relationship REVERSES (Dog → Animal means Handler<Animal> → Handler<Dog>).

```typescript
type Handler<T> = (arg: T) => void;

let animalHandler: Handler<Animal>;
let dogHandler: Handler<Dog>;

// Contravariant (with strictFunctionTypes: true)
dogHandler = animalHandler; // ✓ Can use Animal handler for Dogs
animalHandler = dogHandler; // ❌ Can't use Dog handler for Animals
```

### Invariance (Mutable Containers)

When a container both reads AND writes, neither direction is safe. You can't assign `Cage<Dog>` to `Cage<Animal>` (writing a Cat would break it) or vice versa (reading might give you a Cat when you expect a Dog). The presence of both operations creates invariance - no assignment in either direction.

This is why mutable generic containers are tricky. Read-only containers can be covariant. Write-only can be contravariant. But read-write? Invariant.

```typescript
interface Cage<T> {
  get(): T;
  set(value: T): void;
}

let animalCage: Cage<Animal>;
let dogCage: Cage<Dog>;

animalCage = dogCage; // ❌ Invariant
dogCage = animalCage; // ❌ Invariant

// Why? Both get (covariant) and set (contravariant)
```

## Generic Performance

TypeScript's type checker isn't infinite. Deeply recursive generics can hit instantiation limits. Overly complex types slow down your editor. Most developers never hit these limits, but when you do, understanding what causes slowdown helps you optimize.

These aren't premature optimization concerns - they're "your build takes 5 minutes and your editor freezes" concerns. Know the limits.

### Instantiation Depth Limit

TypeScript stops recursing after 50 levels by default. Try to create infinitely nested types and you'll hit this limit. The fix? Explicitly limit your recursion. For types like DeepPartial, you rarely need more than 5-10 levels anyway. Set a reasonable limit and you'll stay within TypeScript's capabilities.

TypeScript has a limit on type instantiation depth (default 50).

```typescript
// Problematic: Deep recursion
type DeepNested<T, N extends number = 0> = N extends 50
  ? T
  : { value: DeepNested<T /* increment N */> };

// Better: Limit recursion explicitly
type SafeNested<T, Depth extends number = 5> = Depth extends 0
  ? T
  : { value: SafeNested<T /* decrement */> };
```

### Type Instantiation Limit

Some types generate MASSIVE numbers of instantiations. Distributive conditional types over large unions, deeply mapped recursive types - these multiply quickly. The fix is usually avoiding distribution or limiting recursion. Profile with `tsc --extendedDiagnostics` to find the culprits.

```typescript
// Bad: Creates many instantiations
type BadUnion<T> = T extends any ? { [K in keyof T]: BadUnion<T[K]> } : never;

// Better: Avoid unnecessary instantiations
type GoodUnion<T> = T extends object ? { [K in keyof T]: T[K] } : T;
```

### Caching with Type Aliases

TypeScript recomputes complex conditional types every time they're referenced. If you use `T extends string ? string[] : number[]` in 10 places, TypeScript computes it 10 times. Extract to a type alias and it computes once. Simple optimization, significant speedup for complex types.

```typescript
// Slow: Recomputes every time
function process<T>(value: T extends string ? string[] : number[]) {
  // ...
}

// Faster: Cache result
type ProcessType<T> = T extends string ? string[] : number[];

function process<T>(value: ProcessType<T>) {
  // ...
}
```

## Hands-On Exercise 1: Type-Safe Pick Multiple

Create a function that picks multiple properties safely:

```typescript
// Goal:
const user = { id: 1, name: "Alice", age: 30, email: "alice@example.com" };
const picked = pick(user, ["name", "email"]);
// Type: { name: string; email: string }
```

<details>
<summary>Solution</summary>

```typescript
function pick<T, K extends keyof T>(obj: T, keys: K[]): Pick<T, K> {
  const result = {} as Pick<T, K>;

  for (const key of keys) {
    result[key] = obj[key];
  }

  return result;
}

// Usage:
const user = { id: 1, name: "Alice", age: 30, email: "a@example.com" };
const picked = pick(user, ["name", "email"]);
// Type: { name: string; email: string }

pick(user, ["name", "invalid"]); // ❌ Error: 'invalid' not in user
```

</details>

## Hands-On Exercise 2: Generic Cache

Build a type-safe cache with generic keys and values:

```typescript
// Requirements:
// - set<K, V>(key: K, value: V): void
// - get<K>(key: K): V | undefined
// - Type-safe: keys and values must match
```

<details>
<summary>Solution</summary>

```typescript
class TypedCache<Schema extends Record<string, any>> {
  private cache = new Map<keyof Schema, any>();

  set<K extends keyof Schema>(key: K, value: Schema[K]): void {
    this.cache.set(key, value);
  }

  get<K extends keyof Schema>(key: K): Schema[K] | undefined {
    return this.cache.get(key);
  }

  has<K extends keyof Schema>(key: K): boolean {
    return this.cache.has(key);
  }
}

// Usage:
interface CacheSchema {
  user: { id: number; name: string };
  settings: { theme: string };
  count: number;
}

const cache = new TypedCache<CacheSchema>();

cache.set("user", { id: 1, name: "Alice" }); // ✓
cache.set("count", 42); // ✓
cache.set("user", "invalid"); // ❌ Error: wrong type

const user = cache.get("user"); // Type: { id: number; name: string } | undefined
const count = cache.get("count"); // Type: number | undefined
```

</details>

## Hands-On Exercise 3: Async Mapper

Create a generic async map function:

```typescript
// Goal: Map array with async function
async function asyncMap<T, U>(
  arr: T[],
  fn: (item: T) => Promise<U>,
): Promise<U[]> {
  // Implement
}

// Usage:
const ids = [1, 2, 3];
const users = await asyncMap(ids, async (id) => fetchUser(id));
// Type: User[]
```

<details>
<summary>Solution</summary>

```typescript
async function asyncMap<T, U>(
  arr: T[],
  fn: (item: T, index: number) => Promise<U>,
): Promise<U[]> {
  return Promise.all(arr.map(fn));
}

// Alternative: Sequential processing
async function asyncMapSeq<T, U>(
  arr: T[],
  fn: (item: T, index: number) => Promise<U>,
): Promise<U[]> {
  const results: U[] = [];

  for (let i = 0; i < arr.length; i++) {
    results.push(await fn(arr[i], i));
  }

  return results;
}

// Usage:
const ids = [1, 2, 3];
const users = await asyncMap(ids, async (id) => {
  const response = await fetch(`/users/${id}`);
  return response.json();
});
// Type: any[] (would be User[] with proper typing)
```

</details>

## Interview Questions

### Q1: What's the difference between generic constraint and union type?

This tests deep understanding of generics. Many developers think `<T extends string | number>` and `(value: string | number)` are the same. They're not - one preserves the specific type, one collapses to a union. Shows whether you understand how generics flow types through code.

<details>
<summary>Answer</summary>

**Constraint** (`extends`): Restricts possible types, preserves specific type.

```typescript
function log<T extends string | number>(value: T): T {
  console.log(value);
  return value; // Returns exact type (string or number)
}

const x = log("hello"); // Type: "hello" (literal)
const y = log(42); // Type: 42 (literal)
```

**Union**: Single type that's one of several options.

```typescript
function log(value: string | number): string | number {
  console.log(value);
  return value; // Returns union type
}

const x = log("hello"); // Type: string | number
const y = log(42); // Type: string | number
```

**Key difference**: Generic preserves specific type through the function.

</details>

### Q2: When to use default type parameters?

Tests API design judgment. Default parameters make common cases convenient, but bad defaults confuse users or lead to bugs. This question reveals whether you think about developer experience and can balance convenience vs explicitness.

<details>
<summary>Answer</summary>

Use defaults when:

1. Most common case is predictable
2. Making all callsites explicit is tedious
3. Backward compatibility

```typescript
// Good: API response usually unknown
interface ApiResponse<T = unknown> {
  data: T;
  status: number;
}

// Good: State usually empty object
class Component<Props = {}, State = {}> {
  // ...
}

// Bad: No clear default
interface Box<T = any> {
  // 'any' is rarely a good default
  value: T;
}
```

</details>

### Q3: How does TypeScript infer generic types?

This separates those who use generics from those who understand them. Inference is complex - multiple sources, best common type, fallbacks. If you can explain the inference algorithm, you can design APIs that infer correctly and debug when they don't.

<details>
<summary>Answer</summary>

Inference process:

1. From call site arguments
2. Best common type if multiple sources
3. Falls back to constraint or default
4. Last resort: unknown/any

```typescript
function combine<T>(a: T, b: T): T[] {
  return [a, b];
}

combine(1, 2); // T = number
combine("a", 1); // T = string | number (best common)
combine<string>("a", 1); // ❌ Error: 1 not assignable to string

// Can't infer from return type:
const result: string[] = combine(1, 2); // ❌ Still infers T = number
```

</details>

### Q4: Explain variance in generic types

Advanced question that reveals depth of type system knowledge. Variance is subtle and most developers learn it through trial and error. Being able to explain covariance vs contravariance shows you understand the mathematical foundations of type safety, not just TypeScript syntax.

<details>
<summary>Answer</summary>

**Covariant** (read-only): `Dog[]` → `Animal[]`

```typescript
let animals: Animal[] = dogs; // ✓ Reading Dog as Animal is safe
```

**Contravariant** (write-only): `Animal handler` → `Dog handler`

```typescript
let dogHandler: (d: Dog) => void = (a: Animal) => {}; // ✓ Safe
```

**Invariant** (read-write): Neither direction allowed

```typescript
interface Box<T> {
  get(): T; // Covariant position
  set(T): void; // Contravariant position
}
// Result: Invariant (can't assign either direction)
```

</details>

## Key Takeaways

1. **Constraints**: Use `extends` to restrict generic types
2. **keyof**: Create type-safe property accessors
3. **Defaults**: Provide fallbacks for common cases
4. **Inference**: TypeScript infers from arguments, not return types
5. **Variance**:
   - Arrays/Promises: Covariant
   - Function params: Contravariant
   - Mutable: Invariant
6. **Performance**: Limit recursion depth, avoid excessive instantiations
7. **Patterns**: Factory, builder, type-safe caches

## Next Steps

In [Lesson 06: Module Systems](lesson-06-module-systems.md), you'll learn:

- ESM vs CommonJS differences
- Module resolution strategies
- Import/export patterns and best practices
- Dynamic imports and lazy loading
