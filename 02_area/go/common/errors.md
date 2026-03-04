# Go Error Handling

## Quick Reference

| Use case             | Method                       |
| -------------------- | ---------------------------- |
| Return an error      | `errors.New` / `fmt.Errorf`  |
| Wrap with context    | `fmt.Errorf("...: %w", err)` |
| Check error type     | `errors.As`                  |
| Check sentinel error | `errors.Is`                  |
| Custom error type    | implement `error` interface  |
| Unwrap chain         | `errors.Unwrap`              |

## Basics

### 1. Return and check errors

```go
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}

result, err := divide(10, 0)
if err != nil {
    log.Fatal(err)
}
```

### 2. Wrap errors with context

```go
func getUser(id int) (*User, error) {
    user, err := db.Find(id)
    if err != nil {
        return nil, fmt.Errorf("getUser %d: %w", id, err)
    }
    return user, nil
}
```

## Sentinel Errors

### 3. Define and check sentinel errors

```go
var ErrNotFound = errors.New("not found")

func findItem(id int) (*Item, error) {
    if id < 0 {
        return nil, ErrNotFound
    }
    // ...
}

item, err := findItem(-1)
if errors.Is(err, ErrNotFound) {
    // handle not found
}
```

### 4. errors.Is works through wrap chains

```go
wrapped := fmt.Errorf("layer: %w", ErrNotFound)
fmt.Println(errors.Is(wrapped, ErrNotFound)) // true
```

## Custom Error Types

### 5. Custom error with fields

```go
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error: %s - %s", e.Field, e.Message)
}

func validate(name string) error {
    if name == "" {
        return &ValidationError{Field: "name", Message: "required"}
    }
    return nil
}
```

### 6. errors.As to extract custom type

```go
err := validate("")
var ve *ValidationError
if errors.As(err, &ve) {
    fmt.Println(ve.Field)   // "name"
    fmt.Println(ve.Message) // "required"
}
```

### 7. errors.As works through wrap chains

```go
wrapped := fmt.Errorf("handler: %w", &ValidationError{Field: "name", Message: "required"})

var ve *ValidationError
if errors.As(wrapped, &ve) {
    fmt.Println(ve.Field) // "name"
}
```

## Patterns

### 8. Multiple return errors (avoid repetition)

```go
func setup() error {
    if err := connectDB(); err != nil {
        return fmt.Errorf("setup: connectDB: %w", err)
    }
    if err := migrateDB(); err != nil {
        return fmt.Errorf("setup: migrateDB: %w", err)
    }
    return nil
}
```

### 9. errors.Is vs errors.As

```go
// errors.Is  — check if a specific error value is in the chain
errors.Is(err, ErrNotFound)     // true/false

// errors.As  — check if a specific error type is in the chain and extract it
var ve *ValidationError
errors.As(err, &ve)             // true/false, populates ve if true
```
