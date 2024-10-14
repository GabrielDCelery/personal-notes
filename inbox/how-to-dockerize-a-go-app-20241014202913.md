---
title: How to Dockerize a go app
author: GaborZeller
date: 2024-10-14T20-29-13Z
tags:
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
