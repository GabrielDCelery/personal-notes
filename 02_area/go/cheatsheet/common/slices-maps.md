# Go Slices & Maps

## Why

- **Slices are references to arrays** — Slicing (`s[1:3]`) shares the underlying array. Modifying the slice modifies the original. Copy explicitly when you need independence.
- **append may reallocate** — When capacity is exceeded, append allocates a new, larger array and copies. Always use `s = append(s, ...)` — the returned slice may point to new memory.
- **Pre-allocate with make** — `make([]T, 0, n)` avoids repeated reallocations when you know the approximate size. Significant performance difference in hot loops.
- **nil vs empty slice** — Both have len 0 and work with range/append. But `json.Marshal(nil)` produces `null`, while `json.Marshal([]int{})` produces `[]`. This matters for APIs.
- **nil map panics on write** — `var m map[K]V` is nil. Reads return zero values safely, but writes panic. Always initialize with `make` or a literal before writing.
- **Map iteration order is random** — Go randomizes map iteration on purpose. Never rely on key order.

## Quick Reference

| Use case         | Method                              |
| ---------------- | ----------------------------------- |
| Declare slice    | `var s []T` / `make([]T, len, cap)` |
| Append           | `s = append(s, val)`                |
| Copy slice       | `copy(dst, src)`                    |
| Delete element   | `append(s[:i], s[i+1:]...)`         |
| Declare map      | `make(map[K]V)`                     |
| Check key exists | `v, ok := m[key]`                   |
| Delete key       | `delete(m, key)`                    |
| Iterate map      | `for k, v := range m`               |

## Slices

### 1. Declaration and initialization

```go
var s []int              // nil slice (len=0, cap=0)
s := []int{}             // empty slice (not nil)
s := []int{1, 2, 3}     // with values
s := make([]int, 3)      // len=3, cap=3, zero-valued
s := make([]int, 0, 10)  // len=0, cap=10 (pre-allocated)
```

### 2. Append

```go
s = append(s, 1)
s = append(s, 1, 2, 3)
s = append(s, other...)  // spread another slice
```

### 3. Slicing (shares underlying array)

```go
s := []int{1, 2, 3, 4, 5}
a := s[1:3]   // [2, 3] — shares memory with s
b := s[:2]    // [1, 2]
c := s[2:]    // [3, 4, 5]

// To avoid sharing memory, copy:
d := append([]int{}, s[1:3]...)
```

### 4. Copy

```go
src := []int{1, 2, 3}
dst := make([]int, len(src))
copy(dst, src) // dst is independent
```

### 5. Delete element (order preserved)

```go
s := []int{1, 2, 3, 4}
i := 2
s = append(s[:i], s[i+1:]...)
// [1, 2, 4]
```

### 6. Delete element (order not preserved, faster)

```go
s[i] = s[len(s)-1]
s = s[:len(s)-1]
```

### 7. Contains check

```go
// Go 1.21+
import "slices"

slices.Contains([]int{1, 2, 3}, 2) // true
```

### 8. nil vs empty slice (common gotcha)

```go
var s []int      // nil  — json: null
s := []int{}     // empty — json: []

// Both have len=0 and work with append/range
// But nil != empty for json.Marshal and reflect.DeepEqual
```

## Maps

### 9. Declaration and initialization

```go
var m map[string]int         // nil map — reads ok, writes panic
m := map[string]int{}        // empty map
m := make(map[string]int)    // empty map
m := map[string]int{"a": 1}  // with values
```

### 10. Read, write, delete

```go
m["key"] = 42       // set
v := m["key"]       // get (zero value if missing)
delete(m, "key")    // delete
```

### 11. Check if key exists

```go
v, ok := m["key"]
if !ok {
    // key does not exist
}
```

### 12. Iterate (order not guaranteed)

```go
for k, v := range m {
    fmt.Println(k, v)
}
```

### 13. Map of slices (common pattern)

```go
groups := make(map[string][]string)

groups["admin"] = append(groups["admin"], "alice")
groups["admin"] = append(groups["admin"], "bob")
groups["user"] = append(groups["user"], "carol")
```

### 14. Struct as map value (update gotcha)

```go
type Point struct{ X, Y int }

// WRONG — cannot assign to field of map value
m := map[string]Point{"a": {1, 2}}
m["a"].X = 10 // compile error

// RIGHT — reassign the whole value
p := m["a"]
p.X = 10
m["a"] = p

// OR use pointer values
m := map[string]*Point{"a": {1, 2}}
m["a"].X = 10 // works fine
```
