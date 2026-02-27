# Lesson 06: Module Systems

Deep dive into ESM vs CommonJS, module resolution, and import/export patterns.

## ESM vs CommonJS

JavaScript didn't have a module system for its first 15 years. Node.js invented CommonJS in 2009 because `<script>` tags weren't enough for server-side code. Then in 2015, ES6 introduced native modules (ESM). Now we live in the transition period where both exist, and you need to understand both to work effectively with Node.js, TypeScript, bundlers, and published packages. The module system you choose affects bundle size, tree-shaking, top-level await, and how your code runs in production.

### CommonJS (Node.js Traditional)

CommonJS solved the "how do I split my server code into files?" problem when Node.js was created. It's synchronous because Node reads from the local filesystem (fast), uses `require()` for imports and `module.exports` for exports. Most of the npm ecosystem was built on this, and it still powers millions of packages today.

```javascript
// math.js - Exporting
exports.add = (a, b) => a + b;
exports.subtract = (a, b) => a - b;

// Or
module.exports = {
  add: (a, b) => a + b,
  subtract: (a, b) => a - b,
};

// Or default export
module.exports = class Calculator {
  add(a, b) {
    return a + b;
  }
};

// app.js - Importing
const math = require("./math");
const { add } = require("./math");
const Calculator = require("./calculator");
```

**Characteristics**:

- Synchronous loading
- Dynamic (can `require()` conditionally)
- Runtime resolution
- `this` equals `exports`
- File extension optional

### ESM (ECMAScript Modules)

ESM is the JavaScript standard that works everywhere: browsers, Node.js, Deno, bundlers. It's statically analyzable (bundlers can tree-shake), asynchronous (works with network loading in browsers), and supports top-level await. The tradeoff is stricter syntax: you can't conditionally import with static `import` (use dynamic `import()` instead), and Node 16+ requires file extensions.

```typescript
// math.ts - Exporting
export const add = (a: number, b: number) => a + b;
export const subtract = (a: number, b: number) => a - b;

// Or
export { add, subtract };

// Default export
export default class Calculator {
  add(a: number, b: number) {
    return a + b;
  }
}

// app.ts - Importing
import * as math from "./math.js"; // Note: .js extension!
import { add } from "./math.js";
import Calculator from "./calculator.js";
```

**Characteristics**:

- Static analysis possible
- Async loading
- Compile-time resolution
- `this` is `undefined` at top level
- File extension required (in Node16+)

### Side-by-Side Comparison

When you see "tree-shaking doesn't work" or "bundle size is huge," module format is often the culprit. When top-level await fails, you're probably in CommonJS. This table is the cheat sheet for "why isn't this working?"

| Feature         | CommonJS                       | ESM                 |
| --------------- | ------------------------------ | ------------------- |
| Syntax          | `require()` / `module.exports` | `import` / `export` |
| Loading         | Synchronous                    | Async               |
| When resolved   | Runtime                        | Parse time          |
| Tree-shaking    | ❌ No                          | ✓ Yes               |
| Dynamic imports | Always                         | Via `import()`      |
| Top-level await | ❌ No                          | ✓ Yes (Node 14.8+)  |
| File extension  | Optional                       | Required (Node16+)  |
| `this`          | `exports`                      | `undefined`         |

## TypeScript Compilation

TypeScript is a compile-time layer; Node.js, browsers, and bundlers run JavaScript. This gap means you need to tell TypeScript "what JavaScript should I emit?" The `module` setting controls whether your beautiful `import`/`export` becomes `require()`/`exports` or stays as ESM. Get this wrong and your code won't run in production—even if TypeScript compiles successfully.

### Output Module Format

The `module` compiler option determines what JavaScript your TypeScript becomes. Use `CommonJS` if you're targeting older Node or AWS Lambda. Use `ES2020`/`ESNext` if you're using a bundler. Use `NodeNext` if you're shipping ESM to modern Node.js (16+). This isn't about preferences—it's about what your runtime supports.

```json
// tsconfig.json
{
  "compilerOptions": {
    "module": "CommonJS" // or "ES2020", "ESNext", "NodeNext"
  }
}
```

```typescript
// Source (TypeScript)
import { foo } from "./bar";
export const baz = 42;

// Output with "module": "CommonJS"
const bar_1 = require("./bar");
exports.baz = 42;

// Output with "module": "ES2020"
import { foo } from "./bar";
export const baz = 42;
```

### Interop Flags

The "why can't I import express normally?" problem. CommonJS modules export with `module.exports`, but ESM expects `export default`. These worlds don't align. `esModuleInterop` adds runtime helpers so you can write `import express from 'express'` instead of `import * as express from 'express'`. It's not magic—it's just wrapping the CJS export to look like an ESM default.

```json
{
  "compilerOptions": {
    "esModuleInterop": true,
    "allowSyntheticDefaultImports": true
  }
}
```

**Without esModuleInterop**:

```typescript
import * as express from "express"; // Must use namespace import
const app = express();
```

**With esModuleInterop**:

```typescript
import express from "express"; // Can use default import
const app = express();
```

## Module Resolution

You write `import { foo } from './bar'` and TypeScript/Node needs to find the file. Is it `bar.ts`? `bar.js`? `bar/index.ts`? Does it respect package.json `exports`? Do you need a file extension? The `moduleResolution` setting answers these questions and MUST match your runtime environment (Node, bundler, etc.) or imports will compile but fail at runtime.

### Resolution Strategies

Different runtimes have different rules. Webpack doesn't require extensions. Modern Node.js (16+) requires `.js` even for TypeScript files and respects package.json `exports` fields. The `bundler` setting gives you the best of both worlds for webpack/Vite projects: strict typing like `node16` but relaxed extensions like the old `node` mode.

```json
{
  "compilerOptions": {
    "moduleResolution": "node" | "node16" | "nodenext" | "bundler"
  }
}
```

| Strategy              | Use Case          | Behavior                                             |
| --------------------- | ----------------- | ---------------------------------------------------- |
| `node`                | Legacy Node       | Classic Node resolution                              |
| `node16` / `nodenext` | Modern Node (16+) | Respects package.json `exports`, requires extensions |
| `bundler`             | Webpack/Vite      | Like node16 but relaxed (no extensions needed)       |

### Classic Resolution (node)

The old algorithm that powered Node.js and TypeScript for years. It tries a bunch of file extensions, looks for index files, and doesn't require you to specify `.js`. Still common in legacy projects and bundler configs. Permissive but less precise than modern resolution.

```
import { foo } from './bar';

Looks for:
1. ./bar.ts
2. ./bar.tsx
3. ./bar.d.ts
4. ./bar/index.ts
5. ./bar/index.tsx
6. ./bar/index.d.ts
```

### Node16/NodeNext Resolution

The modern, stricter algorithm matching Node.js 16+ behavior. Requires file extensions (yes, even for TypeScript—you write `.js` and TypeScript finds the `.ts` source). Respects package.json `exports` field (so libraries control their public API). Breaking change from classic mode, but aligns TypeScript with actual Node.js runtime behavior. Use this if you're publishing ESM packages or using modern Node.

```typescript
// package.json
{ "type": "module" }

// MUST include extension:
import { foo } from './bar.js';  // ✓ (looks for bar.ts)
import { foo } from './bar';     // ❌ Error

// Respects package.json exports:
import { foo } from 'my-lib';    // Uses "exports" field
```

**Critical**: Use `.js` extension even for `.ts` files.

```typescript
// file: src/utils.ts
export const helper = () => {};

// file: src/app.ts
import { helper } from "./utils.js"; // ✓ Correct
// TypeScript finds utils.ts, emits utils.js
```

### Package Exports Resolution

Library authors want control over their public API. The `exports` field in package.json defines "these are the only entry points." You can't deep-import into `/dist/internal/` anymore—if it's not in `exports`, it doesn't exist. This enables better encapsulation and prevents users from depending on internal modules that might change. Used by modern libraries like `next`, `prisma`, and `vitest`.

```json
// node_modules/my-lib/package.json
{
  "exports": {
    ".": {
      "types": "./dist/index.d.ts",
      "import": "./dist/index.mjs",
      "require": "./dist/index.cjs"
    },
    "./utils": "./dist/utils.js"
  }
}
```

```typescript
import lib from "my-lib"; // Uses exports["."]
import { util } from "my-lib/utils"; // Uses exports["./utils"]
import { foo } from "my-lib/internal"; // ❌ Not in exports
```

## Import/Export Patterns

There are many ways to export the same thing. Named exports are explicit and tree-shakeable. Default exports are convenient but can cause naming inconsistencies (everyone imports with a different name). Re-exports let you build public APIs from multiple internal modules. Understanding these patterns helps you design clean, maintainable module interfaces.

### Named Exports

Named exports are explicit: you see exactly what's exported, and importers use the exact names. This enables tree-shaking (bundlers can remove unused exports) and makes refactoring safer (renaming shows up across the codebase). Use named exports for utilities, functions, and constants—anything that's one of many things in a module.

```typescript
// Inline export
export const PI = 3.14;
export function add(a: number, b: number) {
  return a + b;
}

// Export list
const PI = 3.14;
function add(a: number, b: number) {
  return a + b;
}
export { PI, add };

// Re-export
export { foo, bar } from "./other";
export * from "./other"; // Re-export all
```

### Default Export

One export that represents "the main thing" from this module. Classes, components, and single-purpose modules often use default exports. The downside: everyone can import with a different name (`import Calc from './calculator'` vs `import MyCalc from './calculator'`), making searches and refactoring harder. Some teams ban default exports entirely for this reason.

```typescript
// Only one per module
export default class Calculator {
  // ...
}

// Or
class Calculator {
  // ...
}
export default Calculator;

// Or inline value
export default { version: "1.0.0" };
```

### Import Patterns

The consumer side of exports. Destructuring (`{ foo, bar }`) is common for named exports. Aliasing (`as`) solves naming conflicts. Namespace imports (`* as utils`) group related functions. Side-effect imports (`import './polyfills'`) run code without importing values (useful for global setup, CSS, polyfills).

```typescript
// Named imports
import { foo, bar } from "./module";

// Aliasing
import { foo as f, bar as b } from "./module";

// Namespace import
import * as utils from "./utils";
utils.foo();

// Default import
import Calculator from "./calculator";

// Mixed
import Calculator, { add, subtract } from "./math";

// Side-effects only
import "./polyfills";
```

### Re-exporting

Building a public API from multiple internal modules. You have `./internal/auth.ts`, `./internal/db.ts`, `./internal/logger.ts`, but you want users to import from `'my-lib'`, not `'my-lib/internal/auth'`. Re-exports let you create a single entry point (`index.ts`) that aggregates everything. This is how libraries expose a clean API while organizing code internally.

```typescript
// Re-export named
export { foo, bar } from "./other";

// Re-export all
export * from "./other";

// Re-export with rename
export { foo as newFoo } from "./other";

// Re-export default as named
export { default as Calculator } from "./calculator";
```

## Dynamic Imports

Static imports load everything upfront, bloating your initial bundle. Dynamic imports (`import()`) return a Promise and load modules on-demand. This enables code splitting (ship a 10KB bundle, load 100KB feature only when needed), lazy loading (load when user clicks), and conditional features (load analytics only if user opts in). Essential for performance in large apps.

```typescript
// Static import - always loaded
import { large } from "./large-module";

// Dynamic import - loaded on demand
async function loadFeature() {
  const { large } = await import("./large-module");
  large();
}

// Conditional loading
if (condition) {
  const module = await import("./conditional");
}

// Type-safe dynamic import
type MathModule = typeof import("./math");

async function getMath(): Promise<MathModule> {
  return import("./math");
}
```

### Use Cases for Dynamic Imports

This isn't theoretical—these patterns directly impact bundle size, load time, and user experience. Route-based splitting means your landing page doesn't load the admin panel code. Feature flags let you ship code to production without enabling it. Lazy loading defers non-critical code until it's actually needed. Every big production app uses these patterns.

1. **Code splitting** (reduce initial bundle)
2. **Conditional loading** (feature flags)
3. **Lazy loading** (on user interaction)
4. **Dynamic module paths**

```typescript
// Route-based loading
async function loadRoute(route: string) {
  const module = await import(`./routes/${route}`);
  return module.default;
}

// Feature flags
if (featureFlags.newFeature) {
  const { NewFeature } = await import("./new-feature");
  // Use NewFeature
}
```

## Top-Level Await

ESM's killer feature: you can `await` at the module's top level without wrapping in an async function. Perfect for fetching config, connecting to databases, or loading remote data during module initialization. The catch: it blocks module loading, so if you `await fetch()` at top level, nothing else can import your module until that fetch completes. Use carefully—great for entry points, risky for shared utilities.

Only works in ESM (Node 14.8+, "module": "ES2022"+).

```typescript
// ❌ CommonJS
const data = await fetch("/api/data"); // Error: await in top-level

// ✓ ESM (package.json: { "type": "module" })
const response = await fetch("/api/data");
const data = await response.json();

export const config = data;
```

**Use carefully**: Blocks module execution, affects load time.

## Import Assertions (Type Attributes)

JavaScript engines only understand JavaScript. When you `import data from './data.json'`, the engine needs to know "this is JSON, parse it as JSON" (not as JavaScript). Import assertions/attributes tell the runtime what type of file you're importing. Originally `assert { type: 'json' }`, now moving to `with { type: 'json' }` syntax. Mainly used for JSON in Node 17+ and CSS modules in bundlers.

For importing non-JavaScript files (Node 17+).

```typescript
// JSON
import data from "./data.json" assert { type: "json" };

// CSS (in bundlers)
import styles from "./styles.css" assert { type: "css" };
```

**New syntax (TypeScript 5.3+)**:

```typescript
import data from "./data.json" with { type: "json" };
```

## CommonJS Interop

Reality check: you're writing ESM, but half of npm is still CommonJS. You need to import CJS packages from ESM code and occasionally import ESM from CJS. ESM can import CJS with some quirks (default vs named exports). CJS CANNOT statically import ESM (synchronous can't load asynchronous)—you must use dynamic `import()`. This is the messy middle ground we all live in.

### Importing CommonJS in ESM

CommonJS modules export a single object (`module.exports`). ESM expects either named exports or a default export. `esModuleInterop` bridges the gap by treating `module.exports` as a default export. Without it, you need namespace imports (`import * as foo`). With it, you can use cleaner default imports (`import foo`).

```typescript
// CommonJS module: const foo = { bar: 42 }; module.exports = foo;

// ESM import
import foo from "./cjs-module"; // ✓ Works with esModuleInterop
import * as foo from "./cjs-module"; // ✓ Always works

console.log(foo.bar);
```

### Importing ESM in CommonJS

The hard limitation: CommonJS is synchronous, ESM is asynchronous. You cannot use static `import` in a `.cjs` file or CommonJS-mode `.js` file. The only option is dynamic `import()`, which returns a Promise. This is why migrating to ESM is a one-way door—once a package goes ESM-only, CommonJS consumers must use async imports.

```javascript
// ❌ Can't use static import in CommonJS
import { foo } from "./esm-module"; // Error

// ✓ Use dynamic import
async function load() {
  const { foo } = await import("./esm-module.mjs");
  console.log(foo);
}
```

## Module Augmentation

You're using Express and want to add `req.user` to the Request type. You're using a library that's "almost right" but missing one property. Module augmentation lets you extend third-party types without forking the library or modifying node_modules. TypeScript merges your interface additions with the original declarations. Essential for customizing framework types (Express, Next.js, etc.).

Extend existing modules with new declarations.

```typescript
// node_modules/express/index.d.ts
declare namespace Express {
  interface Request {
    user?: User;
  }
}

// Your code: extend Express.Request
declare module "express" {
  interface Request {
    customProp: string;
  }
}

// Now available:
app.get("/", (req, res) => {
  console.log(req.customProp); // ✓ TypeScript knows about it
});
```

### Global Augmentation

Similar to module augmentation, but for global types like `Window`, `NodeJS.ProcessEnv`, or `globalThis`. You need `declare global { }` to add properties to the global scope. Common use case: adding properties to `window` for third-party scripts, extending `process.env` for environment variables, or declaring global functions from CDN-loaded libraries.

```typescript
// Extend global namespace
declare global {
  interface Window {
    myLib: MyLibrary;
  }
}

// Must export something to be a module
export {};
```

## Hands-On Exercise 1: Module Migration

Convert this CommonJS to ESM:

```javascript
// math.js
const add = (a, b) => a + b;
const subtract = (a, b) => a - b;
exports.add = add;
exports.subtract = subtract;

// app.js
const { add } = require("./math");
console.log(add(1, 2));
```

<details>
<summary>Solution</summary>

```typescript
// math.ts
export const add = (a: number, b: number) => a + b;
export const subtract = (a: number, b: number) => a - b;

// app.ts
import { add } from './math.js';  // Note: .js extension
console.log(add(1, 2));

// package.json
{
  "type": "module"
}

// tsconfig.json
{
  "compilerOptions": {
    "module": "NodeNext",
    "moduleResolution": "NodeNext"
  }
}
```

</details>

## Hands-On Exercise 2: Dynamic Import

Implement lazy loading for a feature:

```typescript
// Goal: Load analytics only when user opts in

interface Analytics {
  track(event: string): void;
}

async function initAnalytics(enabled: boolean): Promise<Analytics | null> {
  // Implement
}
```

<details>
<summary>Solution</summary>

```typescript
// analytics.ts
export interface Analytics {
  track(event: string): void;
  pageView(url: string): void;
}

export function createAnalytics(): Analytics {
  return {
    track(event: string) {
      console.log("Track:", event);
    },
    pageView(url: string) {
      console.log("Page view:", url);
    },
  };
}

// app.ts
async function initAnalytics(enabled: boolean): Promise<Analytics | null> {
  if (!enabled) {
    return null;
  }

  const { createAnalytics } = await import("./analytics.js");
  return createAnalytics();
}

// Usage
const analytics = await initAnalytics(true);
analytics?.track("app_start");
```

</details>

## Interview Questions

### Q1: ESM vs CommonJS - key differences?

This question reveals whether you understand the fundamental tradeoffs in module systems: sync vs async, static vs dynamic, tree-shaking vs runtime resolution. Strong candidates explain the "why" behind each difference and when to use each system.

<details>
<summary>Answer</summary>

| Aspect              | CommonJS    | ESM                |
| ------------------- | ----------- | ------------------ |
| **When evaluated**  | Runtime     | Parse time         |
| **Loading**         | Synchronous | Async              |
| **Tree-shaking**    | No          | Yes                |
| **Dynamic imports** | Built-in    | Via `import()`     |
| **Extensions**      | Optional    | Required (Node16+) |
| **Top-level await** | No          | Yes                |

**Why ESM?**

- Better for bundlers (tree-shaking)
- Standard across environments
- Static analysis enables optimization

</details>

### Q2: Why require .js extension in imports with NodeNext?

Tests understanding of TypeScript's compilation model and the gap between source code (.ts) and runtime code (.js). Candidates who get this wrong often struggle with production deployments where imports mysteriously fail.

<details>
<summary>Answer</summary>

```typescript
import { foo } from "./bar.js"; // Why .js for .ts file?
```

**Reason**: TypeScript emits code that Node will run. Node resolves imports in the _emitted_ .js files.

```typescript
// Source: src/app.ts
import { foo } from "./utils.js";

// Emitted: dist/app.js
import { foo } from "./utils.js"; // Node looks for utils.js
```

TypeScript finds the source (.ts) but emits the extension you specified (.js).

</details>

### Q3: What's esModuleInterop?

A practical question about CommonJS/ESM compatibility. Candidates who've actually shipped Node packages understand this flag intimately. Those who haven't often struggle to explain why `import express from 'express'` doesn't work without it.

<details>
<summary>Answer</summary>

Enables default imports from CommonJS modules.

```typescript
// CommonJS module: module.exports = function() {}

// Without esModuleInterop
import * as express from "express";
const app = express(); // ❌ Error: not callable

// With esModuleInterop: true
import express from "express";
const app = express(); // ✓ Works
```

**What it does**: Adds runtime helpers to make CJS default exports compatible with ESM default imports.

</details>

### Q4: When to use dynamic imports?

Reveals performance awareness and understanding of code splitting. Strong answers mention specific bundle size impacts and real-world scenarios (route splitting, feature flags). Weak answers just say "for lazy loading" without explaining why or when.

<details>
<summary>Answer</summary>

**Use cases**:

1. **Code splitting**: Reduce initial bundle

   ```typescript
   const module = await import("./large-feature");
   ```

2. **Conditional loading**: Feature flags

   ```typescript
   if (featureEnabled) {
     await import("./new-feature");
   }
   ```

3. **Lazy loading**: On-demand

   ```typescript
   button.onclick = async () => {
     const { modal } = await import("./modal");
     modal.show();
   };
   ```

4. **Dynamic paths**:
   ```typescript
   const locale = await import(`./i18n/${lang}.js`);
   ```

</details>

### Q5: How does module augmentation work?

Tests experience with TypeScript's declaration merging and extending third-party types. Common in real codebases (extending Express, Next.js, etc.) but often skipped in tutorials. Candidates who've built production apps have battle scars from this.

<details>
<summary>Answer</summary>

Extend existing module types:

```typescript
// Extend external module
declare module "express" {
  interface Request {
    userId?: string;
  }
}

// Now available:
app.use((req, res, next) => {
  req.userId = "123"; // ✓ TypeScript knows
});
```

**Requirements**:

- Must use `declare module 'exact-module-name'`
- File must be a module (has import/export)
- Works with interface merging

</details>

## Key Takeaways

1. **ESM vs CJS**: ESM is standard, better for bundlers, async loading
2. **Module resolution**: Use `NodeNext` for modern Node, `bundler` for webpack/vite
3. **Extensions**: Required in Node16+ ESM (.js for .ts files)
4. **esModuleInterop**: Essential for importing CJS with default syntax
5. **Dynamic imports**: Use for code splitting, lazy loading, conditional features
6. **Top-level await**: Only in ESM, blocks module loading
7. **Module augmentation**: Extend external module types

## Next Steps

In [Lesson 07: Declaration Files](lesson-07-declaration-files.md), you'll learn:

- Writing .d.ts files
- Ambient declarations
- Triple-slash directives
- DefinitelyTyped contributions
