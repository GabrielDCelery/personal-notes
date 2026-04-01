# Security — Distilled

Security in system design is about reducing attack surface — not eliminating all risk, but making exploitation expensive and blast radius small when something goes wrong.

## The Core Mental Model: Layers of Defence

No single control is sufficient. Security works because an attacker has to defeat multiple independent layers.

```
External traffic
  │
  ├─ Layer 1: Edge (CDN / WAF)        → DDoS mitigation, IP blocking, bot filtering
  ├─ Layer 2: Rate limiting           → abuse prevention, brute force protection
  ├─ Layer 3: TLS / transport         → confidentiality in transit
  ├─ Layer 4: Auth (identity)         → who is this, what are they allowed to do
  ├─ Layer 5: Input validation        → reject malformed or malicious payloads
  ├─ Layer 6: Application logic       → authorisation checks, least privilege
  ├─ Layer 7: Network isolation       → services can't talk to things they shouldn't
  └─ Layer 8: Encryption at rest      → data is useless if storage is compromised
```

Skipping a layer doesn't make the others stronger — it just means one successful attack reaches deeper.

## Rate Limiting

Rate limiting protects your API from abuse: credential stuffing, scraping, accidental runaway clients, and DDoS amplification.

**Two algorithms:**

```
Token bucket:
  Bucket holds N tokens. Refills at rate R/sec.
  Request consumes 1 token. Bucket empty → reject.

  Allows bursts up to N (bucket size).
  Good for: APIs that need to absorb short bursts.

Sliding window:
  Count requests per client in the last T seconds.
  Exceed limit → reject.

  No burst allowance — stricter, simpler to reason about.
  Good for: per-user quotas, billing rate limits.
```

**Where to enforce:**

```
Edge / CDN            → coarse-grained IP-based limits, handles DDoS traffic
                         before it reaches your infrastructure

API Gateway           → per-client / per-API-key limits; right place for
                         most rate limiting logic

App layer             → fine-grained per-resource limits (e.g. "5 login
                         attempts per account per 10 min")
```

**State:** rate limit counters live in Redis. Counter-per-key with TTL is the standard pattern:

```
INCR  rl:{client_id}:{window}
EXPIRE rl:{client_id}:{window} {window_seconds}
```

In multi-region, per-region counters are the pragmatic choice — true global rate limiting requires cross-region coordination, which adds 50–150 ms per request. Accept that a determined attacker can spread load across regions; the real defence is per-IP + per-account limits combined.

**Response:** return `429 Too Many Requests` with `Retry-After` and `X-RateLimit-Remaining` headers. Never silently drop — clients need signal to back off.

## DDoS Protection

A volumetric DDoS attack sends more traffic than your infrastructure can absorb. You can't outprovision a botnet; you defend at the edge.

```
L3/L4 (volumetric):   Gbps of UDP/TCP flood → CDN absorbs (Cloudflare, CloudFront,
                       AWS Shield Advanced). Scrubbing centres divert and drop.

L7 (application):     Legitimate-looking HTTP floods → WAF rules, bot fingerprinting,
                       CAPTCHA challenge, rate limiting.
```

The architectural decision: put a CDN / WAF in front of everything public-facing. This costs $50–500/month and removes entire attack categories from your threat model.

**Amplification:** DNS and NTP can amplify a small spoofed packet into 30–50× the response traffic aimed at a victim. Mitigate by not running open resolvers and by egress-filtering your network (BCP38).

## Network Isolation

Services should not be able to reach infrastructure they have no reason to talk to.

```
Public internet
  │
  ├─ Public subnet:   Load balancers, CDN origin endpoints (public IPs)
  │
  └─ Private subnet:  App servers, databases, caches, queues (no public IPs)
                       Egress via NAT gateway only
```

**Security groups — default deny everything, open only what's needed:**

```
App server SG:   allow 443 from LB SG
                 allow 5432 from app SG (to DB)
                 deny all else

DB SG:           allow 5432 from app SG only
                 deny all — including from internet
```

**Zero trust:** network perimeter is not trusted by default. Every service-to-service call is authenticated and authorised regardless of source IP. Use mTLS between services so network position alone isn't sufficient — the service must present a valid certificate. Appropriate when lateral movement (an internal service being compromised) is a serious threat. Overhead: certificate management, ~0.2 ms for TLS handshake amortised with connection reuse.

## Secrets Management

A secret is any value that grants access: DB passwords, API keys, private keys, OAuth client secrets.

**Where secrets must NOT live:**

```
Source code / git history   → leaked forever once committed, even if later removed
Environment variables in    → visible in process listings, logs, error reports
  uncontrolled form
Config files in the repo    → same as source code
```

**The right pattern — secrets injected at runtime:**

```
At deploy time:
  Secret store (AWS SSM / Secrets Manager / HashiCorp Vault)
    → injected as env var or mounted file into the process at start
    → app reads once at startup, holds in memory
    → never written to disk, never logged

In code:
  config.dbPassword = os.Getenv("DB_PASSWORD")   ✓
  config.dbPassword = "hunter2"                   ✗
```

**Rotation without downtime:**

- DB credentials: use a secrets manager that rotates and updates the running app
- API keys: support multiple active keys simultaneously so old keys continue working during the rotation window
- TLS certificates: automate with Let's Encrypt / ACM; manual cert management always lapses

**Least privilege:** a secret grants access only to what the service needs. The API service's DB credentials should only have `SELECT, INSERT, UPDATE, DELETE` on its own schema — not `CREATE TABLE`, not access to other schemas, never admin credentials.

## Encryption

**In transit:** TLS 1.2+ everywhere, including internal service-to-service traffic. TLS 1.3 is preferred — 1-RTT handshake vs 2-RTT, no weak cipher suites.

```
Browser → LB:           TLS (terminated at LB)
LB → App:               TLS or trusted private network
App → DB:               TLS (often overlooked — enable it)
App → external API:     TLS — verify the cert; never disable certificate verification
```

**At rest:** encrypt database storage and S3 buckets. Protects against physical media theft and backup leaks, but not against a compromised application (the app holds the keys). Use AES-256.

```
AWS:    RDS encryption (KMS-managed), S3 SSE-S3 or SSE-KMS
        Cost: negligible
        When it matters: compliance requirements, protecting backup files

Application-level encryption:
        The app encrypts specific fields before storing.
        Protects against DB compromise — attacker sees ciphertext, not plaintext.
        Cost: key management complexity; can't filter/sort on encrypted fields in SQL.
        Use for: PII, payment data, anything regulated.
```

**Envelope encryption:** don't encrypt data directly with a master key. Encrypt a random data key with the master key (KMS), use the data key to encrypt the data. Rotating the master key only re-encrypts data keys — not the data itself.

```
plaintext  → encrypt with data_key        → ciphertext
data_key   → encrypt with master_key(KMS) → encrypted_data_key

Store:   ciphertext + encrypted_data_key
Decrypt: KMS.decrypt(encrypted_data_key) → data_key → decrypt(ciphertext)
```

## Input Validation

Untrusted input is any data originating outside your system: HTTP request bodies, query params, headers, uploaded files, webhook payloads, data from third-party APIs.

**Validate at the boundary. Trust internal data.**

```
SQL injection:
  db.query("SELECT * FROM users WHERE id = " + userId)
  userId = "1 OR 1=1" → dumps entire table
  Fix: parameterised queries, always.
  db.query("SELECT * FROM users WHERE id = $1", [userId])

XSS:
  Store user input, render it as raw HTML.
  Fix: escape output. Use templating libraries that escape by default.
       Never mark user content as safe/raw.

Path traversal:
  file = "/uploads/" + userInput
  userInput = "../../etc/passwd"
  Fix: validate input is a filename only; use path.Base();
       never construct filesystem paths from user input directly.

SSRF (Server-Side Request Forgery):
  App fetches a URL provided by the user.
  URL = "http://169.254.169.254/latest/meta-data/" (AWS instance metadata)
  Fix: allowlist valid target domains; block private IP ranges (RFC 1918).
```

**Schema validation:** validate the shape of incoming payloads at the API boundary (JSON Schema, Zod, Pydantic). Reject malformed input with `400 Bad Request` early — don't let it reach business logic or the database.

**File uploads:** validate MIME type by reading file headers, not trusting the `Content-Type` header. Store in an isolated bucket; never serve directly from the same origin as your app.

## Security Reference Numbers

```
TLS 1.3 handshake (new connection)     ~1 ms     (1-RTT)
TLS 1.2 handshake (new connection)     ~2 ms     (2-RTT)
TLS with session resumption            ~0.2 ms   (0-RTT)
Redis rate limit check                 ~0.5 ms   (same DC)
KMS decrypt call (envelope key)        ~5 ms     (AWS KMS API)
```

## Key Mental Models

1. **Defence in depth.** No single layer is sufficient — stack independent controls so an attacker must defeat multiple layers.
2. **Rate limit at the edge, enforce at the app.** CDN/WAF for volumetric abuse; app layer for per-account business logic limits.
3. **Secrets live in a secret store, never in code.** Inject at runtime. Rotate without downtime.
4. **Default deny networking.** Open only ports that need to be open; everything else is closed.
5. **Encrypt in transit everywhere, including internal traffic.** At rest for compliance; application-level encryption for sensitive fields.
6. **Parameterise all queries.** SQL injection is entirely preventable. There is no acceptable reason for string-concatenated SQL.
7. **Validate at the boundary, trust internal data.** Validate once at entry, then trust it downstream.
8. **SSRF is the cloud-era path traversal.** Any feature that fetches a user-supplied URL needs an allowlist or private IP blocklist.
9. **Least privilege everywhere.** DB credentials, IAM roles, API keys — each grants access only to what it needs, nothing more.
10. **Zero trust when lateral movement is a threat.** If a compromised internal service would cascade, mTLS between services removes the assumption that internal = trusted.
