# Port Mappings: containerPort vs hostPort

## Overview

Port mappings define how traffic flows from the host machine to the container.

## The Two Ports

- **containerPort**: The port the application listens on **inside the container**
- **hostPort**: The port exposed on the **host machine** (EC2 instance or Fargate task)

## Example

```json
{
  "portMappings": [
    {
      "containerPort": 80,
      "hostPort": 8080
    }
  ]
}
```

**Meaning:**

- Your app inside the container listens on port **80**
- The host machine exposes port **8080** to the outside world
- Traffic to `host:8080` → forwarded to → `container:80`

## Visual Flow

```
Internet/VPC
    ↓
Host Machine (EC2/Fargate) :8080  ← hostPort
    ↓ (port mapping)
Container :80                      ← containerPort
    ↓
Your Application
```

## Common Scenarios

### 1. awsvpc Mode (Fargate/Modern EC2)

```json
{
  "networkMode": "awsvpc",
  "portMappings": [
    { "containerPort": 80 } // hostPort not specified
  ]
}
```

- Task has its own IP address
- Connect directly to task's IP on port 80
- `hostPort` is optional (defaults to same as `containerPort`)

### 2. bridge Mode (Legacy EC2)

```json
{
  "networkMode": "bridge",
  "portMappings": [{ "containerPort": 3000, "hostPort": 8080 }]
}
```

- Multiple tasks can run on same EC2 instance
- Each task needs different `hostPort` to avoid conflicts
- Connect to `ec2-ip:8080` → routes to `container:3000`

### 3. Dynamic Port Mapping (bridge + ALB)

```json
{
  "networkMode": "bridge",
  "portMappings": [
    { "containerPort": 3000, "hostPort": 0 } // 0 = random port
  ]
}
```

- ECS assigns a random available port on the host
- ALB automatically discovers the assigned port
- Allows multiple tasks of same service on one EC2 instance

## Docker Equivalent

### ECS Format

```json
{
  "portMappings": [{ "containerPort": 80, "hostPort": 8080 }]
}
```

### Docker CLI

```bash
docker run -p 8080:80 nginx
#             ↑    ↑
#         hostPort:containerPort
```

### Docker Compose

```yaml
services:
  web:
    image: nginx
    ports:
      - "8080:80"
      #   ↑    ↑
      # host:container
```

## Format Comparison

| ECS                                       | Docker CLI     | Docker Compose | Meaning                            |
| ----------------------------------------- | -------------- | -------------- | ---------------------------------- |
| `containerPort: 80`<br>`hostPort: 8080`   | `-p 8080:80`   | `"8080:80"`    | Host port 8080 → Container port 80 |
| `containerPort: 3000`<br>`hostPort: 3000` | `-p 3000:3000` | `"3000:3000"`  | Same port on both sides            |
| `containerPort: 80`<br>(no hostPort)      | `-p 80:80`     | `"80:80"`      | Direct mapping                     |
| `containerPort: 3000`<br>`hostPort: 0`    | `-p 3000`      | `"3000"`       | Random host port                   |

## Multiple Ports

### ECS

```json
{
  "portMappings": [
    { "containerPort": 80, "hostPort": 8080 },
    { "containerPort": 443, "hostPort": 8443 }
  ]
}
```

### Docker CLI

```bash
docker run -p 8080:80 -p 8443:443 nginx
```

### Docker Compose

```yaml
services:
  web:
    ports:
      - "8080:80"
      - "8443:443"
```

## Real-World Example

Your Node.js app:

```javascript
app.listen(3000); // App listens on port 3000
```

Task definition:

```json
{
  "portMappings": [{ "containerPort": 3000, "hostPort": 8080 }]
}
```

Users access: `http://your-host:8080` → ECS forwards to → container's port 3000 → your app responds

## Key Takeaway

- Docker format: `HOST:CONTAINER` (e.g., `8080:80`)
- ECS lists them separately but means the same thing
- In `awsvpc` mode, you typically omit `hostPort`
- In `bridge` mode, you must specify both to avoid conflicts
