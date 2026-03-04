# Go Crypto

## Quick Reference

| Use case       | Package / Method               |
| -------------- | ------------------------------ |
| SHA-256 hash   | `crypto/sha256`                |
| HMAC           | `crypto/hmac`                  |
| Random bytes   | `crypto/rand`                  |
| Random int     | `crypto/rand` + `math/big`     |
| Bcrypt         | `golang.org/x/crypto/bcrypt`   |
| AES encryption | `crypto/aes` + `crypto/cipher` |
| Base64 encode  | `encoding/base64`              |
| Hex encode     | `encoding/hex`                 |

## Hashing

### 1. SHA-256

```go
data := []byte("hello world")
hash := sha256.Sum256(data)

fmt.Println(hex.EncodeToString(hash[:]))
```

### 2. SHA-256 with streaming (large data)

```go
h := sha256.New()
h.Write([]byte("hello "))
h.Write([]byte("world"))
hash := h.Sum(nil)

fmt.Println(hex.EncodeToString(hash))
```

### 3. MD5 (non-security uses only)

```go
hash := md5.Sum([]byte("hello"))
fmt.Println(hex.EncodeToString(hash[:]))
```

## HMAC

### 4. Create and verify HMAC

```go
key := []byte("secret-key")
message := []byte("hello world")

mac := hmac.New(sha256.New, key)
mac.Write(message)
signature := mac.Sum(nil)

// Verify — constant-time comparison
mac2 := hmac.New(sha256.New, key)
mac2.Write(message)
expected := mac2.Sum(nil)
fmt.Println(hmac.Equal(signature, expected))  // true
```

## Random

### 5. Random bytes

```go
b := make([]byte, 32)
_, err := rand.Read(b)
if err != nil {
    log.Fatal(err)
}
fmt.Println(hex.EncodeToString(b))
```

### 6. Random token (URL-safe)

```go
b := make([]byte, 32)
rand.Read(b)
token := base64.URLEncoding.EncodeToString(b)
```

### 7. Random integer in range

```go
n, err := rand.Int(rand.Reader, big.NewInt(100))  // [0, 100)
fmt.Println(n.Int64())
```

## Bcrypt (password hashing)

```sh
go get golang.org/x/crypto/bcrypt
```

### 8. Hash and verify password

```go
password := []byte("hunter2")

hash, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
if err != nil {
    log.Fatal(err)
}

// Verify
err = bcrypt.CompareHashAndPassword(hash, []byte("hunter2"))
if err != nil {
    // wrong password
}
```

## AES Encryption

### 9. AES-GCM encrypt/decrypt

```go
func encrypt(plaintext, key []byte) ([]byte, error) {
    block, err := aes.NewCipher(key)  // key must be 16, 24, or 32 bytes
    if err != nil {
        return nil, err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }

    nonce := make([]byte, gcm.NonceSize())
    if _, err := rand.Read(nonce); err != nil {
        return nil, err
    }

    return gcm.Seal(nonce, nonce, plaintext, nil), nil  // nonce prepended
}

func decrypt(ciphertext, key []byte) ([]byte, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }

    nonceSize := gcm.NonceSize()
    nonce, ct := ciphertext[:nonceSize], ciphertext[nonceSize:]

    return gcm.Open(nil, nonce, ct, nil)
}
```

## Base64 / Hex Encoding

### 10. Encode and decode

```go
// Base64
encoded := base64.StdEncoding.EncodeToString([]byte("hello"))
decoded, err := base64.StdEncoding.DecodeString(encoded)

// URL-safe base64
encoded := base64.URLEncoding.EncodeToString(data)

// Hex
encoded := hex.EncodeToString([]byte("hello"))
decoded, err := hex.DecodeString(encoded)
```
