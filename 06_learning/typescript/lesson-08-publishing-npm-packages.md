# Lesson 08: Publishing npm Packages

Complete guide to building, versioning, and publishing TypeScript packages to npm.

## Package Setup

Publishing to npm isn't just "run `npm publish`." You need to configure package.json with the right entry points (ESM, CJS, types), specify which files to include (dist, not src), set up build scripts that generate multiple module formats, and handle versioning correctly. Get any of these wrong and your package either won't install, won't import correctly, or will break users' builds. This section covers the complete setup for a professional npm package.

### Project Structure

The physical layout of a publishable package. Source code lives in `src/`, compiled output goes to `dist/`. You publish `dist/` to npm (not `src/`), so users get JavaScript, not TypeScript. Include README (first thing users see), LICENSE (legal protection), and .npmignore (control what gets published). Clean separation between development files and distribution files prevents bloating package size.

```
my-package/
  src/
    index.ts
    utils.ts
  dist/           ← Generated
    index.js
    index.mjs
    index.d.ts
  package.json
  tsconfig.json
  .npmignore
  README.md
  LICENSE
```

### package.json Configuration

The most important file for package distribution. Every field matters: `name` and `version` identify your package, `main`/`module`/`types` tell bundlers and Node where to find code, `exports` controls modern resolution, `files` whitelists what gets published. Missing or misconfigured fields cause "module not found" errors, type import failures, or publish 50MB of node_modules by accident. This is the complete config for a production package.

```json
{
  "name": "@scope/my-package",
  "version": "1.0.0",
  "description": "My awesome TypeScript package",
  "keywords": ["typescript", "utility"],
  "author": "Your Name <email@example.com>",
  "license": "MIT",
  "repository": {
    "type": "git",
    "url": "https://github.com/user/my-package.git"
  },
  "bugs": {
    "url": "https://github.com/user/my-package/issues"
  },
  "homepage": "https://github.com/user/my-package#readme",

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
    },
    "./package.json": "./package.json"
  },

  "files": ["dist", "README.md", "LICENSE"],

  "scripts": {
    "build": "npm run build:cjs && npm run build:esm && npm run build:types",
    "build:cjs": "tsc --module commonjs --outDir dist",
    "build:esm": "tsc --module esnext --outDir dist/esm && node scripts/rename-esm.js",
    "build:types": "tsc --declaration --emitDeclarationOnly --outDir dist",
    "clean": "rm -rf dist",
    "prepublishOnly": "npm run clean && npm run build && npm test",
    "test": "jest"
  },

  "devDependencies": {
    "@types/node": "^20.0.0",
    "typescript": "^5.0.0",
    "jest": "^29.0.0"
  }
}
```

### tsconfig.json

Compiler settings for library code differ from application code. Libraries target older ES versions (ES2020) for compatibility, enable strict mode for quality, generate declaration files (`declaration: true`) for types, and produce source maps for debugging. `skipLibCheck` speeds up compilation. These settings produce clean, compatible JavaScript that works in Node, browsers, and bundlers.

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "module": "ESNext",
    "lib": ["ES2020"],

    "declaration": true,
    "declarationMap": true,
    "sourceMap": true,

    "outDir": "./dist",
    "rootDir": "./src",

    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,

    "moduleResolution": "node"
  },
  "include": ["src/**/*"],
  "exclude": ["node_modules", "dist", "**/*.test.ts"]
}
```

## Dual Package Publishing (ESM + CJS)

The npm ecosystem is split: Node.js 16+ and modern bundlers prefer ESM, but millions of packages and older Node versions use CommonJS. Publishing only ESM alienates half your users. Publishing only CJS blocks tree-shaking and modern features. Dual publishing means compiling your TypeScript twice (once for ESM, once for CJS) and configuring package.json `exports` so each runtime gets the right format. This is the standard for serious npm packages.

### Approach 1: Separate Output Dirs

Organize build output by module format: `dist/esm/` for ES modules, `dist/cjs/` for CommonJS, `dist/types/` for declarations. Clear separation, easy to debug, widely used. The `exports` field maps import conditions (`import` for ESM, `require` for CJS) to the right directory. TypeScript compiles three times with different `--module` and `--outDir` flags.

```json
{
  "exports": {
    ".": {
      "import": "./dist/esm/index.js",
      "require": "./dist/cjs/index.js",
      "types": "./dist/types/index.d.ts"
    }
  }
}
```

**Build scripts**:

```bash
tsc --module commonjs --outDir dist/cjs
tsc --module esnext --outDir dist/esm
tsc --declaration --emitDeclarationOnly --outDir dist/types
```

### Approach 2: File Extension (.mjs)

Node.js uses file extensions to determine module type: `.mjs` is ESM, `.js` is CommonJS (or ESM if `"type": "module"`). Compile both formats to the same directory, then rename ESM files from `.js` to `.mjs`. Simpler file structure than separate directories. Requires a post-build script to rename files. Used by packages like `chalk` and `node-fetch`.

```json
{
  "exports": {
    ".": {
      "import": "./dist/index.mjs",
      "require": "./dist/index.js",
      "types": "./dist/index.d.ts"
    }
  }
}
```

**Build script** (rename-esm.js):

```javascript
import fs from "fs";
import path from "path";

const esmDir = path.join(process.cwd(), "dist/esm");
const files = fs.readdirSync(esmDir);

for (const file of files) {
  if (file.endsWith(".js")) {
    const oldPath = path.join(esmDir, file);
    const newPath = path.join(
      process.cwd(),
      "dist",
      file.replace(".js", ".mjs"),
    );
    fs.renameSync(oldPath, newPath);
  }
}
```

### Approach 3: Use Build Tools

Manual dual compilation is error-prone and tedious. Build tools like `tsup` (popular, fast) and `@microsoft/api-extractor` (for API docs) automate the entire process: they compile TypeScript to ESM and CJS, generate declaration files, bundle dependencies if needed, and produce source maps—all with one command. Modern packages use these tools instead of raw `tsc` commands.

**Using tsup**:

```bash
npm install -D tsup
```

```javascript
// tsup.config.ts
import { defineConfig } from "tsup";

export default defineConfig({
  entry: ["src/index.ts"],
  format: ["cjs", "esm"],
  dts: true,
  splitting: false,
  sourcemap: true,
  clean: true,
});
```

```json
{
  "scripts": {
    "build": "tsup"
  }
}
```

**Using @microsoft/api-extractor** (for API documentation):

```bash
npm install -D @microsoft/api-extractor
```

## Version Management

Version numbers aren't arbitrary—they communicate compatibility. Semver (semantic versioning) is a contract: `MAJOR.MINOR.PATCH` tells users "patch is safe, minor adds features, major breaks things." npm automates version bumps with `npm version patch/minor/major`, which updates package.json, creates git tags, and runs lifecycle hooks. Getting versioning wrong breaks user trust and causes dependency hell.

### Semantic Versioning (semver)

The language of compatibility. `1.2.3` means major version 1 (breaking changes reset this), minor version 2 (new features), patch 3 (bug fixes). Users can safely upgrade patches (1.2.3 → 1.2.4). Minor upgrades (1.2.0 → 1.3.0) add features without breaking. Major upgrades (1.x → 2.0.0) require code changes. `^1.2.3` in package.json means "allow minor and patch updates." This contract enables npm's entire dependency resolution system.

```
MAJOR.MINOR.PATCH

1.2.3
│ │ └─ Patch: Bug fixes
│ └─── Minor: New features (backward compatible)
└───── Major: Breaking changes
```

### npm version Command

Automates version bumps and git workflow. `npm version patch` increments patch (1.0.0 → 1.0.1), updates package.json, commits the change, and creates a git tag (`v1.0.1`). No manual editing, no typos, automatic git integration. The `-m` flag customizes the commit message. Essential for clean release history.

```bash
# Patch (1.0.0 → 1.0.1)
npm version patch

# Minor (1.0.0 → 1.1.0)
npm version minor

# Major (1.0.0 → 2.0.0)
npm version major

# Prerelease (1.0.0 → 1.0.1-0)
npm version prerelease

# Specific version
npm version 2.0.0

# Create git tag
npm version patch -m "Release %s"
```

### Version Hooks

Lifecycle hooks let you automate the entire release workflow. `preversion` runs tests (don't release broken code), `version` builds and stages artifacts, `postversion` pushes to git. One command (`npm version patch`) triggers tests, builds, commits, tags, and pushes. This prevents human error—you can't forget to build, can't skip tests, can't forget to push tags.

```json
{
  "scripts": {
    "preversion": "npm test",
    "version": "npm run build && git add -A dist",
    "postversion": "git push && git push --tags"
  }
}
```

**Workflow**:

1. `npm version patch`
2. Runs `preversion` (tests)
3. Updates version in package.json
4. Runs `version` (builds, stages files)
5. Creates git commit and tag
6. Runs `postversion` (pushes to git)

### Pre-release Versions

Beta/alpha releases for testing before stable. `1.0.1-alpha.0` signals "not production ready." Users opt-in with `npm install my-package@alpha`. You can iterate (`alpha.1`, `alpha.2`) without affecting stable installs. RC (release candidate) is the final pre-release before stable. Essential for safely testing major changes with early adopters.

```bash
# Alpha
npm version prerelease --preid=alpha
# 1.0.0 → 1.0.1-alpha.0

# Beta
npm version prerelease --preid=beta
# 1.0.0 → 1.0.1-beta.0

# RC (Release Candidate)
npm version prerelease --preid=rc
# 1.0.0 → 1.0.1-rc.0
```

### Publishing Pre-releases

Pre-releases use npm tags (not git tags). Publishing with `--tag beta` makes `npm install my-package` still install the stable version, but `npm install my-package@beta` gets the pre-release. This prevents accidental pre-release installs and lets you test with willing users before promoting to `latest`.

```bash
# Publish with tag (not 'latest')
npm publish --tag beta

# Install specific tag
npm install my-package@beta
```

## Publishing Workflow

The actual mechanics of getting your package to npm. You need npm credentials (`npm login`), must test the package contents (`npm pack --dry-run`), should test locally before publishing (`npm pack` and install the tarball), and must handle scoped packages correctly (public vs private). The first publish requires `--access public` for scoped packages. This is the checklist that prevents embarrassing mistakes.

### Initial Setup

You can't publish without authentication. `npm login` prompts for username, password, and email, then stores credentials locally. Scoped packages (`@myorg/package`) can optionally use `--scope` to set a default scope. You only do this once per machine—credentials persist.

```bash
# Login to npm
npm login

# For scoped packages (@scope/package)
npm login --scope=@scope
```

### First Publish

The first time publishing a package. Public packages are default for unscoped names (`my-package`), but scoped packages (`@myorg/my-package`) default to private (requires paid plan). Use `--access public` to make scoped packages public. `npm pack --dry-run` shows what files would be published without actually creating a tarball—critical safety check.

```bash
# Public package
npm publish

# Scoped public package (default is private)
npm publish --access public

# Check what would be published
npm pack --dry-run
```

### .npmignore

Like .gitignore but for npm. Controls what files get published. Common exclusions: `src/` (users don't need TypeScript source), `tests/`, config files (tsconfig.json, .eslintrc), CI configs (.github/). If .npmignore doesn't exist, npm uses .gitignore. Always include: `dist/`, README.md, LICENSE. Forgetting this publishes everything (node_modules, .env files, etc.)—huge security and size risk.

Controls what gets published (similar to .gitignore).

```
# .npmignore
src/
tests/
*.test.ts
.github/
.vscode/
tsconfig.json
jest.config.js
.eslintrc.js
```

**Note**: If .npmignore doesn't exist, npm uses .gitignore.

**Always include**: dist/, README.md, LICENSE

### Publishing Checklist

The safety net against broken publishes. Test first (don't ship bugs), build fresh (don't publish stale code), check contents (`npm pack --dry-run`), test locally (install the tarball in a scratch project, verify imports and types work), then bump version and publish. Following this checklist prevents 90% of "I published a broken version" incidents.

```bash
# 1. Run tests
npm test

# 2. Build
npm run build

# 3. Check package contents
npm pack --dry-run

# 4. Test in local project
npm pack
# Creates my-package-1.0.0.tgz
cd ../test-project
npm install ../my-package/my-package-1.0.0.tgz

# 5. Check types work
# In test project, import and verify types

# 6. Bump version
npm version patch

# 7. Publish
npm publish

# 8. Verify on npm
npm view my-package
```

## Automated Publishing

Manual publishing is error-prone: you forget to test, skip the build, publish from the wrong branch. CI/CD automation ensures every release follows the same process: checkout clean code, install dependencies, run tests, build, publish. GitHub Actions triggers on git releases. Semantic Release automates versioning based on commit messages. These tools make releases boring (which is good).

### GitHub Actions

Trigger npm publish when you create a GitHub release. The workflow checks out code, installs dependencies, runs tests, builds, and publishes—all automatically. Secrets (npm token) are stored in GitHub settings. Zero manual steps, consistent process, audit trail. Most modern packages use this workflow.

```yaml
# .github/workflows/publish.yml
name: Publish to npm

on:
  release:
    types: [created]

jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: "18"
          registry-url: "https://registry.npmjs.org"

      - name: Install dependencies
        run: npm ci

      - name: Run tests
        run: npm test

      - name: Build
        run: npm run build

      - name: Publish
        run: npm publish --access public
        env:
          NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}
```

### Semantic Release

Fully automated versioning and publishing based on commit message conventions. `feat:` triggers minor release, `fix:` triggers patch, `BREAKING CHANGE:` triggers major. No manual `npm version` calls—just commit with the right message and semantic-release handles version bump, changelog generation, git tag, and npm publish. Used by large projects (Babel, Jest) to eliminate human error in releases.

Automates versioning and publishing based on commit messages.

```bash
npm install -D semantic-release @semantic-release/git @semantic-release/changelog
```

```javascript
// .releaserc.js
module.exports = {
  branches: ["main"],
  plugins: [
    "@semantic-release/commit-analyzer",
    "@semantic-release/release-notes-generator",
    "@semantic-release/changelog",
    "@semantic-release/npm",
    "@semantic-release/git",
    "@semantic-release/github",
  ],
};
```

**Commit conventions**:

```bash
feat: add new feature      # → Minor release
fix: resolve bug           # → Patch release
BREAKING CHANGE: ...       # → Major release
```

## Package Distribution

Once published to npm, your package is instantly available on CDNs (unpkg, jsDelivr) for browser usage, can be installed from private registries (GitHub Packages, Artifactory), and supports scoped namespacing (`@myorg/package`). Distribution isn't just npm—it's the entire ecosystem of mirrors, CDNs, and private registries. Understanding these options is essential for both open source and enterprise packages.

### CDN Distribution

Packages automatically available on CDNs for browser usage—no build step needed. unpkg and jsDelivr serve files directly from npm. Users can `<script src="https://unpkg.com/my-package">` in HTML. Useful for demos, documentation, and quick prototyping. CDNs handle caching, compression, and global distribution automatically.

Packages automatically available on CDNs:

**unpkg**:

```html
<script src="https://unpkg.com/my-package@1.0.0/dist/index.js"></script>
```

**jsDelivr**:

```html
<script src="https://cdn.jsdelivr.net/npm/my-package@1.0.0/dist/index.js"></script>
```

### Private Packages

Enterprise packages that shouldn't be public. Publish to private registries (GitHub Packages, npm private, Azure Artifacts, Artifactory). Set `publishConfig.registry` in package.json to override the default npm registry. Requires authentication tokens. Used for proprietary code, internal libraries, and company-specific packages.

```bash
# Publish to private registry
npm publish --registry https://npm.pkg.github.com

# Or configure in package.json
{
  "publishConfig": {
    "registry": "https://npm.pkg.github.com"
  }
}
```

### Scoped Packages

Namespacing for package names. `@myorg/package` prevents naming collisions, groups related packages, and signals ownership. Scoped packages default to private (requires paid plan) unless you use `--access public`. Organizations use scopes for branding (@babel, @angular, @types). Individuals use scopes to avoid name squatting.

```json
{
  "name": "@myorg/my-package",
  "publishConfig": {
    "access": "public" // or "restricted" for private
  }
}
```

## Package Maintenance

After publishing, packages need ongoing care. Deprecate old versions with bugs (`npm deprecate`), unpublish recent mistakes (72-hour window), transfer ownership (`npm owner`). Deprecation is the safe way to warn users away from broken versions without breaking existing installs. Unpublishing is nuclear—only for critical security or legal issues. Good maintenance keeps the ecosystem healthy.

### Deprecating Versions

Mark a version as problematic without breaking existing installs. Users see a warning during install but can still use the deprecated version if needed. Used for versions with critical bugs, security issues, or that were published by mistake. The warning message guides users to the fix (e.g., "use 1.0.1+ instead").

```bash
# Deprecate specific version
npm deprecate my-package@1.0.0 "Critical bug, use 1.0.1+"

# Deprecate all versions
npm deprecate my-package "Package no longer maintained"
```

### Unpublishing

The nuclear option. You can only unpublish within 72 hours of publishing (to prevent breaking the ecosystem). After 72 hours, use deprecation instead. Unpublishing a widely-used package causes massive breakage (see left-pad incident). Only unpublish for critical security issues, accidentally published secrets, or legal problems.

```bash
# Can only unpublish within 72 hours
npm unpublish my-package@1.0.0

# Unpublish entire package (not recommended)
npm unpublish my-package --force
```

### Package Transfer

Manage ownership of published packages. Add collaborators (`npm owner add`) to share publishing rights, remove collaborators who leave the project, list current owners. Essential for open source projects with multiple maintainers or when transferring abandoned packages to new stewards.

```bash
# Add collaborator
npm owner add <username> my-package

# Remove collaborator
npm owner rm <username> my-package

# List owners
npm owner ls my-package
```

## Best Practices

The rules that prevent common mistakes. Use loose version constraints for dependencies (`^4.17.21` allows patches and minors), publish only necessary files (dist/, not src/), provide clear examples in README, follow semver strictly, and always test before publishing. These practices are learned from collective pain—follow them to avoid repeating others' mistakes.

### 1. Version Constraints for Dependencies

Loose constraints (`^4.17.21`) allow users to get bug fixes and features without you republishing. Tight constraints (`4.17.21`) lock users into outdated versions. peerDependencies should be especially loose (wide compatibility range) so your package works with multiple versions of React, Vue, etc. This maximizes compatibility across the ecosystem.

```json
{
  "dependencies": {
    "lodash": "^4.17.21" // ✓ Allow minor/patch
  },
  "devDependencies": {
    "typescript": "^5.0.0" // ✓ Allow minor/patch
  },
  "peerDependencies": {
    "react": "^18.0.0" // ✓ Wide range for compatibility
  }
}
```

### 2. Include Essential Files Only

Publish dist/ (compiled code), README.md (documentation), LICENSE (legal). Exclude src/ (TypeScript source users don't need), tests/ (not needed at runtime), config files (tsconfig, eslint, etc.). Smaller packages install faster, use less disk space, and reduce attack surface. Use the `files` whitelist in package.json—it's safer than .npmignore blacklisting.

```json
{
  "files": ["dist", "README.md", "LICENSE"]
}
```

**Don't include**: src/, tests/, config files

### 3. Provide Examples

Your README is the first thing users see. Show installation (`npm install`), basic usage (import and call), and common patterns. Code examples should be copy-pasteable and complete. Good examples reduce support burden—users find answers in the README instead of filing issues.

```markdown
# README.md

## Installation

\`\`\`bash
npm install my-package
\`\`\`

## Usage

\`\`\`typescript
import { myFunction } from 'my-package';

const result = myFunction('input');
\`\`\`
```

### 4. Semantic Versioning

The contract with your users. Patch (1.0.0 → 1.0.1) means bug fixes only—safe to upgrade immediately. Minor (1.0.0 → 1.1.0) means new features, backward compatible—safe for `^` ranges. Major (1.0.0 → 2.0.0) means breaking changes—requires code changes. Breaking semver breaks trust and causes dependency hell.

- Patch: Bug fixes only
- Minor: New features, backward compatible
- Major: Breaking changes

### 5. Test Before Publishing

The pre-flight checklist. Build fresh (`npm run build`), run tests (`npm test`), check package contents (`npm pack --dry-run`). Catch errors before users do. "Works on my machine" isn't enough—test in a clean install to catch missing files or incorrect paths.

```bash
# Always run:
npm run build
npm test
npm pack --dry-run
```

## Hands-On Exercise: Publish a Package

Create and publish a simple utility package:

**Requirements**:

1. Create a package with one utility function
2. Support both ESM and CommonJS
3. Include TypeScript types
4. Write README with examples
5. Publish to npm (or test with `npm pack`)

<details>
<summary>Solution</summary>

```typescript
// src/index.ts
/**
 * Capitalizes the first letter of a string
 * @param str - Input string
 * @returns Capitalized string
 */
export function capitalize(str: string): string {
  if (!str) return str;
  return str.charAt(0).toUpperCase() + str.slice(1);
}

/**
 * Debounces a function
 * @param fn - Function to debounce
 * @param delay - Delay in milliseconds
 * @returns Debounced function
 */
export function debounce<T extends (...args: any[]) => any>(
  fn: T,
  delay: number,
): (...args: Parameters<T>) => void {
  let timeoutId: ReturnType<typeof setTimeout>;

  return function (...args: Parameters<T>) {
    clearTimeout(timeoutId);
    timeoutId = setTimeout(() => fn(...args), delay);
  };
}
```

```json
// package.json
{
  "name": "@yourusername/utils",
  "version": "1.0.0",
  "description": "Simple utility functions",
  "main": "./dist/index.js",
  "module": "./dist/index.mjs",
  "types": "./dist/index.d.ts",
  "exports": {
    ".": {
      "types": "./dist/index.d.ts",
      "import": "./dist/index.mjs",
      "require": "./dist/index.js"
    }
  },
  "files": ["dist"],
  "scripts": {
    "build": "tsc && tsc --module esnext --outDir dist/esm",
    "prepublishOnly": "npm run build"
  },
  "keywords": ["utils", "typescript"],
  "license": "MIT",
  "devDependencies": {
    "typescript": "^5.0.0"
  }
}
```

```markdown
# @yourusername/utils

Simple utility functions in TypeScript.

## Installation

\`\`\`bash
npm install @yourusername/utils
\`\`\`

## Usage

\`\`\`typescript
import { capitalize, debounce } from '@yourusername/utils';

console.log(capitalize('hello')); // 'Hello'

const search = debounce((query) => {
console.log('Searching:', query);
}, 300);
\`\`\`

## License

MIT
```

```bash
# Build and test locally
npm run build
npm pack

# Test in another project
cd ../test-project
npm install ../utils/yourusername-utils-1.0.0.tgz

# Publish
npm publish --access public
```

</details>

## Interview Questions

### Q1: How to publish dual ESM/CJS package?

Tests understanding of modern package distribution and the `exports` field. Strong candidates explain both the package.json config and the build process. Weak answers might suggest "just use CommonJS" (outdated) or not understand the `exports` field.

<details>
<summary>Answer</summary>

**Method 1: Different extensions**

```json
{
  "main": "./dist/index.js", // CJS
  "module": "./dist/index.mjs", // ESM
  "types": "./dist/index.d.ts",
  "exports": {
    ".": {
      "import": "./dist/index.mjs",
      "require": "./dist/index.js",
      "types": "./dist/index.d.ts"
    }
  }
}
```

**Build**: Compile twice, rename ESM files to .mjs

**Method 2: Separate directories**

```json
{
  "exports": {
    ".": {
      "import": "./dist/esm/index.js",
      "require": "./dist/cjs/index.js",
      "types": "./dist/types/index.d.ts"
    }
  }
}
```

</details>

### Q2: What's the difference between dependencies and peerDependencies?

Reveals understanding of dependency management and package size. Strong answers explain when to use each and the problems they solve (duplication, version conflicts). Common in React library interviews—"should React be a dependency or peerDependency?"

<details>
<summary>Answer</summary>

**dependencies**: Installed with your package

```json
{
  "dependencies": {
    "lodash": "^4.17.21" // Always installed
  }
}
```

**peerDependencies**: User must install (not bundled)

```json
{
  "peerDependencies": {
    "react": "^18.0.0" // User's project must have React
  }
}
```

**Use peerDependencies** for:

- Plugins/extensions (React components, Babel plugins)
- Avoid version conflicts
- Prevent bundling multiple copies

</details>

### Q3: How does npm version work?

Tests understanding of versioning workflow and automation. Strong answers explain the lifecycle hooks (preversion, version, postversion) and how to automate the entire release process. Weak answers might think it just edits package.json.

<details>
<summary>Answer</summary>

```bash
npm version patch  # 1.0.0 → 1.0.1
```

**Process**:

1. Runs `preversion` script
2. Updates package.json version
3. Runs `version` script
4. Commits changes
5. Creates git tag
6. Runs `postversion` script

**Automation**:

```json
{
  "scripts": {
    "preversion": "npm test",
    "version": "npm run build && git add dist",
    "postversion": "git push && git push --tags && npm publish"
  }
}
```

</details>

### Q4: What files should be published?

Reveals practical experience with npm packaging. Strong answers mention the `files` field in package.json and explain why source files should be excluded. Weak answers might not realize `src/` gets published by default (huge security risk if it contains .env files).

<details>
<summary>Answer</summary>

**Include**:

- `dist/` (compiled code)
- `README.md`
- `LICENSE`
- `CHANGELOG.md`

**Exclude** (via .npmignore):

- `src/` (source TypeScript)
- `tests/`
- Config files (tsconfig.json, .eslintrc, etc.)
- `.github/`, `.vscode/`

```json
{
  "files": ["dist", "README.md", "LICENSE"]
}
```

Check before publishing:

```bash
npm pack --dry-run
```

</details>

## Key Takeaways

1. **Dual Publishing**: Support both ESM and CJS with exports field
2. **Types**: Always include .d.ts files (declaration: true)
3. **Versioning**: Follow semver, use npm version commands
4. **Files**: Only publish dist/, exclude source and config
5. **Testing**: Test locally with npm pack before publishing
6. **Automation**: Use GitHub Actions or semantic-release
7. **Documentation**: Clear README with installation and usage examples

## Summary

You now have a comprehensive understanding of:

- Package configuration for publishing
- Dual ESM/CJS distribution
- Versioning and semver
- Publishing workflow and automation
- Package maintenance and best practices

This completes the TypeScript interview prep series. You're now equipped with the knowledge to confidently discuss TypeScript ecosystem, tooling, type system, and publishing practices in interviews.
