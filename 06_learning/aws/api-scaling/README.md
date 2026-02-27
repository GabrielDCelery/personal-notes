# AWS API Scaling Interview Prep

A focused, comprehensive refresher for experienced developers on handling API scaling challenges using AWS services.

## Lesson Series

### Request Entry & Distribution
- [Lesson 01: API Gateway Deep Dive](lesson-01-api-gateway-deep-dive.md) - REST vs HTTP APIs, throttling, caching, integrations, transformations
- [Lesson 02: API Gateway vs Application Load Balancer](lesson-02-api-gateway-vs-alb.md) - Cost comparison, when to use each, hybrid architectures
- [Lesson 03: Load Balancers (ALB vs NLB)](lesson-03-load-balancers-alb-nlb.md) - Layer 7 vs Layer 4, target groups, health checks, routing strategies
- [Lesson 04: CloudFront & Edge Optimization](lesson-04-cloudfront-edge-optimization.md) - Cache behaviors, Lambda@Edge, origin failover, cost optimization

### Compute & Scaling Strategies
- [Lesson 05: Auto Scaling Patterns](lesson-05-auto-scaling-patterns.md) - EC2/ECS scaling, metrics, warm pools, scaling policies
- [Lesson 06: Lambda at Scale](lesson-06-lambda-at-scale.md) - Concurrency management, cold starts, VPC implications, cost optimization

### Resilience & Decoupling
- [Lesson 07: Queue-Based Decoupling](lesson-07-queue-based-decoupling.md) - SQS/SNS/EventBridge, handling spikes, retry strategies, backpressure

## Target Audience

Developers with 2+ years AWS experience preparing for interviews that focus on:
- "How would you handle 1 million requests per second?"
- Scaling strategies and trade-offs
- Cost vs performance optimization
- Service selection and architectural decisions
- Common scaling pitfalls and edge cases

## Usage

Each lesson includes:
- Service comparison tables for quick reference
- Real-world scaling scenarios
- Cost vs performance trade-offs with ✓/❌ patterns
- Interview questions focusing on architectural decisions
- Hands-on exercises designing scalable architectures

## Prerequisites

Basic understanding of:
- AWS fundamentals (VPCs, security groups, IAM)
- HTTP/HTTPS protocols
- Load balancing concepts
- Basic cloud architecture patterns
