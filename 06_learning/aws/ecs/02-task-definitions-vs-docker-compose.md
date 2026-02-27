# Task Definitions vs Docker Compose

## Overview

A **task definition** is much closer to a complete Docker Compose file. It groups multiple containers that run together on the same host.

## Task Definition (ECS)

```json
{
  "family": "my-app",
  "networkMode": "awsvpc",
  "containerDefinitions": [
    {
      "name": "web",
      "image": "nginx:latest",
      "memory": 512,
      "portMappings": [{ "containerPort": 80 }]
    },
    {
      "name": "app",
      "image": "myapp:latest",
      "memory": 1024,
      "environment": [{ "name": "DB_HOST", "value": "localhost" }]
    }
  ],
  "volumes": [{ "name": "shared-data" }]
}
```

## Docker Compose Equivalent

```yaml
services:
  web:
    image: nginx:latest
    ports:
      - "80:80"

  app:
    image: myapp:latest
    environment:
      DB_HOST: localhost

volumes:
  shared-data:
```

## Similarities

| Feature             | Task Definition                     | Docker Compose             |
| ------------------- | ----------------------------------- | -------------------------- |
| Groups containers   | ✓ Multiple container definitions    | ✓ Multiple services        |
| Shared networking   | ✓ Containers can talk via localhost | ✓ Services on same network |
| Shared volumes      | ✓ Volume mounts                     | ✓ Named/bind volumes       |
| Resource allocation | ✓ Task-level + container-level      | ✓ Service-level limits     |

## Key Difference

- **Task definitions**: Production deployments on AWS ECS
- **Docker Compose**: Local development orchestration

## Mapping

- Task Definition ≈ `docker-compose.yml` file
- Container Definition ≈ Single service in `docker-compose.yml`
