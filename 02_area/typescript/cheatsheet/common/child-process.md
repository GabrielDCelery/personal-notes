# TypeScript Child Process

## Why

- **exec vs spawn** — `exec` buffers the entire output into memory and returns it as a string. `spawn` streams stdout/stderr. Use exec for small outputs (git commands, scripts), spawn for large or long-running processes.
- **execFile over exec for safety** — `exec` runs through a shell, which means shell injection is possible if you interpolate user input. `execFile` runs the binary directly with no shell — much safer.
- **Always handle errors and exit codes** — A child process can fail silently. Check `code` (exit code) and `signal` (if killed). Exit code 0 means success, anything else is failure.
- **Use signal to kill** — Pass an `AbortSignal` to kill a child process on timeout or cancellation. Cleaner than manually calling `.kill()`.
- **stdio: "inherit" for interactive** — Connects the child's stdin/stdout/stderr directly to the parent. Use for interactive commands or when you want output to flow through naturally.

## Quick Reference

| Use case                  | Method                              |
| ------------------------- | ----------------------------------- |
| Run and get output        | `execFile` (promise)                |
| Run through shell         | `exec` (promise)                    |
| Stream output             | `spawn`                             |
| Run and wait              | `spawnSync` / `execFileSync`        |
| Shell pipes               | `exec("cmd1 \| cmd2")`             |
| Kill on timeout           | `AbortSignal.timeout(ms)`          |
| Inherit parent stdio      | `spawn(cmd, { stdio: "inherit" })` |

## exec / execFile (buffered output)

### 1. execFile — safe, no shell

```typescript
import { execFile } from "node:child_process";
import { promisify } from "node:util";

const execFileAsync = promisify(execFile);

const { stdout, stderr } = await execFileAsync("git", ["status", "--short"]);
console.log(stdout);
```

### 2. exec — runs through shell (use for pipes, globbing)

```typescript
import { exec } from "node:child_process";
import { promisify } from "node:util";

const execAsync = promisify(exec);

const { stdout } = await execAsync("ls -la | grep .ts");
console.log(stdout);
```

### 3. With options

```typescript
const { stdout } = await execFileAsync("node", ["script.js"], {
  cwd: "/path/to/project",
  env: { ...process.env, NODE_ENV: "test" },
  timeout: 10_000, // kill after 10 seconds
  maxBuffer: 10 * 1024 * 1024, // 10MB max output
});
```

### 4. Handle errors

```typescript
try {
  const { stdout } = await execFileAsync("git", ["diff", "--stat"]);
  console.log(stdout);
} catch (err) {
  // err has stdout, stderr, code, signal
  const { stderr, code } = err as { stderr: string; code: number };
  console.error(`Exit code ${code}: ${stderr}`);
}
```

## spawn (streaming)

### 5. Basic spawn with streaming output

```typescript
import { spawn } from "node:child_process";

const child = spawn("npm", ["test"], { cwd: "./project" });

child.stdout.on("data", (data: Buffer) => {
  process.stdout.write(data);
});

child.stderr.on("data", (data: Buffer) => {
  process.stderr.write(data);
});

child.on("close", (code) => {
  console.log(`Exited with code ${code}`);
});
```

### 6. Inherit stdio (pass-through to parent)

```typescript
import { spawn } from "node:child_process";

const child = spawn("npm", ["run", "build"], {
  stdio: "inherit", // child's stdout/stderr go directly to terminal
  cwd: "./project",
});

child.on("close", (code) => {
  if (code !== 0) process.exit(code!);
});
```

### 7. Spawn and collect output as promise

```typescript
import { spawn } from "node:child_process";

function run(cmd: string, args: string[]): Promise<string> {
  return new Promise((resolve, reject) => {
    const child = spawn(cmd, args);
    const chunks: Buffer[] = [];

    child.stdout.on("data", (data) => chunks.push(data));
    child.on("error", reject);
    child.on("close", (code) => {
      if (code !== 0) reject(new Error(`${cmd} exited with code ${code}`));
      else resolve(Buffer.concat(chunks).toString("utf-8"));
    });
  });
}

const output = await run("git", ["log", "--oneline", "-5"]);
```

## Kill & Timeout

### 8. Kill with AbortSignal

```typescript
import { execFile } from "node:child_process";
import { promisify } from "node:util";

const execFileAsync = promisify(execFile);

try {
  const { stdout } = await execFileAsync("long-running-task", [], {
    signal: AbortSignal.timeout(5000), // kill after 5s
  });
} catch (err) {
  if ((err as NodeJS.ErrnoException).code === "ABORT_ERR") {
    console.error("Process timed out");
  }
}
```

### 9. Manual kill

```typescript
const child = spawn("long-running-task");

setTimeout(() => {
  child.kill("SIGTERM"); // graceful
  setTimeout(() => child.kill("SIGKILL"), 5000); // force after 5s
}, 30_000);
```

## Patterns

### 10. Run shell command helper

```typescript
import { execFile } from "node:child_process";
import { promisify } from "node:util";

const execFileAsync = promisify(execFile);

async function sh(cmd: string, args: string[], options?: { cwd?: string }): Promise<string> {
  const { stdout } = await execFileAsync(cmd, args, {
    cwd: options?.cwd,
    timeout: 30_000,
    maxBuffer: 10 * 1024 * 1024,
  });
  return stdout.trim();
}

const branch = await sh("git", ["rev-parse", "--abbrev-ref", "HEAD"]);
const hash = await sh("git", ["rev-parse", "--short", "HEAD"]);
```

### 11. Run npm/npx scripts

```typescript
import { spawn } from "node:child_process";

function npmRun(script: string, cwd: string): Promise<number> {
  return new Promise((resolve, reject) => {
    const child = spawn("npm", ["run", script], {
      cwd,
      stdio: "inherit",
      shell: process.platform === "win32", // needed on Windows
    });

    child.on("error", reject);
    child.on("close", (code) => resolve(code ?? 1));
  });
}

const exitCode = await npmRun("test", "./project");
if (exitCode !== 0) console.error("Tests failed");
```

### 12. Parallel processes

```typescript
async function runAll(commands: { cmd: string; args: string[] }[]): Promise<void> {
  await Promise.all(
    commands.map(
      ({ cmd, args }) =>
        new Promise<void>((resolve, reject) => {
          const child = spawn(cmd, args, { stdio: "inherit" });
          child.on("error", reject);
          child.on("close", (code) =>
            code === 0 ? resolve() : reject(new Error(`${cmd} exited with ${code}`)),
          );
        }),
    ),
  );
}

await runAll([
  { cmd: "npm", args: ["run", "lint"] },
  { cmd: "npm", args: ["run", "typecheck"] },
  { cmd: "npm", args: ["run", "test"] },
]);
```
