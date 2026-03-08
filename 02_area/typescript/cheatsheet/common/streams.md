# TypeScript Streams

## Why

- **Streams process data in chunks** — Instead of loading an entire file or response into memory, streams handle data piece by piece. Essential for large files, HTTP uploads/downloads, and ETL pipelines.
- **Backpressure is built in** — When a consumer is slower than a producer, Node.js streams automatically pause the producer. Without streams (e.g. manual event handling), you risk unbounded memory growth.
- **pipeline over pipe** — `stream.pipe()` doesn't forward errors or clean up on failure. `pipeline()` from `node:stream/promises` handles errors, destroys streams on failure, and returns a promise.
- **Four stream types** — Readable (source), Writable (sink), Transform (process through), Duplex (both). Most backend work uses Readable and Transform.
- **Async iteration** — Readable streams implement `Symbol.asyncIterator`. You can `for await...of` them directly, which is the simplest way to consume a stream.

## Quick Reference

| Use case              | Method                              |
| --------------------- | ----------------------------------- |
| Read file as stream   | `createReadStream(path)`            |
| Write file as stream  | `createWriteStream(path)`           |
| Pipe with error handling | `pipeline(src, ...transforms, dest)` |
| Transform chunks      | `new Transform({ transform() })` |
| Async iteration       | `for await (const chunk of stream)` |
| Collect into string   | `stream.toArray()` / manual concat  |
| From array/iterable   | `Readable.from(iterable)`          |

## Reading

### 1. Read file as stream

```typescript
import { createReadStream } from "node:fs";

const stream = createReadStream("large-file.log", { encoding: "utf-8" });

for await (const chunk of stream) {
  process.stdout.write(chunk);
}
```

### 2. Read line by line

```typescript
import { createReadStream } from "node:fs";
import { createInterface } from "node:readline";

const rl = createInterface({
  input: createReadStream("data.csv"),
});

for await (const line of rl) {
  const [id, name] = line.split(",");
  console.log({ id, name });
}
```

### 3. Collect stream into string

```typescript
import { Readable } from "node:stream";

async function streamToString(stream: Readable): Promise<string> {
  const chunks: Buffer[] = [];
  for await (const chunk of stream) {
    chunks.push(Buffer.isBuffer(chunk) ? chunk : Buffer.from(chunk));
  }
  return Buffer.concat(chunks).toString("utf-8");
}
```

## Writing

### 4. Write to file stream

```typescript
import { createWriteStream } from "node:fs";

const ws = createWriteStream("output.log");

ws.write("line 1\n");
ws.write("line 2\n");
ws.end(); // signal no more data

// Wait for finish
ws.on("finish", () => console.log("done"));
```

### 5. Create readable from data

```typescript
import { Readable } from "node:stream";

const stream = Readable.from(["chunk1\n", "chunk2\n", "chunk3\n"]);

// From async generator
async function* generate() {
  for (let i = 0; i < 100; i++) {
    yield `line ${i}\n`;
  }
}
const stream = Readable.from(generate());
```

## Pipeline

### 6. pipeline — safe piping with error handling

```typescript
import { createReadStream, createWriteStream } from "node:fs";
import { pipeline } from "node:stream/promises";
import { createGzip, createGunzip } from "node:zlib";

// Compress a file
await pipeline(
  createReadStream("input.txt"),
  createGzip(),
  createWriteStream("input.txt.gz"),
);

// Decompress
await pipeline(
  createReadStream("input.txt.gz"),
  createGunzip(),
  createWriteStream("output.txt"),
);
```

### 7. Download file with pipeline

```typescript
import { createWriteStream } from "node:fs";
import { pipeline } from "node:stream/promises";
import { Readable } from "node:stream";

const response = await fetch("https://example.com/large-file.zip");
if (!response.ok || !response.body) throw new Error("Download failed");

await pipeline(
  Readable.fromWeb(response.body),
  createWriteStream("download.zip"),
);
```

## Transform Streams

### 8. Custom transform

```typescript
import { Transform } from "node:stream";

const upperCase = new Transform({
  transform(chunk, encoding, callback) {
    callback(null, chunk.toString().toUpperCase());
  },
});

await pipeline(
  createReadStream("input.txt"),
  upperCase,
  createWriteStream("output.txt"),
);
```

### 9. JSON line parser transform

```typescript
import { Transform } from "node:stream";

const jsonLineParser = new Transform({
  objectMode: true,
  transform(chunk, encoding, callback) {
    const lines = chunk.toString().split("\n").filter(Boolean);
    for (const line of lines) {
      try {
        this.push(JSON.parse(line));
      } catch {
        // skip invalid lines
      }
    }
    callback();
  },
});
```

### 10. Process NDJSON file

```typescript
import { createReadStream } from "node:fs";
import { createInterface } from "node:readline";

const rl = createInterface({
  input: createReadStream("data.ndjson"),
});

for await (const line of rl) {
  if (!line.trim()) continue;
  const record = JSON.parse(line);
  await processRecord(record);
}
```

## Patterns

### 11. Stream to HTTP response (Express)

```typescript
app.get("/download", (req, res) => {
  res.setHeader("Content-Type", "application/octet-stream");
  res.setHeader("Content-Disposition", "attachment; filename=data.csv");

  const stream = createReadStream("data.csv");
  stream.pipe(res);
  stream.on("error", (err) => {
    console.error(err);
    if (!res.headersSent) res.status(500).end();
  });
});
```

### 12. Process large CSV without loading into memory

```typescript
import { createReadStream } from "node:fs";
import { createInterface } from "node:readline";

async function processLargeCsv(path: string) {
  const rl = createInterface({ input: createReadStream(path) });
  let isHeader = true;
  let headers: string[];

  for await (const line of rl) {
    if (isHeader) {
      headers = line.split(",");
      isHeader = false;
      continue;
    }

    const values = line.split(",");
    const row = Object.fromEntries(headers.map((h, i) => [h, values[i]]));
    await processRow(row);
  }
}
```
