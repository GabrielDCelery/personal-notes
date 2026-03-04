# How NodePort Works

NodePort opens a port on **every individual node** in the cluster, not the entire cluster as a single entity.

## What NodePort Does

When you create a NodePort Service:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: my-app
spec:
  type: NodePort
  ports:
    - port: 80 # ClusterIP port (internal)
      targetPort: 8080 # Pod port
      nodePort: 30080 # Port opened on EVERY node
  selector:
    app: my-app
```

Kubernetes:

1. **Opens port 30080 on every node** in your cluster
2. **Listens for traffic** on that port on all nodes
3. **Forwards traffic** to your Pods, regardless of which node receives the request

## Visual Example

Say you have a 3-node cluster:

```
Node 1 (IP: 192.168.1.10)  ←  Port 30080 opened
Node 2 (IP: 192.168.1.11)  ←  Port 30080 opened
Node 3 (IP: 192.168.1.12)  ←  Port 30080 opened

All nodes listen on port 30080!
```

Your Pods might be running on specific nodes:

```
Node 1: Pod A (my-app)
Node 2: Pod B (my-app)
Node 3: (no pods for my-app)
```

## Traffic Flow - The Magic Part

Here's what's interesting: **You can hit ANY node's IP:port and reach your service**, even if the pod isn't on that node!

```
User → 192.168.1.10:30080 → Routes to Pod A (on Node 1) ✅
User → 192.168.1.11:30080 → Routes to Pod B (on Node 2) ✅
User → 192.168.1.12:30080 → Routes to Pod A or B (cross-node!) ✅
```

**How?** Kubernetes uses **kube-proxy** on each node to:

1. Listen on the NodePort
2. Load balance across all Pods with that selector
3. Route traffic (even across nodes if needed)

## Port Range Restrictions

NodePort services use ports in the range **30000-32767** by default.

```yaml
nodePort: 30080  # ✅ Valid
nodePort: 80     # ❌ Invalid - below 30000
nodePort: 35000  # ❌ Invalid - above 32767
```

You can change this range in the API server config, but it's not recommended.

## Accessing Your App

After creating the NodePort Service:

```bash
$ kubectl get svc my-app
NAME     TYPE       CLUSTER-IP      EXTERNAL-IP   PORT(S)
my-app   NodePort   10.100.200.50   <none>        80:30080/TCP
```

You can access it via:

- `http://192.168.1.10:30080` (Node 1)
- `http://192.168.1.11:30080` (Node 2)
- `http://192.168.1.12:30080` (Node 3)
- Or internally via ClusterIP: `http://10.100.200.50:80`

## Common Usage Pattern

Typically you put your own load balancer in front:

```
Internet
  ↓
Your Load Balancer (HAProxy, NGINX, etc.)
  ↓ ↓ ↓ (distributes to)
Node 1:30080  Node 2:30080  Node 3:30080
  ↓              ↓              ↓
      Your Pods (anywhere in cluster)
```

This way:

- Users hit your load balancer on port 80/443
- Load balancer forwards to NodePorts
- NodePorts route to Pods

## NodePort vs ClusterIP vs LoadBalancer

| Type             | Accessible From     | Use Case                                 |
| ---------------- | ------------------- | ---------------------------------------- |
| **ClusterIP**    | Inside cluster only | Internal services (databases, APIs)      |
| **NodePort**     | Any node's IP:port  | Self-hosted, need external access        |
| **LoadBalancer** | External cloud LB   | Cloud environments, easy external access |

## Security Consideration

NodePort **opens that port on all nodes' network interfaces**, so:

- If your nodes are public-facing → anyone can hit that port
- Use firewall rules to restrict access
- Or use ClusterIP + Ingress + Cloudflare Tunnel for zero exposed ports!

## Summary

**NodePort**:

- Opens the same port on **every node** (not just "the cluster")
- You can access the service through **any node's IP**
- Traffic is routed to Pods **even if they're on different nodes**
- Uses high ports (30000-32767)
- Doesn't require cloud provider
- Often paired with an external load balancer
