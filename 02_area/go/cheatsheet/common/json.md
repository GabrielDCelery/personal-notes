# Go JSON

## Why

- **Struct tags** — Go's JSON encoder uses reflection. Without tags, fields export as their Go name (`CreatedAt`). Tags control the wire format (`created_at`) and hide fields (`-`) without changing Go code.
- **Marshal vs Encoder** — Marshal works on `[]byte` (in-memory). Encoder streams to an `io.Writer`. Use Encoder for HTTP responses — no intermediate byte slice allocation.
- **omitempty** — Prevents sending `"email":""` or `"count":0` over the wire. Useful for PATCH APIs where absence means "don't update".
- **RawMessage** — Defers decoding. Use when you need to inspect a discriminator field (like `type`) before knowing what shape the rest of the JSON is.
- **Custom Marshaler** — The default encoder handles basic types. Implement MarshalJSON/UnmarshalJSON when your type needs a non-obvious wire format (enums as strings, custom date formats).

## Quick Reference

| Use case                 | Method                                     |
| ------------------------ | ------------------------------------------ |
| Encode struct to JSON    | `json.Marshal` / `json.NewEncoder`         |
| Decode JSON to struct    | `json.Unmarshal` / `json.NewDecoder`       |
| Ignore field             | `json:"-"`                                 |
| Omit if zero/empty       | `json:",omitempty"`                        |
| Custom field name        | `json:"my_name"`                           |
| Unknown fields           | `map[string]any` / `json.RawMessage`       |
| Custom marshal/unmarshal | implement `json.Marshaler` / `Unmarshaler` |

## Basics

### 1. Struct tags

```go
type User struct {
    ID        int    `json:"id"`
    Name      string `json:"name"`
    Email     string `json:"email,omitempty"` // omit if empty string
    Password  string `json:"-"`               // never include
    CreatedAt time.Time `json:"created_at"`
}
```

### 2. Marshal (struct → JSON)

```go
u := User{ID: 1, Name: "Alice"}

data, err := json.Marshal(u)
if err != nil {
    return err
}
fmt.Println(string(data)) // {"id":1,"name":"Alice"}
```

### 3. Unmarshal (JSON → struct)

```go
data := []byte(`{"id":1,"name":"Alice"}`)

var u User
if err := json.Unmarshal(data, &u); err != nil {
    return err
}
```

## Streaming (preferred for HTTP)

### 4. Encode to writer

```go
// Writing JSON response
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(u)
```

### 5. Decode from reader

```go
// Reading JSON request body
var u User
if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
    http.Error(w, "invalid body", http.StatusBadRequest)
    return
}
```

## Working with Unknown Structure

### 6. Decode into map

```go
var m map[string]any
json.Unmarshal(data, &m)

name := m["name"].(string)
```

### 7. RawMessage (delay decoding)

```go
type Envelope struct {
    Type    string          `json:"type"`
    Payload json.RawMessage `json:"payload"` // kept as raw JSON
}

var env Envelope
json.Unmarshal(data, &env)

// Decode payload based on type
switch env.Type {
case "user":
    var u User
    json.Unmarshal(env.Payload, &u)
}
```

## Custom Marshal / Unmarshal

### 8. Custom marshaler

```go
type Color int

const (
    Red Color = iota
    Green
    Blue
)

func (c Color) MarshalJSON() ([]byte, error) {
    names := map[Color]string{Red: "red", Green: "green", Blue: "blue"}
    return json.Marshal(names[c])
}

func (c *Color) UnmarshalJSON(data []byte) error {
    var s string
    if err := json.Unmarshal(data, &s); err != nil {
        return err
    }
    codes := map[string]Color{"red": Red, "green": Green, "blue": Blue}
    *c = codes[s]
    return nil
}
```

### 9. Custom time format

```go
type CustomTime struct {
    time.Time
}

func (ct CustomTime) MarshalJSON() ([]byte, error) {
    return json.Marshal(ct.Format("2006-01-02"))
}

func (ct *CustomTime) UnmarshalJSON(data []byte) error {
    var s string
    if err := json.Unmarshal(data, &s); err != nil {
        return err
    }
    t, err := time.Parse("2006-01-02", s)
    if err != nil {
        return err
    }
    ct.Time = t
    return nil
}
```

## Pretty Print

### 10. Indented output

```go
data, _ := json.MarshalIndent(u, "", "  ")
fmt.Println(string(data))
```
