# ECS Storage: EBS, EFS, and Docker Volumes

## Overview

In Docker, you use volumes for persistent storage. In ECS, you have several storage options depending on your launch type (Fargate or EC2).

## Docker Volumes vs ECS Storage

### Docker Compose Example

```yaml
services:
  app:
    volumes:
      - my-data:/var/lib/data # Named volume
      - ./local:/app # Bind mount
```

### ECS Storage Options

| Storage Type                  | Launch Type  | Persistence             | Use Case                          |
| ----------------------------- | ------------ | ----------------------- | --------------------------------- |
| **Ephemeral (Task Storage)**  | Fargate, EC2 | Task lifetime only      | Temporary files, caches           |
| **EFS (Elastic File System)** | Fargate, EC2 | Persistent, shared      | Shared data across tasks          |
| **EBS (Elastic Block Store)** | EC2 only     | Persistent, single host | Database files, dedicated storage |
| **Bind Mounts**               | EC2 only     | Host filesystem         | Access EC2 host files             |
| **Docker Volumes**            | EC2 only     | Persistent on host      | Local named volumes               |

## 1. Ephemeral Storage (Default)

Every task gets temporary storage that exists only during task lifetime.

### Characteristics

**Fargate:**

- 20 GB default (can configure up to 200 GB)
- Shared across all containers in the task
- Deleted when task stops

**EC2:**

- Uses host's disk space
- Deleted when task stops

### Task Definition

```json
{
  "family": "my-app",
  "containerDefinitions": [
    {
      "name": "app",
      "image": "myapp:latest",
      "mountPoints": [
        {
          "sourceVolume": "scratch-space",
          "containerPath": "/tmp/data"
        }
      ]
    }
  ],
  "volumes": [
    {
      "name": "scratch-space"
    }
  ]
}
```

### Docker Compose Equivalent

```yaml
services:
  app:
    volumes:
      - scratch-space:/tmp/data
volumes:
  scratch-space: # Anonymous/temporary volume
```

## 2. EFS (Elastic File System) - RECOMMENDED

**What**: Fully managed NFS file system that can be mounted by multiple tasks simultaneously.

### When to Use

- Shared persistent storage across multiple tasks
- File uploads, user-generated content
- Shared configuration files
- Works with **both Fargate and EC2**

### Advantages

- Survives task restarts
- Multiple tasks can read/write simultaneously
- Scales automatically
- Works across availability zones

### Task Definition

```json
{
  "family": "my-app",
  "containerDefinitions": [
    {
      "name": "app",
      "image": "myapp:latest",
      "mountPoints": [
        {
          "sourceVolume": "efs-storage",
          "containerPath": "/mnt/efs"
        }
      ]
    }
  ],
  "volumes": [
    {
      "name": "efs-storage",
      "efsVolumeConfiguration": {
        "fileSystemId": "fs-12345678",
        "rootDirectory": "/",
        "transitEncryption": "ENABLED",
        "authorizationConfig": {
          "iam": "ENABLED"
        }
      }
    }
  ]
}
```

### Docker Compose Equivalent

```yaml
services:
  app:
    volumes:
      - nfs-data:/mnt/efs
volumes:
  nfs-data:
    driver: local
    driver_opts:
      type: nfs
      o: addr=fs-12345678.efs.us-east-1.amazonaws.com
      device: ":/path"
```

### Setup Requirements

1. Create EFS file system in same VPC
2. Configure security groups to allow NFS traffic (port 2049)
3. Ensure task security group can reach EFS security group

### CDK Example

```typescript
import * as efs from "aws-cdk-lib/aws-efs";
import * as ecs from "aws-cdk-lib/aws-ecs";

// Create EFS
const fileSystem = new efs.FileSystem(this, "EFS", {
  vpc,
  encrypted: true,
  lifecyclePolicy: efs.LifecyclePolicy.AFTER_14_DAYS,
});

// Add to task definition
const volume = {
  name: "efs-storage",
  efsVolumeConfiguration: {
    fileSystemId: fileSystem.fileSystemId,
    transitEncryption: "ENABLED",
  },
};

taskDefinition.addVolume(volume);

container.addMountPoints({
  sourceVolume: "efs-storage",
  containerPath: "/mnt/efs",
  readOnly: false,
});
```

### EFS Access Points (Multi-Tenant Isolation)

```json
{
  "efsVolumeConfiguration": {
    "fileSystemId": "fs-12345678",
    "transitEncryption": "ENABLED",
    "authorizationConfig": {
      "accessPointId": "fsap-12345678",
      "iam": "ENABLED"
    }
  }
}
```

## 3. EBS (Elastic Block Store) - EC2 Only

**What**: Block-level storage volumes attached to EC2 instances.

### When to Use

- High-performance disk I/O
- Database storage (though RDS is usually better)
- **Only available with EC2 launch type**

### Limitations

- Can only be attached to one EC2 instance at a time
- Task must be pinned to specific EC2 instance
- Not available on Fargate
- Complex management

### Task Definition

```json
{
  "family": "my-app",
  "containerDefinitions": [
    {
      "name": "app",
      "mountPoints": [
        {
          "sourceVolume": "ebs-volume",
          "containerPath": "/data"
        }
      ]
    }
  ],
  "volumes": [
    {
      "name": "ebs-volume",
      "host": {
        "sourcePath": "/mnt/ebs-data"
      }
    }
  ]
}
```

**Note**: You must manually attach EBS to EC2 instance and mount it to `/mnt/ebs-data` first.

## 4. Bind Mounts (EC2 Only)

**What**: Mount a directory from the EC2 host into the container.

### When to Use

- Access files on EC2 instance
- Share data between containers on same host
- Docker socket access (`/var/run/docker.sock`)

### Task Definition

```json
{
  "family": "my-app",
  "containerDefinitions": [
    {
      "name": "app",
      "mountPoints": [
        {
          "sourceVolume": "host-logs",
          "containerPath": "/var/log/app"
        }
      ]
    }
  ],
  "volumes": [
    {
      "name": "host-logs",
      "host": {
        "sourcePath": "/var/log/ecs"
      }
    }
  ]
}
```

### Docker Compose Equivalent

```yaml
services:
  app:
    volumes:
      - /var/log/ecs:/var/log/app
```

## 5. Docker Volumes (EC2 Only)

**What**: Named Docker volumes managed by Docker on the EC2 host.

### Task Definition

```json
{
  "family": "my-app",
  "containerDefinitions": [
    {
      "name": "app",
      "mountPoints": [
        {
          "sourceVolume": "docker-volume",
          "containerPath": "/data"
        }
      ]
    }
  ],
  "volumes": [
    {
      "name": "docker-volume",
      "dockerVolumeConfiguration": {
        "scope": "shared",
        "autoprovision": true,
        "driver": "local"
      }
    }
  ]
}
```

### Docker Compose Equivalent

```yaml
services:
  app:
    volumes:
      - docker-volume:/data
volumes:
  docker-volume:
    driver: local
```

## Fargate vs EC2 Storage Comparison

### Fargate Storage Options

```
✓ Ephemeral storage (20-200 GB)
✓ EFS (persistent, shared)
✗ EBS (not supported)
✗ Bind mounts (no host access)
✗ Docker volumes (no host)
```

### EC2 Storage Options

```
✓ Ephemeral storage
✓ EFS (persistent, shared)
✓ EBS (persistent, single host)
✓ Bind mounts (host filesystem)
✓ Docker volumes (Docker-managed)
```

## Common Patterns

### Pattern 1: Shared Persistent Storage (Fargate)

```json
{
  "family": "web-app",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "256",
  "memory": "512",
  "containerDefinitions": [
    {
      "name": "nginx",
      "image": "nginx",
      "mountPoints": [
        {
          "sourceVolume": "uploads",
          "containerPath": "/usr/share/nginx/html/uploads"
        }
      ]
    }
  ],
  "volumes": [
    {
      "name": "uploads",
      "efsVolumeConfiguration": {
        "fileSystemId": "fs-12345678",
        "rootDirectory": "/uploads",
        "transitEncryption": "ENABLED"
      }
    }
  ]
}
```

### Pattern 2: Multi-Container with Shared Volume

**ECS Task Definition:**

```json
{
  "family": "nginx-php",
  "containerDefinitions": [
    {
      "name": "nginx",
      "image": "nginx",
      "mountPoints": [
        {
          "sourceVolume": "php-files",
          "containerPath": "/var/www/html"
        }
      ]
    },
    {
      "name": "php-fpm",
      "image": "php:fpm",
      "mountPoints": [
        {
          "sourceVolume": "php-files",
          "containerPath": "/var/www/html"
        }
      ]
    }
  ],
  "volumes": [
    {
      "name": "php-files",
      "efsVolumeConfiguration": {
        "fileSystemId": "fs-12345678"
      }
    }
  ]
}
```

**Docker Compose Equivalent:**

```yaml
services:
  nginx:
    image: nginx
    volumes:
      - php-files:/var/www/html
  php-fpm:
    image: php:fpm
    volumes:
      - php-files:/var/www/html
volumes:
  php-files:
```

### Pattern 3: Database on EC2 with EBS

```json
{
  "family": "postgres",
  "containerDefinitions": [
    {
      "name": "postgres",
      "image": "postgres:14",
      "mountPoints": [
        {
          "sourceVolume": "db-data",
          "containerPath": "/var/lib/postgresql/data"
        }
      ]
    }
  ],
  "volumes": [
    {
      "name": "db-data",
      "host": {
        "sourcePath": "/mnt/ebs-postgres"
      }
    }
  ]
}
```

## Decision Tree: Which Storage to Use?

```
Need persistent storage?
  │
  ├─ No → Use ephemeral storage (default)
  │
  └─ Yes → Multiple tasks need access?
      │
      ├─ Yes → Use EFS
      │
      └─ No → Using Fargate or EC2?
          │
          ├─ Fargate → Use EFS (only option)
          │
          └─ EC2 → High IOPS needed?
              │
              ├─ Yes → Use EBS
              │
              └─ No → Use EFS or Docker volumes
```

## Best Practices

### 1. Use EFS for Most Persistent Storage Needs

- Works with Fargate and EC2
- Shared across tasks
- Managed by AWS

### 2. Use Ephemeral Storage for Temporary Data

- Caches, build artifacts
- No persistence needed

### 3. Avoid EBS Unless Necessary

- Complex to manage
- Ties task to specific EC2 instance
- Consider RDS/DynamoDB instead for databases

### 4. Enable Encryption for Sensitive Data

```json
{
  "efsVolumeConfiguration": {
    "fileSystemId": "fs-12345678",
    "transitEncryption": "ENABLED" // Encrypts data in transit
  }
}
```

## Summary Table

| Docker Concept        | ECS Equivalent    | Fargate Support |
| --------------------- | ----------------- | --------------- |
| Anonymous volume      | Ephemeral storage | ✓               |
| Named volume (shared) | EFS               | ✓               |
| Named volume (local)  | Docker volume     | ✗ (EC2 only)    |
| Bind mount            | Host path         | ✗ (EC2 only)    |
| Volume driver (NFS)   | EFS               | ✓               |
| Volume driver (local) | EBS               | ✗ (EC2 only)    |

## Key Takeaways

1. **EFS is the recommended solution** for persistent storage in ECS
2. It works with both **Fargate and EC2**
3. **Multiple tasks can access the same EFS filesystem** simultaneously
4. **Ephemeral storage** is fine for temporary data that doesn't need to persist
5. **EBS is rarely needed** - use managed services (RDS, DynamoDB) for databases instead
6. Always enable **encryption in transit** for EFS when handling sensitive data
