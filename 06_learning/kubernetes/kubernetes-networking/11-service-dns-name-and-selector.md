# Service DNS Name and Selector

Understanding what determines the DNS name and what the selector actually selects.

## 1. DNS Name = metadata.name

The **DNS name comes from `metadata.name`** - that's it!

```yaml
apiVersion: v1
kind: Service
metadata:
  name: app2 # ← This becomes the DNS name
spec:
  selector:
    app: app2 # ← This has nothing to do with DNS
```

The DNS name will be: `app2` (or fully: `app2.default.svc.cluster.local`)

**The selector does NOT affect the DNS name at all.**

You could even do this:

```yaml
metadata:
  name: my-awesome-service # ← DNS name
spec:
  selector:
    app: totally-different-name # ← No problem!
```

DNS name would be `my-awesome-service`, even though it selects pods labeled `app: totally-different-name`.

## 2. Selector Selects Pod Labels (Not Deployments!)

The **selector selects Pods with matching labels**, not Deployments.

### The Service Doesn't Know About Deployments

```yaml
apiVersion: v1
kind: Service
metadata:
  name: app2
spec:
  selector:
    app: app2 # ← Looking for Pods with label "app: app2"
```

This Service will select **any Pod** with the label `app: app2`, regardless of:

- Whether it was created by a Deployment
- Whether it was created by a StatefulSet
- Whether it was created manually
- What the Deployment/StatefulSet is named

### How It Works with Deployments

Let's see the full picture:

```yaml
# Deployment creates Pods with labels
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app-deployment # ← Deployment name (doesn't matter to Service)
spec:
  replicas: 3
  selector:
    matchLabels:
      app: app2 # ← Deployment selects its own Pods
  template:
    metadata:
      labels:
        app: app2 # ← Pods get this label
    spec:
      containers:
        - name: app
          image: nginx
---
# Service selects Pods by label
apiVersion: v1
kind: Service
metadata:
  name: app2 # ← DNS name
spec:
  selector:
    app: app2 # ← Selects any Pod with this label
  ports:
    - port: 8080
```

**What happens:**

1. **Deployment creates 3 Pods**, each gets the label `app: app2`

   ```
   Pod: app2-deployment-abc123  (labels: app=app2)
   Pod: app2-deployment-def456  (labels: app=app2)
   Pod: app2-deployment-ghi789  (labels: app=app2)
   ```

2. **Service looks for Pods** with label `app: app2`
   - Finds all 3 Pods
   - Routes traffic to them
   - Load balances across all 3

3. **Other Pods use DNS** `app2` to reach the Service
   - Service load balances to one of the 3 Pods

## Visual Flow

```
Service (name: app2)
  │
  │ DNS: "app2" → ClusterIP: 10.100.2.80
  │
  └─ selector: {app: app2}
       │
       ├─ Finds Pod 1 (app=app2) ✅
       ├─ Finds Pod 2 (app=app2) ✅
       └─ Finds Pod 3 (app=app2) ✅

Deployment (name: my-app-deployment)
  └─ Creates Pods with label {app: app2}
       ├─ Pod 1
       ├─ Pod 2
       └─ Pod 3
```

## Important: Labels Can Match Anything

The Service will select **all Pods** with matching labels, even from different sources:

```yaml
# Deployment A
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-a
spec:
  template:
    metadata:
      labels:
        app: app2 # ← Has matching label
---
# Deployment B (different deployment!)
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-b
spec:
  template:
    metadata:
      labels:
        app: app2 # ← Also has matching label
---
# Service
apiVersion: v1
kind: Service
metadata:
  name: app2
spec:
  selector:
    app: app2 # ← Selects Pods from BOTH deployments!
```

The Service will load balance across Pods from **both deployments** because they both have the label `app: app2`.

## Common Pattern: Names Match

In practice, people usually make the names match for clarity:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app2 # ← Same name
spec:
  selector:
    matchLabels:
      app: app2 # ← Same label
  template:
    metadata:
      labels:
        app: app2 # ← Same label
---
apiVersion: v1
kind: Service
metadata:
  name: app2 # ← Same name (for DNS)
spec:
  selector:
    app: app2 # ← Same selector
```

This makes it easy to understand, but remember:

- **Service DNS name** = `metadata.name` of the Service
- **Service selector** = Finds Pods by their **labels** (not by Deployment name)

## How to Verify What the Service Selects

```bash
# See which Pods the Service is routing to
$ kubectl get endpoints app2

NAME   ENDPOINTS                                    AGE
app2   10.244.1.5:8080,10.244.2.3:8080,10.244.3.7:8080   5m

# See the Pod IPs
$ kubectl get pods -l app=app2 -o wide

NAME                    IP            NODE
app2-deploy-abc123      10.244.1.5    node-1
app2-deploy-def456      10.244.2.3    node-2
app2-deploy-ghi789      10.244.3.7    node-3
```

The **Endpoints** show the actual Pod IPs the Service is routing to.

## Summary

| What         | Determined By           | Purpose                                                             |
| ------------ | ----------------------- | ------------------------------------------------------------------- |
| **DNS name** | Service `metadata.name` | What other Pods call to reach this service                          |
| **Selector** | Service `spec.selector` | Which **Pod labels** to match                                       |
| **Target**   | Pod labels              | The Service selects **Pods** (not Deployments) with matching labels |

So in your example:

- DNS name is `app2` (from `metadata.name`)
- Selector looks for Pods labeled `app: app2`
- Those Pods could come from any Deployment, StatefulSet, or even be manually created - the Service doesn't care!
