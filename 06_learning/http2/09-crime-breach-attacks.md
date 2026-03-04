# What Are CRIME/BREACH Attacks?

## Question
What does this mean: "HTTP/2 deliberately does not compress DATA frames (the body) because: CRIME/BREACH attacks: Compression + encryption can leak information"?

## Answer

Great question! This is about a subtle but serious security vulnerability. Let me explain how compression + encryption can leak secrets.

### The CRIME/BREACH Attack Concept

The attack exploits a seemingly innocent fact: **compressed data size reveals information about the content**.

### How It Works

**Setup:**
1. Attacker can inject partial data into your requests (e.g., via JavaScript on a page you visit)
2. The request also contains a secret (e.g., session cookie)
3. The data gets compressed, then encrypted, then sent
4. Attacker can see the **size** of the encrypted data (but not the content)

**The Attack:**

```
Original request (no compression):
Cookie: sessionid=SECRET123
Custom-Header: A
[Size: 100 bytes after encryption]

Compressed + Encrypted:
Cookie: sessionid=SECRET123
Custom-Header: S
[Size: 95 bytes - no match, no compression benefit]

Compressed + Encrypted:
Cookie: sessionid=SECRET123
Custom-Header: SE
[Size: 93 bytes - getting warmer!]

Compressed + Encrypted:
Cookie: sessionid=SECRET123
Custom-Header: SEC
[Size: 90 bytes - good compression! "SEC" appears twice]
```

**Key insight:** Compression algorithms (like gzip) look for repeated patterns. When the attacker's injected data matches part of the secret, the compressed size **gets smaller** because the pattern appears twice.

### Concrete Example

Let's say your secret cookie is `auth=SuperSecret`

**Attacker tries:**
```
Attempt 1: Inject "a" → Compressed size: 150 bytes
Attempt 2: Inject "S" → Compressed size: 145 bytes ✓ (match!)
Attempt 3: Inject "Su" → Compressed size: 142 bytes ✓ (match!)
Attempt 4: Inject "Sup" → Compressed size: 138 bytes ✓ (match!)
...continue until full secret recovered...
```

By trying each character and measuring the compressed size, the attacker can recover the secret **one byte at a time**, even though everything is encrypted!

### Why Encryption Doesn't Protect Here

Encryption makes the content unreadable, but it **doesn't hide the length**:
- Encrypted data still has observable size
- TLS packets show their length in plaintext
- Attacker can measure: "this request was 150 bytes, that one was 145 bytes"

### The CRIME/BREACH Attacks

**CRIME (2012):**
- Targeted TLS compression
- Compressed entire TLS stream including headers
- **Solution:** Browsers disabled TLS compression

**BREACH (2013):**
- Targeted HTTP-level compression (like `Content-Encoding: gzip`)
- Compressed HTTP response bodies
- Harder to fix because applications control this, not the protocol

### Why HTTP/2 Learned From This

HTTP/2 designers knew about these attacks, so they made conscious choices:

**HPACK (header compression):**
- Uses **static dictionaries** and **indexing**, not pattern-matching compression
- More predictable size behavior
- Carefully designed to avoid CRIME/BREACH-style attacks

**Body (DATA frames):**
- **Not compressed by HTTP/2**
- Left to application layer where developers can:
  - Choose when to compress
  - Avoid compressing secrets
  - Use techniques like separating secrets from user-controllable data

### How To Protect Yourself

If you use `Content-Encoding: gzip`:

1. **Don't compress secrets mixed with user input**
2. **Use CSRF tokens** to prevent attackers from making requests
3. **Separate secrets from reflected content**
4. **Rate limit** to make brute-force attacks impractical
5. **Add random padding** to blur size differences

### Summary

- **Compression reveals patterns** through size changes
- **Encryption doesn't hide size**, so attacker can measure
- **Attacker injects guesses**, observes which guess compresses best
- **HTTP/2 avoided this** by not compressing bodies
- **Application-level compression** (Content-Encoding) is still vulnerable if misused

This is a great example of how security isn't just about encryption - side channels like **size, timing, power consumption** can all leak information!
