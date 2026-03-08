# TypeScript File Operations

## Why

- **Always use node:fs/promises** — The callback-based `fs` API is legacy. The promise-based API is cleaner, works with async/await, and avoids callback hell. Import from `"node:fs/promises"`.
- **Use node: prefix** — `import from "node:fs/promises"` makes it explicit that this is a built-in module, not an npm package. Required in some runtimes, good practice everywhere.
- **utf-8 encoding** — readFile returns a Buffer by default. Pass `"utf-8"` to get a string. Forgetting this is a common source of bugs.
- **Recursive mkdir** — `mkdir` fails if parent dirs don't exist. Always pass `{ recursive: true }` unless you specifically want that check.
- **Streams for large files** — readFile loads the entire file into memory. For large files (logs, CSVs, uploads), use createReadStream to process chunks.

## Quick Reference

| Use case         | Method                             |
| ---------------- | ---------------------------------- |
| Read file        | `readFile(path, "utf-8")`          |
| Write file       | `writeFile(path, data)`            |
| Append to file   | `appendFile(path, data)`           |
| Check exists     | `access(path)` or `stat(path)`     |
| Delete file      | `unlink(path)` or `rm(path)`       |
| Create directory | `mkdir(path, { recursive: true })` |
| List directory   | `readdir(path)`                    |
| File info        | `stat(path)`                       |
| Copy file        | `copyFile(src, dest)`              |
| Rename / move    | `rename(src, dest)`                |

## Reading Files

### 1. Read entire file as string

```typescript
import { readFile } from "node:fs/promises";

const content = await readFile("config.json", "utf-8");
```

### 2. Read as Buffer (binary data)

```typescript
const buffer = await readFile("image.png");
console.log(buffer.length); // bytes
```

### 3. Read JSON file

```typescript
import { readFile } from "node:fs/promises";

async function readJson<T>(path: string): Promise<T> {
  const raw = await readFile(path, "utf-8");
  return JSON.parse(raw) as T;
}

const config = await readJson<Config>("./config.json");
```

## Writing Files

### 4. Write string to file

```typescript
import { writeFile } from "node:fs/promises";

await writeFile("output.txt", "hello world\n");
```

### 5. Write JSON

```typescript
await writeFile("data.json", JSON.stringify(data, null, 2) + "\n");
```

### 6. Append to file

```typescript
import { appendFile } from "node:fs/promises";

await appendFile("log.txt", `${new Date().toISOString()} - event occurred\n`);
```

## Directories

### 7. Create directory (recursive)

```typescript
import { mkdir } from "node:fs/promises";

await mkdir("path/to/nested/dir", { recursive: true }); // no error if exists
```

### 8. List directory contents

```typescript
import { readdir } from "node:fs/promises";

const files = await readdir("./src");
// ["index.ts", "utils.ts", "types.ts"]

// With file type info
const entries = await readdir("./src", { withFileTypes: true });
for (const entry of entries) {
  if (entry.isDirectory()) {
    console.log(`dir: ${entry.name}`);
  } else {
    console.log(`file: ${entry.name}`);
  }
}
```

## File Info & Checks

### 9. Check if file exists

```typescript
import { access, constants } from "node:fs/promises";

async function fileExists(path: string): Promise<boolean> {
  try {
    await access(path, constants.F_OK);
    return true;
  } catch {
    return false;
  }
}
```

### 10. Get file stats

```typescript
import { stat } from "node:fs/promises";

const stats = await stat("file.txt");
stats.size; // bytes
stats.isFile(); // true
stats.isDirectory(); // false
stats.mtime; // last modified (Date)
stats.birthtime; // created (Date)
```

## File Operations

### 11. Copy, rename, delete

```typescript
import { copyFile, rename, rm, unlink } from "node:fs/promises";

await copyFile("src.txt", "dest.txt");
await rename("old.txt", "new.txt"); // also works as move
await unlink("file.txt"); // delete single file
await rm("dir", { recursive: true }); // delete directory and contents
```

## Streams (large files)

### 12. Read large file line by line

```typescript
import { createReadStream } from "node:fs";
import { createInterface } from "node:readline";

const rl = createInterface({
  input: createReadStream("large-file.log"),
});

for await (const line of rl) {
  process.stdout.write(`${line}\n`);
}
```

### 13. Pipe stream to file

```typescript
import { createWriteStream } from "node:fs";
import { pipeline } from "node:stream/promises";

const response = await fetch("https://example.com/large-file.zip");
if (!response.ok || !response.body) throw new Error("Download failed");

await pipeline(response.body, createWriteStream("./download.zip"));
```

## Paths

### 14. Path utilities

```typescript
import { join, resolve, basename, dirname, extname } from "node:path";

join("src", "utils", "index.ts"); // "src/utils/index.ts"
resolve("./src", "index.ts"); // "/absolute/path/src/index.ts"
basename("/path/to/file.txt"); // "file.txt"
basename("/path/to/file.txt", ".txt"); // "file"
dirname("/path/to/file.txt"); // "/path/to"
extname("file.txt"); // ".txt"
```
