# Lesson 06: Lambda at Scale

Critical knowledge about running AWS Lambda at scale - understanding concurrency limits, mitigating cold starts, VPC implications, event source mapping patterns, and cost optimization for serverless workloads.

## The Serverless Scaling Problem

The interviewer asks: "Your Lambda function gets 100,000 invocations per second. What happens?" Most developers say "Lambda scales automatically!" But there's a regional concurrency limit (1,000 by default). Hit that limit and new invocations get throttled (429 errors). Beyond limits, cold starts add 1-3 seconds of latency. VPC Lambdas add 10+ seconds. At scale, these issues compound - your "infinitely scalable" serverless function becomes your bottleneck. Know how to configure concurrency, eliminate cold starts, and optimize for high throughput.

Lambda scales automatically by launching execution environments. But "automatic" doesn't mean "infinite" or "instant." You hit account limits, cold starts slow responses, and costs explode at high volume. Understand how Lambda scaling works, how to configure it, and when Lambda is the wrong choice.

## Concurrency Types

| Type                      | Behavior                          | Use Case                        | Cost Impact              |
| ------------------------- | --------------------------------- | ------------------------------- | ------------------------ |
| **Unreserved**            | Shares account pool               | Default, most functions         | Pay per invoke           |
| **Reserved**              | Dedicated capacity, never shared  | Critical functions (SLA)        | No extra cost            |
| **Provisioned**           | Pre-warmed, always ready          | Latency-sensitive               | Pay for provisioned time |

## Concurrency Limits

### Account-Level Limit

```
Default: 1,000 concurrent executions per region
Can request increase: Up to tens of thousands

Concurrent executions = Invocations per second × Average duration

Example:
  1,000 req/sec × 0.5 sec avg duration = 500 concurrent executions
  (Uses 50% of 1,000 limit)

  5,000 req/sec × 2 sec avg duration = 10,000 concurrent executions
  (Exceeds 1,000 limit → throttled)
```

### Throttling Behavior

```
When limit exceeded:
  - Synchronous invocations (API Gateway, ALB): 429 TooManyRequestsException
  - Asynchronous invocations (S3, SNS, EventBridge): Retried automatically (exponential backoff)
  - Event source mappings (SQS, Kinesis, DynamoDB): Throttled, retries with backoff
```

### Reserved Concurrency

Dedicate concurrency to specific functions. Guarantees capacity but limits maximum concurrent executions.

```yaml
Function:
  Type: AWS::Lambda::Function
  Properties:
    FunctionName: critical-api-function
    Runtime: nodejs18.x
    Handler: index.handler
    # ✓ Reserve 500 concurrent executions
    ReservedConcurrentExecutions: 500

# Effect:
# - This function always has 500 executions available
# - Other functions share remaining 500 (1,000 - 500)
# - This function can NEVER exceed 500 (hard limit)

# Use when:
# - Critical function must always have capacity
# - Want to prevent one function from starving others
# - Limit runaway function (prevent cost explosion)
```

**Example scenario:**

```
Account limit: 1,000
Function A: Reserved 700
Function B: Reserved 200
Function C: No reservation

Available concurrency:
  Function A: 700 (guaranteed)
  Function B: 200 (guaranteed)
  Function C: 100 (remaining pool)

If Function C spikes:
  - Can use up to 100 concurrent executions
  - Won't steal from Function A or B
  - Throttled beyond 100
```

### Provisioned Concurrency

Pre-initialize execution environments (eliminates cold starts).

```yaml
Function:
  Type: AWS::Lambda::Function
  Properties:
    FunctionName: latency-sensitive-api

# Provisioned concurrency
ProvisionedConcurrencyConfig:
  Type: AWS::Lambda::EventInvokeConfig
  Properties:
    FunctionName: !Ref Function
    Qualifier: !Ref FunctionVersion
    ProvisionedConcurrentExecutions: 100

# Cost:
# $0.000004167 per GB-second (4.17× more than on-demand)
# Plus normal invocation cost

# Example:
# 100 provisioned × 1 GB × 730 hours = 73,000 GB-hours
# 73,000 × $0.000015 = $1,095/month (provisioned cost)
# vs ~$250/month on-demand (with cold starts)

# When to use:
# - Latency SLA <100ms (can't tolerate cold starts)
# - High request rate with variable traffic
# - Cost justified by business requirements
```

### Application Auto Scaling for Provisioned Concurrency

```yaml
ScalableTarget:
  Type: AWS::ApplicationAutoScaling::ScalableTarget
  Properties:
    ServiceNamespace: lambda
    ResourceId: !Sub "function:${Function}:${FunctionVersion}"
    ScalableDimension: lambda:function:ProvisionedConcurrentExecutions
    MinCapacity: 10
    MaxCapacity: 100

# ✓ Track utilization
ScalingPolicy:
  Type: AWS::ApplicationAutoScaling::ScalingPolicy
  Properties:
    PolicyType: TargetTrackingScaling
    ScalingTargetId: !Ref ScalableTarget
    TargetTrackingScalingPolicyConfiguration:
      PredefinedMetricSpecification:
        PredefinedMetricType: LambdaProvisionedConcurrencyUtilization
      TargetValue: 0.7  # Keep 70% utilized

# Scales provisioned concurrency based on usage
# Reduces cost during low traffic, maintains capacity during high traffic
```

## Cold Starts

Cold start: Time to initialize execution environment before running your code.

### Cold Start Components

```
Total cold start time:

1. Download code package (100-500ms)
   - Depends on package size
   - Layers downloaded separately

2. Start runtime (50-200ms)
   - Node.js: ~50ms
   - Python: ~100ms
   - Java: ~300-500ms
   - .NET: ~400-600ms

3. Initialize handler code (0-3000ms)
   - Import dependencies
   - Initialize SDK clients
   - Establish connections

Total: 200ms - 4+ seconds (depends on language + dependencies)
```

### Cold Start Mitigation Strategies

**1. Provisioned Concurrency (Best, Most Expensive)**

```yaml
# ✓ Zero cold starts
ProvisionedConcurrentExecutions: 50

# Cost: ~4× more than on-demand
# Benefit: Guaranteed <100ms latency
```

**2. Keep Functions Warm (Medium Cost, Good)**

```yaml
# EventBridge rule to ping function every 5 minutes
WarmupRule:
  Type: AWS::Events::Rule
  Properties:
    ScheduleExpression: rate(5 minutes)
    Targets:
      - Arn: !GetAtt Function.Arn
        Input: '{"warmup": true}'

# Lambda handler
exports.handler = async (event) => {
  // Ignore warmup events
  if (event.warmup) {
    return { statusCode: 200, body: 'Warmup' };
  }

  // Normal processing
  return processRequest(event);
};

# Cost: Minimal (12 × 24 × 30 = 8,640 invocations/month ≈ free)
# Benefit: Reduces cold starts by keeping ~1 environment warm
# Limitation: Only warms 1 instance (doesn't help under load)
```

**3. Minimize Package Size**

```javascript
// ❌ Bad - Large dependencies
const AWS = require('aws-sdk');  // 50+ MB
const lodash = require('lodash');  // Import entire library

// ✓ Good - Minimal dependencies
const { DynamoDB } = require('@aws-sdk/client-dynamodb');  // Only what you need
const get = require('lodash/get');  // Import specific function
```

**4. Optimize Initialization Code**

```javascript
// ❌ Bad - Initialize on every invocation
exports.handler = async (event) => {
  const AWS = require('aws-sdk');  // ← Runs every time
  const dynamodb = new AWS.DynamoDB.DocumentClient();
  const response = await dynamodb.get(params).promise();
  return response;
};

// ✓ Good - Initialize once (outside handler)
const AWS = require('aws-sdk');
const dynamodb = new AWS.DynamoDB.DocumentClient();

exports.handler = async (event) => {
  // SDK client reused across invocations
  const response = await dynamodb.get(params).promise();
  return response;
};
```

**5. Use Lambda SnapStart (Java only)**

```yaml
Function:
  Type: AWS::Lambda::Function
  Properties:
    Runtime: java17
    SnapStart:
      ApplyOn: PublishedVersions

# How it works:
# 1. Lambda initializes function
# 2. Takes snapshot of initialized state
# 3. Restores from snapshot on cold start (10× faster)

# Cold start reduction:
# - Java without SnapStart: 3-5 seconds
# - Java with SnapStart: 200-500ms
```

## VPC Lambda Performance

Lambda functions in VPC need Elastic Network Interfaces (ENIs) to access VPC resources.

### The ENI Problem (Pre-2019)

```
Old behavior (before Sept 2019):
  1. Lambda creates ENI in VPC (30-90 seconds!)
  2. ENI attached to execution environment
  3. Cold start: 30-90 seconds

Problem: Every new concurrent execution needed new ENI
Result: Massive cold starts at scale
```

### Modern VPC Lambda (Post-2019)

```
New behavior (after Sept 2019):
  - Lambda creates shared ENIs per subnet/security group
  - All executions share ENIs
  - Cold start: Same as non-VPC Lambda (~1-2 sec)

VPC Lambda is now fast!
```

### VPC Configuration

```yaml
Function:
  Type: AWS::Lambda::Function
  Properties:
    VpcConfig:
      SubnetIds:
        - subnet-1234567890abcdef0  # Private subnet
        - subnet-0fedcba0987654321  # Private subnet
      SecurityGroupIds:
        - !Ref LambdaSecurityGroup

# ✓ Use private subnets
# ✓ Attach NAT Gateway for internet access
# ❌ Don't use public subnets (Lambda doesn't get public IP)
```

### When to Use VPC Lambda

```
✓ Use VPC when:
  - Access RDS databases in private subnets
  - Access ElastiCache in private subnets
  - Access private APIs (ALB, internal services)
  - Compliance requires VPC (data isolation)

❌ Avoid VPC when:
  - Only accessing public AWS services (DynamoDB, S3, SNS)
  - Don't need VPC resources
  - Simplicity preferred (no VPC = simpler)

Alternative to VPC:
  - Use RDS Proxy (public endpoint, IAM auth)
  - Use DynamoDB instead of RDS
  - Use API Gateway + VPC Link (Lambda outside VPC calls services inside VPC)
```

## Event Source Mappings

Lambda polls event sources and invokes your function with batches.

### SQS Event Source Mapping

```yaml
EventSourceMapping:
  Type: AWS::Lambda::EventSourceMapping
  Properties:
    FunctionName: !Ref Function
    EventSourceArn: !GetAtt MyQueue.Arn
    BatchSize: 10  # Process up to 10 messages per invocation
    MaximumBatchingWindowInSeconds: 5  # Wait 5 sec to fill batch

    # Scaling
    ScalingConfig:
      MaximumConcurrency: 100  # Max 100 concurrent Lambda invocations

    # Error handling
    FunctionResponseTypes:
      - ReportBatchItemFailures  # Partial batch failure support

# How it works:
# 1. Lambda polls SQS (long polling)
# 2. Receives up to 10 messages
# 3. Invokes function with batch
# 4. If success: Deletes messages from queue
# 5. If failure: Messages return to queue (after visibility timeout)
```

**Batch processing:**

```javascript
exports.handler = async (event) => {
  // event.Records contains up to 10 SQS messages
  const results = [];

  for (const record of event.Records) {
    try {
      await processMessage(JSON.parse(record.body));
    } catch (error) {
      console.error('Failed:', record.messageId, error);
      // ✓ Report individual failure (partial batch failure)
      results.push({ itemIdentifier: record.messageId });
    }
  }

  // Return failed message IDs (rest are deleted)
  return { batchItemFailures: results };
};
```

### Kinesis/DynamoDB Streams Event Source Mapping

```yaml
EventSourceMapping:
  Type: AWS::Lambda::EventSourceMapping
  Properties:
    FunctionName: !Ref Function
    EventSourceArn: !GetAtt KinesisStream.Arn
    StartingPosition: LATEST  # or TRIM_HORIZON
    BatchSize: 100  # Process up to 100 records
    MaximumBatchingWindowInSeconds: 10
    ParallelizationFactor: 10  # Process 10 batches per shard in parallel

    # Error handling
    DestinationConfig:
      OnFailure:
        Destination: !GetAtt DLQ.Arn  # Send failed batches to DLQ
    MaximumRetryAttempts: 3
    MaximumRecordAgeInSeconds: 3600  # Drop records older than 1 hour

    # Tumbling window (aggregation)
    TumblingWindowInSeconds: 60  # Aggregate records over 60 sec window
```

**Key difference from SQS:**

```
SQS:
  - Deletes processed messages
  - Messages can be processed by multiple consumers
  - No order guarantee (except FIFO)

Kinesis/DynamoDB Streams:
  - Records remain in stream (deleted after retention period)
  - Each record processed exactly once per shard
  - Order guaranteed per shard
```

## Cost Optimization

### Lambda Pricing

```
Invocation cost: $0.20 per 1 million requests
Duration cost: $0.0000166667 per GB-second

Example calculation:
  1 million invocations
  128 MB memory
  200ms average duration

Invocation: 1M × $0.0000002 = $0.20
Duration: 1M × 0.2 sec × 0.125 GB × $0.0000166667 = $0.417
Total: $0.617 per million invocations

At 1 GB memory:
  Duration: 1M × 0.2 sec × 1 GB × $0.0000166667 = $3.33
  Total: $3.53 per million invocations

8× more expensive for 8× more memory
```

### Memory vs Duration Trade-off

Lambda CPU scales with memory. More memory = faster execution = lower duration cost.

```javascript
// Test with different memory settings:

128 MB: 1,000ms duration
  Cost: 1M × 1 sec × 0.125 GB × $0.0000166667 = $2.08

512 MB: 300ms duration (faster CPU)
  Cost: 1M × 0.3 sec × 0.5 GB × $0.0000166667 = $2.50

1,024 MB: 200ms duration
  Cost: 1M × 0.2 sec × 1 GB × $0.0000166667 = $3.33

3,008 MB: 150ms duration
  Cost: 1M × 0.15 sec × 2.938 GB × $0.0000166667 = $7.35

Sweet spot: 512-1024 MB (balance cost and performance)
```

**Use AWS Lambda Power Tuning tool:**

```bash
# Finds optimal memory configuration
# Tests function at different memory settings
# Recommends best cost/performance balance

npm install -g lambda-power-tuning
lambda-power-tuning --function my-function
```

### Avoid Over-Invocation

```javascript
// ❌ Bad - Lambda invokes Lambda (double cost)
// API Gateway → Lambda A → invoke Lambda B

exports.handlerA = async (event) => {
  const lambda = new AWS.Lambda();
  await lambda.invoke({
    FunctionName: 'function-b',
    InvokeType: 'Event',
    Payload: JSON.stringify(data)
  }).promise();
};

// ✓ Good - Direct integration
// API Gateway → Lambda B (via SQS or direct)

exports.handler = async (event) => {
  const sqs = new AWS.SQS();
  await sqs.sendMessage({
    QueueUrl: process.env.QUEUE_URL,
    MessageBody: JSON.stringify(data)
  }).promise();
  // Lambda B triggered by SQS event source mapping
};

// Or: API Gateway → Lambda B directly (configure in API Gateway)
```

### Reduce Package Size

```bash
# ❌ Bad - Include dev dependencies
package.json:
  "dependencies": {
    "aws-sdk": "^2.1000.0",  # Already in Lambda runtime
    "lodash": "^4.17.21"
  }
  "devDependencies": {
    "typescript": "^4.5.0"  # Accidentally included
  }

Deployment: zip node_modules/ (150 MB)

# ✓ Good - Production only, exclude aws-sdk
package.json:
  "dependencies": {
    "lodash": "^4.17.21"
  }

npm ci --production  # Install prod only
zip deployment.zip index.js node_modules/ (5 MB)

# Even better - Use layers for shared dependencies
aws lambda publish-layer-version \
  --layer-name shared-dependencies \
  --zip-file fileb://layer.zip

# Function only includes code (< 1 MB)
```

## Common Mistakes

### Mistake 1: Not Handling Throttling

```javascript
// ❌ Bad - No retry logic
exports.handler = async (event) => {
  const dynamodb = new AWS.DynamoDB.DocumentClient();
  const result = await dynamodb.get(params).promise();
  return result;
};

// If throttled (429), invocation fails
// Async invocations: Auto-retried (exponential backoff)
// Sync invocations (API Gateway): Returns 502 to client

// ✓ Good - Exponential backoff retry
const retry = require('async-retry');

exports.handler = async (event) => {
  const dynamodb = new AWS.DynamoDB.DocumentClient();

  const result = await retry(async (bail) => {
    try {
      return await dynamodb.get(params).promise();
    } catch (error) {
      if (error.code === 'ProvisionedThroughputExceededException') {
        throw error;  // Retry
      }
      bail(error);  // Don't retry
    }
  }, {
    retries: 5,
    factor: 2,
    minTimeout: 100,
    maxTimeout: 5000
  });

  return result;
};
```

### Mistake 2: Not Monitoring Concurrency

```yaml
# ❌ Bad - No concurrency monitoring
# Function hitting limit, users getting 429s
# No alert

# ✓ Good - Monitor and alert
ConcurrencyAlarm:
  Type: AWS::CloudWatch::Alarm
  Properties:
    AlarmName: lambda-high-concurrency
    MetricName: ConcurrentExecutions
    Namespace: AWS/Lambda
    Dimensions:
      - Name: FunctionName
        Value: !Ref Function
    Statistic: Maximum
    Period: 60
    EvaluationPeriods: 1
    Threshold: 800  # Alert at 80% of 1,000 limit
    ComparisonOperator: GreaterThanThreshold

ThrottlesAlarm:
  Type: AWS::CloudWatch::Alarm
  Properties:
    AlarmName: lambda-throttles
    MetricName: Throttles
    Namespace: AWS/Lambda
    Dimensions:
      - Name: FunctionName
        Value: !Ref Function
    Statistic: Sum
    Period: 300
    EvaluationPeriods: 1
    Threshold: 10  # Alert on any throttles
    ComparisonOperator: GreaterThanThreshold
```

### Mistake 3: VPC Lambda Without NAT Gateway

```yaml
# ❌ Bad - VPC Lambda in public subnet (doesn't work)
Function:
  VpcConfig:
    SubnetIds:
      - subnet-public-1  # ❌ Lambda doesn't get public IP
    SecurityGroupIds:
      - sg-123

# Lambda can't reach internet or AWS services
# Invocations timeout

# ✓ Good - Private subnet with NAT Gateway
Function:
  VpcConfig:
    SubnetIds:
      - subnet-private-1  # ✓ Private subnet
      - subnet-private-2
    SecurityGroupIds:
      - sg-123

# NAT Gateway in public subnet routes traffic to internet
# VPC Endpoints for S3, DynamoDB (avoid NAT costs)
```

### Mistake 4: Not Using Batch Processing

```yaml
# ❌ Bad - Process SQS messages one at a time
EventSourceMapping:
  BatchSize: 1  # Process 1 message per invocation

# 1,000 messages = 1,000 Lambda invocations
# Invocation cost: 1,000 × $0.0000002 = $0.0002
# High overhead (cold starts, initialization)

# ✓ Good - Batch processing
EventSourceMapping:
  BatchSize: 10  # Process 10 messages per invocation

# 1,000 messages = 100 Lambda invocations
# Invocation cost: 100 × $0.0000002 = $0.00002 (10× cheaper)
# Fewer cold starts, better throughput
```

## Hands-On Exercise 1: Design Lambda Scaling Strategy

**Scenario:** Real-time image processing API

**Requirements:**
- Process uploaded images (resize, optimize, generate thumbnails)
- Uploads: 1,000/sec peak, 100/sec off-peak
- Processing time: 2 seconds per image
- Latency SLA: <500ms to start processing (can't tolerate 2-sec cold starts)
- Images stored in S3, trigger Lambda on upload

**Constraints:**
- Account concurrency limit: 1,000
- Budget: Optimize cost while meeting SLA

Design the Lambda configuration (memory, concurrency, triggers, cost).

<details>
<summary>Solution</summary>

**Analysis:**

```
Peak load:
  1,000 uploads/sec × 2 sec processing = 2,000 concurrent executions
  Problem: Exceeds 1,000 account limit

Off-peak load:
  100 uploads/sec × 2 sec = 200 concurrent executions
  No problem: Within limit
```

**Solution: Decouple with SQS**

```yaml
# ✓ S3 → SQS → Lambda (buffering)
S3Bucket:
  NotificationConfiguration:
    QueueConfigurations:
      - Event: s3:ObjectCreated:*
        Queue: !GetAtt ProcessingQueue.Arn

ProcessingQueue:
  Type: AWS::SQS::Queue
  Properties:
    VisibilityTimeout: 30  # 2× function timeout
    MessageRetentionPeriod: 86400  # 24 hours

Function:
  Type: AWS::Lambda::Function
  Properties:
    Runtime: nodejs18.x
    Handler: index.handler
    MemorySize: 2048  # More memory = faster processing
    Timeout: 15
    # ✓ Reserve 900 concurrent executions
    ReservedConcurrentExecutions: 900

EventSourceMapping:
  Type: AWS::Lambda::EventSourceMapping
  Properties:
    FunctionName: !Ref Function
    EventSourceArn: !GetAtt ProcessingQueue.Arn
    BatchSize: 1  # Process 1 image per invocation (CPU-intensive)
    ScalingConfig:
      MaximumConcurrency: 900  # Match reserved concurrency

# ✓ Provisioned concurrency for SLA
ProvisionedConcurrencyConfig:
  ProvisionedConcurrentExecutions: 200  # Handles off-peak without cold starts
  # Auto-scales to 900 during peak

AutoScaling:
  Type: AWS::ApplicationAutoScaling::ScalingPolicy
  Properties:
    PolicyType: TargetTrackingScaling
    TargetTrackingScalingPolicyConfiguration:
      PredefinedMetricSpecification:
        PredefinedMetricType: LambdaProvisionedConcurrencyUtilization
      TargetValue: 0.7  # Keep 70% utilized
```

**Traffic flow:**

```
Peak (1,000 uploads/sec):
  S3 → SQS (buffers 1,000 messages/sec)
  Lambda: 900 concurrent executions (processes 450 images/sec)
  Queue depth: Grows initially, processes over ~2-3 min

Off-peak (100 uploads/sec):
  S3 → SQS
  Lambda: 200 provisioned (processes 100 images/sec)
  Queue depth: 0 (keeps up with traffic)
  No cold starts (provisioned concurrency)
```

**Cost calculation:**

```
Provisioned concurrency (200):
  200 × 2 GB × 730 hours = 292,000 GB-hours
  292,000 × $0.000015 = $4,380/month

On-demand invocations:
  Peak: 1,000 uploads/sec × 3,600 sec × 8 hours × 22 days = 634M invocations
  Off-peak: 100 uploads/sec × 3,600 × 16 × 22 = 127M invocations
  Total: 761M invocations

  Invocation cost: 761M × $0.0000002 = $152
  Duration: 761M × 2 sec × 2 GB × $0.0000166667 = $50,667

  Total: $55,199/month

Alternative (no provisioned, accept cold starts):
  Duration cost: Same $50,667
  No provisioned cost: $0
  Total: $50,819/month

Savings with provisioned: $0 (actually costs MORE)
But: Meets SLA (<500ms, no 2-sec cold starts)
```

**Optimization:**

```yaml
# ✓ Reduce provisioned concurrency (only for SLA)
ProvisionedConcurrentExecutions: 50  # Minimum to meet SLA
# Scale on-demand for rest
# Cost: 50 × 2 GB × 730 = 73,000 GB-hours = $1,095/month
# Savings: $3,285/month vs 200 provisioned
```

</details>

## Hands-On Exercise 2: Debug Lambda Performance

**Problem:** Lambda function takes 5 seconds to process requests. Expected: <500ms.

**Function details:**
- Runtime: Node.js 18.x
- Memory: 128 MB
- Package size: 50 MB
- VPC-enabled
- Code:

```javascript
exports.handler = async (event) => {
  const AWS = require('aws-sdk');
  const dynamodb = new AWS.DynamoDB.DocumentClient();
  const axios = require('axios');

  const userId = event.pathParameters.userId;

  const user = await dynamodb.get({
    TableName: 'Users',
    Key: { userId }
  }).promise();

  const enrichedData = await axios.get(`https://api.external.com/enrich/${userId}`);

  return {
    statusCode: 200,
    body: JSON.stringify({
      user: user.Item,
      enriched: enrichedData.data
    })
  };
};
```

Identify all performance issues and fix them.

<details>
<summary>Solution</summary>

**Issues identified:**

1. ❌ `require()` inside handler (runs every invocation)
2. ❌ Low memory (128 MB = slow CPU)
3. ❌ Large package size (50 MB = slow download during cold start)
4. ❌ VPC-enabled but not needed (DynamoDB is public service)
5. ❌ No connection reuse (new HTTP connections every invocation)

**Fixed code:**

```javascript
// ✓ Move require() outside handler
const { DynamoDB } = require('@aws-sdk/client-dynamodb');
const { DynamoDBDocument } = require('@aws-sdk/lib-dynamodb');
const axios = require('axios');

// ✓ Initialize SDK clients outside handler (reused across invocations)
const client = new DynamoDB({});
const dynamodb = DynamoDBDocument.from(client);

// ✓ Configure HTTP keep-alive (reuse connections)
const http = require('http');
const https = require('https');
const httpAgent = new http.Agent({ keepAlive: true });
const httpsAgent = new https.Agent({ keepAlive: true });
const axiosInstance = axios.create({
  httpAgent,
  httpsAgent
});

exports.handler = async (event) => {
  const userId = event.pathParameters.userId;

  // ✓ Parallel requests (don't wait sequentially)
  const [user, enrichedData] = await Promise.all([
    dynamodb.get({
      TableName: process.env.TABLE_NAME,
      Key: { userId }
    }),
    axiosInstance.get(`https://api.external.com/enrich/${userId}`)
  ]);

  return {
    statusCode: 200,
    body: JSON.stringify({
      user: user.Item,
      enriched: enrichedData.data
    })
  };
};
```

**Infrastructure changes:**

```yaml
Function:
  Type: AWS::Lambda::Function
  Properties:
    Runtime: nodejs18.x
    Handler: index.handler
    # ✓ Increase memory (faster CPU, faster execution)
    MemorySize: 1024  # Was 128 MB
    Timeout: 10
    # ✓ Remove VPC (DynamoDB doesn't need VPC)
    # VpcConfig: REMOVED
    Environment:
      Variables:
        # ✓ Use environment variables
        TABLE_NAME: !Ref UsersTable
        # ✓ Enable HTTP keep-alive
        AWS_NODEJS_CONNECTION_REUSE_ENABLED: "1"

# ✓ Reduce package size
# package.json:
#   - Remove unused dependencies
#   - Use @aws-sdk/client-dynamodb (not aws-sdk)
#   - Bundle with esbuild/webpack (tree-shaking)

# Before: 50 MB → After: 5 MB (10× smaller)
```

**Performance results:**

```
Before:
  Cold start: 3,500ms
    - Download 50 MB package: 1,000ms
    - VPC ENI setup: 1,000ms (old behavior, no longer applies)
    - Initialize SDK: 500ms
    - Execute: 1,000ms (sequential requests, low memory)
  Warm execution: 1,000ms (sequential requests)

After:
  Cold start: 400ms
    - Download 5 MB package: 100ms
    - No VPC: 0ms
    - Initialize SDK (outside handler): 100ms
    - Execute: 200ms (parallel requests, higher memory)
  Warm execution: 200ms (parallel requests)

Improvement:
  Cold start: 8.75× faster (3,500ms → 400ms)
  Warm execution: 5× faster (1,000ms → 200ms)
```

**Cost impact:**

```
Before:
  1M invocations × 1 sec × 0.125 GB = 125,000 GB-sec
  125,000 × $0.0000166667 = $2.08

After:
  1M invocations × 0.2 sec × 1 GB = 200,000 GB-sec
  200,000 × $0.0000166667 = $3.33

Cost increase: $1.25/million invocations (60% more)
But: 5× faster, better user experience

Alternative: Test at 512 MB
  1M × 0.3 sec × 0.5 GB = 150,000 GB-sec = $2.50
  Still 3.3× faster, only 20% cost increase
```

</details>

## Interview Questions

### Q1: How does Lambda concurrency work, and what happens when you hit the limit?

Tests understanding of Lambda scaling and concurrency limits. Shows whether you've dealt with throttling in production.

<details>
<summary>Answer</summary>

**Concurrency = Invocations per second × Average duration (seconds)**

```
Example:
  1,000 req/sec × 0.5 sec avg = 500 concurrent executions
  5,000 req/sec × 2 sec avg = 10,000 concurrent executions
```

**Account limit:**
```
Default: 1,000 concurrent executions per region
Can request increase: Up to tens of thousands

Shared across ALL functions in region (unless reserved)
```

**When limit exceeded:**

```
Synchronous (API Gateway, ALB):
  - Returns 429 TooManyRequestsException
  - Client sees error immediately
  - Must retry

Asynchronous (S3, SNS, EventBridge):
  - Queued for retry (exponential backoff)
  - Retried up to 6 hours
  - Sent to DLQ if configured

Event source mappings (SQS, Kinesis):
  - Lambda stops polling
  - Retries with backoff
  - Messages remain in queue/stream
```

**Mitigation strategies:**

1. **Request limit increase**
   ```bash
   # Request via AWS Support
   # Can get 10,000+
   ```

2. **Reserved concurrency**
   ```yaml
   # Guarantee capacity for critical functions
   ReservedConcurrentExecutions: 500
   ```

3. **Queue-based buffering**
   ```
   API Gateway → SQS → Lambda
   # SQS buffers requests when Lambda throttled
   # Lambda processes when capacity available
   ```

4. **Optimize duration**
   ```javascript
   // Reduce function duration = lower concurrency
   // 5,000 req/sec × 0.5 sec = 2,500 concurrent
   // vs
   // 5,000 req/sec × 0.1 sec = 500 concurrent (5× lower)
   ```

**Monitoring:**

```bash
# CloudWatch metrics
ConcurrentExecutions: Current concurrent invocations
Throttles: Number of throttled invocations

# Alert on high concurrency
aws cloudwatch put-metric-alarm \
  --alarm-name high-concurrency \
  --metric-name ConcurrentExecutions \
  --threshold 800 \
  --comparison-operator GreaterThanThreshold
```

</details>

### Q2: What's the difference between reserved and provisioned concurrency?

Tests understanding of concurrency types and cost implications. Shows whether you've optimized Lambda for production workloads.

<details>
<summary>Answer</summary>

**Reserved Concurrency:**
- Dedicates capacity to a function
- Guarantees executions won't be throttled (up to limit)
- Limits maximum concurrent executions
- No additional cost
- Still subject to cold starts

```yaml
ReservedConcurrentExecutions: 500

Effect:
  - This function ALWAYS has 500 capacity
  - Other functions share remaining (1,000 - 500 = 500)
  - This function CANNOT exceed 500 (hard limit)
  - Cold starts still occur

Cost: No extra charge

Use when:
  - Guarantee critical function has capacity
  - Prevent one function from consuming all concurrency
  - Limit runaway function (cost protection)
```

**Provisioned Concurrency:**
- Pre-initializes execution environments
- Eliminates cold starts
- Always ready (no initialization delay)
- Additional cost (4× on-demand)

```yaml
ProvisionedConcurrentExecutions: 100

Effect:
  - 100 environments pre-initialized and kept warm
  - Zero cold starts for first 100 concurrent executions
  - Beyond 100: Uses on-demand (may have cold starts)

Cost:
  - $0.000015 per GB-second (provisioned)
  - Plus normal invocation and duration costs
  - ~4× more than on-demand

Use when:
  - Latency SLA <100ms (can't tolerate cold starts)
  - Predictable high traffic
  - Cost justified by business requirements
```

**Comparison:**

| Aspect              | Reserved                  | Provisioned                |
| ------------------- | ------------------------- | -------------------------- |
| **Purpose**         | Guarantee capacity        | Eliminate cold starts      |
| **Cold starts**     | Yes                       | No                         |
| **Cost**            | Free                      | ~4× on-demand              |
| **Use case**        | Critical function (SLA)   | Latency-sensitive          |
| **Limits others**   | Yes (reduces shared pool) | No                         |

**Can combine:**

```yaml
Function:
  # Reserve 1,000 capacity (guarantee)
  ReservedConcurrentExecutions: 1000

ProvisionedConcurrency:
  # Keep 200 warm (eliminate cold starts)
  ProvisionedConcurrentExecutions: 200

Effect:
  - First 200 invocations: Pre-warmed (zero cold start)
  - Next 800 invocations: On-demand (may have cold starts)
  - Beyond 1,000: Throttled
```

</details>

### Q3: How would you optimize a Lambda function with consistent 2-second cold starts?

Tests cold start optimization knowledge. Shows whether you understand initialization best practices and architecture patterns.

<details>
<summary>Answer</summary>

**Root cause analysis:**

```
2-second cold start breakdown:
  1. Download code package: 500ms (large package)
  2. Initialize runtime: 100ms
  3. Initialize handler code: 1,400ms (importing dependencies, connecting to services)

Total: 2,000ms
```

**Optimization strategies:**

**1. Minimize package size**

```bash
# Before: Include entire aws-sdk (50+ MB)
const AWS = require('aws-sdk');

# After: Import only needed clients (5 MB)
const { DynamoDB } = require('@aws-sdk/client-dynamodb');

# Bundle and tree-shake
npm install esbuild --save-dev
esbuild index.js --bundle --platform=node --target=node18 --outfile=dist/index.js

# Result: 50 MB → 5 MB (90% reduction)
# Cold start: 500ms → 100ms download time
```

**2. Move initialization outside handler**

```javascript
// ❌ Bad - Initialize on every invocation
exports.handler = async (event) => {
  const AWS = require('aws-sdk');  // Runs every time
  const db = new AWS.DynamoDB.DocumentClient();
  const result = await db.get(params).promise();
};

// ✓ Good - Initialize once (outside handler)
const { DynamoDB } = require('@aws-sdk/client-dynamodb');
const { DynamoDBDocument } = require('@aws-sdk/lib-dynamodb');

const client = new DynamoDB({});
const db = DynamoDBDocument.from(client);

exports.handler = async (event) => {
  // SDK client reused across invocations
  const result = await db.get(params);
};

// Result: 1,400ms → 200ms initialization (7× faster)
```

**3. Use Lambda Layers for shared dependencies**

```yaml
# Move common dependencies to layer
Layer:
  Type: AWS::Lambda::LayerVersion
  Properties:
    Content: layer.zip  # Contains node_modules/
    CompatibleRuntimes:
      - nodejs18.x

Function:
  Layers:
    - !Ref Layer
  # Function package: Only application code (< 1 MB)

# Benefits:
# - Smaller function package (faster download)
# - Shared across functions (downloaded once)
# - Reusable
```

**4. Lazy initialization**

```javascript
// ✓ Defer heavy initialization until needed
let db;

const getDB = () => {
  if (!db) {
    const { DynamoDB } = require('@aws-sdk/client-dynamodb');
    db = new DynamoDB({});
  }
  return db;
};

exports.handler = async (event) => {
  // Only initialize if this code path requires it
  if (event.operation === 'database') {
    const database = getDB();
    return database.query(...);
  }

  // Fast path (no DB initialization)
  return { statusCode: 200, body: 'OK' };
};
```

**5. Use Provisioned Concurrency**

```yaml
# ✓ Eliminate cold starts entirely
ProvisionedConcurrencyConfig:
  ProvisionedConcurrentExecutions: 10

# Cost: 10 × 1 GB × 730 hours = 7,300 GB-hours = $109/month
# Benefit: Zero cold starts

# Auto-scale provisioned concurrency
ScalingPolicy:
  TargetValue: 0.7  # Keep 70% utilized
  # Scales 10 → 50 during high traffic
```

**6. Keep functions warm (budget option)**

```yaml
# EventBridge pings function every 5 min
WarmupRule:
  ScheduleExpression: rate(5 minutes)
  Target: !GetAtt Function.Arn

# Cost: ~8,640 invocations/month ≈ free
# Keeps 1 environment warm (helps, but doesn't scale)
```

**7. Optimize runtime choice**

```
Cold start by runtime:
  Node.js: ~200ms
  Python: ~250ms
  Java: ~800ms (without SnapStart)
  .NET: ~1,000ms

# Use Node.js or Python for latency-sensitive workloads
# Use Java with SnapStart to reduce from 800ms → 200ms
```

**Results:**

```
Before optimizations:
  Cold start: 2,000ms
  Warm execution: 300ms

After optimizations:
  Package size: 50 MB → 5 MB
  Initialization: Moved outside handler
  Cold start: 400ms (5× faster)
  Warm execution: 300ms (unchanged)

With provisioned concurrency:
  Cold start: 0ms (eliminated)
  Cost: +$109/month
```

</details>

### Q4: When should you use Lambda in a VPC vs outside a VPC?

Tests VPC Lambda knowledge and architectural decision-making. Shows whether you understand the trade-offs and alternatives.

<details>
<summary>Answer</summary>

**Use Lambda in VPC when:**

1. **Access private VPC resources**
   ```
   ✓ RDS databases in private subnets
   ✓ ElastiCache clusters
   ✓ Private ALB/NLB endpoints
   ✓ EC2 instances not publicly accessible
   ```

2. **Compliance requirements**
   ```
   ✓ Data must stay within VPC
   ✓ Network isolation required
   ✓ Controlled egress (NAT Gateway firewall)
   ```

**Avoid Lambda in VPC when:**

1. **Only accessing public AWS services**
   ```
   ❌ DynamoDB (public endpoint, no VPC needed)
   ❌ S3 (public endpoint, no VPC needed)
   ❌ SNS, SQS, EventBridge (all public)

   # Use IAM for access control, not VPC
   ```

2. **Simplicity preferred**
   ```
   VPC Lambda adds complexity:
     - Need NAT Gateway ($32/month)
     - Need VPC Endpoints for AWS services ($7/month each)
     - More network configuration
     - More things to troubleshoot
   ```

**VPC Configuration:**

```yaml
Function:
  VpcConfig:
    # ✓ Use private subnets (not public)
    SubnetIds:
      - subnet-private-1
      - subnet-private-2
    SecurityGroupIds:
      - !Ref LambdaSecurityGroup

# ✓ NAT Gateway for internet access
NATGateway:
  Type: AWS::EC2::NatGateway
  Properties:
    SubnetId: subnet-public-1  # NAT in public subnet
    AllocationId: !GetAtt EIP.AllocationId

# ✓ VPC Endpoints to avoid NAT costs
S3Endpoint:
  Type: AWS::EC2::VPCEndpoint
  Properties:
    VpcId: !Ref VPC
    ServiceName: !Sub com.amazonaws.${AWS::Region}.s3
    RouteTableIds:
      - !Ref PrivateRouteTable

DynamoDBEndpoint:
  Type: AWS::EC2::VPCEndpoint
  Properties:
    ServiceName: !Sub com.amazonaws.${AWS::Region}.dynamodb
```

**Alternatives to VPC Lambda:**

1. **RDS Proxy (for databases)**
   ```yaml
   # ✓ Lambda outside VPC → RDS Proxy → RDS in VPC
   RDSProxy:
     Type: AWS::RDS::DBProxy
     Properties:
       DBProxyName: my-rds-proxy
       EngineFamily: POSTGRESQL
       Auth:
         - AuthScheme: SECRETS
           IAMAuth: REQUIRED

   # Benefits:
   # - Lambda outside VPC (simpler, no NAT Gateway)
   # - Connection pooling
   # - IAM authentication
   # - Still secure (proxy handles VPC access)
   ```

2. **API Gateway + VPC Link**
   ```
   Lambda (outside VPC) → API Gateway → VPC Link → ALB (in VPC)

   # Lambda doesn't need VPC access
   # Calls private ALB via API Gateway
   ```

3. **Use DynamoDB instead of RDS**
   ```
   # If possible, use DynamoDB (public endpoint)
   # No VPC needed
   # Serverless, auto-scaling
   # Often cheaper and simpler
   ```

**Cost comparison:**

```
Lambda outside VPC:
  - Lambda: $10/month
  - Total: $10/month

Lambda in VPC:
  - Lambda: $10/month
  - NAT Gateway: $32/month (+ data transfer)
  - VPC Endpoints (S3, DynamoDB): $14/month
  - Total: $56/month

5.6× more expensive for VPC
```

**Decision tree:**

```
Does Lambda need to access resources in VPC?
  └─ Yes → Is it RDS/ElastiCache?
      └─ Yes → Consider RDS Proxy (Lambda outside VPC)
      └─ No → Lambda in VPC

  └─ No → Lambda outside VPC (simpler, cheaper)
```

</details>

## Key Takeaways

1. **Concurrency Limits**: 1,000 default per region, request increase for high-scale workloads
2. **Reserved Concurrency**: Guarantees capacity (no extra cost), still has cold starts
3. **Provisioned Concurrency**: Eliminates cold starts (~4× cost), use for latency SLAs
4. **Cold Starts**: Minimize package size, initialize outside handler, use layers, consider provisioned
5. **VPC Lambda**: Use only when necessary (RDS, ElastiCache), consider alternatives (RDS Proxy)
6. **Event Source Mappings**: Batch processing reduces invocations (10× cost savings)
7. **Cost Optimization**: Right-size memory (512-1024 MB sweet spot), avoid over-invocation
8. **Monitoring**: Track concurrency, throttles, duration - set alarms before hitting limits

## Next Steps

In [Lesson 07: Queue-Based Decoupling](lesson-07-queue-based-decoupling.md), you'll learn:

- SQS vs SNS vs EventBridge - when to use each
- Handling traffic spikes with queues
- Dead letter queues and retry strategies
- FIFO vs Standard queues
- Backpressure handling patterns
- Fan-out architectures for scale
