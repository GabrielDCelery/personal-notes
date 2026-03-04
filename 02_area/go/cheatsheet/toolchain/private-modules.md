# Go Private Modules

## Quick Reference

| Variable       | Purpose                                      |
| -------------- | -------------------------------------------- |
| `GOPRIVATE`    | Skip proxy and checksum DB for these modules |
| `GONOSUMDB`    | Skip checksum DB only                        |
| `GONOSUMCHECK` | Skip checksum verification only              |
| `GOPROXY`      | Module proxy URL (default: proxy.golang.org) |

## Environment Setup

### 1. Mark modules as private

```sh
# Single org
go env -w GOPRIVATE=github.com/yourorg/*

# Multiple orgs
go env -w GOPRIVATE=github.com/yourorg/*,gitlab.com/internal/*

# Entire domain (self-hosted GitLab, Gitea, etc.)
go env -w GOPRIVATE=git.internal.company.com/*
```

`GOPRIVATE` sets both `GONOSUMDB` and `GONOPROXY` — usually this is all you need.

### 2. Corporate proxy setup

```sh
# Use corporate proxy with fallback to direct
go env -w GOPROXY=https://goproxy.company.com,https://proxy.golang.org,direct

# Private modules bypass proxy, everything else uses it
go env -w GOPRIVATE=github.com/yourorg/*
go env -w GOPROXY=https://proxy.golang.org,direct
```

## Git Authentication

### 3. SSH (developer machines)

```gitconfig
# ~/.gitconfig
[url "ssh://git@github.com/"]
    insteadOf = https://github.com/
```

This rewrites HTTPS module paths to SSH so `go get` uses your SSH key.

### 4. HTTPS with token (CI / Docker)

```sh
# ~/.netrc — used by go get for HTTPS auth
machine github.com
login x-access-token
password ${GITHUB_TOKEN}
```

Or via git config:

```sh
git config --global url."https://${GITHUB_TOKEN}@github.com/yourorg/".insteadOf "https://github.com/yourorg/"
```

### 5. GitHub App / Deploy token (CI)

```sh
# GitHub Actions — use the built-in token
git config --global url."https://x-access-token:${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"

# GitLab CI — use CI_JOB_TOKEN
git config --global url."https://gitlab-ci-token:${CI_JOB_TOKEN}@gitlab.com/".insteadOf "https://gitlab.com/"
```

## CI Pipeline Setup

### 6. GitHub Actions example

```yaml
- name: Configure private modules
  run: |
    go env -w GOPRIVATE=github.com/yourorg/*
    git config --global url."https://x-access-token:${{ secrets.GITHUB_TOKEN }}@github.com/".insteadOf "https://github.com/"

- name: Download dependencies
  run: go mod download
```

### 7. GitLab CI example

```yaml
before_script:
  - go env -w GOPRIVATE=gitlab.com/yourorg/*
  - git config --global url."https://gitlab-ci-token:${CI_JOB_TOKEN}@gitlab.com/".insteadOf "https://gitlab.com/"
  - go mod download
```

## Troubleshooting

### 8. Verify settings

```sh
# Check current values
go env GOPRIVATE GOPROXY GONOSUMDB GONOPROXY

# Test fetching a private module
GOFLAGS=-v go get github.com/yourorg/private-lib@latest

# Debug git credential issues
GIT_TRACE=1 go get github.com/yourorg/private-lib@latest
```

### 9. Common errors

| Error                                 | Fix                                         |
| ------------------------------------- | ------------------------------------------- |
| `410 Gone` from proxy                 | Add module to `GOPRIVATE`                   |
| `terminal prompts disabled`           | Configure `.netrc` or git URL rewrite       |
| `unknown revision` for private module | Check git auth — SSH key or token missing   |
| `verifying module: checksum mismatch` | Add module to `GONOSUMCHECK` or `GOPRIVATE` |
