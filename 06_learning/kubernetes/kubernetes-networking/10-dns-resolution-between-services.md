# DNS Resolution Between Services

DNS works **cluster-wide** - any Pod can resolve any Service's DNS name.

## Each Service Gets Its Own DNS Name

```yaml
# Service 1
apiVersion: v1
kind: Service
metadata:
  name: app1
spec:
  selector:
    app: app1
  ports:
    - port: 80
---
# Service 2
apiVersion: v1
kind: Service
metadata:
  name: app2
spec:
  selector:
    app: app2
  ports:
    - port: 8080
```

Kubernetes automatically creates DNS entries:

- `app1` → resolves to app1's ClusterIP
- `app2` → resolves to app2's ClusterIP

## DNS Works Everywhere in the Cluster

**Any Pod can resolve any Service name:**

```bash
# From a Pod in app1
$ kubectl exec -it app1-pod -- curl http://app2:8080
Success!  # app1 can reach app2

# From a Pod in app2
$ kubectl exec -it app2-pod -- curl http://app1:80
Success!  # app2 can reach app1

# From a completely different app
$ kubectl exec -it app3-pod -- curl http://app1:80
Success!  # app3 can reach app1 too
```

## Full DNS Format

The complete DNS name format is:

```
<service-name>.<namespace>.svc.cluster.local
```

For example, if both services are in the `default` namespace:

- Full name: `app1.default.svc.cluster.local`
- Short name: `app1` (works within the same namespace)

## Cross-Namespace DNS

If services are in **different namespaces**, you need to be more specific:

```yaml
# In namespace "frontend"
apiVersion: v1
kind: Service
metadata:
  name: web
  namespace: frontend
---
# In namespace "backend"
apiVersion: v1
kind: Service
metadata:
  name: api
  namespace: backend
```

**From a Pod in the `frontend` namespace:**

```bash
# Same namespace - short name works
$ curl http://web:80
Success!

# Different namespace - need to specify namespace
$ curl http://api.backend:8080
Success!

# Or use the full DNS name
$ curl http://api.backend.svc.cluster.local:8080
Success!
```

## DNS Resolution Patterns

| From Pod in Namespace | Target Service    | DNS Name to Use |
| --------------------- | ----------------- | --------------- |
| default               | app1 (in default) | `app1`          |
| default               | app2 (in default) | `app2`          |
| default               | api (in backend)  | `api.backend`   |
| frontend              | api (in backend)  | `api.backend`   |
| frontend              | web (in frontend) | `web`           |

## Visual Example

```
┌─────────────────────────────────────┐
│  Kubernetes Cluster                 │
│                                     │
│  DNS Server (CoreDNS)               │
│    ↓                                │
│  app1 → 10.100.1.50                 │
│  app2 → 10.100.2.80                 │
│                                     │
│  ┌──────────┐      ┌──────────┐    │
│  │ app1-pod │      │ app2-pod │    │
│  │          │      │          │    │
│  │ Can call:│      │ Can call:│    │
│  │ - app2   │      │ - app1   │    │
│  │ - app1   │      │ - app2   │    │
│  └──────────┘      └──────────┘    │
│                                     │
│  ┌──────────┐                       │
│  │ app3-pod │                       │
│  │          │                       │
│  │ Can call:│                       │
│  │ - app1   │                       │
│  │ - app2   │                       │
│  └──────────┘                       │
└─────────────────────────────────────┘
```

## What DNS Actually Resolves To

When a Pod does `curl http://app1:80`:

1. **DNS lookup** happens: `app1` → `10.100.1.50` (ClusterIP)
2. **Request goes to** the Service's ClusterIP
3. **Service load balances** to one of the app1 Pods
4. **Pod responds** back through the Service

```bash
# You can see this with nslookup inside a Pod
$ kubectl exec -it app2-pod -- nslookup app1

Name:      app1.default.svc.cluster.local
Address:   10.100.1.50  # This is app1's ClusterIP
```

## Important Notes

### 1. Service Names Must Be Unique Per Namespace

```yaml
# ❌ NOT allowed - duplicate names in same namespace
namespace: default
  - service: app1
  - service: app1  # Error!

# ✅ Allowed - same name in different namespaces
namespace: frontend
  - service: app1
namespace: backend
  - service: app1  # OK - different namespace
```

### 2. DNS Only Works for Services, Not Pods

```bash
# ✅ Works - Service DNS
$ curl http://app1:80

# ❌ Doesn't work - Pod names aren't in DNS by default
$ curl http://app1-pod-xyz123:80
```

(There are ways to get Pod DNS with StatefulSets, but that's advanced)

### 3. This Is Why You Don't Hardcode IPs

**Bad:**

```yaml
env:
  - name: API_URL
    value: "http://10.100.2.80:8080" # ❌ ClusterIP can change!
```

**Good:**

```yaml
env:
  - name: API_URL
    value: "http://app2:8080" # ✅ DNS name is stable
```

## Summary

- **Each Service gets its own DNS name** (the service name)
- **DNS resolution works cluster-wide** - any Pod can resolve any Service
- **Same namespace**: Use short name (`app1`)
- **Different namespace**: Use `<service>.<namespace>` (`app1.backend`)
- **DNS is managed by CoreDNS** (runs automatically in your cluster)

So your app1 service and app2 service each have their own DNS names, and **all Pods** in the cluster can use those names to communicate with the services!
