# Deployment Strategies and Load Balancing Algorithms

## Overview

ECS supports multiple deployment strategies including true blue/green deployments with pre-production validation, similar to Kubernetes. ALB also supports different load balancing algorithms beyond simple round-robin.

## ECS Deployment Strategies

### 1. Rolling Update (Default)

The basic gradual replacement strategy:

```json
{
  "deploymentConfiguration": {
    "maximumPercent": 200,
    "minimumHealthyPercent": 100
  }
}
```

**What happens:**

- Start new tasks (up to 200% = 6 tasks for desiredCount: 3)
- Wait for health checks
- Stop old tasks
- Not true blue/green (gradual replacement)

### 2. Blue/Green Deployment (CodeDeploy)

**True blue/green** with validation before cutover:

```json
{
  "deploymentController": {
    "type": "CODE_DEPLOY"
  }
}
```

**How it works:**

```
Step 1: Blue environment (production)
  ALB Listener:80 → Blue Target Group
    ├─→ Task 1 (v1)
    ├─→ Task 2 (v1)
    └─→ Task 3 (v1)

Step 2: Deploy green environment (new version)
  ALB Listener:80 → Blue Target Group (still serving traffic)
    ├─→ Task 1 (v1)
    ├─→ Task 2 (v1)
    └─→ Task 3 (v1)

  ALB Listener:8080 → Green Target Group (test traffic)
    ├─→ Task 4 (v2) ← NEW
    ├─→ Task 5 (v2) ← NEW
    └─→ Task 6 (v2) ← NEW

Step 3: Test green environment
  - Run automated tests against :8080
  - Manual verification
  - Health checks pass

Step 4: Traffic shift (cutover)
  ALB Listener:80 → Green Target Group
    ├─→ Task 4 (v2)
    ├─→ Task 5 (v2)
    └─→ Task 6 (v2)

Step 5: Cleanup
  - Old blue tasks terminated
  - Blue target group removed
```

#### Setting Up Blue/Green with CodeDeploy

**Create Service with CodeDeploy Controller:**

```bash
aws ecs create-service \
  --cluster my-cluster \
  --service-name web-service \
  --task-definition web-app:1 \
  --desired-count 3 \
  --deployment-controller type=CODE_DEPLOY \
  --load-balancers targetGroupArn=arn:...,containerName=web,containerPort=80
```

**Create CodeDeploy Application:**

```bash
# Create application
aws deploy create-application \
  --application-name ecs-blue-green-app \
  --compute-platform ECS

# Create deployment group
aws deploy create-deployment-group \
  --application-name ecs-blue-green-app \
  --deployment-group-name ecs-deployment-group \
  --service-role-arn arn:aws:iam::123456789012:role/CodeDeployRole \
  --ecs-services clusterName=my-cluster,serviceName=web-service \
  --load-balancer-info targetGroupPairInfoList=[{
      targetGroups=[{name=blue-tg},{name=green-tg}],
      prodTrafficRoute={listenerArns=[arn:...:listener/app/my-alb/.../...]}
    }]
```

**Deploy New Version:**

```bash
aws deploy create-deployment \
  --application-name ecs-blue-green-app \
  --deployment-group-name ecs-deployment-group \
  --revision revisionType=AppSpecContent,appSpecContent={
      content="version: 0.0
Resources:
  - TargetService:
      Type: AWS::ECS::Service
      Properties:
        TaskDefinition: arn:aws:ecs:...:task-definition/web-app:2
        LoadBalancerInfo:
          ContainerName: web
          ContainerPort: 80"
    }
```

#### CDK Example: Blue/Green Deployment

```typescript
import * as ecs from "aws-cdk-lib/aws-ecs";
import * as codedeploy from "aws-cdk-lib/aws-codedeploy";
import * as elbv2 from "aws-cdk-lib/aws-elasticloadbalancingv2";

// Create two target groups (blue and green)
const blueTargetGroup = new elbv2.ApplicationTargetGroup(
  this,
  "BlueTargetGroup",
  {
    vpc,
    port: 80,
    targetType: elbv2.TargetType.IP,
    healthCheck: { path: "/health" },
  }
);

const greenTargetGroup = new elbv2.ApplicationTargetGroup(
  this,
  "GreenTargetGroup",
  {
    vpc,
    port: 80,
    targetType: elbv2.TargetType.IP,
    healthCheck: { path: "/health" },
  }
);

// Create ALB with production listener
const lb = new elbv2.ApplicationLoadBalancer(this, "LB", {
  vpc,
  internetFacing: true,
});

const prodListener = lb.addListener("ProdListener", {
  port: 80,
  defaultTargetGroups: [blueTargetGroup],
});

// Test listener for validating green environment
const testListener = lb.addListener("TestListener", {
  port: 8080,
  defaultTargetGroups: [greenTargetGroup],
});

// Create ECS service with CodeDeploy deployment controller
const service = new ecs.FargateService(this, "Service", {
  cluster,
  taskDefinition,
  desiredCount: 3,
  deploymentController: {
    type: ecs.DeploymentControllerType.CODE_DEPLOY,
  },
});

// Attach initial target group
service.attachToApplicationTargetGroup(blueTargetGroup);

// Create CodeDeploy deployment group
const deploymentGroup = new codedeploy.EcsDeploymentGroup(
  this,
  "BlueGreenDG",
  {
    service,
    blueGreenDeploymentConfig: {
      blueTargetGroup,
      greenTargetGroup,
      listener: prodListener,
      testListener,
    },
    deploymentConfig: codedeploy.EcsDeploymentConfig.CANARY_10PERCENT_5MINUTES,
  }
);
```

### 3. Canary Deployment (Gradual Traffic Shift)

CodeDeploy supports **canary** and **linear** traffic shifting.

#### AppSpec File

```yaml
version: 0.0
Resources:
  - TargetService:
      Type: AWS::ECS::Service
      Properties:
        TaskDefinition: "arn:aws:ecs:...:task-definition/web-app:2"
        LoadBalancerInfo:
          ContainerName: "web"
          ContainerPort: 80

Hooks:
  - BeforeInstall: "LambdaFunctionToValidateBeforeInstall"
  - AfterInstall: "LambdaFunctionToValidateAfterInstall"
  - AfterAllowTestTraffic: "LambdaFunctionToValidateAfterTestTraffic"
  - BeforeAllowTraffic: "LambdaFunctionToValidateBeforeAllowingProdTraffic"
  - AfterAllowTraffic: "LambdaFunctionToValidateAfterAllowingProdTraffic"
```

#### Traffic Shift Strategies

**Canary10Percent5Minutes:**

```
Time 0:   10% → Green (new), 90% → Blue (old)
Time 5m:  100% → Green
```

**Canary10Percent15Minutes:**

```
Time 0:    10% → Green, 90% → Blue
Time 15m:  100% → Green
```

**Linear10PercentEvery3Minutes:**

```
Time 0:   10% → Green, 90% → Blue
Time 3m:  20% → Green, 80% → Blue
Time 6m:  30% → Green, 70% → Blue
...
Time 27m: 100% → Green
```

**AllAtOnce:**

```
Time 0:   100% → Green
```

#### Validation Lambda Example

```python
import boto3
import requests

def lambda_handler(event, context):
    # Get test endpoint from CodeDeploy event
    test_endpoint = get_test_endpoint(event)

    # Run tests against green environment
    try:
        response = requests.get(f"{test_endpoint}/health")
        assert response.status_code == 200

        response = requests.get(f"{test_endpoint}/api/version")
        assert response.json()["version"] == "2.0.0"

        # All tests passed
        return {
            'statusCode': 200,
            'body': 'Green environment validated successfully'
        }
    except Exception as e:
        # Tests failed - CodeDeploy will rollback
        raise Exception(f"Validation failed: {str(e)}")
```

### 4. Circuit Breaker (Auto-Rollback)

ECS has a **circuit breaker** that automatically rolls back failed deployments:

```json
{
  "deploymentConfiguration": {
    "deploymentCircuitBreaker": {
      "enable": true,
      "rollback": true
    },
    "maximumPercent": 200,
    "minimumHealthyPercent": 100
  }
}
```

**How it works:**

```
Step 1: Deploy new version
  New tasks: Task4(v2), Task5(v2), Task6(v2)

Step 2: Health checks fail repeatedly
  Task4(v2) ✗ failed health check
  Task5(v2) ✗ failed health check
  Task6(v2) ✗ failed health check

Step 3: Circuit breaker triggers
  - Stops deployment
  - Rolls back to previous task definition (v1)
  - Sends CloudWatch alarm

Step 4: Old version restored
  Task7(v1), Task8(v1), Task9(v1) ← rolled back
```

**CDK Example:**

```typescript
const service = new ecs.FargateService(this, "Service", {
  cluster,
  taskDefinition,
  desiredCount: 3,
  circuitBreaker: { rollback: true }, // Auto-rollback on failure
  deploymentConfiguration: {
    maximumPercent: 200,
    minimumHealthyPercent: 100,
  },
});
```

### 5. External Deployment Controller

For custom deployment logic (e.g., using Kubernetes-style controllers):

```json
{
  "deploymentController": {
    "type": "EXTERNAL"
  }
}
```

- You manage the entire deployment lifecycle
- ECS only launches/stops tasks when you tell it to
- Use for advanced custom deployments

## ALB Load Balancing Algorithms

ALB supports multiple load balancing algorithms:

### 1. Round Robin (Default)

```
Request 1 → Task 1 (10.0.1.50)
Request 2 → Task 2 (10.0.1.51)
Request 3 → Task 3 (10.0.1.52)
Request 4 → Task 1 (10.0.1.50) ← back to first
Request 5 → Task 2 (10.0.1.51)
```

**Characteristics:**

- Simple rotation
- Default for HTTP/HTTPS
- Good for similar request processing times

### 2. Least Outstanding Requests

```
Target Group:
  Task 1: 5 active requests
  Task 2: 2 active requests  ← SELECTED (least busy)
  Task 3: 8 active requests

New request → Task 2
```

**How to configure:**

```bash
aws elbv2 modify-target-group-attributes \
  --target-group-arn arn:... \
  --attributes Key=load_balancing.algorithm.type,Value=least_outstanding_requests
```

**When to use:**

- Requests have varying processing times
- Some requests are much slower than others
- Better distribution for long-running requests

### 3. Flow Hash (NLB only)

For Network Load Balancers (not ALB):

- Hash based on protocol, source/dest IP, source/dest port
- Same client always goes to same target
- Good for connection persistence

## Comparison: Deployment Strategies

| Strategy            | Downtime | Rollback  | Traffic Shift          | Test Before Cutover | Complexity |
| ------------------- | -------- | --------- | ---------------------- | ------------------- | ---------- |
| **Rolling**         | None     | Manual    | Gradual (task by task) | No                  | Low        |
| **Blue/Green**      | None     | Instant   | Instant or gradual     | Yes (test listener) | Medium     |
| **Canary**          | None     | Automatic | Gradual (% based)      | Yes                 | Medium     |
| **Circuit Breaker** | None     | Automatic | Gradual                | No (auto-detects)   | Low        |

## Example: Complete Blue/Green Setup

### 1. Task Definition

```json
{
  "family": "web-app",
  "containerDefinitions": [
    {
      "name": "web",
      "image": "nginx:1.21",
      "portMappings": [{ "containerPort": 80 }],
      "healthCheck": {
        "command": [
          "CMD-SHELL",
          "curl -f http://localhost/health || exit 1"
        ],
        "interval": 30,
        "timeout": 5,
        "retries": 3
      }
    }
  ]
}
```

### 2. AppSpec.yaml for CodeDeploy

```yaml
version: 0.0
Resources:
  - TargetService:
      Type: AWS::ECS::Service
      Properties:
        TaskDefinition: "arn:aws:ecs:us-east-1:123456789012:task-definition/web-app:2"
        LoadBalancerInfo:
          ContainerName: "web"
          ContainerPort: 80
        PlatformVersion: "LATEST"

Hooks:
  - BeforeInstall: "LambdaFunctionToRunBeforeInstall"
  - AfterInstall: "LambdaFunctionToRunAfterInstall"
  - AfterAllowTestTraffic: "LambdaFunctionToTestGreenEnvironment"
  - BeforeAllowTraffic: "LambdaFunctionToRunBeforeProductionTraffic"
  - AfterAllowTraffic: "LambdaFunctionToValidateDeployment"
```

### 3. Deploy with Canary

```bash
aws deploy create-deployment \
  --application-name ecs-app \
  --deployment-group-name ecs-dg \
  --deployment-config-name CodeDeployDefault.ECSCanary10Percent5Minutes \
  --description "Deploying v2 with 10% canary"
```

## Monitoring Deployments

### CloudWatch Metrics

```typescript
// Create alarm for failed deployments
const failedDeploymentAlarm = new cloudwatch.Alarm(this, "FailedDeployment", {
  metric: service.metricFailedDeployments(),
  threshold: 1,
  evaluationPeriods: 1,
  alarmDescription: "Alert on failed ECS deployment",
});

// Monitor CPU during deployment
const cpuAlarm = new cloudwatch.Alarm(this, "HighCPU", {
  metric: service.metricCpuUtilization(),
  threshold: 80,
  evaluationPeriods: 2,
});
```

### Deployment Events

```bash
# Watch deployment progress
aws ecs describe-services \
  --cluster my-cluster \
  --services web-service \
  --query 'services[0].deployments'

# Output:
[
  {
    "id": "ecs-svc/1234567890",
    "status": "PRIMARY",
    "taskDefinition": "web-app:2",
    "desiredCount": 3,
    "runningCount": 3,
    "rolloutState": "COMPLETED"
  }
]
```

## Best Practices

### 1. Use Blue/Green for Critical Services

```typescript
const service = new ecs.FargateService(this, "Service", {
  cluster,
  taskDefinition,
  desiredCount: 3,
  deploymentController: {
    type: ecs.DeploymentControllerType.CODE_DEPLOY,
  },
});
```

### 2. Enable Circuit Breaker as Fallback

```typescript
const service = new ecs.FargateService(this, "Service", {
  cluster,
  taskDefinition,
  desiredCount: 3,
  circuitBreaker: { rollback: true },
});
```

### 3. Configure Proper Health Checks

```json
{
  "healthCheck": {
    "command": ["CMD-SHELL", "curl -f http://localhost/health || exit 1"],
    "interval": 30,
    "timeout": 5,
    "retries": 3,
    "startPeriod": 60
  }
}
```

### 4. Use Least Outstanding Requests for Variable Workloads

```bash
aws elbv2 modify-target-group-attributes \
  --target-group-arn arn:... \
  --attributes Key=load_balancing.algorithm.type,Value=least_outstanding_requests
```

### 5. Set Deregistration Delay for Graceful Shutdown

```typescript
const targetGroup = new elbv2.ApplicationTargetGroup(this, "TG", {
  vpc,
  port: 80,
  targetType: elbv2.TargetType.IP,
  deregistrationDelay: cdk.Duration.seconds(30), // Wait 30s before stopping task
});
```

## Summary Tables

### Deployment Strategies

| Feature                     | Rolling | Blue/Green         | Canary             |
| --------------------------- | ------- | ------------------ | ------------------ |
| **Test before production**  | ✗       | ✓ (test listener)  | ✓ (gradual shift)  |
| **Instant rollback**        | ✗       | ✓                  | ✓                  |
| **Zero downtime**           | ✓       | ✓                  | ✓                  |
| **Resource overhead**       | Low     | High (2x tasks)    | High (2x tasks)    |
| **Complexity**              | Low     | Medium             | Medium             |

### Load Balancing Algorithms

| Algorithm                        | Best For               | How It Works                |
| -------------------------------- | ---------------------- | --------------------------- |
| **Round Robin**                  | Similar request times  | Rotates through targets     |
| **Least Outstanding Requests**   | Variable request times | Routes to least busy target |
| **Flow Hash (NLB)**              | Connection persistence | Same client → same target   |

## Key Takeaways

1. **Blue/Green with CodeDeploy** gives you true pre-production validation (similar to Kubernetes)
2. **Test listener on port 8080** lets you validate green environment before cutover
3. **Circuit Breaker** provides automatic rollback for rolling deployments
4. **Canary deployments** allow gradual traffic shift with validation hooks
5. **Least Outstanding Requests** is better than Round Robin for variable workloads
6. **Health checks** are critical for all deployment strategies
7. **Lambda hooks** enable automated testing during blue/green deployments
8. **Deregistration delay** ensures graceful connection draining during deployments
