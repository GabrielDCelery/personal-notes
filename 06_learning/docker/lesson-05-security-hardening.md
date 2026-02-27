# Lesson 05: Security Hardening

Reducing container attack surface — non-root users, read-only filesystems, capabilities, seccomp, and secrets.

## Why Container Security Is a Distinct Skill

Running as root in a container is not the same as running as root on a bare-metal machine — but it's closer than most developers realise. A container breach combined with a kernel exploit or misconfiguration can mean full host compromise. These aren't theoretical: CVE-2019-5736 (runc escape), CVE-2020-15257 (containerd socket exposure), and others have demonstrated real escape vectors.

Security hardening is also a hiring signal: interviewers at companies with mature DevSecOps practices will ask about these topics specifically.

---

## Non-Root Users

By default, processes inside containers run as root (UID 0). This is bad because:

- Root inside a container can write to any file owned by root on bind-mounted host paths
- Some container escape vulnerabilities only work as root
- Container runtimes have `--privileged` and capabilities that amplify root's power

### Setting a non-root user in the Dockerfile

```dockerfile
FROM node:20-alpine

WORKDIR /app

# Create a dedicated user and group
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

COPY package*.json ./
RUN npm ci --omit=dev

COPY . .
RUN chown -R appuser:appgroup /app    # ← transfer ownership before switching

USER appuser                          # ← drop to non-root

EXPOSE 3000
CMD ["node", "dist/index.js"]
```

### Official images and built-in users

Many official images include a non-root user already:

| Image          | Built-in user | UID  |
| -------------- | ------------- | ---- |
| `node:*`       | `node`        | 1000 |
| `nginx:alpine` | `nginx`       | 101  |
| `postgres:*`   | `postgres`    | 999  |
| `redis:*`      | `redis`       | 999  |

```dockerfile
# ✓ Use the built-in user instead of creating one
FROM node:20-alpine
USER node
```

### Overriding at runtime

```bash
docker run -u 1000:1000 myimage        # numeric UID:GID
docker run --user node myimage          # named user (must exist in container)
```

```yaml
# docker-compose.yml
services:
  api:
    image: myapi
    user: "1000:1000"
```

### Common non-root gotchas

```dockerfile
# ❌ npm ci after USER switch — npm can't write to /app if root owns it
FROM node:20-alpine
WORKDIR /app
USER node                              # ← switched too early
COPY package*.json ./
RUN npm ci                             # ← fails: /app/node_modules owned by root

# ✓ Install as root, then switch
FROM node:20-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --omit=dev
USER node                              # ← switch after install
```

Ports below 1024 require root (`CAP_NET_BIND_SERVICE`) — use a port above 1024 in your app and map externally:

```dockerfile
# ✓ App listens on 3000, mapped to 80 externally
EXPOSE 3000
# docker run -p 80:3000 myimage
```

---

## Read-Only Root Filesystem

`--read-only` mounts the container's root filesystem as read-only. Any write attempt fails with `EROFS`. This prevents an attacker from modifying binaries or writing scripts.

```bash
docker run --read-only myimage
```

```yaml
services:
  api:
    image: myapi
    read_only: true
    tmpfs:
      - /tmp # ← whitelist writable dirs explicitly
      - /run
      - /app/tmp
```

### What breaks (and how to fix it)

| What writes to root FS | Fix                                             |
| ---------------------- | ----------------------------------------------- |
| Node.js tmp files      | `--tmpfs /tmp`                                  |
| npm cache              | Set `npm_config_cache=/tmp/npm`                 |
| PID files              | Use tmpfs `/run`                                |
| Log files              | Write to stdout/stderr instead (best), or tmpfs |
| Static assets (nginx)  | Add tmpfs `/var/cache/nginx` and `/var/run`     |

```yaml
# nginx with read-only root
services:
  nginx:
    image: nginx:alpine
    read_only: true
    tmpfs:
      - /var/cache/nginx
      - /var/run
      - /tmp
```

---

## Linux Capabilities

Linux divides root's privileges into ~40 distinct capabilities. Docker drops most of them by default, keeping a minimal set. You should drop more.

### What Docker keeps by default

```
CHOWN, DAC_OVERRIDE, FSETID, FOWNER, MKNOD, NET_RAW,
SETGID, SETUID, SETFCAP, SETPCAP, NET_BIND_SERVICE,
SYS_CHROOT, KILL, AUDIT_WRITE
```

### Drop everything, add back only what you need

```bash
# ✓ Drop all capabilities, add back only what's needed
docker run \
  --cap-drop ALL \
  --cap-add NET_BIND_SERVICE \   # if binding to port < 1024
  myimage
```

```yaml
services:
  api:
    image: myapi
    cap_drop:
      - ALL
    cap_add:
      - NET_BIND_SERVICE # only if needed
```

### Common dangerous capabilities to drop

| Capability          | What it allows                              | Should you have it?     |
| ------------------- | ------------------------------------------- | ----------------------- |
| `SYS_ADMIN`         | Almost everything (mount, namespaces, etc.) | Almost never            |
| `NET_RAW`           | Raw sockets, packet sniffing                | Only network tools      |
| `SYS_PTRACE`        | Debug/trace processes                       | Only debuggers          |
| `SYS_MODULE`        | Load kernel modules                         | Never in app containers |
| `DAC_OVERRIDE`      | Bypass file permission checks               | Rarely                  |
| `SETUID` / `SETGID` | Change UID/GID                              | Only if app needs it    |

### `--privileged` — never use it in production

`--privileged` gives the container ALL capabilities and disables seccomp, AppArmor, and device restrictions. It's essentially running as root on the host with extra steps.

```bash
# ❌ Never in production
docker run --privileged myimage

# ✓ If you need a specific capability, add it explicitly
docker run --cap-add SYS_PTRACE myimage   # for debuggers
docker run --device /dev/snd myimage      # for audio device access
```

---

## Seccomp Profiles

Seccomp (Secure Computing Mode) is a Linux kernel feature that filters which system calls a process can make. Docker applies a default seccomp profile that blocks ~300 dangerous syscalls out of ~400 total.

```bash
# Verify seccomp is active
docker run --rm alpine cat /proc/1/status | grep Seccomp
# Seccomp: 2   → 2 means filter mode (active)
# Seccomp: 0   → 0 means disabled
```

### Docker's default seccomp profile blocks

- `ptrace`, `kexec_load`, `open_by_handle_at`, `init_module`, `create_module`
- Most socket family creation (Netlink, etc.)
- Clock manipulation syscalls
- Namespace-related admin syscalls

### Custom seccomp profiles

```bash
# Disable seccomp (⚠️ only for debugging)
docker run --security-opt seccomp=unconfined myimage

# Apply custom profile
docker run --security-opt seccomp=/path/to/profile.json myimage
```

Custom profiles are JSON files (Docker provides the default at [moby/moby/profiles/seccomp](https://github.com/moby/moby/blob/master/profiles/seccomp/default.json)). Start with the default and add `"action": "SCMP_ACT_ALLOW"` entries for specific syscalls your app needs.

---

## AppArmor

AppArmor is a Linux MAC (Mandatory Access Control) system. Docker applies a default `docker-default` AppArmor profile that restricts file access patterns, mount operations, and signal sending to processes outside the container.

```bash
# Check AppArmor status
docker run --rm alpine cat /proc/1/attr/current
# docker-default (enforce)
```

```bash
# Disable AppArmor for a container (⚠️ debugging only)
docker run --security-opt apparmor=unconfined myimage

# Custom AppArmor profile
docker run --security-opt "apparmor=my-custom-profile" myimage
```

On RHEL/Fedora/Amazon Linux, SELinux replaces AppArmor:

```bash
docker run --security-opt label=type:container_t myimage
```

---

## Secrets Management at Runtime

### Never in environment variables for sensitive values

```yaml
# ❌ Secret in env var — visible in docker inspect, logs, process listings
environment:
  DATABASE_PASSWORD: mysecret

# ❌ Secret in ARG — stored in image history
ARG DATABASE_PASSWORD
RUN curl -u admin:${DATABASE_PASSWORD} https://api.example.com
```

### Compose secrets (mounted as files)

See Lesson 04 for full Compose secrets setup. The application reads:

```javascript
const password = fs.readFileSync("/run/secrets/db_password", "utf8").trim();
```

### Docker Swarm secrets (encrypted at rest)

```bash
echo "mysupersecretpassword" | docker secret create db_password -
docker service create \
  --secret db_password \
  --name api myapi
```

Secrets in Swarm are encrypted in the Raft log and only decrypted in-memory on worker nodes that need them.

### AWS Secrets Manager / SSM Parameter Store pattern

For production on cloud infrastructure, don't store secrets in Docker at all. Use an init container or entrypoint script to fetch secrets at startup:

```bash
#!/bin/sh
# entrypoint.sh
export DATABASE_URL=$(aws ssm get-parameter \
  --name /myapp/prod/database-url \
  --with-decryption \
  --query Parameter.Value \
  --output text)

exec "$@"
```

---

## Image Scanning

Image scanning detects known CVEs in base image packages and dependencies.

### Trivy (recommended, open source)

```bash
# Scan an image
trivy image myapp:latest

# Scan only critical and high vulnerabilities
trivy image --severity CRITICAL,HIGH myapp:latest

# Scan in CI (non-zero exit on found vulns)
trivy image --exit-code 1 --severity CRITICAL myapp:latest
```

### Docker Scout (built into Docker CLI)

```bash
docker scout cves myapp:latest
docker scout recommendations myapp:latest   # suggests base image updates
```

### What to do with scan results

1. Update base image (`FROM node:20-alpine` → `FROM node:20.11-alpine`)
2. Update OS packages (`RUN apk upgrade --no-cache`)
3. Accept risk for packages with no fix available (document the exception)
4. Pin to specific digest for reproducibility:

```dockerfile
# ✓ Pinned to digest — immune to tag mutation
FROM node:20-alpine@sha256:a1b2c3d4...
```

---

## Rootless Docker

Rootless mode runs the entire Docker daemon as a non-root user. Container processes' root maps to the user's subUID range via user namespaces.

```bash
# Install and start rootless Docker
dockerd-rootless-setuptool.sh install
systemctl --user start docker

# Rootless containers map root to unprivileged host UIDs
docker run --rm busybox id
# uid=0(root) gid=0(root)  ← inside container looks like root
# Host: uid=100000 gid=100000 ← actual host UID
```

**Trade-offs**:

- ✓ Even if container escapes, attacker gets an unprivileged host user
- ❌ Some features require privileged mode (macvlan, overlay networking, host networking)
- ❌ Performance overhead from user namespace mapping
- ❌ Not all CI environments support rootless

---

## Hands-On Exercise 1: Harden an API Container

Apply security hardening to this Compose service. Target: non-root user, read-only FS, cap drop, and no new privileges.

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
    ports:
      - "3000:3000"
    user: "1000:1000" # ✓ non-root
    read_only: true # ✓ read-only root FS
    tmpfs:
      - /tmp:size=50m # ✓ whitelist temp dir
    cap_drop:
      - ALL # ✓ drop all capabilities
    security_opt:
      - no-new-privileges:true # ✓ prevent setuid escalation
    secrets:
      - db_url # ✓ secret as file, not env var

secrets:
  db_url:
    file: ./secrets/database_url.txt
```

And update the Dockerfile:

```dockerfile
FROM node:20-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --omit=dev
COPY . .
RUN chown -R node:node /app
USER node                              # ✓ non-root in image
CMD ["node", "dist/index.js"]
```

**What `no-new-privileges` does**: Prevents the process from gaining more privileges via `setuid` binaries or file capabilities. Even if the container has a setuid binary (like `sudo`), it can't be used to escalate.

</details>

---

## Hands-On Exercise 2: Identify Security Issues

List all security issues in this production Dockerfile and Compose snippet:

```dockerfile
FROM ubuntu:latest
RUN apt-get update && apt-get install -y curl nodejs npm
WORKDIR /app
COPY . .
RUN npm install
EXPOSE 80
CMD ["node", "server.js"]
```

```yaml
services:
  api:
    build: .
    privileged: true
    ports:
      - "80:80"
    environment:
      DB_PASSWORD: mysupersecretpassword
      JWT_SECRET: abc123
```

<details>
<summary>Solution</summary>

**Dockerfile issues:**

1. ❌ `FROM ubuntu:latest` — unpinned tag, gets different packages on each build; ubuntu is large; use `node:20-alpine`
2. ❌ Installing Node.js via apt — gets an old version; use the official Node image
3. ❌ No non-root `USER` — runs as root
4. ❌ `npm install` instead of `npm ci` — non-deterministic, can modify lockfile
5. ❌ `EXPOSE 80` — ports below 1024 require `CAP_NET_BIND_SERVICE`; use 3000 and map externally
6. ❌ `COPY . .` before splitting deps — cache-busting on every change (not a security issue, a performance one)
7. ❌ No `.dockerignore` — `.env`, `node_modules`, secrets may be copied

**Compose issues:**

1. ❌ `privileged: true` — gives ALL capabilities plus disables seccomp/AppArmor; almost never needed
2. ❌ `DB_PASSWORD` and `JWT_SECRET` in plain `environment:` — visible in `docker inspect`, process listings, CI logs
3. ❌ No `cap_drop`, `read_only`, `security_opt`

**Fixed:**

```dockerfile
FROM node:20-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --omit=dev
COPY --chown=node:node . .
USER node
EXPOSE 3000
CMD ["node", "server.js"]
```

```yaml
services:
  api:
    build: .
    ports:
      - "80:3000"
    user: "1000:1000"
    read_only: true
    tmpfs: [/tmp]
    cap_drop: [ALL]
    security_opt: [no-new-privileges:true]
    secrets:
      - db_password
      - jwt_secret

secrets:
  db_password:
    file: ./secrets/db_password
  jwt_secret:
    file: ./secrets/jwt_secret
```

</details>

---

## Interview Questions

### Q1: What is the difference between `--privileged` and `--cap-add SYS_ADMIN`?

Tests whether you know the blast radius of common Docker run flags. Interviewers at security-conscious companies ask this to filter out candidates who cargo-cult `--privileged`.

<details>
<summary>Answer</summary>

`--cap-add SYS_ADMIN` adds a single capability (a very powerful one, but still one). The container still has seccomp, AppArmor, and device restrictions.

`--privileged` does all of the following simultaneously:

- Adds ALL Linux capabilities
- Disables seccomp filtering
- Disables AppArmor/SELinux
- Enables all `/dev` devices
- Allows mounting filesystems
- Allows loading kernel modules

`--privileged` is effectively: "this container can do anything the root user on the host can do." It's sometimes needed for Docker-in-Docker (DinD) or certain system tools, but should never be used for application containers.

If you think you need `--privileged`, you almost certainly need one specific capability:

- Mount filesystems → `SYS_ADMIN` (still dangerous, but scoped)
- Bind port < 1024 → `NET_BIND_SERVICE`
- Debug another process → `SYS_PTRACE`

</details>

---

### Q2: A container is running as root. Why might this still be safer than running as root on bare metal?

A nuanced question that tests whether you understand the layered security model, not just binary "root bad."

<details>
<summary>Answer</summary>

Running as root inside a container is safer than on bare metal because of multiple isolation layers:

1. **Capabilities** — Docker drops ~14 capabilities by default (even as root). Root inside the container can't, for example, load kernel modules (`SYS_MODULE`) or make raw socket connections (`NET_RAW`), unless explicitly granted.

2. **Seccomp** — Docker's default seccomp profile blocks ~300 dangerous syscalls. Even root can't call `kexec_load`, `create_module`, or `open_by_handle_at`.

3. **Namespaces** — The container has its own PID, mount, network, IPC, and UTS namespaces. Root inside can't see or signal host processes.

4. **AppArmor/SELinux** — Additional MAC layer restricting file access and capabilities.

However, "safer" doesn't mean "safe." Historical kernel exploits have broken through these boundaries. Running as root in a container is still a risk multiplier: if any of these isolation mechanisms has a bug, root gives the attacker maximum leverage. Non-root + capability drop + seccomp + AppArmor is defence in depth.

</details>

---

### Q3: What does `--security-opt no-new-privileges` do and when must you use it?

Tests depth of knowledge on Linux privilege escalation vectors. Often missed by candidates who only know user-level hardening.

<details>
<summary>Answer</summary>

`no-new-privileges` sets the `PR_SET_NO_NEW_PRIVS` flag on the container's init process. This prevents any process in the container from gaining new privileges via:

- `setuid` binaries (e.g., `sudo`, `su`, `passwd`)
- File capabilities (`setcap` on binaries)
- `execve` with a more privileged binary

Without this flag, even a non-root user can escalate privileges if a setuid binary is present in the container image:

```bash
# Without no-new-privileges:
docker run -u 1000 myimage sudo id   # might work if sudo is setuid
docker run -u 1000 myimage /usr/bin/newuidmap 1 0 1000 1  # user namespace attack

# With no-new-privileges:
docker run --security-opt no-new-privileges:true -u 1000 myimage sudo id
# ← fails — cannot gain root even via setuid
```

**Always use it when**: running as a non-root user and you want to ensure no escalation path exists. It's a free additional layer that should be default for all application containers.

</details>

---

### Q4: How would you prevent a secret from appearing in `docker inspect` output?

Tests practical secrets management knowledge — interviewers want to know you've thought about this in production contexts.

<details>
<summary>Answer</summary>

`docker inspect` exposes all environment variables in plaintext. To prevent secrets from appearing there:

**Option 1: Compose secrets (files, not env vars)**

```yaml
secrets:
  db_password:
    file: ./secrets/db_password.txt
services:
  api:
    secrets: [db_password]
    # /run/secrets/db_password inside container, NOT in env vars
```

`docker inspect` does not show secret file contents — it only shows mount points.

**Option 2: Fetch at runtime from a secrets manager**

```bash
# Entrypoint fetches from AWS Secrets Manager — never stored in Docker metadata
export DB_PASSWORD=$(aws secretsmanager get-secret-value --secret-id myapp/db --query SecretString --output text)
```

**Option 3: If you must use env vars, use an init process that reads from a file**

```bash
#!/bin/sh
export DB_PASSWORD="$(cat /run/secrets/db_password)"
exec "$@"
```

The environment variable exists inside the process but was never passed to Docker — it doesn't appear in `docker inspect`.

**What doesn't work**:

- `ARG` instead of `ENV` — ARG is in image history
- Encoding/encrypting values in env vars — still visible in inspect, just obscured

</details>

---

## Key Takeaways

1. **Run as non-root** — use `USER` in Dockerfile or `user:` in Compose; match file ownership before switching
2. **`--read-only` + `tmpfs`** — read-only root FS with explicit writable tmp directories prevents runtime modification of binaries
3. **`cap_drop: ALL` then add-back** — most application containers need zero capabilities after startup
4. **`no-new-privileges`** — prevents setuid escalation even when running as non-root
5. **Secrets as files, not env vars** — `/run/secrets/<name>` is not visible in `docker inspect`; env vars are
6. **`--privileged` is a last resort**, not a "fix" — it disables seccomp, AppArmor, and all capability limits simultaneously
7. **Image scanning** (Trivy, Docker Scout) catches known CVEs — run in CI, block on CRITICAL
8. **Pin base image digests** for reproducible, attack-resistant builds

## Next Steps

In [Lesson 06: Production Patterns](lesson-06-production-patterns.md), you'll learn:

- Why PID 1 matters for signal handling and how to fix it with tini or dumb-init
- Writing effective HEALTHCHECK instructions
- Restart policies and what happens at the orchestration layer
- Resource limits (CPU and memory) and the OOM killer
- Logging drivers and how to get logs out of containers in production
