# Lesson 02: API Gateway vs Application Load Balancer

Critical architectural decision - choosing between API Gateway and Application Load Balancer for your API, understanding when to use each, and when to combine them.

## The Fundamental Question

The interviewer asks: "You're building a REST API that needs to handle 50,000 requests per second. Should you use API Gateway or Application Load Balancer?" Most developers immediately pick API Gateway because "API" is in the name. But at high scale, the choice is more nuanced. API Gateway costs $3.50 per million requests. ALB costs $0.008 per hour plus LCU charges. At 50,000 req/sec (4.3 billion/month), that's $15,000/month for API Gateway vs potentially $200/month for ALB. Know when each makes sense.

Both API Gateway and ALB route HTTP requests to backends. Both support SSL termination, health checks, and autoscaling. But they're optimized for different use cases and have vastly different pricing models. Pick wrong and you're either overpaying by 50x or missing critical features.

## Feature Comparison

| Feature                      | API Gateway (REST)          | Application Load Balancer          |
| ---------------------------- | --------------------------- | ---------------------------------- |
| **Pricing model**            | Per-request ($3.50/million) | Hourly + LCU ($0.008/hr + usage)   |
| **Protocol**                 | HTTP/HTTPS only             | HTTP/HTTPS/HTTP2/gRPC              |
| **WebSocket**                | ✓ (separate API type)       | ✓ (native support)                 |
| **Request transformation**   | ✓ (VTL templates)           | ✗ (passthrough only)               |
| **Response caching**         | ✓ (built-in)                | ✗ (need CloudFront or app-level)   |
| **Request validation**       | ✓ (schema validation)       | ✗                                  |
| **Throttling/rate limiting** | ✓ (built-in)                | ✗ (need WAF or app-level)          |
| **API keys & usage plans**   | ✓                           | ✗                                  |
| **Auth**                     | IAM, Cognito, Lambda, JWT   | Cognito (redirect), OIDC           |
| **Custom domains**           | ✓ (via CloudFront/Route53)  | ✓ (direct)                         |
| **Static IP**                | ✗ (need CloudFront)         | ✗ (need NLB or Global Accelerator) |
| **Targets**                  | Lambda, HTTP, AWS services  | EC2, ECS, Lambda, IP addresses     |
| **VPC integration**          | VPC Link or Lambda in VPC   | Native (lives in VPC)              |
| **Max timeout**              | 29 seconds                  | Configurable (up to 4000 sec)      |
| **Max payload**              | 10 MB                       | No hard limit                      |
| **Request routing**          | Path/method/headers         | Path/host/headers/HTTP method      |
| **Health checks**            | ✓ (target-based)            | ✓ (highly configurable)            |
| **Logging**                  | CloudWatch (verbose)        | Access logs to S3 (less detail)    |
| **Metrics**                  | Detailed CloudWatch metrics | Basic CloudWatch metrics           |
| **Cold start**               | N/A (always hot)            | N/A (always hot)                   |
| **Deployment stages**        | ✓ (dev/staging/prod)        | ✗ (use different ALBs)             |
| **Canary deployments**       | ✓ (built-in)                | ✗ (use target groups)              |

## When to Use API Gateway

**Choose API Gateway when you need:**

### 1. Request/Response Transformations

You receive `{ "user_id": "123" }` but your backend expects `{ "userId": "123" }`. With API Gateway, VTL templates handle this. With ALB, you need to write transformation code in every backend or use Lambda.

```java
// ✓ API Gateway VTL transformation (no code)
{
  "userId": "$input.path('$.user_id')",
  "timestamp": $context.requestTimeEpoch
}

// ❌ ALB - need Lambda or app code for transformation
```

### 2. Built-in Throttling & Rate Limiting

You're building a public API and need to protect against abuse. API Gateway throttles at 10,000 req/sec by default and lets you set per-method limits. ALB has no throttling - every request reaches your backend.

```yaml
# ✓ API Gateway - built-in throttling
MethodSettings:
  - ResourcePath: /expensive-endpoint
    HttpMethod: POST
    ThrottlingRateLimit: 100 # 100 req/sec max


# ❌ ALB - need WAF (extra cost) or app-level rate limiting
# WAF cost: $5/month + $1 per million requests
```

### 3. API Keys & Usage Plans

You're monetizing your API with free/paid tiers. API Gateway has built-in API key management and quota enforcement. ALB has no concept of API keys.

```yaml
# ✓ API Gateway - native API key support
UsagePlan:
  Throttle:
    RateLimit: 1000
  Quota:
    Limit: 100000
    Period: MONTH

# ❌ ALB - build your own auth/quota system
```

### 4. Request Validation

You want to reject invalid requests before they hit your backend. API Gateway validates against JSON schemas. ALB passes everything through.

```yaml
# ✓ API Gateway - validates before backend
RequestValidatorId: !Ref Validator
RequestModels:
  application/json: !Ref UserModel

# Request missing required field → 400 Bad Request (never hits backend)

# ❌ ALB - backend must validate every request
```

### 5. Multiple Deployment Stages

You need dev, staging, and prod environments for the same API. API Gateway stages let you deploy the same API definition to multiple environments. ALB requires separate load balancers.

```yaml
# ✓ API Gateway - same API, multiple stages
POST https://api.example.com/dev/users
POST https://api.example.com/staging/users
POST https://api.example.com/prod/users
# ❌ ALB - separate ALBs for each environment
```

### 6. Direct AWS Service Integration

You need to write to SQS, query DynamoDB, or invoke Step Functions directly from HTTP endpoints. API Gateway integrates with AWS services natively. ALB always needs compute (Lambda/EC2/ECS).

```yaml
# ✓ API Gateway - direct SQS integration (no Lambda)
POST /events → API Gateway → SQS

# ❌ ALB - requires Lambda or EC2 proxy
POST /events → ALB → Lambda → SQS
```

## When to Use Application Load Balancer

**Choose ALB when you need:**

### 1. High Request Volume (Cost Sensitive)

You're handling millions of requests per day. API Gateway charges per request. ALB charges hourly + LCU (processing capacity).

```
50,000 requests/second = 4.3 billion requests/month

API Gateway (REST API):
  4,300,000,000 × $0.0000035 = $15,050/month

ALB:
  Base: $0.0225/hour × 730 hours = $16.43/month
  LCU: ~$0.008/LCU-hour × 730 × 25 LCU = $146/month
  Total: ~$162/month

Savings: $14,888/month (99% cheaper)
```

**Breakeven point:** ~30-50 requests/second sustained

### 2. Long-Running Requests (>29 seconds)

You have upload endpoints or long-polling connections. API Gateway times out at 29 seconds. ALB supports up to 4,000 seconds.

```yaml
# ❌ API Gateway - 29 second hard limit
User uploads 500 MB file → 29 seconds → 504 Gateway Timeout

# ✓ ALB - configurable timeout
TargetGroup:
  Properties:
    TargetGroupAttributes:
      - Key: deregistration_delay.timeout_seconds
        Value: "300"  # 5 minutes
```

### 3. Large Payloads (>10 MB)

You're accepting file uploads. API Gateway has a 10 MB payload limit. ALB has no hard limit.

```yaml
# ❌ API Gateway - 10 MB limit
POST /upload (50 MB file) → 413 Payload Too Large
# Must use S3 presigned URLs workaround

# ✓ ALB - no payload limit
POST /upload (50 MB file) → ALB → EC2 → Success
```

### 4. Native VPC Integration

Your API fronts private EC2/ECS services. ALB lives in your VPC and routes directly. API Gateway needs VPC Link (extra cost + complexity).

```yaml
# ❌ API Gateway - requires VPC Link
API Gateway → VPC Link ($0.01/hr) → NLB → EC2
# Extra hop, extra cost, extra latency

# ✓ ALB - native VPC integration
ALB → EC2 (same VPC, direct routing)
```

### 5. Container/EC2 Workloads

You're running containerized services on ECS/EKS or EC2 instances. ALB integrates natively with target groups and container auto-discovery. API Gateway requires extra layers.

```yaml
# ✓ ALB - native ECS integration
ALB → Target Group → ECS Service (dynamic port mapping)
# Auto-discovers containers, handles health checks

# ❌ API Gateway - need Lambda or VPC Link proxy
API Gateway → Lambda → ECS Service
# Or: API Gateway → VPC Link → NLB → ALB → ECS
```

### 6. WebSocket at Scale

You're building chat/real-time apps. API Gateway charges $1/million connections + $0.25/million messages. For high-message-volume apps, ALB WebSocket is cheaper (hourly charge only).

```yaml
# 1 million connected users, 100 messages/user/day = 100M messages/month

API Gateway WebSocket:
  Connections: $1.00 per million = $1/month
  Messages: $0.25 × 100 = $25/month
  Total: $26/month

ALB WebSocket:
  Hourly: $16.43/month
  LCU: ~$73/month (message processing)
  Total: ~$89/month

# API Gateway wins at low message volume
# But if 1,000 messages/user/day (100B messages):
#   API Gateway: $25,000/month
#   ALB: ~$200/month
```

## Cost Comparison Deep Dive

Understanding the cost model difference is critical for architectural decisions.

### API Gateway Pricing (REST API)

```
Cost = Requests × $3.50 per million

Examples:
  1,000 req/sec (2.6B/month):  $9,100/month
  10,000 req/sec (26B/month):  $91,000/month
  50,000 req/sec (130B/month): $455,000/month
```

**Predictable**: Scales linearly with traffic
**Problem**: Gets expensive fast at high scale

### ALB Pricing (LCU Model)

```
Cost = Hourly rate + LCU charges

Hourly: $0.0225/hour = $16.43/month
LCU: $0.008/LCU-hour (charged for highest dimension)

LCU dimensions:
  - New connections: 25/sec = 1 LCU
  - Active connections: 3,000 = 1 LCU
  - Processed bytes: 1 GB/hour = 1 LCU
  - Rule evaluations: 1,000/sec = 1 LCU

You pay for the MAXIMUM dimension
```

**Example calculation:**

```
Traffic: 10,000 req/sec, avg 5 KB response, avg 1 sec duration

New connections: 10,000/sec ÷ 25 = 400 LCU
Active connections: 10,000 concurrent ÷ 3,000 = 3.3 LCU
Processed bytes: 10,000 × 5 KB × 3,600 sec = 180 GB/hour = 180 LCU ← HIGHEST
Rule evaluations: 10,000/sec ÷ 1,000 = 10 LCU

Billed LCU: 180 (the maximum dimension)
Cost: $16.43 + (180 × $0.008 × 730) = $16.43 + $1,051 = $1,067/month

vs API Gateway: 26B req × $0.0000035 = $91,000/month

Savings: $89,933/month (98% cheaper)
```

### Breakeven Analysis

```
API Gateway = ALB when:
$3.50 per million requests = $16.43 base + LCU costs

Low traffic (minimal LCU charges):
  Breakeven ≈ 5,000 requests/month
  5,000 × $0.0000035 = $0.0175 ≈ ALB minimum

Medium traffic (assuming 10 LCU average):
  Breakeven ≈ 30-50 requests/second sustained
  Below this: API Gateway cheaper
  Above this: ALB cheaper
```

## Hybrid Architectures (Using Both)

Sometimes the best architecture uses BOTH services.

### Pattern 1: API Gateway + ALB (Layer Separation)

```
Internet → API Gateway (auth, throttling, caching) → ALB → EC2/ECS

✓ Use case: Public API with heavy traffic
  - API Gateway: Handles auth, rate limiting, caching
  - ALB: Handles load balancing to containers
  - Best of both worlds

Example:
  - API Gateway caches 90% of GET requests (reduces backend load)
  - API Gateway throttles abusive clients (protects ALB)
  - ALB distributes remaining 10% across ECS containers
```

### Pattern 2: CloudFront + ALB (Skip API Gateway)

```
Internet → CloudFront → ALB → EC2/ECS

✓ Use case: High-traffic API, simple routing
  - CloudFront: Caching, SSL, DDoS protection
  - ALB: Load balancing, health checks
  - Skip API Gateway entirely (save per-request costs)

Cost:
  50,000 req/sec:
    CloudFront: ~$500/month (caching + data transfer)
    ALB: ~$200/month
    Total: ~$700/month

  vs API Gateway alone: $15,000/month
  Savings: $14,300/month
```

### Pattern 3: API Gateway for Public, ALB for Internal

```
External clients → API Gateway → Lambda/DynamoDB
Internal services → ALB → ECS

✓ Use case: Microservices architecture
  - Public API: API Gateway (auth, throttling, managed)
  - Internal APIs: ALB (high throughput, low cost)
```

## Common Mistakes

### Mistake 1: Using API Gateway for High-Volume Simple Proxies

```yaml
# ❌ Wrong - Simple Lambda proxy at high scale
Architecture:
  - Traffic: 20,000 req/sec
  - API Gateway → Lambda → DynamoDB
  - Cost: $52,000/month for API Gateway

# ✓ Correct - ALB for high-volume simple routing
Architecture:
  - ALB → Lambda → DynamoDB
  - Cost: $300/month for ALB
  - Savings: $51,700/month
  - Tradeoff: No built-in throttling (add WAF if needed)
```

### Mistake 2: Using ALB When You Need API Features

```yaml
# ❌ Wrong - Reinventing API Gateway features
Architecture:
  - ALB → Lambda (for auth) → Lambda (for rate limiting) → Lambda (business logic)
  - Complexity: High
  - Latency: 3 Lambda cold starts
  - Cost: ALB + 3× Lambda invocations

# ✓ Correct - Use API Gateway's built-in features
Architecture:
  - API Gateway (auth, throttling) → Lambda (business logic)
  - Complexity: Low
  - Built-in features: Auth, throttling, validation
  - Cost: Reasonable for <10,000 req/sec
```

### Mistake 3: Not Considering Hybrid Approach

```yaml
# ❌ Wrong - All-or-nothing thinking
"We have high traffic, so we can't use API Gateway"

# ✓ Correct - Layer the services
CloudFront (caching, 80% hit rate) → API Gateway → ALB → ECS
  - 80% requests served from CloudFront cache (fast, cheap)
  - API Gateway handles 20% (auth, throttling, validation)
  - ALB distributes to containers
  - Best of all worlds
```

## Hands-On Exercise 1: Choose the Right Architecture

For each scenario, choose API Gateway, ALB, or a hybrid approach. Justify your choice with cost and feature considerations.

**Scenario 1: Public REST API**

- Traffic: 100 req/sec average, 500 req/sec peak
- Need: JWT auth, rate limiting (1,000 req/hour per user), response caching
- Backend: Lambda functions

**Scenario 2: Internal Microservices**

- Traffic: 50,000 req/sec between services
- Need: Load balancing, health checks, simple routing
- Backend: ECS containers on Fargate

**Scenario 3: File Upload API**

- Traffic: 1,000 req/day, uploads up to 500 MB
- Need: Auth, progress tracking
- Backend: EC2 instances processing uploads

**Scenario 4: Real-Time Chat API**

- Traffic: 100,000 WebSocket connections, 50 messages/user/day
- Need: Bi-directional messaging, presence
- Backend: ECS containers

**Scenario 5: Third-Party Developer API**

- Traffic: 10,000 req/sec
- Need: API keys, usage quotas (free tier: 1,000 req/day, paid tier: unlimited)
- Backend: Lambda + DynamoDB

<details>
<summary>Solution</summary>

**Scenario 1: API Gateway REST API** ✓

- Traffic: Low enough (100 req/sec = $9/month)
- Built-in JWT auth, rate limiting, caching
- Lambda integration (native)
- Cost: ~$10/month (API GW + Lambda)

**Scenario 2: ALB** ✓

- Traffic: High (50K req/sec = $455K/month for API GW vs $300/month for ALB)
- Internal traffic (no need for API Gateway features)
- ECS native integration
- Cost: ~$300/month

**Scenario 3: ALB** ✓

- Large payloads (500 MB exceeds API Gateway 10 MB limit)
- Long upload times (exceeds API Gateway 29 sec timeout)
- Low traffic (1,000 req/day = minimal ALB cost)
- Cost: ~$20/month (minimal LCU usage)

**Scenario 4: ALB WebSocket** ✓

- High message volume: 100K users × 50 msg/day = 5M msg/day
  - API Gateway: $1 (connections) + $37.50 (messages) = $38.50/month
  - ALB: ~$50/month
- Close in cost, but ALB scales better for high message volume
- If messages increase to 500/user/day:
  - API Gateway: $375/month
  - ALB: ~$80/month

**Scenario 5: API Gateway REST API** ✓

- API keys & usage plans ONLY available in API Gateway
- Monetization/quota enforcement required
- Worth higher cost for built-in features
- Traffic moderate (10K req/sec = $91K/month, but needed for business model)
- Alternative: ALB + custom auth system (complex, not worth it)

</details>

## Hands-On Exercise 2: Cost Optimization

You have this architecture running in production:

```
API Gateway (REST) → Lambda → DynamoDB
```

**Current traffic:**

- 30,000 requests/second sustained
- Average response: 2 KB
- 95% GET requests (read-heavy)
- Current cost: $273,000/month (API Gateway alone)

**Requirements:**

- Must maintain: Auth, rate limiting
- Must reduce cost significantly
- Can tolerate: Some added complexity

Design an optimized architecture and calculate the new cost.

<details>
<summary>Solution</summary>

**Optimized Architecture:**

```
CloudFront (caching) → ALB → Lambda → DynamoDB
    ↓
AWS WAF (rate limiting)
```

**Caching Strategy:**

- Cache GET requests at CloudFront (5-minute TTL)
- Assume 80% cache hit rate (typical for read-heavy APIs)

**New Traffic Flow:**

- 80% requests: CloudFront cache (never hit ALB)
- 20% requests: CloudFront → ALB → Lambda

**Cost Breakdown:**

```
CloudFront:
  Requests: 30K req/sec × 2.6M sec/month = 78B requests
  Cached (80%): 62.4B requests × $0.0000001 = $6,240/month
  Origin fetch (20%): 15.6B requests × $0.0000001 = $1,560/month
  Data transfer: 30K × 2 KB × 2.6M sec × 0.2 (origin) = 31 TB
    First 10 TB: $0.085/GB = $870
    Next 21 TB: $0.080/GB = $1,722
  CloudFront total: ~$10,392/month

ALB (20% of traffic = 6,000 req/sec):
  Hourly: $16.43/month
  LCU (processed bytes): 6K × 2KB × 3,600 = 43.2 GB/hr = 43 LCU
  LCU cost: 43 × $0.008 × 730 = $251/month
  ALB total: ~$267/month

AWS WAF (rate limiting):
  Base: $5/month
  Rules: $1/rule × 3 = $3/month
  Requests: 15.6B × $0.0000006 = $9,360/month
  WAF total: ~$9,368/month

Lambda (unchanged): ~$500/month (estimate)

DynamoDB (80% reduction due to caching): ~$200/month

TOTAL: $20,727/month

Original cost: $273,000/month (API Gateway alone)
New cost: $20,727/month
Savings: $252,273/month (92% reduction)
```

**Trade-offs:**

- ✓ 92% cost savings
- ✓ Better performance (CloudFront edge caching)
- ❌ More complex (3 services vs 1)
- ❌ No built-in API Gateway features (WAF for rate limiting, custom auth)
- ✓ Still have auth (ALB listener rules + Lambda authorizer)
- ✓ Still have rate limiting (WAF)

**Alternative (if complexity is an issue):**

```
CloudFront (caching) → API Gateway → Lambda

Cost:
  CloudFront: ~$10,392/month (same caching)
  API Gateway (20% traffic): 15.6B × $0.0000035 = $54,600/month
  Total: ~$65,000/month

Savings: $208,000/month (76% reduction)
Benefit: Simpler, keeps API Gateway features
```

</details>

## Interview Questions

### Q1: When would you choose ALB over API Gateway for a REST API?

This question tests cost awareness and architectural maturity. Many developers default to API Gateway because "REST API" is in the name, but at high scale, ALB is often better. Shows whether you understand pricing models and can make cost-based decisions.

<details>
<summary>Answer</summary>

**Choose ALB when:**

1. **High request volume** (>50 req/sec sustained)
   - API Gateway: $3.50 per million requests
   - ALB: Hourly + LCU (much cheaper at scale)
   - Example: 10K req/sec = $91K/month (API GW) vs $1K/month (ALB)

2. **Long-running requests** (>29 seconds)
   - API Gateway: Hard 29-second timeout
   - ALB: Up to 4,000 seconds configurable

3. **Large payloads** (>10 MB)
   - API Gateway: 10 MB hard limit
   - ALB: No practical limit

4. **Native VPC integration**
   - ALB lives in VPC, routes directly to targets
   - API Gateway needs VPC Link (extra cost/latency)

5. **Container/EC2 workloads**
   - ALB native ECS/EKS integration
   - API Gateway requires proxy layers

**Choose API Gateway when:**

- Need throttling, caching, request validation (built-in)
- Need API keys and usage plans
- Need request/response transformation
- Traffic is low-moderate (<50 req/sec)
- Want managed service with less operational overhead

**Cost breakeven:** ~30-50 req/sec sustained

</details>

### Q2: How would you design a cost-effective architecture for a high-traffic API (50,000 req/sec)?

Tests ability to design for scale and cost. Shows whether you can combine services strategically and understand caching strategies. The naive answer is "just use API Gateway" - the correct answer layers multiple services.

<details>
<summary>Answer</summary>

**Layered Architecture:**

```
Internet → CloudFront → ALB → ECS/Lambda → Database
            ↓
          WAF (optional)
```

**Why each layer:**

1. **CloudFront** (caching layer)
   - Cache GET requests (80-90% hit rate typical)
   - Edge locations (global distribution)
   - DDoS protection (AWS Shield Standard)
   - Reduces backend load by 80%+
   - Cost: ~$500/month (for 50K req/sec with caching)

2. **AWS WAF** (optional security layer)
   - Rate limiting (if needed)
   - SQL injection/XSS protection
   - Geographic blocking
   - Cost: ~$10K/month (for 50K req/sec)

3. **ALB** (load balancing)
   - Distributes 20% cache misses across targets
   - Health checks, auto-scaling integration
   - WebSocket support
   - Cost: ~$200/month (for 10K req/sec after caching)

4. **Compute** (ECS Fargate or Lambda)
   - ECS: Better for sustained high traffic
   - Lambda: Better for variable/spiky traffic
   - Cost: Depends on complexity

**Total cost: ~$700-1,000/month**

**vs API Gateway alone: $455,000/month**

**Savings: 99.8%**

**Trade-offs:**

- More operational complexity (3+ services)
- No built-in API Gateway features (must implement separately)
- Better performance (edge caching)
- Better cost at scale

**When to add API Gateway:**

- If you need API keys, usage plans, or complex transformations
- Put it between CloudFront and ALB: CloudFront → API Gateway → ALB
- Only pays for cache misses (~20% of traffic)

</details>

### Q3: Explain the LCU pricing model for ALB and how it differs from API Gateway pricing.

Tests deep understanding of AWS pricing models. Most developers know the basics but can't explain LCU dimensions. Shows whether you've actually worked with high-traffic ALBs and optimized costs.

<details>
<summary>Answer</summary>

**ALB LCU (Load Balancer Capacity Unit):**

You pay for the HIGHEST of these 4 dimensions:

1. **New connections/sec:** 25 connections = 1 LCU
2. **Active connections:** 3,000 connections = 1 LCU
3. **Processed bytes:** 1 GB/hour = 1 LCU
4. **Rule evaluations:** 1,000 evaluations/sec = 1 LCU

**Example:**

```
Traffic: 5,000 req/sec, 10 KB average response, 2 sec avg duration

New connections: 5,000/sec ÷ 25 = 200 LCU
Active connections: (5,000 × 2 sec) ÷ 3,000 = 3.3 LCU
Processed bytes: (5K × 10KB × 3,600) = 180 GB/hr = 180 LCU ← HIGHEST
Rule evaluations: 5,000/sec ÷ 1,000 = 5 LCU

Billed: 180 LCU
Cost: $16.43 (base) + (180 × $0.008 × 730) = $1,067/month
```

**API Gateway:**

```
Same traffic: 5,000 req/sec = 13B req/month
Cost: 13B × $0.0000035 = $45,500/month

43× more expensive
```

**Key differences:**

| Aspect             | ALB                        | API Gateway     |
| ------------------ | -------------------------- | --------------- |
| **Cost model**     | Capacity-based (LCU)       | Request-based   |
| **Predictability** | Less (varies by dimension) | High (linear)   |
| **At low scale**   | More expensive (minimum)   | Cheaper         |
| **At high scale**  | Much cheaper               | Expensive       |
| **Optimization**   | Reduce bytes, keep-alive   | Reduce requests |

**Optimization tips for ALB:**

- Use HTTP keep-alive (reduce new connections)
- Enable compression (reduce processed bytes) ← Usually the limiting dimension
- Minimize rule complexity (reduce evaluations)

</details>

### Q4: You have a public API that needs authentication, throttling, and costs $50,000/month in API Gateway charges. How would you optimize it?

Tests practical cost optimization skills. Shows whether you can analyze a real production scenario and propose concrete improvements with trade-off analysis.

<details>
<summary>Answer</summary>

**Step 1: Analyze current costs**

```
$50,000/month ÷ $3.50 per million = 14.3 billion requests/month
14.3B ÷ 2.6M seconds = ~5,500 req/sec average
```

**Step 2: Identify optimization opportunities**

1. **Add CloudFront caching** (if read-heavy)
2. **Switch to HTTP API** (if don't need REST features)
3. **Move to ALB + WAF** (if can handle complexity)
4. **Optimize cache hit rate** (if already using caching)

**Option 1: Add CloudFront Caching** (easiest)

```
Assumption: 70% GET requests, 80% cache hit rate

Architecture:
  CloudFront (caching) → API Gateway → Lambda

Requests hitting API Gateway: 14.3B × (1 - 0.8) = 2.86B/month
API Gateway cost: 2.86B × $0.0000035 = $10,010/month
CloudFront cost: ~$1,000/month (caching + data transfer)

Total: ~$11,000/month
Savings: $39,000/month (78%)

Pros: Simple, keeps API Gateway features
Cons: Only works if cacheable content
```

**Option 2: Switch to HTTP API** (if possible)

```
Check requirements:
  - Do you need response caching? (If yes, can't use HTTP API)
  - Do you need API keys/usage plans? (If yes, can't use HTTP API)
  - Do you need VTL transformations? (If yes, can't use HTTP API)

If NO to all:
  HTTP API cost: 14.3B × $0.000001 = $14,300/month
  Savings: $35,700/month (71%)

Pros: Simple migration, keeps most features
Cons: Loses caching, API keys, request validation
```

**Option 3: Migrate to ALB + WAF** (most savings)

```
Architecture:
  CloudFront (caching) → ALB → Lambda
  ↓
  WAF (throttling)

CloudFront: ~$1,000/month (80% cache hit)
ALB: ~$300/month (for 1,100 req/sec after caching)
WAF: ~$2,000/month (rate limiting rules)
Lambda: Unchanged

Total: ~$3,500/month
Savings: $46,500/month (93%)

Pros: Massive savings, better performance
Cons: More complex, need to rebuild auth/throttling
```

**Recommendation:**

Start with **Option 1** (CloudFront caching):

- Quick win, minimal changes
- $39K/month savings
- If still too expensive, then consider Option 3

**Implementation:**

```yaml
CloudFrontDistribution:
  Properties:
    DefaultCacheBehavior:
      TargetOriginId: APIGateway
      CachePolicyId: CachingOptimized
      AllowedMethods: [GET, HEAD, OPTIONS]
      CachedMethods: [GET, HEAD]
      Compress: true
      ViewerProtocolPolicy: redirect-to-https
```

</details>

## Key Takeaways

1. **Pricing Models**: API Gateway charges per-request ($3.50/million), ALB charges hourly + LCU (~$20-200/month base)
2. **Breakeven Point**: ~30-50 req/sec sustained - below: API Gateway cheaper, above: ALB cheaper
3. **API Gateway Best For**: Low-moderate traffic, need API features (throttling, caching, validation, API keys)
4. **ALB Best For**: High traffic, long requests (>29 sec), large payloads (>10 MB), container workloads
5. **Hybrid Approach**: CloudFront + API Gateway or CloudFront + ALB often optimal (caching reduces costs)
6. **Cost at Scale**: 50K req/sec = $455K/month (API GW) vs $200/month (ALB) - 2,000× difference
7. **WebSocket**: API Gateway charges per message, ALB charges hourly - high message volume favors ALB
8. **Don't Reinvent**: If you need API Gateway features with ALB, layer them (CloudFront → API GW → ALB)

## Next Steps

In [Lesson 03: Load Balancers (ALB vs NLB)](lesson-03-load-balancers-alb-nlb.md), you'll learn:

- Layer 7 vs Layer 4 load balancing
- When to use ALB vs NLB
- Target groups and health check strategies
- Path-based and host-based routing patterns
- Sticky sessions and connection handling
- Performance and scaling characteristics
