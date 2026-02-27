# Go: make() and Capacity

## Declaring with make

`make` only works with **slices, maps, and channels** — not arrays.

### Slice

```go
// make([]T, length, capacity)
s := make([]int, 3)      // [0 0 0]  len=3, cap=3
s := make([]int, 3, 10)  // [0 0 0]  len=3, cap=10 (pre-allocated)

// vs literal
s := []int{1, 2, 3}
```

### Map

```go
// make(map[K]V)
m := make(map[string]int)        // empty, ready to use
m := make(map[string]int, 100)   // hint: pre-allocate for ~100 entries (perf only)

// vs literal
m := map[string]int{"a": 1}
```

### Array — no make, use var or literal

```go
var a [3]int         // [0 0 0]  zero-valued
a := [3]int{1, 2, 3} // [1 2 3]
a := [...]int{1,2,3} // compiler counts the length
```

### Channel

```go
// make(chan T, bufferSize)
ch := make(chan int)    // unbuffered
ch := make(chan int, 5) // buffered, capacity 5
```

### When to use make vs literal

| Situation                             | Use                       |
| ------------------------------------- | ------------------------- |
| Empty map/slice to populate later     | `make`                    |
| Known initial values                  | literal `{}`              |
| Pre-allocate capacity for performance | `make` with capacity hint |

---

## What is Capacity

Capacity is the number of elements the underlying array can hold **before a new allocation is needed**.

A slice has two size concepts:

- `len` — elements currently in use
- `cap` — total space allocated

```go
s := make([]int, 3, 5)
//                ^  ^
//               len cap

fmt.Println(len(s)) // 3
fmt.Println(cap(s)) // 5

// underlying array looks like: [0, 0, 0, _, _]
//                                <--len-->
//                                <----cap---->
```

### append and reallocation

```go
s := make([]int, 3, 5)
s = append(s, 9)    // len=4, cap=5  — no alloc, fits in existing array
s = append(s, 9)    // len=5, cap=5  — no alloc, fits
s = append(s, 9)    // len=6, cap=10 — REALLOC, new backing array
```

When `append` exceeds capacity, Go:

1. Allocates a new, larger backing array
2. Copies all elements over
3. Returns a new slice header pointing to the new array

### Pre-allocating for performance

```go
// Bad — reallocates multiple times as it grows
s := make([]int, 0)
for i := 0; i < 10000; i++ {
    s = append(s, i)
}

// Good — single allocation upfront
s := make([]int, 0, 10000)
for i := 0; i < 10000; i++ {
    s = append(s, i)
}
```

---

## Slice Growth Strategy

Growth is not simply "double it" — it depends on current size and Go version.

- **Small slices**: roughly **2x**
- **Large slices**: slows down to roughly **1.25x**

The threshold and exact formula changed in **Go 1.18** to be a smoother curve.

```
len=1     new cap=1
len=2     new cap=2
len=3     new cap=4
len=5     new cap=8
len=9     new cap=16
len=17    new cap=32
...
len=1025  new cap=1280   ← growth slows here
len=1281  new cap=1696
len=1697  new cap=2048
```

Don't rely on the exact factor — it's a **runtime implementation detail** that can change between Go versions.

Growth is **amortised O(1)** per append — individual appends are cheap on average. If you know the size upfront, always use `make([]int, 0, n)` to avoid any reallocation.
