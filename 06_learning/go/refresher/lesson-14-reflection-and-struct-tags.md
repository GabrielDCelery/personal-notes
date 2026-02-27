# Lesson 14: Reflection & Struct Tags

Reflection lets you inspect and manipulate types and values at runtime — without knowing their concrete types at compile time. It's how `encoding/json`, `database/sql`, ORMs, test assertion libraries, and dependency injection frameworks work. You rarely write reflection code yourself, but you're expected to understand what it's doing and when it's appropriate — and interviewers use reflection questions to probe understanding of Go's type system at a deeper level.

## Why Reflection Exists

Before generics (pre-Go 1.18), reflection was the only way to write code that operated on arbitrary types. Even with generics, reflection fills a different role: generics give you type-safe algorithms over types whose _shape_ you know at compile time; reflection handles truly dynamic behaviour — inspecting struct fields by name, reading tags, constructing values of unknown types.

```go
// Generics: works on types whose shape you know
func Map[T, U any](s []T, f func(T) U) []U { ... }   // ✓ compile-time type safety

// Reflection: inspecting unknown struct fields at runtime
func structToMap(v interface{}) map[string]interface{} {
    // You can't know at compile time what fields `v` has
    // ✓ reflection handles it
}
```

**The cost of reflection**:

- No compile-time type safety — errors surface at runtime as panics
- 10–100x slower than direct field access
- Code is verbose and hard to read
- The `reflect` package API is notoriously confusing

Use reflection only when the type is genuinely unknown at compile time. If you know the types, use direct access or generics.

---

## `reflect.Type` and `reflect.Value`

Every reflection operation starts with one of two entry points:

```go
var x float64 = 3.14

t := reflect.TypeOf(x)    // reflect.Type  — describes the type
v := reflect.ValueOf(x)   // reflect.Value — wraps the value
```

### `reflect.Type`

Describes the type itself — its kind, name, methods, fields:

```go
t := reflect.TypeOf(x)

t.Kind()        // reflect.Float64
t.Name()        // "float64"
t.String()      // "float64"
t.NumMethod()   // number of methods in the method set

// For structs:
t.NumField()            // number of fields
t.Field(0)              // reflect.StructField — first field descriptor
t.FieldByName("Name")   // reflect.StructField, bool
```

### `reflect.Value`

Wraps an actual value — lets you read and (sometimes) write it:

```go
v := reflect.ValueOf(x)

v.Type()          // same as reflect.TypeOf(x)
v.Kind()          // reflect.Float64
v.Float()         // 3.14 — extract the underlying float64
v.Interface()     // interface{} — recover the original value
v.IsValid()       // false if v is the zero reflect.Value
v.IsNil()         // true for nil pointers, channels, maps, slices, funcs
```

### The Interface Dynamic Type Gotcha

`reflect.TypeOf` returns the **dynamic** type of an interface value:

```go
var r io.Reader = os.Stdin

reflect.TypeOf(r)   // *os.File, not io.Reader
// reflect always sees through the interface to the concrete type
```

To get the `reflect.Type` of an interface itself (useful when registering types):

```go
readerType := reflect.TypeOf((*io.Reader)(nil)).Elem()
// (*io.Reader)(nil) is a *io.Reader nil pointer
// .Elem() dereferences the pointer → gives you the io.Reader interface type
```

---

## `reflect.Kind`

`Kind` is the underlying category of a type — not its named type:

```go
type Celsius float64
type UserID  int64

reflect.TypeOf(Celsius(0)).Kind()   // reflect.Float64 — not "Celsius"
reflect.TypeOf(UserID(0)).Kind()    // reflect.Int64   — not "UserID"
reflect.TypeOf("hello").Kind()      // reflect.String
reflect.TypeOf([]int{}).Kind()      // reflect.Slice
reflect.TypeOf(map[string]int{}).Kind() // reflect.Map
```

When writing generic reflection code, branch on `Kind`, not `Type`:

```go
// ✓ Handles all named types based on underlying kind
func printValue(v reflect.Value) {
    switch v.Kind() {
    case reflect.String:
        fmt.Println(v.String())
    case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
        fmt.Println(v.Int())
    case reflect.Float32, reflect.Float64:
        fmt.Println(v.Float())
    case reflect.Bool:
        fmt.Println(v.Bool())
    case reflect.Ptr:
        if v.IsNil() {
            fmt.Println("<nil>")
        } else {
            printValue(v.Elem())   // dereference and recurse
        }
    case reflect.Struct:
        // iterate fields
    }
}
```

**Gotcha**: forgetting to dereference pointers before inspecting struct fields:

```go
v := reflect.ValueOf(&user)        // Kind == Ptr
v.NumField()                       // ❌ panics — Ptr has no fields

v = reflect.ValueOf(&user).Elem()  // ✓ Kind == Struct — dereference first
v.NumField()                       // ✓
```

---

## Traversing Structs

```go
type User struct {
    ID    int64  `json:"id"    db:"id"`
    Name  string `json:"name"  db:"name"`
    email string              // unexported
}

u := User{ID: 1, Name: "alice"}
t := reflect.TypeOf(u)
v := reflect.ValueOf(u)

for i := 0; i < t.NumField(); i++ {
    field := t.Field(i)                    // reflect.StructField
    value := v.Field(i)                    // reflect.Value for this field

    fmt.Printf("Name: %s, Kind: %s, Tag: %s\n",
        field.Name,
        field.Type.Kind(),
        field.Tag,
    )
}
```

### `reflect.StructField`

| Field          | Description                                |
| -------------- | ------------------------------------------ |
| `Name`         | Field name (`"ID"`, `"Name"`)              |
| `Type`         | `reflect.Type` of the field                |
| `Tag`          | The raw struct tag string                  |
| `Index`        | `[]int` — index path (for embedded fields) |
| `Anonymous`    | `true` if the field is embedded            |
| `IsExported()` | `true` if the field is exported (Go 1.17+) |

```go
// ✓ Skip unexported fields
for i := 0; i < t.NumField(); i++ {
    field := t.Field(i)
    if !field.IsExported() {
        continue   // can't read or set unexported fields via reflection
    }
    // ...
}
```

---

## Settability

To set a field via reflection, you must pass a **pointer** to the value:

```go
u := User{Name: "alice"}

// ❌ Not settable — reflect.ValueOf(u) is a copy
v := reflect.ValueOf(u)
v.FieldByName("Name").SetString("bob")   // panics: reflect.Value.SetString using value obtained using unexported field

// ✓ Pass a pointer, dereference with Elem()
v := reflect.ValueOf(&u).Elem()
v.FieldByName("Name").SetString("bob")   // ✓ u.Name is now "bob"
```

**Why**: `reflect.ValueOf(u)` captures a copy. There's nothing to set on a copy — the original wouldn't change. Passing a pointer gives reflect a reference to the original memory.

### `CanSet()` and `CanAddr()`

Always check before setting to avoid panics:

```go
v := reflect.ValueOf(&u).Elem()

field := v.FieldByName("Name")
if !field.CanSet() {
    // unexported field, or not addressable
    return
}
field.SetString("bob")   // ✓ safe
```

### `FieldByName` vs `Field(i)` in Hot Paths

`FieldByName` does a linear search through fields by name — O(n) per call. In a hot path, prefer caching the field index:

```go
// ✓ Cache the index once
t := reflect.TypeOf(User{})
idx := -1
for i := 0; i < t.NumField(); i++ {
    if t.Field(i).Name == "Name" {
        idx = i
        break
    }
}

// Use index on each value
v := reflect.ValueOf(&u).Elem()
v.Field(idx).SetString("bob")   // O(1) field access
```

---

## Struct Tags

Struct tags are raw string literals attached to struct fields. They follow a convention but are not enforced by the compiler.

```go
type User struct {
    ID   int64  `json:"id" db:"user_id" validate:"required"`
    Name string `json:"name,omitempty" db:"name"`
}
```

### Tag Syntax

```
`key:"value" key2:"value2,option1,option2"`
```

- Keys and values are separated by `:`
- Multiple keys separated by spaces
- Options within a value are comma-separated (convention varies by package)
- The entire tag is a raw string literal (backticks)

### Reading Tags

```go
t := reflect.TypeOf(User{})
field, _ := t.FieldByName("Name")

// Get returns "" if the key is absent — can't distinguish missing from empty
tag := field.Tag.Get("json")     // "name,omitempty"

// Lookup returns ("", false) if absent — distinguishes missing from empty
val, ok := field.Tag.Lookup("json")   // "name,omitempty", true
val, ok  = field.Tag.Lookup("xml")    // "", false
```

### Conventional Tag Keys

| Key        | Used by                   |
| ---------- | ------------------------- |
| `json`     | `encoding/json`           |
| `yaml`     | `gopkg.in/yaml.v3`        |
| `db`       | `sqlx`, `gorm`            |
| `xml`      | `encoding/xml`            |
| `validate` | `go-playground/validator` |
| `env`      | various config loaders    |
| `flag`     | various flag parsers      |

### How `encoding/json` Uses Tags

When `json.Marshal` or `json.Unmarshal` encounters a struct, it:

1. Iterates fields via `reflect.Type.NumField()` / `Field(i)`
2. Reads the `json` tag with `field.Tag.Get("json")`
3. Splits on `,` — first part is the JSON key name, rest are options (`omitempty`, `string`)
4. Skips unexported fields and fields tagged `json:"-"`
5. Uses the field name as the JSON key if no tag is present (case-insensitive match on decode)

**Tag typo gotcha**: misspelled tag keys are silently ignored — no compile-time error:

```go
type User struct {
    Name string `jsno:"name"`   // ❌ typo — "jsno" not "json"; field uses its Go name "Name"
}
// Marshals as {"Name":"..."} not {"name":"..."}
```

No tool catches this by default. `go vet` checks `printf`-style format strings but not struct tag keys. `staticcheck` catches some tag issues.

---

## Reflection vs Generics

| Situation                                                        | Use        |
| ---------------------------------------------------------------- | ---------- |
| Type-safe algorithm over a known shape (e.g., filter, map, sort) | Generics   |
| Inspect struct fields by name at runtime                         | Reflection |
| Enumerate struct tags                                            | Reflection |
| Build a value of an unknown type                                 | Reflection |
| Plugin system, registry, framework internals                     | Reflection |
| Constraint to a set of known types                               | Generics   |
| Avoid boxing/unboxing overhead                                   | Generics   |

**What generics cannot do that reflection can**:

- Enumerate the fields of a struct at runtime
- Read struct tags
- Create values of a type you receive as `reflect.Type`
- Implement `encoding/json`-style unmarshaling into arbitrary types

**The hybrid pattern**: use generics for the type-safe outer API, reflection internally:

```go
// Generic outer API — type-safe for callers
func Scan[T any](rows *sql.Rows) (T, error) {
    var result T
    // Internally uses reflection to map columns to fields of T
    return result, scanInto(rows, &result)
}

func scanInto(rows *sql.Rows, dest interface{}) error {
    v := reflect.ValueOf(dest).Elem()   // reflection internally
    // ...
}
```

---

## Performance

Reflection is significantly slower than direct access. The overhead comes from:

- Type assertion at every operation
- Interface boxing/unboxing
- Cannot be inlined by the compiler

| Operation       | Direct | Via Reflection | Ratio  |
| --------------- | ------ | -------------- | ------ |
| Field read      | ~1ns   | ~10–50ns       | 10–50x |
| Field write     | ~1ns   | ~20–60ns       | 20–60x |
| `TypeOf`        | —      | ~5ns           | —      |
| Struct tag read | —      | ~15ns          | —      |

### Cache `reflect.Type` Results

`reflect.TypeOf` is cheap but not free. For types used in tight loops, cache the result:

```go
var userType = reflect.TypeOf(User{})   // ✓ computed once at package init

func process(v interface{}) {
    if reflect.TypeOf(v) != userType {
        // type mismatch
    }
}
```

### Cache Field Indices with `sync.Map`

For serialisation-like hot paths:

```go
var fieldCache sync.Map   // map[reflect.Type][]cachedField

type cachedField struct {
    index int
    name  string
    tag   string
}

func getFields(t reflect.Type) []cachedField {
    if v, ok := fieldCache.Load(t); ok {
        return v.([]cachedField)
    }
    // build and store
    fields := buildFields(t)
    fieldCache.Store(t, fields)
    return fields
}
```

This is exactly what `encoding/json` does internally — it caches the struct layout on the first call and reuses it on subsequent calls.

---

## Hands-On Exercise 1: Struct to Map

Write `StructToMap(v interface{}) (map[string]interface{}, error)` that converts a struct's exported fields to a map, using the `json` tag name (or the field name if no tag is present). Skip fields tagged `json:"-"`.

<details>
<summary>Solution</summary>

```go
func StructToMap(v interface{}) (map[string]interface{}, error) {
    rv := reflect.ValueOf(v)

    // Dereference pointer if necessary
    if rv.Kind() == reflect.Ptr {
        if rv.IsNil() {
            return nil, errors.New("nil pointer")
        }
        rv = rv.Elem()
    }

    if rv.Kind() != reflect.Struct {
        return nil, fmt.Errorf("expected struct, got %s", rv.Kind())
    }

    rt := rv.Type()
    result := make(map[string]interface{}, rt.NumField())

    for i := 0; i < rt.NumField(); i++ {
        field := rt.Field(i)

        // Skip unexported fields
        if !field.IsExported() {
            continue
        }

        // Read json tag
        key := field.Name   // default: Go field name
        if tag, ok := field.Tag.Lookup("json"); ok {
            parts := strings.Split(tag, ",")
            if parts[0] == "-" {
                continue   // skip json:"-"
            }
            if parts[0] != "" {
                key = parts[0]
            }
        }

        result[key] = rv.Field(i).Interface()
    }

    return result, nil
}

// Test
type User struct {
    ID       int64  `json:"id"`
    Name     string `json:"name"`
    Password string `json:"-"`
    internal string
}

u := User{ID: 1, Name: "alice", Password: "secret"}
m, _ := StructToMap(u)
// map[id:1 name:alice] — Password and internal are excluded
```

</details>

## Hands-On Exercise 2: Simple Deep Equal

Write `DeepEqual(a, b interface{}) bool` using reflection that compares two structs field-by-field. It should:

- Return `true` if both are the same type and all exported fields are equal
- Return `false` if types differ or any exported field differs
- Treat unexported fields as irrelevant (skip them)

Do not use `reflect.DeepEqual` — implement it manually.

<details>
<summary>Solution</summary>

```go
func DeepEqual(a, b interface{}) bool {
    va := reflect.ValueOf(a)
    vb := reflect.ValueOf(b)

    // Dereference pointers
    if va.Kind() == reflect.Ptr { va = va.Elem() }
    if vb.Kind() == reflect.Ptr { vb = vb.Elem() }

    // Must be the same kind
    if va.Kind() != vb.Kind() {
        return false
    }

    // Must be the same type
    if va.Type() != vb.Type() {
        return false
    }

    if va.Kind() != reflect.Struct {
        // For non-structs, use Interface() comparison
        return va.Interface() == vb.Interface()
    }

    t := va.Type()
    for i := 0; i < t.NumField(); i++ {
        field := t.Field(i)
        if !field.IsExported() {
            continue   // skip unexported
        }

        fa := va.Field(i).Interface()
        fb := vb.Field(i).Interface()

        if fa != fb {
            return false
        }
    }

    return true
}
```

**Limitations of this implementation**: uses `==` on field values, so it doesn't handle slices, maps, or nested structs correctly (those aren't comparable with `==`). A full implementation would recurse. This is intentional for the exercise — the standard library's `reflect.DeepEqual` handles all these cases.

</details>

---

## Interview Questions

### Q1: What is the difference between `reflect.Type` and `reflect.Value`, and when do you need each?

Interviewers ask this to see if the candidate has actually used the `reflect` package rather than just knowing it exists. Confusing the two is a sign of surface-level familiarity.

<details>
<summary>Answer</summary>

`reflect.Type` describes the **type** — its kind, name, method set, fields (for structs), element type (for slices/maps). It's what you use to inspect structure.

`reflect.Value` wraps an **actual value** — it lets you read and (if addressable) write the data. It carries both a type and a value.

```go
var x int64 = 42

t := reflect.TypeOf(x)   // Type — describes int64 as a concept
v := reflect.ValueOf(x)  // Value — wraps the specific value 42

t.Kind()      // reflect.Int64
t.Name()      // "int64"

v.Kind()      // reflect.Int64 (same)
v.Int()       // 42 — the actual value
v.Type()      // gives you the Type from a Value
```

**When you need each**:

- `reflect.Type`: when you're inspecting struct layout (fields, tags) without touching actual data — e.g., building a schema, registering type mappings, validating that a type satisfies a structure
- `reflect.Value`: when you're reading or writing actual data — e.g., marshaling, unmarshaling, dependency injection

In practice, you usually need both: `TypeOf` to understand structure, `ValueOf` to access data. You can get `Type` from a `Value` via `v.Type()`, so often `reflect.ValueOf` is the single entry point.

</details>

### Q2: Why does setting a struct field via reflection require a pointer, and how do you do it correctly?

A direct test of reflection mechanics — this trips up almost everyone who hasn't used reflection for mutation before.

<details>
<summary>Answer</summary>

`reflect.ValueOf(x)` captures a **copy** of `x`. Setting a field on a copy has no effect on the original — and reflect knows this, so it marks the value as not settable. Calling `Set` panics.

To set a field, you need a `reflect.Value` that is **addressable** — one that refers to actual memory. You get this by passing a pointer and dereferencing:

```go
type User struct{ Name string }
u := User{Name: "alice"}

// ❌ Not settable — v is a copy
v := reflect.ValueOf(u)
v.FieldByName("Name").SetString("bob")   // panic: reflect: reflect.Value.SetString using value obtained using unexported field

// ✓ Settable — pass pointer, dereference with Elem()
v := reflect.ValueOf(&u).Elem()         // Elem() dereferences the pointer
v.FieldByName("Name").SetString("bob")  // u.Name is now "bob"
```

`Elem()` on a pointer `reflect.Value` dereferences it, yielding a settable value backed by the original memory.

Always call `CanSet()` before `Set` when writing library code — unexported fields are not settable even with a pointer:

```go
field := v.FieldByName("Name")
if field.CanSet() {
    field.SetString("bob")
}
```

</details>

### Q3: When should you use reflection instead of generics?

A design question — tests whether the candidate understands where the two features overlap and where they diverge. Expected from senior developers who've worked on both sides of a library API.

<details>
<summary>Answer</summary>

**Use generics when**:

- You know the shape of the type at compile time (it's a slice, a map, has a specific method)
- You want type safety and compiler-checked constraints
- You want performance (generic code can be inlined; reflection cannot)
- Example: `slices.Sort`, `maps.Keys`, a generic `Filter` or `Map` function

**Use reflection when**:

- The type is genuinely unknown at compile time
- You need to enumerate struct fields or read struct tags
- You're building framework-level code: serialisation libraries, ORMs, DI containers, test assertion libraries
- You need to construct values of a type you received dynamically (e.g., a plugin system)
- Example: `encoding/json`, `database/sql` scanning, `reflect.DeepEqual`

**The key distinction**: generics are parametric — you write code that works for any type matching a constraint, but the constraint is fixed at compile time. Reflection is dynamic — you inspect and manipulate types whose structure is not known until runtime.

**The hybrid**: use generics for the caller-facing API (type safety, no casting), use reflection internally for the dynamic work:

```go
func Decode[T any](data []byte) (T, error) {
    var result T
    err := json.Unmarshal(data, &result)   // json.Unmarshal uses reflection internally
    return result, err
}
```

</details>

### Q4: How does `encoding/json` use struct tags, and what happens with a tag typo?

A practical question that reveals understanding of how the standard library actually works, not just what its API looks like.

<details>
<summary>Answer</summary>

When `json.Marshal` or `json.Unmarshal` encounters a struct, it uses reflection to inspect the type. For each exported field it:

1. Reads the raw tag string: `field.Tag.Get("json")` or `field.Tag.Lookup("json")`
2. Splits on `,` — `parts[0]` is the JSON key name, remaining parts are options
3. If `parts[0] == "-"` → field is always skipped
4. If `parts[0] == ""` → use the Go field name as the JSON key
5. Otherwise → use `parts[0]` as the JSON key
6. Recognises options: `omitempty` (skip zero values), `string` (encode number/bool as JSON string)

`encoding/json` caches this analysis using `sync.Map` keyed by `reflect.Type`, so the reflection cost is paid once per type, not once per call.

**Tag typo behaviour**:

```go
type User struct {
    Name string `jsno:"name"`   // typo: "jsno" instead of "json"
}

b, _ := json.Marshal(User{Name: "alice"})
// {"Name":"alice"} — uses the Go field name, not the (misspelled) tag
```

The typo is completely silent. `encoding/json` looks for the key `"json"` — since it's absent, it falls back to the field name. There is no warning, no error, no panic. The JSON output changes silently.

**How to catch it**: `staticcheck` can detect some tag issues. Custom linters or `go-structtag` validation can check tag syntax. In practice, test your JSON output in unit tests — marshal/unmarshal round-trips would catch wrong key names.

</details>

---

## Key Takeaways

1. **`reflect.TypeOf` vs `reflect.ValueOf`**: Type for structure inspection, Value for data access — you usually need both.
2. **Branch on `Kind`, not `Type`**: `Kind` gives the underlying category (Struct, Ptr, Slice); named types like `type UserID int64` have `Kind == Int64`.
3. **Always dereference pointers**: check `v.Kind() == reflect.Ptr` and call `v.Elem()` before accessing struct fields.
4. **Settability requires a pointer**: pass `&v` and call `Elem()` to get an addressable, settable `reflect.Value`. Check `CanSet()` before setting.
5. **`FieldByName` is O(n)**: for hot paths, cache the field index and use `Field(i)`.
6. **Tag typos are silent**: `json:"naem"` is not caught by the compiler or `go vet` — test your serialisation explicitly.
7. **`Get` vs `Lookup`**: `Tag.Get` returns `""` for absent keys; `Tag.Lookup` returns `"", false` — use `Lookup` to distinguish absent from empty.
8. **Reflection vs generics**: generics for type-safe algorithms with known shapes; reflection for runtime struct inspection, tags, and dynamic type construction.
9. **Cache `reflect.Type`**: store the result of `reflect.TypeOf` at package level or in a `sync.Map` — recomputing it per call adds unnecessary overhead.
10. **Reflection is 10–100x slower**: measure before adding reflection to hot paths; code-generation or generics are preferable alternatives when the type is known.

## Next Steps

In [Lesson 15: JSON Encoding Deep Dive](lesson-15-json-encoding.md), you'll learn:

- How `encoding/json` uses struct tags under the hood
- The `omitempty` gotchas that catch everyone at least once
- Custom `MarshalJSON`/`UnmarshalJSON` and the infinite recursion trap
- `json.Number` and large integer precision loss
- `json.RawMessage` for deferred and partial decoding
