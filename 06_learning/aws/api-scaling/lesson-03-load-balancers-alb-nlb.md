# Lesson 03: Load Balancers (ALB vs NLB)

Critical knowledge about Application Load Balancer vs Network Load Balancer - understanding Layer 7 vs Layer 4, choosing the right type, and configuring for scale and performance.

## Layer 7 vs Layer 4: The Fundamental Difference

The interviewer asks: "Should we use Application Load Balancer or Network Load Balancer?" Your answer depends on understanding the OSI model layers where each operates. ALB inspects HTTP headers and makes routing decisions based on URLs, headers, and methods. NLB forwards packets blindly based on IP and port. Choose wrong and you're either overpaying for features you don't need (ALB) or missing critical routing capabilities (NLB).

You're distributing traffic across backend servers. Both ALB and NLB spread requests, but they work at different network layers with vastly different capabilities and performance characteristics. ALB terminates connections, parses HTTP, and routes based on application logic. NLB passes packets through with minimal processing, achieving microsecond latency and millions of requests per second. The choice shapes your entire architecture.

## Feature Comparison

| Feature                       | Application Load Balancer (ALB)         | Network Load Balancer (NLB)              |
| ----------------------------- | --------------------------------------- | ---------------------------------------- |
| **OSI Layer**                 | Layer 7 (Application)                   | Layer 4 (Transport)                      |
| **Protocol**                  | HTTP/HTTPS/HTTP2/gRPC/WebSocket         | TCP/UDP/TLS                              |
| **Routing**                   | Path/host/header/method/query           | IP + port only                           |
| **Latency**                   | ~10-20 ms                               | <1 ms (microseconds)                     |
| **Throughput**                | ~100K req/sec per AZ                    | Millions req/sec per AZ                  |
| **Connection handling**       | Terminates connections (proxy)          | Passthrough (preserves source IP)        |
| **Static IP**                 | ✗ (DNS only)                            | ✓ (Elastic IP support)                   |
| **TLS termination**           | ✓ (ACM integration)                     | ✓ (ACM integration)                      |
| **WebSocket**                 | ✓                                       | ✓                                        |
| **HTTP/2**                    | ✓                                       | ✗                                        |
| **Content-based routing**     | ✓                                       | ✗                                        |
| **Host-based routing**        | ✓                                       | ✗                                        |
| **Path-based routing**        | ✓                                       | ✗                                        |
| **Redirect rules**            | ✓ (HTTP to HTTPS, custom)               | ✗                                        |
| **Fixed response**            | ✓ (return static content)               | ✗                                        |
| **Sticky sessions**           | ✓ (cookie-based)                        | ✓ (flow hash)                            |
| **Health checks**             | HTTP/HTTPS (checks response)            | TCP/HTTP/HTTPS (checks connection)       |
| **Targets**                   | EC2, IP, Lambda, ECS                    | EC2, IP, ECS (no Lambda)                 |
| **Cross-zone**                | ✓ (default on, no charge)               | ✓ (off by default, charges data transfer) |
| **Integration**               | WAF, Cognito, OIDC                      | ✗                                        |
| **PrivateLink**               | ✗                                       | ✓                                        |
| **Pricing**                   | Hourly + LCU                            | Hourly + NLCU (cheaper)                  |
| **Use case**                  | HTTP APIs, microservices, containers    | Extreme performance, static IP, non-HTTP |

## When to Use ALB

**Choose Application Load Balancer when you need:**

### 1. Content-Based Routing

You have multiple microservices behind one domain. Route `/api/users` to the user service, `/api/orders` to the order service, `/api/payments` to the payment service. ALB inspects the URL path and routes accordingly. NLB can't do this - it only sees IP and port.

```yaml
# ✓ ALB - Path-based routing
Listener:
  Rules:
    - Condition:
        PathPattern: /api/users/*
      Action:
        TargetGroup: UserServiceTG

    - Condition:
        PathPattern: /api/orders/*
      Action:
        TargetGroup: OrderServiceTG

    - Condition:
        PathPattern: /api/payments/*
      Action:
        TargetGroup: PaymentServiceTG

# ❌ NLB - Can't route by path (only IP:port)
# Would need separate NLBs or separate ports per service
```

### 2. Host-Based Routing (Multi-Tenancy)

You host multiple customer subdomains on the same infrastructure. Route `tenant1.api.com` to tenant 1's resources, `tenant2.api.com` to tenant 2's resources. ALB reads the Host header. NLB sees all requests identically.

```yaml
# ✓ ALB - Host-based routing
Listener:
  Rules:
    - Condition:
        HostHeader: tenant1.api.com
      Action:
        TargetGroup: Tenant1TG

    - Condition:
        HostHeader: tenant2.api.com
      Action:
        TargetGroup: Tenant2TG

# ❌ NLB - Can't see Host header
```

### 3. HTTP/HTTPS-Specific Features

You need HTTP to HTTPS redirects, custom HTTP error pages, authentication (Cognito/OIDC), or WAF protection. These are Layer 7 features. NLB operates at Layer 4 and doesn't understand HTTP.

```yaml
# ✓ ALB - HTTP to HTTPS redirect
Listener:
  Port: 80
  DefaultAction:
    Type: redirect
    RedirectConfig:
      Protocol: HTTPS
      Port: 443
      StatusCode: HTTP_301

# ✓ ALB - Custom error pages
TargetGroup:
  Properties:
    HealthCheckPath: /health
    Matcher:
      HttpCode: 200
  # Return custom 503 page when no healthy targets

# ❌ NLB - Doesn't understand HTTP, can't redirect or customize responses
```

### 4. Lambda Targets

You want to invoke Lambda functions directly from the load balancer. ALB can route HTTP requests to Lambda. NLB doesn't support Lambda targets (Lambda speaks HTTP, NLB speaks TCP).

```yaml
# ✓ ALB - Lambda integration
TargetGroup:
  Type: AWS::ElasticLoadBalancingV2::TargetGroup
  Properties:
    TargetType: lambda
    Targets:
      - Id: !GetAtt MyLambda.Arn

# ❌ NLB - No Lambda support
```

### 5. Microservices Architecture

You're running containers on ECS/EKS with dynamic port mappings. ALB integrates natively with service discovery and routes to container IPs + dynamic ports. NLB can do this too, but ALB's content-based routing is critical for microservices (route by path/header to different services).

```yaml
# ✓ ALB - Microservices with path routing
/api/users    → UserService containers
/api/orders   → OrderService containers
/api/products → ProductService containers

All on one ALB, routing by path

# ❌ NLB - Would need separate NLBs per service or manual port management
```

## When to Use NLB

**Choose Network Load Balancer when you need:**

### 1. Extreme Performance (Millions RPS)

Your gaming platform handles 10 million connections per second. ALB tops out around 100K req/sec per AZ before you need multiple ALBs. NLB handles millions per second with sub-millisecond latency because it doesn't terminate connections or parse payloads.

```
Performance comparison (per AZ):

ALB:
  - Throughput: ~100,000 req/sec
  - Latency: 10-20 ms (connection termination + parsing)
  - Processing: Parses HTTP headers, routing logic

NLB:
  - Throughput: Millions req/sec
  - Latency: <1 ms (packet forwarding only)
  - Processing: Minimal (IP + port forwarding)
```

### 2. Static IP Addresses

Your enterprise customers require whitelisting specific IPs in their firewalls. ALB uses dynamic IPs (DNS only). NLB supports Elastic IPs - one static IP per AZ.

```yaml
# ❌ ALB - Dynamic IPs, DNS only
ALB DNS: my-alb-1234567890.us-east-1.elb.amazonaws.com
IPs change: 52.1.2.3, 52.1.2.4, 52.1.2.5 (rotates)
Firewall whitelist: Can't whitelist (IPs change)

# ✓ NLB - Static Elastic IPs
NLB with EIP:
  us-east-1a: 52.1.100.1 (static)
  us-east-1b: 52.1.100.2 (static)
  us-east-1c: 52.1.100.3 (static)
Firewall whitelist: Whitelist 3 IPs (never change)
```

### 3. Source IP Preservation

Your application needs the true client IP for geo-blocking, rate limiting, or logging. ALB terminates connections (client connects to ALB, ALB connects to backend) - backend sees ALB's IP. NLB passes connections through - backend sees the original client IP.

```yaml
# ❌ ALB - Connection termination
Client (203.0.113.5) → ALB (10.0.1.10) → Backend
Backend sees source IP: 10.0.1.10 (ALB's private IP)
Workaround: Parse X-Forwarded-For header

# ✓ NLB - Source IP preserved
Client (203.0.113.5) → NLB → Backend
Backend sees source IP: 203.0.113.5 (true client IP)
No header parsing needed
```

### 4. Non-HTTP Protocols

You're load balancing database connections (MySQL, PostgreSQL), message queues (MQTT, AMQP), or custom TCP/UDP protocols. These aren't HTTP. ALB only understands HTTP/HTTPS/HTTP2/gRPC. NLB handles any TCP/UDP traffic.

```yaml
# ✓ NLB - MySQL load balancing
Listener:
  Protocol: TCP
  Port: 3306
  TargetGroup: MySQLServers

# ✓ NLB - UDP traffic (gaming, DNS, VoIP)
Listener:
  Protocol: UDP
  Port: 514  # Syslog
  TargetGroup: SyslogServers

# ❌ ALB - HTTP/HTTPS only
```

### 5. PrivateLink (VPC Endpoint Services)

You're exposing services to other AWS accounts via PrivateLink. Only NLB supports PrivateLink endpoint services. Customers connect to your service through VPC endpoints without traversing the internet.

```yaml
# ✓ NLB - PrivateLink support
VPCEndpointService:
  Type: AWS::EC2::VPCEndpointService
  Properties:
    NetworkLoadBalancerArns:
      - !Ref MyNLB
    AcceptanceRequired: true

# Customer account connects via VPC endpoint (private, secure)

# ❌ ALB - No PrivateLink support
```

### 6. TLS Passthrough

Your backend handles TLS termination (end-to-end encryption). NLB forwards encrypted traffic without decrypting. ALB terminates TLS (decrypts) and re-encrypts to backend - your traffic is briefly decrypted in the load balancer.

```yaml
# ✓ NLB - TLS passthrough
Client → [TLS] → NLB → [TLS] → Backend
Traffic encrypted end-to-end, NLB never sees plaintext

# ❌ ALB - TLS termination
Client → [TLS] → ALB → [TLS] → Backend
ALB decrypts and re-encrypts (traffic briefly in plaintext at ALB)
```

## Target Groups and Health Checks

Both ALB and NLB route traffic to target groups. A target group is a logical grouping of targets (EC2 instances, IPs, containers, Lambda functions) with health check configuration.

### Target Types

| Target Type  | ALB | NLB | Use Case                                    |
| ------------ | --- | --- | ------------------------------------------- |
| **instance** | ✓   | ✓   | EC2 instances (node port)                   |
| **ip**       | ✓   | ✓   | IP addresses (on-prem, containers, peered VPC) |
| **lambda**   | ✓   | ✗   | Lambda functions                            |

```yaml
# EC2 instances
TargetGroup:
  TargetType: instance
  Targets:
    - Id: i-1234567890abcdef0
      Port: 80
    - Id: i-0fedcba0987654321
      Port: 80

# IP addresses (ECS Fargate, EKS)
TargetGroup:
  TargetType: ip
  Targets:
    - Id: 10.0.1.10
      Port: 8080
    - Id: 10.0.1.11
      Port: 8080

# Lambda (ALB only)
TargetGroup:
  TargetType: lambda
  Targets:
    - Id: !GetAtt MyFunction.Arn
```

### Health Check Configuration

Health checks determine if targets are healthy and should receive traffic.

```yaml
TargetGroup:
  Properties:
    # How often to check (default: 30 seconds)
    HealthCheckIntervalSeconds: 30

    # Timeout for health check (default: 5 seconds)
    HealthCheckTimeoutSeconds: 5

    # Consecutive successes to mark healthy (default: 2)
    # Min: 2, Max: 10
    HealthyThresholdCount: 2

    # Consecutive failures to mark unhealthy (default: 2)
    # Min: 2, Max: 10
    UnhealthyThresholdCount: 2

    # ALB - HTTP health check
    HealthCheckProtocol: HTTP
    HealthCheckPath: /health
    Matcher:
      HttpCode: 200-299  # Success codes

    # NLB - TCP health check (simpler)
    HealthCheckProtocol: TCP
    # Just checks if port is open
```

### Health Check Best Practices

```yaml
# ❌ Wrong - Shallow health check
HealthCheckPath: /
Matcher:
  HttpCode: 200

# Service returns 200 even if database is down
# Load balancer thinks it's healthy, sends traffic
# Requests fail at application layer

# ✓ Correct - Deep health check
HealthCheckPath: /health/deep
Matcher:
  HttpCode: 200

# /health/deep endpoint:
# - Checks database connectivity
# - Checks dependent services
# - Returns 503 if unhealthy
# Load balancer drains traffic if unhealthy
```

```yaml
# ❌ Wrong - Aggressive thresholds
HealthCheckIntervalSeconds: 5
HealthyThresholdCount: 2
UnhealthyThresholdCount: 2

# Target marked unhealthy after 10 seconds (2 failures × 5 sec)
# Causes flapping during brief network issues

# ✓ Correct - Conservative thresholds
HealthCheckIntervalSeconds: 30
HealthyThresholdCount: 3
UnhealthyThresholdCount: 3

# Target marked unhealthy after 90 seconds (3 failures × 30 sec)
# More stable, tolerates brief issues
```

### Deregistration Delay (Connection Draining)

When a target is deregistered (instance terminated, health check fails), the load balancer stops sending NEW requests but may have IN-FLIGHT requests. Deregistration delay keeps the target registered for a grace period to complete existing requests.

```yaml
TargetGroup:
  Properties:
    # How long to wait for in-flight requests to complete
    TargetGroupAttributes:
      - Key: deregistration_delay.timeout_seconds
        Value: "300"  # 5 minutes (default)

# Timeline:
# t=0: Target fails health check
# t=0: Load balancer stops sending NEW requests
# t=0 to t=300: Existing connections complete
# t=300: Target fully deregistered, connections closed
```

```yaml
# ❌ Wrong - Too short
deregistration_delay.timeout_seconds: "30"
# Long-running requests (file uploads, processing) get killed after 30 sec

# ✓ Correct - Match request duration
deregistration_delay.timeout_seconds: "300"
# For long-running operations (uploads, batch processing)

# ✓ Correct - Short for quick APIs
deregistration_delay.timeout_seconds: "60"
# For fast APIs where requests complete in <1 second
```

## Routing Strategies (ALB Only)

ALB supports sophisticated routing based on request attributes.

### Path-Based Routing

```yaml
# Route by URL path
Listener:
  Rules:
    # Priority 1: Highest priority
    - Priority: 1
      Condition:
        PathPattern: /api/v2/*
      Action:
        TargetGroup: APIv2TargetGroup

    # Priority 2
    - Priority: 2
      Condition:
        PathPattern: /api/v1/*
      Action:
        TargetGroup: APIv1TargetGroup

    # Priority 3: Static files
    - Priority: 3
      Condition:
        PathPattern: /static/*
      Action:
        TargetGroup: StaticFileServers

    # Default rule (lowest priority)
    - Priority: default
      Action:
        TargetGroup: DefaultTargetGroup
```

### Host-Based Routing

```yaml
# Route by hostname
Listener:
  Rules:
    - Condition:
        HostHeader: api.example.com
      Action:
        TargetGroup: APIServers

    - Condition:
        HostHeader: admin.example.com
      Action:
        TargetGroup: AdminServers

    - Condition:
        HostHeader: "*.example.com"  # Wildcard
      Action:
        TargetGroup: DefaultServers
```

### HTTP Header Routing

```yaml
# Route by custom header
Listener:
  Rules:
    # Mobile app traffic
    - Condition:
        HttpHeader:
          Name: User-Agent
          Values: ["*Mobile*", "*Android*", "*iOS*"]
      Action:
        TargetGroup: MobileBackend

    # API version header
    - Condition:
        HttpHeader:
          Name: X-API-Version
          Values: ["2.0"]
      Action:
        TargetGroup: APIv2
```

### Query String Routing

```yaml
# Route by query parameter
Listener:
  Rules:
    # Beta users
    - Condition:
        QueryString:
          - Key: beta
            Value: "true"
      Action:
        TargetGroup: BetaServers

    # A/B testing
    - Condition:
        QueryString:
          - Key: variant
            Value: "B"
      Action:
        TargetGroup: VariantBServers
```

### Weighted Target Groups (Traffic Splitting)

```yaml
# Canary deployment - 10% traffic to new version
Listener:
  DefaultAction:
    Type: forward
    ForwardConfig:
      TargetGroups:
        - TargetGroupArn: !Ref StableVersion
          Weight: 90
        - TargetGroupArn: !Ref CanaryVersion
          Weight: 10
      TargetGroupStickinessConfig:
        Enabled: true
        DurationSeconds: 3600
```

## Sticky Sessions

Sticky sessions (session affinity) route requests from the same client to the same target.

### ALB Sticky Sessions (Cookie-Based)

```yaml
TargetGroup:
  Properties:
    TargetGroupAttributes:
      - Key: stickiness.enabled
        Value: "true"
      - Key: stickiness.type
        Value: "lb_cookie"  # ALB-generated cookie
      - Key: stickiness.lb_cookie.duration_seconds
        Value: "86400"  # 24 hours

# ALB sets cookie: AWSALB=<encoded-target-info>
# Client includes cookie in subsequent requests
# ALB routes to same target
```

**When to use:**
- ✓ Stateful applications (user sessions stored in memory)
- ✓ WebSocket connections
- ✗ Stateless applications (adds unnecessary complexity)

**Gotchas:**
- ❌ Uneven load distribution (one target gets overloaded)
- ❌ Reduces effectiveness of auto-scaling
- ❌ Longer draining time during deployments

### NLB Sticky Sessions (Flow Hash)

```yaml
TargetGroup:
  Properties:
    TargetGroupAttributes:
      - Key: stickiness.enabled
        Value: "true"
      - Key: stickiness.type
        Value: "source_ip"  # Hash based on source IP

# NLB hashes: (source IP, dest IP, source port, dest port, protocol)
# Same hash → same target
# No cookies (Layer 4)
```

**Difference from ALB:**
- ALB: Cookie-based (application-level, works across different IPs)
- NLB: Flow hash (network-level, same source IP → same target)

## Cross-Zone Load Balancing

Load balancers are deployed across multiple Availability Zones. Without cross-zone load balancing, each AZ's load balancer only routes to targets in that AZ. With cross-zone, traffic is distributed evenly across ALL targets in ALL zones.

### Without Cross-Zone

```
AZ-A: 2 targets, receives 50% of traffic
  → Each target: 25% of total traffic

AZ-B: 8 targets, receives 50% of traffic
  → Each target: 6.25% of total traffic

Problem: Uneven load (AZ-A targets get 4× more traffic)
```

### With Cross-Zone

```
All 10 targets receive equal traffic:
  → Each target: 10% of total traffic

Benefit: Even distribution regardless of AZ target count
```

### Configuration

```yaml
# ALB - Cross-zone ON by default, no charge
LoadBalancer:
  Properties:
    Type: application
  # Cross-zone always enabled, can't disable

# NLB - Cross-zone OFF by default, charges for data transfer
LoadBalancer:
  Properties:
    Type: network
    LoadBalancerAttributes:
      - Key: load_balancing.cross_zone.enabled
        Value: "true"

# Cost: $0.01/GB for cross-zone data transfer
# 100 GB/day cross-zone = $30/month extra
```

**When to enable (NLB):**
- ✓ Uneven target distribution across AZs
- ✓ Even load distribution is critical
- ✗ High data transfer costs outweigh benefits
- ✗ Targets evenly distributed across AZs

## Connection Handling

### ALB - Connection Termination

```
Client → [Connection 1] → ALB → [Connection 2] → Target

ALB maintains two separate connections:
  1. Client to ALB
  2. ALB to target

Benefits:
  - HTTP keep-alive pooling (reuses connections to targets)
  - Buffers slow clients (fast connection to target)
  - Parses HTTP (enables content routing)

Drawbacks:
  - Higher latency (connection termination overhead)
  - Backend sees ALB IP (need X-Forwarded-For for client IP)
```

### NLB - Connection Passthrough

```
Client → [Connection] → NLB → [Same Connection] → Target

NLB forwards packets (doesn't terminate):
  - Preserves source IP
  - Sub-millisecond latency
  - Supports millions of connections

Benefits:
  - Lower latency (no termination)
  - Source IP preserved
  - Higher throughput

Drawbacks:
  - No HTTP parsing (can't route by path/header)
  - No connection pooling
```

## Cost Comparison

Both ALB and NLB use similar pricing models, but NLB is generally cheaper.

### ALB Pricing

```
Hourly: $0.0225/hour
LCU: $0.008/LCU-hour

LCU dimensions (charged for highest):
  - New connections/sec: 25 = 1 LCU
  - Active connections: 3,000 = 1 LCU
  - Processed bytes: 1 GB/hour = 1 LCU
  - Rule evaluations: 1,000/sec = 1 LCU
```

### NLB Pricing

```
Hourly: $0.0225/hour (same as ALB)
NLCU: $0.006/NLCU-hour (25% cheaper than LCU)

NLCU dimensions (charged for highest):
  - New connections/sec: 800 = 1 NLCU (32× more than ALB)
  - Active connections: 100,000 = 1 NLCU (33× more than ALB)
  - Processed bytes: 1 GB/hour = 1 NLCU (same as ALB)
```

### Example Calculation

```
Traffic: 10,000 new connections/sec, 50,000 active connections, 100 GB/hour

ALB:
  New connections: 10,000 ÷ 25 = 400 LCU
  Active connections: 50,000 ÷ 3,000 = 16.7 LCU
  Processed bytes: 100 GB/hour = 100 LCU ← HIGHEST
  Cost: $16.43 + (100 × $0.008 × 730) = $600/month

NLB:
  New connections: 10,000 ÷ 800 = 12.5 NLCU
  Active connections: 50,000 ÷ 100,000 = 0.5 NLCU
  Processed bytes: 100 GB/hour = 100 NLCU ← HIGHEST
  Cost: $16.43 + (100 × $0.006 × 730) = $454/month

Savings: $146/month (24%)
```

**NLB is cheaper when:**
- High connection count (NLB handles more connections per NLCU)
- Processed bytes is the limiting dimension (25% cheaper NLCU)

**ALB costs more for:**
- Rule evaluations (complex routing adds LCU charges)

## Common Mistakes

### Mistake 1: Using NLB for HTTP When ALB Would Work

```yaml
# ❌ Wrong - NLB for microservices routing
Architecture:
  NLB (TCP:80) → Nginx (parses HTTP, routes by path) → Microservices

Problems:
  - Extra hop (Nginx layer)
  - Manage Nginx fleet (scaling, health)
  - Complexity

# ✓ Correct - ALB with path-based routing
Architecture:
  ALB (path-based routing) → Microservices

Benefits:
  - Native path routing
  - No extra layer
  - Simpler architecture
```

### Mistake 2: Using ALB When Performance is Critical

```yaml
# ❌ Wrong - ALB for high-performance TCP
Architecture:
  Gaming server (10M connections/sec) → ALB

Problems:
  - ALB can't handle millions req/sec
  - HTTP parsing overhead (game uses custom TCP)
  - Higher latency

# ✓ Correct - NLB for performance
Architecture:
  Gaming server (10M connections/sec) → NLB

Benefits:
  - Sub-millisecond latency
  - Handles millions connections/sec
  - No HTTP overhead
```

### Mistake 3: Sticky Sessions on Stateless Apps

```yaml
# ❌ Wrong - Sticky sessions for stateless API
TargetGroup:
  Properties:
    TargetGroupAttributes:
      - Key: stickiness.enabled
        Value: "true"

Problems:
  - Uneven load (some targets overloaded)
  - Slower auto-scaling response
  - Longer deployments (drain time)

# ✓ Correct - No sticky sessions for stateless
TargetGroup:
  Properties:
    TargetGroupAttributes:
      - Key: stickiness.enabled
        Value: "false"

Benefits:
  - Even load distribution
  - Fast scaling
  - Quick deployments
```

### Mistake 4: Not Tuning Health Checks

```yaml
# ❌ Wrong - Default health check for critical service
HealthCheckIntervalSeconds: 30
HealthyThresholdCount: 2
UnhealthyThresholdCount: 2

Problems:
  - 60 seconds to detect failure (2 × 30)
  - 60 seconds of failed requests before marked unhealthy
  - Slow recovery (60 seconds to mark healthy)

# ✓ Correct - Tuned for critical service
HealthCheckIntervalSeconds: 10
HealthyThresholdCount: 2
UnhealthyThresholdCount: 2

Benefits:
  - 20 seconds to detect failure (2 × 10)
  - Faster recovery
  - Less downtime
```

## Hands-On Exercise 1: Choose ALB or NLB

For each scenario, choose ALB or NLB and justify your choice.

**Scenario 1: E-commerce API**
- Traffic: 5,000 req/sec HTTP traffic
- Routes: `/api/products`, `/api/users`, `/api/orders` to different microservices
- Need: TLS termination, path-based routing
- Backend: ECS containers

**Scenario 2: Real-Time Gaming**
- Traffic: 1 million concurrent TCP connections, custom binary protocol
- Need: Sub-millisecond latency, static IPs for client whitelisting
- Backend: EC2 instances

**Scenario 3: Internal Database Cluster**
- Traffic: 10,000 connections/sec to PostgreSQL cluster
- Need: Load balance across read replicas, preserve source IP for logging
- Backend: RDS read replicas + EC2 instances

**Scenario 4: WebSocket Chat**
- Traffic: 50,000 concurrent WebSocket connections
- Need: Sticky sessions, path routing (`/chat/rooms/*` to room service)
- Backend: ECS containers

**Scenario 5: Multi-Tenant SaaS**
- Traffic: 10,000 req/sec HTTP
- Need: Route by hostname (`tenant1.app.com`, `tenant2.app.com`)
- Backend: EC2 instances per tenant

<details>
<summary>Solution</summary>

**Scenario 1: ALB** ✓
- HTTP traffic (Layer 7)
- Path-based routing required (`/api/products` → ProductService)
- TLS termination (ALB has ACM integration)
- Microservices architecture (ALB's strength)
- Traffic moderate (5K req/sec well within ALB capacity)

**Scenario 2: NLB** ✓
- Custom TCP protocol (not HTTP, ALB can't handle)
- Extreme performance (1M connections, need sub-ms latency)
- Static IPs required (NLB supports Elastic IPs)
- No need for HTTP features

**Scenario 3: NLB** ✓
- TCP traffic (database protocol)
- Source IP preservation (for logging/audit)
- No HTTP routing needed (all traffic goes to same pool)
- Cost: NLB cheaper for TCP (no HTTP overhead)

**Scenario 4: ALB** ✓
- WebSocket (both support, but ALB has path routing)
- Path routing required (`/chat/rooms/*`)
- Sticky sessions (both support, ALB cookie-based better for WebSocket)
- HTTP-based protocol (WebSocket starts as HTTP upgrade)

**Scenario 5: ALB** ✓
- Host-based routing (NLB can't route by hostname)
- HTTP traffic
- Multi-tenancy (ALB's host-based routing perfect for this)
- Moderate traffic

</details>

## Hands-On Exercise 2: Debug Health Check Issues

You deployed an application behind an ALB. Targets keep flapping between healthy and unhealthy. Debug the issue.

**Current configuration:**

```yaml
TargetGroup:
  HealthCheckProtocol: HTTP
  HealthCheckPath: /
  HealthCheckIntervalSeconds: 10
  HealthCheckTimeoutSeconds: 5
  HealthyThresholdCount: 2
  UnhealthyThresholdCount: 2
  Matcher:
    HttpCode: 200

DeregistrationDelay: 30 seconds
```

**Symptoms:**
- Targets marked unhealthy every 2-3 minutes
- Then marked healthy again 20 seconds later
- Repeat cycle
- Application logs show no errors
- CloudWatch shows no CPU/memory spikes

<details>
<summary>Solution</summary>

**Root Cause Analysis:**

The issue is likely **health check path timing**. The application root path `/` may take >5 seconds to respond (timeout), especially during garbage collection or temporary load spikes.

**Problems with current config:**

1. ✗ Health check path `/` may load heavy resources
2. ✗ Timeout too short (5 sec) - GC pauses cause timeouts
3. ✗ Threshold too aggressive (2 failures = unhealthy in 20 sec)
4. ✗ No dedicated health endpoint

**Fixed configuration:**

```yaml
TargetGroup:
  # ✓ Dedicated lightweight health endpoint
  HealthCheckPath: /health

  # ✓ More conservative timing
  HealthCheckIntervalSeconds: 30
  HealthCheckTimeoutSeconds: 10  # Tolerates brief GC pauses
  HealthyThresholdCount: 3
  UnhealthyThresholdCount: 3

  Matcher:
    HttpCode: 200

  # ✓ Longer deregistration (give time for in-flight requests)
  DeregistrationDelay: 120 seconds
```

**Create `/health` endpoint:**

```javascript
// ✓ Lightweight health check
app.get('/health', (req, res) => {
  // Quick checks only
  res.status(200).send('OK');
});

// ❌ Heavy health check (causes flapping)
app.get('/health', async (req, res) => {
  // These are too slow:
  await db.query('SELECT 1');  // DB latency
  await redis.ping();           // Redis latency
  // Total: >5 seconds possible
});
```

**Verification:**

```bash
# Test health check manually
time curl -i http://backend-ip:8080/health

# Should respond in <1 second
# HTTP/1.1 200 OK
# Real time: 0.234s
```

**Improved configuration prevents flapping:**
- Unhealthy after: 3 failures × 30 sec = 90 seconds (tolerates transient issues)
- Healthy after: 3 successes × 30 sec = 90 seconds (stable recovery)
- Deregistration: 120 seconds (in-flight requests complete)

</details>

## Interview Questions

### Q1: What's the difference between ALB and NLB, and when would you use each?

This is THE classic load balancer question. Tests fundamental understanding of OSI layers and whether you can match technology to use case. The interviewer wants to see if you understand the performance vs features trade-off.

<details>
<summary>Answer</summary>

**ALB (Application Load Balancer) - Layer 7:**
- Operates at HTTP/HTTPS level
- Content-based routing (path, host, header, query string)
- HTTP-specific features (redirects, authentication, WAF)
- Connection termination (proxy)
- ~10-20 ms latency, ~100K req/sec per AZ
- More expensive (higher LCU charges for routing)

**NLB (Network Load Balancer) - Layer 4:**
- Operates at TCP/UDP level
- IP + port routing only
- Extreme performance (millions req/sec, <1 ms latency)
- Connection passthrough (preserves source IP)
- Static IP support (Elastic IPs)
- Cheaper (lower NLCU charges)

**Use ALB when:**
- HTTP/HTTPS traffic
- Need content-based routing (microservices, multi-tenant)
- Need HTTP features (redirects, auth, custom errors)
- Lambda targets
- Moderate traffic (<100K req/sec)

**Use NLB when:**
- Extreme performance required (>100K req/sec, <1ms latency)
- Non-HTTP protocols (TCP, UDP, custom)
- Need static IPs
- Source IP preservation critical
- PrivateLink endpoint services

**Example decision:**
```
Microservices API (HTTP, 10K req/sec, path routing):
  → ALB (content routing needed)

Gaming server (custom TCP, 1M connections, ultra-low latency):
  → NLB (performance + non-HTTP)
```

</details>

### Q2: How does ALB preserve the client's IP address, and why does this matter?

Tests understanding of connection termination and HTTP headers. Shows whether you've debugged issues related to logging, geo-blocking, or rate limiting where client IP is critical.

<details>
<summary>Answer</summary>

**Problem: ALB Connection Termination**

```
Client (203.0.113.5) → ALB (10.0.1.10) → Backend
Backend sees TCP source IP: 10.0.1.10 (ALB's private IP)
```

Backend can't see the real client IP (203.0.113.5) because ALB terminates the connection.

**Solution: X-Forwarded-For Header**

ALB adds HTTP headers with client information:

```
X-Forwarded-For: 203.0.113.5, 10.0.1.10
  - Original client IP
  - ALB IP (if multiple proxies)

X-Forwarded-Port: 443
  - Original port

X-Forwarded-Proto: https
  - Original protocol
```

**Backend code:**

```javascript
// ❌ Wrong - reads ALB IP
const clientIP = req.connection.remoteAddress;
// Returns: 10.0.1.10 (ALB IP)

// ✓ Correct - reads X-Forwarded-For
const clientIP = req.headers['x-forwarded-for']?.split(',')[0];
// Returns: 203.0.113.5 (real client)
```

**Why this matters:**

1. **Logging/Analytics**: Need real client IP for user tracking
2. **Geo-blocking**: Rate limiting based on client IP
3. **Security**: Detect abuse from specific IPs
4. **Compliance**: Audit logs must show real client IP

**NLB alternative:**

NLB preserves source IP (no header parsing needed):
```
Client (203.0.113.5) → NLB → Backend
Backend TCP source: 203.0.113.5 (preserved)
```

</details>

### Q3: Explain sticky sessions and when you should/shouldn't use them.

Tests understanding of stateful vs stateless architecture. Shows whether you've dealt with session management in distributed systems and understand the scaling implications.

<details>
<summary>Answer</summary>

**Sticky Sessions (Session Affinity):**

Route requests from the same client to the same backend target.

**ALB Implementation:**
```yaml
# Cookie-based
TargetGroup:
  TargetGroupAttributes:
    stickiness.enabled: true
    stickiness.type: lb_cookie
    stickiness.lb_cookie.duration_seconds: 86400

# ALB sets cookie: AWSALB=<target-info>
# Subsequent requests → same target
```

**NLB Implementation:**
```yaml
# Flow hash (source IP based)
TargetGroup:
  TargetGroupAttributes:
    stickiness.enabled: true
    stickiness.type: source_ip

# Hash(source IP, dest IP, protocol) → same target
```

**When to use:**

✓ **Stateful applications**
```
User session stored in memory on backend
Login state, shopping cart in memory
WebSocket connections (must reconnect to same server)
```

✓ **Caching on backend**
```
Backend caches user data locally
Stickiness improves cache hit rate
```

**When NOT to use:**

❌ **Stateless applications**
```
Session stored in database/Redis (shared state)
All backends can handle any request
Stickiness adds no value, only drawbacks
```

**Drawbacks:**

1. **Uneven load distribution**
   - Some targets get overloaded
   - Auto-scaling less effective

2. **Slower deployments**
   - Must wait for sessions to drain
   - Longer deregistration delay

3. **Reduced availability**
   - Target fails → users lose sessions
   - Need to re-login

**Best practice: Design for stateless**

```javascript
// ❌ Stateful (requires sticky sessions)
const sessions = {};
app.post('/login', (req, res) => {
  const sessionId = generateId();
  sessions[sessionId] = { userId: req.body.userId };
  res.cookie('sessionId', sessionId);
});

// ✓ Stateless (no sticky sessions needed)
app.post('/login', (req, res) => {
  const token = jwt.sign({ userId: req.body.userId });
  res.json({ token });
});

// Any backend can validate JWT
// No server-side session state
```

</details>

### Q4: How would you handle a health check that keeps marking targets as unhealthy during deployments?

Tests practical troubleshooting and deployment strategy knowledge. Shows whether you understand health check timing, graceful shutdown, and zero-downtime deployments.

<details>
<summary>Answer</summary>

**Problem: Targets marked unhealthy during deployment**

```
Deployment starts
  → Application restarts
  → Health checks fail during startup
  → Target marked unhealthy
  → Load balancer drains traffic
  → Deployment continues
  → Application ready
  → Health checks succeed
  → Target marked healthy
  → 60-90 seconds of downtime
```

**Solutions:**

**1. Increase health check thresholds (temporary fix)**

```yaml
# Before deployment, temporarily adjust:
HealthCheckIntervalSeconds: 30
HealthyThresholdCount: 2
UnhealthyThresholdCount: 5  # ← More tolerant

# 5 failures × 30 sec = 150 sec before unhealthy
# Gives app time to restart
```

**2. Implement graceful shutdown (proper fix)**

```javascript
// Backend: Handle SIGTERM
let isShuttingDown = false;

process.on('SIGTERM', () => {
  isShuttingDown = true;

  // Stop accepting new requests
  server.close(() => {
    // Complete in-flight requests
    // Then exit
    process.exit(0);
  });
});

// Health check returns 503 when shutting down
app.get('/health', (req, res) => {
  if (isShuttingDown) {
    return res.status(503).send('Shutting down');
  }
  res.status(200).send('OK');
});
```

**3. Use deregistration delay**

```yaml
TargetGroup:
  TargetGroupAttributes:
    deregistration_delay.timeout_seconds: "300"

# Timeline:
# 1. Deployment starts → SIGTERM sent
# 2. App /health returns 503
# 3. Load balancer detects unhealthy
# 4. Load balancer stops NEW requests
# 5. In-flight requests complete (up to 300 sec)
# 6. Target deregistered
# 7. Container terminates
```

**4. Blue/Green or Rolling deployments**

```yaml
# Blue/Green: Zero downtime
1. Deploy new version (green) to new target group
2. Health checks pass on green
3. Switch traffic: blue → green
4. Drain blue
5. Terminate blue

# Rolling: Gradual replacement
1. Deploy to 1 instance
2. Wait for health checks
3. Deploy to next instance
4. Repeat (always maintain healthy capacity)
```

**Complete solution:**

```yaml
TargetGroup:
  # Conservative health checks
  HealthCheckIntervalSeconds: 30
  HealthyThresholdCount: 2
  UnhealthyThresholdCount: 3

  # Graceful shutdown period
  TargetGroupAttributes:
    deregistration_delay.timeout_seconds: "120"

# Backend: Graceful shutdown handler
# Deployment: Blue/Green or rolling
```

**Result: Zero downtime deployments**

</details>

## Key Takeaways

1. **Layer 7 vs Layer 4**: ALB operates at HTTP (application), NLB at TCP/UDP (transport)
2. **ALB Best For**: HTTP routing (path/host/header), microservices, Lambda targets, moderate traffic
3. **NLB Best For**: Extreme performance, static IPs, non-HTTP protocols, source IP preservation
4. **Performance**: NLB handles millions req/sec with <1ms latency, ALB ~100K req/sec with 10-20ms
5. **Routing**: ALB has sophisticated routing (path, host, header), NLB only IP:port
6. **Sticky Sessions**: Use for stateful apps (WebSocket, in-memory sessions), avoid for stateless
7. **Health Checks**: Tune thresholds based on app startup time, use dedicated lightweight endpoint
8. **Cost**: NLB ~25% cheaper (NLCU vs LCU), especially for high connection count

## Next Steps

In [Lesson 04: CloudFront & Edge Optimization](lesson-04-cloudfront-edge-optimization.md), you'll learn:

- Cache behaviors and invalidation strategies
- Origin failover and high availability
- Lambda@Edge vs CloudFront Functions
- Geographic restrictions and signed URLs
- Cost optimization through cache hit ratio improvement
