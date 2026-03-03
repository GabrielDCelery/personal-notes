---
title: How to Dockerize a go app
author: GaborZeller
date: 2024-10-14T20-29-13Z
tags:
  - docker
  - go
draft: true
---

# How to Dockerize a go app

```dockerfile
FROM golang:1.23.2-bullseye as build

# add a user called app-user
# disable password authentication, but login is still possible
# no home directory
# no general information set for the account
# set the login shell that prevents logins but allows the user to exist for system purposes
# set uuid to arbitary value (since IDs are generally reserved below 1000)
RUN adduser \
    --disabled-password \
    --no-create-home \
    --gecos "" \
    --shell "/sbin/nologin" \
    --uid 65532 \
    app-user

# set GOPATH and GOCACHE (using ARG for build time only variables)
ARG GOPATH=/root/go
ARG GOCACHE=/root/.cache/go-build

# sets workdir to go source files path
WORKDIR ${GOPATH}/src/app

# copy dependency files to workdir to reduce build time if only code changes
COPY go.mod go.sum ./

# set mounts for better caching
RUN --mount=type=cache,target=${GOPATH}/pkg/mod \
    --mount=type=cache,target=${GOCACHE} \
    go mod download

# verify packages
RUN go mod verify

# copy project to workdir
COPY . .

# disable CGO for static linking to ensure compatibility with distroless
# sets the target os to linux and architecture to amd64
# compile application to /main
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /main ./cmd/get-market-orders/main.go

FROM scratch

# copy the timezone database that are critical for having the correct system time and handling date/time operations
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
# copy SSL certificates to be able to enable HTTPS connections
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# copy the user account information to be able to run system operations
COPY --from=build /etc/passwd /etc/passwd
# copy the group account information to be able to run system operations
COPY --from=build /etc/group /etc/group
# copy the main golang binary from the previous build step
COPY --from=build /main .

USER app-user:app-user

CMD ["/main"]
```

```Dockerfile
FROM golang@sha256:76dfe4aee4c0bf1ecd9666d29d22087eae592f697f449a88c0c7fc81a82faa01 as Builder

RUN apk add --no-cache bash openssh-client git ca-certificates

ENV GOPRIVATE github.com/ticker/
ENV USER=appuser
ENV UID=10001
ARG VERSION

RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

WORKDIR $GOPATH/src/github.com/ticker/services-trakm8-ingress


RUN mkdir -p -m 0600 ~/.ssh && ssh-keyscan github.com >> ~/.ssh/known_hosts
RUN git config --global url."git@github.com:".insteadOf "https://github.com/"

COPY go.mod go.sum ./
RUN --mount=type=ssh go mod download
RUN go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s -X 'main.BuildVersion=$VERSION'" -a -installsuffix cgo -o /go/bin/trkm8-ingress-server

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

COPY --from=builder /go/bin/trkm8-ingress-server /go/bin/trkm8-ingress-server

USER appuser:appuser
ENTRYPOINT ["/go/bin/trkm8-ingress-server"]

```

Stage 1: Builder (Lines 1-31)

Base Image & Dependencies

FROM golang@sha256:76dfe4aee4c0bf1ecd9666d29d22087eae592f697f449a88c0c7fc81a82faa01 as Builder

- Uses a specific Go image by SHA256 hash (pinned for reproducibility)
- Named Builder for the multi-stage build reference

RUN apk add --no-cache bash openssh-client git ca-certificates

- apk = Alpine package manager
- bash - shell for scripts
- openssh-client - needed for SSH key authentication with GitHub
- git - required for Go modules from private repos
- ca-certificates - SSL/TLS certificates for HTTPS connections

Environment Setup

ENV GOPRIVATE github.com/ticker/
ENV USER=appuser
ENV UID=10001
ARG VERSION

- GOPRIVATE - tells Go not to use public proxies for github.com/ticker/ repos (private repos)
- USER/UID - defines the non-root user that will run the app
- VERSION - build argument to inject version info into the binary

Security: Non-Root User Creation

RUN adduser \
 --disabled-password \
 --gecos "" \
 --home "/nonexistent" \
 --shell "/sbin/nologin" \
 --no-create-home \
 --uid "${UID}" \
      "${USER}"
Creates a minimal user account:

- No password (can't login)
- No GECOS info (user details)
- No home directory
- No shell access
- Fixed UID (10001) for consistency across builds
- This user is copied to the final image for running the app

SSH Configuration for Private Repos

RUN mkdir -p -m 0600 ~/.ssh && ssh-keyscan github.com >> ~/.ssh/known_hosts
RUN git config --global url."git@github.com:".insteadOf "https://github.com/"

- Creates .ssh directory with restricted permissions (0600 = owner read/write only)
- Adds GitHub's SSH host keys to avoid "unknown host" prompts
- Configures Git to use SSH instead of HTTPS for GitHub (required for private repos)

Go Dependencies

WORKDIR $GOPATH/src/github.com/ticker/services-trakm8-ingress

COPY go.mod go.sum ./
RUN --mount=type=ssh go mod download
RUN go mod verify

- Sets working directory to standard Go path
- Copies only go.mod and go.sum first (layer caching optimization - dependencies change less often than code)
- --mount=type=ssh - mounts SSH keys from build context (Docker BuildKit feature) to authenticate with private repos
- go mod download - downloads all dependencies
- go mod verify - checks integrity of downloaded modules

Build the Binary

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
 -ldflags="-w -s -X 'main.BuildVersion=$VERSION'" \
 -a -installsuffix cgo \
 -o /go/bin/trkm8-ingress-server

- Copies all source code
- Build flags:
  - CGO_ENABLED=0 - pure Go, no C dependencies (required for scratch base)
  - GOOS=linux - target Linux
  - GOARCH=amd64 - target AMD64 architecture
  - -ldflags="-w -s" - strip debug info (-w) and symbol table (-s) for smaller binary
  - -X 'main.BuildVersion=$VERSION' - inject version string into main.BuildVersion variable
  - -a - force rebuild of all packages
  - -installsuffix cgo - suffix for package installation directory (legacy, mainly for older Go versions)
  - -o /go/bin/trkm8-ingress-server - output binary name

Stage 2: Final Image (Lines 33-42)

Minimal Base

FROM scratch

- scratch is Docker's empty base image (literally nothing)
- Results in the smallest possible image (just the binary + essentials)
- No OS, no shell, no utilities - only what you explicitly copy

Copy Essentials from Builder

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

- ca-certificates.crt - needed for HTTPS/TLS connections to external services
- /etc/passwd - user database (contains the appuser we created)
- /etc/group - group database (contains the appuser group)

COPY --from=builder /go/bin/trkm8-ingress-server /go/bin/trkm8-ingress-server

- Copies the compiled binary from the builder stage

Security & Execution

USER appuser:appuser
ENTRYPOINT ["/go/bin/trkm8-ingress-server"]

- USER - run as non-root user (security best practice)
- ENTRYPOINT - defines the executable that runs when container starts
- Uses exec form (JSON array) to avoid shell overhead

Key Design Decisions

1. Multi-stage build - Keeps final image tiny by discarding build tools
2. Scratch base - Minimal attack surface, smallest possible image
3. Static binary - CGO disabled means no dynamic linking
4. Non-root execution - Security hardening
5. Pinned base image - Reproducible builds
6. SSH mount - Secure access to private repos without embedding keys in image
7. Layer optimization - Dependencies downloaded before code copy for better caching
8. Version injection - Build metadata baked into binary

This results in a production-ready, secure, minimal container image likely under 20MB.
