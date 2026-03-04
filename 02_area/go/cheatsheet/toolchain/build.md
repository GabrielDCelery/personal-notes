# Go Build

## Quick Reference

| Use case            | Command / Method                          |
| ------------------- | ----------------------------------------- |
| Build               | `go build ./...`                          |
| Build with output   | `go build -o myapp ./cmd/myapp`           |
| Cross-compile       | `GOOS=linux GOARCH=amd64 go build`        |
| Inject version      | `go build -ldflags "-X main.version=..."` |
| Build tags          | `go build -tags integration`              |
| Generate code       | `go generate ./...`                       |
| Static binary       | `CGO_ENABLED=0 go build`                  |
| List OS/arch combos | `go tool dist list`                       |

## Cross-Compilation

### 1. Build for different platforms

```sh
# Linux
GOOS=linux GOARCH=amd64 go build -o myapp-linux

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o myapp-mac

# Windows
GOOS=windows GOARCH=amd64 go build -o myapp.exe

# ARM (Raspberry Pi)
GOOS=linux GOARCH=arm GOARM=7 go build -o myapp-arm
```

### 2. Common GOOS/GOARCH combos

| GOOS    | GOARCH | Description              |
| ------- | ------ | ------------------------ |
| linux   | amd64  | Linux x86_64             |
| linux   | arm64  | Linux ARM (AWS Graviton) |
| darwin  | amd64  | macOS Intel              |
| darwin  | arm64  | macOS Apple Silicon      |
| windows | amd64  | Windows x86_64           |

### 3. Static binary (no CGO)

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o myapp
```

> Required for scratch/distroless Docker images.

## Ldflags — Inject Build Info

### 4. Inject version at build time

```go
// main.go
var (
    version = "dev"
    commit  = "none"
    date    = "unknown"
)
```

```sh
go build -ldflags "\
  -X main.version=1.2.3 \
  -X main.commit=$(git rev-parse --short HEAD) \
  -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o myapp
```

### 5. Strip debug info (smaller binary)

```sh
go build -ldflags "-s -w" -o myapp
# -s  strip symbol table
# -w  strip DWARF debug info
```

## Build Tags

### 6. Conditional compilation

```go
//go:build integration

package myapp_test

func TestWithDatabase(t *testing.T) {
    // only runs with: go test -tags integration
}
```

### 7. Platform-specific code

```go
//go:build linux

package myapp

func platformSpecific() {
    // only compiled on Linux
}
```

### 8. Multiple tags

```go
//go:build integration && !race
```

```sh
go build -tags "integration debug"
go test -tags integration ./...
```

## go generate

### 9. Define generators

```go
//go:generate stringer -type=Status
//go:generate mockery --name=UserRepo
//go:generate protoc --go_out=. --go-grpc_out=. api.proto
```

```sh
go generate ./...
```

## Build Cache

### 10. Cache management

```sh
go clean -cache        # clear build cache
go clean -testcache    # clear test cache
go env GOCACHE         # show cache location
```

## Install

### 11. Install binary to GOBIN

```sh
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go env GOBIN           # where binaries go (default: $GOPATH/bin)
```
