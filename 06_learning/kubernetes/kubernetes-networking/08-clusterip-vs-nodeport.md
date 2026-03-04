# ClusterIP vs NodePort

The key difference is **where you can access the service from**.

## ClusterIP (Internal Only)

A **ClusterIP** Service is only accessible **inside the cluster**.

```yaml
apiVersion: v1
kind: Service
metadata:
  name: my-app
spec:
  type: ClusterIP # Default if you don't specify
  ports:
    - port: 80
      targetPort: 8080
  selector:
    app: my-app
```

After creation:

```bash
$ kubectl get svc my-app
NAME     TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)
my-app   ClusterIP   10.100.200.50   <none>        80/TCP
```

**Who can access it:**

- ✅ Other Pods in the cluster can access `http://10.100.200.50:80`
- ✅ Other Pods can use the DNS name `http://my-app:80`
- ❌ You cannot access it from outside the cluster
- ❌ Your browser/laptop cannot reach it

**Use cases:**

- Internal microservices talking to each other
- Databases (PostgreSQL, Redis) that only need cluster access
- Backend APIs that sit behind an Ingress Controller

## NodePort (External Access)

A **NodePort** Service is accessible **from outside the cluster** via any node's IP.

```yaml
apiVersion: v1
kind: Service
metadata:
  name: my-app
spec:
  type: NodePort
  ports:
    - port: 80
      targetPort: 8080
      nodePort: 30080
  selector:
    app: my-app
```

After creation:

```bash
$ kubectl get svc my-app
NAME     TYPE       CLUSTER-IP      EXTERNAL-IP   PORT(S)
my-app   NodePort   10.100.200.50   <none>        80:30080/TCP
```

**Who can access it:**

- ✅ Other Pods can access `http://10.100.200.50:80` (has ClusterIP too!)
- ✅ Other Pods can use DNS `http://my-app:80`
- ✅ **External access via any node: `http://<NodeIP>:30080`**
- ✅ Your browser/laptop can reach `http://192.168.1.10:30080`

**Use cases:**

- Exposing services for external access (when not using a LoadBalancer)
- Development/testing environments
- Self-hosted clusters without cloud load balancers

## Key Differences

| Aspect              | ClusterIP                          | NodePort                             |
| ------------------- | ---------------------------------- | ------------------------------------ |
| **Accessibility**   | Internal only                      | Internal + External                  |
| **IP Address**      | Gets cluster IP (10.x.x.x)         | Gets cluster IP + opens node ports   |
| **Port**            | Any port                           | 30000-32767 range                    |
| **External Access** | None                               | Via `<NodeIP>:<NodePort>`            |
| **Cost**            | Free                               | Free                                 |
| **Security**        | More secure (no external exposure) | Less secure (port open on all nodes) |

## Visual Comparison

**ClusterIP:**

```
Pod A ──────────────► ClusterIP Service (10.100.200.50:80) ──► Pods
                             ▲
Pod B ───────────────────────┘

Your Laptop ✗ (cannot reach)
```

**NodePort:**

```
Pod A ──────────────► ClusterIP Service (10.100.200.50:80) ──► Pods
                             ▲
Pod B ───────────────────────┘

Your Laptop ──► Node 1 (192.168.1.10:30080) ───┐
                                                ├──► Pods
Internet ────► Node 2 (192.168.1.11:30080) ────┘
```

## Important Detail: NodePort Includes ClusterIP

When you create a NodePort Service, you **get both**:

1. A ClusterIP (internal access)
2. NodePort (external access)

NodePort is essentially **ClusterIP + external port mapping**.

## When to Use Each

### Use ClusterIP when:

- Service only needs to be accessed by other services in the cluster
- You're using an Ingress Controller (it accesses services via ClusterIP)
- You're using Cloudflare Tunnel or similar (internal access is enough)
- Security: Don't expose unless necessary

### Use NodePort when:

- You need direct external access without a cloud LoadBalancer
- You're on bare-metal and setting up your own external load balancer
- Development/testing on local clusters
- You want simple external access without additional components

### Use LoadBalancer when:

- You're on a cloud provider and want the easiest external access
- You're willing to pay for cloud load balancers
- You want automatic provisioning and management

## Example Scenario

Let's say you have a web app with a database:

```yaml
# Database - ClusterIP (internal only)
apiVersion: v1
kind: Service
metadata:
  name: postgres
spec:
  type: ClusterIP # Only accessible inside cluster
  ports:
    - port: 5432
  selector:
    app: postgres
---
# Frontend - NodePort (external access)
apiVersion: v1
kind: Service
metadata:
  name: frontend
spec:
  type: NodePort # Accessible externally
  ports:
    - port: 80
      nodePort: 30080
  selector:
    app: frontend
```

Now:

- Frontend pods can connect to `postgres:5432` ✅
- Users can access frontend at `http://node-ip:30080` ✅
- Users **cannot** directly access postgres ✅ (security!)

## Summary

**ClusterIP** = Internal networking only (like a private network)
**NodePort** = ClusterIP + external access on every node (like port forwarding on your router)
