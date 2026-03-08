# npm Package Management

## Quick Reference

| Use case              | Command                  |
| --------------------- | ------------------------ |
| Init project          | `npm init -y`            |
| Install all deps      | `npm install` / `npm ci` |
| Add dependency        | `npm install pkg`        |
| Add dev dependency    | `npm install -D pkg`     |
| Remove dependency     | `npm uninstall pkg`      |
| Update single         | `npm update pkg`         |
| Update all            | `npm update`             |
| Check outdated        | `npm outdated`           |
| Run script            | `npm run script-name`    |
| List installed        | `npm list --depth=0`     |
| Why is dep installed  | `npm explain pkg`        |
| Audit vulnerabilities | `npm audit`              |
| Clean install (CI)    | `npm ci`                 |

## Init & Install

### 1. Initialize a project

```sh
npm init -y
```

### 2. Install dependencies

```sh
npm install express zod           # production deps
npm install -D typescript @types/node @types/express  # dev deps
npm install -D vitest eslint prettier
```

### 3. Version specifiers

```sh
npm install pkg@1.2.3             # exact version
npm install pkg@^1.2.3            # >=1.2.3 <2.0.0 (default)
npm install pkg@~1.2.3            # >=1.2.3 <1.3.0
npm install pkg@latest            # latest stable
npm install pkg@next              # next/preview tag
npm install pkg@github:user/repo  # from GitHub
```

### 4. npm install vs npm ci

```sh
npm install    # installs from package.json, may update lock file
npm ci         # installs from lock file exactly, faster, for CI
```

Always use `npm ci` in CI pipelines — it's deterministic and fails if lock file is out of sync.

## package.json

### 5. Essential fields

```json
{
  "name": "my-service",
  "version": "1.0.0",
  "type": "module",
  "main": "dist/index.js",
  "types": "dist/index.d.ts",
  "scripts": {
    "build": "tsc",
    "start": "node dist/server.js",
    "dev": "tsx watch src/server.ts",
    "test": "vitest run",
    "test:watch": "vitest",
    "lint": "eslint src/",
    "format": "prettier --write src/",
    "typecheck": "tsc --noEmit"
  },
  "engines": {
    "node": ">=20"
  }
}
```

### 6. type: module vs commonjs

| `"type"`               | `.js` files are | `.cjs` files are | `.mjs` files are |
| ---------------------- | --------------- | ---------------- | ---------------- |
| `"module"`             | ESM             | CJS              | ESM              |
| `"commonjs"` (default) | CJS             | CJS              | ESM              |

### 7. exports field (libraries)

```json
{
  "exports": {
    ".": {
      "types": "./dist/index.d.ts",
      "import": "./dist/index.js",
      "require": "./dist/index.cjs"
    },
    "./utils": {
      "types": "./dist/utils.d.ts",
      "import": "./dist/utils.js"
    }
  }
}
```

## Scripts

### 8. npm scripts

```sh
npm run build         # run named script
npm test              # shorthand for npm run test
npm start             # shorthand for npm run start
npm run lint -- --fix # pass args to script (note the --)
```

### 9. Pre/post hooks

```json
{
  "scripts": {
    "prebuild": "rm -rf dist",
    "build": "tsc",
    "postbuild": "cp package.json dist/",
    "pretest": "npm run build"
  }
}
```

### 10. Run multiple scripts

```sh
npm run lint && npm run test        # sequential, stop on fail
npm run lint & npm run test         # parallel (shell)

# Or use npm-run-all / concurrently
npx concurrently "npm:lint" "npm:test"
```

## package-lock.json

### 11. Lock file rules

- **Always commit** `package-lock.json` to version control
- Never edit it manually
- `npm ci` reads it exactly — ensures reproducible builds
- If lock file conflicts: delete it, run `npm install`, commit the new one

## Update & Audit

### 12. Check for updates

```sh
npm outdated                      # show current vs wanted vs latest
npm update                        # update within semver range
npm install pkg@latest            # jump to latest (may be breaking)
```

### 13. Security audit

```sh
npm audit                         # check for known vulnerabilities
npm audit fix                     # auto-fix where possible
npm audit fix --force             # allow breaking changes (careful)
```

## Workspaces (Monorepo)

### 14. Configure workspaces

```json
// root package.json
{
  "workspaces": ["packages/*", "apps/*"]
}
```

```sh
npm install                       # install all workspace deps
npm run build -w packages/shared  # run script in specific workspace
npm run test --workspaces          # run script in all workspaces
npm install zod -w packages/api   # add dep to specific workspace
```

## .npmrc

### 15. Project-level npm config

```ini
# .npmrc
engine-strict=true                # fail if Node version doesn't match engines
save-exact=true                   # pin exact versions (no ^ prefix)
package-lock=true                 # always generate lock file
```

## Useful Commands

### 16. Inspect and debug

```sh
npm list --depth=0                # direct deps only
npm list --all                    # full tree
npm explain pkg                   # why is pkg installed
npm view pkg versions             # available versions
npm cache clean --force           # clear npm cache
npm config list                   # show all config
```
