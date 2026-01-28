---
title: How to construct a README
tags: git, documentation, best-practices
---

# README Standardization Guide

Goal: Document 200 repositories in one month with consistent, useful READMEs.

## Core Principles

1. **Standardize the structure** - Every README follows the same template so developers know exactly where to look
2. **Keep it scannable** - Use headers, bullet points, and tables. No walls of text
3. **Link, don't duplicate** - Point to external docs (runbooks, wikis) rather than maintaining info in multiple places
4. **Focus on the "what now"** - Developers land on a repo asking: "What is this? How do I run it? How do I deploy it?"

## Recommended Template

````markdown
# Project Name

One-sentence description of what this service/library does.

## Overview

- **Purpose**: What problem does this solve?
- **Owners**: Team or individuals responsible
- **Status**: Active / Maintenance / Deprecated
- **Related Repos**: Links to repos this works with

## Quick Start

```sh
# Clone and setup
git clone <repo>
cd <repo>
make setup # or npm install, etc.

# Run locally
make run # or docker-compose up, etc.
```

## Architecture

Brief description of how it works. Include a simple diagram if helpful.

- **Language/Framework**: Go 1.21 / Python 3.11 / etc.
- **Database**: PostgreSQL / Redis / None
- **Message Queue**: Kafka / RabbitMQ / None
- **External APIs**: List any third-party services called

## Configuration

| Variable       | Description                | Required | Default |
| -------------- | -------------------------- | -------- | ------- |
| `DATABASE_URL` | Postgres connection string | Yes      | -       |
| `LOG_LEVEL`    | debug/info/warn/error      | No       | info    |

Secrets are stored in: [Vault path / AWS Secrets Manager / etc.]

## Development

```sh
# Run tests
make test

# Run linter
make lint
```

## Deployment

- **Production**: Deployed via [ArgoCD / Jenkins / GitHub Actions] to [K8s cluster / AWS ECS / etc.]
- **Environments**: dev / staging / prod
- **Runbook**: [Link to ops runbook if exists]

## API Documentation

[Link to OpenAPI spec / Swagger / internal docs]
````

## Examples

### Example 1: Internal API Service

````markdown
# user-service

Handles user authentication, profile management, and session tokens.

## Overview

- **Purpose**: Central auth service for all customer-facing apps
- **Owners**: Platform Team (@platform-team)
- **Status**: Active
- **Related Repos**: [api-gateway](../api-gateway), [notification-service](../notification-service)

## Quick Start

```sh
cp .env.example .env
docker-compose up -d postgres redis
go run cmd/server/main.go
```

Server runs at `http://localhost:8080`

## Architecture

- **Language/Framework**: Go 1.21 / Chi router
- **Database**: PostgreSQL (users, sessions)
- **Cache**: Redis (session tokens)
- **External APIs**: SendGrid (email verification)

## Configuration

| Variable           | Description         | Required |
| ------------------ | ------------------- | -------- |
| `DATABASE_URL`     | Postgres connection | Yes      |
| `REDIS_URL`        | Redis connection    | Yes      |
| `SENDGRID_API_KEY` | Email provider      | Yes      |
| `JWT_SECRET`       | Token signing key   | Yes      |

Secrets in Vault: `secret/prod/user-service`

## Deployment

Deployed via ArgoCD to `platform` namespace.
Runbook: [Confluence - User Service Ops](https://...)
````

### Example 2: Shared Library

````markdown
# go-common

Shared utilities for all Go services: logging, tracing, config loading.

## Overview

- **Purpose**: Reduce boilerplate across Go microservices
- **Owners**: Platform Team
- **Status**: Active
- **Used by**: user-service, order-service, inventory-service

## Installation

```sh
go get github.com/yourcompany/go-common@latest
```

## Usage

```go
import "github.com/yourcompany/go-common/logging"

logger := logging.New("my-service")
logger.Info("server started", "port", 8080)
```

## Modules

- `logging` - Structured logging with trace ID propagation
- `config` - Load config from env vars and Vault
- `middleware` - HTTP middleware for auth, tracing, metrics

## Development

```sh
make test
make lint
```

## Versioning

Uses semantic versioning. Breaking changes require major version bump.
````

### Example 3: Infrastructure / Deployment Repo

````markdown
# infra-kubernetes

Kubernetes manifests and Helm charts for all services.

## Overview

- **Purpose**: GitOps source of truth for cluster state
- **Owners**: SRE Team
- **Status**: Active

## Structure

```
├── base/           # Shared resources (namespaces, RBAC)
├── charts/         # Helm charts
├── environments/
│   ├── dev/
│   ├── staging/
│   └── prod/
└── services/       # Per-service manifests
```

## Making Changes

1. Create branch from `main`
2. Modify manifests in `environments/<env>/`
3. Open PR - ArgoCD will show diff
4. Merge - ArgoCD auto-syncs within 3 minutes

## Adding a New Service

```sh
./scripts/scaffold-service.sh my-new-service
```

Then edit `services/my-new-service/values.yaml`

## Secrets

Secrets are NOT in this repo. They're managed via External Secrets Operator pulling from Vault.
````

## Rollout Strategy

For 200 repos in a month:

1. **Week 1**: Create the template, document 10 critical repos as examples
2. **Week 2-3**: Have each team document their own repos (they know them best)
3. **Week 4**: Review, fill gaps, mark deprecated repos clearly

## Enforcement

Consider adding a CI check that fails if a repo is missing a README or required sections. Example with GitHub Actions:

```yaml
# .github/workflows/readme-check.yml
name: README Check
on: [pull_request]
jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Check README exists
        run: test -f README.md
      - name: Check required sections
        run: |
          grep -q "## Quick Start" README.md
          grep -q "## Configuration" README.md
          grep -q "## Deployment" README.md
```
