# Lesson 01: Image Internals & Optimization

How Docker images actually work under the hood, and how to build fast, small, secure images.

## How Docker Images Are Stored: OverlayFS

You've been using images for years, but do you know why `docker pull` shows all those "layer" lines, or why changing one line in your Dockerfile can bust the entire cache? The answer is OverlayFS — the union filesystem Docker uses on Linux.

Docker images are a stack of read-only layers. When you run a container, Docker adds one writable layer on top. OverlayFS merges these into a single view of the filesystem.

```
Container (writable layer)     ← your running process writes here
─────────────────────────────
Layer 4: COPY . /app           ← read-only
Layer 3: RUN npm ci            ← read-only
Layer 2: COPY package*.json ./ ← read-only
Layer 1: FROM node:20-alpine   ← read-only (the base image)
```

Each layer is a diff (added/modified/deleted files) stored as a tar archive in `/var/lib/docker/overlay2/`. The `lower`, `upper`, and `merged` directories are the OverlayFS mechanism.

**Why this matters for interviews**: Understanding layers explains cache invalidation, copy-on-write behaviour, and why large files added then deleted don't reduce image size.

### Copy-on-Write (CoW)

When a container modifies a file from a read-only layer, OverlayFS copies it into the writable layer first. This means:

- Reading from lower layers is cheap
- First write to a file in a lower layer has overhead (the copy-up)
- Deleting a file from a lower layer in a `RUN` command still keeps it in that layer; only the reference is hidden

```dockerfile
# ❌ This image is still ~500MB - the tarball exists in an earlier layer
FROM ubuntu:22.04
RUN apt-get update && apt-get install -y build-essential
RUN wget https://example.com/large-source.tar.gz
RUN tar -xzf large-source.tar.gz && make && make install
RUN rm large-source.tar.gz  # ← too late, layer 3 still has it

# ✓ Combine into one RUN to keep the tarball out of the final image
FROM ubuntu:22.04
RUN apt-get update && apt-get install -y build-essential \
    && wget https://example.com/large-source.tar.gz \
    && tar -xzf large-source.tar.gz \
    && make && make install \
    && rm -rf large-source.tar.gz /var/lib/apt/lists/*
```

---

## Layer Caching: The Most Important Build Optimisation

Docker caches each layer. If nothing changed since the last build, Docker reuses the cached layer instead of re-running the instruction. Cache invalidation cascades: once one layer is invalidated, every subsequent layer is also rebuilt.

### Cache Invalidation Rules

| Instruction    | When cache is invalidated                            |
| -------------- | ---------------------------------------------------- |
| `FROM`         | Base image digest changes                            |
| `RUN`          | The command string changes                           |
| `COPY` / `ADD` | Any file in the source path changes (checksum-based) |
| `ENV` / `ARG`  | Value changes                                        |
| `WORKDIR`      | Path changes                                         |

### The Golden Rule: Order by change frequency (least to most)

```dockerfile
# ❌ Dependencies re-installed every time source code changes
FROM node:20-alpine
WORKDIR /app
COPY . .                    # ← all source files, changes constantly
RUN npm ci                  # ← invalidated on every code change

# ✓ Dependencies cached until package.json changes
FROM node:20-alpine
WORKDIR /app
COPY package*.json ./       # ← only changes when deps change
RUN npm ci                  # ← cache hit 95% of the time
COPY . .                    # ← only this layer rebuilds on code change
```

---

## Multi-stage Builds

Build-time dependencies don't belong in production images. Multi-stage builds let you use a fat builder image and copy only the output into a minimal runtime image.

```dockerfile
# Stage 1: Builder
FROM node:20 AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build           # produces /app/dist

# Stage 2: Runtime
FROM node:20-alpine AS runtime
WORKDIR /app
COPY --from=builder /app/dist ./dist
COPY --from=builder /app/node_modules ./node_modules
# node_modules from builder already has prod deps only if you used
# npm ci --omit=dev in the builder stage

EXPOSE 3000
CMD ["node", "dist/index.js"]
```

### Multi-stage Patterns

| Pattern                | When to use                                                 |
| ---------------------- | ----------------------------------------------------------- |
| Builder + slim runtime | Compiled languages, transpiled JS/TS                        |
| Test stage             | Run tests in CI without including test deps in final image  |
| Dev stage              | Full dev tooling for local development                      |
| Conditional `--target` | `docker build --target builder` to stop at a specific stage |

```dockerfile
FROM node:20 AS base
WORKDIR /app
COPY package*.json ./

FROM base AS dev
RUN npm ci
COPY . .
CMD ["npm", "run", "dev"]

FROM base AS builder
RUN npm ci --omit=dev       # ✓ skip devDependencies for production
COPY . .
RUN npm run build

FROM node:20-alpine AS prod  # ✓ smallest possible runtime
WORKDIR /app
COPY --from=builder /app/dist ./dist
COPY --from=builder /app/node_modules ./node_modules
CMD ["node", "dist/index.js"]
```

Build targets:

```bash
docker build --target dev -t myapp:dev .
docker build --target prod -t myapp:prod .
```

---

## Base Image Selection

The base image is the biggest lever you have on image size and attack surface.

| Base                                  | Approx size | Use case                                 |
| ------------------------------------- | ----------- | ---------------------------------------- |
| `ubuntu:22.04`                        | ~80MB       | General purpose, familiar tooling        |
| `debian:bookworm-slim`                | ~80MB       | Debian without extras                    |
| `alpine:3.19`                         | ~7MB        | Minimal, musl libc (⚠️ see gotcha below) |
| `node:20-alpine`                      | ~130MB      | Node.js on Alpine                        |
| `node:20-slim`                        | ~240MB      | Node.js on Debian slim                   |
| `gcr.io/distroless/nodejs20-debian12` | ~170MB      | No shell, no package manager             |
| `scratch`                             | 0 bytes     | Statically compiled binaries only        |

**Alpine gotcha**: Alpine uses `musl` libc instead of `glibc`. Most npm packages are fine, but some native modules (like `sharp`) have prebuilt binaries compiled against `glibc` and fail on Alpine. Watch for:

```
Error: /lib/x86_64-linux-gnu/libc.so.6: version 'GLIBC_2.29' not found
```

**Distroless**: No shell (`/bin/sh`), no package manager, no coreutils. Dramatically reduces attack surface. Debugging is harder — use a debug variant in dev:

```dockerfile
FROM gcr.io/distroless/nodejs20-debian12:debug AS debug  # has busybox shell
FROM gcr.io/distroless/nodejs20-debian12 AS prod
```

---

## .dockerignore

`.dockerignore` filters the build context sent to the Docker daemon. This matters because without it, `COPY . .` sends everything — including `node_modules`, `.git`, and secrets.

```
# .dockerignore
node_modules/       # ← re-installed inside container anyway
.git/               # ← contains full repo history, potentially large
.env                # ← never send secrets to build context
dist/               # ← generated output, not needed
**/*.log
**/.DS_Store
coverage/
.nyc_output/
```

**Gotcha**: `.dockerignore` also affects which `COPY` commands can access files. If you ignore a directory, `COPY` can't copy it even if you explicitly try.

---

## Hands-On Exercise 1: Audit and Optimise a Dockerfile

The following Dockerfile works but is poorly optimised. Identify all the issues and rewrite it.

```dockerfile
FROM node:20
WORKDIR /app
COPY . .
RUN npm install
RUN npm run test
RUN npm run build
RUN npm cache clean --force
EXPOSE 3000
CMD ["node", "server.js"]
```

<details>
<summary>Solution</summary>

**Issues:**

1. ❌ `FROM node:20` — heavy image (~1GB), use `node:20-alpine` or `node:20-slim` for production
2. ❌ `COPY . .` before `npm install` — busts the npm cache on every source file change
3. ❌ Tests run inside the production image build — test dependencies end up in the final image
4. ❌ `RUN npm install` instead of `npm ci` — `npm install` can modify package-lock.json; `npm ci` is deterministic
5. ❌ `npm cache clean` is a separate layer — the npm cache from `RUN npm install` already consumed a layer
6. ❌ Not a multi-stage build — dev/test tools and source files are all in the final image

**Fixed:**

```dockerfile
# Stage 1: Install and test
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci                  # ✓ deterministic, cached when package*.json unchanged
COPY . .
RUN npm run test            # ✓ tests in builder, not in final image
RUN npm run build

# Stage 2: Slim runtime
FROM node:20-alpine AS runtime
WORKDIR /app
ENV NODE_ENV=production
COPY package*.json ./
RUN npm ci --omit=dev       # ✓ prod deps only
COPY --from=builder /app/dist ./dist
EXPOSE 3000
CMD ["node", "dist/server.js"]
```

Also add `.dockerignore`:

```
node_modules/
.git/
dist/
coverage/
*.log
```

</details>

---

## Hands-On Exercise 2: Multi-stage Go Build

Write a multi-stage Dockerfile for a Go application that produces the smallest possible final image.

Requirements:

- Use Go 1.22 for building
- The final image must have no shell or package manager
- The binary must run as a non-root user

```go
// main.go - a simple HTTP server
package main

import (
    "fmt"
    "net/http"
)

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "Hello")
    })
    http.ListenAndServe(":8080", nil)
}
```

<details>
<summary>Solution</summary>

```dockerfile
# Stage 1: Build a statically linked binary
FROM golang:1.22-alpine AS builder
WORKDIR /app

# Download dependencies first for caching
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0: static binary (no glibc dependency)
# -ldflags="-w -s": strip debug info and symbol table (~30% smaller)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server .

# Stage 2: Scratch - literally empty
FROM scratch
# Copy timezone data if needed
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
# Copy CA certs for HTTPS calls
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# Copy the binary
COPY --from=builder /app/server /server

# Non-root: scratch has no /etc/passwd, so use numeric UID directly
USER 65534:65534   # nobody:nogroup

EXPOSE 8080
ENTRYPOINT ["/server"]
```

Result: ~10MB final image with no shell attack surface.

**Gotchas:**

- `scratch` has no shell, no DNS config — you need to copy `/etc/resolv.conf` manually if DNS isn't working, or use `FROM gcr.io/distroless/static-debian12` which handles this
- If the app reads files at runtime (templates, configs), `COPY` them into scratch too
- `CGO_ENABLED=0` is critical — without it, Go links against glibc and the binary won't run on scratch

</details>

---

## Interview Questions

### Q1: What is a Docker image layer and why can't you reduce image size by deleting files in a subsequent RUN command?

Interviewers ask this to check whether you understand OverlayFS internals versus just knowing Docker commands. It reveals whether you'll make a common production mistake (leaking build secrets or large files into final images).

<details>
<summary>Answer</summary>

Each `RUN`, `COPY`, and `ADD` instruction creates an immutable layer — a filesystem diff stored as a tar archive. Layers are content-addressed and stacked using OverlayFS (or other union filesystem drivers).

When you delete a file in a subsequent `RUN`, Docker doesn't remove it from the previous layer. It adds a "whiteout" file (`.wh.<filename>`) in the new layer that hides the file at runtime. But the original data is still present in the earlier layer and contributes to the image size.

```dockerfile
# ❌ large-file.tar.gz still exists in layer 2, image is large
RUN wget https://example.com/large-file.tar.gz
RUN rm large-file.tar.gz

# ✓ combined into one layer, file never persists
RUN wget https://example.com/large-file.tar.gz \
    && do-something-with-it \
    && rm large-file.tar.gz
```

This is also why you should never put secrets into intermediate layers even if you delete them — they're still in the image history.

</details>

---

### Q2: When does `COPY package*.json ./` followed by `RUN npm ci` get a cache hit vs a cache miss?

Interviewers use this to see if you understand Docker's build cache invalidation mechanism — a critical skill for keeping CI pipelines fast.

<details>
<summary>Answer</summary>

Docker computes a checksum of the files being copied and compares it to the cached layer. A cache **hit** occurs when:

- The instruction string is identical
- The checksums of all files matched by `package*.json` are identical to the last build

A cache **miss** occurs when:

- `package.json` or `package-lock.json` content has changed
- A new file matches `package*.json` (e.g., you added `package-extra.json`)
- Any parent layer was invalidated (cache invalidation cascades downward)

When there's a cache miss on the `COPY` layer, `RUN npm ci` is also re-run — even if the command string is identical. This is why layer ordering matters: put frequently-changing `COPY` instructions after stable ones.

**Common gotcha**: `ARG` instructions can invalidate cache too. `ARG BUILD_DATE` before a `COPY` will invalidate the copy layer every build if `BUILD_DATE` changes.

</details>

---

### Q3: What's the difference between `CMD` and `ENTRYPOINT`, and when would you use each?

A classic Docker interview question. Interviewers want to see you understand the exec vs shell form distinction and how the two interact, not just give a surface-level answer.

<details>
<summary>Answer</summary>

**ENTRYPOINT** sets the process that always runs. **CMD** provides default arguments that can be overridden at runtime.

|               | ENTRYPOINT                | CMD                             |
| ------------- | ------------------------- | ------------------------------- |
| Overridden by | `docker run --entrypoint` | `docker run <image> <args>`     |
| Typical use   | Fixed executable          | Default args / fallback command |

They interact: if both are set, CMD is passed as arguments to ENTRYPOINT.

```dockerfile
ENTRYPOINT ["node"]
CMD ["dist/index.js"]
# docker run myapp           → node dist/index.js
# docker run myapp dist/other.js → node dist/other.js
```

**Exec form vs shell form**:

```dockerfile
CMD ["node", "server.js"]   # ✓ exec form: node is PID 1, gets SIGTERM
CMD node server.js           # ❌ shell form: /bin/sh -c is PID 1, node is a child
                             #    SIGTERM goes to shell, not node — graceful shutdown breaks
```

Always use exec form in production so your process receives Unix signals correctly.

</details>

---

### Q4: How would you prevent secrets from leaking into a Docker image during the build?

Interviewers ask this in security-focused interviews. It tests whether you know the common mistake (secrets in ENV/ARG/RUN) and the proper solution (BuildKit secrets).

<details>
<summary>Answer</summary>

**The problem**: `ARG` and `ENV` values are baked into the image history. `RUN echo $SECRET` appears in `docker history`. Even `RUN --env SECRET=$SECRET` leaks it.

**Approaches in order of preference**:

1. **BuildKit `--mount=type=secret`** (best, secret never persists to any layer):

```dockerfile
# syntax=docker/dockerfile:1
RUN --mount=type=secret,id=npm_token \
    NPM_TOKEN=$(cat /run/secrets/npm_token) npm ci
```

```bash
docker build --secret id=npm_token,src=.npmrc .
```

2. **Multi-stage + only copy outputs**: Secrets used in a build stage never end up in the final stage if you only `COPY --from=builder` the compiled output.

3. **Build-time environment only, no persistence**: If you must use `ARG`, ensure the secret is only used in a single `RUN` and never in `ENV` (ENV persists to the final image).

```dockerfile
# ❌ Secret in ENV — visible in docker inspect and docker history
ENV API_KEY=supersecret

# ❌ Secret in ARG — visible in docker history
ARG API_KEY
RUN curl -H "Authorization: $API_KEY" https://api.example.com/download
```

</details>

---

## Key Takeaways

1. **OverlayFS layers are immutable** — deleting files in a later layer hides them but doesn't remove them from the image size
2. **Cache invalidation cascades** — once one layer busts, every layer after it rebuilds
3. **Order Dockerfiles by change frequency** — stable deps first, frequently-changing source last
4. **Multi-stage builds** separate build tooling from the runtime image; only copy what you need
5. **`npm ci` not `npm install`** in Dockerfiles — deterministic, respects lockfile, faster
6. **Exec form `CMD ["node", "app.js"]`** is required for signal handling; shell form wraps with `/bin/sh`
7. **`scratch` and distroless** images eliminate shell attack surface at the cost of debuggability
8. **Never put secrets in `ARG`/`ENV`** — use BuildKit `--mount=type=secret` instead
9. **`.dockerignore` reduces build context size** and prevents accidental inclusion of secrets

## Next Steps

In [Lesson 02: Networking Deep Dive](lesson-02-networking-deep-dive.md), you'll learn:

- How Docker implements network namespaces and isolation
- The differences between bridge, host, overlay, and macvlan drivers
- How container DNS resolution works — and why custom networks are required for it
- Exposed vs published ports and what `0.0.0.0` binding means
