# Go Type Assertions and Type Switches

## Quick Reference

| Use case        | Method                     |
| --------------- | -------------------------- |
| Assert type     | `val := i.(Type)`          |
| Safe assert     | `val, ok := i.(Type)`      |
| Type switch     | `switch v := i.(type)`     |
| Check interface | `val, ok := i.(Interface)` |

## Type Assertions

### 1. Basic assertion (panics if wrong type)

```go
var i interface{} = "hello"

s := i.(string)
fmt.Println(s)  // "hello"

// n := i.(int)  // panic: interface conversion
```

### 2. Safe assertion with ok

```go
var i interface{} = "hello"

s, ok := i.(string)
fmt.Println(s, ok)  // "hello" true

n, ok := i.(int)
fmt.Println(n, ok)  // 0 false
```

## Type Switches

### 3. Switch on type

```go
func describe(i interface{}) string {
    switch v := i.(type) {
    case string:
        return fmt.Sprintf("string of length %d", len(v))
    case int:
        return fmt.Sprintf("int with value %d", v)
    case bool:
        return fmt.Sprintf("bool: %t", v)
    case nil:
        return "nil"
    default:
        return fmt.Sprintf("unknown: %T", v)
    }
}
```

### 4. Multiple types in one case

```go
switch v := i.(type) {
case int, int64, float64:
    fmt.Printf("number: %v\n", v)
case string, []byte:
    fmt.Printf("text-like: %v\n", v)
}
```

## Interface Assertions

### 5. Check if value implements an interface

```go
type Stringer interface {
    String() string
}

if s, ok := val.(Stringer); ok {
    fmt.Println(s.String())
}
```

### 6. Assert to error interface

```go
if err, ok := recovered.(error); ok {
    fmt.Println(err.Error())
} else {
    fmt.Println("not an error:", recovered)
}
```

## Patterns

### 7. Handle multiple return types from any

```go
func process(data any) error {
    switch v := data.(type) {
    case []byte:
        return handleBytes(v)
    case string:
        return handleBytes([]byte(v))
    case io.Reader:
        b, err := io.ReadAll(v)
        if err != nil {
            return err
        }
        return handleBytes(b)
    default:
        return fmt.Errorf("unsupported type: %T", data)
    }
}
```

### 8. Unwrap custom error types

```go
func handleError(err error) {
    switch e := err.(type) {
    case *json.SyntaxError:
        fmt.Printf("JSON syntax error at offset %d\n", e.Offset)
    case *os.PathError:
        fmt.Printf("path error: %s on %s\n", e.Err, e.Path)
    default:
        fmt.Println(err)
    }
}
```

> Prefer `errors.As` over type-switching on errors when checking wrapped error chains.
