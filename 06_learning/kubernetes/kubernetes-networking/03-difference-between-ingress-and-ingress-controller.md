# Difference Between Ingress and Ingress Controller

These are two separate but related concepts that work together:

## Ingress (the resource)

An **Ingress** is just a Kubernetes resource definition - a YAML file that describes the routing rules you want.

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-ingress
spec:
  rules:
    - host: example.com
      http:
        paths:
          - path: /api
            pathType: Prefix
            backend:
              service:
                name: api-service
                port:
                  number: 8080
```

This is **declarative configuration** - you're saying "I want traffic to example.com/api to go to api-service." But this YAML file by itself does nothing.

## Ingress Controller (the implementation)

An **Ingress Controller** is the actual software/pod running in your cluster that:

1. **Watches for Ingress resources** being created/updated/deleted
2. **Reads the routing rules** from those Ingress resources
3. **Configures itself** (usually NGINX, Traefik, HAProxy, etc.) to implement those rules
4. **Handles the actual traffic** routing and load balancing

## The Analogy

Think of it like:

- **Ingress** = A recipe (instructions)
- **Ingress Controller** = A chef who reads the recipe and cooks the food

Or in Kubernetes terms:

- **Ingress** = The "what" (what routing do I want)
- **Ingress Controller** = The "how" (the software that makes it happen)

## Key Points

- You can have **many Ingress resources** but you need **at least one Ingress Controller** for them to work
- Kubernetes **does not include an Ingress Controller by default** - you must install one yourself (unlike built-in controllers for Deployments, Services, etc.)
- Different Ingress Controllers support different features via **annotations** in the Ingress resource

```yaml
# Example: NGINX-specific annotation
metadata:
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
```

Without an Ingress Controller installed, your Ingress resources are just dormant YAML files that do nothing!
