# Go Modules

## Quick Reference

| Use case              | Command                           |
| --------------------- | --------------------------------- |
| Init module           | `go mod init module/path`         |
| Add dependency        | `go get pkg@version`              |
| Add to latest         | `go get pkg@latest`               |
| Remove unused deps    | `go mod tidy`                     |
| Download deps         | `go mod download`                 |
| Vendor dependencies   | `go mod vendor`                   |
| Show dependency graph | `go mod graph`                    |
| Verify checksums      | `go mod verify`                   |
| Edit go.mod           | `go mod edit -require pkg@v1.2.3` |
| Why is dep included   | `go mod why pkg`                  |

## go.mod

### 1. Basic go.mod

```
module github.com/myorg/myapp

go 1.25

require (
    github.com/gin-gonic/gin v1.9.1
    go.uber.org/zap v1.27.0
)
```

### 2. Version queries

```sh
go get pkg@v1.2.3          # exact version
go get pkg@latest          # latest stable
go get pkg@v1              # latest v1.x.x
go get pkg@master          # tip of branch
go get pkg@commit-hash     # specific commit
```

### 3. Update dependencies

```sh
go get -u pkg              # update pkg to latest minor/patch
go get -u ./...            # update all direct deps
go get pkg@none            # remove dependency
go mod tidy                # clean up after changes
```

## Replace Directive

### 4. Local development override

```
replace github.com/myorg/shared => ../shared
```

### 5. Fork override

```
replace github.com/original/pkg => github.com/myfork/pkg v1.0.0
```

### 6. Replace via command

```sh
go mod edit -replace github.com/original/pkg=../local-pkg
go mod edit -dropreplace github.com/original/pkg
```

## Vendoring

### 7. Vendor dependencies

```sh
go mod vendor              # copy deps to vendor/
go build -mod=vendor       # build using vendor/
```

Set in go.env or env var:

```sh
GOFLAGS=-mod=vendor        # always use vendor
```

## go.sum

### 8. What go.sum does

- Contains cryptographic checksums for every dependency version
- Ensures reproducible builds
- Commit it to version control
- Never edit manually

```sh
go mod verify              # check go.sum matches downloaded modules
go mod download            # download all deps and populate go.sum
```

## Workspaces (multi-module)

### 9. go.work for monorepos

```sh
go work init ./api ./shared ./worker
```

```
// go.work
go 1.25

use (
    ./api
    ./shared
    ./worker
)
```

```sh
go work sync               # sync go.work with module go.mod files
```

> go.work is for local development. Don't commit it unless the whole team uses it.

## Useful Commands

### 10. Inspect

```sh
go list -m all             # list all dependencies
go list -m -versions pkg   # list available versions
go mod why pkg             # explain why pkg is needed
go mod graph               # full dependency graph
```
