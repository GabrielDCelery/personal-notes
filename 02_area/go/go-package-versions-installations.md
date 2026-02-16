---
title: Go package versions and installations
---

# How to figure out which package version to install

1. Go Package Discovery Site (easiest)

Visit `https://pkg.go.dev/` and search for the package:

- golangci-lint: `https://pkg.go.dev/github.com/golangci/golangci-lint`
- migrate: `https://pkg.go.dev/github.com/golang-migrate/migrate`

The page shows:

- Correct module path (at the top)
- All published versions (Versions tab)
- Import paths for installable commands

2. GitHub Releases Page

Check the releases page directly:

- `https://github.com/golangci/golangci-lint/releases`
- `https://github.com/golang-migrate/migrate/releases`

Look at the tag names (e.g., v1.60.3, v4.19.1)

3. Command Line - List Versions

List all available versions

```sh
go list -m -versions github.com/golangci/golangci-lint
```

Or check what @latest resolves to

```sh
go list -m github.com/golangci/golangci-lint@latest
```

# Finding the correct install path

For CLI tools, check the repo structure:

**Clone and look for main.go**

```sh
git clone https://github.com/golangci/golangci-lint
find . -name main.go
```

**Usually in: ./cmd/<toolname>/main.go**

The install path is: <module-path>/cmd/<toolname>
