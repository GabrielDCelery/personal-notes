# Lesson 06: Production Patterns

Signal handling, health checks, restart policies, resource limits, and logging — the patterns that separate a running container from a production-ready one.

## PID 1 and the Signal Handling Problem

Here is a production incident waiting to happen: your container gets a `SIGTERM` (from a deploy, `docker stop`, or Kubernetes pod eviction) and doesn't shut down gracefully. Requests are dropped, database connections aren't cleaned up, and the orchestrator kills it with `SIGKILL` after 30 seconds.

The cause is almost always the same: your process isn't PID 1, so it never receives the signal.

### What PID 1 does (the init process)

In a normal Linux system, `init` (PID 1) is responsible for:

1. Receiving signals and forwarding them to child processes
2. Reaping zombie processes (calling `waitpid` on exited children)

When Docker sends `SIGTERM` to a container, it sends it to PID 1. If PID 1 is a shell (`/bin/sh -c node server.js`), the shell doesn't forward `SIGTERM` to `node` by default — the signal is silently dropped.

### Shell form vs exec form

```dockerfile
# ❌ Shell form — PID 1 is /bin/sh, node is a child process, SIGTERM dropped
CMD node server.js
# Same as: CMD ["/bin/sh", "-c", "node server.js"]

# ✓ Exec form — node IS PID 1, receives SIGTERM directly
CMD ["node", "server.js"]
```

Verify:

```bash
docker exec mycontainer ps aux
# Shell form:  PID 1 = /bin/sh, PID 7 = node server.js
# Exec form:   PID 1 = node server.js
```

### tini: a proper init for containers

Even with exec form, there's a zombie reaping problem: if your app forks child processes and they exit, they become zombies (kernel keeps their entry in the process table until PID 1 calls `waitpid`). Node.js doesn't do this. Over time, zombie accumulation can exhaust the process table.

`tini` is a tiny init process specifically designed for containers:

```dockerfile
FROM node:20-alpine

# Install tini
RUN apk add --no-cache tini

WORKDIR /app
COPY . .
RUN npm ci --omit=dev

ENTRYPOINT ["/sbin/tini", "--"]   # ✓ tini is PID 1
CMD ["node", "server.js"]          # ← passed as args to tini, tini forwards signals
```

Or use Docker's built-in init:

```bash
docker run --init myimage           # injects tini automatically
```

```yaml
# docker-compose.yml
services:
  api:
    image: myapi
    init: true # ✓ equivalent to --init
```

### Graceful shutdown in Node.js

Your application must handle `SIGTERM`:

```javascript
// server.js
const server = http.createServer(app);

process.on("SIGTERM", () => {
  console.log("SIGTERM received, shutting down gracefully");
  server.close(() => {
    // Close DB connections, flush logs, etc.
    process.exit(0);
  });

  // Force exit if graceful shutdown takes too long
  setTimeout(() => {
    console.error("Forcing exit after timeout");
    process.exit(1);
  }, 10000);
});
```

---

## HEALTHCHECK

`HEALTHCHECK` tells Docker (and orchestrators) whether the container is functioning correctly, not just running.

```dockerfile
HEALTHCHECK --interval=30s --timeout=10s --start-period=15s --retries=3 \
  CMD curl -f http://localhost:3000/health || exit 1
```

| Option           | Default | Meaning                                       |
| ---------------- | ------- | --------------------------------------------- |
| `--interval`     | 30s     | Time between checks                           |
| `--timeout`      | 30s     | How long to wait for a response               |
| `--start-period` | 0s      | Grace period before health checks start       |
| `--retries`      | 3       | Consecutive failures before marking unhealthy |

### Health check states

| State       | Meaning                                               |
| ----------- | ----------------------------------------------------- |
| `starting`  | Container started, `start_period` hasn't elapsed      |
| `healthy`   | Last N checks passed                                  |
| `unhealthy` | Last N checks failed (triggers restart if policy set) |

```bash
docker inspect mycontainer --format '{{ .State.Health.Status }}'
# → healthy
docker inspect mycontainer --format '{{ json .State.Health }}' | jq
# Shows last 5 health check results
```

### Writing a good health endpoint

```javascript
// /health endpoint
app.get("/health", async (req, res) => {
  try {
    await db.query("SELECT 1"); // ← check DB connectivity
    res.json({ status: "ok" });
  } catch (err) {
    res.status(503).json({ status: "error", message: err.message });
  }
});
```

**Gotcha**: `curl -f` returns exit code 1 on HTTP 4xx/5xx — that's what you want. Without `-f`, curl exits 0 even on 500 errors.

```dockerfile
# ✓ -f flag makes curl fail on HTTP errors
HEALTHCHECK CMD curl -f http://localhost:3000/health || exit 1

# ❌ Without -f, always exits 0 even if the app is returning 500
HEALTHCHECK CMD curl http://localhost:3000/health
```

### Health check without curl

Alpine images often lack curl. Use wget or a small script:

```dockerfile
# wget alternative
HEALTHCHECK CMD wget -qO- http://localhost:3000/health || exit 1

# Or use a Node.js script bundled with the app
HEALTHCHECK CMD node /app/healthcheck.js
```

```javascript
// healthcheck.js
const http = require("http");
const req = http.request(
  { host: "localhost", port: 3000, path: "/health", timeout: 2000 },
  (res) => {
    process.exit(res.statusCode === 200 ? 0 : 1);
  },
);
req.on("error", () => process.exit(1));
req.end();
```

---

## Restart Policies

Restart policies control what Docker does when a container exits.

| Policy           | Behaviour                                                        |
| ---------------- | ---------------------------------------------------------------- |
| `no`             | Never restart (default)                                          |
| `always`         | Always restart, regardless of exit code                          |
| `on-failure[:N]` | Restart only on non-zero exit; optionally limit to N retries     |
| `unless-stopped` | Always restart except when explicitly stopped with `docker stop` |

```bash
docker run --restart unless-stopped myimage

# With retry limit
docker run --restart on-failure:5 myimage
```

```yaml
services:
  api:
    image: myapi
    restart: unless-stopped # ✓ survives host reboots and crashes
```

### Restart policy in production vs Kubernetes

In Kubernetes (or ECS), the orchestrator manages restarts at the pod/task level — set `restart: no` or `restart: on-failure` and let the orchestrator decide. Using `restart: always` in Kubernetes can cause duplicate restart loops (Docker AND Kubernetes both restarting).

### Backoff behaviour

Docker implements exponential backoff for restarts: 100ms → 200ms → 400ms → ... → max 1 minute. After 10 seconds of running without crashing, the backoff counter resets.

```bash
# Watch restart behaviour
docker ps  # RESTARTS column shows count
```

---

## Resource Limits

Without limits, a runaway container can consume all host resources and starve other containers (and the host itself).

### Memory limits

```bash
docker run \
  --memory 512m \              # hard limit — OOM killed if exceeded
  --memory-swap 512m \         # ← same as memory = no swap (disable swap)
  --memory-reservation 256m \  # soft limit — preferred minimum
  myimage
```

```yaml
services:
  api:
    image: myapi
    deploy:
      resources:
        limits:
          memory: 512M
        reservations:
          memory: 256M
```

**Note**: `deploy.resources` is only respected by Docker Swarm and `docker compose up`. For standalone `docker run`, use `--memory`.

### Memory and the OOM killer

When a container exceeds its memory limit, the Linux OOM killer terminates the process with signal 9 (`SIGKILL`). This:

- Cannot be caught by the application
- Does not trigger graceful shutdown
- Shows up as `OOMKilled: true` in `docker inspect`

```bash
# Check if a container was OOM killed
docker inspect mycontainer --format '{{ .State.OOMKilled }}'
# → true

# View OOM kill events
dmesg | grep -i 'oom'
```

**Set limits slightly above the application's peak usage**, measured under load testing. Too tight = random OOM kills. Too loose = resource waste.

### CPU limits

```bash
docker run \
  --cpus 0.5 \           # 0.5 = 50% of one CPU core
  --cpu-shares 512 \     # relative weight (default 1024) — soft limit
  myimage
```

```yaml
deploy:
  resources:
    limits:
      cpus: "0.5"
    reservations:
      cpus: "0.25"
```

**`--cpus` vs `--cpu-shares`**:

- `--cpus`: Hard CFS (Completely Fair Scheduler) quota — enforced even when host is idle
- `--cpu-shares`: Relative weight — only matters when CPU is contested

---

## Logging Drivers

By default, Docker uses the `json-file` driver which writes logs to `/var/lib/docker/containers/<id>/<id>-json.log`. Without size limits, these can fill disk.

```bash
docker logs mycontainer              # ← reads from json-file driver
docker logs -f mycontainer           # ← follow (like tail -f)
docker logs --since 1h mycontainer   # ← last hour
docker logs --tail 100 mycontainer   # ← last 100 lines
```

### Configure json-file limits (always do this)

```json
// /etc/docker/daemon.json
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}
```

Per-container override:

```yaml
services:
  api:
    logging:
      driver: json-file
      options:
        max-size: "10m"
        max-file: "3"
```

### Production logging drivers

| Driver      | Destination        | `docker logs` works |
| ----------- | ------------------ | ------------------- |
| `json-file` | Local disk         | ✓ Yes               |
| `none`      | Discarded          | ❌ No               |
| `syslog`    | Host syslog        | ❌ No               |
| `journald`  | systemd journal    | ❌ No               |
| `awslogs`   | AWS CloudWatch     | ❌ No               |
| `fluentd`   | Fluentd aggregator | ❌ No               |
| `splunk`    | Splunk HEC         | ❌ No               |
| `gcplogs`   | GCP Stackdriver    | ❌ No               |

**Important**: When using non-`json-file` drivers, `docker logs` stops working. Log retrieval goes through the destination system instead.

```yaml
# AWS CloudWatch
services:
  api:
    logging:
      driver: awslogs
      options:
        awslogs-region: us-east-1
        awslogs-group: /myapp/api
        awslogs-stream: "{{.Name}}"
        awslogs-create-group: "true"
```

### Logging best practices

```javascript
// ✓ Write logs to stdout/stderr — let Docker handle the rest
console.log(
  JSON.stringify({ level: "info", msg: "Server started", port: 3000 }),
);
console.error(
  JSON.stringify({ level: "error", msg: err.message, stack: err.stack }),
);

// ❌ Don't write to log files inside the container
fs.appendFileSync("/var/log/app.log", logLine); // ← grows unboundedly, not visible to `docker logs`
```

Structured JSON logs (one JSON object per line) are parseable by CloudWatch, Datadog, Loki, and Splunk without custom parsing rules.

---

## Ulimits

Ulimits set kernel-level resource limits for the container process.

```bash
docker run \
  --ulimit nofile=65536:65536 \   # open file descriptors (soft:hard)
  --ulimit nproc=2048:4096 \      # process count
  myimage
```

Common limits to configure for production:

| Limit     | Default         | Recommended for servers             |
| --------- | --------------- | ----------------------------------- |
| `nofile`  | 1024:1024       | 65536:65536                         |
| `nproc`   | 0:0 (unlimited) | 2048:4096                           |
| `memlock` | 64k             | unlimited (for Elasticsearch, etc.) |

```yaml
# daemon.json — set defaults for all containers
{
  "default-ulimits":
    { "nofile": { "Name": "nofile", "Hard": 65536, "Soft": 65536 } },
}
```

---

## Hands-On Exercise 1: Fix Signal Handling

A Node.js API container takes 30 seconds to stop (always waits for Docker's kill timeout). Identify why and fix it.

```dockerfile
FROM node:20-alpine
WORKDIR /app
COPY . .
RUN npm ci --omit=dev
CMD node server.js
```

<details>
<summary>Solution</summary>

**Root cause**: Shell form `CMD node server.js` means `/bin/sh -c node server.js` is PID 1. `SIGTERM` goes to the shell, not to node. The shell doesn't forward it. After Docker's stop timeout (default 10s), Docker sends `SIGKILL`.

**Fix 1: Use exec form**

```dockerfile
CMD ["node", "server.js"]   # ✓ node is PID 1, receives SIGTERM
```

**Fix 2: Add tini for full init capabilities**

```dockerfile
FROM node:20-alpine
RUN apk add --no-cache tini
WORKDIR /app
COPY . .
RUN npm ci --omit=dev
ENTRYPOINT ["/sbin/tini", "--"]
CMD ["node", "server.js"]
```

**Fix 3: Handle SIGTERM in the app**

```javascript
process.on("SIGTERM", () => {
  server.close(() => {
    db.end(() => process.exit(0));
  });
  setTimeout(() => process.exit(1), 10000);
});
```

**Verify the fix:**

```bash
time docker stop mycontainer
# Before fix: ~30 seconds (waits for kill timeout)
# After fix:  ~1-2 seconds (clean shutdown)
```

**Adjust Docker's stop timeout** if graceful shutdown needs more than 10s:

```bash
docker stop --time 30 mycontainer
# or: STOPSIGNAL SIGQUIT in Dockerfile for apps that use SIGQUIT for graceful shutdown
```

</details>

---

## Hands-On Exercise 2: Production-Ready Compose Service

Add health check, restart policy, resource limits, and structured logging to this service:

```yaml
services:
  api:
    image: myapi:latest
    ports:
      - "3000:3000"
    environment:
      DATABASE_URL: postgres://postgres:secret@db/mydb
```

<details>
<summary>Solution</summary>

```yaml
services:
  api:
    image: myapi:latest
    init: true # ✓ tini as PID 1
    ports:
      - "3000:3000"
    environment:
      DATABASE_URL: postgres://postgres:secret@db/mydb
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://localhost:3000/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 15s # ✓ grace period for startup
    restart: unless-stopped # ✓ survives crashes and reboots
    logging:
      driver: json-file
      options:
        max-size: "10m" # ✓ prevent disk fill
        max-file: "5"
    deploy:
      resources:
        limits:
          cpus: "0.5"
          memory: 512M
        reservations:
          cpus: "0.1"
          memory: 128M
    ulimits:
      nofile:
        soft: 65536
        hard: 65536
```

**What each addition does:**

- `init: true` → tini handles PID 1, zombie reaping, signal forwarding
- `healthcheck` → orchestrators and `depends_on` use this to know when the service is ready
- `restart: unless-stopped` → auto-restart on crash; stops when explicitly stopped
- `logging` with limits → prevents log files from filling disk
- `deploy.resources` → prevents container from starving other services
- `ulimits.nofile` → allows the app to handle thousands of concurrent connections

</details>

---

## Interview Questions

### Q1: Why does Docker wait 10 seconds before a container fully stops, and how do you fix it?

Classic operational interview question. The answer reveals whether you understand PID 1 and signal propagation.

<details>
<summary>Answer</summary>

The 10-second wait is Docker's stop timeout. When you run `docker stop`, Docker sends `SIGTERM` to PID 1. If the container hasn't exited after 10 seconds (configurable with `--time`), Docker sends `SIGKILL`.

The reason containers don't shut down on `SIGTERM` is usually one of:

1. **Shell form CMD**: `/bin/sh -c node app.js` — shell is PID 1, doesn't forward `SIGTERM` to node
2. **No signal handler**: App doesn't listen for `SIGTERM` and therefore doesn't initiate shutdown
3. **Signal forwarding issue**: tini not used, or wrong `ENTRYPOINT` form

**Fix:**

1. Use exec form: `CMD ["node", "app.js"]`
2. Add `process.on('SIGTERM', ...)` handler in the application
3. Use `init: true` (tini) so signal forwarding is handled correctly
4. Set `STOPSIGNAL` in the Dockerfile if the app uses a different signal (some apps use `SIGQUIT` or `SIGUSR1` for graceful drain)

```dockerfile
STOPSIGNAL SIGQUIT   # nginx uses SIGQUIT for graceful shutdown
```

</details>

---

### Q2: What happens when a container exceeds its memory limit?

Tests whether you know about the OOM killer and how to diagnose/prevent it.

<details>
<summary>Answer</summary>

The Linux kernel's OOM (Out of Memory) killer terminates the offending process with `SIGKILL`. This:

- Is not catchable by the application (unlike `SIGTERM`)
- Triggers immediately when the cgroup memory limit is exceeded
- Does not trigger graceful shutdown
- Gets recorded by Docker: `docker inspect mycontainer` shows `"OOMKilled": true`

```bash
# Diagnose OOM kill
docker inspect mycontainer | jq '.[0].State'
# "OOMKilled": true, "ExitCode": 137  ← 128 + 9 (SIGKILL)

# Host kernel also logs it
dmesg | grep -i oom-kill
```

**Memory limit options**:

- `--memory`: Hard limit — process is killed when exceeded
- `--memory-swap`: Total memory + swap limit. Set equal to `--memory` to disable swap. Default is 2x memory limit.
- `--memory-reservation`: Soft limit — Docker scheduler hint, not enforced by kernel

**To prevent OOM kills**:

1. Load test to find actual peak memory usage
2. Set limit 20-50% above peak (not too tight)
3. Add heap dump / memory profiling on OOM events in the application
4. Use `--oom-score-adj` to adjust the OOM killer's preference

</details>

---

### Q3: What is the difference between `restart: always` and `restart: unless-stopped`?

Subtle distinction that matters for host reboots and planned maintenance. Interviewers ask this to see if you've thought about operational scenarios.

<details>
<summary>Answer</summary>

Both restart the container on crashes. The difference is in two scenarios:

| Scenario                               | `always`                    | `unless-stopped`                     |
| -------------------------------------- | --------------------------- | ------------------------------------ |
| Container crashed                      | Restart ✓                   | Restart ✓                            |
| `docker stop mycontainer`              | Restart on daemon restart ✓ | Does NOT restart on daemon restart ✓ |
| Host rebooted (Docker daemon restarts) | Restart ✓                   | Does NOT restart ✓                   |

With `unless-stopped`: once you explicitly stop a container with `docker stop`, it stays stopped even after the Docker daemon restarts (e.g., after a host reboot). This is useful for containers you want to control manually.

With `always`: the container always comes back, even if you just stopped it for maintenance. This can be annoying when you want to take a container offline temporarily.

**Recommendation**: Use `unless-stopped` for long-running services. Use `always` only if you never want to stop a container manually (rare).

In Kubernetes and ECS, the orchestrator manages restarts — set `restart: no` and let the pod/task restart policy handle it.

</details>

---

### Q4: How do you get logs out of a container running in production on AWS?

A practical ops question that tests whether you know logging drivers and production log collection patterns.

<details>
<summary>Answer</summary>

The standard pattern for AWS:

**Option 1: `awslogs` logging driver (simplest)**

```yaml
logging:
  driver: awslogs
  options:
    awslogs-region: us-east-1
    awslogs-group: /myapp/api
    awslogs-stream: "{{.Name}}-{{.ID}}"
    awslogs-create-group: "true"
```

The container needs IAM permission `logs:CreateLogGroup`, `logs:CreateLogStream`, `logs:PutLogEvents`.

Downside: `docker logs` no longer works on the host.

**Option 2: Fluentd/Fluent Bit sidecar**
Run Fluent Bit as a sidecar container sharing a log volume, then ship to CloudWatch, Elasticsearch, or S3. More flexible, supports log enrichment, filtering, and multiple destinations.

**Option 3: stdout/stderr → host logging → CloudWatch agent**
Container writes to stdout → Docker json-file driver → CloudWatch agent on host reads the JSON log files and ships them. No special Docker config needed; the CloudWatch agent handles log collection.

**Option 4: In ECS with FireLens**
ECS supports FireLens as a log router — a Fluent Bit sidecar managed by ECS that intercepts container logs before they hit the Docker logging driver. Most flexible for ECS deployments.

**My recommendation**: For ECS, use FireLens. For standalone EC2, use `awslogs` driver. For Kubernetes, use a DaemonSet log collector (Fluent Bit) that reads from the node's `/var/log/containers/`.

</details>

---

## Key Takeaways

1. **Exec form `CMD ["node", "app.js"]`** makes your process PID 1 and receives signals; shell form does not
2. **`init: true` (tini)** handles PID 1 responsibilities: zombie reaping and signal forwarding
3. **`SIGTERM` handler in your app** is required for graceful shutdown — `docker stop` sends SIGTERM, then SIGKILL after timeout
4. **`start_period` in HEALTHCHECK** gives apps time to initialise before health checks start (prevents early false unhealthy)
5. **Memory OOM kill is `SIGKILL`** — not catchable; the process dies immediately; diagnose with `docker inspect` → `OOMKilled`
6. **`restart: unless-stopped`** for services that should auto-restart but respect explicit `docker stop`
7. **json-file logging with `max-size`/`max-file`** is mandatory — without limits, logs fill disk silently
8. **Non-json-file drivers disable `docker logs`** — use CloudWatch, Fluentd, or the host logging agent instead
9. **Write structured JSON to stdout/stderr** — don't write log files inside containers

## Next Steps

In [Lesson 07: BuildKit & Multi-platform Builds](lesson-07-buildkit-and-multi-platform.md), you'll learn:

- BuildKit's `RUN --mount` types for caching, secrets, and SSH forwarding
- How to build images for multiple architectures with `docker buildx`
- Cache backend strategies for CI (registry, GitHub Actions cache, local)
- Image signing with cosign for supply chain security
