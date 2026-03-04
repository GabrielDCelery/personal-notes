# When is Encryption Established?

## Question
When is encryption being established in this?

## Answer

Great question! Encryption (TLS/SSL) is established **after** the TCP connection but **before** any HTTP data flows.

### Complete Sequence with HTTPS

```
Client                          Server
  |                               |
  |=== TCP Handshake ===          |
  |  1. SYN                       |
  |----------------------------->|
  |  2. SYN-ACK                   |
  |<-----------------------------|
  |  3. ACK                       |
  |----------------------------->|
  |                               |
  | ✓ TCP Connection Established |
  |                               |
  |=== TLS Handshake ===          |
  |  4. ClientHello               |
  |----------------------------->|
  |  5. ServerHello + Certificate |
  |<-----------------------------|
  |  6. Key Exchange + Finished   |
  |----------------------------->|
  |  7. Finished                  |
  |<-----------------------------|
  |                               |
  | ✓ Encrypted Channel Ready    |
  |                               |
  |=== Application Data ===       |
  |  8. HTTP Request (encrypted)  |
  |----------------------------->|
  |  9. HTTP Response (encrypted) |
  |<-----------------------------|
```

### TLS Handshake Details (TLS 1.2)

**ClientHello:**
- Supported cipher suites
- TLS version
- Random data for key generation

**ServerHello:**
- Chosen cipher suite
- Server's certificate (public key)
- Random data

**Key Exchange:**
- Client generates a pre-master secret
- Encrypts it with server's public key
- Both sides derive session keys from this

**Finished Messages:**
- Both sides verify the handshake succeeded
- These are the first encrypted messages

### TLS 1.3 (Faster)

TLS 1.3 reduces this to **1 round trip** instead of 2:

```
Client                          Server
  |  ClientHello + KeyShare      |
  |----------------------------->|
  |  ServerHello + KeyShare +    |
  |  Certificate + Finished      |
  |<-----------------------------|
  |  Finished                    |
  |----------------------------->|
  | (Application data can start) |
```

### Total Round Trips for HTTPS

**For TLS 1.2:**
- TCP: 1.5 round trips (SYN, SYN-ACK, ACK+data)
- TLS: 2 round trips
- **Total: ~3.5 round trips before HTTP data**

**For TLS 1.3:**
- TCP: 1.5 round trips
- TLS: 1 round trip
- **Total: ~2.5 round trips before HTTP data**

### HTTP vs HTTPS

- **HTTP (port 80)**: TCP handshake → immediate HTTP data
- **HTTPS (port 443)**: TCP handshake → TLS handshake → encrypted HTTP data

This is why HTTPS connections feel slightly slower on the first request, but modern optimizations like **TLS session resumption** and **0-RTT** (TLS 1.3) help reduce this overhead on subsequent connections!
