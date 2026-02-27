# ECS Network Modes

## Overview

ECS supports four network modes that determine how containers connect to the network.

## Network Modes

| Mode       | Description                                            | Use Case                                       | IP Address                                   |
| ---------- | ------------------------------------------------------ | ---------------------------------------------- | -------------------------------------------- |
| **awsvpc** | Each task gets its own ENI (Elastic Network Interface) | **Fargate (required)**, modern EC2 deployments | Task gets unique private IP from VPC subnet  |
| **bridge** | Docker's default bridge network                        | Legacy EC2 tasks                               | Containers share host's IP, use port mapping |
| **host**   | Container uses host's network directly                 | High-performance networking on EC2             | Container shares host's IP and ports         |
| **none**   | No external networking                                 | Isolated tasks, batch jobs                     | No network access                            |

## awsvpc Mode (Recommended)

```json
{
  "family": "my-app",
  "networkMode": "awsvpc",
  "containerDefinitions": [
    {
      "portMappings": [
        { "containerPort": 80 } // No hostPort needed
      ]
    }
  ]
}
```

**Characteristics:**

- Task gets its own IP address
- Task gets its own security group
- Required for Fargate
- Better security isolation
- Supports AWS service discovery

## bridge Mode (Legacy)

```json
{
  "family": "my-app",
  "networkMode": "bridge",
  "containerDefinitions": [
    {
      "portMappings": [
        { "containerPort": 80, "hostPort": 8080 } // Must map ports
      ]
    }
  ]
}
```

**Characteristics:**

- Default for EC2 launch type
- Containers share host's network namespace
- Requires explicit port mapping
- Multiple tasks can run on same host with different ports

## host Mode

```json
{
  "family": "my-app",
  "networkMode": "host",
  "containerDefinitions": [
    {
      "portMappings": [
        { "containerPort": 80 } // Directly exposed on host
      ]
    }
  ]
}
```

**Characteristics:**

- Container bypasses Docker networking
- Maximum network performance
- No port mapping needed
- Only one task per host per port

## none Mode

```json
{
  "family": "my-app",
  "networkMode": "none",
  "containerDefinitions": [
    {
      // No port mappings
    }
  ]
}
```

**Characteristics:**

- No network interface
- Completely isolated
- Used for batch processing or security-sensitive workloads

## Recommendations

- **Use awsvpc** for all new deployments (Fargate requires it)
- **Use bridge** only for legacy EC2 tasks
- **Use host** only when you need maximum network performance
- **Use none** for isolated batch jobs
