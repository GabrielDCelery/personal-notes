# Lesson 03: Type System Internals

Understanding how TypeScript's type system works under the hood - inference, narrowing, structural typing, and variance.

## Type Inference

You could annotate every variable: `const x: number = 42`. But that's verbose and defeats the point of TypeScript. The compiler is smart enough to figure out types from your code. The question is: when does it infer, what does it infer, and when do you need to step in?

Understanding inference isn't just about writing less code - it's about knowing when TypeScript's guess matches your intent. Sometimes it infers `string` when you need `"success" | "error"`. Sometimes it widens when you need precision. Knowing the rules means you can work with the type system instead of fighting it.

### Basic Inference

```typescript
let x = 42; // Type: number (inferred)
let y = "hello"; // Type: string
let z = [1, 2, 3]; // Type: number[]
let w = { a: 1 }; // Type: { a: number }

// Widening
let a = null; // Type: any (!)
let b = undefined; // Type: any (!)
```

### Best Common Type

When you have an array with different types, what should TypeScript infer? If you have `[new Dog(), new Cat()]`, both extend `Animal`, so you might expect `Animal[]`. But TypeScript doesn't assume parent types - it uses the union of what it sees: `(Dog | Cat)[]`. This is more precise but sometimes surprising. Understanding this helps you know when to explicitly type and when to let inference work.

```typescript
let arr = [1, 2, "three"]; // Type: (number | string)[]

let mixed = [1, null]; // Type: (number | null)[]

class Animal {}
class Dog extends Animal {}
class Cat extends Animal {}
let animals = [new Dog(), new Cat()]; // Type: (Dog | Cat)[]
// Not Animal[] - TypeScript uses concrete types
```

### Contextual Typing

Not all inference flows left-to-right from values. Sometimes TypeScript infers from WHERE you use a value. In `addEventListener('click', (e) => ...)`, you didn't type `e`, but TypeScript knows the callback signature from the `addEventListener` overload. This "backwards" inference from context is powerful - it means less manual typing and more automatic correctness.

```typescript
// Type inferred from context
window.addEventListener("click", (e) => {
  // e is inferred as MouseEvent
  console.log(e.clientX); // ✓ Knows about MouseEvent properties
});

// Array methods
const numbers = [1, 2, 3];
numbers.map((n) => n.toFixed(2)); // n is inferred as number

// Return type inferred
function double(x: number) {
  return x * 2; // Return type: number (inferred)
}
```

## Type Widening

You write `let status = 'success'` and TypeScript infers `string`, not the literal `"success"`. Why? Because `let` means reassignable - you might later do `status = 'error'`. So TypeScript widens the literal to the general type to allow flexibility.

This is usually helpful, but sometimes you WANT the literal type - like when defining a discriminated union or a constant configuration. That's where `const` and `as const` come in. Understanding widening helps you control when TypeScript is precise versus flexible.

```typescript
// Widening
let x = "hello"; // Type: string (not "hello")
let y = 42; // Type: number (not 42)
let z = true; // Type: boolean (not true)

// Prevent widening with const
const a = "hello"; // Type: "hello" (literal)
const b = 42; // Type: 42 (literal)

// Object widening
let obj = { x: 1 }; // Type: { x: number }
const obj2 = { x: 1 }; // Still: { x: number }
// Note: const doesn't prevent widening for objects!

// Prevent with as const
const obj3 = { x: 1 } as const; // Type: { readonly x: 1 }
```

### Interview Question: const vs as const

Here's a common gotcha: `const obj = { x: 1 }` doesn't give you literal types for the properties. `const` prevents reassigning the BINDING (can't do `obj = {...}`), but the object itself is mutable. So `x` stays `number`, not `1`. For deep immutability and literal types, you need `as const`. This comes up constantly in config objects and discriminated unions.

```typescript
const config = {
  apiUrl: "https://api.example.com",
  timeout: 5000,
};
// Type: { apiUrl: string; timeout: number }

const config2 = {
  apiUrl: "https://api.example.com",
  timeout: 5000,
} as const;
// Type: { readonly apiUrl: "https://api.example.com"; readonly timeout: 5000 }

// Usage difference:
config.timeout = 10000; // ✓ Allowed
config2.timeout = 10000; // ❌ Error: readonly
```

## Type Narrowing

You have a value that's `string | number`. Inside an `if (typeof value === 'string')` block, TypeScript knows it's a string. Outside, it's still the union. This isn't magic - it's control flow analysis. TypeScript tracks your runtime checks and refines types based on what's possible in each code path.

Narrowing is how you work with union types safely. Without it, every union would require type assertions. With it, your runtime checks double as compile-time proofs. Master narrowing techniques and you'll write cleaner, safer code with fewer `as` casts.

### typeof Guards

```typescript
function process(value: string | number) {
  if (typeof value === "string") {
    return value.toUpperCase(); // Type: string
  } else {
    return value.toFixed(2); // Type: number
  }
}
```

### instanceof Guards

```typescript
class Dog {
  bark() {}
}
class Cat {
  meow() {}
}

function speak(animal: Dog | Cat) {
  if (animal instanceof Dog) {
    animal.bark(); // Type: Dog
  } else {
    animal.meow(); // Type: Cat
  }
}
```

### in Operator

```typescript
interface Fish {
  swim(): void;
}
interface Bird {
  fly(): void;
}

function move(animal: Fish | Bird) {
  if ("swim" in animal) {
    animal.swim(); // Type: Fish
  } else {
    animal.fly(); // Type: Bird
  }
}
```

### Discriminated Unions (Tagged Unions)

The gold standard for type-safe unions. Add a literal `kind` or `type` field that's unique per variant, and TypeScript can narrow perfectly. Check `if (result.kind === 'success')` and TypeScript knows it's the Success variant with the `data` field. No type casts needed. This pattern is everywhere in production TypeScript - API responses, state machines, Redux actions. Learn this well.

```typescript
interface Success {
  kind: "success";
  data: string;
}
interface Error {
  kind: "error";
  message: string;
}
type Result = Success | Error;

function handle(result: Result) {
  if (result.kind === "success") {
    console.log(result.data); // Type: Success
  } else {
    console.log(result.message); // Type: Error
  }
}
```

### Truthiness Narrowing

Using `if (str)` to check for `string | null` works, but has a gotcha: empty string `''` is falsy. So `if (str)` excludes both `null` AND empty strings. Sometimes that's what you want. Often it's not. Be explicit with `!== null` when you need to allow falsy values. This bug ships to production all the time - user enters empty string, your code treats it as missing data.

```typescript
function printLength(str: string | null) {
  if (str) {
    console.log(str.length); // Type: string
  } else {
    console.log("No string"); // Type: null
  }
}

// Caveat: Excludes falsy values
function printLength2(str: string | null) {
  if (str) {
    console.log(str.length); // Type: string
  }
  // But empty string '' is also falsy!
}

// Better:
function printLength3(str: string | null) {
  if (str !== null) {
    console.log(str.length); // Type: string (includes '')
  }
}
```

### Custom Type Guards

Built-in narrowing (typeof, instanceof) only goes so far. What about validating an API response or checking a complex shape? Type guards with the `is` keyword let you write custom narrowing logic. The function returns a boolean, but TypeScript trusts the `obj is User` annotation and narrows accordingly. It's a contract: you promise the check is correct, TypeScript uses it for narrowing.

```typescript
interface User {
  name: string;
  email: string;
}

// Type predicate
function isUser(obj: any): obj is User {
  return (
    typeof obj === "object" &&
    obj !== null &&
    typeof obj.name === "string" &&
    typeof obj.email === "string"
  );
}

function greet(data: unknown) {
  if (isUser(data)) {
    console.log(data.name); // Type: User
  }
}
```

### Assertion Functions

Type guards narrow inside an if-block. Assertion functions narrow AFTER the call - because if the check fails, they throw. Write `assertIsString(value)` and if execution continues, TypeScript knows `value` is a string. No if-block needed. This is cleaner for invariants: "this should never be null, throw if it is." The `asserts` keyword tells TypeScript about the throw behavior.

```typescript
function assert(condition: any, msg?: string): asserts condition {
  if (!condition) {
    throw new Error(msg);
  }
}

function assertIsString(val: any): asserts val is string {
  if (typeof val !== "string") {
    throw new Error("Not a string");
  }
}

function process(value: string | number) {
  assertIsString(value);
  // Type is now: string (TypeScript knows we threw if not)
  value.toUpperCase();
}
```

## Structural Typing (Duck Typing)

Coming from Java or C#, this feels weird. Two interfaces with identical properties are interchangeable in TypeScript, even if they have different names. It's not about what you CALL a type, it's about what SHAPE it has. "If it walks like a duck and quacks like a duck..."

This matches JavaScript's runtime behavior - objects are just bags of properties. TypeScript mirrors this at compile time. It enables flexibility but can allow unintended compatibility. Understanding structural typing explains why seemingly unrelated types are sometimes assignable.

### Structural vs Nominal

```typescript
// Structural: Compatibility based on shape
interface Point2D {
  x: number;
  y: number;
}

interface Vector {
  x: number;
  y: number;
}

const point: Point2D = { x: 0, y: 0 };
const vec: Vector = point; // ✓ Compatible (same structure)

// Nominal typing (e.g., Java, C#):
// Point2D and Vector would be incompatible despite same shape
```

### Excess Property Checking

Structural typing says "extra properties are fine - you have everything needed." But that would miss typos. Write `{ urll: 'typo' }` for a `{ url: string }` type and structural typing would allow it. The typo'd property is ignored, but your bug persists.

Excess property checking is TypeScript's compromise: for object LITERALS specifically, reject unknown properties. This catches typos at the call site. But assign to a variable first? That bypasses the check because it's no longer a literal. Quirky, but prevents common mistakes.

```typescript
interface Config {
  url: string;
  timeout?: number;
}

// ✓ Allowed (structural)
const config1: Config = {
  url: "https://api.com",
  timeout: 5000,
  retries: 3, // ❌ Error: Object literal may only specify known properties
};

// ✓ Workaround: Assign to variable first
const temp = {
  url: "https://api.com",
  timeout: 5000,
  retries: 3,
};
const config2: Config = temp; // ✓ No error (excess property check bypassed)

// ✓ Type assertion
const config3: Config = {
  url: "https://api.com",
  retries: 3,
} as Config;
```

**Why?** Excess property checking only applies to object literals to catch typos.

### Subtyping

```typescript
interface Animal {
  name: string;
}

interface Dog extends Animal {
  breed: string;
}

const dog: Dog = { name: "Buddy", breed: "Golden" };
const animal: Animal = dog; // ✓ Dog is subtype of Animal (more properties OK)

// Reverse doesn't work:
const animal2: Animal = { name: "Unknown" };
const dog2: Dog = animal2; // ❌ Error: missing 'breed'
```

## Variance

This is where type theory meets practice. If `Dog extends Animal`, is `Array<Dog>` assignable to `Array<Animal>`? What about `Handler<Animal>` to `Handler<Dog>`? The answer depends on whether the type parameter is in a read position (covariant), write position (contravariant), or both (invariant).

Most developers never learn the formal terms, but you FEEL the effects when function types don't assign the way you expect. Understanding variance explains those moments and helps you design better APIs. It's also a common interview topic for senior positions.

### Covariance (Most Common)

Reading from an `Array<Dog>` and expecting an `Animal`? Totally safe - every Dog IS an Animal. So `Dog[]` can be used where `Animal[]` is expected. This is covariance: the type relationship preserves direction (Dog → Animal means Dog[] → Animal[]).

Arrays, Promises, and most read-only generics are covariant. It feels natural because reading more specific data (Dog) as less specific (Animal) is always safe.

```typescript
interface Animal {
  name: string;
}
interface Dog extends Animal {
  breed: string;
}

// Arrays are covariant in read position
let animals: Animal[] = [];
let dogs: Dog[] = [{ name: "Buddy", breed: "Golden" }];

animals = dogs; // ✓ Dog[] is subtype of Animal[]

// Reading is safe:
const animal: Animal = animals[0]; // ✓ Dog is Animal

// But writing is unsafe (in reality):
// animals.push({ name: 'Cat' });  // Would break dogs array!
// TypeScript allows this - limitation
```

### Contravariance (Function Parameters)

This is the mind-bender. A function expecting `Dog` can use a handler that accepts `Animal`? Yes! The handler gets Dogs (which are Animals), so it's safe. But the reverse breaks: an `Animal` handler that calls `.breed` would crash on a Cat.

The type relationship REVERSES for function parameters (Dog → Animal means Handler<Animal> → Handler<Dog>). This is contravariance. It's unintuitive until you think through the safety: broader inputs are safer, narrower inputs are risky. Enable `strictFunctionTypes` to enforce this correctly.

```typescript
type Func<T> = (arg: T) => void;

let animalFunc: Func<Animal> = (animal) => {
  console.log(animal.name);
};

let dogFunc: Func<Dog> = (dog) => {
  console.log(dog.breed);
};

// Contravariance: Can assign Animal handler to Dog position
dogFunc = animalFunc; // ✓ With strictFunctionTypes: true

// Why? Every Dog is an Animal, so Animal handler can handle Dogs
// Reverse is unsafe:
animalFunc = dogFunc; // ❌ Error: Cat doesn't have 'breed'
```

### Invariance (Read-Write)

**Mutable properties** are invariant.

```typescript
interface Box<T> {
  value: T;
  set(val: T): void;
  get(): T;
}

let animalBox: Box<Animal>;
let dogBox: Box<Dog>;

animalBox = dogBox; // ❌ Error: Invariant
dogBox = animalBox; // ❌ Error: Invariant

// Why? Both read and write:
// If allowed: animalBox.set({ name: 'Cat' }) would corrupt dogBox
```

### Bivariance (Legacy)

```typescript
// strictFunctionTypes: false (old behavior)
type Handler<T> = (arg: T) => void;

let animalHandler: Handler<Animal>;
let dogHandler: Handler<Dog>;

// Both allowed (bivariant - unsafe!)
animalHandler = dogHandler; // ⚠️ Allowed
dogHandler = animalHandler; // ⚠️ Allowed

// Always use strictFunctionTypes: true
```

## Type Compatibility

### Function Compatibility

```typescript
type Func1 = (a: number, b: number) => number;
type Func2 = (a: number) => number;

let f1: Func1 = (a, b) => a + b;
let f2: Func2 = (a) => a * 2;

// Can assign fewer parameters
f1 = f2; // ✓ Allowed (extra params ignored)
f2 = f1; // ❌ Error (missing required param)

// Array.forEach example:
[1, 2, 3].forEach((n) => console.log(n)); // ✓ Ignores index, array params
```

### Return Type Compatibility

```typescript
type GetAnimal = () => Animal;
type GetDog = () => Dog;

let getAnimal: GetAnimal;
let getDog: GetDog;

// Return type is covariant
getAnimal = getDog; // ✓ Returning Dog is safe (Dog is Animal)
getDog = getAnimal; // ❌ Error: Returning Animal might not be Dog
```

## Hands-On Exercise 1: Narrowing Challenge

Fix type errors using narrowing:

```typescript
function process(value: string | number | null) {
  // Goal: Call .toUpperCase() if string, .toFixed() if number
  // Skip if null
}
```

<details>
<summary>Solution</summary>

```typescript
function process(value: string | number | null) {
  if (value === null) {
    return;
  }

  if (typeof value === "string") {
    console.log(value.toUpperCase());
  } else {
    console.log(value.toFixed(2));
  }
}
```

</details>

## Hands-On Exercise 2: Type Guard

Write a type guard for this shape:

```typescript
interface ApiSuccess {
  status: "success";
  data: any;
}
interface ApiError {
  status: "error";
  error: string;
}
type ApiResponse = ApiSuccess | ApiError;

// Write: function isSuccess(response: ApiResponse): response is ApiSuccess
```

<details>
<summary>Solution</summary>

```typescript
function isSuccess(response: ApiResponse): response is ApiSuccess {
  return response.status === "success";
}

// Usage:
function handle(response: ApiResponse) {
  if (isSuccess(response)) {
    console.log(response.data); // Type: ApiSuccess
  } else {
    console.log(response.error); // Type: ApiError
  }
}
```

</details>

## Hands-On Exercise 3: Variance

Explain why this works or doesn't:

```typescript
interface Animal {
  name: string;
}
interface Dog extends Animal {
  breed: string;
}

type AnimalFunc = (animal: Animal) => void;
type DogFunc = (dog: Dog) => void;

let f1: AnimalFunc = (a) => console.log(a.name);
let f2: DogFunc = (d) => console.log(d.breed);

// Will these work? Why?
f1 = f2; // ?
f2 = f1; // ?
```

<details>
<summary>Solution</summary>

```typescript
f1 = f2; // ❌ Error
// AnimalFunc might receive Cat, but f2 expects Dog (needs .breed)

f2 = f1; // ✓ Allowed (contravariance)
// DogFunc only receives Dogs, and f1 can handle any Animal
// Since Dog extends Animal, f1 is safe to use

// Parameters are contravariant!
```

</details>

## Interview Questions

### Q1: What's type widening?

This tests whether you understand TypeScript's inference rules versus just "types appear automatically." Knowing widening shows you've debugged issues where literals were needed but weren't inferred, and understand the difference between `let` and `const` at the type level.

<details>
<summary>Answer</summary>

TypeScript converts literal types to general types for mutability.

```typescript
let x = "hello"; // Type: string (widened from "hello")
const y = "hello"; // Type: "hello" (literal)

// Reason: let can be reassigned
x = "world"; // Must allow any string

// Prevent with as const:
let z = "hello" as const; // Type: "hello"
z = "world"; // ❌ Error
```

</details>

### Q2: Explain structural vs nominal typing

Fundamental to understanding TypeScript's type system. Coming from nominal languages (Java, C#), developers are surprised when unrelated types are compatible. This question reveals whether you understand WHY TypeScript chose structural typing (JavaScript compatibility) and can explain the tradeoffs. Senior-level question.

<details>
<summary>Answer</summary>

**Structural** (TypeScript): Compatibility based on shape/structure.

```typescript
interface A {
  x: number;
}
interface B {
  x: number;
}

const a: A = { x: 1 };
const b: B = a; // ✓ Same structure
```

**Nominal** (Java, C#, Flow): Compatibility based on explicit declarations.

```java
class A { int x; }
class B { int x; }

A a = new A();
B b = a;  // ❌ Error: Different types despite same structure
```

TypeScript chose structural for JavaScript's duck-typed nature.

</details>

### Q3: How do discriminated unions work?

This is THE pattern for type-safe unions in production TypeScript. If you don't know discriminated unions, you probably haven't worked on a real TypeScript codebase. Shows you understand practical type narrowing, not just theoretical knowledge. Often followed up with "show me an example from your code."

<details>
<summary>Answer</summary>

Use a common literal property to discriminate union members.

```typescript
type Shape =
  | { kind: "circle"; radius: number }
  | { kind: "square"; size: number };

function area(shape: Shape) {
  if (shape.kind === "circle") {
    return Math.PI * shape.radius ** 2; // Type: circle
  } else {
    return shape.size ** 2; // Type: square
  }
}
```

TypeScript narrows based on the discriminant (`kind`).

</details>

### Q4: What's the difference between type predicates and assertion functions?

These are both custom narrowing techniques but with different control flow. Knowing both shows you've done non-trivial type validation. The interviewer wants to see if you understand HOW narrowing works (control flow analysis) versus just memorizing syntax. Good follow-up: "When would you use each?"

<details>
<summary>Answer</summary>

**Type Predicate**: Returns boolean, narrows type in if-block.

```typescript
function isString(val: any): val is string {
  return typeof val === "string";
}

if (isString(x)) {
  x.toUpperCase(); // Type: string
}
```

**Assertion Function**: Throws on false, narrows type after call.

```typescript
function assertString(val: any): asserts val is string {
  if (typeof val !== "string") throw new Error();
}

assertString(x);
x.toUpperCase(); // Type: string (if we reach here)
```

</details>

### Q5: Explain function parameter contravariance

This is the variance question that actually comes up in practice - functions. If you can explain WHY handler types work this way, you deeply understand TypeScript's type system. Many senior developers can't answer this. It also tests `strictFunctionTypes` knowledge, which many teams don't enable (and should).

<details>
<summary>Answer</summary>

Function parameters are contravariant: can assign handler of supertype to position expecting subtype.

```typescript
type Handler<T> = (arg: T) => void;

let animalHandler: Handler<Animal> = (a) => console.log(a.name);
let dogHandler: Handler<Dog>;

dogHandler = animalHandler; // ✓ Contravariance
```

**Why safe?**

- `dogHandler` receives only Dogs
- `animalHandler` can handle any Animal
- Dog is an Animal, so it works

**Why reverse is unsafe?**

```typescript
animalHandler = dogHandler; // ❌ With strictFunctionTypes
// animalHandler might receive Cat, but dogHandler expects Dog
```

</details>

## Key Takeaways

1. **Inference**: TypeScript infers types; explicit types for clarity
2. **Widening**: Literals widen to general types; prevent with `as const`
3. **Narrowing**: Use typeof, instanceof, in, discriminated unions, type guards
4. **Structural Typing**: Compatibility by shape, not name
5. **Variance**:
   - Covariance: Return types (Dog → Animal)
   - Contravariance: Parameters (Animal → Dog)
   - Invariance: Mutable (neither direction)
6. **Type Guards**: `is` for predicates, `asserts` for assertions
7. **Discriminated Unions**: Use literal discriminant for safe narrowing

## Next Steps

In [Lesson 04: Utility Types & Type Manipulation](lesson-04-utility-types-and-manipulation.md), you'll learn:

- Built-in utility types (Partial, Pick, Record, etc.)
- Mapped types
- Conditional types
- Template literal types
