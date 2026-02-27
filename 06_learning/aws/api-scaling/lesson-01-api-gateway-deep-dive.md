# Lesson 01: API Gateway Deep Dive

Critical knowledge about AWS API Gateway for scaling APIs - choosing the right type, throttling strategies, caching, and integration patterns.

## REST API vs HTTP API vs WebSocket

The interviewer asks: "We need to build an API that handles 100,000 requests per second. Should we use API Gateway?" Your answer depends on understanding the three API Gateway types and when each makes sense. Pick REST API when you don't need its features and you're burning money. Pick HTTP API for simple proxies and you're leaving observability and security on the table. Know the trade-offs.

You're building an API. AWS offers three API Gateway types that look similar but have vastly different capabilities and price points. REST APIs have every feature but cost 3.5x more than HTTP APIs. WebSocket APIs handle bidirectional communication but can't do REST. Choosing wrong means either overpaying for unused features or missing critical capabilities you need later.

| Feature                         | REST API           | HTTP API            | WebSocket API              |
| ------------------------------- | ------------------ | ------------------- | -------------------------- |
| **Cost per million**            | $3.50              | $1.00               | $1.00 (+ messages)         |
| **Max timeout**                 | 29 seconds         | 30 seconds          | 2 hours (idle)             |
| **Auth**                        | All types          | JWT, IAM, Lambda    | IAM, Lambda                |
| **Request validation**          | ✓                  | ✗                   | ✗                          |
| **API keys**                    | ✓                  | ✗                   | ✗                          |
| **Usage plans**                 | ✓                  | ✗                   | ✗                          |
| **Response caching**            | ✓                  | ✗                   | ✗                          |
| **Request/response transforms** | ✓ (VTL)            | ✗                   | ✗                          |
| **Private APIs (VPC)**          | ✓                  | ✓                   | ✗                          |
| **CORS**                        | Manual config      | Auto-config         | N/A                        |
| **Protocol**                    | HTTP/HTTPS         | HTTP/HTTPS          | WebSocket (wss://)         |
| **Use case**                    | Full-featured APIs | Simple proxy/Lambda | Real-time, chat, streaming |

### When to Use Each

**REST API** - Choose when you need:

- Response caching to reduce backend load
- Request validation before hitting backend
- Usage plans and API keys for third-party developers
- Request/response transformations (modify payloads)
- Full CloudWatch metrics and detailed logging

**HTTP API** - Choose when you need:

- Simple Lambda proxy or HTTP proxy
- Lower cost (70% cheaper than REST)
- JWT authorization built-in
- CORS auto-configuration
- Better performance (lower latency)

**WebSocket API** - Choose when you need:

- Real-time bidirectional communication
- Chat applications, live dashboards
- Streaming data to connected clients
- Long-lived connections (up to 2 hours idle)

### Common Mistake: Using REST API for Everything

```yaml
# ❌ Wrong - REST API for simple Lambda proxy
Type: AWS::ApiGateway::RestApi
Properties:
  Name: SimpleLambdaProxy
  # Paying $3.50/million when HTTP API costs $1/million
  # Not using caching, validation, or any REST-only features

# ✓ Correct - HTTP API for simple proxies
Type: AWS::ApiGatewayV2::Api
Properties:
  Name: SimpleLambdaProxy
  ProtocolType: HTTP
  # 70% cost savings, better performance
```

## Throttling, Quotas, and Burst Limits

Your API just got featured on HackerNews. Requests spike from 100/sec to 10,000/sec. Without throttling, your database melts down, costs explode, and your API dies. With proper throttling, legitimate users get through, attackers get 429s, and your backend stays healthy.

API Gateway uses the **token bucket algorithm**. Think of it as a bucket that refills at a steady rate (your throttle limit) and can hold extra tokens (your burst capacity). Each request takes a token. If the bucket is empty, requests get 429 Too Many Requests. This protects your backend while allowing brief traffic spikes.

### Throttle Levels (Priority Order)

API Gateway applies throttles in this order (most specific wins):

| Level             | Scope                        | Default        | Purpose                          |
| ----------------- | ---------------------------- | -------------- | -------------------------------- |
| **Per-method**    | Specific method (GET /users) | None           | Protect expensive operations     |
| **Per-stage**     | Entire stage (prod, dev)     | 10,000 req/sec | Control overall stage throughput |
| **Account-level** | All APIs in region           | 10,000 req/sec | AWS-imposed regional limit       |
| **Usage plan**    | API key consumers            | Custom         | Third-party rate limiting        |

### Token Bucket Algorithm

```
Bucket capacity (burst): 5,000 tokens
Refill rate: 10,000 tokens/second

Second 0: 15,000 requests arrive
  - Use 5,000 tokens from bucket (bucket empty)
  - Use 10,000 tokens from refill
  - Throttle 5,000 requests (429 errors)

Second 1: 8,000 requests arrive
  - Bucket refilled with 10,000 tokens
  - Used 5,000 tokens from bucket
  - Use 8,000 tokens
  - Bucket now has 2,000 tokens
  - All requests succeed
```

### Throttling Configuration

```yaml
# CloudFormation - Method-level throttling
Resources:
  MyApi:
    Type: AWS::ApiGateway::RestApi
    Properties:
      Name: ThrottledAPI

  GetUsersMethod:
    Type: AWS::ApiGateway::Method
    Properties:
      # ✓ Correct - Protect expensive database queries
      HttpMethod: GET
      ResourceId: !Ref UsersResource
      RestApiId: !Ref MyApi
      MethodResponses:
        - StatusCode: 200
        - StatusCode: 429 # Too Many Requests

  ApiStage:
    Type: AWS::ApiGateway::Stage
    Properties:
      StageName: prod
      RestApiId: !Ref MyApi
      # Stage-level throttling
      ThrottlingBurstLimit: 5000 # Burst capacity
      ThrottlingRateLimit: 10000 # Steady-state req/sec
      MethodSettings:
        - ResourcePath: /users
          HttpMethod: GET
          # ✓ Override for expensive endpoint
          ThrottlingBurstLimit: 100
          ThrottlingRateLimit: 200
```

### Usage Plans for Third-Party APIs

You're building a public API for external developers. Free tier users get 1,000 requests/day. Paid users get 100,000/day. Enterprise users get unlimited. Without usage plans, you can't enforce these limits - everyone gets the same quota.

```yaml
# ✓ Correct - Tiered access with API keys
Resources:
  FreeUsagePlan:
    Type: AWS::ApiGateway::UsagePlan
    Properties:
      UsagePlanName: FreeTier
      Throttle:
        BurstLimit: 10
        RateLimit: 5 # 5 req/sec
      Quota:
        Limit: 1000
        Period: DAY
      ApiStages:
        - ApiId: !Ref MyApi
          Stage: !Ref ProdStage

  PaidUsagePlan:
    Type: AWS::ApiGateway::UsagePlan
    Properties:
      UsagePlanName: PaidTier
      Throttle:
        BurstLimit: 1000
        RateLimit: 500
      Quota:
        Limit: 100000
        Period: DAY
      ApiStages:
        - ApiId: !Ref MyApi
          Stage: !Ref ProdStage
```

## Caching Strategies

Your API queries a database that hasn't changed in 5 minutes. You just paid for 100,000 identical database queries. With caching enabled, you pay for 1 query and serve 99,999 responses from cache. But cache configuration is tricky - cache too long and users see stale data, cache too short and you waste money.

API Gateway caching stores responses at the stage level. Enable it and GET requests with the same parameters return cached responses instead of hitting your backend. This slashes backend load and response time. The cost is $0.02/hour for a 0.5GB cache - pennies compared to database costs.

### Cache Configuration

```yaml
# ✓ Correct - Cache GET requests for product catalog
ApiStage:
  Type: AWS::ApiGateway::Stage
  Properties:
    StageName: prod
    RestApiId: !Ref MyApi
    CacheClusterEnabled: true
    CacheClusterSize: "0.5" # 0.5GB, 1.6GB, 6.1GB, 13.5GB, 28.4GB, 58.2GB, 118GB, 237GB
    MethodSettings:
      - ResourcePath: /products/*
        HttpMethod: GET
        CachingEnabled: true
        CacheTtlInSeconds: 300 # 5 minutes
        CacheDataEncrypted: true
        # ✓ Cache based on query parameters
        CacheKeyParameters:
          - method.request.querystring.category
          - method.request.querystring.limit

      - ResourcePath: /orders
        HttpMethod: POST
        # ❌ Don't cache POST/PUT/DELETE
        CachingEnabled: false
```

### Cache Key Parameters

You have an endpoint `/products?category=electronics&limit=10`. Without cache key parameters, ALL requests to `/products` share the same cache entry - users requesting `category=books` get electronics. With cache key parameters, each unique combination gets its own cache entry.

```yaml
# ❌ Wrong - All /products requests share cache
CachingEnabled: true
# User 1: /products?category=electronics  → Cache MISS, query DB, cache result
# User 2: /products?category=books        → Cache HIT, returns electronics ❌

# ✓ Correct - Separate cache per category
CachingEnabled: true
CacheKeyParameters:
  - method.request.querystring.category
# User 1: /products?category=electronics  → Cache MISS (key: electronics)
# User 2: /products?category=books        → Cache MISS (key: books)
# User 3: /products?category=electronics  → Cache HIT (key: electronics) ✓
```

### Cache Invalidation

```bash
# ✓ Invalidate entire cache
aws apigateway flush-stage-cache \
  --rest-api-id abc123 \
  --stage-name prod

# ✓ Client can bypass cache (if allowed)
curl -H "Cache-Control: max-age=0" https://api.example.com/products
```

### Cost Consideration

```
Cache cost: 0.5GB cache = $0.02/hour = $14.40/month

Without cache:
  1M requests/day → 30M requests/month
  Backend: Lambda + DynamoDB = ~$20/month (compute + reads)

With cache (90% hit rate):
  3M requests hit backend (10%)
  27M served from cache
  Cost: $14.40 (cache) + $2 (backend) = $16.40/month
  Savings: $3.60/month + massive backend performance improvement
```

**When NOT to cache:**

- ❌ POST, PUT, DELETE, PATCH requests
- ❌ User-specific data (unless cache key includes user ID)
- ❌ Real-time data (stock prices, live scores)
- ❌ Small APIs with <1,000 requests/day (cache costs more than it saves)

## Integration Types

API Gateway doesn't host your application - it's a front door that routes requests somewhere. That "somewhere" is an integration. Choose wrong and you're writing boilerplate transformation code, paying for unnecessary Lambda invocations, or creating security holes.

| Integration      | Backend            | Use Case                       | Transforms | Cost              |
| ---------------- | ------------------ | ------------------------------ | ---------- | ----------------- |
| **Lambda**       | Lambda function    | Serverless APIs                | ✓ (VTL)    | $0 (API GW only)  |
| **Lambda Proxy** | Lambda function    | Passthrough (no transform)     | ✗          | $0 (API GW only)  |
| **HTTP**         | HTTP endpoint      | Proxy to external API/service  | ✓ (VTL)    | $0 (API GW only)  |
| **HTTP Proxy**   | HTTP endpoint      | Passthrough to HTTP            | ✗          | $0 (API GW only)  |
| **AWS Service**  | DynamoDB, SQS, etc | Direct AWS service integration | ✓ (VTL)    | $0 (API GW only)  |
| **VPC Link**     | Private resources  | Access private ALB/NLB         | ✓          | $0.01/hour + data |
| **Mock**         | None               | Return static response         | ✓          | $0 (API GW only)  |

### Lambda Proxy vs Lambda Integration

```javascript
// Lambda Proxy - API Gateway passes entire request
// ❌ Use when you DON'T need request transformation
// ✓ Use when you want full control in Lambda
exports.handler = async (event) => {
  // event = { httpMethod, path, queryStringParameters, headers, body, ... }
  return {
    statusCode: 200,
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ message: "Hello" }),
  };
};

// Lambda Integration - API Gateway transforms request/response
// ✓ Use when you want to modify requests before Lambda
// ❌ Use when you want simple passthrough (adds complexity)
exports.handler = async (event) => {
  // event = transformed by VTL mapping template
  return { message: "Hello" }; // Response also transformed by VTL
};
```

### AWS Service Integration (Skip Lambda)

You need to put messages in SQS. Most developers write a Lambda function that receives the API request and calls SQS. This works but adds latency, cost, and complexity. API Gateway can write to SQS directly - no Lambda needed.

```yaml
# ❌ Wrong - Unnecessary Lambda
API Gateway → Lambda → SQS
  - Extra latency (Lambda cold start)
  - Extra cost (Lambda invocation)
  - Extra code to maintain

# ✓ Correct - Direct integration
API Gateway → SQS
  - No Lambda needed
  - Lower latency
  - Lower cost
  - Less code

# CloudFormation
Method:
  Type: AWS::ApiGateway::Method
  Properties:
    HttpMethod: POST
    ResourceId: !Ref Resource
    RestApiId: !Ref Api
    Integration:
      Type: AWS
      IntegrationHttpMethod: POST
      Uri: !Sub 'arn:aws:apigateway:${AWS::Region}:sqs:path/${AWS::AccountId}/MyQueue'
      Credentials: !GetAtt ApiGatewayRole.Arn
      RequestParameters:
        integration.request.header.Content-Type: "'application/x-www-form-urlencoded'"
      RequestTemplates:
        application/json: |
          Action=SendMessage&MessageBody=$input.body
```

### VPC Link (Private Resources)

Your API needs to call a private service in a VPC (internal microservices on EC2/ECS behind a private ALB). API Gateway is public by default - it can't reach private resources. Options: (1) Make your internal service public (security risk), (2) Use Lambda in VPC as proxy (complex, slow cold starts), or (3) Use VPC Link to privately connect API Gateway to your internal NLB.

```yaml
# ✓ Correct - VPC Link to private NLB
VpcLink:
  Type: AWS::ApiGateway::VpcLink
  Properties:
    Name: PrivateServicesLink
    TargetArns:
      - !Ref PrivateNLB # Must be NLB (not ALB for REST API)

Method:
  Type: AWS::ApiGateway::Method
  Properties:
    Integration:
      Type: HTTP_PROXY
      ConnectionType: VPC_LINK
      ConnectionId: !Ref VpcLink
      Uri: http://internal-service.local/api
      IntegrationHttpMethod: ANY
```

**Cost**: $0.01/hour (~$7.20/month) + $0.01/GB data transfer

## Request/Response Transformations (VTL)

Your backend expects `{ "userId": "123" }` but your API receives `{ "user_id": "123" }`. Your backend returns XML but your client needs JSON. You could write Lambda functions to transform data, adding latency and cost. Or you can use VTL (Velocity Template Language) mapping templates in API Gateway to transform requests and responses without additional compute.

VTL is powerful but has a learning curve. For simple transformations, it's perfect. For complex logic, consider Lambda.

### Request Mapping Template

```java
// ✓ Transform request body before Lambda
{
  "userId": "$input.path('$.user_id')",
  "timestamp": $context.requestTimeEpoch,
  "sourceIp": "$context.identity.sourceIp",
  "body": $input.json('$')
}

// Input:  { "user_id": "123", "name": "Alice" }
// Output: {
//   "userId": "123",
//   "timestamp": 1677649200000,
//   "sourceIp": "192.168.1.1",
//   "body": { "user_id": "123", "name": "Alice" }
// }
```

### Response Mapping Template

```java
// ✓ Transform DynamoDB response to clean JSON
#set($items = $input.path('$.Items'))
{
  "users": [
    #foreach($item in $items)
    {
      "id": "$item.userId.S",
      "name": "$item.name.S",
      "email": "$item.email.S"
    }#if($foreach.hasNext),#end
    #end
  ],
  "count": $items.size()
}
```

### Common VTL Patterns

```java
// Access query parameters
$input.params('category')  // ?category=electronics

// Access headers
$input.params('Authorization')

// Access path parameters
$input.params('id')  // /users/{id}

// Parse JSON body
$input.path('$.user.email')

// Get entire body as JSON string
$input.json('$')

// Context variables
$context.requestId
$context.identity.userArn
$context.stage
```

## Common Limits and Gotchas

### Hard Limits (Cannot Be Increased)

| Limit                  | Value   | Gotcha                                    |
| ---------------------- | ------- | ----------------------------------------- |
| Timeout                | 29 sec  | Long-running processes must be async      |
| Payload size           | 10 MB   | Large uploads need S3 presigned URLs      |
| WebSocket idle timeout | 2 hours | Connection closes, clients must reconnect |
| WebSocket message size | 128 KB  | Chunk large messages                      |

### Soft Limits (Request Increase via Support)

| Limit                     | Default | Can Request |
| ------------------------- | ------- | ----------- |
| Regional API requests/sec | 10,000  | Up to 100K+ |
| Burst limit               | 5,000   | Up to 10K+  |
| WebSocket connections     | 100,000 | Higher      |

### Gotcha: 29-Second Timeout

```yaml
# ❌ Wrong - Long-running Lambda behind API Gateway
# User uploads video → API Gateway → Lambda processes video
# Lambda takes 2 minutes → API Gateway times out at 29 seconds → User sees 504 Gateway Timeout

# ✓ Correct - Async pattern
# User uploads video → API Gateway → Lambda (starts job, returns immediately)
#                                  → Lambda puts job in SQS
#                                  → Worker Lambda processes video
# User receives job ID, polls status endpoint
```

### Gotcha: 10 MB Payload Limit

```yaml
# ❌ Wrong - Upload 50 MB file through API Gateway
POST /upload
Content-Length: 52428800  # 50 MB
→ 413 Payload Too Large

# ✓ Correct - S3 presigned URL
1. Client requests upload URL: POST /request-upload
2. Lambda generates S3 presigned URL, returns to client
3. Client uploads directly to S3 (no API Gateway)
4. S3 triggers Lambda when upload completes
```

### Gotcha: CloudWatch Logs Cost

API Gateway can log every request. For high-traffic APIs, this generates massive CloudWatch Logs costs. A 1 KB log entry × 10M requests = 10 GB logs = $5/month. Sounds small, but logs are often 5-10 KB each, and costs scale with traffic.

```yaml
# ❌ Wrong - Full logging on high-traffic API
MethodSettings:
  - LoggingLevel: INFO  # Logs every request
  # 10M requests/month × 5 KB/log = 50 GB logs = $25/month

# ✓ Correct - Sample logging or errors only
MethodSettings:
  - LoggingLevel: ERROR  # Only log errors
  # Or use sampling: log 1% of requests
```

## Hands-On Exercise 1: Choose the Right API Type

You're architecting these systems. Choose REST API, HTTP API, or WebSocket API for each, and justify your choice:

1. **Public API for mobile app** - Simple CRUD operations, need JWT auth, 50,000 requests/day
2. **Third-party developer API** - Need API keys, usage quotas (1,000 req/day free, 100,000 req/day paid)
3. **Internal microservices** - Service-to-service calls, IAM auth, need low latency, 1M requests/day
4. **Real-time chat application** - Bi-directional messaging, presence indicators
5. **Product catalog API** - Mostly GET requests, data changes every 10 minutes, 500,000 requests/day

<details>
<summary>Solution</summary>

1. **HTTP API** ✓
   - JWT auth built-in
   - 70% cheaper ($1 vs $3.50 per million)
   - Don't need caching, validation, or usage plans for internal mobile app
   - Better performance

2. **REST API** ✓
   - API keys and usage plans ONLY available in REST API
   - Need quota enforcement (1,000 vs 100,000 req/day)
   - Worth the extra cost for third-party monetization

3. **HTTP API** ✓
   - IAM auth supported
   - Lower latency than REST API
   - 70% cost savings (1M requests = $1 vs $3.50)
   - Don't need REST-only features

4. **WebSocket API** ✓
   - Only option for bi-directional real-time communication
   - Maintains persistent connections for chat
   - Supports broadcasting to connected clients

5. **REST API** ✓
   - Response caching critical (data changes every 10 minutes)
   - 500K requests/day with 90% cache hit rate = 50K backend requests
   - Cache cost ($14/month) << backend cost savings
   - HTTP API doesn't support caching

</details>

## Hands-On Exercise 2: Design Throttling Strategy

You're building a REST API with these endpoints:

- `GET /products` - Returns product catalog (cheap, cached)
- `GET /products/{id}` - Returns single product (cheap, cached)
- `POST /orders` - Creates order (expensive, writes to DB, sends emails)
- `GET /analytics/report` - Generates analytics report (VERY expensive, 10-second query)

**Requirements:**

- Overall API limit: 10,000 req/sec
- Protect expensive endpoints
- Free tier users: 1,000 requests/day
- Paid tier users: 100,000 requests/day

Design the throttling configuration (stage-level, method-level, usage plans).

<details>
<summary>Solution</summary>

```yaml
Resources:
  # Stage-level throttling (applies to all endpoints)
  ProdStage:
    Type: AWS::ApiGateway::Stage
    Properties:
      StageName: prod
      ThrottlingBurstLimit: 5000
      ThrottlingRateLimit: 10000 # 10K req/sec overall
      MethodSettings:
        # ✓ Expensive endpoint - strict limits
        - ResourcePath: /analytics/report
          HttpMethod: GET
          ThrottlingBurstLimit: 50
          ThrottlingRateLimit: 10 # Only 10 req/sec

        # ✓ Write endpoint - moderate limits
        - ResourcePath: /orders
          HttpMethod: POST
          ThrottlingBurstLimit: 500
          ThrottlingRateLimit: 100 # 100 req/sec

        # ✓ Cached reads - default limits OK
        - ResourcePath: /products/*
          HttpMethod: GET
          CachingEnabled: true
          CacheTtlInSeconds: 600 # 10 minutes
          # Uses stage-level throttle (10K req/sec)

  # Usage plans for tiered access
  FreeTierPlan:
    Type: AWS::ApiGateway::UsagePlan
    Properties:
      Throttle:
        BurstLimit: 10
        RateLimit: 5 # 5 req/sec
      Quota:
        Limit: 1000
        Period: DAY

  PaidTierPlan:
    Type: AWS::ApiGateway::UsagePlan
    Properties:
      Throttle:
        BurstLimit: 1000
        RateLimit: 200 # 200 req/sec
      Quota:
        Limit: 100000
        Period: DAY
```

**Rationale:**

- Analytics endpoint: Strictest limits (10 req/sec) prevents DB overload
- Order creation: Moderate limits (100 req/sec) balances throughput with DB capacity
- Product reads: Cached, so high limits OK (most requests don't hit backend)
- Usage plans: Free tier gets basic access, paid tier gets higher limits
- Burst capacity: Allows temporary spikes without rejecting requests

</details>

## Interview Questions

### Q1: When would you choose HTTP API over REST API?

This question tests whether you understand cost-benefit analysis and service selection. Many developers default to REST API because it's the "original" option, but HTTP API is often better. The interviewer wants to see if you make deliberate architectural decisions based on requirements, not just familiarity.

<details>
<summary>Answer</summary>

**Choose HTTP API when:**

- Simple Lambda proxy or HTTP proxy (don't need transformations)
- Need JWT authorization (built-in for HTTP API)
- Cost-sensitive (70% cheaper: $1 vs $3.50 per million requests)
- Want lower latency (HTTP API is faster)
- CORS auto-configuration is valuable
- Don't need: caching, request validation, usage plans, API keys, or VTL transformations

**Choose REST API when:**

- Need response caching (HTTP API doesn't support)
- Need API keys and usage plans (third-party developer APIs)
- Need request validation before backend
- Need VTL request/response transformations
- Need detailed CloudWatch metrics

**Example:**

```
Internal microservices (1M req/day):
  - HTTP API: $1/million = $30/month
  - REST API: $3.50/million = $105/month
  - Savings: $75/month per service
  - Use HTTP API unless you need REST-only features
```

</details>

### Q2: How does API Gateway throttling work, and what happens when limits are exceeded?

Tests understanding of the token bucket algorithm and how to design resilient APIs. Shows whether you've dealt with rate limiting in production and understand how to protect backends from traffic spikes.

<details>
<summary>Answer</summary>

**Token Bucket Algorithm:**

- Bucket holds tokens (burst capacity)
- Tokens refill at steady rate (throttle limit)
- Each request consumes one token
- If bucket empty, request gets 429 Too Many Requests

**Throttle hierarchy (most specific wins):**

1. Method-level (specific endpoint)
2. Usage plan (per API key)
3. Stage-level (entire API)
4. Account-level (regional limit)

**When limits exceeded:**

- Client receives `429 Too Many Requests`
- Response includes `Retry-After` header
- Request NOT forwarded to backend (protects backend)
- Client should implement exponential backoff

**Best practice:**

```yaml
# Set aggressive limits on expensive endpoints
/analytics/report: 10 req/sec
/orders: 100 req/sec
/products: 10,000 req/sec (cached)
```

**Client retry logic:**

```javascript
async function callAPI(url, retries = 3) {
  try {
    return await fetch(url);
  } catch (err) {
    if (err.status === 429 && retries > 0) {
      await sleep(Math.pow(2, 3 - retries) * 1000); // Exponential backoff
      return callAPI(url, retries - 1);
    }
    throw err;
  }
}
```

</details>

### Q3: Explain when you would use direct AWS service integration vs Lambda integration.

Separates developers who reach for Lambda for everything versus those who understand API Gateway's full capabilities. Shows architectural maturity - knowing when to eliminate components rather than adding them.

<details>
<summary>Answer</summary>

**Use direct AWS service integration when:**

- Simple operations (SQS send, DynamoDB get/put)
- No business logic needed
- Want to eliminate Lambda (lower latency, lower cost, simpler)

**Use Lambda integration when:**

- Need business logic, validation, or transformation
- Complex orchestration across multiple services
- Need to handle errors and retries
- Integration requires conditional logic

**Examples:**

```yaml
# ✓ Direct integration - API Gateway → SQS
POST /events → SQS.SendMessage
  - No Lambda needed
  - Lower latency (no cold start)
  - Lower cost (no Lambda invocation)
  - Less code to maintain

# ✓ Lambda integration - Business logic required
POST /orders → Lambda → [Validate, DynamoDB, SQS, SNS, Email]
  - Lambda needed for orchestration
  - Complex multi-step workflow
  - Error handling and rollback logic
```

**Cost comparison (1M requests):**

```
Direct integration:
  API Gateway: $3.50
  SQS: $0.40
  Total: $3.90

With Lambda proxy:
  API Gateway: $3.50
  Lambda: $0.20
  SQS: $0.40
  Total: $4.10 (5% more expensive + higher latency)
```

</details>

### Q4: How would you handle file uploads larger than 10 MB with API Gateway?

Tests practical problem-solving and whether you know API Gateway's hard limits. The 10 MB payload limit is a common production issue. The correct answer shows you've actually built file upload systems at scale.

<details>
<summary>Answer</summary>

**Problem:** API Gateway has a hard 10 MB payload limit. Large file uploads fail with `413 Payload Too Large`.

**Solution: S3 Presigned URLs**

```
Client workflow:
1. Request upload URL: POST /request-upload → API Gateway → Lambda
2. Lambda generates S3 presigned URL (valid for 15 min)
3. Lambda returns presigned URL to client
4. Client uploads directly to S3 using presigned URL (bypasses API Gateway)
5. S3 triggers Lambda on upload completion (S3 event notification)
6. Processing Lambda validates, scans, processes file
```

**Implementation:**

```javascript
// Lambda: Generate presigned URL
const AWS = require("aws-sdk");
const s3 = new AWS.S3();

exports.handler = async (event) => {
  const { filename, contentType } = JSON.parse(event.body);

  const params = {
    Bucket: "my-uploads",
    Key: `uploads/${Date.now()}-${filename}`,
    ContentType: contentType,
    Expires: 900, // 15 minutes
  };

  const uploadUrl = await s3.getSignedUrlPromise("putObject", params);

  return {
    statusCode: 200,
    body: JSON.stringify({ uploadUrl, key: params.Key }),
  };
};

// Client: Upload to S3
const response = await fetch("/request-upload", {
  method: "POST",
  body: JSON.stringify({ filename: "video.mp4", contentType: "video/mp4" }),
});
const { uploadUrl } = await response.json();

await fetch(uploadUrl, {
  method: "PUT",
  body: fileData,
  headers: { "Content-Type": "video/mp4" },
});
```

**Benefits:**

- Supports files up to 5 TB (S3 limit)
- No API Gateway payload limit
- No Lambda timeout (upload goes directly to S3)
- Lower cost (no data transfer through API Gateway)
- Better performance (S3 optimized for large files)

</details>

## Key Takeaways

1. **API Types**: REST API for full features, HTTP API for cost/performance, WebSocket for real-time
2. **Cost**: HTTP API is 70% cheaper ($1 vs $3.50/million) - choose unless you need REST-only features
3. **Throttling**: Token bucket algorithm - set aggressive limits on expensive endpoints
4. **Caching**: REST API only - huge backend savings for read-heavy APIs (90% cache hit rate typical)
5. **Integrations**: Use direct AWS service integration (skip Lambda) for simple operations
6. **VTL Transformations**: Modify requests/responses without Lambda, but has learning curve
7. **Limits**: 29-second timeout, 10 MB payload - design around these hard limits
8. **Usage Plans**: REST API only - required for API keys and quotas

## Next Steps

In [Lesson 02: API Gateway vs ALB)](lesson-02-api-gateway-vs-alb.md), you'll learn:

- When to use ALB vs NLB (Layer 7 vs Layer 4)
- Target groups and health check strategies
- Connection handling and scaling behavior
- Path-based and host-based routing
- Sticky sessions and their performance impact
