# Lesson 08: Debugging & Troubleshooting

Systematic diagnosis of container failures — from layer analysis to kernel-level namespace inspection.

## The Debugging Mindset

Production container issues fall into a small number of categories. Rather than random `docker exec` sessions, work through a systematic checklist:

1. **Is the container running?** (`docker ps`, exit codes)
2. **What are the logs saying?** (`docker logs`)
3. **What is the container's configuration?** (`docker inspect`)
4. **Is the network correct?** (DNS, ports, connectivity)
5. **Is storage correct?** (volumes, permissions)
6. **Is the image correct?** (layer analysis, what's actually inside)
7. **What is the process doing right now?** (`exec`, `strace`, `nsenter`)

---

## `docker inspect`: The First Tool to Reach For

`docker inspect` returns the full JSON configuration and runtime state of any Docker object: containers, images, networks, volumes.

```bash
# Full output (piped to jq for readability)
docker inspect mycontainer | jq

# Specific fields with Go template format
docker inspect mycontainer --format '{{ .State.Status }}'          # running
docker inspect mycontainer --format '{{ .State.ExitCode }}'        # 0, 1, 137, etc.
docker inspect mycontainer --format '{{ .State.OOMKilled }}'       # true if OOM
docker inspect mycontainer --format '{{ .State.Error }}'           # error message
docker inspect mycontainer --format '{{ .NetworkSettings.IPAddress }}'
docker inspect mycontainer --format '{{ json .Mounts }}' | jq
docker inspect mycontainer --format '{{ json .HostConfig.PortBindings }}' | jq
docker inspect mycontainer --format '{{ json .Config.Env }}' | jq  # env vars
```

### Exit codes

| Code | Signal       | Meaning                                              |
| ---- | ------------ | ---------------------------------------------------- |
| 0    | —            | Clean exit                                           |
| 1    | —            | Application error                                    |
| 137  | SIGKILL (9)  | Killed by OOM or `docker kill`/`docker stop` timeout |
| 139  | SIGSEGV (11) | Segmentation fault                                   |
| 143  | SIGTERM (15) | Graceful termination (should be 0 after cleanup)     |

Exit code = 128 + signal number for killed-by-signal.

```bash
# Quick exit code diagnosis
docker inspect mycontainer --format '{{ .State.ExitCode }} OOM={{ .State.OOMKilled }}'
```

---

## `docker logs`: Reading Container Output

```bash
docker logs mycontainer                   # all logs
docker logs -f mycontainer                # follow (like tail -f)
docker logs --tail 100 mycontainer        # last 100 lines
docker logs --since 10m mycontainer       # last 10 minutes
docker logs --since 2024-01-01T00:00:00 mycontainer
docker logs -t mycontainer                # include timestamps

# Both stdout and stderr
docker logs mycontainer 2>&1 | grep ERROR
```

**Gotcha**: If a container exited, `docker logs` still works (reads from the json-file). If you used a non-json-file logging driver (awslogs, fluentd), `docker logs` returns nothing — check your log aggregator.

```bash
# Check which logging driver is in use
docker inspect mycontainer --format '{{ .HostConfig.LogConfig.Type }}'
# → awslogs  ← docker logs won't work
```

---

## `docker exec`: Live Container Inspection

```bash
# Open a shell in a running container
docker exec -it mycontainer sh            # alpine/busybox
docker exec -it mycontainer bash          # debian/ubuntu

# Run a one-off command
docker exec mycontainer env               # list environment variables
docker exec mycontainer ps aux            # list processes
docker exec mycontainer cat /etc/hosts    # read internal files
docker exec mycontainer netstat -tlnp     # open ports inside container

# Run as a different user
docker exec -u root mycontainer sh        # switch to root for admin tasks
```

### When there's no shell (distroless, scratch)

```bash
# ❌ Can't exec into a distroless container
docker exec -it mycontainer sh            # OCI runtime exec failed: no shell

# Option 1: Use debug image variant
# gcr.io/distroless/nodejs20-debian12:debug has busybox shell
docker run gcr.io/distroless/nodejs20-debian12:debug sh

# Option 2: nsenter (see below)

# Option 3: Copy a debugging binary into a temporary container
docker run --pid=container:mycontainer \
  --net=container:mycontainer \
  --volumes-from mycontainer \
  busybox sh                              # ← shares process/net/filesystem with target
```

### `docker run --rm` for one-off debugging

```bash
# Debug a container's filesystem without running the actual command
docker run --rm -it --entrypoint sh myimage

# Debug networking from inside the container's network namespace
docker run --rm --network container:mycontainer nicolaka/netshoot
```

`netshoot` is a container image packed with network debugging tools: `tcpdump`, `nmap`, `dig`, `curl`, `iperf`, `traceroute`.

---

## `nsenter`: Kernel-Level Namespace Debugging

`nsenter` lets you enter the namespaces of a running container from the host. Useful when the container has no shell (distroless) or when you need host-level tools inside the container's namespace.

```bash
# Get the container's PID on the host
PID=$(docker inspect mycontainer --format '{{ .State.Pid }}')

# Enter all namespaces (network, mount, PID, UTS, IPC)
nsenter --target $PID --mount --uts --ipc --net --pid -- sh

# Enter just the network namespace (for network debugging)
nsenter --target $PID --net -- ip addr
nsenter --target $PID --net -- ss -tlnp
nsenter --target $PID --net -- tcpdump -i eth0 -n

# Enter just the mount namespace (to see the container's filesystem)
nsenter --target $PID --mount -- ls /
nsenter --target $PID --mount -- cat /etc/nginx/nginx.conf
```

**Requirements**: Root on the host. `nsenter` is in `util-linux`, available on all major Linux distros.

---

## Layer Analysis with `dive`

`dive` is a TUI tool for exploring image layers — what was added, modified, or deleted in each layer, and what's wasting space.

```bash
# Install
brew install dive                          # macOS
apt-get install dive                       # Linux

# Analyse an image
dive myimage:latest

# CI mode — fails if wasted space > threshold
dive --ci --lowestEfficiency 0.9 myimage:latest
```

### `docker history` without dive

```bash
docker history myimage
# IMAGE          CREATED        CREATED BY                                SIZE
# abc123         2 hours ago    CMD ["node" "dist/index.js"]               0B
# def456         2 hours ago    COPY . .                                   1.2MB
# ghi789         2 hours ago    RUN npm ci --omit=dev                     45.6MB
# ...
```

**Identify large layers**:

```bash
docker history myimage --format '{{.Size}}\t{{.CreatedBy}}' | sort -rh | head -20
```

### Common image bloat causes

| Cause                               | Fix                                                                   |
| ----------------------------------- | --------------------------------------------------------------------- |
| `node_modules` with devDependencies | `npm ci --omit=dev` in production stage                               |
| APT/APK cache not cleaned           | `rm -rf /var/lib/apt/lists/*` in same `RUN`                           |
| Build tools in runtime image        | Multi-stage build                                                     |
| `.git` directory copied             | `.dockerignore` with `.git/`                                          |
| Test fixtures, docs                 | `.dockerignore` or multi-stage                                        |
| npm cache                           | `npm cache clean --force` in same `RUN` (or use `--mount=type=cache`) |

---

## Container Resource Monitoring

```bash
# Real-time stats for all containers
docker stats

# Specific container
docker stats mycontainer

# JSON output for scripting
docker stats mycontainer --no-stream --format json | jq
```

Output fields:

| Field               | Meaning                 |
| ------------------- | ----------------------- |
| `CPU %`             | Percentage of total CPU |
| `MEM USAGE / LIMIT` | Current vs limit        |
| `MEM %`             | Memory usage % of limit |
| `NET I/O`           | Network bytes in/out    |
| `BLOCK I/O`         | Disk bytes read/written |
| `PIDS`              | Number of processes     |

### Beyond `docker stats`

- **cAdvisor**: Runs as a container, exposes Prometheus metrics for all containers
- **Prometheus + Grafana**: Long-term resource trending
- **`pidstat`, `iotop`, `nethogs`** on the host for per-process analysis

```bash
# Check actual memory usage (RSS, not virtual)
docker exec mycontainer cat /sys/fs/cgroup/memory/memory.usage_in_bytes

# CPU throttling (is the container hitting its CPU limit?)
docker exec mycontainer cat /sys/fs/cgroup/cpu/cpu.stat
# nr_throttled > 0 means the container is CPU-limited
```

---

## Common Failure Patterns

### Pattern 1: Container exits immediately (exit 0 or exit 1)

```bash
# Diagnose
docker logs mycontainer              # what did the process output?
docker inspect mycontainer --format '{{ .State.ExitCode }}'

# Container ran fine but CMD exited — e.g., `CMD ls` which finishes instantly
# Fix: make CMD run a long-lived process, not a one-shot command
```

**Common causes**:

- `CMD` runs a script that exits (not a daemon)
- Application crash on startup (missing env var, bad config)
- Entrypoint script exits non-zero before starting the main process

### Pattern 2: OOM Kill (exit 137)

```bash
docker inspect mycontainer --format '{{ .State.OOMKilled }}'   # → true
docker inspect mycontainer --format '{{ .State.ExitCode }}'    # → 137
dmesg | tail -20 | grep -i oom                                  # host kernel logs
```

**Fix**: Increase memory limit or fix memory leak in the application.

### Pattern 3: Port conflict

```bash
# Error: bind: address already in use
docker logs mycontainer | grep -i "address already in use"

# Find what's using the port
ss -tlnp | grep :8080
lsof -i :8080

# Check if another container has it
docker ps --format '{{.Names}}\t{{.Ports}}' | grep 8080
```

### Pattern 4: Permission denied on volume

```bash
# Inside container
docker exec mycontainer ls -la /data
# → drwxr-xr-x 2 root root 4096 ...  ← owned by root, container runs as 1000

# Fix: chown on host
sudo chown -R 1000:1000 /host/data/path

# Or use an entrypoint to fix permissions
docker run --user root -e TARGET_UID=1000 myimage chown -R $TARGET_UID /data
```

### Pattern 5: Container can't reach another container (DNS failure)

```bash
# Check they're on the same network
docker inspect serviceA --format '{{ json .NetworkSettings.Networks }}' | jq 'keys'
docker inspect serviceB --format '{{ json .NetworkSettings.Networks }}' | jq 'keys'

# Test DNS from inside the container
docker exec serviceA nslookup serviceB
docker exec serviceA ping serviceB

# Check embedded DNS server
docker exec serviceA cat /etc/resolv.conf
# Should show: nameserver 127.0.0.11

# Test with netshoot
docker run --rm --network container:serviceA nicolaka/netshoot \
  dig @127.0.0.11 serviceB
```

### Pattern 6: Container starts but health check fails

```bash
# Check health check history
docker inspect mycontainer --format '{{ json .State.Health }}' | jq

# Test the health check command manually inside the container
docker exec mycontainer wget -qO- http://localhost:3000/health

# Common causes:
# - App starts on 0.0.0.0 but health check uses 127.0.0.1 (or vice versa)
# - App not ready within start_period
# - Health check tool (curl/wget) not installed in image
```

### Pattern 7: Image pull failure in production

```bash
# Rate limiting (Docker Hub)
docker pull myimage  # toomanyrequests: You have reached your pull rate limit

# Fix: authenticate to Docker Hub
docker login
# Or use a pull-through cache in ECR/GCR/ACR

# Wrong architecture
docker pull myimage  # exec format error
docker inspect myimage --format '{{ .Architecture }}'
# Fix: build/use a multi-platform image or specify --platform

# Registry auth
docker login myregistry.azurecr.io
# Or via Kubernetes: imagePullSecrets
```

---

## Hands-On Exercise 1: Debug a Broken Container

The following container starts but its web server is unreachable. Find and fix the issue without modifying the Dockerfile.

```bash
docker run -d --name broken -p 8080:8080 nginx:alpine
curl http://localhost:8080   # ❌ connection refused
```

<details>
<summary>Solution</summary>

```bash
# Step 1: Is the container running?
docker ps | grep broken
# ← shows it's up

# Step 2: What do the logs say?
docker logs broken
# nginx: [emerg] bind() to 0.0.0.0:80 failed... or no errors, nginx started fine

# Step 3: What port is nginx actually listening on?
docker exec broken netstat -tlnp
# tcp  0  0 0.0.0.0:80  0.0.0.0:*  LISTEN  nginx

# Step 4: What port mapping did we set?
docker inspect broken --format '{{ json .HostConfig.PortBindings }}' | jq
# {"8080/tcp": [{"HostIp": "", "HostPort": "8080"}]}

# Found it! Nginx listens on port 80, but we mapped host:8080 → container:8080
# The container:8080 has nothing listening on it.

# Fix: correct port mapping
docker stop broken && docker rm broken
docker run -d --name fixed -p 8080:80 nginx:alpine  # ✓ host:8080 → container:80
curl http://localhost:8080   # ✓ works
```

**Lessons**:

- Always verify with `docker exec <container> netstat -tlnp` what ports the process actually listens on
- The `-p` format is `host_port:container_port` — easy to reverse
- `docker ps` shows the mapping but only after the fact

</details>

---

## Hands-On Exercise 2: Image Size Audit

This image is 1.8GB. Identify what's taking the space and reduce it below 200MB.

```dockerfile
FROM node:20
WORKDIR /app
COPY . .
RUN npm install
RUN npm run build
CMD ["node", "dist/index.js"]
```

<details>
<summary>Solution</summary>

```bash
# Step 1: Check layer sizes
docker history myimage --format '{{.Size}}\t{{.CreatedBy}}' | sort -rh | head -10
# 900MB  FROM node:20                    ← base image is huge
# 500MB  COPY . .                        ← node_modules copied from host!
# 300MB  RUN npm install                 ← deps + devDependencies

# Step 2: Check with dive (interactive)
dive myimage

# Step 3: What's in the node_modules?
docker run --rm myimage du -sh /app/node_modules
# → 480MB — includes devDependencies (TypeScript compiler, test tools, etc.)

# Step 4: Check if .dockerignore exists
cat .dockerignore  # → doesn't exist or node_modules not listed!
# node_modules from the host (with devDeps) was COPYed in, then npm install added MORE
```

**Fixed Dockerfile:**

```dockerfile
# syntax=docker/dockerfile:1

FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci                      # install all deps for build
COPY . .
RUN npm run build               # compile TypeScript → dist/

FROM node:20-alpine AS runtime
WORKDIR /app
COPY package*.json ./
RUN npm ci --omit=dev           # prod deps only (~50MB vs 480MB)
COPY --from=builder /app/dist ./dist
CMD ["node", "dist/index.js"]
```

**.dockerignore:**

```
node_modules/
dist/
.git/
*.md
coverage/
.env*
```

**Result:**

- Before: 1.8GB (node:20 base + devDeps + source + host node_modules)
- After: ~180MB (node:20-alpine + prod deps only + compiled dist)

**Verification:**

```bash
docker build -t myimage-optimised .
docker images myimage-optimised
# → 178MB
```

</details>

---

## Interview Questions

### Q1: A container is running but unreachable. Walk me through your diagnosis process.

The classic ops interview question. Tests systematic thinking, not just Docker command knowledge.

<details>
<summary>Answer</summary>

I work through a layered checklist:

**1. Is the container actually running?**

```bash
docker ps | grep mycontainer
# Check status: Up, Exited, Restarting (crash loop)?
```

**2. What do the logs say?**

```bash
docker logs mycontainer --tail 50
# Error messages, startup failures, crash reasons
```

**3. What port is the process listening on?**

```bash
docker exec mycontainer netstat -tlnp
# or: docker exec mycontainer ss -tlnp
# Verify: is the process listening on the right port AND interface (0.0.0.0 vs 127.0.0.1)?
```

**4. Is the port mapping correct?**

```bash
docker inspect mycontainer --format '{{ json .HostConfig.PortBindings }}' | jq
docker port mycontainer
# host:XXXX → container:YYYY — verify both sides
```

**5. Can I reach it from inside the container? (is it the app or the network?)**

```bash
docker exec mycontainer wget -qO- http://localhost:3000/health
# If this works but external access fails → port mapping or firewall issue
# If this fails → app not listening or crashed
```

**6. Can I reach it from the host?**

```bash
curl http://localhost:8080
# If fails → check iptables, Docker port binding
```

**7. Is there a firewall blocking it?**

```bash
iptables -L DOCKER -n  # Docker's iptables rules
ufw status             # UFW rules (note: Docker bypasses UFW)
```

**8. Is it a DNS issue (for inter-container)?**

```bash
docker exec caller nslookup targetservice
docker exec caller ping targetservice
```

</details>

---

### Q2: A container exited with code 137. What happened and how do you investigate?

Tests understanding of exit codes and OOM diagnosis — a real production troubleshooting scenario.

<details>
<summary>Answer</summary>

Exit code 137 = 128 + 9 (SIGKILL). The process was killed by signal 9. In Docker, this usually means:

1. **OOM kill** — exceeded memory limit, kernel OOM killer sent SIGKILL
2. **`docker stop` timeout** — container didn't respond to SIGTERM in time, Docker sent SIGKILL
3. **Manual kill** — someone ran `docker kill mycontainer`

**Investigate:**

```bash
# Check if it was OOM
docker inspect mycontainer --format '{{ .State.OOMKilled }}'
# → true = OOM kill confirmed

# Check host kernel logs for OOM details
dmesg | grep -i 'oom-kill'
dmesg | grep -A5 'mycontainer'
# Shows which process triggered OOM, how much memory was used

# Check if it was a graceful shutdown failure
docker logs mycontainer --tail 20
# If SIGTERM handler ran, you'd see graceful shutdown messages
```

**If OOM kill:**

- Find peak memory: `docker stats mycontainer --no-stream` (if still running in a new instance)
- Increase `--memory` limit
- Profile the application for memory leaks
- Add JVM/Node.js heap limits: `NODE_OPTIONS=--max-old-space-size=512`

**If stop timeout:**

- The app isn't handling SIGTERM — fix signal handling (Lesson 06)
- Increase `docker stop --time 60 mycontainer` if graceful shutdown legitimately takes longer

</details>

---

### Q3: How do you debug a container that has no shell (distroless or scratch)?

Tests senior-level debugging knowledge. Most candidates only know `docker exec sh`.

<details>
<summary>Answer</summary>

Several approaches:

**1. Use the debug image variant**

```bash
# distroless provides a :debug tag with busybox
docker run -it gcr.io/distroless/nodejs20-debian12:debug sh
# Only in dev/staging — never use debug image in production
```

**2. `nsenter` from the host**

```bash
PID=$(docker inspect mycontainer --format '{{ .State.Pid }}')
# Enter all namespaces — you have host tools, container's FS view
nsenter --target $PID --mount --net --pid -- sh
# Now you have a shell with the container's mount namespace
```

**3. Share namespaces with a debugging container**

```bash
docker run -it \
  --pid=container:mycontainer \     # share PID namespace
  --network=container:mycontainer \ # share network namespace
  --volumes-from mycontainer \      # share volumes
  busybox sh                        # busybox has all the tools you need
```

**4. Ephemeral debug container (Kubernetes `kubectl debug`, Docker equivalent)**

```bash
# Temporarily add a debugging sidecar to the running container group
# (more commonly used in Kubernetes as ephemeral containers)
docker run --rm --pid=container:mycontainer ubuntu \
  strace -p $(pgrep -f myapp)
```

**5. Inspect the filesystem without running the container**

```bash
# Mount the container's filesystem
docker export mycontainer | tar -t | grep suspicious-file
docker cp mycontainer:/app/config.yaml ./extracted-config.yaml
```

</details>

---

### Q4: How do you find out why an image is unexpectedly large?

Image size debugging is a common interview task. Tests knowledge of layer analysis tools and Docker internals.

<details>
<summary>Answer</summary>

**Step 1: `docker history`** — quick layer-by-layer size breakdown

```bash
docker history myimage --format '{{.Size}}\t{{.CreatedBy}}' | sort -rh | head -20
# Immediately identifies which layer is the culprit
```

**Step 2: `dive`** — interactive layer explorer

```bash
dive myimage
# Shows exactly which files were added/modified/deleted in each layer
# Highlights wasted space (files added then deleted across layers)
```

**Step 3: `docker export` + analysis**

```bash
docker create --name tmp myimage
docker export tmp | tar -tv | sort -k5 -rn | head -30  # largest files
docker rm tmp
```

**Common culprits found this way:**

| Finding                                | Fix                                                     |
| -------------------------------------- | ------------------------------------------------------- |
| Large `node_modules` with devDeps      | `npm ci --omit=dev` in final stage                      |
| `.git/` directory                      | Add `.git/` to `.dockerignore`                          |
| Build tools (gcc, make) in final image | Multi-stage build                                       |
| APT/APK cache                          | `rm -rf /var/lib/apt/lists/*` in same RUN               |
| Files deleted in later layers          | Combine into single RUN or multi-stage                  |
| Test fixtures, docs                    | `.dockerignore`                                         |
| npm cache                              | `npm cache clean` in same RUN (or `--mount=type=cache`) |
| Log files from build steps             | Multi-stage or combined RUN                             |

**After fixing**, verify the improvement:

```bash
docker images myimage --format '{{ .Size }}'
```

</details>

---

## Key Takeaways

1. **`docker inspect`** is your primary diagnosis tool — use Go templates to extract exactly the field you need
2. **Exit code 137** = SIGKILL (OOM or timeout); check `OOMKilled` field and `dmesg` for confirmation
3. **`docker logs`** requires `json-file` driver — non-file drivers mean logs are in your aggregator, not on the host
4. **`nsenter --target $PID --net`** enters a container's network namespace from the host — works even for distroless images
5. **Sharing namespaces** (`--pid=container:X`, `--network=container:X`) lets a debugging container see a distroless container's environment
6. **`docker history` + `dive`** identify which layer is bloating an image — always check both
7. **Port issues**: verify what the process actually listens on with `netstat -tlnp` inside the container, then verify the `-p` mapping
8. **Systematic diagnosis** beats random `exec` sessions — work through: running? logs? process? ports? network? storage?

---

## Series Complete

You've covered the full advanced Docker skill set:

| Lesson                                              | Topic                                                           |
| --------------------------------------------------- | --------------------------------------------------------------- |
| [01](lesson-01-image-internals-and-optimization.md) | OverlayFS, layer caching, multi-stage builds, distroless        |
| [02](lesson-02-networking-deep-dive.md)             | Network namespaces, bridge/overlay/host, DNS, port binding      |
| [03](lesson-03-volumes-and-storage-strategies.md)   | Named volumes, bind mounts, tmpfs, volume drivers, backup       |
| [04](lesson-04-compose-advanced-patterns.md)        | Profiles, healthcheck depends_on, secrets, override files       |
| [05](lesson-05-security-hardening.md)               | Non-root, read-only FS, capabilities, seccomp, secrets          |
| [06](lesson-06-production-patterns.md)              | PID 1, signal handling, health checks, resource limits, logging |
| [07](lesson-07-buildkit-and-multi-platform.md)      | BuildKit mounts, multi-platform, cache backends, image signing  |
| [08](lesson-08-debugging-and-troubleshooting.md)    | Inspect, exec, nsenter, layer analysis, failure patterns        |
