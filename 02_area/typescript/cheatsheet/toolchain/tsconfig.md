# TypeScript tsconfig.json

## Why

- **tsconfig.json is your project contract** — It defines what TypeScript version features you use, how strict the checking is, and what the output looks like. Get it right once and it prevents entire categories of bugs.
- **strict: true is non-negotiable** — Without it, TypeScript silently allows `any`, skips null checks, and misses common bugs. Always start with strict mode. Relaxing individual checks later is fine.
- **target vs module** — `target` controls what JS syntax is emitted (arrow functions, async/await). `module` controls the import/export system (CommonJS vs ESM). They're independent.
- **paths for import aliases** — Avoid `../../../utils`. Map `@/utils` to your source directory. But remember: `paths` only affects TypeScript — your runtime/bundler needs matching config.
- **skipLibCheck: true** — Skips type-checking `.d.ts` files from node_modules. Dramatically speeds up compilation. Rarely hides real bugs.

## Quick Reference

| Setting             | Purpose                                  |
| ------------------- | ---------------------------------------- |
| `strict`            | Enable all strict type checks            |
| `target`            | JS output version (ES2022, ESNext)       |
| `module`            | Module system (Node16, NodeNext, ESNext) |
| `outDir`            | Where compiled JS goes                   |
| `rootDir`           | Root of source files                     |
| `paths`             | Import path aliases                      |
| `skipLibCheck`      | Skip checking .d.ts files                |
| `esModuleInterop`   | Fixes default import from CJS            |
| `resolveJsonModule` | Allow importing .json files              |
| `declaration`       | Generate .d.ts files                     |

## Backend Starter Configs

### 1. Node.js backend (ESM — recommended)

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "Node16",
    "moduleResolution": "Node16",
    "outDir": "dist",
    "rootDir": "src",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "resolveJsonModule": true,
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true
  },
  "include": ["src"],
  "exclude": ["node_modules", "dist"]
}
```

Requires `"type": "module"` in package.json.

### 2. Node.js backend (CommonJS)

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "CommonJS",
    "moduleResolution": "Node",
    "outDir": "dist",
    "rootDir": "src",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "resolveJsonModule": true,
    "declaration": true,
    "sourceMap": true
  },
  "include": ["src"],
  "exclude": ["node_modules", "dist"]
}
```

### 3. Library (dual CJS + ESM output)

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "Node16",
    "moduleResolution": "Node16",
    "outDir": "dist",
    "rootDir": "src",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true,
    "composite": true
  },
  "include": ["src"],
  "exclude": ["node_modules", "dist", "**/*.test.ts"]
}
```

## Key Settings Explained

### 4. target — what JS to emit

| Target   | Key features available                        |
| -------- | --------------------------------------------- |
| `ES2020` | Optional chaining, nullish coalescing, BigInt |
| `ES2021` | `String.replaceAll`, `Promise.any`            |
| `ES2022` | Top-level await, `Array.at()`, error cause    |
| `ES2023` | `Array.findLast`, `toSorted`, `toReversed`    |
| `ESNext` | Latest — moves with each TS release           |

For Node.js, match your minimum supported Node version:

- Node 18 → `ES2022`
- Node 20 → `ES2023`
- Node 22 → `ES2024` or `ESNext`

### 5. module and moduleResolution

| module     | moduleResolution | Use when                       |
| ---------- | ---------------- | ------------------------------ |
| `Node16`   | `Node16`         | Node.js with ESM (recommended) |
| `NodeNext` | `NodeNext`       | Same but tracks latest Node    |
| `CommonJS` | `Node`           | Legacy CJS projects            |
| `ESNext`   | `Bundler`        | Bundled (esbuild, webpack)     |

### 6. strict — what it enables

```json
{
  "strict": true
  // Equivalent to enabling ALL of these:
  // "strictNullChecks": true,
  // "strictFunctionTypes": true,
  // "strictBindCallApply": true,
  // "strictPropertyInitialization": true,
  // "noImplicitAny": true,
  // "noImplicitThis": true,
  // "alwaysStrict": true,
  // "useUnknownInCatchVariables": true
}
```

## Path Aliases

### 7. Configure import aliases

```json
{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@/*": ["src/*"],
      "@config/*": ["src/config/*"],
      "@utils/*": ["src/utils/*"]
    }
  }
}
```

Runtime resolution requires matching config:

```json
// package.json (Node 20+ with --experimental-specifier-resolution)
{
  "imports": {
    "#*": "./src/*"
  }
}
```

Or use `tsc-alias`, `tsconfig-paths`, or a bundler.

## Project References (Monorepo)

### 8. Composite projects

```json
// packages/shared/tsconfig.json
{
  "compilerOptions": {
    "composite": true,
    "outDir": "dist",
    "rootDir": "src"
  }
}

// packages/api/tsconfig.json
{
  "compilerOptions": {
    "outDir": "dist",
    "rootDir": "src"
  },
  "references": [
    { "path": "../shared" }
  ]
}
```

```sh
tsc --build              # incremental build respecting references
tsc --build --clean      # clean build outputs
```

## Watch Mode

### 9. Development with watch

```sh
tsc --watch                    # recompile on change
tsc --watch --preserveWatchOutput  # don't clear terminal
```

For backend dev, pair with `nodemon` or `tsx`:

```sh
# tsx — run TS directly with watch
npx tsx watch src/server.ts

# nodemon + ts-node
npx nodemon --exec ts-node src/server.ts
```

## Useful Extra Settings

### 10. Additional recommended settings

```json
{
  "compilerOptions": {
    "noUncheckedIndexedAccess": true, // arr[0] is T | undefined
    "noUnusedLocals": true, // error on unused variables
    "noUnusedParameters": true, // error on unused params
    "exactOptionalPropertyTypes": true, // distinguish undefined from missing
    "noFallthroughCasesInSwitch": true, // require break in switch
    "isolatedModules": true, // required for esbuild/swc
    "verbatimModuleSyntax": true // enforce type-only imports
  }
}
```
