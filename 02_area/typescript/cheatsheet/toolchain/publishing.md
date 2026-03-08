# Publishing npm Packages

## Quick Reference

| Task             | Command / convention                     |
| ---------------- | ---------------------------------------- |
| Publish          | `npm publish`                            |
| Scoped package   | `npm publish --access public`            |
| Bump version     | `npm version patch/minor/major`          |
| Preview contents | `npm pack --dry-run`                     |
| Deprecate        | `npm deprecate pkg@version "message"`    |
| Unpublish        | `npm unpublish pkg@version` (within 72h) |

## Pre-Publish Setup

### 1. package.json for a library

```json
{
  "name": "@yourorg/mylib",
  "version": "1.0.0",
  "type": "module",
  "main": "dist/index.js",
  "types": "dist/index.d.ts",
  "exports": {
    ".": {
      "types": "./dist/index.d.ts",
      "import": "./dist/index.js"
    }
  },
  "files": ["dist"],
  "scripts": {
    "build": "tsc",
    "prepublishOnly": "npm run build && npm test"
  },
  "license": "MIT",
  "repository": {
    "type": "git",
    "url": "https://github.com/yourorg/mylib"
  }
}
```

### 2. files field — control what gets published

```json
{
  "files": ["dist", "README.md", "LICENSE"]
}
```

Always check before publishing:

```sh
npm pack --dry-run    # shows exactly what files will be in the tarball
```

### 3. .npmignore (alternative to files)

```
src/
tests/
.github/
*.test.ts
tsconfig.json
.env
coverage/
```

`files` in package.json is preferred over `.npmignore` — whitelist is safer than blacklist.

## Version Bumping

### 4. npm version

```sh
npm version patch     # 1.0.0 → 1.0.1
npm version minor     # 1.0.0 → 1.1.0
npm version major     # 1.0.0 → 2.0.0
npm version prerelease --preid=rc   # 1.0.0 → 1.0.1-rc.0
```

`npm version` updates package.json, creates a git commit, and tags it.

### 5. Semver rules

```
v1.2.3
│ │ │
│ │ └── Patch: bug fixes, no API changes
│ └──── Minor: new features, backwards compatible
└────── Major: breaking changes
```

## Publishing

### 6. First publish

```sh
# Login (one time)
npm login

# Public scoped package
npm publish --access public

# Unscoped package
npm publish
```

### 7. Subsequent releases

```sh
npm version patch     # bump version + git tag
git push && git push --tags
npm publish
```

### 8. Pre-release versions

```sh
npm version prerelease --preid=rc    # 1.0.1-rc.0
npm publish --tag rc                  # publish under "rc" tag

# Users install with:
# npm install @yourorg/mylib@rc
```

Pre-release versions are not installed by default — users must opt in.

## Automation

### 9. GitHub Actions — publish on tag

```yaml
name: Publish
on:
  push:
    tags: ["v*"]

jobs:
  publish:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      id-token: write
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          registry-url: "https://registry.npmjs.org"
      - run: npm ci
      - run: npm test
      - run: npm publish --provenance --access public
        env:
          NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}
```

`--provenance` adds supply chain attestation (npm 9.5+).

### 10. Release workflow

```sh
# 1. Make sure everything is clean
git status    # should be clean

# 2. Run checks
npm test
npm run lint
npm run build

# 3. Bump version
npm version minor    # creates commit + tag

# 4. Push
git push && git push --tags

# 5. Publish (or let CI do it)
npm publish
```

## Deprecation & Unpublish

### 11. Deprecate a version

```sh
npm deprecate @yourorg/mylib@1.2.0 "Security issue, upgrade to 1.2.1"
npm deprecate @yourorg/mylib@"<1.3.0" "Upgrade to >=1.3.0"
```

### 12. Unpublish (last resort)

```sh
# Only within 72 hours of publishing
npm unpublish @yourorg/mylib@1.2.0

# Entire package (very rare)
npm unpublish @yourorg/mylib --force
```

Prefer deprecation over unpublishing — unpublishing breaks downstream users.

## Private Registry

### 13. Publish to private registry

```ini
# .npmrc
@yourorg:registry=https://npm.pkg.github.com
//npm.pkg.github.com/:_authToken=${GITHUB_TOKEN}
```

```sh
npm publish    # publishes to the registry configured for your scope
```
