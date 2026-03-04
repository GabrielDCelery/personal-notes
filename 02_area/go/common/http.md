# Go HTTP Operations

## Quick Reference

| Use case           | Method                             |
| ------------------ | ---------------------------------- |
| Simple GET         | `http.Get`                         |
| Simple POST        | `http.Post`                        |
| Custom request     | `http.NewRequest` + `client.Do`    |
| With headers/auth  | `req.Header.Set` + `client.Do`     |
| With timeout       | `http.Client{Timeout: ...}`        |
| JSON body          | `json.Marshal` + `bytes.NewBuffer` |
| Read response body | `io.ReadAll(resp.Body)`            |

## Making HTTP Requests

### 1. Simple GET (quickest)

```go
resp, err := http.Get("https://example.com/api/items")
if err != nil {
    return err
}
defer resp.Body.Close()

body, err := io.ReadAll(resp.Body)
```

### 2. Simple POST with JSON

```go
payload := []byte(`{"name":"Alice"}`)

resp, err := http.Post("https://example.com/api/items", "application/json", bytes.NewBuffer(payload))
if err != nil {
    return err
}
defer resp.Body.Close()
```

### 3. Custom request with headers

```go
req, err := http.NewRequest(http.MethodGet, "https://example.com/api/items", nil)
if err != nil {
    return err
}

req.Header.Set("Authorization", "Bearer "+token)
req.Header.Set("Accept", "application/json")

resp, err := http.DefaultClient.Do(req)
if err != nil {
    return err
}
defer resp.Body.Close()
```

### 4. POST JSON with custom headers

```go
type Item struct {
    Name string `json:"name"`
}

data, err := json.Marshal(Item{Name: "Alice"})
if err != nil {
    return err
}

req, err := http.NewRequest(http.MethodPost, "https://example.com/api/items", bytes.NewBuffer(data))
if err != nil {
    return err
}
req.Header.Set("Content-Type", "application/json")
req.Header.Set("Authorization", "Bearer "+token)

resp, err := http.DefaultClient.Do(req)
if err != nil {
    return err
}
defer resp.Body.Close()
```

### 5. With timeout (always do this in production)

```go
client := &http.Client{
    Timeout: 10 * time.Second,
}

resp, err := client.Get("https://example.com/api/items")
if err != nil {
    return err
}
defer resp.Body.Close()
```

### 6. Decode JSON response

```go
type Item struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

resp, err := http.Get("https://example.com/api/items/1")
if err != nil {
    return err
}
defer resp.Body.Close()

var item Item
if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
    return err
}
```

### 7. Check status code

```go
resp, err := http.Get("https://example.com/api/items")
if err != nil {
    return err
}
defer resp.Body.Close()

if resp.StatusCode != http.StatusOK {
    return fmt.Errorf("unexpected status: %s", resp.Status)
}
```
