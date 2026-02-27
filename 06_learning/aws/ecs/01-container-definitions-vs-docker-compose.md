# Container Definitions vs Docker Compose

## Overview

A **container definition** in ECS is equivalent to a single **service** in Docker Compose.

## Container Definition (ECS)

A container definition is a JSON object within a task definition that describes how to run a single container:

```json
{
  "name": "web",
  "image": "nginx:latest",
  "memory": 512,
  "cpu": 256,
  "portMappings": [{ "containerPort": 80, "hostPort": 8080 }],
  "environment": [{ "name": "ENV", "value": "prod" }],
  "mountPoints": [{ "sourceVolume": "data", "containerPath": "/var/www" }]
}
```

## Docker Compose Service Equivalent

```yaml
services:
  web:
    image: nginx:latest
    ports:
      - "8080:80"
    environment:
      ENV: prod
    volumes:
      - data:/var/www
```

## Key Takeaway

- **Container Definition** = One service in Docker Compose
- **Task Definition** = Entire Docker Compose file (see next lesson)
