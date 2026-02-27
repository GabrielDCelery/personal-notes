# Volume Sharing and Replication in ECS

## Overview

Understanding how volumes work across tasks vs within tasks is critical for ECS storage architecture. This document clarifies the difference between volume sharing **within a task** and volume isolation **across tasks**.

## The Two Levels of Volume Sharing

### Level 1: Within a Task (Shared)

**All containers in the same task can share volumes.**

```
Task 1:
  ├── Container A ─┐
  └── Container B ─┼─→ Volume (shared)
```

### Level 2: Across Tasks (Isolated for EBS, Shared for EFS)

**EBS volumes are NOT shared across tasks:**

```
Task 1 → EBS Volume 1 (isolated)
Task 2 → EBS Volume 2 (isolated)
Task 3 → EBS Volume 3 (isolated)
```

**EFS volumes ARE shared across tasks:**

```
Task 1 ─┐
Task 2 ─┼─→ EFS File System (shared)
Task 3 ─┘
```

## EBS Volumes with Multiple Replicas

### The Critical Concept

When you have a service with `desiredCount: 3` and EBS volumes configured, you get **3 separate EBS volumes** - one per task.

### Example Service with 3 Replicas

```json
{
  "serviceName": "my-service",
  "taskDefinition": "my-app:1",
  "desiredCount": 3,
  "volumeConfigurations": [
    {
      "name": "data-volume",
      "managedEBSVolume": {
        "size": 20,
        "volumeType": "gp3"
      }
    }
  ]
}
```

### What Actually Happens

```
Service (desiredCount: 3)
  ├── Task 1 → EBS Volume 1 (20 GB, independent)
  ├── Task 2 → EBS Volume 2 (20 GB, independent)
  └── Task 3 → EBS Volume 3 (20 GB, independent)

Total storage provisioned: 60 GB
```

### Key Properties of EBS Volumes

- **Single-attach**: An EBS volume can only be attached to one task at a time
- **Isolated**: Data written to Task 1's volume is NOT visible to Task 2 or Task 3
- **Not replicated**: Each volume starts empty (or from a snapshot if configured)
- **Independent lifecycle**: When a task stops, its EBS volume can be deleted or reused

## Volume Sharing Within a Task

Containers in the same task CAN share volumes by mounting the same volume at different paths.

### Task Definition Example

```json
{
  "family": "shared-volume-example",
  "volumes": [
    {
      "name": "shared-data",
      "configuredAtLaunch": true // EBS placeholder
    }
  ],
  "containerDefinitions": [
    {
      "name": "app",
      "image": "my-app:latest",
      "mountPoints": [
        {
          "sourceVolume": "shared-data",
          "containerPath": "/app/data"
        }
      ]
    },
    {
      "name": "sidecar",
      "image": "log-processor:latest",
      "mountPoints": [
        {
          "sourceVolume": "shared-data",
          "containerPath": "/logs" // Same volume, different path
        }
      ]
    }
  ]
}
```

### What Happens

```
Task 1:
  ├── App Container
  │   └── Mounts: /app/data → EBS Volume 1
  └── Sidecar Container
      └── Mounts: /logs → EBS Volume 1 (same volume!)

Both containers see the same data in real-time
```

### Common Use Cases for Shared Volumes Within Tasks

1. **Log aggregation**: App writes logs, sidecar ships them
2. **Shared cache**: Multiple containers read/write cache files
3. **Data processing pipeline**: One container produces, another consumes
4. **Init container pattern**: Init container prepares config, app reads it

## Volume Configuration Levels

EBS volume configuration is split across two levels, allowing flexibility in deployment.

### Task Definition Level: Declare the Volume

```json
{
  "family": "my-task",
  "volumes": [
    {
      "name": "my-ebs-volume",
      "configuredAtLaunch": true // Placeholder: "I need a volume"
    },
    {
      "name": "my-efs-volume",
      "efsVolumeConfiguration": {
        // Fully configured here
        "fileSystemId": "fs-12345"
      }
    }
  ]
}
```

### Service Level: Configure EBS Specifics

```json
{
  "serviceName": "my-service",
  "taskDefinition": "my-task:1",
  "volumeConfigurations": [
    {
      "name": "my-ebs-volume", // Matches volume name from task definition
      "managedEBSVolume": {
        "size": 20,
        "volumeType": "gp3",
        "iops": 3000,
        "throughput": 125,
        "encrypted": true,
        "fileSystemType": "ext4"
      }
    }
  ]
}
```

### Why the Split?

This allows you to:

- Reuse the same task definition with different EBS sizes per environment
- Override EBS settings at deployment time without updating the task definition
- Use different storage configs for dev/staging/prod with the same task definition

**Note**: EFS and bind mounts are fully configured in the task definition only.

## EBS vs EFS: Sharing Comparison

### EBS (Isolated Per Task)

```json
{
  "volumes": [
    {
      "name": "data",
      "configuredAtLaunch": true
    }
  ]
}
```

**Storage model:**

```
Service (3 replicas)
  Task 1 → EBS Vol 1 (20 GB) ← Isolated
  Task 2 → EBS Vol 2 (20 GB) ← Isolated
  Task 3 → EBS Vol 3 (20 GB) ← Isolated
```

**Characteristics:**

- ❌ Cannot share across tasks
- ✅ Can share within task (multiple containers)
- ✅ High performance (local block storage)
- ❌ Data is isolated per task
- 💰 Cost: 3 × volume size

### EFS (Shared Across All Tasks)

```json
{
  "volumes": [
    {
      "name": "data",
      "efsVolumeConfiguration": {
        "fileSystemId": "fs-12345"
      }
    }
  ]
}
```

**Storage model:**

```
Service (3 replicas)
  Task 1 ─┐
  Task 2 ─┼─→ EFS (20 GB) ← Shared
  Task 3 ─┘
```

**Characteristics:**

- ✅ Can share across tasks
- ✅ Can share within task (multiple containers)
- ✅ Works with Fargate and EC2
- ✅ All tasks see the same data
- 💰 Cost: 1 × storage size (plus per-request fees)

## Decision Matrix

### Use EBS When:

- Task needs **isolated, dedicated storage**
- High IOPS/throughput required
- Running stateful workloads where each replica has its own state
- Running on EC2 (not supported on Fargate for older versions)

### Use EFS When:

- Tasks need **shared access** to the same files
- User uploads, shared configuration, or collaborative data
- Running on Fargate (most common persistent storage option)
- Multi-AZ deployments (EFS spans availability zones)

## Practical Examples

### Example 1: Web App with 3 Replicas (EBS)

```json
{
  "serviceName": "web-app",
  "desiredCount": 3,
  "volumeConfigurations": [
    {
      "name": "cache-volume",
      "managedEBSVolume": {
        "size": 10
      }
    }
  ]
}
```

**Result:**

```
Task 1: /app/cache → EBS Vol 1 (10 GB, Task 1's local cache)
Task 2: /app/cache → EBS Vol 2 (10 GB, Task 2's local cache)
Task 3: /app/cache → EBS Vol 3 (10 GB, Task 3's local cache)

Each task has its own isolated cache
Total storage: 30 GB
```

### Example 2: Web App with Shared Uploads (EFS)

```json
{
  "serviceName": "web-app",
  "desiredCount": 3,
  "taskDefinition": {
    "volumes": [
      {
        "name": "uploads",
        "efsVolumeConfiguration": {
          "fileSystemId": "fs-abc123"
        }
      }
    ]
  }
}
```

**Result:**

```
Task 1: /app/uploads ─┐
Task 2: /app/uploads ─┼─→ EFS (shared upload directory)
Task 3: /app/uploads ─┘

All tasks see the same files
User uploads available to any task
Total storage: Size of EFS filesystem
```

### Example 3: Multi-Container Task with Shared Volume

```json
{
  "family": "nginx-php",
  "volumes": [
    {
      "name": "web-files",
      "efsVolumeConfiguration": {
        "fileSystemId": "fs-xyz789"
      }
    }
  ],
  "containerDefinitions": [
    {
      "name": "nginx",
      "mountPoints": [
        {
          "sourceVolume": "web-files",
          "containerPath": "/usr/share/nginx/html"
        }
      ]
    },
    {
      "name": "php-fpm",
      "mountPoints": [
        {
          "sourceVolume": "web-files",
          "containerPath": "/var/www/html"
        }
      ]
    }
  ]
}
```

**Result:**

```
Task:
  Nginx container: /usr/share/nginx/html ─┐
                                          ├─→ Same EFS volume
  PHP-FPM container: /var/www/html ───────┘

Both containers in the task access the same files
```

## Summary Table

| Aspect                   | EBS                     | EFS                      |
| ------------------------ | ----------------------- | ------------------------ |
| **Sharing within task**  | ✅ Yes                  | ✅ Yes                   |
| **Sharing across tasks** | ❌ No (isolated)        | ✅ Yes                   |
| **Fargate support**      | ✅ Yes (1.4.0+)         | ✅ Yes                   |
| **EC2 support**          | ✅ Yes                  | ✅ Yes                   |
| **Performance**          | High (local)            | Medium (network)         |
| **Configuration level**  | Task def + Service      | Task def only            |
| **Cost model**           | Per volume              | Per GB + requests        |
| **Use case**             | Isolated state per task | Shared data across tasks |

## Key Takeaways

1. **EBS volumes are NOT replicated across tasks** - each task gets its own isolated volume
2. **Containers within a task CAN share volumes** - mount the same volume at different paths
3. **EBS configuration is split** between task definition (declaration) and service (specifics)
4. **EFS is the solution for cross-task sharing** - all tasks access the same filesystem
5. **With 3 replicas and EBS**, you provision 3× the storage (3 independent volumes)
6. **With 3 replicas and EFS**, you provision 1× the storage (1 shared filesystem)
