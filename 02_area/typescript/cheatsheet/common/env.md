# TypeScript Environment Variables

## Why

- **process.env values are always string | undefined** — There's no typed access. You must parse numbers, booleans, and enums yourself. Do it once at startup, not scattered across the codebase.
- **Config object pattern** — Centralizes all env var reads into one place. Fails fast at startup if required vars are missing. The rest of the app imports a typed config object, never touches process.env directly.
- **No .env in production** — dotenv files are for local dev. In production, env vars come from the platform (ECS task def, K8s configmap, systemd unit). Don't ship .env files in your container.
- **Validate early** — If DATABASE_URL is missing, you want to know at startup, not when the first request hits the database 5 minutes later.
- **NODE_ENV is not secure** — Anyone can set it. Don't use it for security decisions. Use it for behavior switches (logging verbosity, error detail).

## Quick Reference

| Use case          | Method                             |
| ----------------- | ---------------------------------- |
| Read a var        | `process.env.KEY`                  |
| Read with default | `process.env.KEY ?? "default"`     |
| Check if set      | `process.env.KEY !== undefined`    |
| Load .env file    | `dotenv` / `--env-file` (Node 20+) |
| Typed config      | validate at startup                |

## Basics

### 1. Read environment variables

```typescript
const port = process.env.PORT; // string | undefined
const host = process.env.HOST ?? "0.0.0.0";
```

### 2. Check if set vs empty

```typescript
if (process.env.DATABASE_URL === undefined) {
  // not set at all
} else if (process.env.DATABASE_URL === "") {
  // set but empty
}
```

### 3. Parse non-string types

```typescript
const port = parseInt(process.env.PORT ?? "3000", 10);
const debug = process.env.DEBUG === "true";
const maxRetries = Number(process.env.MAX_RETRIES ?? "3");
const timeout = parseFloat(process.env.TIMEOUT_SECONDS ?? "30");
```

## Loading .env Files

### 4. Node 20+ built-in --env-file

```sh
node --env-file=.env dist/server.js
node --env-file=.env --env-file=.env.local dist/server.js  # later file wins
```

No code changes needed — vars are available on process.env.

### 5. dotenv package (older Node versions)

```typescript
import "dotenv/config"; // load at import time — put at top of entrypoint

// Or load manually
import { config } from "dotenv";
config({ path: ".env.local" });
```

## Config Pattern

### 6. Typed config with validation

```typescript
interface Config {
  port: number;
  databaseUrl: string;
  redisUrl: string;
  jwtSecret: string;
  debug: boolean;
  nodeEnv: "development" | "production" | "test";
}

function loadConfig(): Config {
  const required = (key: string): string => {
    const value = process.env[key];
    if (value === undefined || value === "") {
      throw new Error(`Missing required env var: ${key}`);
    }
    return value;
  };

  return {
    port: parseInt(process.env.PORT ?? "3000", 10),
    databaseUrl: required("DATABASE_URL"),
    redisUrl: required("REDIS_URL"),
    jwtSecret: required("JWT_SECRET"),
    debug: process.env.DEBUG === "true",
    nodeEnv: (process.env.NODE_ENV as Config["nodeEnv"]) ?? "development",
  };
}

export const config = loadConfig(); // fails fast at import time
```

### 7. Config with zod (if you use zod in the project)

```typescript
import { z } from "zod";

const envSchema = z.object({
  PORT: z.coerce.number().default(3000),
  DATABASE_URL: z.string().url(),
  REDIS_URL: z.string().url(),
  JWT_SECRET: z.string().min(32),
  DEBUG: z
    .enum(["true", "false"])
    .transform((v) => v === "true")
    .default("false"),
  NODE_ENV: z
    .enum(["development", "production", "test"])
    .default("development"),
});

export const config = envSchema.parse(process.env);
```

## Patterns

### 8. .env.example as documentation

```sh
# .env.example — commit this file, never commit .env
PORT=3000
DATABASE_URL=postgres://user:pass@localhost:5432/mydb
REDIS_URL=redis://localhost:6379
JWT_SECRET=your-secret-here-min-32-chars
DEBUG=false
NODE_ENV=development
```

### 9. Per-environment overrides

```
.env                # shared defaults (committed)
.env.local          # local overrides (gitignored)
.env.test           # test environment (committed)
.env.production     # production reference (committed, no secrets)
```

```typescript
import { config } from "dotenv";

const env = process.env.NODE_ENV ?? "development";
config({ path: `.env.${env}` });
config({ path: ".env" }); // fallback — dotenv doesn't overwrite existing vars
```
