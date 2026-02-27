# Lesson 04: CloudFront & Edge Optimization

Critical knowledge about AWS CloudFront for scaling APIs - cache behaviors, origin strategies, Lambda@Edge vs CloudFront Functions, and cost optimization through caching.

## The Edge Caching Problem

The interviewer asks: "Your API gets 100,000 requests per second globally. Users in Asia complain about 500ms latency. How do you fix this?" The answer isn't more servers - it's edge caching. Your API is in us-east-1. Asian users cross the Pacific for every request. CloudFront puts cache servers in 450+ edge locations worldwide. Requests hit the nearest edge, get cached responses in 20ms instead of 500ms. At scale, this also slashes your origin costs by 80-90%.

You're serving content globally. Every request from Europe to your US backend crosses the Atlantic (150ms+ latency). CloudFront is a CDN (Content Delivery Network) that caches content at edge locations near users. But it's not just for static files - modern CloudFront powers dynamic APIs with sophisticated caching, request transformation, and origin failover. Configure it right and you get 10x faster responses plus massive cost savings. Configure it wrong and you're paying for a cache that doesn't cache.

## CloudFront Architecture

```
User (Tokyo) → CloudFront Edge (Tokyo, 5ms)
                  ↓ (cache miss)
              CloudFront Origin (us-east-1, 150ms)
                  ↓
              Your API/ALB

Next request from Tokyo user:
User (Tokyo) → CloudFront Edge (Tokyo, 5ms) → Cache HIT, return immediately
(No origin request, no latency, no origin cost)
```

**Edge locations:** 450+ data centers worldwide (vs 30+ AWS regions)
**Regional edge caches:** Intermediate cache layer between edge and origin
**Origin:** Your source (S3, ALB, API Gateway, custom HTTP server)

## Cache Behaviors and Policies

Cache behaviors define HOW CloudFront caches content. Each behavior matches URL patterns and applies caching rules.

### Default Behavior vs Path Patterns

```yaml
CloudFrontDistribution:
  DefaultCacheBehavior:
    # Matches everything not caught by other behaviors
    TargetOriginId: MyALB
    CachePolicyId: CachingOptimized
    AllowedMethods: [GET, HEAD, OPTIONS, PUT, POST, PATCH, DELETE]
    CachedMethods: [GET, HEAD]

  CacheBehaviors:
    # ✓ Cache static files aggressively
    - PathPattern: /static/*
      TargetOriginId: S3StaticAssets
      CachePolicyId: CachingOptimized
      Compress: true
      MinTTL: 86400 # 24 hours
      DefaultTTL: 604800 # 7 days
      MaxTTL: 31536000 # 1 year

    # ✓ Cache API responses briefly
    - PathPattern: /api/*
      TargetOriginId: MyALB
      CachePolicyId: CachingOptimized
      MinTTL: 0
      DefaultTTL: 300 # 5 minutes
      MaxTTL: 3600 # 1 hour

    # ✓ Don't cache user-specific data
    - PathPattern: /api/users/profile
      TargetOriginId: MyALB
      CachePolicyId: CachingDisabled
      AllowedMethods: [GET, HEAD, OPTIONS, PUT, POST, PATCH, DELETE]
```

### Cache Key Components

CloudFront caches based on a cache key. Same key = cache hit. Different key = cache miss.

**Cache key includes:**

- URL path
- Query strings (optional, configurable)
- Headers (optional, configurable)
- Cookies (optional, configurable)

```yaml
# ❌ Wrong - Include all headers in cache key
CachePolicy:
  ParametersInCacheKeyAndForwardedToOrigin:
    EnableAcceptEncodingGzip: true
    HeadersConfig:
      HeaderBehavior: all  # ❌ Every unique header combo = new cache entry

# Every unique User-Agent = separate cache entry
# Chrome 120, Chrome 121, Firefox, Safari = 100+ cache entries
# Cache hit rate: <5%

# ✓ Correct - Minimal cache key
CachePolicy:
  ParametersInCacheKeyAndForwardedToOrigin:
    EnableAcceptEncodingGzip: true
    HeadersConfig:
      HeaderBehavior: none  # Don't include headers in cache key
    QueryStringsConfig:
      QueryStringBehavior: whitelist
      QueryStrings:
        - category  # Only include relevant query params
        - page

# Cache hit rate: 80%+
```

### Cache Policy Examples

```yaml
# ✓ CachingOptimized - Static assets
CachePolicy:
  Name: StaticAssetsCaching
  MinTTL: 1
  DefaultTTL: 86400  # 1 day
  MaxTTL: 31536000  # 1 year
  ParametersInCacheKeyAndForwardedToOrigin:
    EnableAcceptEncodingGzip: true
    EnableAcceptEncodingBrotli: true
    HeadersConfig:
      HeaderBehavior: none
    QueryStringsConfig:
      QueryStringBehavior: none
    CookiesConfig:
      CookieBehavior: none

# ✓ CachingOptimized with query strings - Product catalog
CachePolicy:
  Name: ProductCatalogCaching
  DefaultTTL: 300  # 5 minutes
  ParametersInCacheKeyAndForwardedToOrigin:
    QueryStringsConfig:
      QueryStringBehavior: whitelist
      QueryStrings:
        - category
        - sort
        - limit
    HeadersConfig:
      HeaderBehavior: whitelist
      Headers:
        - Accept-Language  # Cache per language

# ✓ CachingDisabled - User-specific data
CachePolicy:
  Name: NoCache
  MinTTL: 0
  DefaultTTL: 0
  MaxTTL: 0
```

### Origin Request Policy

Separate from cache policy - controls what CloudFront forwards to origin on cache miss.

```yaml
# ✓ Forward auth headers to origin (not in cache key)
OriginRequestPolicy:
  Name: ForwardAuthHeaders
  HeadersConfig:
    HeaderBehavior: whitelist
    Headers:
      - Authorization # Forward to origin, but don't cache based on it
      - CloudFront-Viewer-Country # Geo info
  QueryStringsConfig:
    QueryStringBehavior: all # Forward all query strings
  CookiesConfig:
    CookieBehavior: none

# Cache key: /api/products?category=electronics
# But forwards Authorization header to origin
# All authenticated users share same cache (if response is same)
```

## TTL Strategy

Time To Live (TTL) controls how long content stays cached.

### TTL Hierarchy

```
1. Cache-Control: max-age=3600 (origin response header)
2. Expires: Thu, 01 Jan 2026 00:00:00 GMT (origin response header)
3. CloudFront DefaultTTL (if no origin headers)

Priority: Origin headers > CloudFront configuration
```

### TTL Configuration

```yaml
# ✓ Static files - Long TTL
PathPattern: /static/*
MinTTL: 86400  # Minimum 1 day
DefaultTTL: 2592000  # Default 30 days
MaxTTL: 31536000  # Maximum 1 year

# Origin returns: Cache-Control: max-age=604800 (7 days)
# Cached for: 7 days (within min/max bounds)

# ✓ API responses - Short TTL
PathPattern: /api/*
MinTTL: 0
DefaultTTL: 300  # 5 minutes
MaxTTL: 3600  # 1 hour

# Origin returns: Cache-Control: max-age=600 (10 min)
# Cached for: 10 minutes (within max bound)

# ✓ Dynamic content - No cache
PathPattern: /api/user/profile
MinTTL: 0
DefaultTTL: 0
MaxTTL: 0
# Or origin returns: Cache-Control: no-cache, no-store
```

### Origin Cache-Control Headers

```javascript
// Backend sets cache headers
app.get("/api/products", (req, res) => {
  // ✓ Public, cacheable for 5 minutes
  res.set("Cache-Control", "public, max-age=300, s-maxage=300");
  res.json(products);
});

app.get("/api/users/profile", (req, res) => {
  // ✓ Private, don't cache
  res.set("Cache-Control", "private, no-cache, no-store, must-revalidate");
  res.json(userProfile);
});

app.get("/static/logo.png", (req, res) => {
  // ✓ Immutable, cache forever
  res.set("Cache-Control", "public, max-age=31536000, immutable");
  res.sendFile(logo);
});
```

**Cache-Control directives:**

- `public`: Cacheable by any cache (CDN, browser)
- `private`: Only cacheable by browser (not CDN)
- `no-cache`: Must revalidate with origin before using cached copy
- `no-store`: Never cache
- `max-age=N`: Cache for N seconds
- `s-maxage=N`: Cache for N seconds (CDN only, overrides max-age)
- `immutable`: Content never changes (skip revalidation)

## Cache Invalidation

You deployed new code. Cached content is stale. Users see old version for hours until TTL expires.

### Invalidation Methods

```bash
# ✓ Invalidate specific paths
aws cloudfront create-invalidation \
  --distribution-id E1234567890ABC \
  --paths "/api/products" "/api/categories"

# Cost: First 1,000 invalidations/month free, then $0.005 each

# ✓ Invalidate wildcard
aws cloudfront create-invalidation \
  --distribution-id E1234567890ABC \
  --paths "/api/*"

# Invalidates all /api/* paths (counts as 1 invalidation)

# ❌ Expensive - Invalidate everything
aws cloudfront create-invalidation \
  --distribution-id E1234567890ABC \
  --paths "/*"

# Invalidates entire distribution (use sparingly)
```

### Versioned URLs (Better Than Invalidation)

```javascript
// ❌ Wrong - Invalidate on every deployment
// static/app.js  (need to invalidate after each deploy)

// ✓ Correct - Versioned URLs
// static/app.v1.2.3.js
// static/app.123abc.js (git hash)
// static/app.js?v=1.2.3

// New version = new URL = automatic cache miss
// No invalidation needed
// Old version still cached (supports rollback)

app.get("/index.html", (req, res) => {
  res.send(`
    <script src="/static/app.${process.env.GIT_HASH}.js"></script>
  `);
});

// Cache headers:
// - /static/app.*.js → Cache-Control: max-age=31536000, immutable
// - /index.html → Cache-Control: max-age=300 (short TTL)
```

## Origin Failover

Your origin goes down. CloudFront keeps serving stale content from cache, but cache misses fail. Origin failover lets CloudFront try a secondary origin automatically.

### Origin Group Configuration

```yaml
OriginGroups:
  - Id: MyOriginGroup
    FailoverCriteria:
      StatusCodes:
        - 500
        - 502
        - 503
        - 504
        - 404 # Optional: treat 404 as failure
    Members:
      - OriginId: PrimaryOrigin
      - OriginId: SecondaryOrigin

Origins:
  # Primary origin (us-east-1)
  - Id: PrimaryOrigin
    DomainName: api-primary.example.com
    CustomOriginConfig:
      HTTPPort: 80
      HTTPSPort: 443
      OriginProtocolPolicy: https-only

  # Secondary origin (us-west-2)
  - Id: SecondaryOrigin
    DomainName: api-secondary.example.com
    CustomOriginConfig:
      HTTPPort: 80
      HTTPSPort: 443
      OriginProtocolPolicy: https-only

DefaultCacheBehavior:
  TargetOriginId: MyOriginGroup # Use origin group (not single origin)
```

**Failover behavior:**

```
Request → CloudFront edge (cache miss)
  ↓
Try PrimaryOrigin → 503 Service Unavailable
  ↓
Try SecondaryOrigin → 200 OK
  ↓
Cache response, return to user

Primary recovers → CloudFront automatically switches back on next cache miss
```

### Failover Best Practices

```yaml
# ✓ Add origin shield to reduce origin load
CacheBehavior:
  TargetOriginId: MyOriginGroup
  OriginShieldEnabled: true
  OriginShieldRegion: us-east-1

# ✓ Configure custom error pages
CustomErrorResponses:
  - ErrorCode: 503
    ResponseCode: 503
    ResponsePagePath: /errors/503.html
    ErrorCachingMinTTL: 10 # Don't cache errors long

  - ErrorCode: 500
    ResponseCode: 500
    ResponsePagePath: /errors/500.html
    ErrorCachingMinTTL: 0 # Don't cache at all
```

## Lambda@Edge vs CloudFront Functions

Both execute code at CloudFront edge locations. Different capabilities and cost.

### Feature Comparison

| Feature                | CloudFront Functions            | Lambda@Edge                       |
| ---------------------- | ------------------------------- | --------------------------------- |
| **Runtime**            | JavaScript (ECMAScript 5.1)     | Node.js, Python                   |
| **Execution location** | Edge locations (450+)           | Regional edge caches (13)         |
| **Max duration**       | <1 ms                           | 5 sec (viewer), 30 sec (origin)   |
| **Max memory**         | 2 MB                            | 128 MB - 10 GB                    |
| **Network access**     | ✗                               | ✓                                 |
| **File system access** | ✗                               | ✓ (/tmp)                          |
| **Environment vars**   | ✗                               | ✓                                 |
| **Triggers**           | Viewer request, viewer response | All 4 events                      |
| **Cost**               | $0.10 per 1M invocations        | $0.60 per 1M + $0.00005001/GB-sec |
| **Use case**           | Simple transforms, redirects    | Complex logic, API calls          |

### CloudFront Functions Use Cases

```javascript
// ✓ URL rewriting
function handler(event) {
  var request = event.request;
  var uri = request.uri;

  // Redirect /old-path to /new-path
  if (uri === "/old-path") {
    request.uri = "/new-path";
  }

  // Append index.html to directory requests
  if (uri.endsWith("/")) {
    request.uri += "index.html";
  }

  return request;
}

// ✓ Add security headers (viewer response)
function handler(event) {
  var response = event.response;
  response.headers = response.headers || {};

  response.headers["strict-transport-security"] = {
    value: "max-age=63072000; includeSubdomains; preload",
  };
  response.headers["x-content-type-options"] = {
    value: "nosniff",
  };
  response.headers["x-frame-options"] = {
    value: "DENY",
  };

  return response;
}

// ✓ A/B testing (simple)
function handler(event) {
  var request = event.request;
  var headers = request.headers;

  // Route 10% to variant B
  if (Math.random() < 0.1) {
    request.uri = request.uri.replace("/api/", "/api/variant-b/");
  }

  return request;
}
```

### Lambda@Edge Use Cases

```javascript
// ✓ Authentication at edge
exports.handler = async (event) => {
  const request = event.Records[0].cf.request;
  const headers = request.headers;

  const authHeader = headers.authorization?.[0]?.value;

  if (!authHeader) {
    return {
      status: "401",
      statusDescription: "Unauthorized",
      body: "Authentication required",
    };
  }

  // Verify JWT (can make external calls)
  try {
    const token = authHeader.replace("Bearer ", "");
    await verifyJWT(token); // ✓ Can call external services
    return request;
  } catch (err) {
    return {
      status: "403",
      statusDescription: "Forbidden",
      body: "Invalid token",
    };
  }
};

// ✓ Image resizing at edge
const AWS = require("aws-sdk");
const sharp = require("sharp");

exports.handler = async (event) => {
  const response = event.Records[0].cf.response;

  // Only process images
  if (
    response.status !== "200" ||
    !response.headers["content-type"]?.[0]?.value.startsWith("image/")
  ) {
    return response;
  }

  const width = event.Records[0].cf.request.querystring.split("=")[1];

  if (width) {
    const imageBuffer = Buffer.from(response.body, "base64");
    const resizedImage = await sharp(imageBuffer)
      .resize(parseInt(width))
      .toBuffer();

    response.body = resizedImage.toString("base64");
    response.bodyEncoding = "base64";
  }

  return response;
};
```

### Event Types

```
Viewer Request (before cache lookup)
  ↓
[CloudFront Cache]
  ↓ (cache miss)
Origin Request (before origin)
  ↓
[Origin]
  ↓
Origin Response (after origin, before caching)
  ↓
[CloudFront Cache] (response cached)
  ↓
Viewer Response (before returning to client)

CloudFront Functions: Viewer Request, Viewer Response only
Lambda@Edge: All 4 events
```

## Geographic Restrictions

Block or allow requests based on user location.

### Geo-Blocking Configuration

```yaml
Restrictions:
  GeoRestriction:
    # ✓ Whitelist - Only allow specific countries
    RestrictionType: whitelist
    Locations:
      - US
      - CA
      - GB
      - DE

    # Or blacklist - Block specific countries
    # RestrictionType: blacklist
    # Locations:
    #   - CN
    #   - RU
```

**Limitations:**

- Country-level only (not city/region)
- Based on IP geolocation (not 100% accurate)
- Returns 403 Forbidden for blocked countries

### Advanced Geo-Routing (Lambda@Edge)

```javascript
// ✓ Route by country to different origins
exports.handler = async (event) => {
  const request = event.Records[0].cf.request;
  const country = request.headers["cloudfront-viewer-country"]?.[0]?.value;

  // Route EU users to EU origin (GDPR compliance)
  if (["DE", "FR", "GB", "IT", "ES"].includes(country)) {
    request.origin.custom.domainName = "api-eu.example.com";
  }
  // Route Asia users to Asia origin
  else if (["JP", "CN", "IN", "SG"].includes(country)) {
    request.origin.custom.domainName = "api-asia.example.com";
  }

  return request;
};
```

## Signed URLs and Signed Cookies

Control access to private content without making origin publicly accessible.

### Signed URLs

```javascript
// Backend generates signed URL
const AWS = require("aws-sdk");
const cloudFront = new AWS.CloudFront.Signer(
  process.env.CLOUDFRONT_KEY_PAIR_ID,
  privateKey,
);

app.post("/request-video-access", (req, res) => {
  const videoKey = "videos/lesson-01.mp4";

  const signedUrl = cloudFront.getSignedUrl({
    url: `https://cdn.example.com/${videoKey}`,
    expires: Math.floor(Date.now() / 1000) + 3600, // 1 hour
  });

  res.json({ url: signedUrl });
});

// Client uses signed URL
const response = await fetch("/request-video-access", { method: "POST" });
const { url } = await response.json();

// url = https://cdn.example.com/videos/lesson-01.mp4?Expires=...&Signature=...&Key-Pair-Id=...
video.src = url;
```

### Signed Cookies (Multiple Files)

```javascript
// ✓ For multiple files (entire directory)
app.post("/request-course-access", (req, res) => {
  const policy = JSON.stringify({
    Statement: [
      {
        Resource: "https://cdn.example.com/courses/aws-course/*",
        Condition: {
          DateLessThan: {
            "AWS:EpochTime": Math.floor(Date.now() / 1000) + 86400, // 24 hours
          },
        },
      },
    ],
  });

  const signature = cloudFront.getSignedCookie({
    policy: policy,
  });

  // Set cookies
  res.cookie("CloudFront-Policy", signature["CloudFront-Policy"]);
  res.cookie("CloudFront-Signature", signature["CloudFront-Signature"]);
  res.cookie("CloudFront-Key-Pair-Id", signature["CloudFront-Key-Pair-Id"]);

  res.json({ message: "Access granted" });
});

// Client can now access any file under /courses/aws-course/*
// https://cdn.example.com/courses/aws-course/lesson-01.mp4
// https://cdn.example.com/courses/aws-course/lesson-02.mp4
```

## Cost Optimization

CloudFront costs: Data transfer out + HTTP/HTTPS requests

### Pricing (Simplified)

```
Data Transfer Out (per GB):
  First 10 TB: $0.085/GB
  Next 40 TB: $0.080/GB
  Over 150 TB: $0.060/GB

Requests:
  HTTP: $0.0075 per 10,000 requests
  HTTPS: $0.0100 per 10,000 requests

Field-Level Encryption: $0.02 per 10,000 requests
Invalidations: First 1,000/month free, then $0.005 each
```

### Cost Optimization Strategies

**1. Maximize Cache Hit Ratio**

```
Without caching:
  100M requests/month → 100M origin requests
  Origin cost (ALB + compute): $1,000/month

With 90% cache hit ratio:
  100M requests → 10M origin requests (90M cached)
  Origin cost: $100/month
  CloudFront cost: ~$100/month (caching + data transfer)
  Total: $200/month

Savings: $800/month (80%)
```

**2. Compress Content**

```yaml
CacheBehavior:
  Compress: true # ✓ Enable compression


# Before: 1 MB response = 1 MB data transfer = $0.085
# After (gzip): 200 KB response = 0.2 MB data transfer = $0.017
# Savings: 80% data transfer cost
```

**3. Use Origin Shield**

```yaml
# ✓ Reduces origin requests with additional cache layer
CacheBehavior:
  OriginShieldEnabled: true
  OriginShieldRegion: us-east-1

# Cost: $0.01/10,000 requests to Origin Shield
# Benefit: Reduces origin load (fewer ALB LCU charges)

# Example:
# Without Origin Shield:
#   Edge locations: 450
#   Cache miss at each edge: 450 origin requests
#   Origin cost: High (many connections)

# With Origin Shield:
#   Edge locations: 450 → Origin Shield: 1
#   Origin Shield → Origin: 1 request
#   Origin cost: Much lower
```

**4. Cache Error Responses**

```yaml
# ❌ Wrong - Don't cache errors
CustomErrorResponses:
  - ErrorCode: 500
    ErrorCachingMinTTL: 0  # Every request hits origin

# ✓ Correct - Cache errors briefly
CustomErrorResponses:
  - ErrorCode: 500
    ErrorCachingMinTTL: 10  # Cache 10 seconds
    ResponsePagePath: /errors/500.html

# Protects origin from repeated requests during outage
```

## Hands-On Exercise 1: Design Cache Strategy

You're launching an e-commerce API with these endpoints:

**Endpoints:**

- `GET /api/products` - Product catalog (updates hourly)
- `GET /api/products/{id}` - Single product (updates hourly)
- `GET /api/products?category=electronics&page=1` - Filtered products
- `POST /api/orders` - Create order (never cache)
- `GET /api/users/profile` - User profile (user-specific)
- `GET /static/images/*` - Product images (immutable)

**Requirements:**

- Global users (low latency critical)
- Product data changes hourly
- User profiles are private
- Images never change (versioned URLs)

Design CloudFront cache behaviors with TTLs and cache keys.

<details>
<summary>Solution</summary>

```yaml
CloudFrontDistribution:
  Origins:
    - Id: APIOrigin
      DomainName: api.example.com
    - Id: S3StaticAssets
      DomainName: static-assets.s3.amazonaws.com

  # ✓ Static images - Aggressive caching
  CacheBehaviors:
    - PathPattern: /static/images/*
      TargetOriginId: S3StaticAssets
      CachePolicyId: CachingOptimized
      MinTTL: 86400
      DefaultTTL: 31536000 # 1 year
      MaxTTL: 31536000
      Compress: true
      ParametersInCacheKeyAndForwardedToOrigin:
        EnableAcceptEncodingGzip: true
        EnableAcceptEncodingBrotli: true
        HeadersConfig:
          HeaderBehavior: none
        QueryStringsConfig:
          QueryStringBehavior: none
        CookiesConfig:
          CookieBehavior: none

    # ✓ Product catalog - Cache with query strings
    - PathPattern: /api/products*
      TargetOriginId: APIOrigin
      CachePolicyId: CustomCaching
      DefaultTTL: 3600 # 1 hour (matches update frequency)
      Compress: true
      AllowedMethods: [GET, HEAD, OPTIONS]
      CachedMethods: [GET, HEAD]
      ParametersInCacheKeyAndForwardedToOrigin:
        QueryStringsConfig:
          QueryStringBehavior: whitelist
          QueryStrings:
            - category
            - page
            - sort
        HeadersConfig:
          HeaderBehavior: none
        CookiesConfig:
          CookieBehavior: none

    # ✓ User profile - Don't cache (user-specific)
    - PathPattern: /api/users/profile
      TargetOriginId: APIOrigin
      CachePolicyId: CachingDisabled
      AllowedMethods: [GET, HEAD, OPTIONS, PUT, POST, PATCH, DELETE]
      OriginRequestPolicy:
        HeadersConfig:
          HeaderBehavior: whitelist
          Headers:
            - Authorization # Forward auth, but don't cache

    # ✓ Order creation - Don't cache
    - PathPattern: /api/orders
      TargetOriginId: APIOrigin
      CachePolicyId: CachingDisabled
      AllowedMethods: [POST, PUT, DELETE, GET, HEAD, OPTIONS]

  # Default behavior for everything else
  DefaultCacheBehavior:
    TargetOriginId: APIOrigin
    CachePolicyId: CachingOptimized
    DefaultTTL: 300 # 5 minutes
```

**Backend cache headers:**

```javascript
// Product endpoints
app.get("/api/products", (req, res) => {
  res.set("Cache-Control", "public, max-age=3600, s-maxage=3600");
  res.json(products);
});

// User profile (private, no cache)
app.get("/api/users/profile", (req, res) => {
  res.set("Cache-Control", "private, no-cache, no-store");
  res.json(userProfile);
});

// Static images (immutable)
// S3 bucket configured with: Cache-Control: public, max-age=31536000, immutable
```

**Expected cache hit ratio:**

- Static images: 99% (rarely change)
- Product catalog: 80-90% (popular products/categories)
- User profiles: 0% (not cached)
- Overall: 70-80% (depending on traffic mix)

</details>

## Hands-On Exercise 2: Calculate Cost Savings

**Current architecture (no CloudFront):**

- Traffic: 50,000 req/sec globally
- 80% GET requests (cacheable)
- Average response size: 50 KB
- Users distributed: 30% US, 30% Europe, 30% Asia, 10% other
- ALB in us-east-1
- Current monthly cost: $15,000 (ALB + data transfer + compute)

**Proposed: Add CloudFront**

Calculate the new cost and savings.

<details>
<summary>Solution</summary>

**Traffic Analysis:**

```
Total requests/month: 50K req/sec × 2.6M sec = 130B requests
Cacheable (80% GET): 104B requests
Non-cacheable (20%): 26B requests

Assume 85% cache hit ratio (well-optimized caching)
```

**Current Costs (No CloudFront):**

```
Data Transfer Out:
  130B requests × 50 KB = 6,500 TB/month
  From us-east-1 to:
    US (30%): 1,950 TB × $0.09/GB = $180,225
    Europe (30%): 1,950 TB × $0.09/GB = $180,225
    Asia (30%): 1,950 TB × $0.14/GB = $280,350
    Other (10%): 650 TB × $0.12/GB = $80,080
  Total data transfer: $720,880/month

ALB:
  Processed bytes: 6,500 TB/month = 270 TB/hour = 270 NLCU (limiting)
  Cost: $16 + (270 × $0.008 × 730) = $1,577/month

Compute (EC2/ECS): ~$5,000/month (estimate)

Total: $727,457/month
```

**With CloudFront:**

```
CloudFront Requests:
  Total: 130B requests
  HTTPS: 130B × $0.0100/10K = $130,000/month

CloudFront Data Transfer:
  Total served: 130B × 50 KB = 6,500 TB
  First 10 TB: $0.085/GB = $870
  Next 40 TB: $0.080/GB = $3,277
  Next 100 TB: $0.060/GB = $6,144
  Remaining 6,350 TB: $0.060/GB = $391,680
  Total: $401,971/month

Origin Requests (15% cache miss):
  26B non-cacheable + (104B × 0.15 cacheable misses) = 41.6B origin requests

Origin Data Transfer:
  41.6B × 50 KB = 2,080 TB
  All from CloudFront edge to origin (in same region): $0.02/GB = $42,496

ALB (reduced load):
  Processed bytes: 2,080 TB/month = 86.7 TB/hour = 87 NLCU
  Cost: $16 + (87 × $0.008 × 730) = $523/month

Compute (30% reduction due to caching): ~$3,500/month

Total with CloudFront:
  CloudFront: $531,971
  Origin transfer: $42,496
  ALB: $523
  Compute: $3,500
  Total: $578,490/month

Savings: $727,457 - $578,490 = $148,967/month (20% reduction)
```

**Better Savings with Compression:**

```
Enable compression (reduces transfer by ~70%):
  Response size: 50 KB → 15 KB (compressed)

CloudFront Data Transfer:
  6,500 TB × 0.3 (after compression) = 1,950 TB
  Cost: ~$120,000/month (vs $401,971)

Total with CloudFront + compression: ~$297,000/month

Savings: $727,457 - $297,000 = $430,457/month (59% reduction)
```

**Key takeaways:**

- CloudFront alone: 20% savings
- CloudFront + compression: 59% savings
- Additional benefits: Better global latency (500ms → 50ms for Asian users)

</details>

## Interview Questions

### Q1: How does CloudFront caching work, and what determines the cache key?

Tests understanding of caching fundamentals and cache key optimization. Shows whether you've actually tuned cache hit ratios in production or just used default settings.

<details>
<summary>Answer</summary>

**How caching works:**

```
1. Request arrives at CloudFront edge location
2. CloudFront generates cache key from:
   - URL path
   - Query strings (configurable)
   - Headers (configurable)
   - Cookies (configurable)
3. CloudFront checks cache using key
4. Cache HIT: Return cached response
5. Cache MISS: Forward to origin, cache response, return to user
```

**Cache key components:**

```yaml
CachePolicy:
  ParametersInCacheKeyAndForwardedToOrigin:
    # Query strings
    QueryStringsConfig:
      QueryStringBehavior: whitelist
      QueryStrings: [category, page] # Only include these

    # Headers
    HeadersConfig:
      HeaderBehavior: whitelist
      Headers: [Accept-Language] # Cache per language

    # Cookies
    CookiesConfig:
      CookieBehavior: none # Don't include cookies
```

**Cache key optimization:**

```
❌ Bad cache key (low hit rate):
  /api/products + ALL query strings + ALL headers + ALL cookies
  Example keys:
    - /api/products?category=electronics&tracking_id=abc123&timestamp=...
    - /api/products?category=electronics&tracking_id=def456&timestamp=...
  Result: Different cache entries for same content (tracking_id varies)
  Hit rate: <10%

✓ Good cache key (high hit rate):
  /api/products + [category, page] query strings only
  Example keys:
    - /api/products?category=electronics&page=1
    - /api/products?category=electronics&page=1  ← Cache HIT
  Result: Ignore tracking_id, timestamp (not in cache key)
  Hit rate: 80%+
```

**Separate cache key from origin request:**

```yaml
# Cache based on: /api/products?category=electronics
# But forward ALL query strings to origin (for analytics)
CachePolicy:
  QueryStringsConfig:
    QueryStringBehavior: whitelist
    QueryStrings: [category, page]

OriginRequestPolicy:
  QueryStringsConfig:
    QueryStringBehavior: all # Forward everything to origin
```

**Best practices:**

- Minimize cache key (only include what varies the response)
- Use whitelist (not all) for query strings/headers
- Don't include auth tokens in cache key (forward via OriginRequestPolicy)
- Monitor cache hit ratio (target 80%+)

</details>

### Q2: What's the difference between Lambda@Edge and CloudFront Functions, and when would you use each?

Tests understanding of edge compute options and cost optimization. Shows whether you know the performance and capability trade-offs.

<details>
<summary>Answer</summary>

**Key differences:**

| Aspect           | CloudFront Functions      | Lambda@Edge              |
| ---------------- | ------------------------- | ------------------------ |
| **Performance**  | <1 ms                     | 5-30 sec                 |
| **Scale**        | Millions req/sec          | Thousands req/sec        |
| **Execution**    | Every edge (450+)         | Regional edge (13)       |
| **Runtime**      | JS (ECMAScript 5.1)       | Node.js, Python          |
| **Capabilities** | Simple transforms         | Complex logic, API calls |
| **Network**      | ✗ No external calls       | ✓ Can call external APIs |
| **Cost**         | $0.10/million invocations | $0.60/million + compute  |
| **Max size**     | 10 KB                     | 50 MB                    |

**Use CloudFront Functions for:**

```javascript
// ✓ URL rewrites
function handler(event) {
  var request = event.request;
  request.uri = request.uri.replace(/^\/old/, "/new");
  return request;
}

// ✓ Add/modify headers
function handler(event) {
  var response = event.response;
  response.headers["x-custom-header"] = { value: "test" };
  return response;
}

// ✓ Simple redirects
function handler(event) {
  var request = event.request;
  if (request.uri === "/old-page") {
    return {
      statusCode: 301,
      statusDescription: "Moved Permanently",
      headers: { location: { value: "/new-page" } },
    };
  }
  return request;
}
```

**Use Lambda@Edge for:**

```javascript
// ✓ Authentication (external JWT verification)
exports.handler = async (event) => {
  const request = event.Records[0].cf.request;
  const token = request.headers.authorization?.[0]?.value;

  // Call external service to verify token
  const isValid = await verifyTokenWithAuthService(token);

  if (!isValid) {
    return { status: "403", body: "Forbidden" };
  }
  return request;
};

// ✓ Dynamic content generation
exports.handler = async (event) => {
  // Query database
  const data = await dynamodb
    .get({ TableName: "Products", Key: { id: "123" } })
    .promise();

  return {
    status: "200",
    body: JSON.stringify(data.Item),
  };
};

// ✓ Image processing
const sharp = require("sharp");
exports.handler = async (event) => {
  const response = event.Records[0].cf.response;
  const width = event.Records[0].cf.request.querystring.split("=")[1];

  const imageBuffer = Buffer.from(response.body, "base64");
  const resized = await sharp(imageBuffer).resize(parseInt(width)).toBuffer();

  response.body = resized.toString("base64");
  return response;
};
```

**Cost comparison:**

```
100M requests/month

CloudFront Functions:
  100M × $0.10/million = $10/month

Lambda@Edge:
  100M × $0.60/million = $60/month
  + Compute: 128MB × 100ms avg × 100M × $0.00005001/GB-sec = $64/month
  Total: $124/month

12× more expensive
```

**Decision tree:**

- Need external API calls? → Lambda@Edge
- Need packages (sharp, axios)? → Lambda@Edge
- Simple request/response transform? → CloudFront Functions
- Cost-sensitive + simple logic? → CloudFront Functions
- Need <1ms latency? → CloudFront Functions

</details>

### Q3: How would you design a caching strategy to achieve 80%+ cache hit ratio for an API?

Tests practical caching knowledge and optimization skills. Shows whether you understand the balance between cache TTL, cache key design, and business requirements.

<details>
<summary>Answer</summary>

**Strategy: Layered caching approach**

**1. Analyze traffic patterns**

```
Identify cacheable vs non-cacheable:
  ✓ Cacheable:
    - Product catalog (changes hourly)
    - Category pages (changes daily)
    - Public user profiles (changes on update)
    - Static assets (images, CSS, JS)

  ✗ Non-cacheable:
    - User-specific data (shopping cart, private profile)
    - POST/PUT/DELETE requests
    - Real-time data (stock prices, live scores)
    - Admin endpoints

Measure traffic distribution:
  If 80% requests are GET + public → high cache potential
  If 50% requests are user-specific → lower cache potential
```

**2. Optimize cache key**

```yaml
# ❌ Bad - Include everything
CachePolicy:
  QueryStringsConfig:
    QueryStringBehavior: all  # Every unique query = new cache entry
  HeadersConfig:
    HeaderBehavior: all  # Every User-Agent = new cache entry
  CookiesConfig:
    CookieBehavior: all  # Every cookie variation = new cache entry

Result: Cache hit rate <10%

# ✓ Good - Minimal cache key
CachePolicy:
  QueryStringsConfig:
    QueryStringBehavior: whitelist
    QueryStrings:
      - category  # Only include params that change response
      - page
      - sort
      # Exclude: tracking_id, session_id, timestamp, etc.

  HeadersConfig:
    HeaderBehavior: whitelist
    Headers:
      - Accept-Language  # Only if content varies by language
      # Exclude: User-Agent, Accept-Encoding (CloudFront handles)

  CookiesConfig:
    CookieBehavior: none  # Don't include cookies in cache key

Result: Cache hit rate 70-90%
```

**3. Set appropriate TTLs**

```yaml
# Static assets (immutable)
/static/*:
  TTL: 1 year
  Hit ratio: 99%

# Product catalog (updates hourly)
/api/products:
  TTL: 1 hour
  Hit ratio: 80-90%

# Homepage (updates every 5 min)
/:
  TTL: 5 minutes
  Hit ratio: 60-70%

# User-specific (don't cache)
/api/users/profile:
  TTL: 0
  Hit ratio: 0%
```

**4. Use Origin Shield**

```yaml
# Adds regional cache layer
OriginShieldEnabled: true
OriginShieldRegion: us-east-1

# Without Origin Shield:
#   450 edge locations × cache miss = 450 origin requests

# With Origin Shield:
#   450 edges → 1 Origin Shield → 1 origin request
#   Improves hit ratio by 5-10%
```

**5. Implement cache-friendly design**

```javascript
// ✓ Version static assets
// /static/app.v1.2.3.js (can cache forever)

// ✓ Normalize query parameters
// Sort params alphabetically (prevents duplicate cache entries)
const url = `/api/products?${new URLSearchParams(
  Object.entries(params).sort(),
)}`;

// ✓ Set proper Cache-Control headers
app.get("/api/products", (req, res) => {
  res.set("Cache-Control", "public, max-age=3600, s-maxage=3600");
  // s-maxage overrides max-age for CDN
});

// ✓ Use stale-while-revalidate
res.set("Cache-Control", "public, max-age=60, stale-while-revalidate=3600");
// Serve stale content while fetching fresh copy in background
```

**6. Monitor and optimize**

```bash
# CloudWatch metrics to monitor:
- CacheHitRate (target: 80%+)
- OriginLatency (should decrease with higher hit rate)
- 4xx/5xx errors (watch for cache-related issues)

# Calculate hit ratio:
CacheHitRate = (BytesDownloaded - BytesUploaded) / BytesDownloaded × 100%
```

**Expected results:**

```
Well-optimized API:
  Static assets: 99% hit rate
  Public API endpoints: 80-90% hit rate
  Overall: 85%+ hit rate

Benefits:
  - Origin load reduced 85%
  - Latency reduced (edge cache <50ms vs origin 200ms)
  - Cost reduced (fewer origin requests, less data transfer)
```

</details>

### Q4: Your CloudFront distribution has a 30% cache hit ratio. How would you troubleshoot and improve it?

Tests practical troubleshooting skills and understanding of common caching pitfalls. Shows whether you've actually debugged caching issues in production.

<details>
<summary>Answer</summary>

**Step 1: Analyze cache key configuration**

```bash
# Check cache policy
aws cloudfront get-cache-policy --id <policy-id>

# Common issues:
❌ QueryStringBehavior: all (every unique query creates new cache entry)
❌ HeaderBehavior: all (every User-Agent variation creates new cache entry)
❌ CookieBehavior: all (session cookies kill cache hit rate)
```

**Step 2: Review CloudWatch metrics**

```bash
# Detailed cache metrics by path
aws cloudwatch get-metric-statistics \
  --namespace AWS/CloudFront \
  --metric-name CacheHitRate \
  --dimensions Name=DistributionId,Value=<id> Name=Path,Value=/api/*

# Identify paths with low hit rate
# Example results:
#   /api/products: 80% hit rate ✓
#   /api/users: 5% hit rate ❌ (investigate why)
```

**Step 3: Analyze access logs**

```bash
# Enable access logs
aws cloudfront update-distribution --id <id> \
  --logging-config Enabled=true,Bucket=logs.s3.amazonaws.com,Prefix=cloudfront/

# Analyze logs
# Look for:
1. X-Cache header: Hit vs Miss vs RefreshHit
2. Query string patterns (are there random params?)
3. User-Agent variety (too many variations?)
4. Cookie patterns (session cookies in requests?)

# Query with Athena:
SELECT
  cs_uri_stem,
  COUNT(*) as requests,
  SUM(CASE WHEN x_edge_result_type = 'Hit' THEN 1 ELSE 0 END) as hits,
  SUM(CASE WHEN x_edge_result_type = 'Hit' THEN 1 ELSE 0 END) * 100.0 / COUNT(*) as hit_rate
FROM cloudfront_logs
GROUP BY cs_uri_stem
ORDER BY requests DESC;
```

**Step 4: Common problems and fixes**

**Problem 1: Random query parameters**

```bash
# Bad: /api/products?category=electronics&_=1234567890&session=abc123
# Different on every request (timestamp, session ID)

# Fix: Whitelist only meaningful parameters
CachePolicy:
  QueryStringsConfig:
    QueryStringBehavior: whitelist
    QueryStrings: [category, page, sort]
    # Exclude: _, session, tracking_id, timestamp
```

**Problem 2: Cookies in cache key**

```bash
# Bad: Caching based on session cookies
# Every user gets different cache entry

# Fix: Remove cookies from cache key
CachePolicy:
  CookiesConfig:
    CookieBehavior: none  # Don't cache based on cookies

# But still forward auth cookies to origin:
OriginRequestPolicy:
  CookiesConfig:
    CookieBehavior: whitelist
    Cookies: [session_id, auth_token]
```

**Problem 3: Vary header**

```javascript
// Backend returns Vary header
res.set("Vary", "User-Agent, Accept-Encoding");

// CloudFront creates separate cache entry for each User-Agent
// Hundreds of variations → low hit rate

// Fix: Remove Vary or be specific
res.set("Vary", "Accept-Encoding"); // Only vary on encoding
// Or remove entirely if content doesn't vary
```

**Problem 4: Short TTL**

```javascript
// Bad: TTL too short
res.set("Cache-Control", "max-age=10"); // 10 seconds

// Cache expires quickly → frequent origin requests → low hit rate

// Fix: Longer TTL based on update frequency
res.set("Cache-Control", "public, max-age=3600"); // 1 hour
// Use stale-while-revalidate for best of both worlds
res.set("Cache-Control", "public, max-age=300, stale-while-revalidate=3600");
```

**Problem 5: Non-cacheable responses**

```javascript
// Backend sends Cache-Control: no-cache
res.set("Cache-Control", "no-cache, no-store, must-revalidate");

// CloudFront respects this (doesn't cache)

// Fix: Only set no-cache for truly dynamic content
if (isUserSpecific) {
  res.set("Cache-Control", "private, no-cache");
} else {
  res.set("Cache-Control", "public, max-age=3600");
}
```

**Step 5: Implement fixes and monitor**

```yaml
# Before:
CacheHitRate: 30%
OriginRequests: 70M/month

# After optimizations:
CachePolicy:
  QueryStringsConfig: whitelist (category, page, sort only)
  HeadersConfig: none
  CookiesConfig: none
CacheTTL: Increased to 1 hour for public content

# Results:
CacheHitRate: 85%
OriginRequests: 15M/month (78% reduction)
```

**Monitoring:**

```bash
# Set CloudWatch alarm for cache hit rate
aws cloudwatch put-metric-alarm \
  --alarm-name low-cache-hit-rate \
  --metric-name CacheHitRate \
  --namespace AWS/CloudFront \
  --statistic Average \
  --period 3600 \
  --threshold 80 \
  --comparison-operator LessThanThreshold
```

</details>

## Key Takeaways

1. **Edge Caching**: CloudFront serves content from 450+ edge locations (vs 30+ AWS regions) for lower latency
2. **Cache Key**: Minimize cache key (path + selective query/headers) to maximize hit ratio (target 80%+)
3. **TTL Strategy**: Static assets (1 year), public API (5-60 min), user-specific (no cache)
4. **Cache Invalidation**: Use versioned URLs instead of invalidations (cheaper, enables rollback)
5. **Origin Failover**: Configure secondary origin for high availability (automatic failover on 5xx)
6. **Edge Compute**: CloudFront Functions (<1ms, $0.10/M) for simple transforms, Lambda@Edge for complex logic
7. **Cost**: 70-80% savings possible with caching + compression (6,500 TB → 1,950 TB compressed)
8. **Monitoring**: Track cache hit ratio, origin latency, analyze logs to optimize

## Next Steps

In [Lesson 05: Auto Scaling Patterns](lesson-05-auto-scaling-patterns.md), you'll learn:

- EC2 Auto Scaling vs ECS Service Auto Scaling
- Target tracking vs step scaling vs scheduled scaling
- Choosing the right scaling metrics
- Warm pools and lifecycle hooks
- Cool-down periods and scaling performance
- Common mistakes and anti-patterns
