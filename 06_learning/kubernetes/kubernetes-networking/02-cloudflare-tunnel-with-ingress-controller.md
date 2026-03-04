# Combining Cloudflare Tunnel with an Ingress Controller

Yes, you can absolutely combine Cloudflare Tunnel with an Ingress Controller! There are a few ways to architect this:

## Option 1: Cloudflare Tunnel → Ingress Controller (Recommended)

Point your Cloudflare Tunnel at your Ingress Controller, and let the Ingress Controller handle all internal routing.

**Benefits:**

- No public IP or open ports needed
- Keep your existing Ingress rules and configuration
- Single tunnel endpoint, Ingress Controller handles all routing
- Best of both: Cloudflare's security + Kubernetes-native routing

**Setup:**

```yaml
# cloudflared config
ingress:
  - hostname: example.com
    service: http://nginx-ingress-controller.ingress-nginx.svc.cluster.local:80
  - hostname: "*.example.com"
    service: http://nginx-ingress-controller.ingress-nginx.svc.cluster.local:80
  - service: http_status:404
```

The tunnel sends all traffic to your Ingress Controller, which then routes based on your existing Ingress resources.

## Option 2: Cloudflare Tunnel → Direct to Services

Bypass the Ingress Controller and route directly from Cloudflare to specific services.

**Benefits:**

- Simpler if you only have a few services
- One less hop in the request path

**Drawbacks:**

- Need multiple tunnel rules
- Lose Ingress Controller features (middleware, auth, rate limiting)
- Can't use Ingress resources

```yaml
ingress:
  - hostname: api.example.com
    service: http://api-service.default.svc.cluster.local:8080
  - hostname: web.example.com
    service: http://web-service.default.svc.cluster.local:80
```

## Option 3: Hybrid Approach

Use Cloudflare Tunnel for external access and keep Ingress Controller for internal or other external paths.

**When to use:**

- Development/staging environments (tunnel) vs production (traditional ingress)
- Different security requirements for different services

## Recommendation

Go with **Option 1** if you already have an Ingress Controller - it's the cleanest architecture and maintains separation of concerns. Cloudflare handles the secure tunnel and DDoS protection, while your Ingress Controller manages application-level routing within your cluster.
