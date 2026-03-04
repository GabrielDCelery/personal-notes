# What is a LoadBalancer Service?

A **LoadBalancer Service** is one of the ways Kubernetes can expose your application to the outside world.

## What It Is

In Kubernetes, a **Service** is an abstraction that defines how to access a set of Pods. There are different types:

- **ClusterIP** - Only accessible inside the cluster (default)
- **NodePort** - Opens a port on every node
- **LoadBalancer** - Creates an external load balancer

## What LoadBalancer Service Does

When you create a Service of type `LoadBalancer`, Kubernetes:

1. **Requests a load balancer** from your cloud provider (AWS ELB/ALB, GCP Load Balancer, Azure Load Balancer)
2. **Gets an external IP address** assigned to that load balancer
3. **Configures the load balancer** to forward traffic to your Pods
4. **Updates the Service** with the external IP

```yaml
apiVersion: v1
kind: Service
metadata:
  name: my-app
spec:
  type: LoadBalancer
  ports:
    - port: 80 # External port on the load balancer
      targetPort: 8080 # Port your Pod listens on
  selector:
    app: my-app
```

After creation:

```bash
$ kubectl get svc my-app
NAME     TYPE           CLUSTER-IP      EXTERNAL-IP      PORT(S)
my-app   LoadBalancer   10.100.200.50   203.0.113.42     80:31234/TCP
```

Now anyone can access your app at `http://203.0.113.42`

## What the Load Balancer Actually Does

The cloud load balancer:

1. **Distributes traffic** across multiple Pods (if you have replicas)
2. **Health checks** - Only sends traffic to healthy Pods
3. **Provides a stable entry point** - Even as Pods are created/destroyed
4. **Handles failover** - Automatically removes unhealthy Pods from rotation

**Example traffic flow:**

```
User's Browser (Internet)
  ↓
External Load Balancer (203.0.113.42:80)
  ↓ ↓ ↓ (distributes across)
Pod 1 (10.244.1.5:8080)
Pod 2 (10.244.2.3:8080)
Pod 3 (10.244.3.7:8080)
```

## Important Points

### Only Works in Cloud Environments

LoadBalancer Services **only work** if you're running Kubernetes on a cloud provider (AWS, GCP, Azure) or using something like MetalLB for bare-metal clusters.

If you try this on a local cluster (minikube, kind):

```bash
$ kubectl get svc my-app
NAME     TYPE           EXTERNAL-IP
my-app   LoadBalancer   <pending>     # Stuck forever!
```

### Cost Implications

**Each LoadBalancer Service creates a separate cloud load balancer**, which costs money:

- AWS ELB: ~$16-25/month per load balancer
- GCP Load Balancer: ~$18/month + traffic
- Azure Load Balancer: ~$18/month + traffic

If you have 10 services, that's 10 load balancers = $$$$

## Why Use an Ingress Controller Instead?

This is why Ingress Controllers are popular - you get:

**With LoadBalancer Services (without Ingress):**

- 10 services = 10 LoadBalancers = 10 external IPs = $$$
- No path-based routing
- No hostname-based routing

**With Ingress Controller + 1 LoadBalancer:**

- 1 LoadBalancer points to Ingress Controller
- Ingress Controller routes to all your services based on rules
- Cost: Just 1 cloud load balancer
- Features: Path routing, hostname routing, SSL termination, etc.

## Summary

A **LoadBalancer Service**:

- Is a Kubernetes Service type
- Automatically provisions a cloud load balancer
- Gives you an external IP to access your app
- Costs money per Service
- Is often replaced by using an Ingress Controller with a single LoadBalancer

With **Cloudflare Tunnel**, you don't need LoadBalancer Services at all since the tunnel provides the external access!
