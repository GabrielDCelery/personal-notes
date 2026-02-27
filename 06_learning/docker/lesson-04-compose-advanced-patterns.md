# Lesson 04: Compose Advanced Patterns

Beyond `docker compose up` — profiles, health-checked dependencies, secrets, override files, and DRY configurations.

## The Problem with Simple Compose Files

Most developers write a single `docker-compose.yml` that works locally but fails in CI, staging, or production because:

- Dev tooling (pgAdmin, mailhog) ends up in production
- Secrets are in plaintext environment variables
- Services start before their dependencies are actually ready (not just running)
- Config for dev and prod are duplicated or missing

This lesson covers the patterns that solve these problems.

---

## Profiles: Environment-Specific Service Sets

Profiles let you define services that only start when explicitly requested. Without a profile, a service always starts.

```yaml
services:
  api:
    image: myapi
    # no profile → always starts

  postgres:
    image: postgres:16
    # no profile → always starts

  pgadmin:
    image: dpage/pgadmin4
    profiles: [tools] # ← only starts when 'tools' profile is active

  mailhog:
    image: mailhog/mailhog
    profiles: [tools, dev] # ← starts with 'tools' OR 'dev' profile
```

```bash
# Start core services only
docker compose up -d

# Start core + tools
docker compose --profile tools up -d

# Start core + dev tools
docker compose --profile dev up -d

# Multiple profiles
docker compose --profile tools --profile monitoring up -d

# Via environment variable
COMPOSE_PROFILES=tools,monitoring docker compose up -d
```

### Profile Gotcha: depends_on with profiled services

```yaml
services:
  api:
    depends_on:
      - postgres
    profiles: [] # ← always on

  postgres:
    profiles: [db] # ← only with 'db' profile


# ❌ This fails if you start api without the 'db' profile:
# api depends on postgres, but postgres isn't started
```

When a service depends on a profiled service, you must also activate that profile, or the dependency must be in a shared profile.

---

## `depends_on` with Health Check Conditions

The most common Compose mistake: thinking `depends_on` means "wait for the service to be ready."

```yaml
# ❌ This only waits for the container to START, not for Postgres to be READY
services:
  api:
    depends_on:
      - postgres

  postgres:
    image: postgres:16
```

By the time `api` starts, the Postgres process may still be initialising. The result: `ECONNREFUSED` on the first database query, a crash, and a restart loop.

### `condition: service_healthy`

```yaml
services:
  api:
    depends_on:
      postgres:
        condition: service_healthy # ✓ waits for health check to pass
      redis:
        condition: service_healthy

  postgres:
    image: postgres:16
    environment:
      POSTGRES_PASSWORD: secret
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 5s
      retries: 5
      start_period: 10s # ← grace period before first check

  redis:
    image: redis:7-alpine
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 3
```

### Condition options

| Condition                        | Meaning                                         |
| -------------------------------- | ----------------------------------------------- |
| `service_started`                | Container has started (default) — NOT readiness |
| `service_healthy`                | Health check is passing                         |
| `service_completed_successfully` | Container exited with code 0 (for init jobs)    |

```yaml
# Run a database migration before starting the app
services:
  api:
    depends_on:
      migrate:
        condition: service_completed_successfully
      postgres:
        condition: service_healthy

  migrate:
    image: myapi
    command: ["node", "migrate.js"] # runs once and exits
    depends_on:
      postgres:
        condition: service_healthy

  postgres:
    image: postgres:16
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      retries: 5
```

---

## Secrets and Configs

### The problem with environment variables for secrets

```yaml
# ❌ Secrets in environment variables
services:
  api:
    environment:
      DATABASE_PASSWORD: mysupersecretpassword # visible in docker inspect, docker compose config
```

`docker inspect` and `docker compose config` reveal all environment variables in plaintext.

### Compose Secrets

Secrets are mounted as files inside containers at `/run/secrets/<name>`. They are never exposed as environment variables.

```yaml
services:
  api:
    image: myapi
    secrets:
      - db_password
      - api_key

secrets:
  db_password:
    file: ./secrets/db_password.txt # ← read from file
  api_key:
    environment: API_KEY # ← read from host environment variable
```

Inside the container:

```bash
cat /run/secrets/db_password   # → mysupersecretpassword
cat /run/secrets/api_key       # → abc123xyz
```

In your application code:

```javascript
// ✓ Read from file, not from process.env
const dbPassword = fs.readFileSync("/run/secrets/db_password", "utf8").trim();
```

### Compose Configs

Configs are like secrets but for non-sensitive configuration files (they're still accessible via `docker config inspect`).

```yaml
services:
  nginx:
    image: nginx:alpine
    configs:
      - source: nginx_conf
        target: /etc/nginx/nginx.conf

configs:
  nginx_conf:
    file: ./nginx/nginx.conf
```

Configs support templating and can be rotated without rebuilding images.

---

## Extension Fields (`x-`)

YAML anchors are powerful for DRY Compose files. Combine with Compose's extension field convention:

```yaml
x-logging: &default-logging # ← YAML anchor
  driver: json-file
  options:
    max-size: "10m"
    max-file: "3"

x-common-env: &common-env
  NODE_ENV: production
  LOG_LEVEL: info

services:
  api:
    image: myapi
    logging: *default-logging # ← YAML alias (reference)
    environment:
      <<: *common-env # ← merge into environment
      PORT: "3000"

  worker:
    image: myworker
    logging: *default-logging
    environment:
      <<: *common-env
      QUEUE: jobs
```

Extension fields starting with `x-` are ignored by Compose but valid YAML — you can put anchors there without Compose trying to parse them as services.

---

## Override Files

Compose merges multiple files together. The base file defines the canonical structure; override files add or modify fields.

```
docker-compose.yml           ← base (checked into git)
docker-compose.override.yml  ← auto-loaded locally (git-ignored)
docker-compose.prod.yml      ← production overrides
docker-compose.ci.yml        ← CI overrides
```

**`docker-compose.override.yml`** is automatically merged when you run `docker compose up` with no `-f` flag — perfect for local dev overrides without modifying the base file.

```yaml
# docker-compose.yml (base — checked in)
services:
  api:
    image: myapi:${TAG:-latest}
    environment:
      NODE_ENV: production

  postgres:
    image: postgres:16
    volumes:
      - pgdata:/var/lib/postgresql/data

volumes:
  pgdata:
```

```yaml
# docker-compose.override.yml (dev overrides — git-ignored)
services:
  api:
    build: . # ← build locally instead of pulling image
    volumes:
      - ./src:/app/src # ← live reload
    environment:
      NODE_ENV: development
      DEBUG: "api:*"
    ports:
      - "9229:9229" # ← debugger port

  postgres:
    ports:
      - "5432:5432" # ← expose for local DB clients
```

```yaml
# docker-compose.prod.yml (production overrides)
services:
  api:
    deploy:
      replicas: 3
      resources:
        limits:
          cpus: "0.5"
          memory: 512M
    logging:
      driver: awslogs
      options:
        awslogs-group: /myapp/api
        awslogs-region: us-east-1
```

Explicit merge:

```bash
# Production deploy
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# CI build
docker compose -f docker-compose.yml -f docker-compose.ci.yml run --rm test
```

### Merge rules

| Field type                                      | Merge behaviour                        |
| ----------------------------------------------- | -------------------------------------- |
| Scalars (strings, numbers)                      | Override file wins                     |
| Sequences (ports, volumes, environment as list) | Concatenated (union)                   |
| Maps (environment as key=value)                 | Merged, override wins on conflict      |
| `command` / `entrypoint`                        | Override file wins (replaces entirely) |

---

## Environment Variable Precedence

Compose evaluates environment variables in this order (highest to lowest priority):

1. Values set directly in the `environment:` key in the Compose file
2. Values from the `env_file:` file specified in the service
3. Values from the `.env` file in the project directory
4. Shell environment variables

```yaml
services:
  api:
    environment:
      PORT: "3000" # ← highest priority, always wins
    env_file:
      - .env.local # ← second priority
      - .env # ← third priority (read top-to-bottom)
```

```bash
# .env (checked in, defaults)
DATABASE_URL=postgres://localhost/mydb
LOG_LEVEL=info

# .env.local (git-ignored, personal overrides)
DATABASE_URL=postgres://localhost/mydb_dev
DEBUG=true

# Shell override (highest for non-Compose-file vars)
export PORT=4000
docker compose up  # PORT from shell, DATABASE_URL from .env.local
```

**Gotcha**: The `.env` file is read by Compose itself for variable substitution in the Compose file (`${VARIABLE}`). It is NOT automatically passed to containers — you must explicitly reference it in `env_file:` or `environment:`.

---

## Compose Watch (Develop Mode)

Compose Watch (Docker Desktop 4.24+, Compose v2.22+) is a more capable alternative to bind mounts for development. It syncs files, triggers rebuilds, and restarts services without full bind mounts.

```yaml
services:
  api:
    build: .
    develop:
      watch:
        - action: sync # ← sync files without restart
          path: ./src
          target: /app/src
        - action: rebuild # ← rebuild image on Dockerfile change
          path: Dockerfile
        - action: sync+restart # ← sync and restart the container
          path: ./config
          target: /app/config
```

```bash
docker compose watch             # ← start with file watching
```

More efficient than bind mounts: files are synced via tar over the Docker socket rather than mounting a host directory.

---

## Hands-On Exercise 1: Health-Checked Service Startup

Fix this Compose file so that the API doesn't start until both Postgres and Redis are ready to accept connections.

```yaml
services:
  api:
    image: myapi
    depends_on:
      - postgres
      - redis

  postgres:
    image: postgres:16
    environment:
      POSTGRES_PASSWORD: secret

  redis:
    image: redis:7-alpine
```

<details>
<summary>Solution</summary>

```yaml
services:
  api:
    image: myapi
    depends_on:
      postgres:
        condition: service_healthy # ✓ wait for health check
      redis:
        condition: service_healthy

  postgres:
    image: postgres:16
    environment:
      POSTGRES_PASSWORD: secret
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 5s
      retries: 5
      start_period: 10s # ← give Postgres time to initialise

  redis:
    image: redis:7-alpine
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5
```

**Why `start_period` matters**: Postgres takes several seconds to initialise its data directory on first run. Without `start_period`, health checks start immediately and fail (unhealthy) before the process is even ready, causing Compose to mark it as failed immediately.

**Verify health status:**

```bash
docker compose ps
# NAME        STATUS
# postgres    Up 30 seconds (healthy)
# redis       Up 30 seconds (healthy)
# api         Up 5 seconds
```

</details>

---

## Hands-On Exercise 2: Dev/Prod Split

Refactor this monolithic Compose file into a base + override pattern:

```yaml
# Everything in one file — don't do this
services:
  api:
    build: .
    image: myapi:latest
    ports:
      - "3000:3000"
      - "9229:9229" # debugger
    volumes:
      - ./src:/app/src
    environment:
      NODE_ENV: development
      DATABASE_URL: postgres://postgres:secret@postgres/mydb
      DEBUG: "api:*"

  postgres:
    image: postgres:16
    environment:
      POSTGRES_PASSWORD: secret
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data

volumes:
  pgdata:
```

<details>
<summary>Solution</summary>

```yaml
# docker-compose.yml (base — checked in)
services:
  api:
    image: myapi:${TAG:-latest}
    environment:
      DATABASE_URL: postgres://postgres:secret@postgres/mydb

  postgres:
    image: postgres:16
    environment:
      POSTGRES_PASSWORD: secret
    volumes:
      - pgdata:/var/lib/postgresql/data

volumes:
  pgdata:
```

```yaml
# docker-compose.override.yml (dev — git-ignored)
services:
  api:
    build: . # ✓ build locally in dev
    ports:
      - "3000:3000"
      - "9229:9229" # ✓ debugger only in dev
    volumes:
      - ./src:/app/src # ✓ live reload in dev
    environment:
      NODE_ENV: development
      DEBUG: "api:*"

  postgres:
    ports:
      - "5432:5432" # ✓ expose for local DB clients
```

```yaml
# docker-compose.prod.yml (production)
services:
  api:
    environment:
      NODE_ENV: production
    logging:
      driver: json-file
      options:
        max-size: "10m"
        max-file: "5"

  postgres:
    # no ports: published in production — no direct DB access
```

Usage:

```bash
# Local dev (auto-loads override.yml)
docker compose up -d

# Production
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# CI (no override.yml loaded when -f is explicit)
docker compose -f docker-compose.yml run --rm api npm test
```

</details>

---

## Interview Questions

### Q1: What is the difference between `depends_on` with `service_started` vs `service_healthy`?

A very common interview topic because many developers have debugged race conditions in Compose. It tests practical experience.

<details>
<summary>Answer</summary>

`depends_on` with `service_started` (the default) waits for the container to be in a "running" state — i.e., the process started. It does NOT mean the process inside is ready to accept connections.

`service_healthy` waits for the container's health check to return a healthy status. This requires a `healthcheck:` block on the dependency service.

The classic race condition:

```
postgres container starts → still initialising
api starts (depends_on: postgres) → tries to connect → ECONNREFUSED
api crashes → Docker restarts it
postgres finishes initialising → api connects on restart
```

With `service_healthy`, Compose waits until Postgres's `pg_isready` check passes before starting the API. No crash loop.

**Important caveat**: `service_healthy` is Compose-only. In production orchestration (Kubernetes, ECS), the application itself should handle retries with exponential backoff — don't rely on the orchestrator to sequence startup.

</details>

---

### Q2: How does Compose handle merging of `environment:` keys across override files?

Tests whether you understand Compose merge semantics — important when debugging why a config value isn't what you expect.

<details>
<summary>Answer</summary>

Compose merges the two forms differently:

**Map form** (key: value) — keys are merged, override file wins on conflict:

```yaml
# base
environment:
  FOO: base
  BAR: base

# override
environment:
  FOO: override   # ← wins
  BAZ: new

# result
environment:
  FOO: override
  BAR: base
  BAZ: new
```

**List form** (`- KEY=value`) — lists are concatenated, later value wins for duplicate keys:

```yaml
# base
environment:
  - FOO=base
  - BAR=base

# override
environment:
  - FOO=override

# result (last value wins for FOO)
environment:
  - FOO=base
  - BAR=base
  - FOO=override   # ← FOO appears twice, but last one is used by most runtimes
```

**Don't mix forms across files** — Compose converts them all to map internally, but the interaction can be surprising.

**Precedence** (highest to lowest):

1. Inline `environment:` in Compose file
2. `env_file:` entries (files listed top-to-bottom, last wins)
3. `.env` file in project directory
4. Shell environment (for variable interpolation in Compose file, not for container env)

</details>

---

### Q3: What is the `.env` file and how does it differ from `env_file:` in Compose?

Commonly confused. Interviewers ask this to see if you know the difference between Compose-level interpolation and container-level environment variables.

<details>
<summary>Answer</summary>

**`.env` file**: Used by Compose itself for variable substitution in the Compose YAML (`${VARIABLE}`). Not automatically passed to containers.

```yaml
# .env
TAG=v1.2.3
PORT=3000

# docker-compose.yml
services:
  api:
    image: myapi:${TAG}         # ← .env is used here (Compose interpolation)
    ports:
      - "${PORT}:3000"          # ← and here
```

**`env_file:`**: Passes variables from a file into the container's environment. The variables must be explicitly listed or the file explicitly referenced.

```yaml
services:
  api:
    env_file:
      - .env # ← NOW the container also gets these vars
      - .env.local
```

**The gotcha**: Many developers assume `.env` automatically becomes the container's environment because it does in tools like Vite or Next.js. In Compose, `.env` is only for Compose-level variable substitution unless you also add `env_file: - .env`.

|                                 | `.env`                        | `env_file:`           |
| ------------------------------- | ----------------------------- | --------------------- |
| Who reads it                    | Compose (before parsing YAML) | Container runtime     |
| Used for                        | YAML interpolation (`${VAR}`) | Container environment |
| Auto-loaded                     | ✓ Yes                         | ❌ Must be specified  |
| Overrides inline `environment:` | ❌ No                         | ❌ No (inline wins)   |

</details>

---

### Q4: How would you structure a Compose file for a project that needs different configurations for local dev, CI, and production?

An architecture question that tests practical Compose experience. Interviewers want to see a clean pattern, not a single file with conditionals.

<details>
<summary>Answer</summary>

Use the **base + override pattern** with one file per environment:

```
docker-compose.yml           ← base: production-like defaults, no local dev specifics
docker-compose.override.yml  ← dev: auto-loaded, git-ignored
docker-compose.ci.yml        ← CI: in-memory DB, no volumes, fast startup
docker-compose.prod.yml      ← prod: logging drivers, resource limits, no build context
```

**Base file rules**:

- Use production-suitable images (`image:`, not `build:`)
- No published ports except what's genuinely needed
- No source code bind mounts
- Define volumes and networks

**Override file rules**:

- Add `build: .` for local building
- Add dev-only ports and volumes
- Add debugging tools

**CI file**:

```yaml
# docker-compose.ci.yml
services:
  postgres:
    tmpfs: /var/lib/postgresql/data # ← in-memory, faster, no persistence needed
  api:
    command: npm test # ← run tests instead of the server
```

Usage:

```bash
# Dev (auto-loads override.yml)
docker compose up

# CI
docker compose -f docker-compose.yml -f docker-compose.ci.yml run api

# Prod
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

This approach keeps git history clean (base file is always valid), allows per-developer overrides (`.override.yml` is git-ignored), and makes CI/CD pipelines explicit.

</details>

---

## Key Takeaways

1. **`depends_on` with `service_started` is not readiness** — use `condition: service_healthy` with a `healthcheck:` block
2. **Profiles keep dev tools out of production** — annotate pgAdmin, mailhog, etc. with `profiles: [tools]`
3. **Secrets are files, not env vars** — mounted at `/run/secrets/<name>`, never in `docker inspect`
4. **Base + override pattern** separates environment concerns cleanly; `.override.yml` is auto-loaded and git-ignored
5. **`.env` is for Compose YAML interpolation**, not automatic container environment — use `env_file:` for the latter
6. **Extension fields (`x-`) + YAML anchors** eliminate duplication in logging, resource limits, environment blocks
7. **Merge semantics matter**: sequences are concatenated, maps are merged with override winning — know the rules to avoid debugging surprises
8. **`service_completed_successfully`** is the right condition for one-shot init containers (migrations, seed scripts)

## Next Steps

In [Lesson 05: Security Hardening](lesson-05-security-hardening.md), you'll learn:

- Running containers as non-root users and why it matters
- Read-only root filesystems and what breaks when you enable them
- Linux capabilities — what they are, what Docker drops by default, and what to drop further
- Seccomp profiles, AppArmor, and the proper way to manage runtime secrets
