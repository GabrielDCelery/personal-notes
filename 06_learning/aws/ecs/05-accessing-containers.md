# Accessing Running Containers in ECS

## Overview

You cannot use traditional SSH to access ECS containers. Instead, AWS provides **ECS Exec**, which works like `docker exec`.

## ECS Exec (Recommended)

Works for both **Fargate** and **EC2** launch types.

### Step 1: Enable ECS Exec on the Service

```bash
aws ecs update-service \
  --cluster my-cluster \
  --service my-service \
  --enable-execute-command
```

### Step 2: Configure IAM Permissions

Your task IAM role needs these permissions:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ssmmessages:CreateControlChannel",
        "ssmmessages:CreateDataChannel",
        "ssmmessages:OpenControlChannel",
        "ssmmessages:OpenDataChannel"
      ],
      "Resource": "*"
    }
  ]
}
```

### Step 3: Install Session Manager Plugin

```bash
# Ubuntu/Debian
curl "https://s3.amazonaws.com/session-manager-downloads/plugin/latest/ubuntu_64bit/session-manager-plugin.deb" -o "session-manager-plugin.deb"
sudo dpkg -i session-manager-plugin.deb

# macOS
brew install --cask session-manager-plugin
```

### Step 4: Connect to Container

```bash
# List running tasks
aws ecs list-tasks --cluster my-cluster --service my-service

# Execute interactive shell
aws ecs execute-command \
  --cluster my-cluster \
  --task <task-id> \
  --container my-container \
  --interactive \
  --command "/bin/sh"
```

## Enable ECS Exec in Infrastructure as Code

### AWS CLI (Run Task)

```bash
aws ecs run-task \
  --cluster my-cluster \
  --task-definition my-task:1 \
  --enable-execute-command
```

### CDK (TypeScript)

```typescript
const taskDefinition = new ecs.FargateTaskDefinition(this, "Task", {
  // ...
});

const service = new ecs.FargateService(this, "Service", {
  cluster,
  taskDefinition,
  enableExecuteCommand: true, // Enable ECS Exec
});
```

### Terraform

```hcl
resource "aws_ecs_service" "app" {
  name            = "my-service"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.app.arn

  enable_execute_command = true
}
```

## Alternative: EC2 Launch Type Only

If using **EC2 launch type**, you can SSH to the EC2 instance, then use Docker:

```bash
# SSH to EC2 instance
ssh ec2-user@<ec2-ip>

# List containers
docker ps

# Exec into container
docker exec -it <container-id> /bin/sh
```

## Comparison Table

| Method                | Launch Type              | Use Case                          | Setup                                |
| --------------------- | ------------------------ | --------------------------------- | ------------------------------------ |
| **ECS Exec**          | Fargate, EC2             | Modern, recommended               | Requires IAM permissions, SSM plugin |
| **SSH → Docker exec** | EC2 only                 | Legacy, when ECS Exec unavailable | Requires SSH access to EC2           |
| **Traditional SSH**   | Either (not recommended) | Anti-pattern                      | Requires SSH server in container     |

## Debugging Without Shell Access

If ECS Exec isn't available:

```bash
# View logs
aws logs tail /ecs/my-service --follow

# Describe task for status and errors
aws ecs describe-tasks --cluster my-cluster --tasks <task-id>

# View CloudWatch logs
aws logs get-log-events \
  --log-group-name /ecs/my-service \
  --log-stream-name ecs/my-container/<task-id>
```

## Docker Equivalent

ECS Exec is the AWS equivalent of:

```bash
docker exec -it <container-id> /bin/sh
```

## Common Issues

### "ExecuteCommandAgent not running"

The container must be running for a few seconds before ECS Exec is available. Wait and retry.

### "Session Manager plugin not found"

Install the Session Manager plugin (see Step 3 above).

### "User not authorized"

Ensure your IAM user/role has `ecs:ExecuteCommand` permission:

```json
{
  "Effect": "Allow",
  "Action": "ecs:ExecuteCommand",
  "Resource": "*"
}
```

## Best Practices

- Enable ECS Exec only when needed (security consideration)
- Use CloudWatch Logs for routine debugging
- Prefer logging and metrics over interactive debugging
- Ensure your container image has a shell (`/bin/sh` or `/bin/bash`)
