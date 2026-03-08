# ESLint & Prettier

## Quick Reference

| Tool              | Purpose                                       |
| ----------------- | --------------------------------------------- |
| ESLint            | Finds bugs and enforces code patterns         |
| Prettier          | Formats code (spacing, quotes, semicolons)    |
| typescript-eslint | ESLint rules that understand TypeScript types |

Rule of thumb: Prettier handles formatting, ESLint handles logic and correctness. Don't fight over formatting in ESLint.

## ESLint (Flat Config — v9+)

### 1. Install

```sh
npm install -D eslint @eslint/js typescript-eslint
```

### 2. eslint.config.mjs

```javascript
import eslint from "@eslint/js";
import tseslint from "typescript-eslint";

export default tseslint.config(
  eslint.configs.recommended,
  ...tseslint.configs.recommended,
  {
    ignores: ["dist/", "node_modules/", "coverage/"],
  },
);
```

### 3. With stricter type-checked rules

```javascript
import eslint from "@eslint/js";
import tseslint from "typescript-eslint";

export default tseslint.config(
  eslint.configs.recommended,
  ...tseslint.configs.recommendedTypeChecked,
  {
    languageOptions: {
      parserOptions: {
        projectService: true,
        tsconfigRootDir: import.meta.dirname,
      },
    },
  },
  {
    ignores: ["dist/", "node_modules/", "coverage/", "*.config.*"],
  },
);
```

Type-checked rules catch more bugs (floating promises, unsafe any usage) but are slower.

### 4. Run ESLint

```sh
npx eslint src/                   # lint src directory
npx eslint src/ --fix             # auto-fix what it can
npx eslint src/ --max-warnings=0  # fail on warnings (good for CI)
```

### 5. Useful rules to enable

```javascript
{
  rules: {
    "@typescript-eslint/no-unused-vars": ["error", { argsIgnorePattern: "^_" }],
    "@typescript-eslint/no-floating-promises": "error",  // requires type-checked
    "@typescript-eslint/no-misused-promises": "error",   // requires type-checked
    "@typescript-eslint/await-thenable": "error",        // requires type-checked
    "no-console": "warn",
    "eqeqeq": "error",
    "no-throw-literal": "error",
  },
}
```

### 6. Disable for a line or block

```typescript
// eslint-disable-next-line @typescript-eslint/no-unused-vars
const _placeholder = true;

/* eslint-disable no-console */
console.log("debug");
console.log("more debug");
/* eslint-enable no-console */
```

## Prettier

### 7. Install

```sh
npm install -D prettier eslint-config-prettier
```

`eslint-config-prettier` disables ESLint rules that conflict with Prettier.

### 8. .prettierrc

```json
{
  "semi": true,
  "singleQuote": false,
  "trailingComma": "all",
  "printWidth": 100,
  "tabWidth": 2
}
```

### 9. .prettierignore

```
dist
node_modules
coverage
package-lock.json
```

### 10. Run Prettier

```sh
npx prettier --write src/         # format files
npx prettier --check src/         # check without writing (CI)
```

## Integrating ESLint + Prettier

### 11. Add prettier to ESLint config

```javascript
import eslint from "@eslint/js";
import tseslint from "typescript-eslint";
import prettierConfig from "eslint-config-prettier";

export default tseslint.config(
  eslint.configs.recommended,
  ...tseslint.configs.recommended,
  prettierConfig, // must be last — disables conflicting rules
  {
    ignores: ["dist/", "node_modules/", "coverage/"],
  },
);
```

## package.json Scripts

### 12. Common script setup

```json
{
  "scripts": {
    "lint": "eslint src/",
    "lint:fix": "eslint src/ --fix",
    "format": "prettier --write src/",
    "format:check": "prettier --check src/",
    "check": "npm run typecheck && npm run lint && npm run format:check",
    "typecheck": "tsc --noEmit"
  }
}
```

## CI Integration

### 13. GitHub Actions

```yaml
- name: Lint and format check
  run: |
    npm run typecheck
    npm run lint -- --max-warnings=0
    npm run format:check
```

### 14. Pre-commit hook (with lint-staged)

```sh
npm install -D lint-staged husky
npx husky init
echo "npx lint-staged" > .husky/pre-commit
```

```json
// package.json
{
  "lint-staged": {
    "*.{ts,tsx}": ["eslint --fix", "prettier --write"],
    "*.{json,md,yml}": ["prettier --write"]
  }
}
```
