# KafkaJS

```sh
npm install kafkajs
```

> Pure JavaScript Kafka client. No native dependencies. Supports producer, consumer, and admin APIs.

## Setup

```typescript
import { Kafka, logLevel } from "kafkajs";

const kafka = new Kafka({
  clientId: "my-service",
  brokers: ["localhost:9092"],
  logLevel: logLevel.WARN,
});
```

## Producer

```typescript
const producer = kafka.producer();

await producer.connect();

// Send a message
await producer.send({
  topic: "my-topic",
  messages: [
    {
      key: "user-123",
      value: JSON.stringify({ event: "created", userId: "123" }),
    },
  ],
});

// Send batch
await producer.send({
  topic: "my-topic",
  messages: [
    { key: "1", value: JSON.stringify(event1) },
    { key: "2", value: JSON.stringify(event2) },
  ],
});

// Send to multiple topics
await producer.sendBatch({
  topicMessages: [
    { topic: "topic-a", messages: [{ value: "msg1" }] },
    { topic: "topic-b", messages: [{ value: "msg2" }] },
  ],
});

await producer.disconnect();
```

## Consumer

```typescript
const consumer = kafka.consumer({ groupId: "my-group" });

await consumer.connect();
await consumer.subscribe({ topic: "my-topic", fromBeginning: false });

await consumer.run({
  eachMessage: async ({ topic, partition, message }) => {
    const key = message.key?.toString();
    const value = message.value?.toString();
    console.log(`${topic}[${partition}] key=${key} value=${value}`);
  },
});
```

## Consumer — batch processing

```typescript
await consumer.run({
  eachBatch: async ({ batch, resolveOffset, heartbeat }) => {
    for (const message of batch.messages) {
      await processMessage(message);
      resolveOffset(message.offset); // mark as processed
      await heartbeat(); // keep consumer alive during long batches
    }
  },
});
```

## Consumer — manual commit

```typescript
const consumer = kafka.consumer({
  groupId: "my-group",
  // autoCommit is enabled by default
});

// Disable auto-commit for at-least-once processing
await consumer.run({
  autoCommit: false,
  eachMessage: async ({ topic, partition, message }) => {
    await processMessage(message);

    await consumer.commitOffsets([
      { topic, partition, offset: (Number(message.offset) + 1).toString() },
    ]);
  },
});
```

## Graceful shutdown

```typescript
const signals: NodeJS.Signals[] = ["SIGINT", "SIGTERM"];

for (const signal of signals) {
  process.on(signal, async () => {
    await consumer.disconnect();
    await producer.disconnect();
    process.exit(0);
  });
}
```

## Error handling & retries

```typescript
await consumer.run({
  eachMessage: async ({ topic, partition, message }) => {
    try {
      await processMessage(message);
    } catch (err) {
      console.error(`Failed to process message`, {
        topic,
        partition,
        offset: message.offset,
        error: err,
      });

      // Option 1: Send to dead letter topic
      await producer.send({
        topic: `${topic}.dlq`,
        messages: [{ key: message.key, value: message.value }],
      });

      // Option 2: throw to trigger retry (consumer restarts from last commit)
      // throw err;
    }
  },
});
```

## Admin (create topics)

```typescript
const admin = kafka.admin();
await admin.connect();

await admin.createTopics({
  topics: [{ topic: "my-topic", numPartitions: 3, replicationFactor: 1 }],
});

const topics = await admin.listTopics();
const metadata = await admin.fetchTopicMetadata({ topics: ["my-topic"] });

await admin.disconnect();
```

## Headers

```typescript
// Produce with headers
await producer.send({
  topic: "my-topic",
  messages: [
    {
      key: "user-123",
      value: JSON.stringify(event),
      headers: {
        "correlation-id": correlationId,
        "event-type": "user.created",
      },
    },
  ],
});

// Read headers
await consumer.run({
  eachMessage: async ({ message }) => {
    const eventType = message.headers?.["event-type"]?.toString();
    const correlationId = message.headers?.["correlation-id"]?.toString();
  },
});
```
