# Go Commands

## Quick Reference

| Command         | Purpose                                  |
| --------------- | ---------------------------------------- |
| `go mod init`   | Initialize a new module                  |
| `go mod tidy`   | Add missing / remove unused dependencies |
| `go get`        | Add or update a dependency               |
| `go build`      | Compile packages and dependencies        |
| `go run`        | Compile and run a program                |
| `go test`       | Run tests                                |
| `go vet`        | Report likely mistakes in code           |
| `go fmt`        | Format source code                       |
| `go doc`        | Show documentation for a symbol          |
| `go install`    | Compile and install a binary to `$GOBIN` |
| `go work`       | Manage multi-module workspaces           |
| `go generate`   | Run code generation directives           |
| `go env`        | Print Go environment variables           |
| `go tool pprof` | Profile CPU, memory, goroutines          |
| `go mod vendor` | Copy dependencies into `vendor/`         |

## Module Management

### 1. Initialize a module

```sh
go mod init github.com/yourorg/yourproject
```

### 2. Add / update dependencies

```sh
# Add a specific dependency
go get github.com/lib/pq

# Add at a specific version
go get github.com/lib/pq@v1.10.9

# Update to latest
go get -u github.com/lib/pq

# Update all direct and indirect dependencies
go get -u ./...
```

### 3. Clean up go.mod and go.sum

```sh
# Remove unused deps, add missing ones, update go.sum
go mod tidy
```

### 4. Vendor dependencies (CI / air-gapped builds)

```sh
go mod vendor
go build -mod=vendor ./...
```

### 5. List module dependencies

```sh
# Direct and indirect deps
go list -m all

# Check for available updates
go list -m -u all

# Why is a dependency needed?
go mod why github.com/lib/pq
```

## Build & Run

### 6. Build a binary

```sh
# Build current package
go build -o myapp .

# Build with version info (common in prod)
# ldflags passes flags into the linker at build time
# eg. -X main.version=1.2.3 sets var version string in package main
go build -ldflags "-X main.version=1.2.3 -X main.commit=$(git rev-parse --short HEAD)" -o myapp .

# Cross-compile
# check go doc runtime
GOOS=linux GOARCH=amd64 go build -o myapp-linux .
GOOS=darwin GOARCH=arm64 go build -o myapp-darwin .
GOOS=windows GOARCH=amd64 go build -o myapp.exe .
```

### 7. Run without building a binary

```sh
go run .
go run ./cmd/server
```

### 8. Install a tool globally

```sh
# Install to $GOBIN (default $GOPATH/bin)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install local project binary
go install ./cmd/myapp
```

## Testing

### 9. Run tests

```sh
# All tests in current module
go test ./...

# Specific package
go test ./internal/auth

# Verbose output
go test -v ./...

# Run specific test by name
go test -run TestUserCreate ./internal/auth

# With race detector (always use in CI)
go test -race ./...

# Short mode (skip long-running tests)
go test -short ./...
```

### 10. Test coverage

```sh
# Print coverage percentage
go test -cover ./...

# Generate coverage profile and view in browser
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Coverage for specific packages only
go test -coverpkg=./internal/... -coverprofile=coverage.out ./...
```

### 11. Benchmarks

```sh
# Run all benchmarks
go test -bench=. ./...

# Run specific benchmark
go test -bench=BenchmarkParse -benchmem ./internal/parser

# Compare runs (install benchstat first)
go test -bench=. -count=10 ./... > old.txt
# make changes
go test -bench=. -count=10 ./... > new.txt
benchstat old.txt new.txt
```

## Code Quality

### 12. Format code

```sh
# Format all files (writes changes in place)
gofmt -w .

# Using go fmt (formats entire packages)
go fmt ./...
```

### 13. Vet — catch common mistakes

```sh
go vet ./...
```

Catches: printf format mismatches, unreachable code, bad struct tags, mutex copy, etc.

### 14. Generate code

```sh
# Run all //go:generate directives in the module
go generate ./...
```

Common with: mockery, stringer, protobuf, sqlc, ent.

## Profiling & Debugging

### 15. CPU and memory profiling

```sh
# Generate a CPU profile from tests
go test -cpuprofile=cpu.out -bench=. ./...
go tool pprof cpu.out

# Generate a memory profile
go test -memprofile=mem.out -bench=. ./...
go tool pprof mem.out

# Inside pprof interactive mode
# top10        — show top 10 functions
# web          — open flame graph in browser
# list funcName — show annotated source
```

### 16. Profile a running service (net/http/pprof)

```sh
# Fetch 30s CPU profile from a running service
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Heap profile
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine dump
curl http://localhost:6060/debug/pprof/goroutine?debug=2
```

### 17. Trace execution

```sh
go test -trace=trace.out ./...
go tool trace trace.out
```

## Documentation

### 18. View docs from terminal

```sh
# Package-level docs
go doc fmt

# Specific function
go doc fmt.Println

# Unexported symbols too
go doc -all fmt

# Start local doc server
go doc -http=:6060
```

## Environment & Troubleshooting

### 19. Check Go environment

```sh
# Print all Go env vars
go env

# Specific variable
go env GOPATH
go env GOBIN

# Override for current shell
export GOBIN=$HOME/bin
```

### 20. Clean build cache

```sh
# Remove build cache
go clean -cache

# Remove test cache (forces tests to re-run)
go clean -testcache

# Remove module download cache
go clean -modcache
```

## Workspaces (multi-module)

### 21. Manage workspaces

```sh
# Initialize a workspace
go work init ./service-a ./service-b

# Add a module to the workspace
go work use ./service-c

# Sync workspace deps
go work sync
```

Use workspaces when developing multiple modules locally that depend on each other.
