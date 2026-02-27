# Load Balancing with Application Load Balancer (ALB)

## Overview

When using **awsvpc mode** (each task gets its own IP), ECS automatically manages load balancer target groups by registering and deregistering task IPs.

## How Service-Level Load Balancing Works

### The Flow

```
Internet
    ↓
Application Load Balancer (ALB)
    ↓
Target Group (managed by ECS)
    ├─→ Task 1 (IP: 10.0.1.50:80)
    ├─→ Task 2 (IP: 10.0.1.51:80)
    └─→ Task 3 (IP: 10.0.1.52:80)
```

## Automatic Target Registration

ECS **automatically registers and deregisters** task IPs with the target group:

1. **Task starts** → ECS registers task's IP:port to target group
2. **Health check passes** → ALB starts sending traffic to task
3. **Task stops/fails** → ECS deregisters task's IP from target group
4. **New task starts** → Process repeats

## Service Definition with Load Balancer

```json
{
  "serviceName": "web-service",
  "cluster": "my-cluster",
  "taskDefinition": "web-app:1",
  "desiredCount": 3,
  "launchType": "FARGATE",
  "networkConfiguration": {
    "awsvpcConfiguration": {
      "subnets": ["subnet-abc123", "subnet-def456"],
      "securityGroups": ["sg-task123"],
      "assignPublicIp": "DISABLED"
    }
  },
  "loadBalancers": [
    {
      "targetGroupArn": "arn:aws:elasticloadbalancing:...",
      "containerName": "web",
      "containerPort": 80
    }
  ],
  "healthCheckGracePeriodSeconds": 60
}
```

### Key Configuration Elements

- **targetGroupArn**: The ALB target group where tasks will be registered
- **containerName**: Which container in the task definition to route traffic to
- **containerPort**: The port your application listens on (not hostPort)

## Setup Steps

### 1. Create Target Group

```bash
aws elbv2 create-target-group \
  --name my-web-targets \
  --protocol HTTP \
  --port 80 \
  --vpc-id vpc-123456 \
  --target-type ip \
  --health-check-path /health
```

**Important**: For awsvpc mode, target type must be **ip** (not instance).

### 2. Create Load Balancer

```bash
aws elbv2 create-load-balancer \
  --name my-alb \
  --subnets subnet-abc123 subnet-def456 \
  --security-groups sg-alb123
```

### 3. Create Listener

```bash
aws elbv2 create-listener \
  --load-balancer-arn arn:aws:elasticloadbalancing:... \
  --protocol HTTP \
  --port 80 \
  --default-actions Type=forward,TargetGroupArn=arn:...
```

### 4. Create ECS Service

```bash
aws ecs create-service \
  --cluster my-cluster \
  --service-name web-service \
  --task-definition web-app:1 \
  --desired-count 3 \
  --launch-type FARGATE \
  --network-configuration "awsvpcConfiguration={subnets=[subnet-abc],securityGroups=[sg-123]}" \
  --load-balancers targetGroupArn=arn:...,containerName=web,containerPort=80
```

ECS now automatically:

- Launches 3 tasks with unique IPs
- Registers each task IP to the target group
- Monitors task health
- Replaces failed tasks and updates target group

## Target Types: IP vs Instance

### IP Target Type (awsvpc mode)

```
Target Group (type: ip)
  ├─→ 10.0.1.50:80 (Task 1)
  ├─→ 10.0.1.51:80 (Task 2)
  └─→ 10.0.1.52:80 (Task 3)
```

- Used with **awsvpc** network mode
- Registers task's private IP directly
- Works with Fargate and EC2
- **Recommended for modern deployments**

### Instance Target Type (bridge mode)

```
Target Group (type: instance)
  ├─→ EC2 Instance i-123:32768 (Task 1 - dynamic port)
  ├─→ EC2 Instance i-123:32769 (Task 2 - dynamic port)
  └─→ EC2 Instance i-456:32770 (Task 3 - dynamic port)
```

- Used with **bridge** network mode
- Registers EC2 instance ID + dynamic port
- Only works with EC2 launch type
- Legacy approach

## Health Checks

The load balancer continuously health checks each task:

```json
{
  "healthCheckPath": "/health",
  "healthCheckIntervalSeconds": 30,
  "healthCheckTimeoutSeconds": 5,
  "healthyThresholdCount": 2,
  "unhealthyThresholdCount": 3
}
```

**Flow:**

1. ALB sends GET request to `http://task-ip:80/health` every 30 seconds
2. If task responds with 200 OK → healthy
3. If task fails 3 checks → marked unhealthy, removed from rotation
4. ECS notices unhealthy task → stops and replaces it

## CDK Example

```typescript
import * as ecs from "aws-cdk-lib/aws-ecs";
import * as elbv2 from "aws-cdk-lib/aws-elasticloadbalancingv2";

// Create ALB
const lb = new elbv2.ApplicationLoadBalancer(this, "LB", {
  vpc,
  internetFacing: true,
});

// Create target group
const targetGroup = new elbv2.ApplicationTargetGroup(this, "TargetGroup", {
  vpc,
  port: 80,
  protocol: elbv2.ApplicationProtocol.HTTP,
  targetType: elbv2.TargetType.IP, // For awsvpc mode
  healthCheck: {
    path: "/health",
    interval: cdk.Duration.seconds(30),
  },
});

// Add listener
const listener = lb.addListener("Listener", {
  port: 80,
  defaultTargetGroups: [targetGroup],
});

// Create Fargate service
const service = new ecs.FargateService(this, "Service", {
  cluster,
  taskDefinition,
  desiredCount: 3,
});

// Connect service to load balancer
service.attachToApplicationTargetGroup(targetGroup);
```

**What happens:**

- Service launches 3 Fargate tasks
- Each task gets unique IP (e.g., 10.0.1.50, 10.0.1.51, 10.0.1.52)
- ECS registers all 3 IPs to target group
- ALB distributes traffic across them

## Traffic Flow Example

```
1. User request: http://my-alb-123.us-east-1.elb.amazonaws.com
   ↓
2. DNS resolves to ALB: 54.123.45.67
   ↓
3. ALB receives request
   ↓
4. ALB checks target group health:
   - 10.0.1.50:80 ✓ healthy
   - 10.0.1.51:80 ✓ healthy
   - 10.0.1.52:80 ✓ healthy
   ↓
5. ALB picks target (round-robin): 10.0.1.51:80
   ↓
6. ALB forwards request to Task 2 at 10.0.1.51:80
   ↓
7. Task processes and responds
   ↓
8. ALB returns response to user
```

## Scaling Example

When you scale from 3 to 5 tasks:

```bash
aws ecs update-service \
  --cluster my-cluster \
  --service web-service \
  --desired-count 5
```

**What happens:**

```
Before (3 tasks):
Target Group
  ├─→ 10.0.1.50:80 (Task 1)
  ├─→ 10.0.1.51:80 (Task 2)
  └─→ 10.0.1.52:80 (Task 3)

After (5 tasks):
Target Group
  ├─→ 10.0.1.50:80 (Task 1) ✓ existing
  ├─→ 10.0.1.51:80 (Task 2) ✓ existing
  ├─→ 10.0.1.52:80 (Task 3) ✓ existing
  ├─→ 10.0.1.53:80 (Task 4) ← newly registered
  └─→ 10.0.1.54:80 (Task 5) ← newly registered
```

ECS automatically:

1. Starts 2 new tasks
2. Waits for health checks to pass
3. Registers new task IPs to target group
4. ALB starts sending traffic to all 5 tasks

## Rolling Deployment Example

When you deploy a new task definition version:

```bash
aws ecs update-service \
  --cluster my-cluster \
  --service web-service \
  --task-definition web-app:2
```

**What happens (with default rolling update):**

```
Step 1: Start new tasks
  Old: Task1(v1), Task2(v1), Task3(v1)
  New: Task4(v2), Task5(v2) ← starting

Step 2: Wait for health checks
  Old: Task1(v1), Task2(v1), Task3(v1) ← still serving traffic
  New: Task4(v2) ✓, Task5(v2) ✓ ← healthy, now serving traffic

Step 3: Stop old tasks
  Old: Task1(v1) ← stopping, Task2(v1) ← stopping, Task3(v1) ← stopping
  New: Task4(v2) ✓, Task5(v2) ✓ ← serving all traffic

Step 4: Start remaining new tasks
  New: Task4(v2), Task5(v2), Task6(v2) ← all serving traffic
```

ECS automatically:

- Deregisters old task IPs from target group before stopping them
- Waits for connection draining (default 300 seconds)
- Ensures minimum healthy percentage maintained

## Security Groups

With awsvpc mode, you need two security groups:

### ALB Security Group

```
Inbound:
- Port 80/443 from 0.0.0.0/0 (internet)

Outbound:
- Port 80 to Task Security Group
```

### Task Security Group

```
Inbound:
- Port 80 from ALB Security Group

Outbound:
- Port 443 to 0.0.0.0/0 (for external API calls)
```

### Example

```typescript
// ALB security group
const albSg = new ec2.SecurityGroup(this, "AlbSg", {
  vpc,
  allowAllOutbound: true,
});
albSg.addIngressRule(ec2.Peer.anyIpv4(), ec2.Port.tcp(80));

// Task security group
const taskSg = new ec2.SecurityGroup(this, "TaskSg", {
  vpc,
  allowAllOutbound: true,
});
taskSg.addIngressRule(albSg, ec2.Port.tcp(80));
```

## Service Discovery (Alternative to ALB)

For **service-to-service** communication, you can use AWS Cloud Map instead of load balancers:

```typescript
const service = new ecs.FargateService(this, "Service", {
  cluster,
  taskDefinition,
  desiredCount: 3,
  cloudMapOptions: {
    name: "api",
    dnsRecordType: ecs.DnsRecordType.A,
    dnsTtl: cdk.Duration.seconds(10),
  },
});
```

**Result:**

- Service registered as `api.local` in Cloud Map
- DNS query returns all task IPs: `10.0.1.50, 10.0.1.51, 10.0.1.52`
- Client applications use DNS-based load balancing
- No ALB needed for internal services

## Comparison: awsvpc vs bridge Mode Load Balancing

### awsvpc Mode (Modern)

```
ALB Target Group (type: ip)
  ├─→ Task 1: 10.0.1.50:80
  ├─→ Task 2: 10.0.1.51:80
  └─→ Task 3: 10.0.1.52:80
```

**Characteristics:**

- Each task has unique IP
- Direct IP registration
- Works with Fargate
- Simpler security group management

### bridge Mode (Legacy)

```
ALB Target Group (type: instance)
  ├─→ EC2 i-123:32768 → Task 1
  ├─→ EC2 i-123:32769 → Task 2
  └─→ EC2 i-456:32770 → Task 3
```

**Characteristics:**

- Tasks share EC2 instance IP
- Dynamic port mapping required
- Only works with EC2
- More complex setup

## Summary Table

| Aspect             | How It Works                                         |
| ------------------ | ---------------------------------------------------- |
| **Task IPs**       | Each task gets unique IP in awsvpc mode              |
| **Registration**   | ECS automatically registers task IPs to target group |
| **Deregistration** | ECS automatically deregisters when task stops        |
| **Health Checks**  | ALB continuously checks each task IP                 |
| **Target Type**    | Must use `ip` target type (not `instance`)           |
| **Scaling**        | New tasks automatically added to target group        |
| **Deployment**     | Old tasks drained, new tasks added seamlessly        |

## Key Takeaways

1. **ECS manages the target group** - you don't manually register IPs
2. **Target type must be `ip`** for awsvpc mode
3. **Each task gets its own IP** and is registered individually
4. **Health checks determine** which tasks receive traffic
5. **Scaling and deployments** are handled automatically by ECS
6. **Security groups** control traffic flow from ALB to tasks
7. **Service Discovery** is an alternative for internal service-to-service communication

The beauty of this integration is that **you never manually manage IPs** - ECS handles all registration/deregistration automatically as tasks come and go.
