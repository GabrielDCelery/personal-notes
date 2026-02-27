# Go: Header Types

Only **strings** and **interfaces** share the same "header" pattern as slices — a value type that bundles a pointer with metadata.

## String header — `{ptr, len}`

```go
type StringHeader struct {
    Data uintptr  // pointer to bytes
    Len  int      // number of bytes
}
```

Like a slice but without `cap` since strings are immutable — no need to track allocated space.

```go
s := "hello"
// ┌──────────┬─────┐
// │  ptr     │ len │
// │  0xc000  │  5  │
// └──────────┴─────┘
//      │
//      ▼  (read-only memory)
//  h e l l o
```

## Interface header — `{type, data}`

```go
type InterfaceHeader struct {
    Type uintptr  // pointer to type descriptor
    Data uintptr  // pointer to concrete value
}
```

```go
var a Animal = &Dog{Name: "Rex"}
// ┌──────────┬──────────┐
// │  *Dog    │  0xc000  │
// │  type    │  data    │
// └──────────┴──────────┘
```

## Maps and channels are NOT headers

They are a **single pointer** to an internal runtime structure — no bundled metadata exposed to you:

```go
m := make(map[string]int)  // single ptr → runtime hash table
ch := make(chan int)        // single ptr → runtime channel struct
```

## Summary

| Type        | Internal structure | Words | Immutable? |
|-------------|-------------------|-------|------------|
| `[]T`       | `{ptr, len, cap}` | 3     | No         |
| `string`    | `{ptr, len}`      | 2     | Yes        |
| `interface` | `{type, data}`    | 2     | No         |
| `map[K]V`   | single pointer    | 1     | No         |
| `chan T`    | single pointer    | 1     | No         |

The header types (`[]T`, `string`, `interface`) are value types that *contain* a pointer — copying them is cheap but you get a new header pointing to the same underlying data.
