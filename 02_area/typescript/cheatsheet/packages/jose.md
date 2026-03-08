# jose (JWT)

```sh
npm install jose
```

> Universal JavaScript JWT library. Works in Node.js, browsers, Deno, Bun. No native dependencies. Supports JWS, JWE, JWK, JWKS.

## Sign a JWT

```typescript
import { SignJWT } from "jose";

const secret = new TextEncoder().encode(process.env.JWT_SECRET);

const token = await new SignJWT({ userId: "123", role: "admin" })
  .setProtectedHeader({ alg: "HS256" })
  .setIssuedAt()
  .setExpirationTime("2h")
  .setIssuer("my-service")
  .setAudience("my-api")
  .sign(secret);
```

## Verify a JWT

```typescript
import { jwtVerify } from "jose";

const secret = new TextEncoder().encode(process.env.JWT_SECRET);

try {
  const { payload } = await jwtVerify(token, secret, {
    issuer: "my-service",
    audience: "my-api",
  });

  console.log(payload.userId); // "123"
  console.log(payload.role); // "admin"
  console.log(payload.exp); // expiration timestamp
} catch (err) {
  // JWTExpired, JWTClaimValidationFailed, JWSSignatureVerificationFailed, etc.
  console.error("Invalid token:", err);
}
```

## Asymmetric Keys (RS256)

```typescript
import { SignJWT, jwtVerify, importPKCS8, importSPKI } from "jose";

// Load keys
const privateKey = await importPKCS8(process.env.JWT_PRIVATE_KEY!, "RS256");
const publicKey = await importSPKI(process.env.JWT_PUBLIC_KEY!, "RS256");

// Sign with private key
const token = await new SignJWT({ userId: "123" })
  .setProtectedHeader({ alg: "RS256" })
  .setExpirationTime("1h")
  .sign(privateKey);

// Verify with public key
const { payload } = await jwtVerify(token, publicKey);
```

## JWKS (JSON Web Key Set)

```typescript
import { createRemoteJWKSet, jwtVerify } from "jose";

// Fetch and cache keys from an identity provider
const JWKS = createRemoteJWKSet(
  new URL("https://auth.example.com/.well-known/jwks.json"),
);

const { payload } = await jwtVerify(token, JWKS, {
  issuer: "https://auth.example.com/",
  audience: "my-api",
});
```

## Decode Without Verification (inspect only)

```typescript
import { decodeJwt, decodeProtectedHeader } from "jose";

const payload = decodeJwt(token); // { userId: "123", exp: ..., iat: ... }
const header = decodeProtectedHeader(token); // { alg: "HS256", typ: "JWT" }

// WARNING: no signature verification — only use for debugging or routing
```

## Express Auth Middleware

```typescript
import { jwtVerify } from "jose";

const secret = new TextEncoder().encode(process.env.JWT_SECRET);

async function authMiddleware(req: Request, res: Response, next: NextFunction) {
  const header = req.headers.authorization;
  if (!header?.startsWith("Bearer ")) {
    res.status(401).json({ error: "Missing token" });
    return;
  }

  try {
    const token = header.slice(7);
    const { payload } = await jwtVerify(token, secret);
    (req as any).user = payload;
    next();
  } catch {
    res.status(401).json({ error: "Invalid token" });
  }
}
```

## Refresh Token Pattern

```typescript
async function generateTokens(userId: string) {
  const accessToken = await new SignJWT({ userId })
    .setProtectedHeader({ alg: "HS256" })
    .setExpirationTime("15m")
    .sign(secret);

  const refreshToken = await new SignJWT({ userId, type: "refresh" })
    .setProtectedHeader({ alg: "HS256" })
    .setExpirationTime("7d")
    .sign(refreshSecret);

  return { accessToken, refreshToken };
}

async function refreshAccessToken(refreshToken: string) {
  const { payload } = await jwtVerify(refreshToken, refreshSecret);
  if (payload.type !== "refresh") throw new Error("Invalid token type");

  return new SignJWT({ userId: payload.userId })
    .setProtectedHeader({ alg: "HS256" })
    .setExpirationTime("15m")
    .sign(secret);
}
```

## Common Algorithms

| Algorithm | Type                      | Use when                              |
| --------- | ------------------------- | ------------------------------------- |
| `HS256`   | Symmetric (shared secret) | Single service, simple setup          |
| `RS256`   | Asymmetric (RSA)          | Multiple services verify, one signs   |
| `ES256`   | Asymmetric (ECDSA)        | Same as RS256 but smaller keys/tokens |
| `EdDSA`   | Asymmetric (Ed25519)      | Best performance, modern choice       |
