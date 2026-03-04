# What is a Kubernetes Ingress Controller?

A Kubernetes Ingress Controller is a specialized load balancer that manages external access to services within a Kubernetes cluster. Here's what you need to know:

## What It Does

- **Routes external HTTP/HTTPS traffic** to internal services based on rules you define
- **Acts as a reverse proxy** and load balancer at the application layer (Layer 7)
- **Manages SSL/TLS termination** for secure connections
- **Provides a single entry point** to your cluster instead of exposing each service individually

## How It Works

1. You define **Ingress resources** (YAML manifests) that specify routing rules:
   - Which hostnames/domains to accept
   - Which URL paths map to which services
   - SSL certificates to use

2. The **Ingress Controller** watches for these Ingress resources and:
   - Configures itself to implement the routing rules
   - Updates its configuration dynamically as rules change
   - Routes incoming traffic to the appropriate backend services

## Common Ingress Controllers

- **NGINX Ingress Controller** - Most popular, based on NGINX
- **Traefik** - Modern, automatic service discovery
- **HAProxy Ingress** - High performance
- **AWS ALB Ingress Controller** - For AWS Application Load Balancer
- **GCE Ingress Controller** - For Google Cloud Load Balancer

## Example

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: example-ingress
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
          - path: /
            pathType: Prefix
            backend:
              service:
                name: web-service
                port:
                  number: 80
```

This routes `example.com/api/*` to the api-service and `example.com/*` to the web-service.
