# Lesson 07: BuildKit & Multi-platform Builds

BuildKit's advanced features, multi-architecture builds with buildx, and CI cache strategies.

## Why BuildKit Changed Everything

The classic Docker build engine was sequential, single-threaded, and had no way to handle secrets safely. BuildKit (enabled by default since Docker 23.0) rewrites the build execution model:

- **Parallel stage execution** — independent stages build concurrently
- **Better cache semantics** — more granular invalidation
- **`RUN --mount`** — cache mounts, secret mounts, SSH mounts without leaking to layers
- **Heredoc syntax** — inline files in Dockerfiles
- **Multi-platform builds** — build for arm64 from an x86 machine

```bash
# Verify BuildKit is enabled
docker buildx version
# github.com/docker/buildx v0.x.y ...

# Force enable for older Docker
DOCKER_BUILDKIT=1 docker build .

# Or set in daemon.json
{ "features": { "buildkit": true } }
```

---

## `RUN --mount` Types

The `--mount` flag on `RUN` instructions mounts resources during the build step without creating a new layer from them.

### `--mount=type=cache`

Persists a directory between builds. The cache is keyed and reused across runs — perfect for package manager caches.

```dockerfile
# syntax=docker/dockerfile:1

# ✓ npm cache persists between builds — `npm ci` is still deterministic
RUN --mount=type=cache,target=/root/.npm \
    npm ci

# ✓ apt cache — don't need to clean up /var/lib/apt/lists every time
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt,sharing=locked \
    apt-get update && apt-get install -y curl

# ✓ Go module cache
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -o /app/server .

# ✓ pip cache
RUN --mount=type=cache,target=/root/.cache/pip \
    pip install -r requirements.txt
```

**Cache sharing modes:**

| Mode      | Behaviour                                    |
| --------- | -------------------------------------------- |
| `shared`  | Multiple builds can use simultaneously       |
| `private` | One build at a time, others create new cache |
| `locked`  | One build at a time, others wait             |

**The cache mount does NOT invalidate** when package files change — you still control that with `COPY package*.json ./` before the `RUN`. The cache mount just speeds up `npm ci` by reusing the download cache.

```dockerfile
FROM node:20-alpine

WORKDIR /app
COPY package*.json ./

# Cache hit if package.json unchanged + npm cache exists on disk
RUN --mount=type=cache,target=/root/.npm \
    npm ci

COPY . .
RUN npm run build
```

### `--mount=type=secret`

Injects a secret into the build step without it ever being written to any layer.

```dockerfile
# syntax=docker/dockerfile:1
FROM node:20-alpine

WORKDIR /app
COPY package*.json ./

# Read .npmrc from a secret — never appears in image history
RUN --mount=type=secret,id=npmrc,target=/root/.npmrc \
    npm ci
```

```bash
# Pass the secret at build time
docker build \
  --secret id=npmrc,src=$HOME/.npmrc \
  .

# Or from environment variable
echo "//registry.npmjs.org/:_authToken=${NPM_TOKEN}" > /tmp/npmrc
docker build --secret id=npmrc,src=/tmp/npmrc .
```

**Verify the secret doesn't leak:**

```bash
docker history myimage   # ← secret is not visible
docker run --rm myimage cat /root/.npmrc  # ← file doesn't exist in container
```

### `--mount=type=ssh`

Forwards the host's SSH agent into the build for private repository access — without embedding the private key.

```dockerfile
# syntax=docker/dockerfile:1
FROM golang:1.22-alpine

# Add GitHub to known_hosts
RUN apk add --no-cache openssh-client \
    && mkdir -p -m 0700 ~/.ssh \
    && ssh-keyscan github.com >> ~/.ssh/known_hosts

WORKDIR /app
COPY go.mod go.sum ./

# SSH agent forwarded — go mod download can access private repos
RUN --mount=type=ssh \
    go mod download

COPY . .
RUN go build -o /app/server .
```

```bash
# Start ssh-agent and add your key
eval $(ssh-agent)
ssh-add ~/.ssh/id_rsa

# Build with SSH agent forwarding
docker build --ssh default .
```

### `--mount=type=bind`

Mounts a file/directory from the build context or another image without copying it into the layer (read-only by default).

```dockerfile
# Run tests using source files without COPYing them into the layer
RUN --mount=type=bind,target=/app/src,source=src \
    node /app/src/test.js
```

---

## Heredoc Syntax (Dockerfile 1.4+)

Heredocs eliminate the need for escaped multi-line `RUN` commands and allow inline file creation.

```dockerfile
# syntax=docker/dockerfile:1.4

# ✓ Heredoc for multi-line RUN
RUN <<EOF
set -e
apt-get update
apt-get install -y curl git
rm -rf /var/lib/apt/lists/*
EOF

# ✓ Create a file inline
COPY <<EOF /etc/myapp/config.yaml
port: 3000
log_level: info
EOF

# ✓ Create an executable script inline
COPY <<'EOF' /usr/local/bin/healthcheck.sh
#!/bin/sh
wget -qO- http://localhost:3000/health || exit 1
EOF
RUN chmod +x /usr/local/bin/healthcheck.sh
```

Single-quoted `<<'EOF'` prevents variable expansion (useful for scripts with `$VAR`).

---

## Build Arguments vs Build Secrets

|                             | `ARG`                                   | `--mount=type=secret`     |
| --------------------------- | --------------------------------------- | ------------------------- |
| Visible in `docker history` | ✓ Yes (after use)                       | ❌ No                     |
| Persists in layers          | If used in `RUN` env                    | ❌ No                     |
| Can be overridden at build  | ✓ Yes                                   | ✓ Yes                     |
| Good for                    | Build-time config (versions, platforms) | Credentials, tokens, keys |

```dockerfile
# ✓ ARG for non-sensitive build configuration
ARG NODE_ENV=production
ARG APP_VERSION=unknown
ENV NODE_ENV=${NODE_ENV}
LABEL version="${APP_VERSION}"

# ❌ ARG for secrets — visible in history
ARG NPM_TOKEN
RUN npm config set //registry.npmjs.org/:_authToken ${NPM_TOKEN}  # ← in history!

# ✓ Secret mount for credentials
RUN --mount=type=secret,id=npm_token \
    NPM_TOKEN=$(cat /run/secrets/npm_token) \
    npm config set //registry.npmjs.org/:_authToken $NPM_TOKEN \
    && npm ci
```

---

## Multi-platform Builds with `docker buildx`

`buildx` is Docker's extended build CLI plugin that supports multi-platform builds, multiple cache backends, and enhanced output formats.

### Why multi-platform matters

- Apple Silicon (arm64) development vs x86 production → crashes or emulation overhead
- Deploying to AWS Graviton (arm64) → 40% cheaper, 20% faster than x86 equivalents
- IoT/embedded → arm/v7, arm/v6 targets

### Building for multiple platforms

```bash
# Create a builder that supports multi-platform
docker buildx create --name mybuilder --use
docker buildx inspect --bootstrap   # ← starts the builder container

# Build for multiple platforms in one command
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --tag myregistry/myapp:latest \
  --push \                           # ← must push (can't load multi-platform to local)
  .

# Build for one platform, load locally
docker buildx build \
  --platform linux/arm64 \
  --load \
  --tag myapp:arm64 \
  .
```

### How it works: QEMU emulation

For building arm64 images on an x86 machine, buildx uses QEMU (Quick Emulator) via binfmt_misc to transparently execute arm64 binaries.

```bash
# Install QEMU emulators (needed on Linux, pre-installed on Docker Desktop)
docker run --privileged --rm tonistiigi/binfmt --install all
```

**Performance gotcha**: Building under QEMU is slow (10-50x slower than native). For large projects, use native arm64 build machines or a BuildKit builder farm with actual arm64 nodes.

### Platform-conditional Dockerfile logic

```dockerfile
FROM --platform=$BUILDPLATFORM golang:1.22 AS builder

ARG TARGETPLATFORM    # e.g., linux/arm64
ARG TARGETOS          # e.g., linux
ARG TARGETARCH        # e.g., arm64

WORKDIR /app
COPY . .

# Cross-compile for the target platform natively
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    CGO_ENABLED=0 \
    go build -o /app/server .

# ← $BUILDPLATFORM is the builder's platform (x86), but we cross-compiled
# ← No QEMU needed — Go cross-compilation is native

FROM --platform=$TARGETPLATFORM alpine:3.19
COPY --from=builder /app/server /server
ENTRYPOINT ["/server"]
```

`$BUILDPLATFORM` vs `$TARGETPLATFORM`:

- `$BUILDPLATFORM`: The machine running the build (x86 in this case)
- `$TARGETPLATFORM`: The image's target (arm64 in this case)

Using `FROM --platform=$BUILDPLATFORM` for the builder stage means it runs natively (fast), and you cross-compile to the target. Much faster than QEMU emulation.

---

## Build Cache Backends

BuildKit can export and import cache from various backends, enabling fast CI builds.

| Backend    | How to use                       | Best for                       |
| ---------- | -------------------------------- | ------------------------------ |
| `inline`   | Embedded in the pushed image     | Simple workflows, small images |
| `registry` | Separate cache image in registry | CI with a registry             |
| `local`    | Local filesystem directory       | Single-machine, large caches   |
| `gha`      | GitHub Actions cache             | GitHub Actions workflows       |
| `s3`       | AWS S3 bucket                    | AWS-based CI                   |
| `azblob`   | Azure Blob Storage               | Azure-based CI                 |

### Registry cache (most common for CI)

```bash
# Build, populate cache, push
docker buildx build \
  --cache-to type=registry,ref=myregistry/myapp:cache,mode=max \
  --cache-from type=registry,ref=myregistry/myapp:cache \
  --tag myregistry/myapp:latest \
  --push \
  .
```

`mode=max` caches all intermediate layers (including build stages), not just the final image layers. Slower first push, much faster subsequent builds.

### GitHub Actions cache

```yaml
# .github/workflows/build.yml
- name: Build and push
  uses: docker/build-push-action@v5
  with:
    context: .
    push: true
    tags: myregistry/myapp:latest
    cache-from: type=gha
    cache-to: type=gha,mode=max
```

The `gha` backend uses GitHub Actions' built-in cache service (up to 10GB free).

### Bake files (parallel multi-image builds)

`docker buildx bake` is like `make` for Docker builds — define multiple image targets and build them in parallel:

```hcl
# docker-bake.hcl
group "default" {
  targets = ["api", "worker", "scheduler"]
}

target "api" {
  context = "./services/api"
  tags = ["myregistry/api:latest"]
  cache-from = ["type=registry,ref=myregistry/api:cache"]
  cache-to = ["type=registry,ref=myregistry/api:cache,mode=max"]
}

target "worker" {
  context = "./services/worker"
  tags = ["myregistry/worker:latest"]
  cache-from = ["type=registry,ref=myregistry/worker:cache"]
  cache-to = ["type=registry,ref=myregistry/worker:cache,mode=max"]
}
```

```bash
docker buildx bake --push   # ← builds api, worker, scheduler in parallel
```

---

## Image Signing with cosign

Supply chain security: signing images ensures that what you pull is what was built in CI, not something tampered with.

```bash
# Install cosign
brew install sigstore/tap/cosign

# Sign an image (keyless, using OIDC identity)
cosign sign --yes myregistry/myapp@sha256:abc123...

# Verify before deploy
cosign verify \
  --certificate-identity-regexp="https://github.com/myorg/myrepo/.*" \
  --certificate-oidc-issuer="https://token.actions.githubusercontent.com" \
  myregistry/myapp:latest
```

In GitHub Actions:

```yaml
- name: Sign the image
  env:
    DIGEST: ${{ steps.build.outputs.digest }}
  run: |
    cosign sign --yes myregistry/myapp@${DIGEST}
```

The signature is stored as a separate OCI artifact in the registry alongside the image.

---

## Hands-On Exercise 1: Optimise a CI Build

A Node.js project takes 8 minutes in CI because `npm ci` runs every build. Rewrite the Dockerfile and CI config to use BuildKit cache mounts and registry cache.

```dockerfile
# Current Dockerfile (no cache optimisation)
FROM node:20-alpine
WORKDIR /app
COPY . .
RUN npm ci
RUN npm run build
```

<details>
<summary>Solution</summary>

```dockerfile
# syntax=docker/dockerfile:1

FROM node:20-alpine AS builder
WORKDIR /app

# Step 1: Layer deps separately so npm ci is cached when package.json unchanged
COPY package*.json ./

# Step 2: Cache mount so the npm download cache persists across builds
RUN --mount=type=cache,target=/root/.npm \
    npm ci

COPY . .
RUN npm run build

FROM node:20-alpine AS runtime
WORKDIR /app
COPY package*.json ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci --omit=dev
COPY --from=builder /app/dist ./dist
CMD ["node", "dist/index.js"]
```

```yaml
# .github/workflows/build.yml
- name: Set up Docker Buildx
  uses: docker/setup-buildx-action@v3

- name: Build and push
  uses: docker/build-push-action@v5
  with:
    context: .
    push: true
    tags: myregistry/myapp:${{ github.sha }}
    cache-from: type=gha
    cache-to: type=gha,mode=max
```

**Result**: If `package.json` hasn't changed, `npm ci` is skipped (layer cache hit). Even if it does change, the npm download cache mount means packages are read from disk rather than downloaded.

Typical improvement: 8 minutes → 2-3 minutes on cache hit, 4-5 minutes on partial miss.

</details>

---

## Hands-On Exercise 2: Multi-platform Build

Your team is moving to AWS Graviton (arm64) to reduce costs. Currently building for linux/amd64 only. Set up a multi-platform build that produces a manifest list for both architectures.

```bash
# Current build command
docker build -t myregistry/myapp:latest --push .
```

<details>
<summary>Solution</summary>

```bash
# Step 1: Create a buildx builder with multi-platform support
docker buildx create \
  --name multi-platform-builder \
  --driver docker-container \
  --use

# Step 2: Bootstrap (pulls the builder image)
docker buildx inspect --bootstrap

# Step 3: Build for both platforms and push
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --tag myregistry/myapp:latest \
  --cache-from type=registry,ref=myregistry/myapp:cache \
  --cache-to type=registry,ref=myregistry/myapp:cache,mode=max \
  --push \
  .
```

```yaml
# .github/workflows/multi-platform.yml
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up QEMU (for arm64 emulation)
        uses: docker/setup-qemu-action@v3

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Build and push multi-platform image
        uses: docker/build-push-action@v5
        with:
          context: .
          platforms: linux/amd64,linux/arm64
          push: true
          tags: myregistry/myapp:latest
          cache-from: type=gha
          cache-to: type=gha,mode=max
```

**For a Go service (faster, using cross-compilation instead of QEMU)**:

```dockerfile
# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS builder
ARG TARGETOS TARGETARCH
WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 \
    go build -ldflags="-w -s" -o /server .

FROM --platform=$TARGETPLATFORM scratch
COPY --from=builder /server /server
ENTRYPOINT ["/server"]
```

This builds natively on the build machine (no QEMU), cross-compiling to the target arch. Much faster.

</details>

---

## Interview Questions

### Q1: How do BuildKit cache mounts differ from layer caching, and when do you use each?

Tests whether you understand the two distinct caching mechanisms in BuildKit and can explain the trade-off.

<details>
<summary>Answer</summary>

**Layer caching** is Docker's standard mechanism: if the instruction and its inputs (files, command string) haven't changed since the last build, the cached layer is reused and the instruction isn't re-run.

**Cache mounts** (`--mount=type=cache`) are persistent directories that survive across builds, like a shared disk. They're used for package manager caches (npm, pip, apt) that are separate from the layer contents.

Key differences:

|                        | Layer cache                  | Cache mount                     |
| ---------------------- | ---------------------------- | ------------------------------- |
| Invalidated when       | Instruction or input changes | Never (manual `--no-cache`)     |
| Where stored           | Layer overlay filesystem     | BuildKit content store          |
| Visible in final image | ✓ Yes                        | ❌ No                           |
| Shared across builds   | ✓ Yes (if nothing changed)   | ✓ Always                        |
| Use for                | Static, deterministic steps  | Package manager download caches |

**Example**: `RUN npm ci` with layer caching re-runs if `package.json` changes. With a cache mount on `/root/.npm`, even when re-running, packages that were previously downloaded are read from disk rather than re-fetched from the internet. You get both: layer cache for no-change cases, cache mount for faster re-runs when deps change.

</details>

---

### Q2: What is a manifest list and how does multi-platform docker buildx produce one?

Tests understanding of Docker's image format internals — an increasingly common interview topic as arm64 becomes mainstream.

<details>
<summary>Answer</summary>

A **manifest** is a JSON document describing an image: its layers, config, and metadata. Stored in a registry by digest (`sha256:...`).

A **manifest list** (or OCI Image Index) is a manifest that points to multiple platform-specific manifests:

```json
{
  "manifests": [
    {
      "platform": { "os": "linux", "architecture": "amd64" },
      "digest": "sha256:aaa..."
    },
    {
      "platform": { "os": "linux", "architecture": "arm64" },
      "digest": "sha256:bbb..."
    }
  ]
}
```

When you `docker pull myimage:latest` on an arm64 machine, the registry returns the manifest list, the Docker client checks the platform, and pulls the arm64-specific manifest.

`docker buildx build --platform linux/amd64,linux/arm64 --push` creates one manifest per platform, then creates a manifest list pointing to both, and pushes all three to the registry under the same tag.

You can inspect a manifest list:

```bash
docker manifest inspect myregistry/myapp:latest
# Shows the manifest list with platform entries

docker buildx imagetools inspect myregistry/myapp:latest
# More detailed, shows layer info per platform
```

**Why `--push` is required**: Loading a multi-platform image locally is ambiguous (which platform?). The registry is the intermediary that handles manifest list storage. Use `--load` for single-platform builds.

</details>

---

### Q3: A secret is needed during `docker build` to authenticate to a private npm registry. What's the safest way to provide it?

Security and BuildKit knowledge combined. Tests whether you know the wrong approaches and why.

<details>
<summary>Answer</summary>

**Wrong approaches:**

1. `ARG NPM_TOKEN` → visible in `docker history`
2. `COPY .npmrc ./` → `.npmrc` with token ends up in the image layer (even if deleted later)
3. `ENV NPM_TOKEN=xxx` → persists in the image; visible in `docker inspect`

**Correct approach: BuildKit secret mount**

```dockerfile
# syntax=docker/dockerfile:1
FROM node:20-alpine
WORKDIR /app
COPY package*.json ./

RUN --mount=type=secret,id=npmrc,target=/root/.npmrc \
    npm ci
```

```bash
# Build — secret is read from the file, never stored in any layer
docker build \
  --secret id=npmrc,src=$HOME/.npmrc \
  --tag myapp:latest \
  .
```

The secret is:

- Only available during that specific `RUN` instruction
- Never written to any layer (confirmed with `docker history`)
- Not visible in `docker inspect`
- Cleaned up automatically after the instruction completes

**In CI (GitHub Actions):**

```yaml
- name: Build
  run: |
    echo "//registry.npmjs.org/:_authToken=${{ secrets.NPM_TOKEN }}" > /tmp/npmrc
    docker build --secret id=npmrc,src=/tmp/npmrc -t myapp .
    rm /tmp/npmrc
```

</details>

---

### Q4: Your Go service builds fine on your M2 Mac (arm64) but crashes on the Linux CI server (amd64). What are the likely causes and how do you prevent this?

A practical cross-platform debugging question that tests real-world multi-platform experience.

<details>
<summary>Answer</summary>

**Most likely cause**: The image was built for `linux/arm64` on the Mac but the CI server pulls and runs it on `linux/amd64`. The binary is the wrong architecture.

```bash
# Verify the image architecture
docker inspect myapp:latest --format '{{ .Architecture }}'
# → arm64  ← wrong for an amd64 CI server
```

**Other causes:**

- C library differences: code using CGo links against macOS system libraries, not Linux
- Platform-specific syscalls or filesystem behaviour
- Test environments that pass on arm64 but fail on amd64 due to endianness or alignment

**Prevention:**

1. **Build multi-platform in CI** (don't push arm64-only images):

```bash
docker buildx build --platform linux/amd64,linux/arm64 --push .
```

2. **Cross-compile for Go instead of using QEMU**:

```dockerfile
FROM --platform=$BUILDPLATFORM golang:1.22 AS builder
ARG TARGETOS TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build -o /server .
```

3. **Test both architectures in CI** using matrix builds:

```yaml
strategy:
  matrix:
    platform: [linux/amd64, linux/arm64]
```

4. **Lock to a platform** in Compose for local dev to match production:

```yaml
services:
  api:
    platform: linux/amd64 # ← always run as amd64, even on Mac
```

</details>

---

## Key Takeaways

1. **`--mount=type=cache`** persists package manager caches across builds without adding them to image layers
2. **`--mount=type=secret`** injects secrets during build without writing them to any layer or appearing in `docker history`
3. **`--mount=type=ssh`** forwards the SSH agent for private repo access without embedding private keys
4. **Cache modes**: `mode=max` caches all intermediate stages; use it with registry caches in CI for maximum reuse
5. **Multi-platform builds** use QEMU emulation by default — use `$BUILDPLATFORM` + cross-compilation for speed
6. **Manifest lists** let a single image tag serve multiple architectures; `docker buildx build --platform ... --push` creates them
7. **`FROM --platform=$BUILDPLATFORM`** in builder stages means the builder runs natively (not emulated) even when targeting a different arch
8. **`docker buildx bake`** builds multiple images in parallel from a declarative HCL or JSON config

## Next Steps

In [Lesson 08: Debugging & Troubleshooting](lesson-08-debugging-and-troubleshooting.md), you'll learn:

- How to inspect containers, networks, and volumes at depth with `docker inspect`
- Using `docker exec` and `nsenter` for live debugging
- Layer analysis tools (dive) to diagnose image size issues
- Container resource metrics and profiling
- Systematic diagnosis of the most common production failure patterns
