# ECS Architecture Hierarchy

## Overview

Understanding the relationship between clusters, services, tasks, and containers is fundamental to working with ECS.

## Hierarchy Diagram

```
Cluster
  └── Service (manages multiple tasks)
       └── Task (running instance of task definition)
            └── Container Definition (specification)
                 └── Container (actual running container)
```

## Component Breakdown

### Cluster

**What**: A logical grouping of resources (EC2 instances or Fargate capacity)

**Purpose**: Isolate and organize your applications

**Example**:

```bash
# You might have clusters for different environments
my-app-dev-cluster
my-app-staging-cluster
my-app-prod-cluster
```

**Analogy**: A data center building

---

### Service

**What**: Manages and maintains a desired number of tasks running continuously

**Purpose**: Ensures your application stays running, handles load balancing, auto-scaling

**Example**:

```json
{
  "serviceName": "web-service",
  "cluster": "my-cluster",
  "taskDefinition": "my-app:5",
  "desiredCount": 3,  // Keep 3 tasks running at all times
  "loadBalancers": [...],
  "autoScaling": {...}
}
```

**Analogy**: A department that ensures 3 workers are always on duty

**Key Responsibilities**:

- Maintains desired task count
- Replaces failed tasks
- Integrates with load balancers
- Handles rolling deployments
- Auto-scaling

---

### Task Definition

**What**: Blueprint/template that describes how to run your application

**Purpose**: Defines containers, resources, networking, volumes (like a Docker Compose file)

**Example**:

```json
{
  "family": "my-app",
  "containerDefinitions": [
    {
      "name": "web",
      "image": "nginx:latest",
      "memory": 512
    },
    {
      "name": "app",
      "image": "myapp:latest",
      "memory": 1024
    }
  ]
}
```

**Analogy**: Job description / instruction manual

**Versioning**: `my-app:1`, `my-app:2`, `my-app:3` (family:revision)

---

### Task

**What**: Running instance of a task definition

**Purpose**: The actual containers running your application

**Example**:

```
Task ID: abc123 (running my-app:5)
  ├── Container: web (nginx)
  └── Container: app (myapp)
```

**Analogy**: An actual worker following the job instructions

**Lifecycle**: Created by service or run standalone, runs until stopped or fails

---

### Container Definition

**What**: Specification for a single container within a task definition

**Purpose**: Describes image, ports, environment variables, resources for one container

**Example**:

```json
{
  "name": "web",
  "image": "nginx:latest",
  "memory": 512,
  "portMappings": [{ "containerPort": 80 }],
  "environment": [{ "name": "ENV", "value": "prod" }]
}
```

**Analogy**: Individual responsibilities within the job description

---

### Container

**What**: Actual running container (Docker container)

**Purpose**: Executes your application code

**Analogy**: Person doing that specific responsibility

---

## Visual Example

```
┌─────────────────────────────────────────────────────────┐
│ Cluster: production-cluster                             │
│                                                          │
│  ┌────────────────────────────────────────────────────┐ │
│  │ Service: web-service                                │ │
│  │ - Desired count: 3                                  │ │
│  │ - Task definition: my-app:5                         │ │
│  │ - Load balancer: ALB                                │ │
│  │                                                      │ │
│  │  ┌─────────────────────────────────────────────┐   │ │
│  │  │ Task 1 (running instance of my-app:5)       │   │ │
│  │  │  ├── Container: nginx (port 80)              │   │ │
│  │  │  └── Container: app (port 3000)              │   │ │
│  │  └─────────────────────────────────────────────┘   │ │
│  │                                                      │ │
│  │  ┌─────────────────────────────────────────────┐   │ │
│  │  │ Task 2 (running instance of my-app:5)       │   │ │
│  │  │  ├── Container: nginx (port 80)              │   │ │
│  │  │  └── Container: app (port 3000)              │   │ │
│  │  └─────────────────────────────────────────────┘   │ │
│  │                                                      │ │
│  │  ┌─────────────────────────────────────────────┐   │ │
│  │  │ Task 3 (running instance of my-app:5)       │   │ │
│  │  │  ├── Container: nginx (port 80)              │   │ │
│  │  │  └── Container: app (port 3000)              │   │ │
│  │  └─────────────────────────────────────────────┘   │ │
│  └────────────────────────────────────────────────────┘ │
│                                                          │
│  ┌────────────────────────────────────────────────────┐ │
│  │ Service: worker-service                             │ │
│  │ - Desired count: 2                                  │ │
│  │ - Task definition: worker:3                         │ │
│  │                                                      │ │
│  │  ┌─────────────────────────────────────────────┐   │ │
│  │  │ Task 1 (running instance of worker:3)       │   │ │
│  │  │  └── Container: background-worker            │   │ │
│  │  └─────────────────────────────────────────────┘   │ │
│  │                                                      │ │
│  │  ┌─────────────────────────────────────────────┐   │ │
│  │  │ Task 2 (running instance of worker:3)       │   │ │
│  │  │  └── Container: background-worker            │   │ │
│  │  └─────────────────────────────────────────────┘   │ │
│  └────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

## Key Relationships

### Task Definition → Container Definition

- A task definition **contains** one or more container definitions
- It's the template/blueprint
- Versioned as `family:revision`

### Service → Task

- A service **creates and manages** multiple tasks
- Ensures desired count is maintained
- Handles rolling deployments when you update the task definition
- Replaces failed tasks automatically

### Cluster → Service

- A cluster **hosts** multiple services
- Provides the compute resources (EC2 or Fargate)
- Organizes services by environment or application

### Task → Container

- A task is a running instance that **executes** the containers defined in container definitions
- All containers in a task run on the same host
- Containers in a task share networking (can communicate via localhost)

## Practical Workflow

### 1. Create a Task Definition (Blueprint)

```bash
aws ecs register-task-definition --cli-input-json file://task-def.json
# Result: my-app:1 (family:revision)
```

### 2. Create a Service (Manager)

```bash
aws ecs create-service \
  --cluster production-cluster \
  --service-name web-service \
  --task-definition my-app:1 \
  --desired-count 3
```

This will:

1. Use the `my-app:1` task definition as a blueprint
2. Launch 3 tasks (each runs the containers defined in the task definition)
3. Keep those 3 tasks running continuously
4. If a task dies, the service automatically starts a replacement

### 3. Update the Application

```bash
# Register new version of task definition
aws ecs register-task-definition --cli-input-json file://task-def-v2.json
# Result: my-app:2

# Update service to use new version
aws ecs update-service \
  --cluster production-cluster \
  --service web-service \
  --task-definition my-app:2
```

The service will:

- Perform a rolling deployment
- Start new tasks with `my-app:2`
- Stop old tasks with `my-app:1`
- Maintain availability during the update

## Service vs Standalone Task

### Service: Long-Running Applications

**Use for**:

- Web servers, APIs, microservices
- Applications that should always be running
- Applications that need load balancing
- Applications that need auto-scaling

**Features**:

- Maintains desired count
- Integrates with load balancers
- Auto-scaling support
- Health checks and auto-recovery

### Standalone Task: One-Off Jobs

```bash
aws ecs run-task \
  --cluster my-cluster \
  --task-definition batch-job:1 \
  --count 1
```

**Use for**:

- Batch jobs
- Database migrations
- Scheduled tasks (via CloudWatch Events)
- One-time operations

**Features**:

- Runs once and stops
- No automatic restart
- No load balancer integration

## Summary Table

| Component                | Type     | Count                       | Purpose                 |
| ------------------------ | -------- | --------------------------- | ----------------------- |
| **Cluster**              | Grouping | 1 per environment           | Organize resources      |
| **Service**              | Manager  | N per cluster               | Keep tasks running      |
| **Task Definition**      | Template | Versioned (family:revision) | Define how to run       |
| **Task**                 | Instance | N per service               | Actually running        |
| **Container Definition** | Spec     | 1+ per task definition      | Describe each container |
| **Container**            | Process  | 1+ per task                 | The actual workload     |

## Common Patterns

### Pattern 1: Simple Web Application

```
Cluster: prod
  └── Service: web (desired: 3)
       └── Task Definition: web:5
            └── Container: nginx
```

### Pattern 2: Microservice with Sidecar

```
Cluster: prod
  └── Service: api (desired: 5)
       └── Task Definition: api:10
            ├── Container: app (main application)
            └── Container: envoy (sidecar proxy)
```

### Pattern 3: Multi-Service Application

```
Cluster: prod
  ├── Service: frontend (desired: 2)
  │    └── Task Definition: frontend:3
  │         └── Container: react-app
  │
  ├── Service: backend (desired: 3)
  │    └── Task Definition: backend:7
  │         └── Container: node-api
  │
  └── Service: worker (desired: 2)
       └── Task Definition: worker:2
            └── Container: background-processor
```

## Key Takeaways

1. **Cluster** = Compute resources (where tasks run)
2. **Service** = Task lifecycle manager (keeps them running)
3. **Task Definition** = Blueprint (what to run)
4. **Task** = Running instance (actually running)
5. **Container Definition** = Individual container spec
6. **Container** = The actual process

The service uses the task definition as a blueprint to create and manage tasks, which run the containers you've defined.
