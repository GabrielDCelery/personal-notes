# Go Docker Builds

## Quick Reference

| Technique               | Purpose                                |
| ----------------------- | -------------------------------------- |
| `CGO_ENABLED=0`         | Static binary, no libc dependency      |
| Multi-stage build       | Small final image (scratch/distroless) |
| `go mod download` layer | Cache dependencies between builds      |
| `.dockerignore`         | Keep context small                     |

## Multi-Stage Builds

### 1. Basic production Dockerfile

```dockerfile
# -- Build stage --
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build static binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server

# -- Runtime stage --
FROM gcr.io/distroless/static-debian12

COPY --from=builder /app/server /server

ENTRYPOINT ["/server"]
```

### 2. With build args (version, commit)

```dockerfile
FROM golang:1.22-alpine AS builder

ARG VERSION=dev
ARG COMMIT=unknown

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build \
    -ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT}" \
    -o /app/server ./cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=builder /app/server /server
ENTRYPOINT ["/server"]
```

```sh
docker build \
  --build-arg VERSION=1.2.3 \
  --build-arg COMMIT=$(git rev-parse --short HEAD) \
  -t myapp:1.2.3 .
```

### 3. With private modules

```dockerfile
FROM golang:1.22-alpine AS builder

ARG GITHUB_TOKEN
RUN apk add --no-cache git
RUN git config --global url."https://x-access-token:${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"

ENV GOPRIVATE=github.com/yourorg/*

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /app/server ./cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=builder /app/server /server
ENTRYPOINT ["/server"]
```

```sh
# Pass token at build time — never bake into final image
docker build --build-arg GITHUB_TOKEN=$(gh auth token) -t myapp .
```

The multi-stage build ensures the token only exists in the builder stage, not the final image.

## Base Image Choices

### 4. Which base image to use

| Image                               | Size   | Use when                      |
| ----------------------------------- | ------ | ----------------------------- |
| `scratch`                           | 0 MB   | Fully static binary, no shell |
| `gcr.io/distroless/static-debian12` | ~2 MB  | Static binary, need CA certs  |
| `gcr.io/distroless/base-debian12`   | ~20 MB | Need glibc (CGO)              |
| `alpine`                            | ~7 MB  | Need shell for debugging      |

### 5. Using scratch (smallest possible)

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app/server ./cmd/server

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/server /server
ENTRYPOINT ["/server"]
```

Copy CA certs from the builder if your app makes HTTPS calls.

## Optimisation

### 6. Dependency caching with Docker layer cache

```dockerfile
# These layers only rebuild when go.mod/go.sum change
COPY go.mod go.sum ./
RUN go mod download

# This layer rebuilds on any source change
COPY . .
RUN CGO_ENABLED=0 go build -o /app/server ./cmd/server
```

### 7. .dockerignore

```
.git
.github
*.md
docs/
vendor/
bin/
tmp/
.env
```

### 8. BuildKit cache mount (faster rebuilds)

```dockerfile
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -o /app/server ./cmd/server
```

Requires `DOCKER_BUILDKIT=1` or Docker 23+.

## CI Patterns

### 9. GitHub Actions — build and push

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

### 10. Non-root user (security hardening)

```dockerfile
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /app/server /server

USER nonroot:nonroot
ENTRYPOINT ["/server"]
```
