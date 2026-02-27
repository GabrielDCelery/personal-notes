# Lesson 10: ARG and ENV

How Docker injects values at build time and runtime — and why confusing the two causes broken images and leaked secrets.

---

## The Problem You Hit Eventually

You have a CI pipeline that builds Docker images. Different environments need different API URLs, feature flags, or version strings. You reach for environment variables — the obvious solution.

Then you discover that some variables need to exist only while the image is being built (compiler flags, dependency versions, registry URLs), while others need to persist into the running container (app config, feature flags, log levels). Using the wrong instruction in the wrong place produces either an empty variable at runtime or a secret baked permanently into the image.

`ARG` and `ENV` look similar. They are fundamentally different.

---

## ARG — Build-Time Variables

`ARG` declares a variable that exists only during `docker build`. It is the mechanism for parameterising Dockerfiles without hardcoding values.

```dockerfile
ARG NODE_VERSION=20
FROM node:${NODE_VERSION}-alpine

ARG APP_ENV=production
RUN echo "Building for: $APP_ENV"
```

Pass values at build time with `--build-arg`:

```sh
docker build --build-arg NODE_VERSION=22 --build-arg APP_ENV=staging .
```

| Property                          | Behaviour                                          |
| --------------------------------- | -------------------------------------------------- |
| Scope                             | Build process only                                 |
| Available to `FROM`               | Yes (declared before `FROM`)                       |
| Available to `RUN`, `COPY`, `ADD` | Yes                                                |
| Available in running container    | No                                                 |
| Visible in `docker inspect`       | No                                                 |
| Visible in `docker history`       | **Yes — sensitive values are exposed**             |
| Default value                     | Optional — omitting `--build-arg` uses the default |

---

## ENV — Image-Level Environment Variables

`ENV` sets variables that are baked into the image and available to every container that runs from it.

```dockerfile
ENV NODE_ENV=production
ENV PORT=3000 LOG_LEVEL=info   # multiple on one line
```

| Property                       | Behaviour                           |
| ------------------------------ | ----------------------------------- |
| Scope                          | Image and all containers            |
| Available during build         | Yes — subsequent `RUN` steps see it |
| Available in running container | Yes                                 |
| Visible in `docker inspect`    | **Yes — values are exposed**        |
| Overridable at runtime         | Yes — `docker run -e VAR=value`     |
| Default value                  | The value set in the Dockerfile     |

---

## How They Interact

The common pattern: accept a value at build time via `ARG`, then promote it into the image with `ENV`.

```dockerfile
ARG APP_VERSION
ENV APP_VERSION=$APP_VERSION
```

```sh
docker build --build-arg APP_VERSION=1.4.2 -t myapp:1.4.2 .
docker run myapp  # APP_VERSION=1.4.2 available inside container
```

Without `ENV`, `APP_VERSION` disappears after the build. With `ENV`, it persists into the runtime environment.

---

## ARG Scope Rules

### Before vs After FROM

`ARG` declared before `FROM` is only available to `FROM` itself. It goes out of scope the moment the `FROM` instruction completes.

```dockerfile
ARG BASE_VERSION=20          # ✓ usable in FROM
FROM node:${BASE_VERSION}-alpine

RUN echo $BASE_VERSION       # ❌ empty — ARG is out of scope
```

To use it after `FROM`, re-declare it (no default needed — it inherits the value):

```dockerfile
ARG BASE_VERSION=20
FROM node:${BASE_VERSION}-alpine

ARG BASE_VERSION             # re-declare, no default
RUN echo $BASE_VERSION       # ✓ 20
```

### Multi-Stage Builds

Each build stage has its own scope. An `ARG` declared in one stage is not available in another. Pre-`FROM` `ARG` values can be re-declared in any stage.

```dockerfile
ARG REGISTRY=docker.io

FROM ${REGISTRY}/node:20-alpine AS builder
ARG REGISTRY                  # re-declare to use below
RUN echo "Building from: $REGISTRY"

FROM ${REGISTRY}/node:20-alpine AS runner
ARG REGISTRY                  # must re-declare again — separate stage scope
RUN echo "Running from: $REGISTRY"
```

### ARG and Layer Caching

Docker invalidates the build cache when an `ARG` value changes. This means changing `--build-arg` can bust the cache for all subsequent layers, even layers that do not use that variable.

```dockerfile
ARG BUILD_DATE                # ❌ if this changes every build, all layers below are rebuilt
RUN npm ci                    # cache busted even though BUILD_DATE is irrelevant here
```

Place cache-busting `ARG` instructions as late as possible:

```dockerfile
RUN npm ci                    # ✓ cached — BUILD_DATE not declared yet

ARG BUILD_DATE
LABEL build.date=$BUILD_DATE  # cache busts here, but npm ci is already cached
```

---

## ENV Scope and Override Precedence

`ENV` values set in the Dockerfile are defaults. They can be overridden at runtime with `docker run -e`:

```sh
docker run -e LOG_LEVEL=debug myapp
```

The precedence chain, highest to lowest:

```
docker run -e VAR=value          ← highest: runtime override
docker run --env-file .env       ← file-based override
ENV VAR=value in Dockerfile      ← image default
system environment (on host)     ← not propagated unless explicitly passed
```

`ENV` values are inherited by all child processes inside the container, including processes started by shell scripts, package runners, and init systems.

---

## The Secret Problem

Neither `ARG` nor `ENV` is safe for secrets.

```dockerfile
ARG DB_PASSWORD
ENV DB_PASSWORD=$DB_PASSWORD    # ❌ visible in docker inspect
RUN echo $DB_PASSWORD           # ❌ visible in docker history
```

**`docker inspect` exposes ENV:**

```sh
docker inspect mycontainer | jq '.[0].Config.Env'
# ["DB_PASSWORD=hunter2", "NODE_ENV=production", ...]
```

**`docker history` exposes ARG values used in RUN:**

```sh
docker history myimage
# IMAGE   CREATED  CREATED BY
# sha256  ...      RUN echo hunter2   ← the value is in the layer metadata
```

Even if you `unset` the variable in a later layer, the value is preserved in the earlier layer's metadata. Docker images are immutable layer stacks — you cannot retroactively scrub a layer.

**The correct approach for secrets:** use BuildKit's `--secret` flag.

```dockerfile
# syntax=docker/dockerfile:1
RUN --mount=type=secret,id=db_password \
    DB_PASSWORD=$(cat /run/secrets/db_password) \
    && ./configure --db-password="$DB_PASSWORD"
```

```sh
docker build --secret id=db_password,src=./db_password.txt .
```

The secret is mounted as a temporary file during that `RUN` step only. It never appears in the layer, the history, or `docker inspect`.

---

## ENV and Image Size

Each `ENV` instruction creates a new image layer (in older BuildKit modes). More importantly, `ENV` values add metadata overhead and increase the risk of accidentally shipping configuration that should vary per environment.

Prefer injecting runtime config via `docker run -e` or `--env-file` rather than baking it into the image. Baked `ENV` values make the same image harder to reuse across environments.

```dockerfile
# ❌ Hard to reuse across environments
ENV DATABASE_URL=postgres://prod-host:5432/mydb
ENV CACHE_URL=redis://prod-cache:6379

# ✓ Provide sensible defaults, override at runtime
ENV DATABASE_URL=postgres://localhost:5432/mydb
ENV CACHE_URL=redis://localhost:6379
```

---

## Hands-On Exercise 1: Fix the Broken Multi-Stage ARG

This Dockerfile fails to pass the registry argument into the second stage. Find the problem and fix it.

```dockerfile
ARG REGISTRY=docker.io

FROM ${REGISTRY}/node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM ${REGISTRY}/node:20-alpine AS runner
WORKDIR /app
COPY --from=builder /app/dist ./dist
RUN echo "Registry used: $REGISTRY"
CMD ["node", "dist/server.js"]
```

```sh
docker build --build-arg REGISTRY=myregistry.example.com .
# Expected: "Registry used: myregistry.example.com"
# Actual:   "Registry used: "
```

<details>
<summary>Solution</summary>

**Problem:**

`ARG REGISTRY` declared before `FROM` is only available to `FROM` instructions. Each stage starts with a clean scope. `$REGISTRY` in the `runner` stage is empty because it was never re-declared there.

The same issue affects the `builder` stage — `$REGISTRY` is empty inside the `RUN echo` equivalent if you added one.

**Fixed:**

```dockerfile
ARG REGISTRY=docker.io

FROM ${REGISTRY}/node:20-alpine AS builder
ARG REGISTRY                         # ✓ re-declare to use within this stage
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM ${REGISTRY}/node:20-alpine AS runner
ARG REGISTRY                         # ✓ re-declare again for this stage
WORKDIR /app
COPY --from=builder /app/dist ./dist
RUN echo "Registry used: $REGISTRY"  # ✓ now correctly outputs the value
CMD ["node", "dist/server.js"]
```

Each `ARG REGISTRY` after a `FROM` is a re-declaration that pulls in the value supplied at build time, without needing to repeat the default.

</details>

---

## Hands-On Exercise 2: Identify the Secret Leak

This Dockerfile builds an app that needs an API key during the build to download a private package. Find all the ways the secret leaks and rewrite it to be safe.

```dockerfile
FROM node:20-alpine

ARG NPM_TOKEN
ENV NPM_TOKEN=$NPM_TOKEN

RUN echo "//registry.npmjs.org/:_authToken=$NPM_TOKEN" > ~/.npmrc
RUN npm install
RUN rm ~/.npmrc

ENV NODE_ENV=production
CMD ["node", "server.js"]
```

<details>
<summary>Solution</summary>

**Leaks:**

1. ❌ `ARG NPM_TOKEN` — the value is visible in `docker history` because it is used in a `RUN` command
2. ❌ `ENV NPM_TOKEN=$NPM_TOKEN` — the token is baked into the image and visible in `docker inspect`
3. ❌ `RUN rm ~/.npmrc` — removing the file in a separate layer does not remove it from the previous layer. The `.npmrc` file with the token exists permanently in the layer created by the `RUN echo` step

**Fixed using BuildKit secrets:**

```dockerfile
# syntax=docker/dockerfile:1
FROM node:20-alpine

RUN --mount=type=secret,id=npm_token \
    echo "//registry.npmjs.org/:_authToken=$(cat /run/secrets/npm_token)" > ~/.npmrc \
    && npm install \
    && rm ~/.npmrc

ENV NODE_ENV=production
CMD ["node", "server.js"]
```

```sh
docker build --secret id=npm_token,src=./npm_token.txt .
```

Everything happens in a single `RUN` step: the secret is mounted, `.npmrc` is written, `npm install` runs, and `.npmrc` is deleted — all within one layer. The secret never appears in image metadata or history.

</details>

---

## Interview Questions

### Q1: What is the difference between ARG and ENV in a Dockerfile?

Interviewers ask this to test whether you understand Docker's build vs runtime model. It also reveals whether you have shipped real images to production or just run tutorials.

<details>
<summary>Answer</summary>

`ARG` variables exist only during `docker build`. They are passed via `--build-arg` and are not available in running containers. They are useful for parameterising the build itself — base image versions, registry URLs, compile-time flags.

`ENV` variables are baked into the image and available in every container that runs from it. They appear in `docker inspect` and are inherited by all processes in the container. They can be overridden at runtime with `docker run -e`.

The key distinction: `ARG` is build-time only, `ENV` is image-level and persists into runtime. To get a build-time value into a running container, you must explicitly copy it: `ENV MY_VAR=$MY_ARG`.

</details>

---

### Q2: Why is it unsafe to pass secrets via ARG or ENV?

Interviewers ask this to assess your security awareness around image builds. It catches developers who know the syntax but have not thought through what ends up in the image.

<details>
<summary>Answer</summary>

`ENV` values are stored in the image's configuration metadata and visible to anyone who can run `docker inspect`. They appear in plain text in the output.

`ARG` values that are used in `RUN` commands are embedded in the layer's metadata and visible via `docker history`. Even if you delete the file that contained the secret in a later `RUN` step, the secret persists in the earlier layer — Docker images are immutable layer stacks.

The safe alternative is BuildKit's `--secret` flag with `RUN --mount=type=secret`. The secret is mounted as a temporary file only during that specific `RUN` step. It never appears in any layer, history output, or `docker inspect` result.

</details>

---

### Q3: Why do ARG values declared before FROM go out of scope after FROM?

Interviewers ask this to probe depth of understanding of Docker's build model. It distinguishes developers who have debugged multi-stage builds from those who have only used single-stage ones.

<details>
<summary>Answer</summary>

Docker's build model treats each `FROM` as the start of a new build stage with its own isolated scope. The pre-`FROM` scope is a special global scope that only exists to allow parameterising the `FROM` instruction itself — for dynamic base image selection.

When `FROM` is processed, a new stage begins and inherits nothing from the pre-`FROM` scope automatically. This is intentional: stages are meant to be composable and independently reproducible. Automatic inheritance would create implicit coupling between stages.

To use a pre-`FROM` `ARG` value within a stage, you re-declare it with a bare `ARG VARNAME` (no default). Docker then pulls in the value that was supplied at build time. You must do this in every stage that needs the value.

</details>

---

### Q4: How does changing a --build-arg value affect Docker's layer cache?

Interviewers ask this because cache invalidation bugs in CI pipelines are common and expensive. Understanding this shows you have operated Docker at scale, not just locally.

<details>
<summary>Answer</summary>

Docker invalidates the cache for a layer when any input to that layer changes. For `ARG`, changing the value via `--build-arg` invalidates the cache for the `ARG` instruction itself and all subsequent layers in that stage — even layers that do not reference the variable.

This means a poorly placed `ARG` can force expensive operations like `npm install` or `apt-get` to re-run on every build, even when their inputs have not changed.

The fix is to declare `ARG` instructions as late as possible, after the layers you want to keep cached:

```dockerfile
RUN npm ci                    # ✓ cached — no ARG declared yet

ARG BUILD_DATE                # cache busts here
LABEL build.date=$BUILD_DATE  # only this and subsequent layers rebuild
```

In practice: declare `ARG` values that change frequently (build timestamps, git SHA) as late as possible. Declare stable values (registry, base version) at the top where they are needed.

</details>

---

## Key Takeaways

1. **ARG is build-time only** — not available in running containers, not in `docker inspect`
2. **ENV is image-level** — baked in, persists to all containers, visible in `docker inspect`
3. **To bridge the gap**: use `ENV MY_VAR=$MY_ARG` to promote a build arg into the runtime environment
4. **Pre-FROM ARG scope is limited** — only available to `FROM`; re-declare after `FROM` to use within a stage
5. **Each stage is isolated** — re-declare `ARG` in every stage that needs the value
6. **ARG values bust the cache** — place frequently-changing `ARG` declarations late to preserve cached layers
7. **Neither ARG nor ENV is safe for secrets** — both leak through `docker history` or `docker inspect`
8. **Use BuildKit `--secret`** for secrets — mounted per `RUN` step, never written to any layer
9. **ENV values are overridable at runtime** — use `docker run -e` or `--env-file` to inject per-environment config
10. **Baking environment-specific config into ENV is an anti-pattern** — prefer runtime injection to keep images portable across environments

---

## Next Steps

In [Lesson 11: COPY, ADD, and Build Context](lesson-11-copy-add-and-build-context.md), you will learn:

- How the build context works and why large contexts slow builds
- The difference between `COPY` and `ADD` and when each is appropriate
- `.dockerignore` patterns and their effect on context size and cache
- `COPY --from` for copying across stages and from external images
