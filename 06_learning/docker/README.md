# Docker Advanced Interview Prep

A focused, senior-level refresher for developers with 2+ years of Docker experience preparing for technical interviews.

## Lesson Series

### Image Internals & Building
- [Lesson 01: Image Internals & Optimization](lesson-01-image-internals-and-optimization.md) - OverlayFS, layer caching, multi-stage builds, distroless, BuildKit
- [Lesson 07: BuildKit & Multi-platform Builds](lesson-07-buildkit-and-multi-platform.md) - RUN --mount, build secrets, buildx, QEMU, multi-platform, cache backends

### Networking
- [Lesson 02: Networking Deep Dive](lesson-02-networking-deep-dive.md) - Network namespaces, bridge/overlay/host/macvlan, DNS, inter-container communication

### Storage
- [Lesson 03: Volumes & Storage Strategies](lesson-03-volumes-and-storage-strategies.md) - Named volumes vs bind mounts vs tmpfs, volume drivers, backup/restore

### Compose & Orchestration
- [Lesson 04: Compose Advanced Patterns](lesson-04-compose-advanced-patterns.md) - Profiles, healthcheck depends_on, secrets/configs, override files, env var precedence

### Security & Production
- [Lesson 05: Security Hardening](lesson-05-security-hardening.md) - Non-root user, read-only FS, seccomp, capabilities, secrets management, rootless Docker
- [Lesson 06: Production Patterns](lesson-06-production-patterns.md) - PID 1/signal handling, health checks, restart policies, resource limits, logging drivers
- [Lesson 09: Exec Form, Shell Form, and PID 1](lesson-09-exec-form-shell-form-and-pid1.md) - execve() internals, shell parsing pipeline, signal delivery, SIGKILL timeout, exec entrypoint pattern
- [Lesson 10: ARG and ENV](lesson-10-arg-and-env.md) - Build-time vs runtime variables, scope rules, multi-stage ARG inheritance, cache invalidation, secret leaks

### Debugging
- [Lesson 08: Debugging & Troubleshooting](lesson-08-debugging-and-troubleshooting.md) - docker inspect, nsenter, layer analysis, stats, common failure patterns

## Target Audience

Developers with 2+ years of Docker experience who need a focused refresher covering:

- What you might have forgotten or never dug into deeply
- Interview-critical internals (how Docker actually works)
- Production edge cases and gotchas
- Security and hardening patterns expected at senior level

## Usage

Each lesson includes:

- Quick reference tables for fast lookup
- Code examples annotated with ✓ (correct) and ❌ (wrong)
- Interview questions with collapsible answers
- Hands-on exercises with collapsible solutions
- Key takeaways for last-minute review
