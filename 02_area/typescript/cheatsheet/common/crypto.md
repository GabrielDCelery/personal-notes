# TypeScript Crypto

## Why

- **node:crypto not Math.random()** — Math.random() is a predictable PRNG. Use `crypto.randomBytes` or `crypto.randomUUID` for tokens, keys, and anything security-related.
- **bcrypt/scrypt for passwords** — SHA-256 is fast by design, which makes brute-force easy. Use bcrypt or scrypt — they're intentionally slow and include a salt. Never store passwords with SHA/MD5.
- **timingSafeEqual for verification** — Regular `===` comparison is vulnerable to timing attacks. `crypto.timingSafeEqual` uses constant-time comparison. Use it for HMAC verification and token comparison.
- **AES-256-GCM over AES-CBC** — GCM provides both encryption and authentication (integrity check). CBC only encrypts — tampering goes undetected. Always prefer authenticated encryption.
- **IV/nonce must be unique** — Reusing an IV with the same key in AES-GCM completely breaks the security. Generate a fresh random IV for every encryption.

## Quick Reference

| Use case            | Method                                 |
| ------------------- | -------------------------------------- |
| SHA-256 hash        | `crypto.createHash("sha256")`          |
| HMAC                | `crypto.createHmac("sha256", key)`     |
| Random bytes        | `crypto.randomBytes(32)`               |
| Random UUID         | `crypto.randomUUID()`                  |
| Password hash       | `bcrypt` or `crypto.scrypt`            |
| AES encryption      | `crypto.createCipheriv`                |
| Base64 encode       | `Buffer.from(data).toString("base64")` |
| Hex encode          | `Buffer.from(data).toString("hex")`    |
| Timing-safe compare | `crypto.timingSafeEqual`               |

## Hashing

### 1. SHA-256

```typescript
import { createHash } from "node:crypto";

const hash = createHash("sha256").update("hello world").digest("hex");
// "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
```

### 2. Hash with streaming (large data)

```typescript
import { createHash } from "node:crypto";
import { createReadStream } from "node:fs";
import { pipeline } from "node:stream/promises";

const hash = createHash("sha256");
await pipeline(createReadStream("large-file.bin"), hash);
console.log(hash.digest("hex"));
```

### 3. MD5 (non-security uses only — checksums, cache keys)

```typescript
const hash = createHash("md5").update("hello").digest("hex");
```

## HMAC

### 4. Create and verify HMAC

```typescript
import { createHmac, timingSafeEqual } from "node:crypto";

const secret = "my-secret-key";
const message = "hello world";

// Create
const signature = createHmac("sha256", secret).update(message).digest("hex");

// Verify — timing-safe
function verifyHmac(
  message: string,
  signature: string,
  secret: string,
): boolean {
  const expected = createHmac("sha256", secret).update(message).digest();
  const actual = Buffer.from(signature, "hex");
  if (expected.length !== actual.length) return false;
  return timingSafeEqual(expected, actual);
}
```

## Random

### 5. Random bytes and tokens

```typescript
import { randomBytes, randomUUID } from "node:crypto";

const bytes = randomBytes(32);
const token = randomBytes(32).toString("hex"); // 64 char hex string
const urlSafe = randomBytes(32).toString("base64url"); // URL-safe token
const uuid = randomUUID(); // "a1b2c3d4-..."
```

### 6. Random integer in range

```typescript
import { randomInt } from "node:crypto";

const n = randomInt(100); // [0, 100)
const n = randomInt(10, 20); // [10, 20)
```

## Password Hashing

### 7. bcrypt (via bcryptjs — pure JS, no native deps)

```sh
npm install bcryptjs
npm install -D @types/bcryptjs
```

```typescript
import { hash, compare } from "bcryptjs";

const hashed = await hash("hunter2", 12); // 12 rounds
const valid = await compare("hunter2", hashed); // true
```

### 8. scrypt (built-in, no dependencies)

```typescript
import { scrypt, randomBytes, timingSafeEqual } from "node:crypto";
import { promisify } from "node:util";

const scryptAsync = promisify(scrypt);

async function hashPassword(password: string): Promise<string> {
  const salt = randomBytes(16).toString("hex");
  const derived = (await scryptAsync(password, salt, 64)) as Buffer;
  return `${salt}:${derived.toString("hex")}`;
}

async function verifyPassword(
  password: string,
  stored: string,
): Promise<boolean> {
  const [salt, hash] = stored.split(":");
  const derived = (await scryptAsync(password, salt, 64)) as Buffer;
  return timingSafeEqual(derived, Buffer.from(hash, "hex"));
}
```

## AES Encryption

### 9. AES-256-GCM encrypt/decrypt

```typescript
import { createCipheriv, createDecipheriv, randomBytes } from "node:crypto";

function encrypt(plaintext: string, key: Buffer): string {
  const iv = randomBytes(12); // 96-bit IV for GCM
  const cipher = createCipheriv("aes-256-gcm", key, iv);

  const encrypted = Buffer.concat([
    cipher.update(plaintext, "utf-8"),
    cipher.final(),
  ]);
  const authTag = cipher.getAuthTag();

  // Pack: iv + authTag + ciphertext
  return Buffer.concat([iv, authTag, encrypted]).toString("base64");
}

function decrypt(packed: string, key: Buffer): string {
  const data = Buffer.from(packed, "base64");

  const iv = data.subarray(0, 12);
  const authTag = data.subarray(12, 28);
  const encrypted = data.subarray(28);

  const decipher = createDecipheriv("aes-256-gcm", key, iv);
  decipher.setAuthTag(authTag);

  return decipher.update(encrypted) + decipher.final("utf-8");
}

// Key must be 32 bytes for AES-256
const key = randomBytes(32);
const encrypted = encrypt("secret message", key);
const decrypted = decrypt(encrypted, key);
```

## Base64 / Hex Encoding

### 10. Encode and decode

```typescript
// Base64
const encoded = Buffer.from("hello").toString("base64"); // "aGVsbG8="
const decoded = Buffer.from(encoded, "base64").toString("utf-8"); // "hello"

// URL-safe base64
const urlSafe = Buffer.from("hello").toString("base64url"); // "aGVsbG8"

// Hex
const hex = Buffer.from("hello").toString("hex"); // "68656c6c6f"
const fromHex = Buffer.from(hex, "hex").toString("utf-8"); // "hello"
```
