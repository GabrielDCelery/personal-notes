# Lesson 09: Module System & Toolchain

Go modules replaced `GOPATH`-based development and gave Go a first-class dependency management system. Understanding `go.mod`, `go.sum`, major version suffixes, and the broader toolchain is expected from senior Go developers — especially in interviews at companies where you'd own a non-trivial Go codebase.

## `go.mod`: The Module Definition

Every Go module has a `go.mod` file at its root. It's the source of truth for the module's identity and dependencies.

```go
module github.com/acme/myservice

go 1.22

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/jmoiron/sqlx v1.3.5
    golang.org/x/exp v0.0.0-20231006
)

require (
    // indirect: dependencies of dependencies
    github.com/bytedance/sonic v1.10.1 // indirect
    github.com/go-playground/validator/v10 v10.15.5 // indirect
)

replace github.com/acme/internal => ../internal // local replace directive

exclude github.com/badlibrary/v1 v1.0.0 // never use this version
```

### Key Directives

| Directive | Purpose                                                               |
| --------- | --------------------------------------------------------------------- |
| `module`  | Module path (must match directory structure for imports)              |
| `go`      | Minimum Go version; also controls language semantics                  |
| `require` | Direct and indirect dependencies with exact versions                  |
| `replace` | Substitute a module with a different path or local directory          |
| `exclude` | Prevent a specific version from being selected                        |
| `retract` | Mark versions as withdrawn (authors use this to retract bad releases) |

### The `go` Directive Matters

The `go` directive isn't just documentation — it changes language semantics:

```
go 1.21   # loop variable semantics, slices.Sort, maps.Keys etc.
go 1.22   # range over integers, loop variable is per-iteration
go 1.23   # iterator protocol (range over func)
```

If you set `go 1.18` in `go.mod`, you can't use language features from 1.21 even if the compiler is newer.

## `go.sum`: The Integrity Database

`go.sum` records cryptographic hashes of every module version and its `go.mod`:

```
github.com/gin-gonic/gin v1.9.1 h1:4idEAncQnU5cB7BeOkPtxjfCSye0AAm1R0RVIqJ+Jmg=
github.com/gin-gonic/gin v1.9.1/go.mod h1:hPrL7YrpYKXt5YId3A/Tnip5kqbEAP+KLuI3SUcPTeU=
```

**Two entries per module version**:

- `h1:...` — hash of the module's source tree (the zip file)
- `h1:.../go.mod` — hash of just the `go.mod` file

**Purpose**: Ensures that a given version of a module always has the same content. If a maintainer modifies a tag on GitHub, your build fails because the hash won't match. This prevents supply chain attacks via tag tampering.

**Never edit `go.sum` manually.** Always commit both `go.mod` and `go.sum` to version control.

## Semantic Versioning and Version Selection

Go modules follow semantic versioning (`vMAJOR.MINOR.PATCH`):

- `MAJOR`: Breaking changes
- `MINOR`: New backwards-compatible features
- `PATCH`: Bug fixes

### Minimum Version Selection (MVS)

When multiple dependencies require different versions of the same module, Go uses **Minimum Version Selection**: it picks the minimum version that satisfies all requirements.

```
Module A requires gin v1.9.0
Module B requires gin v1.9.1
→ Go selects gin v1.9.1 (minimum that satisfies both)
```

MVS is **reproducible**: the same `go.mod` always selects the same versions. It doesn't automatically upgrade to newer versions — you must explicitly run `go get` to upgrade.

```sh
go get github.com/gin-gonic/gin@latest   # upgrade to latest
go get github.com/gin-gonic/gin@v1.9.1  # specific version
go get -u ./...                          # upgrade all deps
go mod tidy                              # remove unused deps, add missing ones
```

## Major Version Suffixes

Breaking changes require a new major version. In Go modules, major versions ≥ 2 must be reflected in the **import path**.

```go
// v1 (original)
import "github.com/acme/mylib"

// v2 (breaking changes - different import path)
import "github.com/acme/mylib/v2"

// v3
import "github.com/acme/mylib/v3"
```

And in `go.mod`:

```
require github.com/acme/mylib/v2 v2.1.0
```

**Why**: Two major versions of the same library are considered **different modules**. Your code can import both simultaneously (e.g., for migration), and they don't conflict.

```go
import (
    mylibv1 "github.com/acme/mylib"
    mylibv2 "github.com/acme/mylib/v2"
)
```

**The module path must include `/v2`**: The `go.mod` of a v2 module must start with:

```
module github.com/acme/mylib/v2
```

## Go Workspaces (`go.work`)

Workspaces (Go 1.18+) allow you to work on multiple modules simultaneously without `replace` directives.

**The problem they solve**: You're developing `myapp` and `mylib` together. Changes to `mylib` need to be visible in `myapp` before you release. The old solution was `replace github.com/acme/mylib => ../mylib` in `go.mod` — which you'd have to remove before committing.

**Workspace solution**:

```
workspace/
├── go.work         # workspace definition
├── myapp/
│   └── go.mod
└── mylib/
    └── go.mod
```

```go
// go.work
go 1.22

use (
    ./myapp
    ./mylib
)
```

```sh
go work init ./myapp ./mylib   # create go.work
go work use ./anothermodule    # add a module
```

With `go.work` in place, `myapp`'s `import "github.com/acme/mylib"` automatically uses the local `./mylib` directory. No edits to `go.mod` needed. **Don't commit `go.work` to version control** (add to `.gitignore`) — it's a local development aid.

## Build Tags

Build tags (also called build constraints) control which files are included in a build.

### New Syntax (Go 1.17+)

```go
//go:build linux && amd64
// (optionally followed by old-style: // +build linux amd64)

package main
```

```go
//go:build !windows   // exclude on Windows
//go:build go1.21     // only on Go 1.21+
//go:build integration // custom tag
```

### Common Uses

```go
// platform-specific implementation
//go:build windows

package fs

func tempDir() string { return os.Getenv("TEMP") }
```

```go
// integration tests (not run by default)
//go:build integration

package myservice_test

func TestAgainstRealDB(t *testing.T) { ... }
```

```sh
go test -tags integration ./...  # include files with integration tag
go build -tags linux             # specific platform tag
```

### File Naming Convention (Alternative to Tags)

Go also respects OS/arch suffixes in filenames:

```
file_linux.go         # only on Linux
file_windows.go       # only on Windows
file_linux_amd64.go   # only on Linux/amd64
```

No build tag needed — the naming convention is sufficient.

## `go:generate`

`//go:generate` embeds code generation commands in source files:

```go
//go:generate stringer -type=Status
//go:generate mockgen -source=interfaces.go -destination=mocks/mocks.go

type Status int
const (
    Active Status = iota
    Inactive
    Deleted
)
```

```sh
go generate ./...  # runs all //go:generate commands
```

Common generators:

- `stringer`: generates `String() string` for `iota` enums
- `mockgen`: generates mocks for interfaces (gomock)
- `protoc`: generates Go code from protobuf definitions
- `sqlc`: generates type-safe Go code from SQL queries

**Convention**: Run `go generate` manually (or in CI before tests), and commit the generated files. Generated files have a header: `// Code generated by stringer; DO NOT EDIT.`

## Static Analysis Tools

### `go vet`

Built-in, always run:

```sh
go vet ./...
```

Catches: mutex copies, incorrect `Printf` format strings, unreachable code, `sync.WaitGroup` misuse, shadowed imports.

### `staticcheck`

The most comprehensive Go static analyzer ([staticcheck.io](https://staticcheck.io)):

```sh
go install honnef.co/go/tools/cmd/staticcheck@latest
staticcheck ./...
```

Catches: deprecated API usage, incorrect time format strings, unused code, performance improvements, correctness issues that `go vet` misses.

### `golangci-lint`

Meta-linter that runs many linters simultaneously:

```sh
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
golangci-lint run ./...
```

Includes `staticcheck`, `errcheck` (unchecked errors), `gosimple`, `unused`, and dozens more. Configure in `.golangci.yml`.

### `gopls`

The official Go language server for IDEs. Not a linter you run manually, but powers your editor's autocompletion, go-to-definition, hover docs, and real-time diagnostics.

### Recommended CI Pipeline

```yaml
# .github/workflows/ci.yml
- run: go vet ./...
- run: staticcheck ./...
- run: go test -race -count=1 ./...
- run: go test -bench=. -benchmem ./... # optional, catch regressions
```

## Useful `go` Commands Reference

```sh
# Module management
go mod init github.com/org/repo  # initialize new module
go mod tidy                       # sync go.mod and go.sum with imports
go mod download                   # download all deps to cache
go mod verify                     # verify hashes match go.sum
go mod graph                      # print module dependency graph
go mod why -m github.com/foo/bar  # why is this module needed?

# Dependency management
go get github.com/foo/bar@v1.2.3  # add or upgrade
go get github.com/foo/bar@none    # remove
go get -u ./...                   # upgrade all

# Build and run
go build ./...                    # build all packages
go run .                          # run main package
go install github.com/tool@latest # install a tool to $GOPATH/bin

# Code quality
go vet ./...
go test ./...
go test -race ./...
go test -bench=. -benchmem ./...
go build -gcflags="-m" ./...     # escape analysis

# Formatting
gofmt -w .
goimports -w .                    # gofmt + organize imports
```

## Hands-On Exercise 1: Reading `go.mod`

Analyze this `go.mod` and answer the questions:

```
module github.com/acme/api-gateway

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/acme/auth-lib/v3 v3.2.0
    golang.org/x/crypto v0.14.0
)

require (
    github.com/bytedance/sonic v1.10.1 // indirect
    golang.org/x/net v0.17.0 // indirect
    golang.org/x/sys v0.13.0 // indirect
)

replace github.com/acme/auth-lib/v3 => ../auth-lib
```

**Questions**:

1. What import path would you use to import `auth-lib`?
2. Why is `auth-lib` in a `replace` directive? What does this indicate?
3. What are the indirect dependencies and why are they listed?
4. What happens if you run `go mod tidy` and the `replace` directive is left in?

<details>
<summary>Solution</summary>

1. **Import path**: `github.com/acme/auth-lib/v3` — major version suffix must be included in the import path for v3.

2. **The `replace` directive** substitutes the module with a local directory (`../auth-lib`). This indicates the developer is working on both modules simultaneously. It's a workspace-like setup using the older pattern. This should NOT be committed to a shared repository — it will break other developers' builds if they don't have `../auth-lib` at that relative path. The modern alternative is `go.work`.

3. **Indirect dependencies** are transitive dependencies — dependencies of `gin`, `auth-lib`, or `golang.org/x/crypto`. They're listed in `go.mod` when they're needed for the build but not directly imported in this module. `go mod tidy` adds and removes them automatically.

4. **`go mod tidy` with `replace`**: `go mod tidy` will keep the `replace` directive — it doesn't remove replace directives. However, it will verify that all imports in the code match the modules in `require`, and will clean up unused indirect dependencies. The `replace` for `auth-lib` stays.

</details>

## Hands-On Exercise 2: Module Upgrade Plan

You're upgrading a dependency from v1 to v2. The library changed its import path from `github.com/foo/cache` to `github.com/foo/cache/v2`. Describe the steps.

<details>
<summary>Solution</summary>

**Step 1**: Check the v2 changelog for breaking changes.

**Step 2**: Add the v2 dependency:

```sh
go get github.com/foo/cache/v2@latest
```

**Step 3**: Update all imports in your code:

```sh
# Find files importing the old path
grep -r "github.com/foo/cache" --include="*.go" .

# Update each file's imports
# Before: import "github.com/foo/cache"
# After:  import "github.com/foo/cache/v2"
```

**Step 4**: Fix any API changes (breaking changes between v1 and v2).

**Step 5**: Run tests:

```sh
go test ./...
go build ./...
```

**Step 6**: Remove the old v1 dependency if no longer used:

```sh
go mod tidy
```

**Step 7**: Verify `go.mod` has the v2 entry and v1 is gone:

```
require github.com/foo/cache/v2 v2.x.x
```

**Note**: You can use both v1 and v2 simultaneously during migration:

```go
import (
    cachev1 "github.com/foo/cache"        // old code still using v1
    cachev2 "github.com/foo/cache/v2"     // new code using v2
)
```

</details>

## Interview Questions

### Q1: What is the difference between `go.mod` and `go.sum`, and why are both needed?

A fundamental question about the module system's design goals.

<details>
<summary>Answer</summary>

**`go.mod`**: Defines the module's identity and its dependencies with version constraints. It's the human-readable, manually-edited file. `require` entries specify the minimum acceptable version; the actual version selected may be higher (MVS).

**`go.sum`**: Records the cryptographic hashes of every dependency version's source tree and `go.mod` file. It's automatically maintained by the `go` tool. Never edit it manually.

**Why both are needed**:

- `go.mod` alone tells you _which_ version to use, but not _what_ that version contains
- `go.sum` ensures that the version you download today is byte-for-byte identical to the version you downloaded when you first added it

Without `go.sum`, a malicious actor could replace a tagged release on GitHub with different code. The hash mismatch would fail the build. With `go.sum`, the build is reproducible and tamper-evident.

Both files must be committed to version control. `go.sum` should never be `.gitignore`d — doing so removes the tamper-detection guarantee.

</details>

### Q2: Why do Go modules require a `/v2` suffix in the import path for major versions?

A design question — interviewers use this to see if you understand the trade-offs.

<details>
<summary>Answer</summary>

The rule is called "import compatibility rule": if a package's import path doesn't change, it must be backwards compatible.

**The problem without `/v2`**: If `github.com/foo/lib` releases a v2 with breaking changes but keeps the same import path, any dependency that imports `github.com/foo/lib` would get different behaviour depending on which version is selected. Two packages in the same binary could import different major versions but be unable to pass values between them (since `v1.User` and `v2.User` would be different types).

**With `/v2` in the import path**:

- `github.com/foo/lib` and `github.com/foo/lib/v2` are treated as completely separate modules
- A binary can import both simultaneously with no conflicts
- Types from v1 and v2 are distinct — you can't accidentally mix them
- MVS can satisfy both independently

**Practical implication**: Major version bumps in Go are explicit in the code. You can't accidentally get a breaking API change because the import path itself changes when the API breaks. This enables gradual migration — old code uses `lib`, new code uses `lib/v2`.

</details>

### Q3: What are Go workspaces and when do you use them?

Tests knowledge of modern Go development workflows.

<details>
<summary>Answer</summary>

Go workspaces (`go.work`) allow you to simultaneously develop multiple interdependent modules without modifying their `go.mod` files. Introduced in Go 1.18.

**The problem they solve**: When you're developing `myapp` (which imports `mylib`) and `mylib` simultaneously, you need `myapp` to use your local, in-progress version of `mylib`. The old solution was adding a `replace` directive to `myapp/go.mod` pointing to `../mylib` — which you'd have to revert before committing.

**Workspace approach**:

```sh
# In the parent directory containing both modules:
go work init ./myapp ./mylib
```

Creates `go.work`:

```
go 1.22
use ./myapp
use ./mylib
```

Now building `myapp` automatically uses the local `mylib`. No changes to either `go.mod`.

**Key points**:

- `go.work` is a local file — add it to `.gitignore`
- Works across all `go` commands: `go build`, `go test`, `go run`
- `GOWORK=off` disables workspace mode for a command
- You can reference modules outside the repo (not just siblings)

**vs `replace` directive**: Workspaces are cleaner for local development — no risk of accidentally committing `replace` directives. Use `replace` only for permanent redirects (like forking a dead library).

</details>

### Q4: What static analysis tools should every Go project run, and what does each catch?

A practical DevOps/quality question — tests whether you build quality into your development process.

<details>
<summary>Answer</summary>

**Mandatory (zero overhead, always run)**:

- **`go vet`**: Built-in. Catches: mutex copies, wrong Printf verb/arg count, unreachable code, misuse of `sync` primitives. Run as part of `go test` by default.

- **`go test -race`**: Race detector. Catches concurrent data races that cause intermittent, hard-to-reproduce bugs. ~5-10x overhead, appropriate for CI test runs.

**Strongly recommended**:

- **`staticcheck`**: Comprehensive static analysis. Catches: deprecated APIs, inefficient code patterns, incorrect string format usage, unused code, real correctness issues. More thorough than `go vet`.

- **`golangci-lint`**: Meta-linter running 50+ linters including `staticcheck`, `errcheck` (unchecked errors), `gosimple`, `unused`. Configure via `.golangci.yml` to tune which linters run.

- **`gofmt` / `goimports`**: Format checking. `goimports` also manages import organization. Run as a CI check: `gofmt -l .` returns non-zero if any file needs formatting.

**Optional but valuable**:

- **`govulncheck`**: Official Go vulnerability scanner. Checks your imports against the Go vulnerability database.

- **`deadcode`**: Finds unreachable code.

**Minimal CI pipeline**:

```yaml
go vet ./...
staticcheck ./...
go test -race ./...
```

</details>

## Key Takeaways

1. **`go.mod`**: Module identity, Go version, and dependency requirements — commit it.
2. **`go.sum`**: Cryptographic hashes guaranteeing reproducible builds and tamper detection — commit it, never edit manually.
3. **MVS**: Go selects the minimum version satisfying all constraints — no surprise upgrades, fully reproducible.
4. **`/v2` in import path**: Major version changes must update the import path — the import compatibility rule enforces this.
5. **`go mod tidy`**: Syncs `go.mod` and `go.sum` with actual imports — run before every commit.
6. **`replace` directive**: For local development or permanent forks — don't commit if it points to a local path.
7. **`go.work`**: Modern replacement for `replace` in multi-module development — `.gitignore` it.
8. **Build tags**: `//go:build platform` or `//go:build tag` for conditional compilation.
9. **`go vet` + `staticcheck`**: Non-negotiable in CI — they catch real bugs, not just style.
10. **`go:generate`**: Embed generation commands in source; commit generated files.

---

## Series Complete

You've covered the nine core topics for senior Go interviews:

| Lesson                                              | Topic                          |
| --------------------------------------------------- | ------------------------------ |
| [01](lesson-01-interfaces-and-type-system.md)       | Interfaces & the Type System   |
| [02](lesson-02-error-handling-patterns.md)          | Error Handling Patterns        |
| [03](lesson-03-goroutines-and-scheduler.md)         | Goroutines & the Scheduler     |
| [04](lesson-04-sync-primitives-and-patterns.md)     | Sync Primitives & Patterns     |
| [05](lesson-05-memory-model-and-escape-analysis.md) | Memory Model & Escape Analysis |
| [06](lesson-06-context-and-cancellation.md)         | Context & Cancellation         |
| [07](lesson-07-generics.md)                         | Generics                       |
| [08](lesson-08-testing-and-benchmarking.md)         | Testing & Benchmarking         |
| [09](lesson-09-module-system-and-toolchain.md)      | Module System & Toolchain      |

**Channels** are covered separately in the [`go/channels/`](../channels/) series.
