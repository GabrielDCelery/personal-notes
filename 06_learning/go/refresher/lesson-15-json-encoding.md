# Lesson 15: JSON Encoding Deep Dive

`encoding/json` is one of the most-used packages in Go, and also one of the most misused. The default behaviour is sensible for simple cases, but the edge cases — zero-value omission, interface marshaling, large integer precision, custom encoders — trip up even experienced developers. Interviewers probe this because incorrect JSON handling often leads to silent data corruption that only surfaces in production.

## Struct Tags

Struct tags control how `encoding/json` maps Go fields to JSON keys.

```go
type User struct {
    ID        int64     `json:"id"`                   // rename: "ID" → "id"
    Name      string    `json:"name,omitempty"`        // omit if empty string
    Password  string    `json:"-"`                     // always omit
    Dash      string    `json:"-,"`                    // field named "-" (trailing comma)
    CreatedAt time.Time `json:"created_at"`
    internal  string                                   // unexported: always omitted
}
```

### Decoding Is Case-Insensitive

When unmarshaling, `encoding/json` matches JSON keys to struct fields case-insensitively, with the tagged name taking precedence:

```go
type User struct {
    Name string `json:"name"`
}

// All of these unmarshal into Name:
// {"name": "alice"}
// {"Name": "alice"}
// {"NAME": "alice"}
```

### Embedded Structs

Fields of embedded structs are promoted to the top level:

```go
type Timestamps struct {
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type User struct {
    ID   int64  `json:"id"`
    Name string `json:"name"`
    Timestamps           // ✓ fields promoted: JSON has created_at, updated_at at top level
}

// {"id": 1, "name": "alice", "created_at": "...", "updated_at": "..."}
```

**Gotcha**: If the outer struct and embedded struct both have a field with the same JSON name, the outer struct's field wins (shadowing). If two embedded structs have the same name at the same depth, both are silently omitted.

---

## `omitempty` Gotchas

`omitempty` omits the field if the value is the **zero value** for its type:

| Type                   | Omitted when                                   |
| ---------------------- | ---------------------------------------------- |
| `string`               | `""`                                           |
| `int`, `float64`, etc. | `0`                                            |
| `bool`                 | `false`                                        |
| pointer                | `nil`                                          |
| slice                  | `nil` (not empty `[]T{}`)                      |
| map                    | `nil` (not empty `map[K]V{}`)                  |
| struct                 | **never omitted** — zero structs are not empty |

```go
type Config struct {
    Timeout  int        `json:"timeout,omitempty"`   // omitted when 0
    Debug    bool       `json:"debug,omitempty"`      // omitted when false
    Tags     []string   `json:"tags,omitempty"`       // omitted when nil, NOT when []
    Metadata *Meta      `json:"metadata,omitempty"`   // omitted when nil pointer
    Options  Options    `json:"options,omitempty"`    // ❌ NEVER omitted — struct
}
```

### Omitting a Zero-Value Struct

Use a pointer:

```go
type Config struct {
    Options *Options `json:"options,omitempty"`   // ✓ omitted when nil
}

cfg := Config{}                    // Options is nil → field omitted
cfg := Config{Options: &Options{}} // Options is non-nil → field included (even if zero)
```

### `time.Time` Zero Value Is Not Omitted

```go
type Event struct {
    Name      string    `json:"name"`
    DeletedAt time.Time `json:"deleted_at,omitempty"`   // ❌ NOT omitted when zero
}

// A zero time.Time marshals to "0001-01-01T00:00:00Z" — it is a struct, not nil
// ✓ Use a pointer:
type Event struct {
    Name      string     `json:"name"`
    DeletedAt *time.Time `json:"deleted_at,omitempty"`   // omitted when nil
}
```

---

## Custom Marshaling

Implement `json.Marshaler` and `json.Unmarshaler` to control serialization completely.

```go
type Marshaler interface {
    MarshalJSON() ([]byte, error)
}

type Unmarshaler interface {
    UnmarshalJSON([]byte) error
}
```

### The Infinite Recursion Trap

```go
type Money struct {
    Amount   int64
    Currency string
}

// ❌ Infinite recursion — MarshalJSON calls json.Marshal(m) which calls MarshalJSON again
func (m Money) MarshalJSON() ([]byte, error) {
    return json.Marshal(m)   // stack overflow
}
```

**Fix: use a type alias** to break the recursion. A type alias has no methods, so `json.Marshal` uses default struct marshaling:

```go
// ✓ Type alias has no methods — no recursion
func (m Money) MarshalJSON() ([]byte, error) {
    type plain Money   // alias: same fields, no methods
    return json.Marshal(struct {
        plain
        Amount string `json:"amount"`   // override Amount field with formatted string
    }{
        plain:  plain(m),
        Amount: fmt.Sprintf("%d.%02d", m.Amount/100, m.Amount%100),
    })
}
```

### Full Custom Marshaling Example

```go
type Status int

const (
    StatusActive Status = iota
    StatusInactive
    StatusDeleted
)

func (s Status) MarshalJSON() ([]byte, error) {
    switch s {
    case StatusActive:
        return []byte(`"active"`), nil
    case StatusInactive:
        return []byte(`"inactive"`), nil
    case StatusDeleted:
        return []byte(`"deleted"`), nil
    default:
        return nil, fmt.Errorf("unknown status: %d", s)
    }
}

func (s *Status) UnmarshalJSON(data []byte) error {
    var str string
    if err := json.Unmarshal(data, &str); err != nil {
        return err
    }
    switch str {
    case "active":
        *s = StatusActive
    case "inactive":
        *s = StatusInactive
    case "deleted":
        *s = StatusDeleted
    default:
        return fmt.Errorf("unknown status: %q", str)
    }
    return nil
}
```

**Note**: `UnmarshalJSON` must have a pointer receiver — it modifies the value.

---

## Marshaling Interfaces and `any`

### Marshaling an Interface Value

`json.Marshal` encodes the **dynamic** (concrete) type of an interface value:

```go
type Animal interface{ Sound() string }
type Dog struct{ Name string }
type Cat struct{ Name string }

func (d Dog) Sound() string { return "woof" }
func (c Cat) Sound() string { return "meow" }

var a Animal = Dog{Name: "Rex"}
json.Marshal(a)   // {"Name":"Rex"} — marshals the Dog, not the Animal interface
```

### Unmarshaling Into `interface{}`

When you unmarshal into `interface{}` (or `any`), Go uses these default Go types:

| JSON type      | Go type                                   |
| -------------- | ----------------------------------------- |
| `true`/`false` | `bool`                                    |
| number         | `float64` — **always**, even for integers |
| string         | `string`                                  |
| array          | `[]interface{}`                           |
| object         | `map[string]interface{}`                  |
| null           | `nil`                                     |

```go
var result interface{}
json.Unmarshal([]byte(`{"id": 123, "name": "alice"}`), &result)

m := result.(map[string]interface{})
id := m["id"].(float64)   // ✓ it's float64, not int — even though JSON has no decimal
// id == 123.0

// ❌ This panics:
id := m["id"].(int)
```

### Large Integer Precision Loss

`float64` has 53 bits of mantissa. Integers larger than 2^53 lose precision when decoded as `float64`:

```go
data := []byte(`{"id": 9007199254740993}`)   // 2^53 + 1 — cannot be represented in float64

var result map[string]interface{}
json.Unmarshal(data, &result)
fmt.Println(result["id"])   // 9.007199254740992e+15 — WRONG, last digit lost
```

---

## `json.Number`

Use `json.Decoder.UseNumber()` to decode numbers as `json.Number` (a string alias) instead of `float64`:

```go
dec := json.NewDecoder(strings.NewReader(`{"id": 9007199254740993}`))
dec.UseNumber()   // ✓ numbers decoded as json.Number, not float64

var result map[string]interface{}
dec.Decode(&result)

n := result["id"].(json.Number)
id, _ := n.Int64()    // ✓ 9007199254740993 — exact
f, _  := n.Float64()  // also available
s     := n.String()   // "9007199254740993" — the raw string
```

**Alternatively**, use a typed struct field:

```go
type Response struct {
    ID int64 `json:"id"`   // ✓ decoded directly as int64 — no precision issue
}
```

`json.Number` is mainly useful when you're decoding into `interface{}` and need to preserve number precision.

---

## Unknown Fields and Partial Decoding

### `DisallowUnknownFields`

By default, unknown JSON keys are silently ignored. To reject them:

```go
dec := json.NewDecoder(r)
dec.DisallowUnknownFields()

var user User
if err := dec.Decode(&user); err != nil {
    // returns error if JSON contains keys not in User struct
}
```

Useful for strict API input validation. Not appropriate for forward-compatible protocols where you expect to add fields.

### `json.RawMessage` for Deferred Decoding

`json.RawMessage` is `[]byte` — it captures a JSON value verbatim, deferring actual decoding:

```go
type Event struct {
    Type    string          `json:"type"`
    Payload json.RawMessage `json:"payload"`   // decoded later based on Type
}

var event Event
json.Unmarshal(data, &event)

switch event.Type {
case "user_created":
    var payload UserCreatedPayload
    json.Unmarshal(event.Payload, &payload)
case "order_placed":
    var payload OrderPayload
    json.Unmarshal(event.Payload, &payload)
}
```

**Also useful for re-encoding**: marshaling a struct that contains pre-encoded JSON without re-encoding it.

---

## Encoder Options

```go
enc := json.NewEncoder(w)

enc.SetIndent("", "  ")        // pretty-print with 2-space indent
enc.SetIndent("", "\t")        // pretty-print with tabs
enc.SetEscapeHTML(false)       // ✓ don't escape <, >, & — important for non-HTML output
```

**`SetEscapeHTML(false)`**: By default, `json.Marshal` and `json.NewEncoder` escape `<`, `>`, and `&` to `\u003c`, `\u003e`, `\u0026`. This is safe for embedding JSON in HTML but produces ugly output in APIs and logs. Disable it when the output is not HTML-embedded:

```go
// ❌ Default escaping in API response
json.NewEncoder(w).Encode(data)
// {"url":"https://example.com\u0026ref=1"}

// ✓ Disable HTML escaping
enc := json.NewEncoder(w)
enc.SetEscapeHTML(false)
enc.Encode(data)
// {"url":"https://example.com&ref=1"}
```

---

## Performance

`encoding/json` uses reflection on every call. For hot paths, this matters.

### Avoid Repeated `json.Marshal` in Loops

```go
// ❌ Reflects on MyStruct type on every iteration
for _, item := range items {
    data, _ := json.Marshal(item)
    w.Write(data)
}

// ✓ Use an encoder — reuses internal state
enc := json.NewEncoder(w)
enc.SetEscapeHTML(false)
for _, item := range items {
    enc.Encode(item)
}
```

### Pre-allocate Buffers with `sync.Pool`

```go
var bufPool = sync.Pool{
    New: func() interface{} { return new(bytes.Buffer) },
}

func marshal(v interface{}) ([]byte, error) {
    buf := bufPool.Get().(*bytes.Buffer)
    buf.Reset()
    defer bufPool.Put(buf)

    enc := json.NewEncoder(buf)
    enc.SetEscapeHTML(false)
    if err := enc.Encode(v); err != nil {
        return nil, err
    }
    return buf.Bytes(), nil
}
```

### Code-Generated Encoders

For maximum performance, skip reflection entirely:

- **[easyjson](https://github.com/mailru/easyjson)**: generates type-specific marshal/unmarshal code at compile time
- **[sonic](https://github.com/bytedance/sonic)**: JIT-based JSON library, drop-in replacement for `encoding/json`
- **[jsoniter](https://github.com/json-iterator/go)**: reflection-based but faster, mostly API-compatible

Typical speedup: 2–5x for easyjson/jsoniter, up to 10x for sonic on large payloads. Only worth it when profiling shows JSON is a bottleneck.

---

## Hands-On Exercise 1: Custom Marshaler Without Recursion

Implement `MarshalJSON` for a `Duration` type that marshals as a human-readable string (`"1h30m"`) rather than nanoseconds, and `UnmarshalJSON` that parses it back. Avoid the infinite recursion trap.

```go
type Duration struct {
    time.Duration
}
```

<details>
<summary>Solution</summary>

```go
type Duration struct {
    time.Duration
}

func (d Duration) MarshalJSON() ([]byte, error) {
    return json.Marshal(d.Duration.String())   // ✓ marshal a string, not Duration itself
}

func (d *Duration) UnmarshalJSON(data []byte) error {
    var s string
    if err := json.Unmarshal(data, &s); err != nil {
        return err
    }
    dur, err := time.ParseDuration(s)
    if err != nil {
        return fmt.Errorf("invalid duration %q: %w", s, err)
    }
    d.Duration = dur
    return nil
}

// Usage
type Config struct {
    Timeout Duration `json:"timeout"`
}

cfg := Config{Timeout: Duration{5 * time.Minute}}
data, _ := json.Marshal(cfg)
// {"timeout":"5m0s"}

var cfg2 Config
json.Unmarshal(data, &cfg2)
// cfg2.Timeout.Duration == 5 * time.Minute
```

**Why no recursion**: `MarshalJSON` calls `json.Marshal(d.Duration.String())` — it marshals a `string`, not a `Duration`. `time.Duration` has no `MarshalJSON` method, so there's no recursive call.

</details>

## Hands-On Exercise 2: Heterogeneous JSON Array with `RawMessage`

Decode this JSON where each element has a different shape depending on `"type"`:

```json
[
  { "type": "text", "content": "hello" },
  { "type": "image", "url": "https://example.com/img.png", "width": 800 }
]
```

Define Go types and unmarshal the array into a `[]interface{}` where each element is the correct concrete type.

<details>
<summary>Solution</summary>

```go
type TextBlock struct {
    Content string `json:"content"`
}

type ImageBlock struct {
    URL   string `json:"url"`
    Width int    `json:"width"`
}

type RawBlock struct {
    Type    string          `json:"type"`
    Payload json.RawMessage `json:"-"`   // capture everything
}

func unmarshalBlocks(data []byte) ([]interface{}, error) {
    // First pass: decode to get raw messages
    var raws []json.RawMessage
    if err := json.Unmarshal(data, &raws); err != nil {
        return nil, err
    }

    results := make([]interface{}, 0, len(raws))

    for _, raw := range raws {
        // Decode just the type field
        var typed struct {
            Type string `json:"type"`
        }
        if err := json.Unmarshal(raw, &typed); err != nil {
            return nil, err
        }

        // Decode the full object based on type
        switch typed.Type {
        case "text":
            var b TextBlock
            if err := json.Unmarshal(raw, &b); err != nil {
                return nil, err
            }
            results = append(results, b)
        case "image":
            var b ImageBlock
            if err := json.Unmarshal(raw, &b); err != nil {
                return nil, err
            }
            results = append(results, b)
        default:
            return nil, fmt.Errorf("unknown block type: %q", typed.Type)
        }
    }

    return results, nil
}
```

**Key technique**: Decode the array as `[]json.RawMessage` first — each element is kept as raw bytes. Then decode each element twice: once for the discriminator field (`type`), once for the full struct. This avoids decoding into `map[string]interface{}` and losing type safety.

</details>

---

## Interview Questions

### Q1: What does `omitempty` do, and what is its biggest gotcha?

Asked frequently because `omitempty` misuse causes incorrect API responses — fields appear when they shouldn't, or don't appear when they should. It tests attention to Go's type system.

<details>
<summary>Answer</summary>

`omitempty` omits a field from the JSON output if its value is the zero value for that type: `""` for strings, `0` for numbers, `false` for bools, `nil` for pointers, slices, and maps.

**The biggest gotcha**: structs are never considered empty by `omitempty` — even a zero-value struct is included:

```go
type Meta struct{ Tags []string }

type Response struct {
    Meta Meta `json:"meta,omitempty"`   // ❌ NEVER omitted — struct is never "empty"
}

// Always marshals as {"meta":{"tags":null}}

// ✓ Fix: use a pointer
type Response struct {
    Meta *Meta `json:"meta,omitempty"`   // omitted when nil
}
```

**Second gotcha**: `time.Time` is a struct — `omitempty` doesn't omit zero times. Use `*time.Time`.

**Third gotcha**: an empty (non-nil) slice `[]string{}` is not omitted by `omitempty` — only a nil slice is. This matters when you explicitly initialize a slice to empty:

```go
type Response struct {
    Tags []string `json:"tags,omitempty"`
}

r := Response{Tags: []string{}}    // NOT omitted — slice is non-nil, just empty
r := Response{}                     // omitted — Tags is nil
```

</details>

### Q2: How does `encoding/json` handle numbers when unmarshaling into `interface{}`?

A precision question — tests whether the candidate has hit the `float64` integer precision problem in production or has read the docs carefully enough to know about it.

<details>
<summary>Answer</summary>

All JSON numbers are decoded as `float64` when the target type is `interface{}`. This is because JSON has a single `number` type with no distinction between integers and floats.

```go
var result map[string]interface{}
json.Unmarshal([]byte(`{"id": 42}`), &result)

id := result["id"]   // float64(42), not int(42)
```

**The precision problem**: `float64` has 53 bits of mantissa. Integers larger than 2^53 (9,007,199,254,740,992) cannot be represented exactly in `float64`. Twitter famously had to switch tweet IDs to strings in their JSON API because JavaScript (which also uses float64) was corrupting large integer IDs.

**Solutions**:

1. **Typed struct field**: `ID int64 \`json:"id"\``— decoded directly as`int64`, no precision loss
2. **`json.Decoder.UseNumber()`**: numbers decoded as `json.Number` (a string alias), then convert with `.Int64()` or `.Float64()` as needed
3. **String encoding**: encode large integers as JSON strings — universally safe but changes the API contract

For APIs dealing with large IDs (database primary keys, snowflake IDs), always use typed struct fields or `UseNumber()`.

</details>

### Q3: Why does calling `json.Marshal(s)` inside `MarshalJSON` cause infinite recursion, and how do you fix it?

A gotcha question — directly tests whether the candidate has implemented custom marshalers before.

<details>
<summary>Answer</summary>

When `json.Marshal(v)` is called and `v` implements `json.Marshaler`, `encoding/json` calls `v.MarshalJSON()`. If `MarshalJSON` calls `json.Marshal(s)` where `s` is the same type (or the same value), the cycle repeats indefinitely, causing a stack overflow.

```go
func (m Money) MarshalJSON() ([]byte, error) {
    return json.Marshal(m)   // ❌ m implements Marshaler → calls MarshalJSON → calls json.Marshal(m) → ...
}
```

**Fix: type alias**. A type alias defined inside the function has the same fields but inherits no methods from the original type:

```go
func (m Money) MarshalJSON() ([]byte, error) {
    type plain Money          // same fields as Money, but no MarshalJSON method
    return json.Marshal(plain(m))   // ✓ plain doesn't implement Marshaler — uses default struct encoding
}
```

This is the canonical Go pattern. It's also useful when you want to customize only part of the marshaling — define the alias, wrap it in an anonymous struct, override specific fields:

```go
func (m Money) MarshalJSON() ([]byte, error) {
    type plain Money
    return json.Marshal(struct {
        plain
        Amount string `json:"amount"`   // shadow the original Amount field with formatted string
    }{
        plain:  plain(m),
        Amount: fmt.Sprintf("%.2f", float64(m.Amount)/100),
    })
}
```

</details>

### Q4: What does `enc.SetEscapeHTML(false)` do and when should you use it?

A practical question about JSON output correctness — shows awareness of encoding/json's default behaviour and its trade-offs.

<details>
<summary>Answer</summary>

By default, `json.Marshal` and `json.NewEncoder` escape three characters to their Unicode escape sequences:

| Character | Escaped as |
| --------- | ---------- |
| `<`       | `\u003c`   |
| `>`       | `\u003e`   |
| `&`       | `\u0026`   |

This is done to make JSON safe for direct embedding in HTML `<script>` tags, preventing XSS attacks if the JSON is rendered in a browser context.

```go
data := map[string]string{"url": "https://example.com?a=1&b=2"}
b, _ := json.Marshal(data)
// {"url":"https://example.com?a=1\u0026b=2"}  — & is escaped
```

**When to disable**:

```go
enc := json.NewEncoder(w)
enc.SetEscapeHTML(false)
enc.Encode(data)
// {"url":"https://example.com?a=1&b=2"}  — ✓ readable
```

Disable HTML escaping when:

- The JSON is consumed by an API client, not rendered in HTML
- The output goes to logs where `\u0026` is ugly and confusing
- You're comparing JSON output in tests and the escaping causes mismatches

Keep HTML escaping (the default) when:

- The JSON may be embedded directly in HTML templates
- You're unsure of the context — safety over aesthetics

Note: `json.Marshal` has no option to disable escaping — you must use `json.NewEncoder` with `SetEscapeHTML(false)`.

</details>

---

## Key Takeaways

1. **`omitempty` never omits structs**: use `*MyStruct` (pointer) if you want a struct field omitted when zero. Same applies to `time.Time`.
2. **Nil vs empty slice with `omitempty`**: `nil` is omitted, `[]T{}` is not — keep this in mind when initialising slices.
3. **Type alias trick**: inside `MarshalJSON`, use `type plain MyType` to call `json.Marshal` without recursion.
4. **`interface{}` numbers are `float64`**: use typed struct fields or `dec.UseNumber()` when precision matters for large integers.
5. **`json.RawMessage`**: defer decoding until you know the concrete type — essential for discriminated unions and plugin-style JSON.
6. **`SetEscapeHTML(false)`**: disable HTML escaping for non-HTML output — `&` and `<` stay readable in API responses and logs.
7. **`DisallowUnknownFields`**: opt-in to strict parsing for input validation; not appropriate for forward-compatible protocols.
8. **`UnmarshalJSON` needs pointer receiver**: modifies the value in place — value receiver silently does nothing.
9. **Embedded struct fields are promoted**: useful for reusing common field groups, but shadowing rules can cause fields to silently disappear.
10. **Reflection cost**: `encoding/json` reflects on every call; use `json.NewEncoder` for repeated encoding, `sync.Pool` for buffer reuse, and code-generated encoders for hot paths.

## Next Steps

In [Lesson 16: database/sql Patterns](lesson-16-database-sql.md), you'll learn:

- Why `sql.DB` is a pool, not a connection, and how to configure it
- The correct transaction pattern with deferred rollback
- `sql.Rows` gotchas — close, drain, and check `rows.Err()`
- `sql.Null*` types vs pointer types for nullable columns
- How context cancellation propagates to in-flight queries
