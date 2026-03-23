# Auth Patterns — Distilled

Auth is deciding "who is this request from, and what are they allowed to do?" — at minimal latency cost, at every request, across every service.

## The Core Decision: Stateless vs Stateful

**Stateless (JWT):** identity is encoded in the token. Verification is local — no network call.
**Stateful (session):** token is an opaque ID. Identity lives in a store. Every request pays a lookup.

```
Stateless (JWT):
  Request → App → verify signature locally → done
                                              ~0.1 ms

Stateful (session):
  Request → App → Redis lookup → done
                                  ~1–2 ms   (same DC)
  Request → App → DB lookup → done
                                  ~5 ms     (slow path)
```

The difference compounds in microservices: 5 services each doing a session lookup = 5 round trips per request.

**Default to JWT for stateless APIs. Default to sessions for server-rendered apps where instant revocation matters.**

## JWT Anatomy

A JWT is three base64url-encoded JSON objects separated by dots: `header.payload.signature`

```
Header:    { "alg": "RS256", "typ": "JWT" }
Payload:   { "sub": "user_123", "roles": ["admin"], "exp": 1711000000 }
Signature: RSASHA256(base64(header) + "." + base64(payload), privateKey)
```

Verification is pure CPU: decode the payload, verify the signature against a locally held public key. No network call. The payload is readable by anyone — **never put secrets in it.**

The public key is fetched once from the auth server's JWKS endpoint and cached locally:

```
GET https://auth.example.com/.well-known/jwks.json
→ { "keys": [{ "kty": "RSA", "kid": "key-1", "n": "...", "e": "AQAB" }] }
```

The `kid` claim in the JWT header identifies which key to verify against — allows key rotation without downtime.

| Claim | Meaning                          |
| ----- | -------------------------------- |
| `sub` | Subject — the user ID            |
| `exp` | Expiry — Unix timestamp          |
| `iat` | Issued at                        |
| `iss` | Issuer — who created it          |
| `aud` | Audience — who it's intended for |

## The Revocation Problem

**JWTs can't be revoked before expiry.** This is the fundamental trade-off.

Once issued, the token is valid until `exp`. If a user logs out, gets compromised, or has permissions changed, the JWT is still accepted everywhere until it expires.

Solutions — in order of preference:

```
Short expiry (15 min) + refresh tokens     → accept a small revocation window
Token blocklist in Redis                   → instant revocation, reintroduces network call
Opaque reference tokens + introspection    → full revocation, full latency cost
```

**Short expiry + refresh rotation is the right default.** A 15-minute window is acceptable for most apps. Only use a blocklist if you have a hard revocation requirement (financial, healthcare).

## Access Tokens and Refresh Tokens

An **access token is a JWT** — self-contained, verified locally, short-lived.
A **refresh token is an opaque string** — a random ID stored server-side, looked up in DB or Redis.

They use different formats because they have different requirements:

|               | Access token       | Refresh token               |
| ------------- | ------------------ | --------------------------- |
| **Format**    | JWT                | Opaque random string        |
| **Lifetime**  | 15 min             | 7 days                      |
| **Verified**  | Locally (CPU only) | Server-side DB/Redis lookup |
| **Revocable** | No                 | Yes — delete the row        |
| **Exposure**  | Every API request  | Only on token refresh       |

The refresh token travels rarely (only when the access token expires), so it's safer to make it long-lived. Because it lives server-side, it can be revoked instantly.

```
Login:
  Client → Auth Server → access_token (JWT, 15 min) + refresh_token (opaque, 7 days)

Normal requests (for 15 min):
  Client → Your API (access_token) → verified locally, auth server not involved

Access token expires:
  Client → Auth Server (refresh_token) → new access_token + new refresh_token
                                          (rotate: old refresh token invalidated)

Repeat until refresh token expires or user logs out (refresh token deleted).
```

Refresh token rotation: each use issues a new refresh token and invalidates the old one. If a stolen token is used, the legitimate user's next refresh fails — detectable.

## Bearer Tokens and API Keys

**Bearer token** is a transport mechanism, not a format. It means "whoever holds this string, trust them."

```
Authorization: Bearer eyJhbGci...
```

Your access token (a JWT) is sent as a Bearer token. The three concepts are different dimensions of the same thing:

```
format:    JWT            (how the token is structured)
role:      access token   (what it represents)
transport: Bearer token   (how it travels in HTTP)
```

**API keys** serve a similar purpose but for a different use case — machine-to-machine (M2M) auth where there's no human login flow.

|                 | Bearer token (JWT)              | API key                                   |
| --------------- | ------------------------------- | ----------------------------------------- |
| **Lifetime**    | Short-lived (15 min)            | Long-lived (months/years)                 |
| **Contains**    | Identity + permissions + expiry | Opaque ID, looked up server-side          |
| **Issued by**   | Login / OAuth flow              | Manually generated in a dashboard         |
| **Who uses it** | Users (after login)             | Services, developers integrating your API |
| **Revocation**  | Expires naturally               | Must be explicitly revoked                |

```
User flow:     login → JWT (15 min) → refresh → repeat
API key flow:  generate once → store securely → send on every request
```

API keys are essentially long-lived opaque tokens. Because they don't expire, they're higher risk if leaked — scope them to specific permissions and support rotation.

## Where to Store Tokens

| Storage                  | XSS risk                   | CSRF risk                                 | Use for        |
| ------------------------ | -------------------------- | ----------------------------------------- | -------------- |
| **httpOnly cookie**      | None (JS can't read)       | Yes — mitigate with SameSite + CSRF token | Refresh tokens |
| **Memory (JS variable)** | Low (lost on page refresh) | None                                      | Access tokens  |
| **localStorage**         | High (any JS can read)     | None                                      | Avoid          |
| **sessionStorage**       | High                       | None                                      | Avoid          |

**httpOnly + SameSite=Strict cookie for refresh tokens. In-memory for access tokens. Never localStorage.**

CSRF with cookies: `SameSite=Strict` blocks cross-origin requests entirely. For legitimate cross-origin flows, use `SameSite=Lax` + a CSRF token in a request header alongside the cookie.

## Sessions

Session auth: the server stores state, the client holds an opaque session ID in a cookie.

```
Login:   Client → Server → store { session_id: "abc", user_id: 123, roles: [...] } in Redis
         Server → Client: Set-Cookie: session_id=abc; HttpOnly; Secure

Request: Client → Server (cookie auto-sent) → Redis lookup → identity
```

Advantages over JWT: instant revocation (delete the session row), no token size limits, simpler mental model.

**Sessions need shared storage.** Store in Redis, not local memory — requests hit different servers. A session in local memory is invisible to the other 9 app servers behind your load balancer.

## OAuth / OIDC

OAuth delegates auth to a third party (Google, GitHub, Auth0). You get an access token for their API and — with OIDC — an `id_token` containing the user's identity.

```
User → Your App → redirect to Google
Google: user logs in, grants consent
Google → Your App: authorization_code (short-lived, single-use)
Your App → Google: exchange code for tokens  (server-side)
Google → Your App: access_token + id_token
Your App: verify id_token, create your own session or JWT
```

**The code exchange must happen server-side** — never expose client secrets in a browser.

Cost: the OAuth callback involves 1–2 HTTP calls to the provider (10–100 ms each). **Verify once at login, then issue your own session or JWT.** Never call the provider on every request.

## Microservices

```
Incoming request
  → Gateway: verify JWT once
  → pass trusted internal token downstream

  Service A → Service B → Service C
  (no auth calls — trust the internal token)
```

Never make each downstream service independently call an external auth provider. Each hop adds 10–100 ms and creates a dependency on the auth provider's availability.

For the internal token: pass the original JWT (downstream services verify the signature locally) or issue an internal opaque token that services trust without verification.

## Auth Latency Reference

```
Scale: log10  |  0.1ms    1ms       10ms      100ms |
              0----+--------+---------+---------+---+

JWT (local verify)          |██                  | ~0.1 ms   no network call
Session → Redis             |████████            | ~1–2 ms   one round trip
Session → DB                |████████████        | ~5 ms     network + query
OAuth introspect            |████████████████████| 10–100 ms HTTP call to provider
```

## Key Mental Models

1. **JWT = no network call.** Verification is CPU-only — decode, check signature against cached public key, read claims.
2. **Access token = JWT. Refresh token = opaque string.** Different formats because they have different requirements.
3. **Bearer token is transport, not format.** Your JWT access token travels as a Bearer token in the Authorization header.
4. **API keys are for M2M, not users.** Long-lived, manually issued, no login flow — higher risk if leaked.
5. **Short expiry + refresh rotation is the standard.** 15-min access token, 7-day rotating refresh token.
6. **httpOnly cookie for refresh tokens. Memory for access tokens. Never localStorage.**
7. **JWTs can't be revoked before expiry.** Accept the window, or use a blocklist (which reintroduces a network call). If you need instant revocation on every request, use sessions instead.
8. **Verify once at the gateway.** Don't re-verify at every downstream service.
9. **OAuth: exchange the code server-side, then issue your own token.** Don't call the provider per request.
10. **Sessions need shared storage.** Redis, not local memory — load balancers don't guarantee the same server.
