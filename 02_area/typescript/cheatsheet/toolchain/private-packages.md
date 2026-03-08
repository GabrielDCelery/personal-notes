# Private npm Packages

## Quick Reference

| Registry         | Use when                         |
| ---------------- | -------------------------------- |
| GitHub Packages  | Already on GitHub, scoped to org |
| npm private      | Paying for npm Teams/Org         |
| Verdaccio        | Self-hosted, on-prem             |
| AWS CodeArtifact | AWS ecosystem                    |
| GitLab Registry  | Already on GitLab                |

## .npmrc Configuration

### 1. Scope to a private registry

```ini
# .npmrc (project root — commit this)
@yourorg:registry=https://npm.pkg.github.com
```

This routes all `@yourorg/*` packages to GitHub Packages. Everything else goes to the public npm registry.

### 2. Auth token (developer machines)

```ini
# ~/.npmrc (user home — never commit this)
//npm.pkg.github.com/:_authToken=ghp_xxxxxxxxxxxx
```

### 3. Multiple registries

```ini
# .npmrc
@yourorg:registry=https://npm.pkg.github.com
@otherog:registry=https://registry.npmjs.org

# Auth for each
//npm.pkg.github.com/:_authToken=${GITHUB_TOKEN}
//registry.npmjs.org/:_authToken=${NPM_TOKEN}
```

## GitHub Packages

### 4. Setup

```ini
# .npmrc
@yourorg:registry=https://npm.pkg.github.com
```

```json
// package.json of the private package
{
  "name": "@yourorg/shared-utils",
  "version": "1.0.0",
  "publishConfig": {
    "registry": "https://npm.pkg.github.com"
  }
}
```

### 5. Auth for developers

```sh
# Login once
npm login --registry=https://npm.pkg.github.com
# Username: your-github-username
# Password: ghp_xxxxxxxxxxxx (classic PAT with read:packages, write:packages)
```

### 6. Publish

```sh
npm publish
```

## CI Pipeline Setup

### 7. GitHub Actions — install private deps

```yaml
- uses: actions/setup-node@v4
  with:
    node-version: 20
    registry-url: "https://npm.pkg.github.com"
    scope: "@yourorg"

- run: npm ci
  env:
    NODE_AUTH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

`GITHUB_TOKEN` is automatically available in GitHub Actions and has read access to packages in the same org.

### 8. GitHub Actions — publish private package

```yaml
- uses: actions/setup-node@v4
  with:
    node-version: 20
    registry-url: "https://npm.pkg.github.com"
    scope: "@yourorg"

- run: npm ci
- run: npm publish
  env:
    NODE_AUTH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### 9. GitLab CI

```yaml
before_script:
  - echo "@yourorg:registry=https://${CI_SERVER_HOST}/api/v4/projects/${CI_PROJECT_ID}/packages/npm/" >> .npmrc
  - echo "//${CI_SERVER_HOST}/api/v4/projects/${CI_PROJECT_ID}/packages/npm/:_authToken=${CI_JOB_TOKEN}" >> .npmrc
  - npm ci
```

## Docker with Private Packages

### 10. Build with private registry access

```dockerfile
FROM node:20-slim AS builder

ARG NPM_TOKEN

WORKDIR /app

COPY package.json package-lock.json .npmrc ./
RUN echo "//npm.pkg.github.com/:_authToken=${NPM_TOKEN}" >> .npmrc && \
    npm ci && \
    rm -f .npmrc

COPY tsconfig.json ./
COPY src/ src/
RUN npm run build
RUN npm ci --omit=dev

FROM node:20-slim
WORKDIR /app
COPY --from=builder /app/node_modules node_modules/
COPY --from=builder /app/dist dist/
COPY --from=builder /app/package.json .
USER node
CMD ["node", "dist/server.js"]
```

```sh
docker build --build-arg NPM_TOKEN=$(gh auth token) -t myapp .
```

Multi-stage build ensures the token only exists in the builder stage.

## Troubleshooting

### 11. Common errors

| Error                             | Fix                                                              |
| --------------------------------- | ---------------------------------------------------------------- |
| `401 Unauthorized`                | Check auth token in ~/.npmrc                                     |
| `403 Forbidden`                   | Token lacks permissions (needs read:packages / write:packages)   |
| `404 Not Found`                   | Package name must match GitHub org scope (@yourorg/pkg)          |
| `E401` in CI                      | Ensure NODE_AUTH_TOKEN or NPM_TOKEN env var is set               |
| `UNABLE_TO_VERIFY_LEAF_SIGNATURE` | Corporate proxy — set `strict-ssl=false` in .npmrc (last resort) |

### 12. Debug npm registry issues

```sh
# Check what registry a package resolves to
npm config get @yourorg:registry

# Check auth
npm whoami --registry=https://npm.pkg.github.com

# Verbose install
npm ci --loglevel verbose

# Check which .npmrc files are loaded
npm config list
```
