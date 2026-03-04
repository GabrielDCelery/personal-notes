# Go Pointers

## Why

- **Go is pass-by-value** — Everything is copied when passed to a function. Pointers let you avoid copying large structs and let functions modify the caller's data.
- **Slices and maps are already references** — They contain internal pointers, so you don't need to pass `*[]int`. But append can reallocate, so always return the new slice.
- **nil pointer = absence** — Returning `*User` lets you return nil for "not found" without a separate bool. But always check for nil before dereferencing.
- **Pointer for optional fields** — `*string` distinguishes "not provided" (nil) from "provided but empty" (""). Common in PATCH/update APIs.
- **new vs &** — `new(T)` allocates and returns a pointer to a zero value. `&T{...}` does the same but lets you set fields. The `&` form is idiomatic for structs.

## Quick Reference

| Use case          | Method                    |
| ----------------- | ------------------------- |
| Get pointer       | `p := &val`               |
| Dereference       | `*p`                      |
| Allocate with new | `p := new(Type)`          |
| Nil check         | `if p != nil`             |
| Pointer receiver  | `func (t *Type) Method()` |
| Value receiver    | `func (t Type) Method()`  |

## Basics

### 1. Create and dereference

```go
x := 42
p := &x          // p is *int, points to x
fmt.Println(*p)  // 42 — dereference
*p = 100         // modifies x through pointer
fmt.Println(x)   // 100
```

### 2. new vs &

```go
// Both create a pointer to a zero-valued int
p1 := new(int)       // *int, points to 0
p2 := &int{}         // won't compile — can't take address of primitive literal

// For structs, & is idiomatic
u := &User{Name: "Alice"}  // preferred
u := new(User)              // less common, zero-valued
```

### 3. Nil pointers

```go
var p *int          // nil by default
fmt.Println(p)      // <nil>

// Dereferencing nil panics
// *p = 42          // panic: runtime error

if p != nil {
    fmt.Println(*p)
}
```

## Value vs Pointer Receivers

### 4. When to use pointer receivers

```go
// Pointer receiver — use when:
// - Method modifies the struct
// - Struct is large (avoids copying)
// - Consistency (if one method needs pointer, use for all)
func (u *User) SetName(name string) {
    u.Name = name
}

// Value receiver — use when:
// - Method only reads, struct is small
// - You want the type to be usable as a map key or in comparisons
func (u User) String() string {
    return u.Name
}
```

## Passing Pointers

### 5. Functions that modify values

```go
func increment(n *int) {
    *n++
}

x := 5
increment(&x)
fmt.Println(x)  // 6
```

### 6. Slices and maps are already references

```go
// No need to pass pointers for slices and maps
func appendItem(items []string, item string) []string {
    return append(items, item)  // append may reallocate, so return the new slice
}

func setKey(m map[string]int, k string, v int) {
    m[k] = v  // modifies the original map directly
}
```

## Patterns

### 7. Optional values with pointers

```go
type UpdateRequest struct {
    Name  *string  // nil means "don't update"
    Email *string
}

func StringPtr(s string) *string { return &s }

req := UpdateRequest{
    Name: StringPtr("Alice"),
    // Email is nil — not updated
}
```

### 8. Pointer in return to signal absence

```go
func findUser(id int) *User {
    if id < 0 {
        return nil
    }
    return &User{ID: id}
}

u := findUser(1)
if u == nil {
    // not found
}
```
