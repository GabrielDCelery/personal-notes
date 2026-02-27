# Lesson 01: package.json Deep Dive

Critical knowledge about package.json for TypeScript developers - dependency management, versioning, and publishing.

## Dependency Types

When you install a package, npm needs to know: "Should this be installed when someone uses my package? When they develop it? Should they provide it themselves?" Getting this wrong means bloated bundles, missing dependencies at runtime, or version conflicts that break your app. Each dependency type tells npm a different story about how your package relates to its dependencies.

| Type                   | When to Use                     | Installed                | Bundled                 |
| ---------------------- | ------------------------------- | ------------------------ | ----------------------- |
| `dependencies`         | Runtime requirements            | Always                   | Yes (if published)      |
| `devDependencies`      | Build/test tools only           | Only in dev              | No                      |
| `peerDependencies`     | Plugin/extension deps           | Manual (shows warning)   | No                      |
| `optionalDependencies` | Nice-to-have, fallback if fails | Tries, continues on fail | Yes                     |
| `bundledDependencies`  | Force specific versions         | Always                   | Yes (literally bundled) |

### Common Mistakes

```json
{
  "dependencies": {
    "typescript": "^5.0.0" // ❌ Wrong - TypeScript is a build tool
  },
  "devDependencies": {
    "typescript": "^5.0.0", // ✓ Correct
    "express": "^4.18.0" // ❌ Wrong - express is needed at runtime
  }
}
```

### Peer Dependencies Deep Dive

Imagine you're building a React component library. If you bundle React as a regular dependency, every app using your library ends up with TWO copies of React - theirs and yours. This doesn't just waste space; React literally breaks because it relies on singleton state for its reconciliation algorithm. Two React instances = cryptic errors and broken hooks.

Peer dependencies solve this by saying "I need React to work, but I expect YOU (the app) to provide it." This way, there's only one React instance, your library uses whatever version the app has, and everyone shares the same copy.

```json
{
  "name": "my-react-component",
  "peerDependencies": {
    "react": "^18.0.0", // User must have React 18+
    "react-dom": "^18.0.0"
  },
  "peerDependenciesMeta": {
    "react-dom": {
      "optional": true // Won't warn if missing
    }
  },
  "devDependencies": {
    "react": "^18.0.0", // Still needed for development
    "react-dom": "^18.0.0"
  }
}
```

**Why?** Prevents version conflicts. If your library bundled React, users would have two React copies.

What happens if a peer dependency is out of date?

It should fail and block the installation. There are three workarounds:

1. `npm install --legacy-peer-deps` - installs the peer dependency separately
2. `npm install --force` - force installs despite errors (not recommended)
3. fix peer dependency declaration, e.g.

```json
{
  "peerDependencies": {
    "react": "^17.0.0 || ^18.0.0 || ^19.0.0"
  }
}
```

## Semver (Semantic Versioning)

You depend on a library at version 1.4.2. The maintainer releases 1.4.3 with a critical security fix. Should your app automatically get it? What about 1.5.0 with new features? Or 2.0.0 that completely changes the API?

Semantic versioning and version ranges solve this by encoding meaning into version numbers. The format tells you what changed, and range operators (`^`, `~`) let you automatically get safe updates while blocking breaking changes. Get this right and your dependencies stay secure and up-to-date. Get it wrong and your app breaks on `npm install`.

Format: `MAJOR.MINOR.PATCH` (e.g., `1.4.2`)

| Version          | Meaning               | Allows          |
| ---------------- | --------------------- | --------------- |
| `1.4.2`          | Exact version         | Only 1.4.2      |
| `^1.4.2`         | Compatible (caret)    | >=1.4.2 <2.0.0  |
| `~1.4.2`         | Approximately (tilde) | >=1.4.2 <1.5.0  |
| `>=1.4.2 <2.0.0` | Range                 | 1.4.2 to <2.0.0 |
| `*` or `latest`  | Latest                | Any version     |

### Interview Question: ^ vs ~

```bash
# Current version: 1.4.2
^1.4.2  →  Can install 1.4.3, 1.5.0, 1.99.0  (NOT 2.0.0)
~1.4.2  →  Can install 1.4.3, 1.4.99         (NOT 1.5.0)

# Current version: 0.4.2 (pre-1.0)
^0.4.2  →  Can install 0.4.3, 0.4.99         (NOT 0.5.0) ⚠️ Different behavior!
~0.4.2  →  Can install 0.4.3, 0.4.99         (NOT 0.5.0)
```

**Gotcha**: For `0.x.y` versions, `^` behaves like `~` because any minor change can be breaking.

### Lock Files

"Works on my machine" isn't just about code - it's about dependencies too. Your package.json says `"lodash": "^4.17.0"`, which could resolve to 4.17.15 on your machine today and 4.17.21 on CI tomorrow. If 4.17.21 has a bug, your builds suddenly fail and nobody knows why.

Lock files freeze the ENTIRE dependency tree - every package, every transitive dependency, down to the exact version and download URL. This means your teammate, your CI server, and production all get the exact same dependency versions you tested with. No surprises, no "but it worked yesterday."

| File                | Manager | Purpose                     |
| ------------------- | ------- | --------------------------- |
| `package-lock.json` | npm     | Locks exact dependency tree |
| `yarn.lock`         | yarn    | Locks exact dependency tree |
| `pnpm-lock.yaml`    | pnpm    | Locks exact dependency tree |

**Critical**: Always commit lock files. They ensure reproducible builds.

## Entry Points for Publishing

Here's the nightmare: Node.js uses CommonJS. Browsers use ESM. Bundlers want ESM for tree-shaking. TypeScript needs `.d.ts` files. Your library needs to work everywhere.

The old solution was to pick one format and let tools figure it out. The modern solution is to provide EVERYTHING - CommonJS for old Node, ESM for modern Node and bundlers, and TypeScript types - and let the environment pick what it needs. The `exports` field is the control panel that makes this work, telling each consumer exactly which file to load based on how they're importing your package.

```json
{
  "name": "my-library",
  "version": "1.0.0",

  // Legacy Node.js (CommonJS)
  "main": "./dist/index.js",

  // Legacy bundlers (ESM)
  "module": "./dist/index.mjs",

  // TypeScript types
  "types": "./dist/index.d.ts",

  // Modern - most important for Node 16+
  "exports": {
    ".": {
      "types": "./dist/index.d.ts",
      "import": "./dist/index.mjs", // ESM
      "require": "./dist/index.js", // CommonJS
      "default": "./dist/index.js"
    },
    "./package.json": "./package.json",
    "./utils": {
      "types": "./dist/utils.d.ts",
      "import": "./dist/utils.mjs",
      "require": "./dist/utils.js"
    }
  }
}
```

### Export Conditions Order Matters

Tools read the `exports` object from top to bottom and use the first condition they understand. Put `types` first so TypeScript finds your type definitions before trying to infer types from JavaScript. Put `import` before `require` to give ESM precedence. The `default` at the end catches everything else. Wrong order means TypeScript might skip your types or Node might load the wrong module format.

```json
{
  "exports": {
    ".": {
      "types": "./dist/index.d.ts", // ✓ Types first (for TypeScript)
      "import": "./dist/index.mjs", // ✓ ESM
      "require": "./dist/index.js", // ✓ CJS
      "default": "./dist/index.js" // ✓ Fallback
    }
  }
}
```

### Subpath Exports

Before `exports`, users could reach into your package and import ANY file: `import secret from 'my-lib/src/internal/secret.js'`. If you refactor and move that file, their code breaks. If it was never meant to be public, tough luck - it's now part of your public API forever.

Subpath exports give you control. You explicitly list what's importable. Everything else is blocked. This lets you refactor internals freely, prevent users from depending on implementation details, and design a clean public API. It's encapsulation for npm packages.

```json
{
  "exports": {
    ".": "./dist/index.js",
    "./utils": "./dist/utils.js",
    "./internal/*": null // ❌ Block access to internals
  }
}
```

```typescript
// Users can import:
import lib from "my-library"; // ✓ Works
import utils from "my-library/utils"; // ✓ Works
import foo from "my-library/internal/foo"; // ❌ Error!
```

## Publishing-Critical Fields

```json
{
  "name": "@scope/package-name", // Scoped package
  "version": "1.0.0",
  "description": "Shows in npm search",
  "keywords": ["searchable", "terms"],
  "author": "Your Name <email@example.com>",
  "license": "MIT",
  "repository": {
    "type": "git",
    "url": "https://github.com/user/repo.git"
  },
  "homepage": "https://github.com/user/repo#readme",
  "bugs": {
    "url": "https://github.com/user/repo/issues"
  },

  // What gets published
  "files": ["dist", "README.md", "LICENSE"],

  // Prevent accidental publishing
  "private": true, // Set to false when ready to publish

  // npm version compatibility
  "engines": {
    "node": ">=16.0.0",
    "npm": ">=8.0.0"
  },

  // Type of module system
  "type": "module" // or "commonjs" (default)
}
```

### Files Field

```json
{
  "files": ["dist"] // Only publishes dist/ directory
}
```

**Always included** (can't exclude):

- package.json
- README
- LICENSE
- CHANGELOG

**Always excluded** (can't include):

- node_modules
- .git

**Use `.npmignore`** to exclude additional files (similar to .gitignore).

## Scripts & Lifecycle Hooks

You're about to publish your package. Did you build it? Did you run tests? Did you update the changelog? Relying on memory for this checklist is how broken packages get published.

Lifecycle hooks automate the checklist. `prepublishOnly` runs your build and tests before npm lets you publish. `prepare` ensures git dependencies get compiled after install. You codify the workflow once, and npm enforces it forever. No more "oops, forgot to build" moments after publishing.

### Common Scripts

```json
{
  "scripts": {
    "build": "tsc",
    "test": "jest",
    "lint": "eslint src",
    "format": "prettier --write src",

    // Composite scripts
    "clean": "rm -rf dist",
    "prebuild": "npm run clean", // Runs BEFORE build
    "postbuild": "npm run test", // Runs AFTER build

    // Development
    "dev": "tsc --watch",
    "start": "node dist/index.js"
  }
}
```

### npm Lifecycle Hooks

Automatically run in this order:

```
npm install:
  preinstall → install → postinstall

npm publish:
  prepublishOnly → prepare → prepublish (deprecated) → publish → postpublish
```

### Publishing Workflow

The difference between `prepublishOnly` and `prepare` trips up even experienced developers. Here's the scenario: someone installs your package directly from GitHub (not npm). If you only use `prepublishOnly`, the TypeScript source never gets compiled - they get raw `.ts` files and your package breaks.

`prepare` runs on BOTH `npm publish` AND `npm install` from git, ensuring compilation happens whenever needed. `prepublishOnly` runs ONLY on publish, making it perfect for checks that should block publishing (like tests) but don't need to run for git installs.

```json
{
  "scripts": {
    "clean": "rm -rf dist",
    "build": "tsc",
    "test": "jest",

    // Runs before npm publish (one-time publish only)
    "prepublishOnly": "npm run build && npm test",

    // Runs on both npm install and npm publish
    "prepare": "npm run build"
  }
}
```

**Key difference**:

- `prepublishOnly`: Only before `npm publish`
- `prepare`: Before `npm publish` AND after `npm install` (useful for git dependencies)

## Package Types

Node.js has a problem: `.js` files are ambiguous. Is `export const foo = 'bar'` in `index.js` valid? Depends on whether you're using ESM or CommonJS. For years, Node assumed CommonJS. Now with ESM support, it needs to know which syntax to expect.

The `type` field in package.json is the signal. `"type": "module"` means "treat .js files as ESM". `"type": "commonjs"` (or omitting it) means "treat .js as CommonJS". No more ambiguity. Extensions like `.mjs` and `.cjs` override this on a per-file basis when you need to mix both in one project.

```json
{
  "type": "module" // Default: "commonjs"
}
```

| Type         | .js files are | .mjs | .cjs     |
| ------------ | ------------- | ---- | -------- |
| `"module"`   | ESM           | ESM  | CommonJS |
| `"commonjs"` | CommonJS      | ESM  | CommonJS |

```javascript
// package.json: { "type": "module" }
// index.js
export const foo = "bar"; // ✓ Valid ESM

// package.json: { "type": "commonjs" } or missing
// index.js
module.exports = { foo: "bar" }; // ✓ Valid CJS
```

## Hands-On Exercise 1: Dependency Audit

Analyze a package.json and identify issues:

```json
{
  "dependencies": {
    "react": "^18.0.0",
    "typescript": "^5.0.0",
    "axios": "^1.0.0"
  },
  "devDependencies": {
    "@types/react": "^18.0.0",
    "jest": "^29.0.0",
    "dotenv": "^16.0.0"
  }
}
```

<details>
<summary>Solution</summary>

**Issues**:

1. ❌ `typescript` should be in `devDependencies` (build tool)
2. ⚠️ `dotenv` might be needed in `dependencies` if loading env vars at runtime
3. ⚠️ If this is a library (not an app), `react` should be a `peerDependency`

**Fixed**:

```json
{
  "dependencies": {
    "axios": "^1.0.0",
    "dotenv": "^16.0.0"
  },
  "devDependencies": {
    "typescript": "^5.0.0",
    "@types/react": "^18.0.0",
    "jest": "^29.0.0",
    "react": "^18.0.0"
  },
  "peerDependencies": {
    "react": "^18.0.0"
  }
}
```

</details>

## Hands-On Exercise 2: Configure Dual Package

Create package.json for a library that supports both ESM and CommonJS.

**Requirements**:

- Build outputs to `dist/`
- Support TypeScript
- Provide both ESM (.mjs) and CJS (.js) builds

<details>
<summary>Solution</summary>

```json
{
  "name": "my-dual-package",
  "version": "1.0.0",
  "type": "module",
  "main": "./dist/index.js",
  "module": "./dist/index.mjs",
  "types": "./dist/index.d.ts",
  "exports": {
    ".": {
      "types": "./dist/index.d.ts",
      "import": "./dist/index.mjs",
      "require": "./dist/index.js",
      "default": "./dist/index.js"
    }
  },
  "files": ["dist"],
  "scripts": {
    "build": "tsc && tsc --module esnext --outDir dist/esm && mv dist/esm/index.js dist/index.mjs",
    "prepublishOnly": "npm run build"
  },
  "devDependencies": {
    "typescript": "^5.0.0"
  }
}
```

</details>

## Interview Questions

### Q1: What's the difference between ^ and ~ in package.json?

<details>
<summary>Answer</summary>

- `^1.4.2`: Allows updates that don't change the leftmost non-zero digit
  - Allows: 1.4.3, 1.5.0, 1.99.0
  - Blocks: 2.0.0
- `~1.4.2`: Allows patch updates only
  - Allows: 1.4.3, 1.4.99
  - Blocks: 1.5.0

**Exception**: For 0.x.y, `^0.4.2` behaves like `~0.4.2` (only allows patch updates) because pre-1.0 versions can have breaking changes in minor versions.

</details>

### Q2: Why use peerDependencies instead of dependencies?

Interviewers ask this to see if you understand package architecture and have published libraries (not just apps). It reveals whether you know the difference between building something standalone versus building something that plugs into an ecosystem. The answer shows you've dealt with the pain of dependency conflicts in real projects.

<details>
<summary>Answer</summary>

Prevents version conflicts for plugins/extensions. Example:

If your React component library used regular dependencies:

```
App depends on React 18.0.0
Your library bundles React 18.2.0
→ Two React copies in bundle! (Breaks React)
```

With peerDependencies:

```
App depends on React 18.0.0
Your library requires React ^18.0.0 (peer)
→ Single React copy shared
```

</details>

### Q3: What's the purpose of the "exports" field?

This question tests whether you're keeping up with modern npm practices. The `exports` field is relatively new (Node 12+) and many developers still use the old `main` field. Knowing `exports` signals you understand dual-package publishing, encapsulation, and the ESM transition - all critical for publishing production packages.

<details>
<summary>Answer</summary>

1. **Encapsulation**: Block access to internal modules
2. **Multiple entry points**: Different files for different imports
3. **Conditional exports**: Serve different files for ESM vs CommonJS
4. **Future-proof**: Standard way to define package structure

```json
{
  "exports": {
    ".": "./index.js", // import 'pkg'
    "./utils": "./utils.js", // import 'pkg/utils'
    "./internal/*": null // Block 'pkg/internal/*'
  }
}
```

</details>

### Q4: When does "prepare" script run vs "prepublishOnly"?

This catches developers who've only worked on apps (where build happens locally) versus publishing libraries (where builds need to happen for consumers too). It's a subtle distinction that causes real bugs when packages don't build correctly from git dependencies. Shows you've debugged "package works from npm but breaks from git" issues.

<details>
<summary>Answer</summary>

- `prepare`: Runs on:
  - `npm install` (after installing from git)
  - `npm publish`
  - Local `npm install`

- `prepublishOnly`: Runs only on:
  - `npm publish`

**Use case**:

- `prepublishOnly`: Build + test before publishing
- `prepare`: Build (for git dependencies that need compilation)

</details>

## Key Takeaways

1. **Dependencies**: Runtime vs dev vs peer - know the difference
2. **Semver**: `^` for minor updates, `~` for patches, exact for critical deps
3. **Exports**: Modern way to define package entry points (supports dual ESM/CJS)
4. **Files**: Control what gets published to avoid bloat
5. **Lock files**: Always commit them for reproducible builds
6. **Scripts**: Use lifecycle hooks (prepublishOnly, prepare) for automation
7. **Type field**: Controls .js file interpretation (module vs commonjs)

## Next Steps

In [Lesson 02: tsconfig.json Mastery](lesson-02-tsconfig-mastery.md), you'll learn:

- Critical compiler options and their implications
- Module resolution strategies
- Project references for monorepos
- Performance optimization
