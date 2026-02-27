# Lesson 07: Declaration Files

Master .d.ts files - writing type definitions, ambient declarations, and contributing to DefinitelyTyped.

## What Are Declaration Files?

TypeScript needs to type-check JavaScript code, but JavaScript has no type annotations. Declaration files (.d.ts) are the bridge: they describe the shape of JavaScript code without changing it. When you install a library (lodash, express, react), TypeScript needs to know "what does this function take and return?" Declaration files answer that question. They're also how you publish types alongside your own JavaScript packages.

Declaration files (.d.ts) provide type information for JavaScript code.

```typescript
// math.js (JavaScript)
export function add(a, b) {
  return a + b;
}

// math.d.ts (Type declarations)
export function add(a: number, b: number): number;
```

**Use cases**:

1. Type JavaScript libraries
2. Publish types with npm packages
3. Share types across projects
4. Contribute to DefinitelyTyped

## Basic Declaration Syntax

Declaration files use `declare` to say "this exists at runtime, here's its type." You're not implementing anything—just describing what already exists. Think of it as writing a contract for code that lives elsewhere. The syntax mirrors regular TypeScript but without implementations: functions have no body, classes have no logic, variables have no values.

### Function Declarations

Functions in .d.ts files describe the signature without implementation. You specify parameter types, return type, overloads (multiple signatures), and generics. This is how libraries document their API surface—`declare function` means "at runtime there's a function called this, here's how to call it."

```typescript
// Functions
declare function greet(name: string): string;
declare function optional(required: string, optional?: number): void;
declare function overloaded(x: number): number;
declare function overloaded(x: string): string;

// With generics
declare function identity<T>(value: T): T;
```

### Variable Declarations

You're declaring "this variable exists" without providing a value. Use `const` for immutable, `let` for mutable. Useful for global variables from script tags (jQuery, analytics libraries) or configuration objects loaded at runtime. The type describes what shape the variable has when the JavaScript runs.

```typescript
// Variables
declare const version: string;
declare let count: number;

// Objects
declare const config: {
  apiUrl: string;
  timeout: number;
};
```

### Class Declarations

Describe the public interface of a class: constructor signature, methods, properties, static members. No method bodies—just the contract. This is how you type JavaScript classes or describe class instances from libraries. Mark properties `readonly` if they shouldn't be reassigned.

```typescript
declare class EventEmitter {
  // Constructor
  constructor(options?: { maxListeners?: number });

  // Methods
  on(event: string, listener: Function): this;
  emit(event: string, ...args: any[]): boolean;

  // Properties
  readonly eventNames: string[];

  // Static members
  static defaultMaxListeners: number;
}
```

### Interface & Type Declarations

Pure type definitions—no runtime representation. Interfaces and types in .d.ts files work exactly like regular TypeScript. Use them to describe object shapes, unions, generics, and complex types. These are the building blocks for your function and class declarations.

```typescript
interface User {
  id: number;
  name: string;
  email?: string;
}

type Status = "pending" | "active" | "inactive";

type Result<T> = { success: true; data: T } | { success: false; error: string };
```

## Module Declarations

Different module systems (ESM, CommonJS, UMD) have different declaration patterns. ESM uses `export`/`export default`. CommonJS uses `export =` and namespaces. UMD supports both module and global usage. The pattern you choose must match how the JavaScript module actually exports at runtime—get this wrong and imports won't type-check correctly.

### ES Module

The modern standard. Use `export` for named exports, `export default` for default export. This matches how ESM JavaScript works—TypeScript will understand `import { foo } from 'lib'` and `import Lib from 'lib'` based on your declarations.

```typescript
// my-lib.d.ts
export interface Config {
  apiKey: string;
}

export function initialize(config: Config): void;

export default class MyLib {
  constructor(config: Config);
  run(): void;
}
```

### CommonJS Module

The Node.js legacy pattern. `export =` declares what `module.exports` is. `declare namespace` adds properties to the export. This is how you type old Node libraries that use `module.exports = MyClass` and attach extra properties like `MyClass.helper = function() {}`. Consumers use `import MyLib = require('my-lib')`.

```typescript
// my-lib.d.ts
export = MyLib;

declare class MyLib {
  constructor(config: MyLib.Config);
  run(): void;
}

declare namespace MyLib {
  interface Config {
    apiKey: string;
  }
  function helper(): void;
}
```

Usage:

```typescript
import MyLib = require("my-lib");
const lib = new MyLib({ apiKey: "xxx" });
```

### UMD (Universal Module Definition)

The "works everywhere" pattern for libraries that support both module imports and script tag globals. `export as namespace MyLib` means "if loaded via `<script>`, it's available as `window.MyLib`." Used by libraries like jQuery, lodash (when using UMD builds) that need maximum compatibility.

```typescript
// my-lib.d.ts
export as namespace MyLib; // Global variable when loaded via script tag

export interface Config {
  apiKey: string;
}

export function initialize(config: Config): void;
```

Usage:

```typescript
// As module
import { initialize } from "my-lib";

// As global (script tag)
MyLib.initialize({ apiKey: "xxx" });
```

## Ambient Declarations

"Ambient" means "already exists in the environment." These declarations don't generate JavaScript—they tell TypeScript about runtime things it can't see: global variables from script tags, untyped npm packages, browser APIs, Node.js builtins. Without ambient declarations, TypeScript would error on `window.myLib` or `import 'untyped-package'`.

Declare types for things that exist at runtime but have no TypeScript source.

### Ambient Modules

You `npm install untyped-library` but there's no @types package. TypeScript errors on `import { foo } from 'untyped-library'`. Ambient module declarations solve this: you write the types yourself in a .d.ts file. Now TypeScript understands imports from that module. Quick fix for untyped dependencies.

```typescript
// globals.d.ts
declare module "untyped-library" {
  export function doSomething(value: string): number;
  export const version: string;
}

// Now can import:
import { doSomething } from "untyped-library";
```

### Wildcard Module Declarations

You `import styles from './app.css'` or `import logo from './logo.png'` in webpack/vite projects. TypeScript doesn't know how to type CSS or PNG imports—it only understands .ts/.tsx. Wildcard declarations (`declare module '*.css'`) tell TypeScript "any .css import is this type." Essential for bundler setups that transform non-JS files.

```typescript
// For importing non-JS files
declare module "*.css" {
  const content: { [className: string]: string };
  export default content;
}

declare module "*.png" {
  const value: string;
  export default value;
}

declare module "*.json" {
  const value: any;
  export default value;
}

// Usage:
import styles from "./app.css";
import logo from "./logo.png";
import data from "./config.json";
```

### Global Augmentation

You need to add properties to `window`, type environment variables in `process.env`, or declare global variables from third-party scripts. `declare global { }` lets you extend built-in global types. Common for typing analytics snippets, feature flags loaded via script tag, or environment variables in Node.js apps.

```typescript
// Extend global namespace
declare global {
  interface Window {
    myApp: {
      version: string;
      config: Record<string, any>;
    };
  }

  namespace NodeJS {
    interface ProcessEnv {
      DATABASE_URL: string;
      API_KEY: string;
    }
  }

  var DEBUG: boolean;
}

// Must export to be treated as module
export {};
```

Usage:

```typescript
window.myApp.version; // ✓ Type-safe
process.env.DATABASE_URL; // ✓ Type-safe
if (DEBUG) {
  // ✓ Type-safe
  console.log("Debug mode");
}
```

## Triple-Slash Directives

Legacy compiler instructions using special comments (`/// <reference ...>`). Before tsconfig.json, these were how you included types. Now mostly obsolete—prefer tsconfig `types` and `lib` options. Still seen in older codebases and generated .d.ts files. They must be at the top of the file before any code.

Special comments that provide instructions to the compiler.

### Reference Types

Tells TypeScript to include types from a specific @types package. Useful in .d.ts files that need Node.js types but don't have a tsconfig.json. Modern projects configure this in tsconfig instead.

```typescript
/// <reference types="node" />

// Now Node.js types are available
const buffer: Buffer = Buffer.from("hello");
```

### Reference Path

Include types from a specific file path. Useful when you have hand-written .d.ts files that aren't auto-discovered. Can create brittle dependencies—prefer proper module imports when possible.

```typescript
/// <reference path="./custom-types.d.ts" />

// Include types from another file
```

### Reference Lib

Include specific TypeScript lib files (DOM types, ES2015 types, etc.). Useful in .d.ts files for libraries. Modern codebases use tsconfig.json `lib` option instead.

```typescript
/// <reference lib="es2015" />
/// <reference lib="dom" />

// Include specific library types
```

**Note**: Usually prefer tsconfig.json `types` and `lib` options instead.

## Publishing Type Definitions

When you publish a TypeScript or JavaScript package, users need types. Two options: bundle .d.ts files with your package (preferred, always in sync), or publish separately to @types (for JavaScript libraries maintained by others). Bundled types mean zero extra install, immediate updates. Separate @types means community can add types to any JavaScript library.

### Option 1: Bundle with Package

The gold standard. TypeScript generates .d.ts files during build (`tsc`), you include them in your npm package, and point `"types"` in package.json to the entry point. Users install your package and get types automatically—no separate `@types/your-package` install. Types are always in sync with your code. Every new TypeScript package should do this.

```
my-package/
  dist/
    index.js
    index.d.ts    ← Generated by TypeScript
  src/
    index.ts
  package.json
  tsconfig.json
```

**package.json**:

```json
{
  "name": "my-package",
  "main": "./dist/index.js",
  "types": "./dist/index.d.ts",
  "files": ["dist"],
  "scripts": {
    "build": "tsc"
  }
}
```

**tsconfig.json**:

```json
{
  "compilerOptions": {
    "declaration": true, // Generate .d.ts
    "declarationMap": true, // Generate .d.ts.map
    "outDir": "./dist"
  }
}
```

### Option 2: Separate @types Package

For JavaScript libraries without built-in types. Community contributors write types in DefinitelyTyped (massive GitHub repo), which publishes to `@types/library-name`. Users `npm install @types/library-name` separately. Types can lag behind the library or become outdated. If you maintain the library, prefer Option 1. If you're typing someone else's library, contribute to DefinitelyTyped.

For JavaScript libraries without built-in types.

```
@types/my-library/
  index.d.ts
  package.json
```

**Published to**: `@types/my-library` on npm (via DefinitelyTyped).

## Writing Quality Declarations

Good declarations are specific, use overloads for precise types, mark readonly where appropriate, and document with JSDoc. Bad declarations use `any`, miss overloads (forcing unions where overloads are clearer), or misrepresent mutability. Quality matters—these types are your API's documentation and users' type safety depends on accuracy.

### Do's

Specific types over `any`. Overloads for different call signatures. Readonly for immutable properties. Generics for reusable patterns. JSDoc for documentation (shows in IDE hovers). These practices make your library easy to use and hard to misuse.

```typescript
// ✓ Specific types
export function parse(input: string): ParsedResult;

// ✓ Proper overloads
export function get(key: string): string | undefined;
export function get(key: string, fallback: string): string;

// ✓ Readonly where appropriate
export interface Config {
  readonly apiKey: string;
}

// ✓ Generics for reusable types
export function map<T, U>(arr: T[], fn: (item: T) => U): U[];

// ✓ Document with JSDoc
/**
 * Fetches user by ID
 * @param id - User identifier
 * @returns User object or null if not found
 */
export function fetchUser(id: number): Promise<User | null>;
```

### Don'ts

`any` destroys type safety. Missing overloads force users into awkward type guards. Incorrect mutability (`apiKey` should be readonly but isn't) allows bugs. Poor generic constraints (`<T>` with no constraints when `<T extends Processable>` would be clearer) obscure intent. These mistakes make your library harder to use correctly.

```typescript
// ❌ Overly permissive
export function doSomething(input: any): any;

// ❌ Missing overloads
export function get(key: string, fallback?: string): string | undefined;
// Better: Use overloads for different return types

// ❌ Incorrect mutability
export interface Config {
  apiKey: string; // Can be reassigned
}

// ❌ Poor generic constraints
export function process<T>(value: T): T;
// Better: Add meaningful constraints
export function process<T extends Processable>(value: T): T;
```

## Advanced Patterns

Real-world declaration files use sophisticated patterns: namespace merging (jQuery-style APIs), conditional types (generic helpers), branded types (prevent mixing IDs). These patterns solve actual problems: typing libraries with complex APIs, providing utility types, preventing runtime bugs with compile-time checks.

### Namespace Merging

jQuery-style pattern: `$(selector)` is a function, but `$.ajax()` is a static method. This requires merging an interface (the instance) with a namespace (the static methods). Interface merging lets you describe this shape. Useful for libraries with both instance and static APIs.

```typescript
// Combine interface and namespace
export interface JQuery {
  text(): string;
  text(value: string): this;
}

export namespace JQuery {
  interface AjaxSettings {
    url: string;
    method?: string;
  }

  function ajax(settings: AjaxSettings): Promise<any>;
}

// Usage:
declare const $: JQuery;
$.text("hello");
JQuery.ajax({ url: "/api" });
```

### Conditional Types in Declarations

Generic helpers that adapt based on input type. `AsyncOrSync<T>` wraps non-promises in `Promise` but leaves promises alone. Useful for library functions that normalize sync/async values. Conditional types in declarations enable flexible, type-safe APIs.

```typescript
export type AsyncOrSync<T> = T extends Promise<any> ? T : Promise<T>;

export function wrap<T>(value: T): AsyncOrSync<T>;

// Usage:
const a = wrap(42); // Promise<number>
const b = wrap(Promise.resolve(42)); // Promise<number>
```

### Branded Types

String IDs all look the same to TypeScript: `UserId` and `PostId` are both `string`. Branded types add phantom properties (`{ __brand: 'UserId' }`) that prevent mixing. You can't pass `PostId` where `UserId` is expected—compiler error. Zero runtime cost, massive safety improvement. Used for IDs, URLs, file paths—anywhere primitive types need distinction.

```typescript
// Prevent mixing different ID types
export type UserId = string & { __brand: "UserId" };
export type PostId = string & { __brand: "PostId" };

export function getUserById(id: UserId): User;
export function getPostById(id: PostId): Post;

// Can't accidentally mix:
declare const userId: UserId;
declare const postId: PostId;

getUserById(userId); // ✓
getUserById(postId); // ❌ Error
```

## DefinitelyTyped Contribution

DefinitelyTyped is the massive community-maintained repo of types for JavaScript libraries. When a library has no built-in types, someone writes them and publishes to `@types/library-name`. Contributing means writing declarations, tests, and following strict guidelines. Your PR gets reviewed, merged, and auto-published to npm. Millions of developers use these types daily.

### Project Structure

Standard DefinitelyTyped layout: `index.d.ts` for declarations, `-tests.ts` for type tests (code that must compile), `tsconfig.json` for compiler settings. The tests aren't runtime tests—they're code that proves your types work by compiling successfully.

```
 DefinitelyTyped/
  types/
    my-library/
      index.d.ts
      my-library-tests.ts
      tsconfig.json
      tslint.json
```

### Example Declaration

Header comments with library info, project URL, and author. Then standard declarations matching the library's API. Keep it simple, accurate, and well-documented. Check existing @types packages for style guidance.

```typescript
// index.d.ts
// Type definitions for my-library 1.0
// Project: https://github.com/user/my-library
// Definitions by: Your Name <https://github.com/yourusername>
// Definitions: https://github.com/DefinitelyTyped/DefinitelyTyped

export interface Options {
  timeout?: number;
  retries?: number;
}

export function connect(url: string, options?: Options): Promise<Connection>;

export class Connection {
  send(data: string): void;
  close(): void;
}
```

### Test File

Example usage that must type-check. This proves your declarations work. Not runtime tests—just code that compiles. If this file compiles, your types are correct.

```typescript
// my-library-tests.ts
import { connect, Options } from "my-library";

const options: Options = {
  timeout: 5000,
  retries: 3,
};

connect("ws://localhost", options).then((conn) => {
  conn.send("hello");
  conn.close();
});
```

### tsconfig.json

Strict compiler settings required by DefinitelyTyped. `noImplicitAny`, `strictNullChecks`, etc. ensure high-quality types. `noEmit` because you're not generating JavaScript—just type-checking.

```json
{
  "compilerOptions": {
    "module": "commonjs",
    "lib": ["es6"],
    "noImplicitAny": true,
    "noImplicitThis": true,
    "strictNullChecks": true,
    "strictFunctionTypes": true,
    "types": [],
    "noEmit": true,
    "forceConsistentCasingInFileNames": true
  },
  "files": ["index.d.ts", "my-library-tests.ts"]
}
```

## Hands-On Exercise 1: Write Declaration File

Create type definitions for this JavaScript module:

```javascript
// cache.js
class Cache {
  constructor(options) {
    this.maxSize = options?.maxSize || 100;
    this.data = new Map();
  }

  set(key, value, ttl) {
    this.data.set(key, { value, expires: Date.now() + (ttl || 0) });
  }

  get(key) {
    const item = this.data.get(key);
    if (!item) return undefined;
    if (item.expires && item.expires < Date.now()) {
      this.data.delete(key);
      return undefined;
    }
    return item.value;
  }

  clear() {
    this.data.clear();
  }
}

module.exports = Cache;
```

<details>
<summary>Solution</summary>

```typescript
// cache.d.ts
export = Cache;

declare class Cache<K = string, V = any> {
  constructor(options?: Cache.Options);

  set(key: K, value: V, ttl?: number): void;
  get(key: K): V | undefined;
  clear(): void;

  readonly maxSize: number;
}

declare namespace Cache {
  interface Options {
    maxSize?: number;
  }
}
```

Usage:

```typescript
import Cache = require("./cache");

const cache = new Cache<string, number>({ maxSize: 50 });
cache.set("count", 42, 1000);
const value = cache.get("count"); // number | undefined
```

</details>

## Hands-On Exercise 2: Augment Express

Add custom properties to Express Request:

```typescript
// Goal: Make these type-safe
app.get("/", (req, res) => {
  const userId = req.userId; // Should be string
  const session = req.session; // Should be Session
});
```

<details>
<summary>Solution</summary>

```typescript
// types/express.d.ts
import { Session } from "./session";

declare global {
  namespace Express {
    interface Request {
      userId: string;
      session: Session;
    }
  }
}

// Must export to be treated as module
export {};
```

Or with module augmentation:

```typescript
// types/express-augmentation.d.ts
import { Session } from "./session";

declare module "express-serve-static-core" {
  interface Request {
    userId: string;
    session: Session;
  }
}
```

</details>

## Interview Questions

### Q1: When to use declare keyword?

Tests understanding of ambient declarations vs regular code. Strong candidates explain "declare" is for describing existing runtime things, not implementing new things. Weak answers confuse it with regular `const`/`function` declarations.

<details>
<summary>Answer</summary>

Use `declare` for ambient declarations - types for things that exist at runtime.

```typescript
// Declaring global variable from script tag
declare const jQuery: {
  (selector: string): any;
};

// Declaring module without implementation
declare module "my-untyped-lib" {
  export function doThing(): void;
}

// Declaring global namespace
declare namespace MyGlobal {
  function init(): void;
}
```

**Don't use** in regular .ts files (implementations) - only in .d.ts files.

</details>

### Q2: How to type a library with both default and named exports?

Tests experience with CommonJS module patterns and the `export =` syntax. This pattern is common in older Node libraries and confuses many developers. Strong answers show both CommonJS and ESM solutions.

<details>
<summary>Answer</summary>

```typescript
// Library: module.exports = foo; module.exports.bar = bar;

// Declaration:
export = MyLib;

declare function MyLib(): void;

declare namespace MyLib {
  export function bar(): void;
  export const version: string;
}

// Usage:
import MyLib = require("my-lib");

MyLib(); // Default
MyLib.bar(); // Named
MyLib.version; // Named
```

For ESM:

```typescript
declare const myLib: {
  (): void;
  bar(): void;
  version: string;
};

export default myLib;
export const bar: typeof myLib.bar;
export const version: string;
```

</details>

### Q3: What's the difference between types and @types?

Reveals practical understanding of npm package types vs community types. Candidates who've published packages understand built-in types are better (always in sync). Those who've only consumed libraries might not realize @types can lag.

<details>
<summary>Answer</summary>

**Built-in types** (with package):

```json
// package.json
{
  "name": "my-lib",
  "types": "./dist/index.d.ts"
}
```

Published together, always in sync.

**@types packages** (separate):

```bash
npm install @types/my-lib
```

Community-maintained, may lag behind library updates.

**Preference**: Use built-in types when available. Use @types for libraries without types.

</details>

### Q4: How do triple-slash directives work?

Tests knowledge of TypeScript history and legacy patterns. Strong answers explain modern alternatives (tsconfig.json). Weak answers might think triple-slash directives are still the primary way to configure TypeScript.

<details>
<summary>Answer</summary>

Special comments for compiler instructions.

```typescript
/// <reference types="node" />
// Include @types/node

/// <reference path="./types.d.ts" />
// Include specific file

/// <reference lib="es2015" />
// Include specific lib
```

**Modern alternative**: Use tsconfig.json:

```json
{
  "compilerOptions": {
    "types": ["node"],
    "lib": ["ES2015"]
  }
}
```

</details>

## Key Takeaways

1. **.d.ts files**: Provide types for JavaScript code
2. **declare**: For ambient declarations (global, modules)
3. **Module declarations**: Support ESM, CommonJS, UMD
4. **Global augmentation**: Extend Window, NodeJS.ProcessEnv, etc.
5. **Publishing**: Bundle .d.ts or publish to @types
6. **Quality**: Specific types, overloads, readonly, generics
7. **DefinitelyTyped**: Community type definitions

## Next Steps

In [Lesson 08: Publishing npm Packages](lesson-08-publishing-npm-packages.md), you'll learn:

- Build setup for dual ESM/CJS publishing
- Version management and semver
- npm publishing workflow
- Package distribution best practices
