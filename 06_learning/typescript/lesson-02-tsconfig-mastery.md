# Lesson 02: tsconfig.json Mastery

Deep dive into TypeScript compiler configuration - the options that matter for production code and interviews.

## Critical Compiler Options

### Strict Type Checking

TypeScript without strict mode is JavaScript with autocomplete. It lets `null` slip through as strings, allows `any` to spread like wildfire, and won't catch when you call a function with wrong arguments. These aren't just theoretical issues - they're production bugs waiting to happen.

Strict mode flips TypeScript from "helpful suggestions" to "your code is provably safer." Yes, it makes migration harder. Yes, you'll fight with the compiler initially. But every error it catches at compile time is a bug that never makes it to production. This is why every serious TypeScript codebase enables it.

```json
{
  "compilerOptions": {
    "strict": true // Enables ALL strict options below
  }
}
```

| Option                         | What it does                          | Common pitfall                        |
| ------------------------------ | ------------------------------------- | ------------------------------------- |
| `strict`                       | Enables all strict options            | Turn this on, fix issues individually |
| `noImplicitAny`                | Error on `any` inference              | `function foo(x)` → error             |
| `strictNullChecks`             | null/undefined are distinct types     | `string` ≠ `string \| null`           |
| `strictFunctionTypes`          | Function parameters are contravariant | Catches unsafe callbacks              |
| `strictBindCallApply`          | Type-check bind/call/apply            | `fn.call(this, 'wrong')` → error      |
| `strictPropertyInitialization` | Class properties must be initialized  | Must set in constructor or `!:`       |
| `noImplicitThis`               | Error on implicit `this: any`         | Methods need explicit `this` param    |
| `alwaysStrict`                 | Emit `"use strict"`                   | Good for ES5 targets                  |

### Example: strictNullChecks Impact

The billion-dollar mistake (Tony Hoare's words, not mine) is allowing null to inhabit any type. In most languages, a string can secretly be null, leading to "Cannot read property of undefined" errors that plague JavaScript. `strictNullChecks` fixes this by making TypeScript's type system null-aware: if something can be null, you must explicitly say so and check for it.

```typescript
// strictNullChecks: false
let name: string = null; // ✓ Allowed (dangerous!)
name.toLowerCase(); // Runtime error!

// strictNullChecks: true
let name: string = null; // ❌ Error: Type 'null' is not assignable to type 'string'
let name: string | null = null; // ✓ Correct
if (name !== null) {
  name.toLowerCase(); // ✓ Safe, TypeScript knows name is string here
}
```

## Module System Options

### module & target

You write modern TypeScript with async/await and import statements. But your code needs to run on Node 14, support IE11 browsers, and work with both CommonJS and ESM consumers. How do you write once and run everywhere?

That's where `target` and `module` split the problem: `target` controls JavaScript syntax (can we use arrow functions?), while `module` controls import/export format (CommonJS `require` or ESM `import`?). These are independent choices because syntax and module system are orthogonal concerns. Old browsers need old syntax but modern bundlers want ESM for tree-shaking.

```json
{
  "compilerOptions": {
    "target": "ES2020", // Output JavaScript version
    "module": "NodeNext", // Module system
    "lib": ["ES2020"] // Available APIs
  }
}
```

| Option   | Value                                      | Use Case                   |
| -------- | ------------------------------------------ | -------------------------- |
| `target` | `ES5`, `ES6`, `ES2020`, `ESNext`           | Browser/Node compatibility |
| `module` | `CommonJS`, `ESNext`, `NodeNext`, `Node16` | Module system              |
| `lib`    | `ES2015`, `DOM`, `ES2020`, etc.            | Available global APIs      |

### Module Strategies

```typescript
// module: "CommonJS"
import { foo } from "./bar";
// Becomes:
const bar_1 = require("./bar");

// module: "ESNext" or "ES2020"
import { foo } from "./bar";
// Stays as:
import { foo } from "./bar";

// module: "NodeNext" (recommended for modern Node)
// Respects package.json "type" field
```

### Interview Question: target vs lib

This question separates developers who cargo-cult config from those who understand transpilation. You can write code using `Promise` (ES2015) and compile it to ES5 syntax - as long as a polyfill provides `Promise` at runtime. `lib` tells TypeScript "these APIs will exist" while `target` says "compile to this syntax." They're independent because polyfills provide APIs, not syntax.

```json
{
  "target": "ES5", // Output code will be ES5
  "lib": ["ES2015", "DOM"] // But can USE ES2015 APIs in source
}
```

TypeScript will:

- Compile to ES5 syntax
- Allow ES2015 APIs (Promise, Map, etc.) in your code
- Assume polyfills exist at runtime

> [!WARN]
> lib does not install anything, just tells the transpiler that certain language specific APIs (e.g. Promise.resolve) will be available at runtime but they need to be provided separately

## Module Resolution

### Resolution Strategies

When you write `import { foo } from './bar'`, how does TypeScript find the file? Should it look for `bar.ts`, `bar/index.ts`, or `bar.js`? Does it need the extension in the import? Different environments have different rules, and TypeScript needs to match them or your code won't run.

The resolution strategy tells TypeScript how to interpret imports. `node` uses classic Node.js rules (extensions optional, checks index files). `node16`/`nodenext` uses modern Node ESM rules (extensions required, respects package.json exports). `bundler` is for webpack/Vite where the bundler handles everything. Pick wrong and your editor shows green checkmarks while your runtime throws "module not found" errors.

```json
{
  "moduleResolution": "node" | "node16" | "nodenext" | "bundler"
}
```

| Strategy              | When to Use           | Behavior                        |
| --------------------- | --------------------- | ------------------------------- |
| `node` (legacy)       | Old projects          | Classic Node resolution         |
| `node16` / `nodenext` | Modern Node (16+)     | Respects package.json `exports` |
| `bundler`             | Webpack/Vite projects | Like `node16` but relaxed       |

### Resolution Example

```
src/
  utils/
    index.ts
  app.ts
```

```typescript
// app.ts
import { helper } from "./utils";

// moduleResolution: "node"
// Resolves to: ./utils/index.ts ✓

// moduleResolution: "node16" with type: "module"
// ❌ Error! Must use explicit extension:
import { helper } from "./utils/index.js"; // ✓
// Note: .js extension even though file is .ts!
```

### Critical Gotcha: File Extensions with Node16

This trips up everyone the first time. You're writing TypeScript, the file is `bar.ts`, but you must import it as `bar.js`. It feels wrong. Why would you reference a .js file that doesn't exist?

The answer is that TypeScript isn't running your code - Node is. TypeScript compiles `bar.ts` to `bar.js`, and Node's ESM loader requires explicit extensions. So TypeScript makes you write the import as it will appear in the compiled output. The source file is `.ts`, but the import path is what Node will see: `.js`. Once you understand this, it makes sense. But the first time? Total headache.

```typescript
// tsconfig: { "module": "NodeNext", "moduleResolution": "NodeNext" }
// package.json: { "type": "module" }

// ❌ Wrong
import { foo } from "./bar";
import { foo } from "./bar.ts";

// ✓ Correct
import { foo } from "./bar.js"; // Yes, .js even for .ts files!
```

**Why?** TypeScript emits .js files, and Node resolves the emitted .js files.

## Path Mapping

Deep directory trees lead to import hell: `'../../../../components/Button'`. Move a file and all those relative paths break. Worse, you can't tell at a glance whether you're importing from your code or from node_modules. It's brittle, ugly, and hard to refactor.

Path mapping creates import aliases that look like package names: `@components/Button`, `@utils/helpers`. It's cleaner, refactor-friendly, and immediately obvious what you're importing. But here's the catch: TypeScript only uses these for type-checking. Your compiled JavaScript still has those paths, and Node doesn't know what `@components` means. You need runtime support from your bundler or a tool like `tsconfig-paths`.

```json
{
  "compilerOptions": {
    "baseUrl": "./src",
    "paths": {
      "@utils/*": ["utils/*"],
      "@components/*": ["components/*"],
      "@/*": ["./*"]
    }
  }
}
```

```typescript
// Before
import { Button } from "../../../components/Button";

// After
import { Button } from "@components/Button";
```

### Important: Paths Don't Rewrite Imports!

TypeScript uses `paths` for **type-checking only**. At runtime, you need a bundler or runtime tool to resolve them.

**Solutions**:

- Bundlers (Webpack, Vite): Handle automatically
- Node: Use `tsconfig-paths` package
- Build: Use `tsc-alias` to rewrite after compilation

## Project References (Monorepos)

Imagine a monorepo with 50 packages. Every time you change one file, TypeScript rechecks ALL 50 packages. Your editor crawls. Builds take minutes. You need incremental builds, but TypeScript doesn't know which packages depend on what.

Project references solve this by letting you declare dependencies between packages. TypeScript can then rebuild only what changed and what depends on it. Package A imports Package B? TypeScript knows to rebuild B before A. It builds in parallel where possible, caches what hasn't changed, and your monorepo stays fast even as it grows. This is how companies like Microsoft scale TypeScript to massive codebases.

### Project Structure

```
monorepo/
  packages/
    core/
      tsconfig.json
      src/
    utils/
      tsconfig.json
      src/
  tsconfig.json  (root)
```

### Root tsconfig.json

```json
{
  "files": [],
  "references": [{ "path": "./packages/core" }, { "path": "./packages/utils" }]
}
```

### packages/core/tsconfig.json

```json
{
  "compilerOptions": {
    "composite": true, // Required for project references
    "declaration": true, // Required for project references
    "declarationMap": true, // Helpful for debugging
    "outDir": "./dist",
    "rootDir": "./src"
  },
  "include": ["src/**/*"],
  "references": [
    { "path": "../utils" } // core depends on utils
  ]
}
```

### Building with References

```bash
# Build all projects in correct order
tsc --build

# Build with watch mode
tsc --build --watch

# Clean build
tsc --build --clean
```

**Benefits**:

- Faster incremental builds
- Better editor performance
- Enforces dependency boundaries
- Separate compilation per package

## Essential Compiler Options Reference

### Output Control

```json
{
  "compilerOptions": {
    "outDir": "./dist", // Output directory
    "rootDir": "./src", // Input directory (mirrors structure)
    "declaration": true, // Generate .d.ts files
    "declarationMap": true, // Generate .d.ts.map for IDE navigation
    "sourceMap": true, // Generate .js.map for debugging
    "removeComments": true, // Strip comments from output
    "emitDeclarationOnly": false // Only emit .d.ts (no .js)
  }
}
```

### Interop & Compatibility

```json
{
  "compilerOptions": {
    "esModuleInterop": true, // Enable default imports from CommonJS
    "allowSyntheticDefaultImports": true, // Type-check synthetic defaults
    "isolatedModules": true, // Each file must be self-contained (for bundlers)
    "forceConsistentCasingInFileNames": true, // Prevent case-sensitivity issues
    "skipLibCheck": true // Skip type-checking .d.ts files (faster builds)
  }
}
```

### Critical: esModuleInterop

```typescript
// Without esModuleInterop
import * as express from "express";
const app = express(); // ❌ Error: express is not a function

// With esModuleInterop: true
import express from "express";
const app = express(); // ✓ Works!
```

### Strict Checking (Beyond strict: true)

```json
{
  "compilerOptions": {
    "strict": true, // Base strict checking
    "noUnusedLocals": true, // Error on unused variables
    "noUnusedParameters": true, // Error on unused function params
    "noImplicitReturns": true, // All code paths must return
    "noFallthroughCasesInSwitch": true, // Switch cases must break/return
    "noUncheckedIndexedAccess": true, // array[i] returns T | undefined
    "noPropertyAccessFromIndexSignature": true, // obj.foo vs obj["foo"]
    "allowUnusedLabels": false,
    "allowUnreachableCode": false
  }
}
```

### Example: noImplicitReturns

```typescript
// noImplicitReturns: true
function getUser(id: number): User {
  if (id > 0) {
    return fetchUser(id);
  }
  // ❌ Error: Not all code paths return a value
}

// Fix:
function getUser(id: number): User {
  if (id > 0) {
    return fetchUser(id);
  }
  throw new Error("Invalid ID"); // ✓ All paths covered
}
```

### Example: noUncheckedIndexedAccess

```typescript
// noUncheckedIndexedAccess: false
const arr = [1, 2, 3];
const val = arr[10]; // Type: number (dangerous!)
val.toFixed(); // Runtime error!

// noUncheckedIndexedAccess: true
const arr = [1, 2, 3];
const val = arr[10]; // Type: number | undefined ✓
val.toFixed(); // ❌ Error: possibly undefined
if (val !== undefined) {
  val.toFixed(); // ✓ Safe
}
```

## Include/Exclude Patterns

```json
{
  "include": [
    "src/**/*" // All files in src/
  ],
  "exclude": [
    "node_modules", // Always excluded by default
    "dist",
    "**/*.spec.ts", // Exclude test files
    "**/__tests__/**"
  ]
}
```

**Default exclude**: `node_modules`, `bower_components`, `jspm_packages`, and `outDir`.

## Hands-On Exercise 1: Fix Strict Errors

Given this code, enable `strict: true` and fix all errors:

```typescript
function getUserName(user) {
  if (user.name) {
    return user.name.toUpperCase();
  }
}

class Person {
  name: string;
  age: number;
}

const names: string[] = ["Alice", "Bob"];
console.log(names[5].toLowerCase());
```

<details>
<summary>Solution</summary>

```typescript
// Fix 1: Add types (noImplicitAny)
interface User {
  name?: string;
}

function getUserName(user: User): string | undefined {
  if (user.name) {
    return user.name.toUpperCase();
  }
  return undefined; // noImplicitReturns
}

// Fix 2: Initialize class properties (strictPropertyInitialization)
class Person {
  name: string;
  age: number;

  constructor(name: string, age: number) {
    this.name = name;
    this.age = age;
  }
}

// Or use definite assignment assertion
class Person {
  name!: string; // I promise to initialize this
  age!: number;
}

// Fix 3: Handle undefined (noUncheckedIndexedAccess)
const names: string[] = ["Alice", "Bob"];
const name = names[5];
if (name !== undefined) {
  console.log(name.toLowerCase());
}
// Or
console.log(names[5]?.toLowerCase());
```

</details>

## Hands-On Exercise 2: Configure for Modern Node

Create tsconfig.json for a Node 18+ project with:

- Full strict checking
- ESM modules
- Path aliases
- Source maps for debugging

<details>
<summary>Solution</summary>

```json
{
  "compilerOptions": {
    // Language & Environment
    "target": "ES2022",
    "lib": ["ES2022"],
    "module": "NodeNext",
    "moduleResolution": "NodeNext",

    // Strict Checking
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noImplicitReturns": true,
    "noFallthroughCasesInSwitch": true,
    "noUncheckedIndexedAccess": true,

    // Interop
    "esModuleInterop": true,
    "allowSyntheticDefaultImports": true,
    "isolatedModules": true,
    "forceConsistentCasingInFileNames": true,

    // Output
    "outDir": "./dist",
    "rootDir": "./src",
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true,

    // Path Mapping
    "baseUrl": "./src",
    "paths": {
      "@/*": ["./*"],
      "@utils/*": ["utils/*"]
    },

    // Performance
    "skipLibCheck": true
  },
  "include": ["src/**/*"],
  "exclude": ["node_modules", "dist", "**/*.spec.ts"]
}
```

**package.json**:

```json
{
  "type": "module",
  "scripts": {
    "build": "tsc",
    "dev": "tsc --watch"
  }
}
```

</details>

## Interview Questions

### Q1: What's the difference between "target" and "module"?

This question reveals whether you understand what TypeScript actually does (transpile) versus what you assume it does (magic). Many developers treat these as interchangeable or don't realize they're independent concerns. The answer shows you've thought about how TypeScript compiles to different environments.

<details>
<summary>Answer</summary>

- **target**: JavaScript version of emitted code (ES5, ES2020, etc.)
  - Controls syntax: arrow functions, classes, async/await
  - Example: `target: "ES5"` converts `() => {}` to `function() {}`

- **module**: Module system for imports/exports
  - Controls: `import`/`export` vs `require`/`module.exports`
  - Example: `module: "CommonJS"` converts `import` to `require()`

```json
{
  "target": "ES5", // Old browsers
  "module": "CommonJS", // Node.js
  "lib": ["ES2015", "DOM"] // Can use Promise, Map, etc.
}
```

</details>

### Q2: Why use "isolatedModules": true?

Interviewers ask this to see if you've worked with modern build tools. It shows whether you understand the difference between TypeScript's compiler (tsc) and faster single-file transpilers like esbuild or SWC. If you know why this option exists, you've likely debugged build setup or optimized compilation speed.

<details>
<parameter name="new_string">Answer</summary>

Required for build tools that transpile files independently (Babel, esbuild, SWC).

**What it enforces**:

```typescript
// ❌ Error with isolatedModules: true
const enum Foo {
  A = 1,
} // Can't inline across files

export { SomeType }; // ❌ Re-exporting type without type keyword

// ✓ Correct
export type { SomeType };
```

**Why?** Single-file transpilers can't resolve cross-file type references.

</details>

### Q3: When to use "skipLibCheck"?

This separates pragmatic developers from purists. It's a performance vs correctness tradeoff that has a right answer for production (skip it), but junior developers often think "more checking = better." The question tests whether you understand when to optimize compilation speed and can justify tradeoffs.

<details>
<summary>Answer</summary>

**Always set to `true` for faster builds.**

Skips type-checking in `.d.ts` files from `node_modules`.

**Why?**

- Faster compilation (skip checking 3rd party types)
- Avoid errors in dependencies you can't fix
- Your code is still fully type-checked

**When not to use**: If debugging type errors in dependencies.

</details>

### Q4: What does "composite": true do?

<details>
<summary>Answer</summary>

Enables TypeScript project references. Required for multi-package projects.

**Effects**:

- Enables incremental compilation
- Generates `.tsbuildinfo` cache file
- Requires `declaration: true`
- Enforces `rootDir` constraints

**Use case**: Monorepos where package A depends on package B.

```json
{
  "composite": true,
  "declaration": true
}
```

</details>

### Q5: Why .js extension for .ts imports with NodeNext?

This is the #1 "gotcha" question for modern Node + TypeScript setups. Every developer encounters this, gets confused, thinks it's a bug, and eventually learns why. Answering this shows you've worked with ESM in Node, understand the compilation model, and know why TypeScript made this design choice (even if you don't like it).

<details>
<parameter name="new_string">Answer</summary>

```typescript
// file.ts
import { foo } from "./bar.js"; // Not a typo!
```

**Reason**: TypeScript emits `.js` files, and Node resolves imports in the emitted code.

```typescript
// Source: import './bar.js'
// Emitted: import './bar.js'  (unchanged)
// Node resolves: ./bar.js ✓
```

If you wrote `'./bar.ts'`, Node would look for `bar.ts.js` (wrong).

</details>

## Common Pitfalls

### 1. Path Mapping Without Runtime Support

```json
{
  "paths": {
    "@utils/*": ["utils/*"]
  }
}
```

```typescript
import { foo } from "@utils/bar"; // ✓ TypeScript happy
// Runtime: ❌ Error: Cannot find module '@utils/bar'
```

**Fix**: Use bundler or `tsconfig-paths` for Node.

### 2. Mixing ESM and CommonJS

```json
// package.json
{ "type": "module" }

// tsconfig.json
{ "module": "CommonJS" }  // ❌ Conflict!
```

**Fix**: Use `"module": "NodeNext"` to respect package.json.

### 3. Forgetting Declaration Files

```json
{
  "declaration": false // ❌ Libraries need .d.ts files!
}
```

**Fix**: Always `"declaration": true` for publishable packages.

## Recommended Configs

### For Libraries

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "module": "ESNext",
    "declaration": true,
    "declarationMap": true,
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true
  }
}
```

### For Node.js Apps (Modern)

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "NodeNext",
    "moduleResolution": "NodeNext",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "sourceMap": true,
    "outDir": "./dist"
  }
}
```

### For Frontend (React/Vue)

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "module": "ESNext",
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "jsx": "react-jsx", // or "preserve" for Vue
    "moduleResolution": "bundler",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "isolatedModules": true
  }
}
```

## Key Takeaways

1. **strict: true** - Always enable, fix issues incrementally
2. **module** - Use `NodeNext` for modern Node, `ESNext` for bundlers
3. **moduleResolution** - `node16`/`nodenext` for modern Node, `bundler` for webpack/vite
4. **esModuleInterop** - Essential for importing CommonJS modules
5. **isolatedModules** - Required for fast transpilers (Babel, esbuild)
6. **skipLibCheck** - Always true for performance
7. **paths** - Type-checking only, need runtime support
8. **Project references** - Use composite for monorepos

## Next Steps

In [Lesson 03: Type System Internals](lesson-03-type-system-internals.md), you'll learn:

- Type inference and widening
- Type narrowing techniques
- Structural vs nominal typing
- Variance (covariance, contravariance)
