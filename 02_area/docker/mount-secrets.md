The --secret flag is a mapping instruction it got two parts. One is the instruction at build.

```sh
  --secret id=GITHUB_TOKEN,env=GITHUB_TOKEN
```

- env=GITHUB_TOKEN — read the value from this environment variable on the host
- id=GITHUB_TOKEN — expose it inside the build under this name

The id in the Dockerfile is the lookup key

```Dockerfile
  RUN --mount=type=secret,id=GITHUB_TOKEN,required=true \
    export GIT_TOKEN=$(cat /run/secrets/GITHUB_TOKEN)

```

When Docker sees id=GITHUB_TOKEN on the mount, it looks up the secret that was registered under that name by --secret
id=GITHUB_TOKEN,... and mounts it as a file at /run/secrets/GITHUB_TOKEN. The cat reads that file.

```sh
host env var          --secret flag         Dockerfile mount       file in build
GITHUB_TOKEN    →    id=GITHUB_TOKEN    →   id=GITHUB_TOKEN    →  /run/secrets/GITHUB_TOKEN
(the value)          (registers it)         (mounts it)           (cat reads it)
```

So the full chain is:

host env var --secret flag Dockerfile mount file in build

GITHUB_TOKEN → id=GITHUB_TOKEN → id=GITHUB_TOKEN → /run/secrets/GITHUB_TOKEN

(the value) (registers it) (mounts it) (cat reads it)

Why a file and not an env var?

BuildKit deliberately doesn't inject secrets as environment variables. Env vars can leak into child processes, logs, and docker inspect
output. The tmpfs file at /run/secrets/ only exists for the duration of that single RUN step and is never written into the image layer,
which is the whole point of the secret mount.
