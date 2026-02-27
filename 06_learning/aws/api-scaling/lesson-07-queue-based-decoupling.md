# Lesson 07: Queue-Based Decoupling

Critical knowledge about using SQS, SNS, and EventBridge to decouple services - handling traffic spikes, implementing retry strategies, choosing between messaging patterns, and building resilient architectures that scale.

## The Tight Coupling Problem

The interviewer asks: "Your API calls a slow third-party service that takes 10 seconds to respond. Users are getting timeouts. How do you fix this?" The naive answer is "increase the timeout." The correct answer is "decouple with a queue." API Gateway has a 29-second timeout. Users won't wait 10 seconds for a response. Even worse, a spike in traffic creates a spike in third-party calls, overwhelming their API and yours. Queues break this tight coupling - API returns immediately, workers process asynchronously, third-party calls are throttled. Learn when to decouple and which AWS service to use.

You're building a distributed system. Service A calls Service B directly. When Service B is slow or down, Service A is slow or down. When traffic spikes, both services spike. This tight coupling creates cascading failures and makes scaling hard. Queues sit between services, buffering requests and enabling independent scaling. But AWS has three messaging services - SQS, SNS, EventBridge - each for different patterns.

## SQS vs SNS vs EventBridge

| Feature                    | SQS (Queue)                    | SNS (Pub/Sub)                  | EventBridge (Event Bus)           |
| -------------------------- | ------------------------------ | ------------------------------ | --------------------------------- |
| **Pattern**                | Point-to-point (queue)         | Pub/sub (topics)               | Event routing (bus)               |
| **Consumers**              | One consumer per message       | Multiple subscribers           | Multiple targets per rule         |
| **Message retention**      | 1 min - 14 days                | No retention (deliver or drop) | No retention (routes to targets)  |
| **Message size**           | 256 KB                         | 256 KB                         | 256 KB                            |
| **Delivery**               | At-least-once (Standard)       | At-least-once                  | At-least-once                     |
| **Ordering**               | ✓ (FIFO queues)                | ✗ (no ordering guarantee)      | ✗ (no ordering)                   |
| **Filtering**              | ✗ (consumers filter)           | ✓ (filter policies)            | ✓ (event patterns)                |
| **DLQ**                    | ✓                              | ✓ (subscription-level)         | ✓ (target-level)                  |
| **Schema registry**        | ✗                              | ✗                              | ✓                                 |
| **Cost**                   | $0.40 per million              | $0.50 per million              | $1.00 per million                 |
| **Use case**               | Work queues, job processing    | Fan-out, notifications         | Event-driven architectures, routing|

## SQS (Simple Queue Service)

Point-to-point messaging. Producers send messages to queue. Consumers poll queue and process messages.

### Standard vs FIFO Queues

| Feature                | Standard Queue                  | FIFO Queue                        |
| ---------------------- | ------------------------------- | --------------------------------- |
| **Throughput**         | Unlimited (nearly)              | 300 msg/sec (3,000 with batching) |
| **Ordering**           | Best-effort                     | Guaranteed (per message group)    |
| **Delivery**           | At-least-once (may duplicate)   | Exactly-once                      |
| **Use case**           | High throughput, order not critical | Order matters, no duplicates   |
| **Cost**               | $0.40 per million               | $0.50 per million                 |

### Standard Queue Configuration

```yaml
StandardQueue:
  Type: AWS::SQS::Queue
  Properties:
    QueueName: orders-processing-queue
    # Message retention (default 4 days, max 14 days)
    MessageRetentionPeriod: 345600  # 4 days

    # Visibility timeout (how long message is hidden after being received)
    VisibilityTimeout: 30  # 30 seconds

    # Receive message wait time (long polling)
    ReceiveMessageWaitTimeSeconds: 20  # 20 seconds (long polling)

    # Dead letter queue
    RedrivePolicy:
      deadLetterTargetArn: !GetAtt DLQ.Arn
      maxReceiveCount: 3  # Move to DLQ after 3 failed attempts

# Dead letter queue
DLQ:
  Type: AWS::SQS::Queue
  Properties:
    QueueName: orders-dlq
    MessageRetentionPeriod: 1209600  # 14 days (max retention)
```

### FIFO Queue Configuration

```yaml
FIFOQueue:
  Type: AWS::SQS::Queue
  Properties:
    QueueName: payments-processing.fifo  # Must end with .fifo
    FifoQueue: true

    # Content-based deduplication (optional)
    ContentBasedDeduplication: true

    # Deduplication scope
    DeduplicationScope: messageGroup  # or queue

    # FIFO throughput limit
    FifoThroughputLimit: perMessageGroupId  # or perQueue

    VisibilityTimeout: 60
    ReceiveMessageWaitTimeSeconds: 20

    RedrivePolicy:
      deadLetterTargetArn: !GetAtt PaymentsDLQ.Arn
      maxReceiveCount: 5
```

### Producing Messages

```javascript
const AWS = require('aws-sdk');
const sqs = new AWS.SQS();

// ✓ Send message to Standard queue
const sendMessage = async (queueUrl, message) => {
  const params = {
    QueueUrl: queueUrl,
    MessageBody: JSON.stringify(message),
    // Optional: Delay delivery by 10 seconds
    DelaySeconds: 10,
    // Optional: Message attributes
    MessageAttributes: {
      Priority: {
        DataType: 'String',
        StringValue: 'high'
      }
    }
  };

  const result = await sqs.sendMessage(params).promise();
  console.log('Message sent:', result.MessageId);
};

// ✓ Send message to FIFO queue
const sendFIFOMessage = async (queueUrl, message, orderId) => {
  const params = {
    QueueUrl: queueUrl,
    MessageBody: JSON.stringify(message),
    // ✓ Message group ID (ensures ordering within group)
    MessageGroupId: orderId,
    // ✓ Deduplication ID (prevents duplicates within 5 min)
    MessageDeduplicationId: `${orderId}-${Date.now()}`
  };

  await sqs.sendMessage(params).promise();
};

// ✓ Batch send (up to 10 messages, more efficient)
const sendBatch = async (queueUrl, messages) => {
  const entries = messages.map((msg, index) => ({
    Id: `${index}`,
    MessageBody: JSON.stringify(msg)
  }));

  const params = {
    QueueUrl: queueUrl,
    Entries: entries
  };

  const result = await sqs.sendMessageBatch(params).promise();
  console.log('Batch sent:', result.Successful.length, 'messages');
};
```

### Consuming Messages

```javascript
// ✓ Receive and process messages
const processMessages = async (queueUrl) => {
  const params = {
    QueueUrl: queueUrl,
    MaxNumberOfMessages: 10,  # Receive up to 10 messages
    WaitTimeSeconds: 20,  // Long polling (reduces empty responses)
    VisibilityTimeout: 30  // Hide message for 30 sec while processing
  };

  while (true) {
    const data = await sqs.receiveMessage(params).promise();

    if (!data.Messages) {
      continue;  // No messages, keep polling
    }

    // Process messages in parallel
    await Promise.all(data.Messages.map(async (message) => {
      try {
        const body = JSON.parse(message.Body);
        await processOrder(body);

        // ✓ Delete message after successful processing
        await sqs.deleteMessage({
          QueueUrl: queueUrl,
          ReceiptHandle: message.ReceiptHandle
        }).promise();

      } catch (error) {
        console.error('Processing failed:', error);
        // Don't delete - message returns to queue after visibility timeout
        // After maxReceiveCount, moves to DLQ
      }
    }));
  }
};

// ✓ Change visibility timeout (extend processing time)
const extendVisibility = async (queueUrl, receiptHandle) => {
  await sqs.changeMessageVisibility({
    QueueUrl: queueUrl,
    ReceiptHandle: receiptHandle,
    VisibilityTimeout: 60  // Extend by 60 more seconds
  }).promise();
};
```

### Visibility Timeout Deep Dive

```
Timeline:
t=0: Message received by Consumer A
t=0: Message hidden (visibility timeout = 30 sec)
t=0-30: Consumer A processes message
t=15: Consumer A processing takes longer than expected
t=15: Consumer A calls extendVisibility (now hidden until t=75)
t=25: Consumer A completes, deletes message

Alternative timeline (failure):
t=0: Message received by Consumer A
t=0-30: Consumer A crashes during processing
t=30: Visibility timeout expires, message returns to queue
t=30: Message received by Consumer B (retry)
t=50: Consumer B completes, deletes message

If fails 3 times (maxReceiveCount):
  → Moved to Dead Letter Queue
```

## SNS (Simple Notification Service)

Pub/sub messaging. Publishers send messages to topics. Subscribers receive copies of all messages.

### SNS Topic Configuration

```yaml
Topic:
  Type: AWS::SNS::Topic
  Properties:
    TopicName: order-events
    DisplayName: Order Events

    # Delivery policy
    DeliveryPolicy:
      http:
        defaultHealthyRetryPolicy:
          minDelayTarget: 1
          maxDelayTarget: 60
          numRetries: 5
          numNoDelayRetries: 0
          backoffFunction: exponential

# ✓ SQS subscription (reliable delivery)
SQSSubscription:
  Type: AWS::SNS::Subscription
  Properties:
    Protocol: sqs
    TopicArn: !Ref Topic
    Endpoint: !GetAtt OrderProcessingQueue.Arn
    # ✓ Filter policy (only receive relevant messages)
    FilterPolicy:
      eventType:
        - order.created
        - order.updated
    # ✓ Raw message delivery (no SNS wrapper)
    RawMessageDelivery: true

# ✓ Lambda subscription
LambdaSubscription:
  Type: AWS::SNS::Subscription
  Properties:
    Protocol: lambda
    TopicArn: !Ref Topic
    Endpoint: !GetAtt NotificationFunction.Arn
    FilterPolicy:
      eventType:
        - order.completed

# ✓ Email subscription (dev/ops alerts)
EmailSubscription:
  Type: AWS::SNS::Subscription
  Properties:
    Protocol: email
    TopicArn: !Ref Topic
    Endpoint: team@example.com
    FilterPolicy:
      severity:
        - critical
```

### Publishing to SNS

```javascript
const sns = new AWS.SNS();

// ✓ Publish message
const publishMessage = async (topicArn, message) => {
  const params = {
    TopicArn: topicArn,
    Message: JSON.stringify(message),
    Subject: 'Order Created',
    // ✓ Message attributes for filtering
    MessageAttributes: {
      eventType: {
        DataType: 'String',
        StringValue: 'order.created'
      },
      priority: {
        DataType: 'Number',
        StringValue: '1'
      }
    }
  };

  const result = await sns.publish(params).promise();
  console.log('Published:', result.MessageId);
};

// ✓ Publish with different messages per protocol
const publishMultiProtocol = async (topicArn, order) => {
  const params = {
    TopicArn: topicArn,
    Subject: 'Order #' + order.id,
    // Default message (fallback)
    Message: JSON.stringify(order),
    // Protocol-specific messages
    MessageStructure: 'json',
    Message: JSON.stringify({
      default: JSON.stringify(order),  // Fallback
      sqs: JSON.stringify(order),  // Raw data for SQS
      email: `Order ${order.id} has been created for ${order.customer}`,  // Human-readable
      lambda: JSON.stringify({ eventType: 'order.created', order })  // Structured for Lambda
    })
  };

  await sns.publish(params).promise();
};
```

### Filter Policies

```yaml
# ✓ Exact match
FilterPolicy:
  eventType:
    - order.created  # Only receive order.created events

# ✓ Multiple values (OR)
FilterPolicy:
  eventType:
    - order.created
    - order.updated  # Receive created OR updated

# ✓ Numeric matching
FilterPolicy:
  price:
    - numeric: [">=", 100]  # Only orders >= $100

# ✓ Prefix matching
FilterPolicy:
  region:
    - prefix: "us-"  # us-east-1, us-west-2, etc.

# ✓ Complex filtering
FilterPolicy:
  eventType:
    - order.created
  price:
    - numeric: [">=", 100, "<=", 1000]
  region:
    - "us-east-1"
    - "us-west-2"
```

### SNS Fan-Out Pattern

```yaml
# ✓ One message → Multiple consumers
# API Gateway → SNS Topic → [SQS, Lambda, Email]

# Publisher
POST /orders → Lambda → SNS.publish()

# Subscribers
SNS Topic → SQS (order processing)
          → Lambda (send notification)
          → Email (admin alert)
          → SQS (inventory update)
          → SQS (analytics)

# Benefits:
# - Decouple services (each can fail independently)
# - Scale independently (processing queue can backlog)
# - Add new consumers without changing publisher
```

## EventBridge

Event-driven architecture with sophisticated routing and filtering.

### Event Bus and Rules

```yaml
# Custom event bus
EventBus:
  Type: AWS::Events::EventBus
  Properties:
    Name: orders-event-bus

# ✓ Rule: Route high-value orders to priority queue
HighValueOrderRule:
  Type: AWS::Events::Rule
  Properties:
    EventBusName: !Ref EventBus
    EventPattern:
      source:
        - com.myapp.orders
      detail-type:
        - OrderCreated
      detail:
        total:
          - numeric: [">", 1000]  # Orders > $1,000
    Targets:
      - Arn: !GetAtt PriorityQueue.Arn
        Id: PriorityQueueTarget

# ✓ Rule: Route international orders
InternationalOrderRule:
  Type: AWS::Events::Rule
  Properties:
    EventBusName: !Ref EventBus
    EventPattern:
      source:
        - com.myapp.orders
      detail:
        country:
          - anything-but: ["US"]  # Not US
    Targets:
      - Arn: !GetAtt InternationalQueue.Arn

# ✓ Rule: Send all orders to analytics
AllOrdersRule:
  Type: AWS::Events::Rule
  Properties:
    EventBusName: !Ref EventBus
    EventPattern:
      source:
        - com.myapp.orders
    Targets:
      - Arn: !GetAtt AnalyticsStream.Arn
```

### Publishing Events

```javascript
const eventbridge = new AWS.EventBridge();

// ✓ Put event
const publishEvent = async (order) => {
  const params = {
    Entries: [{
      Time: new Date(),
      Source: 'com.myapp.orders',
      DetailType: 'OrderCreated',
      Detail: JSON.stringify({
        orderId: order.id,
        customerId: order.customerId,
        total: order.total,
        country: order.shippingAddress.country,
        items: order.items
      }),
      EventBusName: 'orders-event-bus'
    }]
  };

  const result = await eventbridge.putEvents(params).promise();
  console.log('Event published:', result.Entries[0].EventId);
};

// ✓ Batch events (up to 10)
const publishBatch = async (orders) => {
  const entries = orders.map(order => ({
    Time: new Date(),
    Source: 'com.myapp.orders',
    DetailType: 'OrderCreated',
    Detail: JSON.stringify(order)
  }));

  await eventbridge.putEvents({ Entries: entries }).promise();
};
```

## Handling Traffic Spikes with Queues

### Without Queue (Tight Coupling)

```
API Gateway → Lambda → External API (slow, rate-limited)

Spike: 10,000 req/sec
  → 10,000 concurrent Lambda invocations
  → 10,000 concurrent external API calls
  → External API throttles/fails
  → Lambda fails
  → Users get errors

Problems:
  - No buffering
  - Overwhelming external service
  - No retry mechanism
  - Cascading failures
```

### With Queue (Decoupled)

```
API Gateway → Lambda (fast) → SQS → Lambda (worker) → External API

Spike: 10,000 req/sec
  1. API Lambda: Receives request, writes to SQS, returns 202 Accepted (< 10ms)
  2. SQS: Buffers 10,000 messages
  3. Worker Lambda: Processes at safe rate (100 req/sec)
  4. External API: Receives 100 req/sec (within limits)
  5. Queue drains over ~100 seconds

Benefits:
  - API responds immediately (no waiting)
  - External API protected (throttled consumption)
  - Automatic retries (if processing fails)
  - No cascading failures
```

**Implementation:**

```javascript
// API Lambda - Fast response
exports.apiHandler = async (event) => {
  const order = JSON.parse(event.body);

  // ✓ Write to SQS and return immediately
  await sqs.sendMessage({
    QueueUrl: process.env.QUEUE_URL,
    MessageBody: JSON.stringify(order)
  }).promise();

  return {
    statusCode: 202,  // Accepted
    body: JSON.stringify({
      message: 'Order received, processing started',
      orderId: order.id
    })
  };
};

// Worker Lambda - Processes queue
exports.workerHandler = async (event) => {
  for (const record of event.Records) {
    const order = JSON.parse(record.body);

    try {
      // Call external API
      await axios.post('https://external-api.com/process', order);

      // Success - SQS auto-deletes message
    } catch (error) {
      // Failure - message returns to queue
      // After 3 failures, goes to DLQ
      throw error;
    }
  }
};
```

## Retry Strategies

### Exponential Backoff

```javascript
// ✓ Retry with exponential backoff
const retry = async (fn, maxRetries = 5) => {
  for (let attempt = 0; attempt < maxRetries; attempt++) {
    try {
      return await fn();
    } catch (error) {
      if (attempt === maxRetries - 1) {
        throw error;  // Final attempt failed
      }

      const delay = Math.min(1000 * Math.pow(2, attempt), 30000);  // Max 30 sec
      console.log(`Retry ${attempt + 1} after ${delay}ms`);
      await sleep(delay);
    }
  }
};

// Usage
await retry(async () => {
  return await externalAPI.call(data);
});
```

### SQS Automatic Retry

```yaml
# ✓ Configure retry behavior
Queue:
  VisibilityTimeout: 30  # Hide message for 30 sec during processing

  RedrivePolicy:
    maxReceiveCount: 3  # Retry up to 3 times
    deadLetterTargetArn: !GetAtt DLQ.Arn

# Timeline:
# t=0: Message received, processing fails
# t=30: Visibility timeout expires, message returns to queue (attempt 1)
# t=30: Message received again, processing fails
# t=60: Message returns (attempt 2)
# t=60: Message received again, processing fails
# t=90: Message returns (attempt 3, maxReceiveCount reached)
# t=90: Message moved to DLQ
```

### Dead Letter Queue Analysis

```javascript
// ✓ Monitor DLQ for failures
const monitorDLQ = async (dlqUrl) => {
  const messages = await sqs.receiveMessage({
    QueueUrl: dlqUrl,
    MaxNumberOfMessages: 10,
    WaitTimeSeconds: 20
  }).promise();

  if (messages.Messages) {
    for (const msg of messages.Messages) {
      const body = JSON.parse(msg.Body);
      console.error('Failed message:', body);

      // Log to CloudWatch, alert team, etc.
      await logFailure(body);

      // ✓ After analysis, delete from DLQ
      await sqs.deleteMessage({
        QueueUrl: dlqUrl,
        ReceiptHandle: msg.ReceiptHandle
      }).promise();
    }
  }
};

// ✓ Replay DLQ messages (after fixing issue)
const replayDLQ = async (dlqUrl, mainQueueUrl) => {
  const messages = await sqs.receiveMessage({
    QueueUrl: dlqUrl,
    MaxNumberOfMessages: 10
  }).promise();

  if (messages.Messages) {
    for (const msg of messages.Messages) {
      // Send back to main queue
      await sqs.sendMessage({
        QueueUrl: mainQueueUrl,
        MessageBody: msg.Body
      }).promise();

      // Delete from DLQ
      await sqs.deleteMessage({
        QueueUrl: dlqUrl,
        ReceiptHandle: msg.ReceiptHandle
      }).promise();
    }
  }
};
```

## Backpressure Handling

When consumers can't keep up with producers, queue depth grows. Handle this before running out of memory/storage.

### Monitor Queue Depth

```yaml
# ✓ CloudWatch alarm on queue depth
QueueDepthAlarm:
  Type: AWS::CloudWatch::Alarm
  Properties:
    AlarmName: high-queue-depth
    MetricName: ApproximateNumberOfMessagesVisible
    Namespace: AWS/SQS
    Dimensions:
      - Name: QueueName
        Value: !GetAtt Queue.QueueName
    Statistic: Average
    Period: 300
    EvaluationPeriods: 2
    Threshold: 10000  # Alert when > 10K messages
    ComparisonOperator: GreaterThanThreshold
```

### Scale Consumers Based on Queue Depth

```yaml
# ✓ Auto-scale Lambda concurrency
ScalingTarget:
  Type: AWS::ApplicationAutoScaling::ScalableTarget
  Properties:
    ServiceNamespace: lambda
    ResourceId: !Sub "function:${WorkerFunction}:provisioned-concurrency"
    ScalableDimension: lambda:function:ProvisionedConcurrentExecutions
    MinCapacity: 10
    MaxCapacity: 100

ScalingPolicy:
  Type: AWS::ApplicationAutoScaling::ScalingPolicy
  Properties:
    PolicyType: TargetTrackingScaling
    TargetTrackingScalingPolicyConfiguration:
      CustomizedMetricSpecification:
        MetricName: ApproximateNumberOfMessagesVisible
        Namespace: AWS/SQS
        Dimensions:
          - Name: QueueName
            Value: !GetAtt Queue.QueueName
        Statistic: Average
      TargetValue: 100  # Keep ~100 messages per Lambda

# Or scale ECS tasks
ECSScalingPolicy:
  Type: AWS::ApplicationAutoScaling::ScalingPolicy
  Properties:
    PolicyType: TargetTrackingScaling
    TargetTrackingScalingPolicyConfiguration:
      CustomizedMetricSpecification:
        # Queue depth ÷ desired tasks
        # Target: Each task processes ~100 messages
```

### Rate Limiting Producers

```javascript
// ✓ Use SQS SendMessage delay
const sendWithBackoff = async (queueUrl, message) => {
  const queueDepth = await getQueueDepth(queueUrl);

  let delay = 0;
  if (queueDepth > 10000) {
    delay = 300;  // 5 min delay if queue is large
  } else if (queueDepth > 5000) {
    delay = 60;  // 1 min delay
  }

  await sqs.sendMessage({
    QueueUrl: queueUrl,
    MessageBody: JSON.stringify(message),
    DelaySeconds: delay
  }).promise();
};
```

## Common Patterns

### 1. Request-Response with Correlation ID

```javascript
// ✓ Async request-response pattern
const requestId = uuid();

// 1. Send request
await sqs.sendMessage({
  QueueUrl: REQUEST_QUEUE_URL,
  MessageBody: JSON.stringify({ orderId, requestId }),
  MessageAttributes: {
    ReplyTo: {
      DataType: 'String',
      StringValue: RESPONSE_QUEUE_URL
    }
  }
}).promise();

// 2. Poll response queue for matching requestId
const pollResponse = async (requestId) => {
  const timeout = Date.now() + 30000;  // 30 sec timeout

  while (Date.now() < timeout) {
    const messages = await sqs.receiveMessage({
      QueueUrl: RESPONSE_QUEUE_URL,
      WaitTimeSeconds: 5
    }).promise();

    for (const msg of messages.Messages || []) {
      const response = JSON.parse(msg.Body);
      if (response.requestId === requestId) {
        await sqs.deleteMessage({
          QueueUrl: RESPONSE_QUEUE_URL,
          ReceiptHandle: msg.ReceiptHandle
        }).promise();
        return response;
      }
    }
  }

  throw new Error('Response timeout');
};
```

### 2. Priority Queues

```javascript
// ✓ Multiple queues for different priorities
const routeByPriority = async (order) => {
  let queueUrl;

  if (order.total > 1000) {
    queueUrl = HIGH_PRIORITY_QUEUE_URL;
  } else if (order.total > 100) {
    queueUrl = MEDIUM_PRIORITY_QUEUE_URL;
  } else {
    queueUrl = LOW_PRIORITY_QUEUE_URL;
  }

  await sqs.sendMessage({
    QueueUrl: queueUrl,
    MessageBody: JSON.stringify(order)
  }).promise();
};

// Workers process high-priority queue first
// Lambda concurrency: 100 high, 50 medium, 20 low
```

### 3. Circuit Breaker

```javascript
// ✓ Stop processing if external service is down
class CircuitBreaker {
  constructor(threshold = 5, timeout = 60000) {
    this.failureCount = 0;
    this.threshold = threshold;
    this.timeout = timeout;
    this.state = 'CLOSED';  // CLOSED, OPEN, HALF_OPEN
    this.nextAttempt = Date.now();
  }

  async execute(fn) {
    if (this.state === 'OPEN') {
      if (Date.now() < this.nextAttempt) {
        throw new Error('Circuit breaker is OPEN');
      }
      this.state = 'HALF_OPEN';
    }

    try {
      const result = await fn();
      this.onSuccess();
      return result;
    } catch (error) {
      this.onFailure();
      throw error;
    }
  }

  onSuccess() {
    this.failureCount = 0;
    this.state = 'CLOSED';
  }

  onFailure() {
    this.failureCount++;
    if (this.failureCount >= this.threshold) {
      this.state = 'OPEN';
      this.nextAttempt = Date.now() + this.timeout;
    }
  }
}

// Usage
const breaker = new CircuitBreaker();

exports.handler = async (event) => {
  for (const record of event.Records) {
    try {
      await breaker.execute(async () => {
        return await externalAPI.call(record.body);
      });
    } catch (error) {
      if (error.message === 'Circuit breaker is OPEN') {
        // Stop processing, return messages to queue
        throw error;
      }
      // Other errors: continue processing
    }
  }
};
```

## Hands-On Exercise 1: Choose the Right Service

For each scenario, choose SQS, SNS, or EventBridge and justify your choice.

**Scenario 1: Order Processing**
- API receives orders
- Orders must be processed exactly once, in order (per customer)
- Multiple downstream services need order data (inventory, shipping, analytics)

**Scenario 2: Real-Time Notifications**
- Send notifications when events occur
- Multiple channels: Email, SMS, Push, Webhook
- Some subscribers only care about specific events (high-value orders)

**Scenario 3: Event-Driven Microservices**
- 20+ microservices
- Each service publishes/consumes events
- Need sophisticated routing (e.g., "route payment events to fraud detection if amount > $500")
- Need schema validation

**Scenario 4: Async Job Processing**
- Upload images, process in background
- Workers scale based on queue depth
- Need retry on failure (up to 3 times)
- Low cost priority

<details>
<summary>Solution</summary>

**Scenario 1: SQS FIFO + SNS Fan-Out**

```yaml
# ✓ SQS FIFO for order processing (exactly-once, ordered)
OrderQueue:
  Type: AWS::SQS::Queue
  Properties:
    QueueName: orders.fifo
    FifoQueue: true

# ✓ SNS for fan-out to downstream services
OrderTopic:
  Type: AWS::SNS::Topic

# Subscriptions
InventoryQueue:
  Type: AWS::SQS::Queue
ShippingQueue:
  Type: AWS::SQS::Queue
AnalyticsQueue:
  Type: AWS::SQS::Queue

# Flow:
# API → SQS FIFO (order processing) → Lambda → SNS Topic → [Inventory, Shipping, Analytics queues]
```

**Why:**
- SQS FIFO: Guarantees exactly-once processing, order per customer (MessageGroupId)
- SNS fan-out: Multiple downstream services get notified
- Each service has own SQS queue (can process at own pace, retry independently)

**Scenario 2: SNS with Filter Policies**

```yaml
# ✓ SNS Topic with filtered subscriptions
NotificationTopic:
  Type: AWS::SNS::Topic

# Email subscription (all events)
EmailSubscription:
  Protocol: email
  Endpoint: team@example.com

# SMS subscription (high-value only)
SMSSubscription:
  Protocol: sms
  Endpoint: "+1234567890"
  FilterPolicy:
    orderValue:
      - numeric: [">", 1000]

# Push notification (customer-specific)
PushSubscription:
  Protocol: lambda
  Endpoint: !GetAtt PushFunction.Arn

# Webhook (third-party integration)
WebhookSubscription:
  Protocol: https
  Endpoint: https://partner.com/webhook
```

**Why:**
- SNS: Perfect for fan-out notifications
- Filter policies: Subscribers receive only relevant events
- Multiple protocols: Email, SMS, Lambda, HTTPS all supported
- Cost-effective: $0.50/million (cheaper than EventBridge)

**Scenario 3: EventBridge**

```yaml
# ✓ EventBridge for sophisticated event routing
EventBus:
  Type: AWS::Events::EventBus

# Rule: High-value payments to fraud detection
FraudRule:
  Type: AWS::Events::Rule
  Properties:
    EventPattern:
      source: [com.myapp.payments]
      detail-type: [PaymentProcessed]
      detail:
        amount:
          - numeric: [">", 500]
    Targets:
      - Arn: !GetAtt FraudDetectionQueue.Arn

# Schema registry for validation
Schema:
  Type: AWS::EventSchemas::Schema
  Properties:
    RegistryName: myapp-schemas
    Type: OpenApi3
    Content: # JSON Schema definition
```

**Why:**
- EventBridge: Best for complex event-driven architectures
- Sophisticated routing: Content-based filtering on nested fields
- Schema registry: Validates events, generates code
- Native integrations: 20+ AWS services
- Worth higher cost ($1/million) for large microservices architecture

**Scenario 4: SQS Standard**

```yaml
# ✓ SQS Standard for simple job queue
ImageProcessingQueue:
  Type: AWS::SQS::Queue
  Properties:
    VisibilityTimeout: 300  # 5 min processing time
    RedrivePolicy:
      maxReceiveCount: 3
      deadLetterTargetArn: !GetAtt DLQ.Arn
```

**Why:**
- SQS Standard: Cheapest option ($0.40/million)
- Order doesn't matter (image processing)
- Built-in retry (up to 3 attempts)
- Auto-scales Lambda based on queue depth
- Simple, proven pattern

</details>

## Hands-On Exercise 2: Debug Queue Performance

**Problem:** SQS queue has 50,000 messages backlog and growing. Worker Lambda processes messages but queue isn't draining.

**Current configuration:**

```yaml
Queue:
  VisibilityTimeout: 300  # 5 minutes
  ReceiveMessageWaitTimeSeconds: 0  # No long polling

WorkerFunction:
  MemorySize: 128 MB
  Timeout: 30
  ReservedConcurrentExecutions: 10

EventSourceMapping:
  BatchSize: 1  # Process 1 message per invocation
```

**Worker code:**

```javascript
exports.handler = async (event) => {
  for (const record of event.Records) {
    await processImage(JSON.parse(record.body));  // Takes ~5 seconds
  }
};
```

**Symptoms:**
- Queue depth: 50,000 messages
- Worker Lambda: Processes ~2 messages/sec
- Expected: 10 messages/sec

Identify all issues and fix them.

<details>
<summary>Solution</summary>

**Issues identified:**

1. ❌ `BatchSize: 1` (inefficient, high invocation cost)
2. ❌ `ReservedConcurrentExecutions: 10` (too low)
3. ❌ `ReceiveMessageWaitTimeSeconds: 0` (short polling, inefficient)
4. ❌ Processing synchronously (no parallelization)
5. ❌ Low memory (128 MB = slow CPU)

**Fixed configuration:**

```yaml
Queue:
  VisibilityTimeout: 60  # ✓ Reduced (matches function timeout × 2)
  ReceiveMessageWaitTimeSeconds: 20  # ✓ Long polling

WorkerFunction:
  MemorySize: 1024 MB  # ✓ More memory = faster CPU
  Timeout: 30
  ReservedConcurrentExecutions: 100  # ✓ Increased capacity

EventSourceMapping:
  BatchSize: 10  # ✓ Process 10 messages per invocation
  ScalingConfig:
    MaximumConcurrency: 100  # ✓ Scale up to 100 concurrent invocations
```

**Fixed code:**

```javascript
exports.handler = async (event) => {
  // ✓ Process messages in parallel
  await Promise.all(event.Records.map(async (record) => {
    try {
      await processImage(JSON.parse(record.body));
    } catch (error) {
      console.error('Failed:', record.messageId, error);
      // Throw to return message to queue for retry
      throw error;
    }
  }));
};
```

**Performance analysis:**

```
Before:
  Concurrency: 10
  Batch size: 1
  Processing time: 5 sec
  Throughput: 10 concurrent × 1 message × (1 / 5 sec) = 2 messages/sec
  Time to drain 50K messages: 50,000 ÷ 2 = 25,000 seconds = 7 hours

After:
  Concurrency: 100
  Batch size: 10
  Processing time: 5 sec (parallel processing within batch)
  Throughput: 100 concurrent × 10 messages × (1 / 5 sec) = 200 messages/sec
  Time to drain 50K messages: 50,000 ÷ 200 = 250 seconds = 4 minutes

100× faster!
```

**Cost comparison:**

```
Before:
  50K messages ÷ 1 per invocation = 50,000 invocations
  Invocation cost: 50,000 × $0.0000002 = $0.01
  Duration: 50,000 × 5 sec × 0.125 GB = 31,250 GB-sec = $0.52
  Total: $0.53

After:
  50K messages ÷ 10 per invocation = 5,000 invocations
  Invocation cost: 5,000 × $0.0000002 = $0.001
  Duration: 5,000 × 5 sec × 1 GB = 25,000 GB-sec = $0.42
  Total: $0.42

20% cheaper + 100× faster
```

</details>

## Key Takeaways

1. **SQS**: Point-to-point queues, use for work distribution, job processing, decoupling services
2. **SNS**: Pub/sub fan-out, use for notifications, multiple consumers, event broadcasting
3. **EventBridge**: Event routing with sophisticated filtering, use for event-driven architectures
4. **FIFO vs Standard**: FIFO for order/exactly-once (300 msg/sec), Standard for high throughput (unlimited)
5. **Visibility Timeout**: Set to 2× function timeout, prevents duplicate processing
6. **Long Polling**: Always use (`ReceiveMessageWaitTimeSeconds: 20`) - reduces cost, improves efficiency
7. **Batch Processing**: Use `BatchSize: 10` - reduces invocations 10×, processes in parallel
8. **DLQ**: Always configure for failed messages, monitor and replay after fixing issues

## Conclusion

You've completed the AWS API Scaling series! You now understand:

- **API Gateway vs ALB**: When to use each, cost optimization strategies
- **Load Balancers**: ALB vs NLB, target groups, health checks
- **CloudFront**: Edge caching, Lambda@Edge, cost optimization through cache hit ratio
- **Auto Scaling**: Target tracking, warm pools, lifecycle hooks, metrics selection
- **Lambda at Scale**: Concurrency management, cold start mitigation, VPC considerations
- **Queue-Based Decoupling**: SQS, SNS, EventBridge patterns for resilient architectures

These patterns combine to build APIs that scale to millions of requests while controlling costs. Practice designing systems that layer these services strategically - CloudFront caching reduces ALB load, queues decouple services and handle spikes, auto scaling adjusts capacity automatically.

The key to scaling is not just knowing the services, but knowing when to use each and how they work together. Keep practicing with real scenarios and you'll be ready for any scaling interview question.
