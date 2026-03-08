# BullMQ

```sh
npm install bullmq
```

> Redis-backed job queue for Node.js. Supports delayed jobs, retries, priorities, rate limiting, and repeatable jobs.

## Setup

```typescript
import { Queue, Worker } from "bullmq";

const connection = { host: "localhost", port: 6379 };

const queue = new Queue("emails", { connection });
const worker = new Worker("emails", processJob, { connection });
```

## Add Jobs

```typescript
// Simple job
await queue.add("send-welcome", { userId: "123", email: "alice@example.com" });

// With options
await queue.add(
  "send-welcome",
  { userId: "123" },
  {
    delay: 5000, // delay 5 seconds
    attempts: 3, // retry up to 3 times
    backoff: { type: "exponential", delay: 1000 },
    priority: 1, // lower = higher priority
    removeOnComplete: true,
    removeOnFail: 1000, // keep last 1000 failed jobs
  },
);

// Bulk add
await queue.addBulk([
  { name: "send-welcome", data: { userId: "1" } },
  { name: "send-welcome", data: { userId: "2" } },
  { name: "send-welcome", data: { userId: "3" } },
]);
```

## Process Jobs (Worker)

```typescript
import { Worker, Job } from "bullmq";

const worker = new Worker(
  "emails",
  async (job: Job) => {
    switch (job.name) {
      case "send-welcome":
        await sendWelcomeEmail(job.data.userId, job.data.email);
        break;
      case "send-receipt":
        await sendReceipt(job.data.orderId);
        break;
    }

    // Return value is stored as job result
    return { sent: true };
  },
  {
    connection,
    concurrency: 5, // process 5 jobs at a time
  },
);

worker.on("completed", (job) => {
  console.log(`Job ${job.id} completed`);
});

worker.on("failed", (job, err) => {
  console.error(`Job ${job?.id} failed:`, err.message);
});
```

## Job Progress

```typescript
// In worker
const worker = new Worker("import", async (job) => {
  const items = job.data.items;
  for (let i = 0; i < items.length; i++) {
    await processItem(items[i]);
    await job.updateProgress(Math.round(((i + 1) / items.length) * 100));
  }
});

// Listen for progress
worker.on("progress", (job, progress) => {
  console.log(`Job ${job.id}: ${progress}%`);
});
```

## Delayed Jobs

```typescript
// Process in 10 minutes
await queue.add(
  "reminder",
  { userId: "123" },
  {
    delay: 10 * 60 * 1000,
  },
);
```

## Repeatable Jobs (cron)

```typescript
await queue.upsertJobScheduler(
  "daily-report",
  { pattern: "0 9 * * *" }, // every day at 9:00 AM
  { name: "generate-report", data: { type: "daily" } },
);

await queue.upsertJobScheduler(
  "cleanup",
  { every: 60000 }, // every 60 seconds
  { name: "cleanup-expired", data: {} },
);

// Remove a repeatable job
await queue.removeJobScheduler("daily-report");
```

## Rate Limiting

```typescript
const worker = new Worker("api-calls", processApiCall, {
  connection,
  limiter: {
    max: 10, // max 10 jobs
    duration: 1000, // per 1 second
  },
});
```

## Events

```typescript
import { QueueEvents } from "bullmq";

const queueEvents = new QueueEvents("emails", { connection });

queueEvents.on("completed", ({ jobId, returnvalue }) => {
  console.log(`Job ${jobId} completed with`, returnvalue);
});

queueEvents.on("failed", ({ jobId, failedReason }) => {
  console.error(`Job ${jobId} failed:`, failedReason);
});

queueEvents.on("delayed", ({ jobId, delay }) => {
  console.log(`Job ${jobId} delayed by ${delay}ms`);
});
```

## Graceful Shutdown

```typescript
process.on("SIGTERM", async () => {
  await worker.close(); // wait for current jobs to finish
  await queue.close();
  process.exit(0);
});
```

## Job Inspection

```typescript
const job = await queue.getJob(jobId);
console.log(job?.data);
console.log(job?.progress);
console.log(await job?.getState()); // "completed" | "failed" | "delayed" | "active" | "waiting"

// Get job counts
const counts = await queue.getJobCounts(
  "active",
  "completed",
  "failed",
  "delayed",
  "waiting",
);
```
