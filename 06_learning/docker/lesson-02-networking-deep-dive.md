# Lesson 02: Networking Deep Dive

How Docker networking works under the hood — namespaces, drivers, DNS, and inter-container communication.

## Network Namespaces: The Foundation

Every container gets its own network namespace — a kernel-level isolation boundary that gives it a private network stack: its own interfaces, routing table, iptables rules, and loopback. Containers are networkically isolated by default; they can't see each other's traffic unless you connect them.

```bash
# See the network namespace of a running container
docker inspect <container_id> --format '{{ .NetworkSettings.SandboxKey }}'
# → /var/run/docker/netns/abc123

# Enter the container's network namespace from the host
nsenter --net=/var/run/docker/netns/abc123 ip addr
# Shows the container's network interfaces from the host
```

Docker creates a virtual ethernet pair (`veth`) for each container:

- One end (`eth0`) goes into the container's namespace
- The other end goes into the host and is connected to a Docker bridge (`docker0`)

```
Host network namespace
┌──────────────────────────────────────────┐
│  docker0 (bridge)  172.17.0.1            │
│     │                                    │
│   veth0a ──────── veth0b (container eth0)│ 172.17.0.2
│   veth1a ──────── veth1b (container eth0)│ 172.17.0.3
└──────────────────────────────────────────┘
```

---

## Network Drivers Comparison

| Driver    | Isolation               | Use case                           | Performance     |
| --------- | ----------------------- | ---------------------------------- | --------------- |
| `bridge`  | Per-container namespace | Default; single host               | NAT overhead    |
| `host`    | Shares host namespace   | Max performance, no isolation      | Native          |
| `overlay` | Per-container namespace | Multi-host (Swarm/Kubernetes)      | Slight overhead |
| `macvlan` | Per-container namespace | Container needs real MAC/IP on LAN | Native          |
| `ipvlan`  | Shares host MAC         | Like macvlan, avoids MAC flooding  | Native          |
| `none`    | Loopback only           | Maximum isolation                  | N/A             |

---

## Bridge Networks

### Default Bridge (`docker0`)

All containers get attached to `docker0` by default. This is the oldest and most limited option:

- No automatic DNS — containers can only reach each other by IP
- Legacy `--link` flag adds `/etc/hosts` entries (deprecated)
- `172.17.0.0/16` subnet by default

```bash
docker run -d --name app1 nginx
docker run -it --rm busybox ping app1  # ❌ fails — no DNS on default bridge
docker run -it --rm busybox ping 172.17.0.2  # ✓ works by IP (fragile)
```

### Custom Bridge Networks

Creating your own bridge network enables automatic DNS resolution by container name. This is why you should always use custom networks in Compose.

```bash
docker network create mynet

docker run -d --name db --network mynet postgres:16
docker run -d --name api --network mynet myapp

# Inside api container:
ping db          # ✓ resolves via Docker's embedded DNS server (127.0.0.11)
ping api         # ✓ also resolves its own name
```

Docker's embedded DNS server listens on `127.0.0.11` inside each container. It answers queries for container names, service names (in Compose), and aliases.

```bash
# Inspect DNS inside a container
docker exec api cat /etc/resolv.conf
# nameserver 127.0.0.11
# options ndots:0
```

### Network Aliases

A container can have multiple DNS names:

```bash
docker run -d \
  --network mynet \
  --network-alias cache \
  --network-alias redis \
  redis:7
```

In Docker Compose, `links` and `depends_on` don't enable DNS — the service name is the DNS name automatically.

---

## Host Networking

The container shares the host's network namespace entirely — no `veth` pair, no NAT, no bridge.

```bash
docker run --network host nginx
# nginx now binds to host's port 80 directly — no -p flag needed (or possible)
```

**When to use host networking**:

- High-throughput scenarios where NAT overhead matters (e.g., network proxies)
- Monitoring agents that need to see host network interfaces
- Avoid it for application containers — breaks portability and isolation

```bash
# ❌ -p flag is ignored (silently) with --network host
docker run --network host -p 8080:80 nginx  # ⚠️ -p has no effect
```

---

## Overlay Networks (Swarm/Multi-host)

Overlay networks extend a bridge network across multiple hosts using VXLAN encapsulation. Required for Docker Swarm; Kubernetes uses its own CNI plugins instead.

```
Host A                          Host B
┌──────────────────┐            ┌──────────────────┐
│ container1       │            │ container2       │
│ 10.0.0.2         │            │ 10.0.0.3         │
│      │           │            │      │           │
│   overlay0       │────VXLAN───│   overlay0       │
│      │           │            │      │           │
│   eth0 (host)    │            │   eth0 (host)    │
└──────────────────┘            └──────────────────┘
```

VXLAN wraps Layer 2 frames inside UDP packets (port 4789). The overlay driver maintains a distributed key-value store for MAC-to-host mapping.

---

## Macvlan Networks

Each container gets a real MAC address and appears as a physical device on the LAN. The host acts as a trunk:

```bash
docker network create -d macvlan \
  --subnet=192.168.1.0/24 \
  --gateway=192.168.1.1 \
  -o parent=eth0 \
  mymacvlan
```

**Use cases**: Legacy apps that require a specific IP on the physical network, IoT, network appliances.

**Gotcha**: On most switches, you need to enable promiscuous mode on the host NIC. Also, by default the host cannot communicate with macvlan containers (they're on different sub-interfaces).

---

## Exposed vs Published Ports

This is one of the most confused topics at interviews.

|                            | What it does                                                       | Visible from                          |
| -------------------------- | ------------------------------------------------------------------ | ------------------------------------- |
| `EXPOSE 8080` (Dockerfile) | Documents intent, enables inter-container routing on custom bridge | Only other containers on same network |
| `-p 8080:80` / `--publish` | Creates iptables NAT rule, maps host port to container port        | Host and external machines            |
| `-P`                       | Auto-publishes all `EXPOSE`d ports to random host ports            | Host and external machines            |

```bash
# ❌ Common misconception: EXPOSE alone does NOT make a port accessible from the host
# The following exposes port 80 to other containers only
docker run --network mynet nginx   # EXPOSE 80 is in the nginx image

# ✓ To access from the host machine:
docker run -p 8080:80 --network mynet nginx

# Check what's actually published
docker port <container>
```

**What `-p 0.0.0.0:8080:80` means**: Docker binds on all host interfaces. To restrict to localhost only:

```bash
docker run -p 127.0.0.1:8080:80 nginx   # ✓ only accessible from host itself
docker run -p 8080:80 nginx              # ← binds 0.0.0.0 — accessible from network
```

This is a **security issue** in dev environments: `docker run -p 5432:5432 postgres` exposes your database to anyone on your network.

---

## Inter-Container Communication Patterns

### Same network, by name (preferred)

```yaml
# docker-compose.yml
services:
  api:
    image: myapi
    networks: [backend]

  db:
    image: postgres:16
    networks: [backend]

networks:
  backend:
```

`api` can reach `db` at `db:5432`. No port publishing needed.

### Multiple networks (network segmentation)

```yaml
services:
  nginx:
    networks: [frontend, backend] # ← bridge between networks

  api:
    networks: [backend] # ← not directly reachable from internet

  db:
    networks: [backend] # ← only api can reach db

networks:
  frontend:
  backend:
```

### Host-to-container on Linux (from host to container)

```bash
# Direct to container IP (no NAT)
docker inspect mycontainer --format '{{ .NetworkSettings.IPAddress }}'
curl http://172.17.0.2:8080

# Or use published ports
curl http://localhost:8080
```

### Container-to-host on Linux

Use the special hostname `host-gateway` (Docker 20.10+):

```bash
docker run --add-host host.docker.internal:host-gateway myapp
# Inside container: curl http://host.docker.internal:5432 reaches the host
```

On Docker Desktop (Mac/Windows), `host.docker.internal` is automatically available.

---

## Hands-On Exercise 1: Debug a Network Connectivity Issue

Two containers are on different networks and can't communicate. Diagnose and fix:

```bash
docker network create net-a
docker network create net-b
docker run -d --name serviceA --network net-a nginx
docker run -d --name serviceB --network net-b curlimages/curl sleep 3600

# Inside serviceB:
docker exec serviceB curl http://serviceA  # ❌ fails
```

What are two ways to fix this? What's the difference?

<details>
<summary>Solution</summary>

**Option 1: Connect serviceA to net-b**

```bash
docker network connect net-b serviceA
docker exec serviceB curl http://serviceA  # ✓ now works
```

serviceA is now on both networks. DNS in net-b resolves "serviceA" to the interface IP on net-b.

**Option 2: Connect serviceB to net-a**

```bash
docker network connect net-a serviceB
docker exec serviceB curl http://serviceA  # ✓ works via net-a
```

**The difference**: Think about which container should be "reachable" vs which should be the "caller." For security, attach the reaching-out container to the serving container's network, not the other way around. This way, you control which networks expose which services.

**Diagnosing the original failure:**

```bash
# Check what networks each container is on
docker inspect serviceA --format '{{ json .NetworkSettings.Networks }}' | jq
docker inspect serviceB --format '{{ json .NetworkSettings.Networks }}' | jq

# Verify DNS resolution inside the container
docker exec serviceB nslookup serviceA  # ← will fail (wrong network)
```

</details>

---

## Hands-On Exercise 2: Restrict Database Access

Design a Compose network layout for a three-tier app (nginx → api → postgres) where:

- Only nginx's port 80 is published to the host
- api cannot be reached directly from the host
- postgres cannot be reached from nginx or the host

<details>
<summary>Solution</summary>

```yaml
services:
  nginx:
    image: nginx:alpine
    ports:
      - "80:80" # ✓ only nginx is published
    networks:
      - frontend # ✓ shares network with api only
    depends_on:
      - api

  api:
    image: myapi:latest
    networks:
      - frontend # ✓ reachable by nginx
      - backend # ✓ can reach postgres
    # No ports: published — not directly accessible from host

  postgres:
    image: postgres:16
    environment:
      POSTGRES_PASSWORD: secret
    networks:
      - backend # ✓ only api can reach it
    # No ports: published — completely isolated from host

networks:
  frontend:
  backend:
```

**Verification:**

```bash
docker compose up -d

# ✓ nginx reachable from host
curl http://localhost

# ❌ api not reachable from host (no published port)
curl http://localhost:3000   # connection refused

# ❌ postgres not reachable from host
psql -h localhost -p 5432 postgres  # connection refused

# ✓ api can reach postgres
docker compose exec api psql -h postgres -U postgres
```

**Why this matters**: The default Compose network puts ALL services on one shared network. Anyone who can reach the host can reach postgres if you accidentally publish its port. Network segmentation is defence in depth.

</details>

---

## Interview Questions

### Q1: What is the difference between the default bridge network and a user-defined bridge network?

This question tests whether you know why Compose-created networks work for service discovery but the default `docker0` bridge doesn't. It's asked to see if you understand embedded DNS.

<details>
<summary>Answer</summary>

| Feature                                 | Default bridge (`docker0`) | User-defined bridge         |
| --------------------------------------- | -------------------------- | --------------------------- |
| DNS resolution by name                  | ❌ No — IP only            | ✓ Yes via `127.0.0.11`      |
| `--link` required for name resolution   | ✓ (deprecated)             | ❌ Not needed               |
| Network isolation from other containers | ❌ All containers share it | ✓ Only connected containers |
| Configurable subnet/gateway             | ❌ Fixed at creation       | ✓ Yes                       |
| Automatic connect/disconnect at runtime | ❌                         | ✓ `docker network connect`  |

The key difference is Docker's embedded DNS server (`127.0.0.11`). It only resolves container names within user-defined networks. On the default bridge, you must use IPs or the deprecated `--link` flag (which just writes `/etc/hosts` entries).

Docker Compose always creates a user-defined bridge for each project, which is why `ping db` works inside a Compose-managed container.

</details>

---

### Q2: A container publishes port 5432 (`-p 5432:5432`). Is the database accessible from the internet?

Security-focused question that tests whether you understand Docker's iptables manipulation and what `-p` actually does to the host firewall.

<details>
<summary>Answer</summary>

**Yes, potentially** — and this surprises many developers.

`docker run -p 5432:5432 postgres` causes Docker to insert iptables rules in the `DOCKER` chain that bypass the host's `INPUT` chain. This means even if your host firewall (`ufw`, `firewalld`) has a rule blocking port 5432, Docker's rule can override it.

```bash
# Docker adds rules like this (simplified):
iptables -t nat -A DOCKER -p tcp --dport 5432 -j DNAT --to-destination 172.17.0.2:5432
iptables -A DOCKER -d 172.17.0.2/32 -p tcp --dport 5432 -j ACCEPT
```

These rules are in the `FORWARD` chain, which `ufw`'s default rules don't cover for `DOCKER` chain.

**To restrict access:**

```bash
# Bind only to localhost
docker run -p 127.0.0.1:5432:5432 postgres   # ✓ not accessible from outside host
```

This is a well-known security gotcha. Docker's interaction with iptables is one reason some teams add `--iptables=false` in `/etc/docker/daemon.json` and manage rules manually.

</details>

---

### Q3: How does container DNS resolution work and what is `127.0.0.11`?

Interviewers ask this to see whether you understand what's happening when `curl http://db` works inside a container — not just that it works, but the mechanism.

<details>
<summary>Answer</summary>

Every container on a user-defined network has `/etc/resolv.conf` pointing to `127.0.0.11` — Docker's embedded DNS server (also called the "local DNS resolver" or "resolver").

When a container makes a DNS query:

1. The query goes to `127.0.0.11` (intercepted by a iptables rule that redirects it to the Docker daemon's internal resolver)
2. The resolver checks if the name matches a container name, service name, or network alias in the same network
3. If yes, returns the container's network-local IP
4. If no, forwards to the upstream DNS configured on the host (or custom DNS servers configured via `docker network create --dns`)

```bash
docker exec mycontainer cat /etc/resolv.conf
# nameserver 127.0.0.11
# options ndots:0

# The actual interception — Docker adds this iptables rule:
# iptables -t nat -A OUTPUT -d 127.0.0.11 -p udp --dport 53 -j DNAT --to <actual-resolver-port>
```

`ndots:0` means the resolver tries the exact name first before appending search domains — important for short names like `db` or `redis`.

</details>

---

### Q4: When would you use `--network host` and what are the risks?

Asked to check whether you reach for `--network host` for the right reasons, and whether you understand why it's a security trade-off.

<details>
<summary>Answer</summary>

**Use host networking when:**

- The container needs maximum network performance (no veth/NAT overhead)
- The container needs to bind to many ports dynamically (e.g., a range scanner, Prometheus node exporter)
- The container needs to monitor host network interfaces (e.g., `tcpdump`, `iptables`)

**Risks:**

- No network isolation — a compromised container can bind to any host port
- Port conflicts — container and host share the same port namespace
- Not portable — Linux-only (macOS Docker Desktop uses a VM; host networking goes to the VM, not the Mac)
- `-p` flags are silently ignored, which confuses users

```bash
# ✓ Valid use: Prometheus node exporter needs to read /proc/net/*
docker run --network host prom/node-exporter

# ❌ Bad use: web app that "doesn't need port mapping overhead"
# The NAT overhead is typically microseconds — not worth the isolation loss
```

**Alternatives before reaching for `--network host`**:

- macvlan for physical LAN presence
- Custom bridge for low-overhead inter-container communication (no NAT between containers on the same network)

</details>

---

## Key Takeaways

1. **Network namespaces** give each container an isolated network stack — containers can't see each other's traffic by default
2. **Custom bridge networks** enable DNS resolution by container name via `127.0.0.11`; the default bridge does not
3. **`EXPOSE` is documentation** — it does not publish ports to the host; `-p` does
4. **`-p` bypasses host firewalls** via iptables — binding to `127.0.0.1:port:port` restricts to localhost
5. **Overlay networks** use VXLAN to extend Layer 2 across hosts — used in Swarm, not in Kubernetes (Kubernetes uses CNI)
6. **Network segmentation with Compose** — put services on separate networks to enforce least-privilege access
7. **`host.docker.internal`** resolves to the host from inside a container (built-in on Docker Desktop, add via `--add-host` on Linux)
8. **Multiple networks per container** is the correct way to bridge traffic between network segments

## Next Steps

In [Lesson 03: Volumes & Storage Strategies](lesson-03-volumes-and-storage-strategies.md), you'll learn:

- How Docker's union filesystem interacts with volumes
- Named volumes vs bind mounts vs tmpfs — when to use each
- Volume drivers for NFS, cloud storage, and distributed filesystems
- Backup, restore, and migration patterns for stateful containers
