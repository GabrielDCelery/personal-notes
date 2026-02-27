# AWS ECS Learning Guide

A comprehensive guide to understanding AWS Elastic Container Service (ECS) from a Docker perspective.

## Lessons

### [01. Container Definitions vs Docker Compose](./01-container-definitions-vs-docker-compose.md)

Learn how ECS container definitions map to Docker Compose services.

**Key Points:**

- Container definition = Single service in Docker Compose
- Describes one container: image, ports, env vars, resources
- Part of a larger task definition

### [02. Task Definitions vs Docker Compose](./02-task-definitions-vs-docker-compose.md)

Understand how ECS task definitions compare to complete Docker Compose files.

**Key Points:**

- Task definition ≈ docker-compose.yml file
- Groups multiple containers that run together
- Defines networking, volumes, and resources
- Used for production deployments on AWS

### [03. Network Modes](./03-network-modes.md)

Explore the four ECS network modes and when to use each.

**Key Points:**

- **awsvpc** (recommended): Each task gets its own IP and security group
- **bridge**: Legacy mode, shares host IP with port mapping
- **host**: Maximum performance, bypasses Docker networking
- **none**: No networking, for isolated tasks
- Fargate only supports awsvpc mode

### [04. Port Mappings](./04-port-mappings.md)

Learn how containerPort and hostPort work in ECS.

**Key Points:**

- **containerPort**: Port your app listens on inside the container
- **hostPort**: Port exposed on the host machine
- Docker format: `HOST:CONTAINER` (e.g., `8080:80`)
- ECS format: separate fields for same concept
- In awsvpc mode, hostPort is optional

### [05. Accessing Containers](./05-accessing-containers.md)

Discover how to access running ECS containers without SSH.

**Key Points:**

- Use **ECS Exec** instead of traditional SSH
- Works like `docker exec` for interactive shell access
- Requires IAM permissions and Session Manager plugin
- Available for both Fargate and EC2 launch types
- Alternative: SSH to EC2 then use Docker (EC2 only)

### [06. ECS Hierarchy](./06-ecs-hierarchy.md)

Understand the relationship between clusters, services, tasks, and containers.

**Key Points:**

- **Cluster**: Logical grouping of resources (like a data center)
- **Service**: Manages desired number of tasks (like a department ensuring workers)
- **Task Definition**: Blueprint/template (like a job description)
- **Task**: Running instance of task definition (like an actual worker)
- **Container Definition**: Specification for one container
- **Container**: The actual running workload

**Hierarchy:**

```
Cluster
  └── Service (manages multiple tasks)
       └── Task (running instance)
            └── Container Definition (specification)
                 └── Container (actual running container)
```

### [07. Storage: EBS, EFS, and Volumes](./07-storage-ebs-efs-volumes.md)

Learn about persistent storage options in ECS.

**Key Points:**

- **Ephemeral storage**: Temporary, deleted when task stops
- **EFS** (recommended): Persistent, shared across tasks, works with Fargate
- **EBS**: Persistent, single host, EC2 only
- **Bind mounts**: Access EC2 host filesystem, EC2 only
- **Docker volumes**: Docker-managed volumes, EC2 only

**Decision Tree:**

```
Need persistent storage?
  └─ Yes → Multiple tasks need access?
      └─ Yes → Use EFS (works with Fargate & EC2)
```

### [08. Load Balancing with ALB](./08-load-balancing-with-alb.md)

Understand how ECS integrates with Application Load Balancers.

**Key Points:**

- ECS **automatically registers/deregisters** task IPs with target groups
- Target type must be **ip** (not instance) for awsvpc mode
- Each task gets unique IP and is registered individually
- Health checks determine which tasks receive traffic
- Security groups control traffic flow from ALB to tasks
- Service Discovery (Cloud Map) is an alternative for internal communication

**Traffic Flow:**

```
Internet → ALB → Target Group (managed by ECS)
                    ├─→ Task 1 (10.0.1.50:80)
                    ├─→ Task 2 (10.0.1.51:80)
                    └─→ Task 3 (10.0.1.52:80)
```

### [09. Deployment Strategies and Load Balancing](./09-deployment-strategies-and-load-balancing.md)

Master blue/green deployments and load balancing algorithms.

**Key Points:**

- **Rolling update**: Gradual replacement, low complexity
- **Blue/Green** (CodeDeploy): True pre-production validation with test listener
- **Canary**: Gradual traffic shift (10%, 20%, etc.) with validation
- **Circuit breaker**: Automatic rollback on deployment failure
- **Round Robin**: Default algorithm, rotates through targets
- **Least Outstanding Requests**: Routes to least busy target (better for variable workloads)

**Blue/Green Flow:**

```
1. Blue (v1) serving production on port 80
2. Green (v2) deployed, accessible on port 8080 for testing
3. Tests pass, traffic shifted from port 80 to green
4. Blue (v1) tasks terminated
```

### [10. Volume Sharing and Replication](./10-volume-sharing-and-replication.md)

Understand how volumes are shared within tasks vs isolated across tasks.

**Key Points:**

- **EBS volumes are NOT shared across tasks** - each task gets its own isolated volume
- **Containers within a task CAN share volumes** - mount the same volume at different paths
- **EBS configuration is split** between task definition (declaration) and service (specifics)
- **EFS is the solution for cross-task sharing** - all tasks access the same filesystem
- **With 3 replicas and EBS**: 3 separate volumes (3× storage cost)
- **With 3 replicas and EFS**: 1 shared filesystem (1× storage cost)

**Storage Model:**

```
EBS (Isolated):
  Task 1 → EBS Vol 1 (isolated)
  Task 2 → EBS Vol 2 (isolated)
  Task 3 → EBS Vol 3 (isolated)

EFS (Shared):
  Task 1 ─┐
  Task 2 ─┼─→ EFS (shared)
  Task 3 ─┘
```

## Quick Reference

### Fargate vs EC2 Capabilities

| Feature            | Fargate        | EC2                                              |
| ------------------ | -------------- | ------------------------------------------------ |
| Network Mode       | awsvpc only    | awsvpc, bridge, host, none                       |
| Persistent Storage | Ephemeral, EFS | Ephemeral, EFS, EBS, bind mounts, Docker volumes |
| ECS Exec           | ✓              | ✓                                                |
| SSH Access         | ✗              | ✓ (to host)                                      |
| Custom AMI         | ✗              | ✓                                                |
| Management         | Fully managed  | Self-managed                                     |

### Common Patterns

#### Simple Web Application

```
Cluster: prod
  └── Service: web (desired: 3)
       └── Task Definition: web:5
            └── Container: nginx
```

#### Microservice with Sidecar

```
Cluster: prod
  └── Service: api (desired: 5)
       └── Task Definition: api:10
            ├── Container: app (main)
            └── Container: envoy (sidecar proxy)
```

#### Multi-Service Application

```
Cluster: prod
  ├── Service: frontend (ALB → port 80)
  ├── Service: backend (ALB → port 3000)
  └── Service: worker (no load balancer)
```

## Docker to ECS Mapping

| Docker Concept        | ECS Equivalent                    | Notes                            |
| --------------------- | --------------------------------- | -------------------------------- |
| `docker-compose.yml`  | Task Definition                   | Blueprint for running containers |
| `services:` section   | Container Definitions             | Individual container specs       |
| `docker run`          | Run Task                          | One-off task execution           |
| `docker-compose up`   | Create Service                    | Long-running application         |
| `-p 8080:80`          | containerPort: 80, hostPort: 8080 | Port mapping                     |
| Named volume (shared) | EFS                               | Persistent, multi-task storage   |
| Bind mount            | Host path (EC2 only)              | Access host filesystem           |
| `docker exec -it`     | ECS Exec                          | Interactive shell access         |
| `docker ps`           | List Tasks                        | See running containers           |

## Best Practices

### 1. Networking

- ✓ Use **awsvpc** mode for all new deployments
- ✓ Use **ip** target type for ALB target groups
- ✓ Configure security groups to allow ALB → Task traffic

### 2. Storage

- ✓ Use **EFS** for persistent storage (works with Fargate)
- ✓ Use **ephemeral storage** for temporary files
- ✗ Avoid EBS unless you have specific high-IOPS requirements

### 3. Deployments

- ✓ Use **Blue/Green** for critical services
- ✓ Enable **circuit breaker** for automatic rollback
- ✓ Configure proper **health checks** (path, interval, thresholds)
- ✓ Use **Least Outstanding Requests** for variable workloads

### 4. Load Balancing

- ✓ Set appropriate **deregistration delay** (30-300 seconds)
- ✓ Use **Service Discovery** (Cloud Map) for internal service-to-service communication
- ✓ Configure **health check grace period** to allow tasks to start

### 5. Security

- ✓ Use **IAM roles** for task permissions (taskRoleArn)
- ✓ Enable **EFS encryption** in transit
- ✓ Use **Secrets Manager** or **SSM Parameter Store** for sensitive data
- ✓ Never hardcode credentials in task definitions

### 6. Monitoring

- ✓ Enable **Container Insights** for detailed metrics
- ✓ Configure **CloudWatch alarms** for deployment failures
- ✓ Use **CloudWatch Logs** for centralized logging
- ✓ Monitor **CPU and memory utilization** for right-sizing

## Common Commands

### AWS CLI

```bash
# List clusters
aws ecs list-clusters

# List services in a cluster
aws ecs list-services --cluster my-cluster

# List tasks in a service
aws ecs list-tasks --cluster my-cluster --service-name my-service

# Describe a task
aws ecs describe-tasks --cluster my-cluster --tasks <task-id>

# Update service to new task definition
aws ecs update-service \
  --cluster my-cluster \
  --service my-service \
  --task-definition my-app:2

# Scale service
aws ecs update-service \
  --cluster my-cluster \
  --service my-service \
  --desired-count 5

# Execute command in container
aws ecs execute-command \
  --cluster my-cluster \
  --task <task-id> \
  --container my-container \
  --interactive \
  --command "/bin/sh"

# View logs
aws logs tail /ecs/my-service --follow
```

### Troubleshooting

#### Task Won't Start

1. Check task definition for errors
2. Verify IAM roles exist and have correct permissions
3. Check if ECR image exists and is accessible
4. Verify subnets have available IP addresses
5. Check security group rules

#### Task Starts But Immediately Stops

1. Check CloudWatch logs for application errors
2. Verify health check configuration
3. Check if container has enough memory/CPU
4. Review task stopped reason: `aws ecs describe-tasks`

#### Load Balancer Not Routing Traffic

1. Verify target group has healthy targets
2. Check security group rules (ALB → Task)
3. Verify health check path returns 200 OK
4. Check if tasks are in correct subnets
5. Verify listener rules are correct

#### Deployment Stuck

1. Check if new tasks are failing health checks
2. Verify sufficient capacity in cluster
3. Check CloudWatch Events for errors
4. Review deployment configuration (min/max healthy percent)

## Additional Resources

- [AWS ECS Documentation](https://docs.aws.amazon.com/ecs/)
- [AWS ECS Best Practices Guide](https://docs.aws.amazon.com/AmazonECS/latest/bestpracticesguide/)
- [AWS CDK ECS Patterns](https://docs.aws.amazon.com/cdk/api/v2/docs/aws-cdk-lib.aws_ecs_patterns-readme.html)
- [Docker Documentation](https://docs.docker.com/)
- [AWS Application Load Balancer](https://docs.aws.amazon.com/elasticloadbalancing/latest/application/)

## Summary

ECS is AWS's container orchestration service that translates Docker concepts into AWS-managed infrastructure:

- **Task definitions** are like Docker Compose files
- **Services** keep your containers running and handle deployments
- **awsvpc networking** gives each task its own IP address
- **EFS** provides persistent storage compatible with Fargate
- **ALB integration** automatically manages load balancer targets
- **Blue/Green deployments** enable safe production releases with validation
- **ECS Exec** provides shell access without SSH

The key advantage of ECS is that you **never manually manage IPs, instances, or deployments** - AWS handles the infrastructure while you focus on your application.
