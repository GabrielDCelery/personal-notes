# How Traffic Gets Directed to the Ingress Controller

The Ingress Controller itself needs to be exposed somehow. Here are the common methods:

## 1. LoadBalancer Service (Most Common in Cloud)

The Ingress Controller is exposed via a Service of type `LoadBalancer`:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: ingress-nginx-controller
spec:
  type: LoadBalancer # Creates a cloud load balancer
  ports:
    - port: 80
      targetPort: 80
    - port: 443
      targetPort: 443
  selector:
    app: ingress-nginx
```

**Flow:**
Internet → Cloud Load Balancer (external IP) → Ingress Controller Pod → Your Services

**Pros:** Simple, automatic external IP
**Cons:** Costs money (each LoadBalancer = a cloud LB resource)

## 2. NodePort Service (Self-Hosted/Bare Metal)

```yaml
apiVersion: v1
kind: Service
metadata:
  name: ingress-nginx-controller
spec:
  type: NodePort # Opens port on all nodes
  ports:
    - port: 80
      nodePort: 30080 # Accessible on ANY node at this port
    - port: 443
      nodePort: 30443
  selector:
    app: ingress-nginx
```

**Flow:**
Internet → Your external LB/router → Any Node IP:30080 → Ingress Controller Pod → Your Services

**Pros:** Free, no cloud dependency
**Cons:** Need to manage external load balancer yourself, ports are in high range (30000-32767)

## 3. HostNetwork (Direct Host Access)

The Ingress Controller pod runs in the host's network namespace:

```yaml
spec:
  hostNetwork: true # Pod uses node's network directly
  containers:
    - name: nginx-ingress-controller
      ports:
        - containerPort: 80
          hostPort: 80 # Binds to node's port 80
        - containerPort: 443
          hostPort: 443
```

**Flow:**
Internet → Node IP:80/443 → Ingress Controller Pod (same network) → Your Services

**Pros:** Standard ports (80/443), no extra Service needed
**Cons:** Only one Ingress Controller per node, security concerns

## 4. External Load Balancer (Your Own)

You run your own HAProxy, NGINX, or hardware LB outside the cluster:

**Flow:**
Internet → Your Load Balancer → NodePort or directly to nodes → Ingress Controller → Services

## Real-World Example with Cloudflare Tunnel

Since you mentioned Cloudflare Tunnel, here's how it fits:

```
Internet
  ↓
Cloudflare Edge
  ↓
Cloudflare Tunnel (cloudflared pod in cluster)
  ↓
Ingress Controller Service (ClusterIP is fine!)
  ↓
Ingress Controller Pod
  ↓
Your backend Services
```

**With Cloudflare Tunnel, you don't need LoadBalancer or NodePort!** The tunnel connects from inside the cluster, so a simple ClusterIP Service works:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: ingress-nginx-controller
spec:
  type: ClusterIP # Only accessible inside cluster
  ports:
    - port: 80
    - port: 443
  selector:
    app: ingress-nginx
```

The cloudflared pod connects to this internal Service, and Cloudflare handles all external routing. No public IPs needed!
