# Go Version Management

## Quick Reference

| Method                         | Best for                       |
| ------------------------------ | ------------------------------ |
| `toolchain` in go.mod          | Per-project version pinning    |
| `go install golang.org/dl/...` | Multiple versions side by side |
| `GOTOOLCHAIN`                  | Override toolchain selection   |

## Toolchain Directive (Go 1.21+)

### 1. Pin Go version in go.mod

```
module github.com/yourorg/myapp

go 1.22.0
toolchain go1.22.4
```

- `go` — minimum required version
- `toolchain` — exact version to use when available

When someone runs `go build`, Go automatically downloads the specified toolchain if needed.

### 2. Set toolchain via CLI

```sh
# Update go.mod toolchain directive
go get toolchain@go1.22.4

# Update minimum Go version
go get go@1.22.0
```

### 3. GOTOOLCHAIN environment variable

```sh
# Always use this exact version
go env -w GOTOOLCHAIN=go1.22.4

# Use local version, allow auto-download if go.mod needs newer
go env -w GOTOOLCHAIN=local+auto

# Never auto-download — fail if local version is too old
go env -w GOTOOLCHAIN=local
```

| Value        | Behaviour                                   |
| ------------ | ------------------------------------------- |
| `auto`       | Download toolchain if go.mod needs newer    |
| `local`      | Use installed version only, fail if too old |
| `local+auto` | Prefer local, download if needed            |
| `go1.22.4`   | Always use this specific version            |

## Multiple Versions Side by Side

### 4. Install additional Go versions

```sh
# Install Go 1.21.6 alongside your main install
go install golang.org/dl/go1.21.6@latest
go1.21.6 download

# Use it
go1.21.6 version
go1.21.6 build ./...
go1.21.6 test ./...
```

### 5. Switch default version

```sh
# Check current
go version

# If managed via OS package manager
# Ubuntu/Debian
sudo apt install golang-1.22
sudo update-alternatives --config go

# macOS
brew install go@1.22
brew link go@1.22 --force
```

## CI Version Pinning

### 6. GitHub Actions

```yaml
- uses: actions/setup-go@v5
  with:
    go-version-file: go.mod # reads from go.mod toolchain directive
    cache: true

# Or pin explicitly
- uses: actions/setup-go@v5
  with:
    go-version: "1.22.4"
    cache: true
```

### 7. GitLab CI

```yaml
image: golang:1.22.4-alpine

build:
  script:
    - go build ./...
```

### 8. Check version compatibility in code

```go
//go:build go1.22
```

Build constraint ensures the file only compiles with Go 1.22+.
