# Go Generics

## Quick Reference

| Use case             | Syntax                                       |
| -------------------- | -------------------------------------------- | --------- |
| Generic function     | `func Foo[T any](v T) T`                     |
| Generic struct       | `type Box[T any] struct`                     |
| Type constraint      | `[T int                                      | float64]` |
| Interface constraint | `[T fmt.Stringer]`                           |
| Built-in constraints | `comparable`, `any`                          |
| constraints package  | `constraints.Ordered`, `constraints.Integer` |
| Instantiate          | inferred or explicit `Foo[int](...)`         |

## Functions

### 1. Basic generic function

```go
func Map[T, U any](s []T, f func(T) U) []U {
    result := make([]U, len(s))
    for i, v := range s {
        result[i] = f(v)
    }
    return result
}

// Usage — type inferred
doubled := Map([]int{1, 2, 3}, func(n int) int { return n * 2 })
names := Map([]User{...}, func(u User) string { return u.Name })
```

### 2. Filter

```go
func Filter[T any](s []T, f func(T) bool) []T {
    var result []T
    for _, v := range s {
        if f(v) {
            result = append(result, v)
        }
    }
    return result
}

evens := Filter([]int{1, 2, 3, 4}, func(n int) bool { return n%2 == 0 })
```

### 3. Reduce

```go
func Reduce[T, U any](s []T, init U, f func(U, T) U) U {
    result := init
    for _, v := range s {
        result = f(result, v)
    }
    return result
}

sum := Reduce([]int{1, 2, 3}, 0, func(acc, n int) int { return acc + n })
```

## Type Constraints

### 4. Union constraint

```go
type Number interface {
    int | int32 | int64 | float32 | float64
}

func Sum[T Number](s []T) T {
    var total T
    for _, v := range s {
        total += v
    }
    return total
}

Sum([]int{1, 2, 3})       // 6
Sum([]float64{1.1, 2.2})  // 3.3
```

### 5. Underlying type constraint (tilde)

```go
type Celsius float64
type Fahrenheit float64

// ~float64 matches any type whose underlying type is float64
type Float interface {
    ~float32 | ~float64
}

func Double[T Float](v T) T {
    return v * 2
}

Double(Celsius(100)) // works — underlying type is float64
```

### 6. comparable (for maps/equality)

```go
func Contains[T comparable](s []T, val T) bool {
    for _, v := range s {
        if v == val {
            return true
        }
    }
    return false
}

Contains([]string{"a", "b"}, "b") // true
Contains([]int{1, 2, 3}, 4)       // false
```

### 7. Interface as constraint

```go
func Stringify[T fmt.Stringer](s []T) []string {
    result := make([]string, len(s))
    for i, v := range s {
        result[i] = v.String()
    }
    return result
}
```

## Generic Types

### 8. Generic struct

```go
type Stack[T any] struct {
    items []T
}

func (s *Stack[T]) Push(v T) {
    s.items = append(s.items, v)
}

func (s *Stack[T]) Pop() (T, bool) {
    var zero T
    if len(s.items) == 0 {
        return zero, false
    }
    v := s.items[len(s.items)-1]
    s.items = s.items[:len(s.items)-1]
    return v, true
}

// Usage
s := Stack[int]{}
s.Push(1)
s.Push(2)
v, ok := s.Pop() // 2, true
```

### 9. Generic result type

```go
type Result[T any] struct {
    Value T
    Err   error
}

func (r Result[T]) Unwrap() (T, error) {
    return r.Value, r.Err
}

func fetchUser(id int) Result[User] {
    u, err := db.Get(id)
    return Result[User]{Value: u, Err: err}
}
```

## Useful Utilities

### 10. Must (panic on error)

```go
func Must[T any](v T, err error) T {
    if err != nil {
        panic(err)
    }
    return v
}

// Usage
data := Must(os.ReadFile("config.json"))
```

### 11. Keys and Values from map

```go
func Keys[K comparable, V any](m map[K]V) []K {
    keys := make([]K, 0, len(m))
    for k := range m {
        keys = append(keys, k)
    }
    return keys
}

func Values[K comparable, V any](m map[K]V) []V {
    vals := make([]V, 0, len(m))
    for _, v := range m {
        vals = append(vals, v)
    }
    return vals
}
```

### 12. Pointer helper

```go
// Useful when you need a pointer to a literal value
func Ptr[T any](v T) *T {
    return &v
}

name := Ptr("Alice")  // *string
count := Ptr(42)      // *int
```
