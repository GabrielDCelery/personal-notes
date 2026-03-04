# Go Enums (iota)

> Go has no enum keyword. Use `const` blocks with `iota` instead.

## Why

- **iota resets per const block** — It's a compile-time counter starting at 0. Adding a new value in the middle renumbers everything below it — be careful with persisted data.
- **Skip zero with blank identifier** — The zero value of an int is 0. If `Pending = 0`, an uninitialized variable looks like a valid Pending status. Skip 0 with `_ = iota` so zero means "not set".
- **Custom String() method** — Without it, `fmt.Println(Active)` prints `1`. Implementing Stringer makes logs and debugging readable.
- **String-based enums** — When the value needs to be human-readable in JSON or databases without a custom marshaler. Trades type safety (anyone can cast a string) for simplicity.
- **Bitmask with 1 << iota** — Each value is a power of 2, so you can combine them with `|` and check with `&`. Classic for permissions and feature flags.

## Quick Reference

| Use case        | Method                        |
| --------------- | ----------------------------- |
| Sequential ints | `iota`                        |
| Start from 1    | `_ = iota` then values        |
| Bitmask / flags | `1 << iota`                   |
| String enum     | implement `String()` method   |
| Validate        | define a sentinel `_maxValue` |

## Basics

### 1. Simple enum with iota

```go
type Status int

const (
    Pending  Status = iota  // 0
    Active                  // 1
    Inactive                // 2
)
```

### 2. Start from 1 (skip zero value)

```go
type Role int

const (
    _     Role = iota  // skip 0
    Admin              // 1
    User               // 2
    Guest              // 3
)
```

### 3. String representation

```go
func (s Status) String() string {
    switch s {
    case Pending:
        return "pending"
    case Active:
        return "active"
    case Inactive:
        return "inactive"
    default:
        return fmt.Sprintf("Status(%d)", s)
    }
}

fmt.Println(Active)  // "active"
```

## Patterns

### 4. Bitmask / flags

```go
type Permission int

const (
    Read    Permission = 1 << iota  // 1
    Write                           // 2
    Execute                         // 4
)

perms := Read | Write
fmt.Println(perms&Read != 0)   // true — has read
fmt.Println(perms&Execute != 0) // false — no execute
```

### 5. String-based enum

```go
type Color string

const (
    Red   Color = "red"
    Green Color = "green"
    Blue  Color = "blue"
)
```

### 6. Validation with sentinel

```go
type Direction int

const (
    North Direction = iota
    South
    East
    West
    _directionEnd  // unexported sentinel
)

func (d Direction) Valid() bool {
    return d >= North && d < _directionEnd
}
```

### 7. JSON marshal/unmarshal

```go
type Status int

const (
    Pending Status = iota
    Active
    Inactive
)

var statusNames = map[Status]string{
    Pending:  "pending",
    Active:   "active",
    Inactive: "inactive",
}

var statusValues = map[string]Status{
    "pending":  Pending,
    "active":   Active,
    "inactive": Inactive,
}

func (s Status) MarshalJSON() ([]byte, error) {
    return json.Marshal(statusNames[s])
}

func (s *Status) UnmarshalJSON(data []byte) error {
    var name string
    if err := json.Unmarshal(data, &name); err != nil {
        return err
    }
    val, ok := statusValues[name]
    if !ok {
        return fmt.Errorf("unknown status: %s", name)
    }
    *s = val
    return nil
}
```
