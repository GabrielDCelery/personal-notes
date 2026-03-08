# Node.js Docker Builds

## Quick Reference

| Technique                | Purpose                               |
| ------------------------ | ------------------------------------- |
| Multi-stage build        | Small final image, no devDependencies |
| `npm ci` layer           | Cache dependencies between builds     |
| `node:*-slim` / `alpine` | Smaller base images                   |
| `.dockerignore`          | Keep context small                    |
| Non-root user            | Security hardening                    |

## Multi-Stage Builds

### 1. Basic production Dockerfile

```dockerfile
# -- Build stage --
FROM node:20-slim AS builder

WORKDIR /app

# Cache dependencies
COPY package.json package-lock.json ./
RUN npm ci

# Build
COPY tsconfig.json ./
COPY src/ src/
RUN npm run build

# Prune devDependencies
RUN npm ci --omit=dev

# -- Runtime stage --
FROM node:20-slim

WORKDIR /app

COPY --from=builder /app/node_modules node_modules/
COPY --from=builder /app/dist dist/
COPY --from=builder /app/package.json .

USER node
EXPOSE 3000

CMD ["node", "dist/server.js"]
```

### 2. With build args (version, commit)

```dockerfile
FROM node:20-slim AS builder

ARG VERSION=dev
ARG COMMIT=unknown

WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci

COPY tsconfig.json ./
COPY src/ src/
RUN npm run build

RUN npm ci --omit=dev

FROM node:20-slim
WORKDIR /app

COPY --from=builder /app/node_modules node_modules/
COPY --from=builder /app/dist dist/
COPY --from=builder /app/package.json .

ENV APP_VERSION=${VERSION} APP_COMMIT=${COMMIT}

USER node
CMD ["node", "dist/server.js"]
```

```sh
docker build \
  --build-arg VERSION=1.2.3 \
  --build-arg COMMIT=$(git rev-parse --short HEAD) \
  -t myapp:1.2.3 .
```

### 3. With private npm registry

```dockerfile
FROM node:20-slim AS builder

ARG NPM_TOKEN

WORKDIR /app
COPY package.json package-lock.json .npmrc ./
RUN echo "//registry.npmjs.org/:_authToken=${NPM_TOKEN}" >> .npmrc && \
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
docker build --build-arg NPM_TOKEN=$(npm config get //registry.npmjs.org/:_authToken) -t myapp .
```

Multi-stage build ensures the token only exists in the builder stage.

## Base Image Choices

### 4. Which base image to use

| Image                                 | Size    | Use when                                           |
| ------------------------------------- | ------- | -------------------------------------------------- |
| `node:20-slim`                        | ~200 MB | Default choice, Debian-based                       |
| `node:20-alpine`                      | ~130 MB | Smaller, but musl libc (some native modules break) |
| `node:20`                             | ~1 GB   | Need full OS tools (avoid in prod)                 |
| `gcr.io/distroless/nodejs20-debian12` | ~130 MB | No shell, maximum security                         |

### 5. Alpine (smaller but watch for native module issues)

```dockerfile
FROM node:20-alpine AS builder
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY tsconfig.json ./
COPY src/ src/
RUN npm run build
RUN npm ci --omit=dev

FROM node:20-alpine
WORKDIR /app
COPY --from=builder /app/node_modules node_modules/
COPY --from=builder /app/dist dist/
COPY --from=builder /app/package.json .
USER node
CMD ["node", "dist/server.js"]
```

## Optimisation

### 6. Dependency caching with Docker layer cache

```dockerfile
# These layers only rebuild when package.json/lock change
COPY package.json package-lock.json ./
RUN npm ci

# This layer rebuilds on any source change
COPY tsconfig.json ./
COPY src/ src/
RUN npm run build
```

### 7. .dockerignore

```
node_modules
dist
.git
.github
*.md
.env
.env.*
coverage
.vscode
.idea
```

### 8. BuildKit cache mount (faster rebuilds)

```dockerfile
RUN --mount=type=cache,target=/root/.npm \
    npm ci
```

Requires `DOCKER_BUILDKIT=1` or Docker 23+.

## Security

### 9. Non-root user

```dockerfile
# node:* images include a "node" user (uid 1000)
FROM node:20-slim

WORKDIR /app
COPY --from=builder --chown=node:node /app .

USER node
CMD ["node", "dist/server.js"]
```

### 10. Health check

```dockerfile
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s \
  CMD node -e "fetch('http://localhost:3000/health').then(r => process.exit(r.ok ? 0 : 1))"
```

## CI Patterns

### 11. GitHub Actions — build and push

```yaml
- name: Build and push
  uses: docker/build-push-action@v5
  with:
    context: .
    push: true
    tags: ghcr.io/yourorg/myapp:${{ github.sha }}
    build-args: |
      VERSION=${{ github.ref_name }}
      COMMIT=${{ github.sha }}
    cache-from: type=gha
    cache-to: type=gha,mode=max
```

### 12. Docker Compose for local dev

```yaml
# docker-compose.yml
services:
  app:
    build: .
    ports:
      - "3000:3000"
    environment:
      - DATABASE_URL=postgres://user:pass@db:5432/mydb
      - REDIS_URL=redis://redis:6379
    depends_on:
      - db
      - redis

  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: pass
      POSTGRES_DB: mydb
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

volumes:
  pgdata:
```
