# Go Structs

## Why

- **No constructors** — Go has no `new` keyword like Java/C#. The `NewX` function pattern is just a convention. Use it when you need to set defaults, validate, or allocate internal state.
- **Pointer receivers** — Use pointer receivers when the method modifies the struct or the struct is large. If any method needs a pointer receiver, use pointer receivers for all methods on that type.
- **Embedding is composition, not inheritance** — Embedded fields promote their methods. The outer struct "has-a" inner struct, not "is-a". There is no polymorphism through embedding.
- **Struct tags** — Metadata read via reflection. The `json`, `db`, and `yaml` tags are the most common. They control serialization without changing your field names.
- **Functional options** — The `WithX(val)` pattern solves the "too many constructor parameters" problem. Each option is self-documenting and optional, with sensible defaults.

## Quick Reference

| Use case            | Method                               |
| ------------------- | ------------------------------------ |
| Define struct       | `type Name struct { ... }`           |
| Create instance     | `Name{field: val}`                   |
| Pointer to struct   | `&Name{field: val}`                  |
| Constructor pattern | `func NewName(...) *Name`            |
| Embedding           | anonymous field inside struct        |
| Tags                | backtick annotations on fields       |
| Method              | `func (r Receiver) Method() { ... }` |

## Basics

### 1. Define and create

```go
type User struct {
    ID    int
    Name  string
    Email string
}

u := User{ID: 1, Name: "Alice", Email: "alice@example.com"}
u := User{}           // zero value — all fields zeroed
p := &User{ID: 1}     // pointer to struct
```

### 2. Constructor pattern

```go
func NewUser(name, email string) *User {
    return &User{
        ID:    generateID(),
        Name:  name,
        Email: email,
    }
}
```

### 3. Methods — value vs pointer receiver

```go
// Value receiver — works on a copy
func (u User) FullName() string {
    return u.Name
}

// Pointer receiver — can modify the struct
func (u *User) SetEmail(email string) {
    u.Email = email
}
```

> If any method needs a pointer receiver, use pointer receivers for all methods on that type for consistency.

## Struct Tags

### 4. Common tags

```go
type User struct {
    ID        int    `json:"id" db:"id"`
    Name      string `json:"name" db:"name"`
    Email     string `json:"email,omitempty" db:"email"`
    Password  string `json:"-"`                          // excluded from JSON
    CreatedAt string `json:"created_at" db:"created_at"`
}
```

### 5. Read tags with reflection

```go
t := reflect.TypeOf(User{})
field, _ := t.FieldByName("Email")
fmt.Println(field.Tag.Get("json"))  // "email,omitempty"
fmt.Println(field.Tag.Get("db"))    // "email"
```

## Embedding

### 6. Embed structs (composition)

```go
type Address struct {
    City    string
    Country string
}

type User struct {
    Name string
    Address              // embedded — fields promoted
}

u := User{
    Name:    "Alice",
    Address: Address{City: "London", Country: "UK"},
}

fmt.Println(u.City)     // promoted — no need for u.Address.City
```

### 7. Embed interfaces

```go
type Logger interface {
    Log(msg string)
}

type Service struct {
    Logger               // embedded interface — must be set before use
    Name string
}

s := Service{Logger: myLogger, Name: "auth"}
s.Log("started")        // delegates to embedded Logger
```

## Patterns

### 8. Functional options

```go
type Server struct {
    host    string
    port    int
    timeout time.Duration
}

type Option func(*Server)

func WithPort(port int) Option {
    return func(s *Server) { s.port = port }
}

func WithTimeout(d time.Duration) Option {
    return func(s *Server) { s.timeout = d }
}

func NewServer(host string, opts ...Option) *Server {
    s := &Server{host: host, port: 8080, timeout: 30 * time.Second}
    for _, opt := range opts {
        opt(s)
    }
    return s
}

s := NewServer("localhost", WithPort(9090), WithTimeout(5*time.Second))
```

### 9. Compare structs

```go
// Structs are comparable if all fields are comparable
a := User{ID: 1, Name: "Alice"}
b := User{ID: 1, Name: "Alice"}
fmt.Println(a == b)  // true

// Structs with slices/maps are NOT comparable with ==
// Use reflect.DeepEqual or compare manually
```

### 10. Anonymous structs (inline, one-off)

```go
point := struct {
    X, Y int
}{X: 10, Y: 20}

// Common in tests and JSON decoding
var resp struct {
    Status string `json:"status"`
    Count  int    `json:"count"`
}
json.Unmarshal(data, &resp)
```
