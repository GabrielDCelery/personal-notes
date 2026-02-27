# Lesson 03: Volumes & Storage Strategies

Where data lives, how it survives container restarts, and how to choose the right storage primitive for each use case.

## Why Storage Is Tricky in Containers

Containers are ephemeral by design. Every file written inside a container goes to the writable layer — and disappears when the container is removed. For databases, uploads, logs, and config files, you need data to outlive the container lifecycle.

Docker has three storage mechanisms, and choosing the wrong one causes subtle production bugs:

| Type             | Where data lives                 | Managed by | Persists after `rm` |
| ---------------- | -------------------------------- | ---------- | ------------------- |
| Writable layer   | OverlayFS on host                | Docker     | ❌ No               |
| **Named volume** | `/var/lib/docker/volumes/`       | Docker     | ✓ Yes               |
| **Bind mount**   | Any host path you specify        | You        | ✓ Yes               |
| **tmpfs**        | Host RAM (never written to disk) | OS         | ❌ No (intentional) |

---

## Named Volumes

Named volumes are managed by Docker. You don't need to know where on the host they live — Docker handles the path.

```bash
# Create explicitly
docker volume create pgdata

# Or create implicitly at run time
docker run -d \
  -v pgdata:/var/lib/postgresql/data \
  postgres:16

# List volumes
docker volume ls

# Inspect (see actual host path)
docker volume inspect pgdata
# → Mountpoint: /var/lib/docker/volumes/pgdata/_data
```

### When to use named volumes

- Databases (Postgres, MySQL, Redis persistence)
- Any data that needs to outlive a specific container but doesn't need to be edited from the host
- Shared data between multiple containers

```yaml
# docker-compose.yml
services:
  postgres:
    image: postgres:16
    volumes:
      - pgdata:/var/lib/postgresql/data # ✓ named volume

volumes:
  pgdata: # ← declares it at the project level
```

**Compose volume lifecycle**: `docker compose down` does NOT remove named volumes. You must explicitly `docker compose down -v` or `docker volume rm pgdata` to delete data. This is intentional — it prevents accidental data loss.

---

## Bind Mounts

Bind mounts map a specific host directory into the container. The host directory must exist (or Docker will create it as root-owned, which causes permission issues).

```bash
docker run -d \
  -v /home/gaze/myapp/data:/app/data \
  myapp
```

### When to use bind mounts

- **Local development**: mount source code so the container sees live changes
- **Config files**: inject environment-specific configs without rebuilding
- **Logs**: write directly to a host path monitored by a log shipper

```yaml
# docker-compose.yml — development pattern
services:
  api:
    image: myapi:dev
    volumes:
      - ./src:/app/src # ✓ live reload during development
      - ./config/dev.yaml:/app/config.yaml:ro # ✓ read-only config
```

### Bind mount gotchas

**Permission mismatch**: The container process runs as a UID (e.g., `node` user is UID 1000). If the host directory is owned by root, the container can't write to it.

```bash
# ❌ Directory created by Docker as root, node user can't write
mkdir -p /app/data
docker run -u 1000 -v /app/data:/data myapp  # permission denied

# ✓ Fix: match ownership
chown 1000:1000 /app/data
# Or use :z or :Z for SELinux relabelling on RHEL/Fedora
docker run -v /app/data:/data:z myapp
```

**Overwrites container files**: If the host path is empty, the bind mount hides whatever was in the container at that path. This is different from named volumes (which copy container files to an empty volume on first run).

```bash
# ❌ If ./node_modules doesn't exist on host, it overwrites the container's node_modules
docker run -v ./node_modules:/app/node_modules myapp

# ✓ Use an anonymous volume to preserve container's node_modules
docker run \
  -v ./:/app \
  -v /app/node_modules \  # ← anonymous volume "masks" the bind mount at this path
  myapp
```

---

## tmpfs Mounts

Stores data in host RAM, never written to disk. Automatically cleared when the container stops.

```bash
docker run --tmpfs /tmp:rw,noexec,nosuid,size=100m myapp
```

### When to use tmpfs

- Session data, scratch files, or build intermediates you don't want on disk
- Secrets that must never be written to disk (e.g., decrypted keys in memory only)
- High-frequency temp writes (RAM is faster than disk I/O)

```yaml
services:
  api:
    image: myapi
    tmpfs:
      - /tmp:size=50m,mode=1777
      - /run:size=10m
```

---

## Volume Options: Propagation, SELinux, and Consistency

Mount options are colon-separated suffixes:

```bash
-v /host/path:/container/path[:options]
```

| Option       | Meaning                                                   |
| ------------ | --------------------------------------------------------- |
| `ro`         | Read-only inside the container                            |
| `rw`         | Read-write (default)                                      |
| `z`          | SELinux: relabel with shared label (multiple containers)  |
| `Z`          | SELinux: relabel with private label (single container)    |
| `delegated`  | macOS: container writes may lag behind host (performance) |
| `cached`     | macOS: host view may lag behind container (performance)   |
| `consistent` | macOS: fully consistent (default, slowest)                |

```yaml
volumes:
  - ./nginx.conf:/etc/nginx/nginx.conf:ro # ✓ prevent accidental overwrites
  - ./certs:/etc/ssl/certs:ro,z # ✓ read-only + SELinux relabel
```

---

## Volume Drivers

The default `local` driver stores on the host filesystem. Drivers extend this to remote or distributed storage.

| Driver                   | Storage             | Use case                          |
| ------------------------ | ------------------- | --------------------------------- |
| `local`                  | Host filesystem     | Default, single-host              |
| `local` with NFS options | NFS share           | Multi-host shared storage         |
| `rclone` (plugin)        | S3, GCS, Azure Blob | Cloud object storage              |
| `vieux/sshfs` (plugin)   | Remote SSH          | Development, simple remote access |
| `convoy` (plugin)        | EBS, NFS            | AWS EBS per-container volumes     |

### NFS volume example

```bash
docker volume create \
  --driver local \
  --opt type=nfs \
  --opt o=addr=192.168.1.100,rw \
  --opt device=:/exports/mydata \
  nfs-volume

docker run -v nfs-volume:/data myapp
```

Or in Compose:

```yaml
volumes:
  nfs-data:
    driver: local
    driver_opts:
      type: nfs
      o: addr=192.168.1.100,rw,nfsvers=4
      device: ":/exports/mydata"
```

---

## Backup and Restore Patterns

### Backup a named volume

```bash
# Run a temporary container that mounts the volume and tarballs it to stdout
docker run --rm \
  -v pgdata:/data:ro \
  -v $(pwd)/backups:/backup \
  busybox \
  tar czf /backup/pgdata-$(date +%Y%m%d).tar.gz -C /data .
```

### Restore a named volume

```bash
# Create a fresh volume
docker volume create pgdata-restored

# Extract backup into it
docker run --rm \
  -v pgdata-restored:/data \
  -v $(pwd)/backups:/backup:ro \
  busybox \
  tar xzf /backup/pgdata-20240101.tar.gz -C /data
```

### Postgres-specific backup (preferred for databases)

```bash
# Dump (logical backup, portable)
docker exec postgres pg_dump -U postgres mydb > mydb.sql

# Restore
cat mydb.sql | docker exec -i postgres psql -U postgres mydb
```

### Migrate volumes between hosts

```bash
# On source host:
docker run --rm -v pgdata:/data ubuntu tar czf - /data \
  | ssh user@destination-host "docker run --rm -i -v pgdata:/data ubuntu tar xzf - -C /"
```

---

## Hands-On Exercise 1: Volume Debugging

A developer reports that their Postgres container keeps losing data on restart. Diagnose and fix:

```bash
docker run -d \
  --name postgres \
  -e POSTGRES_PASSWORD=secret \
  postgres:16
```

Then after `docker rm postgres` and recreating, all data is gone.

<details>
<summary>Solution</summary>

**Root cause**: No volume is configured. All data goes to the writable layer which is deleted with the container.

```bash
# Confirm: no volumes mounted
docker inspect postgres --format '{{ json .Mounts }}' | jq
# → []  (empty — no volumes)
```

**Fix:**

```bash
docker run -d \
  --name postgres \
  -e POSTGRES_PASSWORD=secret \
  -v pgdata:/var/lib/postgresql/data \   # ✓ named volume
  postgres:16

# Verify the volume exists and has data
docker volume inspect pgdata
docker volume ls | grep pgdata
```

**Safe removal** (without losing data):

```bash
docker rm postgres           # ✓ container removed, volume pgdata persists
docker rm -v postgres        # ❌ removes container AND its anonymous volumes (not named)
docker volume rm pgdata      # ← explicitly remove the volume
```

**Common confusion**: `docker rm -v` only removes anonymous volumes (those without a name), not named volumes. Named volumes require explicit `docker volume rm`.

</details>

---

## Hands-On Exercise 2: Dev vs Prod Volume Strategy

Design volume configuration for a Node.js web app with these requirements:

- **Development**: Live code reloading, no rebuilds
- **Production**: No source code on host, fast immutable image
- The app writes uploaded files to `/app/uploads`
- Logs are written to `/app/logs`

<details>
<summary>Solution</summary>

```yaml
# docker-compose.yml (development)
services:
  api:
    image: myapi:dev
    volumes:
      - ./src:/app/src # ✓ bind mount for live reload
      - ./src:/app/src:cached # (macOS: cached is faster for read-heavy)
      - /app/node_modules # ✓ anonymous volume preserves node_modules
      - uploads:/app/uploads # ✓ named volume for uploads (shared w/ other services)
      - ./logs:/app/logs # ✓ bind mount logs to host for easy access

volumes:
  uploads:
```

```yaml
# docker-compose.prod.yml (production)
services:
  api:
    image: myapi:prod # ✓ immutable image with source baked in
    volumes:
      # No source code bind mount        # ✓ source is in the image
      - uploads:/app/uploads # ✓ named volume persists between deploys
      - logs:/app/logs # ✓ named volume (or use logging driver instead)

volumes:
  uploads:
  logs:
```

Deploy production:

```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

**Key decisions:**

- `node_modules` anonymous volume prevents the host's missing/different `node_modules` from masking the container's installed deps
- Named `uploads` volume persists files between container replacements (deploys)
- In production, prefer a logging driver (`awslogs`, `fluentd`) over a volume for logs

</details>

---

## Interview Questions

### Q1: What is the difference between a named volume and a bind mount, and when would you choose one over the other?

The classic storage question. Interviewers want to see you understand that "named volume" isn't just a shorthand — it has different behaviour around initial population and permission management.

<details>
<summary>Answer</summary>

**Named volume**: Docker manages the path (`/var/lib/docker/volumes/<name>/_data`). Docker copies container files into an **empty** named volume on first mount (unlike bind mounts).

**Bind mount**: You control the host path. Docker does NOT pre-populate — if the host directory is empty, it hides the container's files at that path.

|                       | Named volume                                                   | Bind mount                            |
| --------------------- | -------------------------------------------------------------- | ------------------------------------- |
| Host path             | Docker-managed                                                 | You specify                           |
| Initial population    | ✓ Copies from container if empty                               | ❌ Hides container files              |
| Permission management | Docker handles                                                 | You handle                            |
| Portability           | ✓ Docker manages location                                      | ❌ Specific host path required        |
| Use in production     | ✓ Databases, uploads, state                                    | ✓ Configs, secrets injected at deploy |
| Use in development    | ⚠️ Not ideal for source code (can't pre-populate live changes) | ✓ Live code reloading                 |

**Choose named volume when**: data needs to persist beyond container lifecycle, you don't need to edit files directly from the host, or you want Docker to handle the initial setup (like Postgres data directory).

**Choose bind mount when**: you need to inject existing host files (configs, certs), or you want live-reloading during development.

</details>

---

### Q2: Why does `docker compose down` not delete your database data?

Tests whether you understand volume lifecycle and how to avoid accidentally nuking production data.

<details>
<summary>Answer</summary>

`docker compose down` removes containers and networks created by Compose, but **not named volumes**. This is intentional — Compose treats named volumes as independently managed state that outlives any particular deployment.

To also remove volumes:

```bash
docker compose down -v            # removes anonymous AND named volumes
docker volume rm myproject_pgdata # removes a specific named volume
```

**Compose volume naming**: Compose prefixes volumes with the project name (directory name by default):

```yaml
volumes:
  pgdata: # → becomes myproject_pgdata on disk
```

You can override with `name:` to share volumes across projects:

```yaml
volumes:
  pgdata:
    name: shared-pgdata # ← no project prefix, shared across compose files
```

**Gotcha with `docker compose down -v`**: It removes ALL volumes defined in the compose file, including those you might want to keep. Be careful in production scripts.

</details>

---

### Q3: A container process writes a file to `/tmp` and the file grows unboundedly. What happens and how do you prevent it?

Tests knowledge of container resource limits and storage implications — a real production incident pattern.

<details>
<summary>Answer</summary>

By default, `/tmp` inside a container writes to the **writable layer** (OverlayFS). This counts against the container's overlay storage and ultimately the host's disk space in `/var/lib/docker/overlay2/`. Docker doesn't set any default size limit on the writable layer.

If the file grows unboundedly:

1. The host disk fills up
2. Docker daemon may start failing
3. Other containers on the same host are affected

**Prevention options:**

1. **tmpfs** — RAM-backed, automatically cleared when container stops, size-limited:

```bash
docker run --tmpfs /tmp:size=100m myapp
```

2. **Device mapper storage limits** (legacy, daemon-level):

```json
// /etc/docker/daemon.json
{
  "storage-driver": "devicemapper",
  "storage-opts": ["dm.basesize=20G"]
}
```

3. **Log rotation** — if the file is a log, use logging driver options:

```yaml
logging:
  driver: json-file
  options:
    max-size: "10m"
    max-file: "3"
```

4. **Application-level cleanup** — the correct fix for unbounded data is the application itself

</details>

---

### Q4: How do you handle a volume that needs to be shared between multiple containers with different UIDs?

A production operations question that reveals whether you understand the Unix permission model in containerised environments.

<details>
<summary>Answer</summary>

Volumes are raw host filesystem directories — they have standard Unix ownership and permissions. When two containers with different UIDs both need write access, you have several options:

**Option 1: Use a shared GID and `setgid`**

```dockerfile
# Both containers use GID 1000
RUN groupadd -g 1000 shared && \
    usermod -aG shared appuser

# On the volume directory, set group-write and setgid:
RUN chmod g+s /shared-data
```

**Option 2: Entrypoint fixup script**

```bash
#!/bin/sh
# entrypoint.sh — run as root, then drop privileges
chown -R ${APP_UID:-1000}:${APP_GID:-1000} /data
exec gosu ${APP_UID:-1000} "$@"
```

**Option 3: Run as root (⚠️ security trade-off)**
Only acceptable for internal tooling, never for internet-facing services.

**Option 4: Use a sidecar to manage permissions**
A short-lived init container (like in Kubernetes `initContainers`) that runs `chmod`/`chown` before the main container starts.

**Option 5: Named pipe / Unix socket**
For IPC rather than file sharing — one container writes to a socket, the other reads from it. Avoids shared filesystem permissions entirely.

</details>

---

## Key Takeaways

1. **Writable layer data is ephemeral** — deleted when the container is removed; use volumes for anything that must persist
2. **Named volumes are populated from the container on first use** — bind mounts are not (they hide container files)
3. **`docker compose down` does not remove named volumes** — add `-v` flag or `docker volume rm` explicitly
4. **Bind mounts in production are dangerous** — host path must exist and have correct permissions; use named volumes for stateful data
5. **tmpfs for secrets and scratch space** — never written to disk, automatically cleared, size-limited
6. **Volume drivers extend storage to NFS, cloud, distributed filesystems** — use when containers run across multiple hosts
7. **Backup volumes via temporary containers** — mount the volume read-only, tar to stdout/host path
8. **Permission mismatches are the #1 bind mount issue** — match host directory UID to container process UID

## Next Steps

In [Lesson 04: Compose Advanced Patterns](lesson-04-compose-advanced-patterns.md), you'll learn:

- Compose profiles for environment-specific service sets
- `depends_on` with health check conditions (not just startup order)
- Secrets and configs — the right way to inject sensitive data
- Extension fields (`x-`) for DRY Compose files
- Override files for dev/prod configuration splits
