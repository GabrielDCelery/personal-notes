# Go Linting

## Quick Reference

| Tool              | Purpose                                |
| ----------------- | -------------------------------------- |
| `golangci-lint`   | Meta-linter, runs many linters at once |
| `go vet`          | Built-in, catches common mistakes      |
| `staticcheck`     | Advanced static analysis               |
| `gofmt` / `goimports` | Code formatting                   |

## golangci-lint

### 1. Install

```sh
# Binary (recommended — don't go install, it's unsupported)
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# macOS
brew install golangci-lint
```

### 2. Run

```sh
# Lint entire project
golangci-lint run

# Specific packages
golangci-lint run ./internal/...

# Show all issues (not just new ones)
golangci-lint run --new=false

# Fix auto-fixable issues
golangci-lint run --fix
```

### 3. Configuration (.golangci.yml)

```yaml
run:
  timeout: 5m
  go: "1.22"

linters:
  enable:
    - errcheck        # unchecked errors
    - govet           # go vet checks
    - staticcheck     # advanced static analysis
    - unused          # unused code
    - gosimple        # simplifications
    - ineffassign     # useless assignments
    - typecheck       # type checking
    - gocritic        # opinionated style checks
    - revive          # flexible linter (replaces golint)
    - misspell        # spelling mistakes in comments
    - prealloc        # suggest slice preallocation
    - nilerr          # returning nil after checking err != nil
    - errname         # error naming conventions
    - exhaustive      # exhaustive switch statements

linters-settings:
  govet:
    enable-all: true
  errcheck:
    check-type-assertions: true
    check-blank: true
  gocritic:
    enabled-tags:
      - diagnostic
      - performance
      - style
  revive:
    rules:
      - name: unexported-return
        disabled: true

issues:
  exclude-rules:
    # Test files can use dot imports
    - path: _test\.go
      linters:
        - revive
      text: "dot-imports"
    # Test files don't need error checks on everything
    - path: _test\.go
      linters:
        - errcheck
```

### 4. Useful linters by category

**Bug prevention:**

| Linter        | Catches                                    |
| ------------- | ------------------------------------------ |
| `errcheck`    | Unchecked error returns                    |
| `staticcheck` | Misuse of stdlib, deprecated funcs         |
| `nilerr`      | Returning nil instead of err after check   |
| `bodyclose`   | Unclosed HTTP response bodies              |
| `sqlclosecheck` | Unclosed SQL rows/statements             |
| `contextcheck` | Non-inherited context in goroutines       |

**Performance:**

| Linter      | Catches                         |
| ----------- | ------------------------------- |
| `prealloc`  | Slices that could be preallocated |
| `ineffassign` | Assignments that are never read |

**Style / consistency:**

| Linter      | Catches                             |
| ----------- | ----------------------------------- |
| `revive`    | Configurable style rules            |
| `gocritic`  | Opinionated suggestions             |
| `misspell`  | Typos in comments and strings       |
| `errname`   | Error vars should start with `Err`  |
| `exhaustive`| Non-exhaustive switch on enum types |

### 5. Suppress false positives

```go
// Suppress for a specific line
//nolint:errcheck
_ = writer.Close()

// Suppress with reason (preferred)
//nolint:errcheck // best-effort cleanup, error doesn't matter
_ = writer.Close()

// Suppress entire function
//nolint:cyclop // complex but readable
func handleRequest() { ... }
```

## CI Integration

### 6. GitHub Actions

```yaml
- name: golangci-lint
  uses: golangci/golangci-lint-action@v6
  with:
    version: latest
    args: --timeout=5m
```

### 7. GitLab CI

```yaml
lint:
  image: golangci/golangci-lint:latest
  script:
    - golangci-lint run --timeout=5m
```

### 8. Pre-commit hook

```sh
# .git/hooks/pre-commit or via pre-commit framework
golangci-lint run --new-from-rev=HEAD~1
```

Only lints changed code — keeps the hook fast.

## Formatting

### 9. gofmt vs goimports

```sh
# gofmt — standard formatting
gofmt -w .

# goimports — gofmt + manages import grouping
go install golang.org/x/tools/cmd/goimports@latest
goimports -w .
```

Most teams use `goimports` — it auto-adds missing imports and groups them:

```go
import (
    // stdlib
    "fmt"
    "net/http"

    // third-party
    "github.com/gin-gonic/gin"

    // internal
    "github.com/yourorg/myapp/internal/auth"
)
```
