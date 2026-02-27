# Lesson 12: Slices, Maps & String Internals

Go's three most-used data structures — slices, maps, and strings — all look simple on the surface but have internal mechanics that trip up even experienced developers. Understanding the backing array, hash-table growth, and UTF-8 byte layout is the difference between writing correct, efficient Go and writing code that silently corrupts data, panics in production, or confuses you during a code review.

## Slice Internals: The Three-Field Header

A slice is not an array. Every slice value is a small struct with three fields:

```
┌─────────┬────────┬──────────┐
│ pointer │ length │ capacity │
└─────────┴────────┴──────────┘
    │
    ▼
┌───┬───┬───┬───┬───┬───┐
│ 0 │ 1 │ 2 │ 3 │ 4 │ 5 │  ← backing array (may be larger than len)
└───┴───┴───┴───┴───┴───┘
```

```go
s := make([]int, 3, 6)
// pointer → underlying array
// len = 3  (elements you can access via s[i])
// cap = 6  (elements before a new allocation is needed)

fmt.Println(len(s)) // 3
fmt.Println(cap(s)) // 6
```

`len` is the number of elements visible through the slice. `cap` is how many elements the backing array can hold starting from the slice's pointer. You can re-slice up to `cap`:

```go
s2 := s[:5]   // ✓ valid — within capacity
s3 := s[:7]   // ❌ panic: slice bounds out of range [:7] with capacity 6
```

### How `append` Works

`append` checks whether `len < cap`. If there is room, it writes into the existing backing array and returns a slice with incremented `len`. If `len == cap`, it allocates a new backing array, copies all elements, then appends.

```go
s := make([]int, 3, 4)   // len=3, cap=4

// Room available — no allocation
s = append(s, 99)         // len=4, cap=4, same backing array

// No room — new allocation
s = append(s, 100)        // len=5, cap=8 (or similar), NEW backing array
```

**This is why you must always capture the return value of `append`.**

```go
// ❌ Wrong — original slice unchanged if reallocation occurred
append(s, 99)

// ✓ Correct
s = append(s, 99)
```

### Capacity Growth Algorithm

| Current capacity | Approximate growth strategy          |
| ---------------- | ------------------------------------ |
| < 256            | Doubles (×2)                         |
| 256 – ~2048      | Grows by ~1.25x–2x (blended formula) |
| > 2048           | Grows by ~1.25x                      |

The exact formula changed in Go 1.18 (from a hard threshold of 1024 to a smooth curve). Small slices grow fast (doubles), large slices grow more conservatively.

**Practical implication**: Pre-allocate with `make([]T, 0, n)` when you know the final size:

```go
// ❌ Triggers repeated reallocations
result := []int{}
for i := range items {
    result = append(result, process(items[i]))
}

// ✓ One allocation, no copies
result := make([]int, 0, len(items))
for i := range items {
    result = append(result, process(items[i]))
}
```

---

## Append Aliasing Gotcha

When two slices share the same backing array, a write through one is visible through the other.

```go
a := make([]int, 3, 5)   // len=3, cap=5
a[0], a[1], a[2] = 1, 2, 3

b := a[:3]   // b shares a's backing array

b = append(b, 99)
fmt.Println(a)       // [1 2 3]    — a's len is still 3, doesn't see 99
fmt.Println(b)       // [1 2 3 99] — b sees it

a2 := a[:4]          // a2 sees position [3] = 99 (written by b's append!)
fmt.Println(a2)      // [1 2 3 99] — a2 sees b's write
```

### The Fix: Full Slice Expression (`s[low:high:max]`)

```go
a := make([]int, 3, 5)
a[0], a[1], a[2] = 1, 2, 3

// ✓ Cap b at len=3, so any append to b must allocate a new array
b := a[0:3:3]   // len=3, cap=3

b = append(b, 99)   // cap exceeded → new backing array allocated
fmt.Println(a)      // [1 2 3] — untouched
fmt.Println(b)      // [1 2 3 99] — b's own array
```

```go
// ❌ Caller can mutate the original backing array via append
func first3(s []int) []int {
    return s[:3]
}

// ✓ Caller's append won't touch the original
func first3(s []int) []int {
    return s[0:3:3]
}
```

---

## nil vs Empty Slice

```go
var s1 []int         // nil slice:   ptr=nil, len=0, cap=0
s2 := []int{}        // empty slice: ptr≠nil, len=0, cap=0
s3 := make([]int, 0) // empty slice: ptr≠nil, len=0, cap=0
```

All three: `len==0`, safe to `range` over, `append` works identically.

**JSON marshaling** — this is the one that bites you in APIs:

```go
var s1 []int      // nil slice
s2 := []int{}     // empty slice

b1, _ := json.Marshal(s1)   // → null
b2, _ := json.Marshal(s2)   // → []
```

**`reflect.DeepEqual`**:

```go
reflect.DeepEqual([]int(nil), []int{})  // false
```

| Situation                                      | Use                       |
| ---------------------------------------------- | ------------------------- |
| Returning from a function that found nothing   | `nil` is fine             |
| Returning JSON that clients expect to be array | `[]T{}` or `make([]T, 0)` |
| Need `reflect.DeepEqual` to match empty slice  | `[]T{}`                   |

---

## Copy Semantics

### `copy(dst, src)`

`copy` copies `min(len(dst), len(src))` elements. **It does not grow `dst`.**

```go
src := []int{1, 2, 3, 4, 5}

// ❌ dst has len=0, copy copies 0 elements
dst1 := make([]int, 0, 5)
n := copy(dst1, src)
fmt.Println(n, dst1) // 0 []

// ✓ dst has len=5
dst2 := make([]int, 5)
n = copy(dst2, src)
fmt.Println(n, dst2) // 5 [1 2 3 4 5]
```

### 2D Slice Pitfall

```go
// ❌ All rows share the same backing array
matrix := make([][]int, 3)
backing := make([]int, 9)
for i := range matrix {
    matrix[i] = backing[i*3 : i*3+3]
}

// ✓ Each row has its own backing array
matrix := make([][]int, 3)
for i := range matrix {
    matrix[i] = make([]int, 3)
}
```

### Passing Slices to Functions

The **header is copied**, but the **backing array is shared**:

```go
func double(s []int) {
    for i := range s {
        s[i] *= 2   // ✓ modifies the caller's backing array
    }
}

s := []int{1, 2, 3}
double(s)
fmt.Println(s) // [2 4 6] — caller sees the mutation
```

Functions can mutate elements. They cannot grow the slice the caller sees unless they return the new slice.

---

## Map Internals

Go maps are hash tables implemented with **buckets**, where each bucket holds up to 8 key-value pairs.

### Why Iteration Order Is Randomized

Since Go 1.0, map iteration order is deliberately randomized. The Go team introduced this intentionally — before Go 1, maps had a quasi-stable iteration order that programs accidentally relied on.

```go
m := map[string]int{"a": 1, "b": 2, "c": 3}

// Different order every run — by design
for k, v := range m {
    fmt.Println(k, v)
}
```

**Gotcha in tests**: Sort the keys first:

```go
// ❌ Flaky test — map order is non-deterministic
keys := []string{}
for k := range m { keys = append(keys, k) }

// ✓ Stable
keys := make([]string, 0, len(m))
for k := range m { keys = append(keys, k) }
sort.Strings(keys)
```

### Load Factor and Growth

Pre-size maps to avoid rehashing:

```go
// ❌ Repeated growth triggers rehashing
m := make(map[string]int)

// ✓ Hint avoids rehashing
m := make(map[string]int, expectedSize)
```

### Zero Value for Missing Keys

```go
m := map[string]int{}
v := m["missing"]          // v = 0, no panic

v, ok := m["missing"]      // v=0, ok=false — use this to distinguish missing from zero
v, ok  = m["present"]      // v=..., ok=true
```

---

## Map Gotchas

### Concurrent Read/Write: Fatal Panic, Not Just a Data Race

```go
// ❌ Fatal: "concurrent map read and map write" — not recoverable
go func() { m["a"] = 1 }()
go func() { _ = m["b"] }()
```

```go
// ✓ Protected with mutex
var mu sync.RWMutex
mu.Lock(); m["a"] = 1; mu.Unlock()
mu.RLock(); v := m["a"]; mu.RUnlock()
```

### Assigning to a Struct Field in a Map

```go
type Point struct{ X, Y int }
m := map[string]Point{"p": {1, 2}}

// ❌ Compile error: cannot assign to m["p"].X
m["p"].X = 10

// ✓ Read, modify, write back
p := m["p"]; p.X = 10; m["p"] = p

// ✓ Or use a pointer map
pm := map[string]*Point{"p": {1, 2}}
pm["p"].X = 10   // valid — pm["p"] is a pointer
```

### nil Map: Read vs Write

```go
var m map[string]int   // nil map

v := m["key"]          // ✓ safe — returns 0
v, ok := m["key"]      // ✓ safe — v=0, ok=false
delete(m, "key")       // ✓ safe — no-op on nil map

m["key"] = 1           // ❌ panic: assignment to entry in nil map
```

---

## String Internals

A Go string is an **immutable** sequence of bytes with a two-field header (pointer + length). No `cap` field — strings cannot be grown.

**`len(s)` returns bytes, not characters:**

```go
s := "Hello, 世界"
fmt.Println(len(s))             // 13 — 7 ASCII bytes + 3+3 bytes for 世界
fmt.Println(len([]rune(s)))     // 9 — 9 Unicode code points
```

### String/[]byte Conversion Cost

Converting between `string` and `[]byte` **copies the data**:

```go
s := "hello"
b := []byte(s)   // copy — b is independent of s
b[0] = 'H'
fmt.Println(s)   // "hello" — unaffected
```

Use `strings.Builder` for incremental construction:

```go
// ❌ O(n²) allocations
result := ""
for _, s := range parts { result += s }

// ✓ Single allocation at the end
var sb strings.Builder
for _, s := range parts { sb.WriteString(s) }
result := sb.String()
```

---

## Rune Iteration

### `s[i]` gives a byte

```go
s := "café"
fmt.Println(s[3])          // 195 (first byte of 'é', not the character)
fmt.Printf("%c\n", s[3])  // Ã  — garbage
```

### `range s` gives rune + byte offset

```go
s := "café"
for i, r := range s {
    fmt.Printf("byte offset %d: %c (U+%04X)\n", i, r, r)
}
// byte offset 0: c (U+0063)
// byte offset 1: a (U+0061)
// byte offset 2: f (U+0066)
// byte offset 3: é (U+00E9)   ← é is 2 bytes; next char would be at offset 5
```

`range` decodes UTF-8 automatically. `i` is the **byte offset**, not the rune index.

```go
import "unicode/utf8"

s := "café"
fmt.Println(len(s))                      // 5 bytes
fmt.Println(utf8.RuneCountInString(s))   // 4 characters
```

### Multi-byte UTF-8 Characters

| Character | Code Point | UTF-8 Bytes           | `len` contribution |
| --------- | ---------- | --------------------- | ------------------ |
| `A`       | U+0041     | `0x41`                | 1                  |
| `é`       | U+00E9     | `0xC3 0xA9`           | 2                  |
| `世`      | U+4E16     | `0xE4 0xB8 0x96`      | 3                  |
| `😀`      | U+1F600    | `0xF0 0x9F 0x98 0x80` | 4                  |

```go
// ❌ Slicing by byte index mid-character corrupts the string
truncated := s[:3]   // might cut 'é' in half

// ✓ Convert to runes, slice, convert back
runes := []rune(s)
truncated := string(runes[:2])
```

### Quick Reference: String Operations

| Operation                          | What you get                                            |
| ---------------------------------- | ------------------------------------------------------- |
| `s[i]`                             | Byte at position i (`byte`/`uint8`)                     |
| `s[i:j]`                           | Byte sub-slice as string (may corrupt multi-byte chars) |
| `range s`                          | `(byteOffset int, r rune)` — UTF-8 safe                 |
| `len(s)`                           | Byte count                                              |
| `utf8.RuneCountInString(s)`        | Character (rune) count                                  |
| `[]rune(s)`                        | Slice of Unicode code points (copies)                   |
| `[]byte(s)`                        | Byte slice (copies)                                     |
| `strings.Builder` / `bytes.Buffer` | Efficient string construction                           |

---

## Hands-On Exercise 1: Slice Aliasing Bug Hunt

The following code has a subtle aliasing bug. Identify it and fix it.

```go
func getFirstN(s []int, n int) []int {
    return s[:n]
}

func main() {
    data := []int{1, 2, 3, 4, 5}
    first3 := getFirstN(data, 3)

    first3 = append(first3, 99)

    fmt.Println("data:", data)     // What does this print?
    fmt.Println("first3:", first3)
}
```

**Questions**:

1. What does `data` print, and why?
2. How do you fix `getFirstN` so that appending to the result never affects `data`?
3. What would happen if `data` was created with `make([]int, 5, 5)`?

<details>
<summary>Solution</summary>

**1. What `data` prints**:

```
data:  [1 2 3 99 5]
first3: [1 2 3 99]
```

`getFirstN(data, 3)` returns `data[:3]` — pointer=&data[0], len=3, cap=5. When `append(first3, 99)` runs, `len(3) < cap(5)`, so no reallocation occurs. It writes `99` into `data[3]`, mutating `data`.

**2. Fix using the full slice expression**:

```go
// ✓ Cap forces copy on next append
func getFirstN(s []int, n int) []int {
    return s[0:n:n]   // len=n, cap=n
}
```

Alternative — explicit copy:

```go
func getFirstN(s []int, n int) []int {
    result := make([]int, n)
    copy(result, s[:n])
    return result
}
```

**3. With `make([]int, 5, 5)`**:

`data` has `len=5, cap=5`. The slice `data[:3]` still has `cap=5`. The append STILL writes into `data[3]` — the bug is independent of whether `data`'s len matches its cap.

</details>

## Hands-On Exercise 2: Unicode String Processor

Write `truncateString(s string, maxRunes int) string` that truncates to at most `maxRunes` Unicode characters without corrupting multi-byte characters. Then write `charFrequency(s string) map[rune]int` that counts occurrences of each Unicode character.

```go
truncateString("café 世界", 4)    // "café"
truncateString("café 世界", 100)  // "café 世界"
charFrequency("hello")            // map[h:1 e:1 l:2 o:1]
```

<details>
<summary>Solution</summary>

```go
import "unicode/utf8"

func truncateString(s string, maxRunes int) string {
    if utf8.RuneCountInString(s) <= maxRunes {
        return s
    }
    count := 0
    for i := range s {   // i is the byte offset of each rune
        if count == maxRunes {
            return s[:i]  // slice at the byte boundary — always safe
        }
        count++
    }
    return s
}

func charFrequency(s string) map[rune]int {
    freq := make(map[rune]int)
    for _, r := range s {   // range decodes UTF-8 automatically
        freq[r]++
    }
    return freq
}
```

**Key points**:

- `for i := range s` — `i` is the byte offset of the rune boundary. Slicing `s[:i]` is always safe.
- `utf8.RuneCountInString` counts characters correctly for multi-byte input.
- `for _, r := range s` — `r` is the decoded `rune`. This is the idiomatic way to iterate over characters.

</details>

---

## Interview Questions

### Q1: You have two slices pointing to the same backing array. After appending to one, you check both — results are unexpected. Walk me through what happened and how to prevent it.

Interviewers ask this to probe whether you understand that a slice is a view into an array, not an independent collection. Developers who haven't hit this bug treat slices as value types, which leads to subtle data corruption that only surfaces under specific capacity conditions.

<details>
<summary>Answer</summary>

A slice header has three fields: pointer, length, capacity. When you sub-slice (`b := a[:3]`), `b` gets the same pointer and a new length — both point at the same memory.

When you `append(b, x)`:

- If `len(b) < cap(b)` — no reallocation. The value is written at `backing[len(b)]`. If `a`'s capacity extends past `len(b)`, reslicing `a` beyond its original length exposes the value you appended to `b`.
- If `len(b) == cap(b)` — reallocation occurs. `b` now has its own backing array and subsequent writes no longer affect `a`.

**Prevention** — use the full slice expression:

```go
b := a[0:3:3]   // len=3, cap=3 — any append copies to a new array
```

Or copy explicitly when you want complete independence immediately:

```go
b := make([]int, 3)
copy(b, a[:3])
```

</details>

### Q2: What is the difference between a nil slice and an empty slice, and when does it actually matter?

Interviewers ask this to distinguish developers who know the language deeply from those who just write working code. Many experienced Go developers have shipped JSON APIs that return `null` for empty arrays because of this distinction.

<details>
<summary>Answer</summary>

Both have `len == 0`, are safe to `range` over, and `append` works identically on both. The differences:

**JSON marshaling** — most impactful:

```go
var s []int       // nil → JSON null
s2 := []int{}     // non-nil → JSON []
```

Client code expecting an array and receiving `null` often panics. Always return `[]T{}` or `make([]T, 0)` in API response types.

**`reflect.DeepEqual`**:

```go
reflect.DeepEqual([]int(nil), []int{})  // false
```

Test assertions using `reflect.DeepEqual` (testify's `assert.Equal`) will fail when comparing a nil return against `[]int{}`.

**Nil check**:

```go
var s []int; s == nil    // true
s2 := []int{}; s2 == nil // false
```

Rule of thumb: use `nil` internally as a sentinel for "not yet initialized"; use `[]T{}` for anything that crosses an API boundary (JSON, gRPC, function returns that callers range over).

</details>

### Q3: Why does Go panic on concurrent map access, and what are your options for making a map safe for concurrent use?

Tests understanding of Go's concurrency model and the deliberate design decision to fatal-panic on map races rather than silently corrupt data.

<details>
<summary>Answer</summary>

Go maps are not safe for concurrent use by design. The Go team chose not to build locking into every map operation — that would add overhead to all single-goroutine uses. Instead, the runtime detects concurrent access and causes a fatal, non-recoverable panic (`concurrent map read and map write`).

**Why a fatal panic rather than silent corruption**: Concurrent map access can corrupt internal hash table structures causing infinite loops, wrong reads, or memory corruption. Making it loud and immediate is safer than silent data corruption.

**Options for concurrent use**:

**1. `sync.RWMutex`** — general purpose, reads don't block each other:

```go
var mu sync.RWMutex
var m = make(map[string]int)

// Write
mu.Lock(); m[k] = v; mu.Unlock()

// Read
mu.RLock(); v := m[k]; mu.RUnlock()
```

**2. `sync.Map`** — optimized for specific patterns:

```go
var m sync.Map
m.Store("key", 42)
v, ok := m.Load("key")
m.Range(func(k, v any) bool { return true })
```

`sync.Map` is NOT a general replacement for `map + mutex`. It is optimized for: (a) keys written once and read many times, (b) many goroutines operating on disjoint key sets. For general workloads, `map + RWMutex` is often faster and always clearer.

</details>

### Q4: What does `range` give you when iterating over a string, and why does `s[i]` give a different result?

Tests understanding of UTF-8 encoding and the byte/rune distinction. Developers who only handle ASCII never notice the difference — until their code encounters non-ASCII input in production and silently corrupts it.

<details>
<summary>Answer</summary>

**`s[i]` gives a `byte` (`uint8`)** — the raw byte at that position. For ASCII (U+0000–U+007F) this equals the Unicode code point. For multi-byte characters, you get one byte of a multi-byte UTF-8 sequence, which is not a valid character:

```go
s := "café"
fmt.Println(s[3])         // 195 (0xC3 — first byte of 'é')
fmt.Printf("%c\n", s[3]) // Ã   — garbage
```

**`range s` gives `(byteOffset int, r rune)`** — the runtime decodes UTF-8 and yields the full Unicode code point:

```go
for i, r := range "café" {
    fmt.Printf("offset %d: %c\n", i, r)
}
// offset 0: c
// offset 1: a
// offset 2: f
// offset 3: é   ← 'é' occupies bytes 3–4; next rune starts at byte 5
```

`i` increments by the byte width of each rune (1–4 bytes), not by 1.

**`len(s)` returns bytes**; for character count use `utf8.RuneCountInString(s)`.

**Practical rules**:

- Use `range s` for character-level iteration — never `for i := 0; i < len(s); i++`
- For character-level slicing: convert to `[]rune`, slice, convert back
- `s[i:j]` is a byte slice — only safe if `i` and `j` are rune boundaries (i.e., obtained from `range`)

</details>

---

## Key Takeaways

1. **Slice header**: pointer + length + capacity — a slice is a view, not an independent container. Copying a slice copies the header, not the backing array.
2. **`append` allocates only when `len == cap`**: if there is room, it writes into the existing array — potentially overwriting data visible through other slices sharing that array.
3. **Full slice expression `s[low:high:max]`**: caps capacity so the next append always copies, preventing aliasing bugs when returning sub-slices.
4. **nil vs empty slice**: both have `len=0` and behave identically for `append` and `range`, but `nil` marshals to JSON `null`. Always return `[]T{}` in API responses.
5. **Map iteration order is randomized by design**: never write order-dependent code; sort keys explicitly when a stable order is required.
6. **Concurrent map access causes a fatal panic**: not just a data race — a deliberate runtime detection. Use `sync.RWMutex` for general concurrent maps, `sync.Map` for append-mostly or disjoint-key workloads.
7. **Assigning to a struct field in a map is a compile error**: map values are not addressable. Read, modify, and write back — or use a pointer map.
8. **`len(s)` counts bytes, not characters**: for Unicode-correct character counting use `utf8.RuneCountInString(s)`.
9. **`s[i]` gives a byte; `range s` gives a rune**: byte indexing corrupts multi-byte characters. Use `range` or convert to `[]rune` for character-level operations.
10. **String/[]byte conversion copies**: use `strings.Builder` for incremental string construction to avoid O(n²) allocations.

## Next Steps

In [Lesson 13: HTTP Server Patterns](lesson-13-http-server-patterns.md), you'll learn:

- The `http.Handler` interface and the `HandlerFunc` adapter
- Middleware chains with `func(http.Handler) http.Handler`
- `http.ServeMux` routing rules and Go 1.22 method+path patterns
- The four server timeouts and what each protects against
- Graceful shutdown with `server.Shutdown(ctx)` and signal handling
- Response writing gotchas: `WriteHeader` order, double writes, status capture in middleware
