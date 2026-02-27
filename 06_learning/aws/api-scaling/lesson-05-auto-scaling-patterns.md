# Lesson 05: Auto Scaling Patterns

Critical knowledge about AWS Auto Scaling - choosing the right scaling type, configuring policies, optimizing metrics, and avoiding common pitfalls that cause slow scaling or over-provisioning.

## The Scaling Problem

The interviewer asks: "Your API gets 1,000 requests per second at 9 AM and 10,000 requests per second at noon. How do you handle this?" Running 10 instances 24/7 wastes $7,000/month during low traffic. Running 1 instance means users get 503 errors at peak. Auto scaling launches instances when load increases and terminates them when load decreases. But default settings scale too slowly (5-10 minutes to add capacity) or too aggressively (constant thrashing). Know how to tune scaling policies for your traffic patterns.

You're running a web application. Traffic varies by time of day, day of week, and sudden spikes (HackerNews, sales events). Manually adjusting instance count is slow and wastes money. Auto scaling adjusts capacity automatically based on metrics (CPU, requests, queue depth). But auto scaling has many knobs - which metrics to track, how fast to scale, when to scale in vs out. Choose wrong and you're either paying for idle instances or dropping requests during spikes.

## Auto Scaling Types

| Type                        | Use Case                      | Scaling Target    | Metrics                         |
| --------------------------- | ----------------------------- | ----------------- | ------------------------------- |
| **EC2 Auto Scaling**        | EC2 instances in Auto Scaling Group | Instance count | CPU, network, custom metrics    |
| **ECS Service Auto Scaling**| ECS tasks (Fargate or EC2)    | Task count        | CPU, memory, ALB requests, custom|
| **Application Auto Scaling**| DynamoDB, Aurora, Lambda      | Varies            | Read/write capacity, concurrent executions|

## EC2 Auto Scaling

### Auto Scaling Group (ASG) Configuration

```yaml
AutoScalingGroup:
  Type: AWS::AutoScaling::AutoScalingGroup
  Properties:
    AutoScalingGroupName: my-asg
    VPCZoneIdentifier:
      - subnet-1234567890abcdef0
      - subnet-0fedcba0987654321
    LaunchTemplate:
      LaunchTemplateId: !Ref MyLaunchTemplate
      Version: $Latest

    # Capacity limits
    MinSize: 2  # Always keep at least 2 instances (HA)
    MaxSize: 20  # Never exceed 20 instances (cost protection)
    DesiredCapacity: 4  # Start with 4 instances

    # Health checks
    HealthCheckType: ELB  # Use load balancer health checks
    HealthCheckGracePeriod: 300  # 5 min warmup before health checks

    # Termination
    TerminationPolicies:
      - OldestLaunchConfiguration  # Terminate oldest first during scale-in

    # Metrics
    MetricsCollection:
      - Granularity: 1Minute
        Metrics:
          - GroupDesiredCapacity
          - GroupInServiceInstances
```

### Launch Template

```yaml
LaunchTemplate:
  Type: AWS::EC2::LaunchTemplate
  Properties:
    LaunchTemplateName: my-app-template
    LaunchTemplateData:
      ImageId: ami-0c55b159cbfafe1f0  # Amazon Linux 2
      InstanceType: t3.medium
      IamInstanceProfile:
        Arn: !GetAtt InstanceProfile.Arn

      # ✓ User data for application startup
      UserData:
        Fn::Base64: !Sub |
          #!/bin/bash
          yum update -y
          yum install -y docker
          systemctl start docker
          docker run -d -p 80:3000 my-app:latest

          # ✓ Signal Auto Scaling when ready
          /opt/aws/bin/cfn-signal -e $? --stack ${AWS::StackName} \
            --resource AutoScalingGroup --region ${AWS::Region}

      # Monitoring
      Monitoring:
        Enabled: true  # Detailed CloudWatch metrics (1-min intervals)

      # Network
      NetworkInterfaces:
        - DeviceIndex: 0
          AssociatePublicIpAddress: false  # Private instances behind ALB
          Groups:
            - !Ref InstanceSecurityGroup
```

## Scaling Policies

### 1. Target Tracking Scaling (Recommended)

Maintains a target value for a metric. Auto Scaling automatically calculates how many instances needed.

```yaml
# ✓ Track average CPU utilization at 70%
TargetTrackingScalingPolicy:
  Type: AWS::AutoScaling::ScalingPolicy
  Properties:
    AutoScalingGroupName: !Ref AutoScalingGroup
    PolicyType: TargetTrackingScaling
    TargetTrackingConfiguration:
      PredefinedMetricSpecification:
        PredefinedMetricType: ASGAverageCPUUtilization
      TargetValue: 70.0
      # Scale out quickly, scale in slowly
      ScaleOutCooldown: 60  # Wait 60 sec before next scale-out
      ScaleInCooldown: 300  # Wait 5 min before scale-in

# ✓ Track ALB requests per target
ALBRequestCountPolicy:
  Type: AWS::AutoScaling::ScalingPolicy
  Properties:
    AutoScalingGroupName: !Ref AutoScalingGroup
    PolicyType: TargetTrackingScaling
    TargetTrackingConfiguration:
      PredefinedMetricSpecification:
        PredefinedMetricType: ALBRequestCountPerTarget
        ResourceLabel: !Join
          - '/'
          - - !GetAtt ALB.LoadBalancerFullName
            - !GetAtt TargetGroup.TargetGroupFullName
      TargetValue: 1000.0  # 1,000 requests/target

# ✓ Custom metric - SQS queue depth
SQSQueueDepthPolicy:
  Type: AWS::AutoScaling::ScalingPolicy
  Properties:
    AutoScalingGroupName: !Ref AutoScalingGroup
    PolicyType: TargetTrackingScaling
    TargetTrackingConfiguration:
      CustomizedMetricSpecification:
        MetricName: ApproximateNumberOfMessagesVisible
        Namespace: AWS/SQS
        Statistic: Average
        Dimensions:
          - Name: QueueName
            Value: !GetAtt MyQueue.QueueName
      TargetValue: 100.0  # Keep ~100 messages per instance
```

**How target tracking works:**

```
Current: 4 instances, avg CPU 85%
Target: 70% CPU

Calculation:
  Needed capacity = Current capacity × (Current metric / Target metric)
  Needed = 4 × (85 / 70) = 4.86 → rounds to 5 instances

Auto Scaling adds 1 instance

After scale-out:
  5 instances, avg CPU ~68% (below target, good)
```

### 2. Step Scaling

Scale by different amounts based on alarm thresholds.

```yaml
# ✓ Scale out based on CPU thresholds
ScaleOutPolicy:
  Type: AWS::AutoScaling::ScalingPolicy
  Properties:
    AutoScalingGroupName: !Ref AutoScalingGroup
    PolicyType: StepScaling
    AdjustmentType: PercentChangeInCapacity
    MetricAggregationType: Average
    StepAdjustments:
      # CPU 70-80%: Add 10%
      - MetricIntervalLowerBound: 0
        MetricIntervalUpperBound: 10
        ScalingAdjustment: 10

      # CPU 80-90%: Add 20%
      - MetricIntervalLowerBound: 10
        MetricIntervalUpperBound: 20
        ScalingAdjustment: 20

      # CPU >90%: Add 30%
      - MetricIntervalLowerBound: 20
        ScalingAdjustment: 30

# CloudWatch Alarm triggers policy
HighCPUAlarm:
  Type: AWS::CloudWatch::Alarm
  Properties:
    AlarmName: high-cpu-alarm
    MetricName: CPUUtilization
    Namespace: AWS/EC2
    Statistic: Average
    Period: 60  # 1 minute
    EvaluationPeriods: 2  # 2 consecutive periods
    Threshold: 70
    ComparisonOperator: GreaterThanThreshold
    AlarmActions:
      - !Ref ScaleOutPolicy

# ✓ Scale in conservatively
ScaleInPolicy:
  Type: AWS::AutoScaling::ScalingPolicy
  Properties:
    AutoScalingGroupName: !Ref AutoScalingGroup
    PolicyType: StepScaling
    AdjustmentType: ChangeInCapacity
    StepAdjustments:
      # CPU <30%: Remove 1 instance
      - MetricIntervalUpperBound: 0
        ScalingAdjustment: -1

LowCPUAlarm:
  Type: AWS::CloudWatch::Alarm
  Properties:
    MetricName: CPUUtilization
    Statistic: Average
    Period: 300  # 5 minutes (longer than scale-out)
    EvaluationPeriods: 3  # 15 min total (conservative)
    Threshold: 30
    ComparisonOperator: LessThanThreshold
    AlarmActions:
      - !Ref ScaleInPolicy
```

### 3. Scheduled Scaling

Scale based on predictable patterns.

```yaml
# ✓ Scale up before business hours
MorningScaleUp:
  Type: AWS::AutoScaling::ScheduledAction
  Properties:
    AutoScalingGroupName: !Ref AutoScalingGroup
    DesiredCapacity: 10
    MinSize: 10
    MaxSize: 20
    Recurrence: "0 8 * * MON-FRI"  # 8 AM weekdays

# ✓ Scale down after business hours
EveningScaleDown:
  Type: AWS::AutoScaling::ScheduledAction
  Properties:
    AutoScalingGroupName: !Ref AutoScalingGroup
    DesiredCapacity: 2
    MinSize: 2
    MaxSize: 10
    Recurrence: "0 18 * * MON-FRI"  # 6 PM weekdays

# ✓ Weekend minimum
WeekendScale:
  Type: AWS::AutoScaling::ScheduledAction
  Properties:
    AutoScalingGroupName: !Ref AutoScalingGroup
    DesiredCapacity: 2
    MinSize: 2
    MaxSize: 5
    Recurrence: "0 0 * * SAT"  # Midnight Saturday
```

## Choosing the Right Metric

| Metric                         | When to Use                        | Pros                    | Cons                           |
| ------------------------------ | ---------------------------------- | ----------------------- | ------------------------------ |
| **CPU Utilization**            | CPU-bound workloads                | Simple, built-in        | Doesn't reflect user load      |
| **Network In/Out**             | Network-intensive apps             | Reflects traffic        | Varies by request size         |
| **ALB Request Count**          | Web apps behind ALB                | Directly tracks load    | Doesn't account for request complexity|
| **ALB Target Response Time**   | Latency-sensitive apps             | User experience metric  | Can be noisy                   |
| **SQS Queue Depth**            | Worker pools processing queue      | Directly tracks backlog | Need to tune target value      |
| **Custom Metric**              | Specific workload requirements     | Flexible                | More complex to implement      |

### CPU Utilization

```yaml
# ✓ Good for: Compute-intensive apps (video encoding, ML inference)
TargetTrackingConfiguration:
  PredefinedMetricSpecification:
    PredefinedMetricType: ASGAverageCPUUtilization
  TargetValue: 70.0

# Why 70% target?
# - 30% headroom for spikes
# - Allows brief bursts without scaling
# - Not too low (wastes capacity) or too high (slow response)

# ❌ Bad for: I/O-bound apps (database queries, API calls)
# - CPU may be 10% while app is slow (waiting on external services)
```

### ALB Request Count Per Target

```yaml
# ✓ Good for: Web APIs, HTTP services
TargetTrackingConfiguration:
  PredefinedMetricSpecification:
    PredefinedMetricType: ALBRequestCountPerTarget
    ResourceLabel: !Sub
      - "${LoadBalancerFullName}/${TargetGroupFullName}"
      - LoadBalancerFullName: !GetAtt ALB.LoadBalancerFullName
        TargetGroupFullName: !GetAtt TargetGroup.TargetGroupFullName
  TargetValue: 1000.0  # 1,000 req/min per instance

# How to calculate target:
# 1. Load test: Find max sustainable requests per instance
#    Example: Instance handles 2,000 req/min before latency spikes
# 2. Add headroom: 2,000 × 0.7 = 1,400 req/min
# 3. Set target: 1,400 (70% of capacity)

# ✓ Scales with actual load (user requests)
# ❌ Doesn't account for request complexity (lightweight vs heavy)
```

### SQS Queue Depth

```yaml
# ✓ Good for: Worker pools, async processing
TargetTrackingConfiguration:
  CustomizedMetricSpecification:
    MetricName: ApproximateNumberOfMessagesVisible
    Namespace: AWS/SQS
    Statistic: Average
    Dimensions:
      - Name: QueueName
        Value: !GetAtt ProcessingQueue.QueueName
  TargetValue: 100.0

# How to calculate target:
# 1. Measure throughput: Instance processes 10 messages/sec = 600/min
# 2. Desired lag: Want to process backlog in 10 minutes
# 3. Target queue depth: 100 messages per instance (10 min × 10 msg/sec)

# Benefits:
# - Scales based on work to be done
# - Prevents queue from growing unbounded
# - Natural backpressure mechanism
```

### Custom Metric (Application-Specific)

```javascript
// ✓ Example: Active WebSocket connections
const AWS = require('aws-sdk');
const cloudwatch = new AWS.CloudWatch();

setInterval(async () => {
  const activeConnections = getActiveWebSocketConnections();

  await cloudwatch.putMetricData({
    Namespace: 'MyApp',
    MetricData: [{
      MetricName: 'ActiveConnections',
      Value: activeConnections,
      Unit: 'Count',
      Timestamp: new Date()
    }]
  }).promise();
}, 60000);  // Every minute
```

```yaml
# Scale based on custom metric
CustomMetricPolicy:
  Type: AWS::AutoScaling::ScalingPolicy
  Properties:
    PolicyType: TargetTrackingScaling
    TargetTrackingConfiguration:
      CustomizedMetricSpecification:
        MetricName: ActiveConnections
        Namespace: MyApp
        Statistic: Average
      TargetValue: 5000.0  # 5,000 connections per instance
```

## Warm Pools

Pre-initialized instances ready to join ASG quickly. Reduces scale-out time from 5 minutes to 30 seconds.

```yaml
AutoScalingGroup:
  Properties:
    # ... other properties
    WarmPoolConfiguration:
      MinSize: 2  # Keep 2 instances warm
      MaxGroupPreparedCapacity: 5  # Max warm + in-service instances
      PoolState: Stopped  # Stop instances (cheaper than Running)
      InstanceReusePolicy:
        ReuseOnScaleIn: true  # Return instances to warm pool on scale-in

# Cost comparison:
# Running warm pool: 2 instances × $0.05/hr = $0.10/hr = $73/month
# Stopped warm pool: 2 instances × $0 (CPU) + EBS = ~$10/month
# Benefit: Scale out in 30 sec instead of 5 min
```

**Warm pool states:**
- `Stopped`: Instances stopped (no compute charge, only EBS)
- `Running`: Instances running (full charge, instant scale-out)
- `Hibernated`: Instances hibernated (RAM to EBS, resume faster than Stopped)

## Lifecycle Hooks

Execute custom actions during instance launch or termination.

```yaml
LaunchLifecycleHook:
  Type: AWS::AutoScaling::LifecycleHook
  Properties:
    AutoScalingGroupName: !Ref AutoScalingGroup
    LifecycleTransition: autoscaling:EC2_INSTANCE_LAUNCHING
    DefaultResult: ABANDON  # Terminate if hook times out
    HeartbeatTimeout: 300  # 5 minutes to complete hook
    NotificationTargetARN: !GetAtt LifecycleQueue.Arn
    RoleARN: !GetAtt LifecycleRole.Arn

TerminateLifecycleHook:
  Type: AWS::AutoScaling::LifecycleHook
  Properties:
    AutoScalingGroupName: !Ref AutoScalingGroup
    LifecycleTransition: autoscaling:EC2_INSTANCE_TERMINATING
    DefaultResult: CONTINUE  # Terminate even if hook times out
    HeartbeatTimeout: 300
    NotificationTargetARN: !GetAtt LifecycleQueue.Arn
    RoleARN: !GetAtt LifecycleRole.Arn
```

**Lifecycle hook flow:**

```
Instance Launching:
  1. EC2 instance starts
  2. Lifecycle hook: Pending:Wait
  3. Lambda processes hook (warm up app, register with service discovery)
  4. Lambda completes hook: autoscaling:CONTINUE
  5. Instance enters InService

Instance Terminating:
  1. Auto Scaling starts termination
  2. Lifecycle hook: Terminating:Wait
  3. Lambda processes hook (deregister, drain connections, backup data)
  4. Lambda completes hook: autoscaling:CONTINUE
  5. Instance terminates
```

**Lambda handler for lifecycle hook:**

```javascript
const AWS = require('aws-sdk');
const autoscaling = new AWS.AutoScaling();

exports.handler = async (event) => {
  const message = JSON.parse(event.Records[0].body);
  const { LifecycleHookName, AutoScalingGroupName, EC2InstanceId, LifecycleTransition } = message;

  try {
    if (LifecycleTransition === 'autoscaling:EC2_INSTANCE_LAUNCHING') {
      // ✓ Warm up application
      await warmUpApplication(EC2InstanceId);

      // ✓ Register with service discovery
      await registerInstance(EC2InstanceId);
    } else if (LifecycleTransition === 'autoscaling:EC2_INSTANCE_TERMINATING') {
      // ✓ Drain connections
      await drainConnections(EC2InstanceId);

      // ✓ Backup data
      await backupInstanceData(EC2InstanceId);
    }

    // Complete lifecycle action
    await autoscaling.completeLifecycleAction({
      LifecycleHookName,
      AutoScalingGroupName,
      LifecycleActionResult: 'CONTINUE',
      InstanceId: EC2InstanceId
    }).promise();

  } catch (error) {
    console.error('Lifecycle hook failed:', error);

    // ABANDON if launch failed, CONTINUE if termination (terminate anyway)
    await autoscaling.completeLifecycleAction({
      LifecycleHookName,
      AutoScalingGroupName,
      LifecycleActionResult: LifecycleTransition.includes('LAUNCHING') ? 'ABANDON' : 'CONTINUE',
      InstanceId: EC2InstanceId
    }).promise();
  }
};
```

## ECS Service Auto Scaling

Similar to EC2 Auto Scaling but for ECS tasks.

```yaml
ECSService:
  Type: AWS::ECS::Service
  Properties:
    Cluster: !Ref ECSCluster
    TaskDefinition: !Ref TaskDefinition
    DesiredCount: 4
    LaunchType: FARGATE

# Scaling target
ScalableTarget:
  Type: AWS::ApplicationAutoScaling::ScalableTarget
  Properties:
    ServiceNamespace: ecs
    ResourceId: !Sub service/${ECSCluster}/${ECSService.Name}
    ScalableDimension: ecs:service:DesiredCount
    MinCapacity: 2
    MaxCapacity: 20
    RoleARN: !GetAtt AutoScalingRole.Arn

# ✓ Track CPU utilization
CPUScalingPolicy:
  Type: AWS::ApplicationAutoScaling::ScalingPolicy
  Properties:
    PolicyName: cpu-scaling
    ServiceNamespace: ecs
    ScalableDimension: ecs:service:DesiredCount
    ResourceId: !Sub service/${ECSCluster}/${ECSService.Name}
    PolicyType: TargetTrackingScaling
    TargetTrackingScalingPolicyConfiguration:
      PredefinedMetricSpecification:
        PredefinedMetricType: ECSServiceAverageCPUUtilization
      TargetValue: 70.0
      ScaleOutCooldown: 60
      ScaleInCooldown: 300

# ✓ Track memory utilization
MemoryScalingPolicy:
  Type: AWS::ApplicationAutoScaling::ScalingPolicy
  Properties:
    PolicyType: TargetTrackingScaling
    TargetTrackingScalingPolicyConfiguration:
      PredefinedMetricSpecification:
        PredefinedMetricType: ECSServiceAverageMemoryUtilization
      TargetValue: 80.0

# ✓ Track ALB requests
ALBScalingPolicy:
  Type: AWS::ApplicationAutoScaling::ScalingPolicy
  Properties:
    PolicyType: TargetTrackingScaling
    TargetTrackingScalingPolicyConfiguration:
      PredefinedMetricSpecification:
        PredefinedMetricType: ALBRequestCountPerTarget
        ResourceLabel: !Sub
          - "${ALBFullName}/${TargetGroupFullName}"
          - ALBFullName: !GetAtt ALB.LoadBalancerFullName
            TargetGroupFullName: !GetAtt TargetGroup.TargetGroupFullName
      TargetValue: 1000.0
```

**ECS vs EC2 Auto Scaling:**

| Aspect              | EC2 Auto Scaling              | ECS Service Auto Scaling       |
| ------------------- | ----------------------------- | ------------------------------ |
| **Unit**            | EC2 instances                 | ECS tasks                      |
| **Scale time**      | 3-5 min (boot + app startup)  | 30 sec - 2 min (Fargate)       |
| **Metrics**         | Instance-level (CPU, network) | Task-level (CPU, mem) + ALB    |
| **Bin packing**     | N/A                           | ✓ (multiple tasks per instance)|
| **Cost**            | Pay for instances             | Pay for tasks (Fargate) or instances (EC2)|

## Common Mistakes

### Mistake 1: Scaling Too Slowly

```yaml
# ❌ Wrong - Slow scale-out
ScalingPolicy:
  TargetValue: 70
  ScaleOutCooldown: 300  # 5 min between scale-outs

Alarm:
  Period: 300  # 5 min
  EvaluationPeriods: 3  # 15 min total before alarm

# Timeline:
# t=0: Load spike, CPU 85%
# t=15: Alarm triggers (after 3× 5min periods)
# t=20: Instance launches (5 min boot)
# t=20-25: Traffic still overloaded
# Result: 20-25 minutes to handle spike

# ✓ Correct - Fast scale-out
ScalingPolicy:
  TargetValue: 70
  ScaleOutCooldown: 60  # 1 min between scale-outs

Alarm:
  Period: 60  # 1 min
  EvaluationPeriods: 2  # 2 min total

# Timeline:
# t=0: Load spike
# t=2: Alarm triggers
# t=7: Instance ready (5 min boot)
# Result: 7 minutes to handle spike (3× faster)
```

### Mistake 2: Aggressive Scale-In

```yaml
# ❌ Wrong - Aggressive scale-in (flapping)
ScaleInCooldown: 60  # 1 min
LowCPUAlarm:
  Period: 60
  EvaluationPeriods: 1
  Threshold: 40

# Behavior:
# t=0: CPU drops to 35%, scale in
# t=5: CPU rises to 75% (fewer instances), scale out
# t=10: CPU drops to 35%, scale in
# Constant thrashing, poor user experience

# ✓ Correct - Conservative scale-in
ScaleInCooldown: 300  # 5 min
LowCPUAlarm:
  Period: 300  # 5 min
  EvaluationPeriods: 3  # 15 min total
  Threshold: 30  # Lower threshold

# Scales in slowly, avoids thrashing
```

### Mistake 3: Wrong Metric

```yaml
# ❌ Wrong - CPU metric for I/O-bound app
# App makes slow database queries
# CPU: 15% (waiting on DB)
# Latency: 5 seconds
# Auto Scaling: "CPU low, no need to scale"

# ✓ Correct - Latency metric
TargetTrackingConfiguration:
  PredefinedMetricSpecification:
    PredefinedMetricType: ALBTargetResponseTime
  TargetValue: 1.0  # 1 second target

# Or custom metric (queue depth, active connections, etc.)
```

### Mistake 4: No Warm Pool or Lifecycle Hooks

```yaml
# ❌ Wrong - Cold start every scale-out
# Instance launches (2 min)
# App downloads dependencies (1 min)
# App builds cache (2 min)
# Total: 5 minutes to serve traffic

# ✓ Correct - Warm pool + lifecycle hook
WarmPoolConfiguration:
  MinSize: 2  # Pre-warmed instances
  PoolState: Stopped

LifecycleHook:
  # Pre-download dependencies
  # Pre-build cache
  # Ready in 30 seconds
```

## Hands-On Exercise 1: Design Scaling Strategy

**Scenario:** E-commerce API

**Traffic patterns:**
- Weekday 9 AM - 5 PM: 5,000 req/sec
- Weekday evenings: 2,000 req/sec
- Weekday nights: 500 req/sec
- Weekends: 1,000 req/sec
- Black Friday: 20,000 req/sec (predictable spike)

**Application:**
- Each instance handles 500 req/sec at 70% CPU
- Boot time: 3 minutes
- Application startup: 1 minute

**Requirements:**
- Handle traffic without errors
- Minimize cost
- Handle Black Friday spike

Design the auto scaling configuration (ASG sizing, scaling policies, metrics, scheduled actions).

<details>
<summary>Solution</summary>

```yaml
AutoScalingGroup:
  MinSize: 2  # HA minimum
  MaxSize: 50  # Protection (20K req/sec ÷ 500 = 40 instances + buffer)
  DesiredCapacity: 4  # Start conservatively

# ✓ Target tracking - Primary scaling
TargetTrackingPolicy:
  PolicyType: TargetTrackingScaling
  TargetTrackingConfiguration:
    PredefinedMetricSpecification:
      PredefinedMetricType: ALBRequestCountPerTarget
      ResourceLabel: !Sub "${ALB}/${TargetGroup}"
    TargetValue: 350.0  # 500 req/sec capacity × 0.7 = 350 target
    ScaleOutCooldown: 60  # Fast scale-out
    ScaleInCooldown: 300  # Slow scale-in

# ✓ Scheduled scaling - Predictable patterns

# Morning scale-up (before business hours)
MorningScaleUp:
  Type: AWS::AutoScaling::ScheduledAction
  Properties:
    DesiredCapacity: 10  # 5K req/sec ÷ 500 = 10 instances
    MinSize: 10
    MaxSize: 50
    Recurrence: "0 8 * * MON-FRI"  # 8 AM weekdays

# Evening scale-down
EveningScaleDown:
  Type: AWS::AutoScaling::ScheduledAction
  Properties:
    DesiredCapacity: 4  # 2K req/sec ÷ 500 = 4 instances
    MinSize: 2
    MaxSize: 20
    Recurrence: "0 17 * * MON-FRI"  # 5 PM weekdays

# Night scale-down
NightScaleDown:
  Type: AWS::AutoScaling::ScheduledAction
  Properties:
    DesiredCapacity: 2  # 500 req/sec ÷ 500 = 1, but 2 for HA
    MinSize: 2
    MaxSize: 10
    Recurrence: "0 22 * * *"  # 10 PM daily

# Weekend scale
WeekendScale:
  Type: AWS::AutoScaling::ScheduledAction
  Properties:
    DesiredCapacity: 2  # 1K req/sec ÷ 500 = 2 instances
    MinSize: 2
    MaxSize: 10
    Recurrence: "0 0 * * SAT"  # Midnight Saturday

# Black Friday preparation (day before)
BlackFridayPrep:
  Type: AWS::AutoScaling::ScheduledAction
  Properties:
    DesiredCapacity: 40  # 20K req/sec ÷ 500 = 40 instances
    MinSize: 40  # Prevent scale-in during event
    MaxSize: 50
    # Nov 24, 2023, midnight (adjust yearly)
    StartTime: "2023-11-24T00:00:00Z"

# Black Friday end
BlackFridayEnd:
  Type: AWS::AutoScaling::ScheduledAction
  Properties:
    DesiredCapacity: 10
    MinSize: 2
    MaxSize: 50
    StartTime: "2023-11-25T00:00:00Z"

# ✓ Warm pool for fast scale-out
WarmPoolConfiguration:
  MinSize: 5  # Keep 5 instances warm
  PoolState: Stopped  # Save cost
  InstanceReusePolicy:
    ReuseOnScaleIn: true

# ✓ Lifecycle hook to warm up app
LaunchLifecycleHook:
  LifecycleTransition: autoscaling:EC2_INSTANCE_LAUNCHING
  HeartbeatTimeout: 300
  # Lambda pre-downloads dependencies, warms cache
```

**Cost calculation:**

```
Without scheduled scaling (always run for peak):
  10 instances × 24 hours × 30 days × $0.05/hr = $360/month

With scheduled scaling:
  Weekday business hours (9 AM - 5 PM): 10 instances × 8 hr × 22 days = 1,760 hrs
  Weekday evenings (5 PM - 10 PM): 4 instances × 5 hr × 22 days = 440 hrs
  Weekday nights (10 PM - 9 AM): 2 instances × 11 hr × 22 days = 484 hrs
  Weekends: 2 instances × 48 hr × 8 days = 768 hrs
  Total: 3,452 hrs × $0.05/hr = $173/month

Savings: $187/month (52%)
```

**Black Friday handling:**
- Scheduled action scales to 40 instances at midnight
- Warm pool provides 5 instances immediately if spike comes early
- Target tracking adds capacity if needed (up to max 50)
- Result: Ready for 20K req/sec with no manual intervention

</details>

## Hands-On Exercise 2: Debug Slow Scaling

**Problem:** Auto scaling takes 8-10 minutes to respond to traffic spikes, causing 503 errors.

**Current configuration:**

```yaml
AutoScalingGroup:
  MinSize: 2
  MaxSize: 10
  HealthCheckGracePeriod: 600  # 10 minutes

ScalingPolicy:
  TargetValue: 70
  ScaleOutCooldown: 300

HighCPUAlarm:
  MetricName: CPUUtilization
  Period: 300  # 5 minutes
  EvaluationPeriods: 2  # 10 minutes total
  Threshold: 70

LaunchTemplate:
  UserData: |
    #!/bin/bash
    yum update -y
    yum install -y docker
    docker pull my-app:latest
    docker run -d -p 80:3000 my-app:latest
```

Identify all issues and fix them.

<details>
<summary>Solution</summary>

**Issues identified:**

1. ❌ Alarm takes 10 minutes to trigger (Period × EvaluationPeriods)
2. ❌ Long health check grace period (10 minutes)
3. ❌ Docker pull in UserData (slow, unpredictable)
4. ❌ No warm pool (cold start every time)
5. ❌ No signal to Auto Scaling when ready

**Fixed configuration:**

```yaml
AutoScalingGroup:
  MinSize: 2
  MaxSize: 10
  # ✓ Reduce grace period
  HealthCheckGracePeriod: 120  # 2 minutes (enough for app startup)

  # ✓ Add warm pool
  WarmPoolConfiguration:
    MinSize: 2  # Keep 2 instances warm
    PoolState: Stopped
    InstanceReusePolicy:
      ReuseOnScaleIn: true

ScalingPolicy:
  TargetValue: 70
  # ✓ Faster scale-out
  ScaleOutCooldown: 60  # 1 minute instead of 5

HighCPUAlarm:
  MetricName: CPUUtilization
  # ✓ Faster alarm
  Period: 60  # 1 minute instead of 5
  EvaluationPeriods: 2  # 2 minutes total instead of 10
  Threshold: 70

LaunchTemplate:
  # ✓ Pre-baked AMI with Docker image
  ImageId: ami-with-preinstalled-app  # Custom AMI with app pre-loaded

  UserData: |
    #!/bin/bash
    # Just start the app (already installed in AMI)
    docker run -d -p 80:3000 my-app:latest

    # ✓ Signal Auto Scaling when ready
    /opt/aws/bin/cfn-signal -e $? \
      --stack ${AWS::StackName} \
      --resource AutoScalingGroup \
      --region ${AWS::Region}

# ✓ Lifecycle hook for final warmup
LifecycleHook:
  LifecycleTransition: autoscaling:EC2_INSTANCE_LAUNCHING
  HeartbeatTimeout: 120  # 2 minutes
  # Lambda handler warms up app (makes health check requests)
```

**Timeline comparison:**

```
Before (8-10 minutes):
  t=0: Traffic spike, CPU rises
  t=10: Alarm triggers (2 periods × 5 min)
  t=10: Instance launches
  t=12: OS boots
  t=13: Docker pull (1 min)
  t=14: App starts
  t=24: Health check grace period ends
  Total: ~14-24 minutes

After (2-3 minutes):
  t=0: Traffic spike
  t=2: Alarm triggers (2 periods × 1 min)
  t=2: Warm pool instance starts (stopped → running: 30 sec)
  t=2.5: Instance running
  t=2.5: Lifecycle hook warms app (30 sec)
  t=3: Instance InService
  Total: ~3 minutes

Improvement: 4-8× faster
```

**Additional optimizations:**

```yaml
# ✓ Use larger instance types (faster boot)
InstanceType: c5.large  # Instead of t3.micro

# ✓ Enable detailed monitoring (1-min metrics)
Monitoring:
  Enabled: true

# ✓ Use predictive scaling (AWS predicts traffic)
PredictiveScalingConfiguration:
  MetricSpecifications:
    - TargetValue: 70
      PredefinedMetricPairSpecification:
        PredefinedMetricType: ASGCPUUtilization
  Mode: ForecastAndScale  # Pre-scale before spike
```

</details>

## Interview Questions

### Q1: What's the difference between target tracking and step scaling, and when would you use each?

Tests understanding of scaling policy types and when to apply each. Shows whether you've tuned scaling policies in production.

<details>
<summary>Answer</summary>

**Target Tracking:**
- Maintains specific metric value (e.g., CPU at 70%)
- Auto Scaling calculates needed capacity automatically
- Simpler configuration (just set target)
- Best for most use cases

**Step Scaling:**
- Scale by different amounts based on thresholds
- More control over scaling behavior
- More complex configuration
- Best for specific requirements

**Example:**

```yaml
# Target Tracking (simple)
TargetTrackingConfiguration:
  PredefinedMetricType: ASGAverageCPUUtilization
  TargetValue: 70.0
  # Auto Scaling figures out: "Need 5 instances to maintain 70% CPU"

# Step Scaling (granular control)
StepAdjustments:
  - MetricIntervalLowerBound: 0
    MetricIntervalUpperBound: 10
    ScalingAdjustment: 1  # CPU 70-80%: Add 1 instance

  - MetricIntervalLowerBound: 10
    MetricIntervalUpperBound: 20
    ScalingAdjustment: 2  # CPU 80-90%: Add 2 instances

  - MetricIntervalLowerBound: 20
    ScalingAdjustment: 5  # CPU >90%: Add 5 instances (emergency)
```

**Use target tracking when:**
- Standard use case (CPU, requests, queue depth)
- Want simple configuration
- Metric correlates linearly with capacity needed
- Don't need fine-grained control

**Use step scaling when:**
- Need aggressive scale-out for critical spikes
- Different scaling rates for different thresholds
- Complex business logic (e.g., scale differently based on time of day)
- Combining multiple metrics (one policy per metric)

**Recommendation:** Start with target tracking (90% of use cases), switch to step only if needed.

</details>

### Q2: How would you scale an application that processes messages from an SQS queue?

Tests understanding of queue-based scaling patterns. Shows whether you know how to use backlog metrics for scaling.

<details>
<summary>Answer</summary>

**Best metric: Queue depth (backlog per instance)**

```yaml
# Calculate target queue depth:
# 1. Measure throughput: Instance processes 10 messages/sec
# 2. Desired lag: Process backlog in 5 minutes
# 3. Target: 10 msg/sec × 300 sec = 3,000 messages per instance

TargetTrackingConfiguration:
  CustomizedMetricSpecification:
    MetricName: BacklogPerInstance
    Namespace: MyApp
    Statistic: Average
  TargetValue: 3000.0

# Custom metric calculation (CloudWatch Math):
BacklogPerInstance = ApproximateNumberOfMessagesVisible / DesiredCapacity

# Alternative: Use built-in metric
TargetTrackingConfiguration:
  CustomizedMetricSpecification:
    MetricName: ApproximateNumberOfMessagesVisible
    Namespace: AWS/SQS
    Dimensions:
      - Name: QueueName
        Value: !GetAtt MyQueue.QueueName
  TargetValue: 3000.0  # Assumes 1 instance initially
```

**Why not CPU?**
```
CPU-based scaling:
  - Queue has 10,000 messages
  - 1 instance at 100% CPU
  - Auto Scaling: "CPU high, add instances"
  - 2 instances at 100% CPU
  - Still backlog of 10,000 messages
  - Doesn't scale enough

Queue depth scaling:
  - Queue has 10,000 messages
  - Target: 3,000 per instance
  - Needed: 10,000 ÷ 3,000 = 4 instances
  - Scales correctly to clear backlog
```

**Implementation:**

```javascript
// Worker application
const AWS = require('aws-sdk');
const sqs = new AWS.SQS();

const processMessages = async () => {
  while (true) {
    const messages = await sqs.receiveMessage({
      QueueUrl: process.env.QUEUE_URL,
      MaxNumberOfMessages: 10,
      WaitTimeSeconds: 20  // Long polling
    }).promise();

    if (!messages.Messages) continue;

    for (const message of messages.Messages) {
      try {
        await processMessage(message);

        await sqs.deleteMessage({
          QueueUrl: process.env.QUEUE_URL,
          ReceiptHandle: message.ReceiptHandle
        }).promise();
      } catch (error) {
        console.error('Failed to process:', error);
        // Message returns to queue for retry
      }
    }
  }
};
```

**Benefits:**
- Scales based on actual work to be done
- Prevents queue from growing unbounded
- Automatic backpressure
- Cost-efficient (only scale when needed)

</details>

### Q3: What are warm pools and when should you use them?

Tests understanding of advanced scaling features. Shows whether you know how to optimize scale-out time.

<details>
<summary>Answer</summary>

**Warm pools:** Pre-initialized EC2 instances ready to join the Auto Scaling Group quickly.

**Problem they solve:**

```
Without warm pool:
  t=0: Scale-out triggered
  t=0-2: EC2 instance launches
  t=2-3: OS boots
  t=3-4: Application starts
  t=4-6: Application warms up (download deps, build cache)
  t=6: Instance ready
  Total: 6 minutes

With warm pool:
  t=0: Scale-out triggered
  t=0: Warm instance starts (already initialized)
  t=0.5: Instance running (stopped → running = 30 sec)
  t=0.5: Instance ready
  Total: 30 seconds

12× faster scale-out
```

**Configuration:**

```yaml
WarmPoolConfiguration:
  MinSize: 5  # Number of instances to keep warm
  MaxGroupPreparedCapacity: 20  # Max (warm + in-service)
  PoolState: Stopped  # or Running, Hibernated
  InstanceReusePolicy:
    ReuseOnScaleIn: true  # Return instances to pool on scale-in
```

**Pool states:**

```
Stopped:
  - No compute charge (only EBS storage)
  - Start time: ~30 seconds
  - Cost: $10-20/month per instance (EBS only)
  - Best for: Cost-sensitive, can tolerate 30 sec startup

Running:
  - Full compute charge
  - Start time: Instant
  - Cost: Same as in-service instance
  - Best for: Ultra-low latency requirements

Hibernated:
  - No compute charge
  - Start time: ~1 minute (faster than Stopped)
  - Cost: EBS + RAM snapshot storage
  - Best for: Balance between cost and speed
```

**When to use warm pools:**

✓ **Long application startup time**
```
App startup: >2 minutes (download deps, build cache)
Benefit: Pre-initialize in warm pool
```

✓ **Unpredictable traffic spikes**
```
Traffic can spike 10× in seconds (viral posts, flash sales)
Benefit: Instant capacity available
```

✓ **Cost-sensitive with occasional spikes**
```
Usually 2 instances, spike to 20 instances
Benefit: Keep 5 in warm pool (Stopped state)
Cost: Minimal ($50/month for 5 stopped instances)
```

❌ **When NOT to use:**

```
- Fast application startup (<1 min)
- Predictable traffic (use scheduled scaling instead)
- Very high scale (warm pool limited to hundreds of instances)
```

**Cost comparison:**

```
Scenario: Need to scale from 2 → 20 instances during spike

Without warm pool:
  - Scale-out time: 5 minutes
  - 503 errors during scale-out
  - Lost revenue: Unknown

With warm pool (5 stopped instances):
  - Warm pool cost: 5 × $10/month = $50/month
  - Scale-out time: 30 seconds (for first 5), then gradual
  - Reduced 503 errors
  - ROI: If prevents one major incident, easily worth it
```

</details>

### Q4: How would you handle a Black Friday traffic spike that's 10× normal load?

Tests practical scaling knowledge and whether you understand combining multiple scaling strategies. Shows ability to plan for extreme events.

<details>
<summary>Answer</summary>

**Multi-layered approach:**

**1. Scheduled Scaling (Pre-scale)**

```yaml
# Start scaling the night before
BlackFridayPrep:
  Type: AWS::AutoScaling::ScheduledAction
  Properties:
    DesiredCapacity: 80  # 10× normal (8 instances)
    MinSize: 80  # Prevent auto scale-in during event
    MaxSize: 100
    StartTime: "2023-11-23T23:00:00Z"  # 11 PM night before

# Gradual ramp-down after event
BlackFridayEnd:
  StartTime: "2023-11-24T23:59:00Z"  # Midnight after
  DesiredCapacity: 20
  MinSize: 2
  MaxSize: 100
```

**2. Target Tracking (Handle unexpected spikes beyond 10×)**

```yaml
# Still use target tracking as safety net
TargetTrackingPolicy:
  TargetValue: 70  # CPU target
  ScaleOutCooldown: 60  # Fast scale-out
  # Can scale to Max: 100 if needed
```

**3. Warm Pool (Instant capacity)**

```yaml
WarmPoolConfiguration:
  MinSize: 20  # Extra 20 instances ready
  PoolState: Running  # Keep running for instant availability
  # Cost during event: Worth it for zero-downtime
```

**4. CloudFront Caching (Reduce origin load)**

```yaml
# Enable aggressive caching
CacheBehavior:
  DefaultTTL: 60  # Cache product pages for 1 minute
  # 90% cache hit rate = 10× load → 1× origin load
  # Effective load: Same as normal day
```

**5. Load Testing (Pre-event validation)**

```bash
# Week before: Load test at 15× capacity
artillery run --count 100 --num 1000 load-test.yml

# Verify:
# - All instances handle load
# - No errors at 15× traffic
# - Latency stays below SLA
# - Database can handle load
```

**6. Monitoring & Alerts**

```yaml
# Alert on key metrics
HighLatencyAlarm:
  Threshold: 1000  # 1 second
  EvaluationPeriods: 1  # Immediate alert

High5xxAlarm:
  Threshold: 100  # 100 errors/min
  EvaluationPeriods: 1

NearMaxCapacityAlarm:
  MetricName: GroupInServiceInstances
  Threshold: 90  # Alert at 90 instances (near 100 max)
```

**7. Database Scaling**

```yaml
# Pre-scale RDS
# Modify instance class: db.t3.large → db.r5.4xlarge
# Increase read replicas: 2 → 10

# Or use Aurora Serverless Auto Scaling
ServerlessScalingConfiguration:
  MinCapacity: 8
  MaxCapacity: 128  # 16× normal
```

**8. Runbook & Team**

```
Pre-event checklist:
  □ Scheduled scaling configured
  □ Warm pool running
  □ CloudFront cache optimized
  □ Database scaled
  □ Load testing completed
  □ Monitoring dashboards ready
  □ On-call team briefed
  □ Rollback plan documented

During event:
  - Monitor dashboard every 15 min
  - Be ready to manually scale if needed
  - Have database admin on standby
  - Document any issues for post-mortem
```

**Timeline:**

```
T-7 days: Load testing
T-3 days: Final configuration review
T-1 day: Pre-scale databases
T-12 hours: Scheduled scaling to 80 instances
T-1 hour: Team monitoring begins
T=0: Black Friday starts
  - 80 instances handling 10× load
  - CloudFront absorbs 90% of requests
  - Effective origin load: Normal
  - Zero errors
T+24 hours: Gradual scale-down
T+48 hours: Back to normal capacity
```

**Cost:**

```
Normal day: 8 instances × 24 hr = 192 instance-hours × $0.05 = $9.60

Black Friday:
  Pre-scale: 80 instances × 24 hr = 1,920 instance-hours × $0.05 = $96
  Warm pool: 20 instances × 24 hr = 480 instance-hours × $0.05 = $24
  Database: $200 (temporary scale-up)
  Total: $320

Extra cost: $310 for the day

Revenue impact if site goes down:
  Lost sales: $50,000 - $500,000
  ROI on scaling: 160× - 1,600×
```

</details>

## Key Takeaways

1. **Policy Types**: Target tracking (simple, 90% of cases), step scaling (fine control), scheduled (predictable patterns)
2. **Metrics**: ALB requests (web apps), CPU (compute-bound), queue depth (workers), custom (specific needs)
3. **Scale Out vs In**: Fast scale-out (60 sec cooldown), slow scale-in (300 sec cooldown) prevents thrashing
4. **Warm Pools**: Pre-initialized instances reduce scale-out from 5 min to 30 sec (12× faster)
5. **Lifecycle Hooks**: Execute custom logic during launch/termination (warmup, deregistration, backup)
6. **Scheduled Scaling**: Pre-scale for predictable patterns (business hours, Black Friday)
7. **Common Mistakes**: Slow alarm periods, aggressive scale-in, wrong metrics, no warm pool
8. **Cost Optimization**: Scheduled scaling saves 50%+ vs running for peak capacity 24/7

## Next Steps

In [Lesson 06: Lambda at Scale](lesson-06-lambda-at-scale.md), you'll learn:

- Concurrency limits and reservation strategies
- Cold start mitigation techniques
- VPC Lambda performance implications
- Event source mappings (SQS, Kinesis, DynamoDB Streams)
- Cost vs performance trade-offs for serverless at scale
