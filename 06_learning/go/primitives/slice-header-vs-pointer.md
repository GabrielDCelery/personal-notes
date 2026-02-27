# Go: Slice Header vs Pointer

A pointer is just a memory address — one field pointing to some data.

A slice header is a **struct with three fields** that Go manages internally:

```go
// What a slice actually is under the hood
type SliceHeader struct {
    Data uintptr  // pointer to the underlying array
    Len  int      // number of elements in use
    Cap  int      // total allocated space
}
```

So when you write:

```go
s := []int{1, 2, 3}
```

What you actually have in memory is:

```
s (slice header, lives on stack)
┌──────────┬─────┬─────┐
│  ptr     │ len │ cap │
│  0xc000  │  3  │  3  │
└──────────┴─────┴─────┘
     │
     ▼
0xc000 (underlying array, lives on heap)
┌───┬───┬───┐
│ 1 │ 2 │ 3 │
└───┴───┴───┘
```

## Why the distinction matters

When you pass a slice to a function, the **header is copied** (all 3 fields), but both headers point to the **same underlying array**:

```go
func modify(s []int) {
    s[0] = 99        // modifies shared array — caller sees it
    s = append(s, 4) // changes s.Len and s.Data locally only — caller does NOT see it
}
```

A plain pointer `*[]int` passes the address of the header itself, so the function can change `len`, `cap`, and `ptr` in the caller's copy:

```go
func grow(s *[]int) {
    *s = append(*s, 4) // replaces the caller's entire header
}
```

## Summary

|                        | Pointer `*T`     | Slice header `[]T`    |
| ---------------------- | ---------------- | --------------------- |
| Size                   | 1 word (8 bytes) | 3 words (24 bytes)    |
| Contains               | memory address   | ptr + len + cap       |
| Mutation of elements   | yes              | yes (shared array)    |
| Mutation of length/cap | yes (via `*`)    | no (header is a copy) |

The slice header is Go's way of bundling the pointer with its metadata so you don't have to manage length and capacity manually yourself.
