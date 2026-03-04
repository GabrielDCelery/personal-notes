# Publishing Go Modules

## Quick Reference

| Task                | Command / convention                       |
| ------------------- | ------------------------------------------ |
| Release a version   | `git tag v1.2.3 && git push origin v1.2.3` |
| Major version (v2+) | Change module path to `.../v2`             |
| Retract a version   | Add `retract` directive to go.mod          |
| Check on proxy      | `https://pkg.go.dev/your/module`           |

## Tagging Releases

### 1. Basic release

```sh
# Tag and push
git tag v1.0.0
git push origin v1.0.0

# The Go proxy indexes the version automatically
# Users can now: go get github.com/yourorg/mylib@v1.0.0
```

### 2. Semver rules

```
v1.2.3
│ │ │
│ │ └── Patch: bug fixes, no API changes
│ └──── Minor: new features, backwards compatible
└────── Major: breaking changes
```

Pre-release and metadata:

```sh
git tag v1.0.0-rc.1      # pre-release
git tag v1.0.0-beta.2     # pre-release
```

Pre-release versions are not selected by `@latest` — users must request them explicitly.

### 3. Tagging submodules (monorepo)

```
myrepo/
├── go.mod          # github.com/yourorg/myrepo
├── sdk/
│   └── go.mod      # github.com/yourorg/myrepo/sdk
```

```sh
# Root module
git tag v1.2.0

# Submodule — prefix tag with directory path
git tag sdk/v1.0.0
git push origin v1.2.0 sdk/v1.0.0
```

## Major Versions (v2+)

### 4. Breaking changes require new major version path

```
# go.mod
module github.com/yourorg/mylib/v2

go 1.22
```

Users import the new major version explicitly:

```go
import "github.com/yourorg/mylib/v2"
```

### 5. Directory-based approach (alternative)

```
mylib/
├── go.mod       # module github.com/yourorg/mylib (v0/v1)
├── v2/
│   └── go.mod   # module github.com/yourorg/mylib/v2
```

Most projects use the major-version-in-path approach (change `go.mod`) rather than subdirectories.

### 6. Tag a v2 release

```sh
# After updating go.mod to .../v2
git tag v2.0.0
git push origin v2.0.0
```

## Retracting Versions

### 7. Retract a broken release

```go
// go.mod
module github.com/yourorg/mylib

go 1.22

// Security vulnerability in JSON parsing
retract v1.3.0

// Broken between these versions
retract [v1.4.0, v1.4.3]
```

```sh
# Must publish a new version containing the retract directive
git tag v1.3.1
git push origin v1.3.1
```

Users on retracted versions see a warning when running `go get` or `go list`.

## Module Proxy & Discovery

### 8. How the Go proxy works

```
go get github.com/yourorg/mylib@v1.2.0
    │
    ▼
proxy.golang.org  ← caches module, serves to users
    │
    ▼
sum.golang.org    ← verifies checksum for supply chain security
    │
    ▼
pkg.go.dev        ← generates documentation automatically
```

After tagging and pushing, the module appears on `pkg.go.dev` within minutes (or on first `go get`).

### 9. Force proxy to index your module

```sh
# Trigger indexing
GOPROXY=proxy.golang.org go list -m github.com/yourorg/mylib@v1.2.0
```

### 10. Check your module on pkg.go.dev

```
https://pkg.go.dev/github.com/yourorg/mylib
https://pkg.go.dev/github.com/yourorg/mylib@v1.2.0
```

Write good doc comments — they render directly on pkg.go.dev:

```go
// Package mylib provides utilities for parsing configuration files.
//
// It supports JSON, YAML, and TOML formats.
package mylib
```

## Checklist Before Publishing

### 11. Pre-release checklist

```sh
# Ensure tests pass
go test -race ./...

# Vet for common mistakes
go vet ./...

# Tidy dependencies
go mod tidy

# Verify no replace directives left in go.mod
grep "replace" go.mod  # should return nothing

# Tag and push
git tag v1.0.0
git push origin v1.0.0
```

Never publish a module with `replace` directives — consumers can't use them.
