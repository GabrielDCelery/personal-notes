# How Container Log Forwarding Works

## Starting at the bottom: stdout and stderr

When your process writes to stdout (fd 1) or stderr (fd 2), it's writing to a **pipe** that Docker controls. Your process has no knowledge of what happens on the other end of that pipe — it writes and forgets.

Docker has a goroutine per container called the **log copier** that sits in a read loop on those pipes. It reads bytes, frames them into log messages by splitting on newlines, attaches metadata (timestamp, stream name, container ID), and hands each message to the configured **log driver**.

```
container process writes to stdout
    └──> pipe
              └──> Docker log copier goroutine (inside dockerd)
                        └──> log driver (e.g. awslogs, json-file)
```

One thing worth knowing: Docker splits on `\n`, so each line becomes a separate log message. If a line exceeds 16KB it gets flushed with a `Partial: true` flag. This is why multiline logs (stack traces) get fragmented — they arrive as many separate messages rather than one coherent block.

## Processes, daemons, and drivers

Before going further it's worth clarifying three terms that are easy to conflate.

A **process** is an OS-level concept — a running program with its own PID, memory space, file descriptors, and lifecycle managed by the kernel.

A **daemon** is just a process that runs in the background, isn't attached to a terminal, and typically starts at boot and runs continuously. The `d` suffix in `dockerd`, `sshd`, `nginx` is a Unix naming convention signaling exactly this. There's no separate software abstraction called a daemon — it's just a word for that kind of process.

A **driver** is a software pattern concept — code that implements a specific interface. It has no PID, no independent memory space, no lifecycle of its own. It lives inside whatever process loaded it. A useful test: does it have a PID? If yes, it's a process. If no, it's just code running inside something else.

So `dockerd` is a daemon (a background process) that contains log drivers (code). The drivers are not processes — they're structs in Go with methods, living in dockerd's memory, executed by dockerd's goroutines.

## The log driver

The log driver is not a separate process. It's code that runs **inside the Docker daemon (dockerd)** as part of the same goroutine machinery. The interface is simple:

```go
type Logger interface {
    Log(*Message) error
    Close() error
}
```

Each driver implements `Log()` differently:

- `json-file` — writes to disk on the host
- `awslogs` — buffers messages and calls the CloudWatch `PutLogEvents` API
- `awsfirelens` — writes messages to a Unix socket inside the task

The driver may spin up its own background goroutines too. `awslogs` for example has a separate goroutine that flushes buffered messages to CloudWatch on a timer, independent of the goroutine reading the pipe.

All of it lives inside dockerd — same process, same memory space, no separate PID.

## json-file log limits

By default `json-file` has no size limit — it will grow until it fills the disk, which is a real problem in production. You configure limits either globally in `/etc/docker/daemon.json` or per container:

```yaml
# docker-compose.yml
services:
  my-service:
    image: my-image
    logging:
      driver: json-file
      options:
        max-size: "10m"
        max-file: "3"
```

```sh
# docker run
docker run --log-driver json-file --log-opt max-size=10m --log-opt max-file=3 my-image
```

- `max-size` — max size of a single log file before it rotates
- `max-file` — how many rotated files to keep

With the above you'd have at most 30MB of logs per container (3 files × 10MB). When a new file is created the oldest is deleted. In production it's good practice to either set these limits or switch to a driver like `awslogs` that ships logs off the host entirely.

## How awslogs ships to CloudWatch

The CloudWatch call is a `PutLogEvents` — an HTTPS POST with a batch of log events (timestamp + message string) to a specific log group and stream. Auth uses AWS Signature V4 derived from the task's IAM role, which ECS makes available via a credentials endpoint. Docker's `awslogs` driver handles all of this: buffering, batching, signing, and retrying. Your app knows nothing about it.

## Why Datadog needs something different

`awslogs` works out of the box because CloudWatch is an AWS service and IAM just handles auth. For Datadog — a third-party SaaS — there's no built-in driver in dockerd, and there's no IAM equivalent. Someone has to make an HTTPS POST to `http-intake.logs.datadoghq.com` with your API key, and that's not something dockerd knows how to do.

On EC2 you could install a Datadog agent on the host, which runs alongside dockerd and can read container logs from the Docker socket. On **Fargate** that's not possible — AWS manages the host and dockerd, you can't touch them.

## FireLens: the Fargate solution

AWS added `awsfirelens` as a built-in driver in the version of dockerd they run on Fargate. Instead of shipping logs directly to a remote service, it hands them off to a sidecar container inside your task — which you control — and that sidecar makes the API call to wherever you want.

The handoff mechanism between dockerd and the sidecar is a **Unix socket**. A Unix socket is a file on disk that two processes on the same machine use to send data to each other. It works like a network socket — one side listens, one side connects and writes — except there's no IP, no TCP, no network stack. The kernel moves bytes directly between the two processes in memory. The file itself contains no data, it's just an address (like an IP:port but as a filesystem path).

```
container stdout
    └──> awsfirelens driver (inside dockerd, AWS-managed)
              └──> Unix socket  ← bytes cross the process boundary here
                        └──> Fluent Bit sidecar (inside your task, you control this)
                                  └──> POST to Datadog / CloudWatch / wherever
```

This is also why drivers like `awslogs` and `json-file` don't need a socket — the log copier goroutine calls the driver code directly inside dockerd. No process boundary, no socket needed. The socket only appears when logs need to leave dockerd entirely.

Access to a Unix socket is controlled by standard Linux file permissions — `chmod` and `chown` work on them like any other file. The Docker socket (`/var/run/docker.sock`) is a well-known example: it's owned by `root` and group `docker`, which is why you need to be in the `docker` group to run Docker commands without `sudo`. Anyone who can write to that socket can control dockerd — effectively root access on the machine.

ECS handles the FireLens wiring automatically — it creates the socket, configures the path, and injects the Fluent Bit config so both sides connect without manual setup.

## What Fluent Bit is

Fluent Bit is a lightweight log processor written in C (~5-10 MB memory footprint). It has a simple pipeline model:

```
INPUT → PARSER → FILTER → OUTPUT
```

- **INPUT**: where logs come from — in this case, the Unix socket
- **PARSER**: optionally parse raw text into structured fields (e.g. parse JSON log lines into key/value pairs)
- **FILTER**: transform or enrich — add fields, redact values, drop records
- **OUTPUT**: a plugin per destination — each one knows that service's API and handles batching, retries, and auth

The Datadog output plugin calls `POST /api/v2/logs` on Datadog's HTTP intake with your API key in the header and a JSON array of log events in the body. Fluent Bit batches records by count or time window, fires the request, and retries with backoff on failure.

## The universal pattern

This read-and-forward structure appears everywhere in logging, not just Docker:

| System            | Reader                                        | Forwarder                                              |
| ----------------- | --------------------------------------------- | ------------------------------------------------------ |
| Docker            | log copier goroutine in dockerd               | log driver (awslogs, firelens)                         |
| Kubernetes        | kubelet writes logs to `/var/log/containers/` | Fluent Bit / Datadog agent DaemonSet tails the files   |
| Traditional Linux | journald / syslog collects from processes     | rsyslog / filebeat reads and forwards                  |
| AWS Lambda        | Lambda runtime captures stdout                | Ships to CloudWatch automatically (AWS owns the stack) |

The reason it's always decoupled like this is that **log generation and log shipping have different failure modes**. Your app shouldn't block or crash because the log destination is slow or unreachable. The copier acts as a buffer — your process writes and forgets, and the shipper handles batching, retries, and backpressure independently.

## The sidecar is just the execution environment

On Fargate, you can't run anything on the host. The sidecar is the only place you're allowed to run your own code. FireLens isn't a fundamentally different approach to log routing — it's the same read-and-forward pattern, just with the forwarder running inside the task because that's the only option available to you.

```
AWS controls                    │  You control
                                │
dockerd on managed Fargate host │  Fluent Bit sidecar inside your task
  awsfirelens driver            │    └──> Datadog output plugin
  forwards to Unix socket  ─────┼──>       └──> POST https://http-intake.logs.datadoghq.com
                                │               DD-API-KEY: your key
```
