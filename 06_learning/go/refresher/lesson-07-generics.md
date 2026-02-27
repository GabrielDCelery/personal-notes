# Lesson 07: Generics

Go added generics in 1.18 after a decade of "just use interface{}." The result is more principled than many expected: not a full type-level programming system, but a pragmatic addition that solves the real pain points — type-safe collections, functional helpers, ordered types — without the complexity of C++ templates. Knowing when generics improve your code vs when they add unnecessary complexity is a key senior Go skill.

## Why Generics? The Problem They Solve

Before generics, writing a `Min` function meant either:

```go
// ❌ Option 1: One function per type (combinatorial explosion)
func MinInt(a, b int) int { if a < b { return a }; return b }
func MinFloat64(a, b float64) float64 { if a < b { return a }; return b }
func MinString(a, b string) string { if a < b { return a }; return b }

// ❌ Option 2: interface{} (loses type safety, requires runtime type assertions)
func Min(a, b interface{}) interface{} {
    // How do you compare? You'd need reflect or a switch on types
}
```

With generics:

```go
// ✓ One function, type-safe, no runtime overhead
func Min[T constraints.Ordered](a, b T) T {
    if a < b {
        return a
    }
    return b
}

Min(1, 2)           // T inferred as int
Min(1.5, 2.5)       // T inferred as float64
Min("apple", "banana") // T inferred as string
```

## Type Parameters

Type parameters are declared in square brackets before the function parameter list:

```go
func FunctionName[T constraint](params) returnType
```

```go
// Single type parameter
func Map[T, U any](s []T, f func(T) U) []U {
    result := make([]U, len(s))
    for i, v := range s {
        result[i] = f(v)
    }
    return result
}

// Usage - type inference works here:
doubled := Map([]int{1, 2, 3}, func(n int) int { return n * 2 })
strs := Map([]int{1, 2, 3}, func(n int) string { return fmt.Sprintf("%d", n) })
```

### Generic Types (Structs)

```go
type Stack[T any] struct {
    items []T
}

func (s *Stack[T]) Push(item T) {
    s.items = append(s.items, item)
}

func (s *Stack[T]) Pop() (T, bool) {
    if len(s.items) == 0 {
        var zero T
        return zero, false // return zero value for T
    }
    item := s.items[len(s.items)-1]
    s.items = s.items[:len(s.items)-1]
    return item, true
}

// Usage:
s := Stack[int]{}      // explicit type (required for types, unlike functions)
s.Push(1)
s.Push(2)
v, _ := s.Pop()        // v is int
```

**Note**: Type inference works for generic functions but NOT for generic types. You must specify the type argument when instantiating a generic type.

## Constraints

A **constraint** defines what type arguments are allowed. It's an interface.

### `any`

The broadest constraint — any type is allowed. Used when you only need zero values or storage.

```go
func First[T any](s []T) (T, bool) {
    if len(s) == 0 {
        var zero T
        return zero, false
    }
    return s[0], true
}
```

### `comparable`

Types that support `==` and `!=`. Required for map keys and set implementations.

```go
func Contains[T comparable](s []T, v T) bool {
    for _, item := range s {
        if item == v {
            return true
        }
    }
    return false
}

// Works with any comparable type:
Contains([]int{1, 2, 3}, 2)           // true
Contains([]string{"a", "b"}, "c")    // false
// Contains([][]int{{1}}, []int{1})   // ❌ compile error: slice is not comparable
```

### `constraints.Ordered`

Types that support `<`, `>`, `<=`, `>=` — all numeric types and strings. From the `golang.org/x/exp/constraints` package (or define your own).

```go
import "golang.org/x/exp/constraints"

func Max[T constraints.Ordered](a, b T) T {
    if a > b {
        return a
    }
    return b
}
```

**Or define it yourself** (the standard approach before `golang.org/x/exp` was available):

```go
type Ordered interface {
    ~int | ~int8 | ~int16 | ~int32 | ~int64 |
        ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
        ~float32 | ~float64 |
        ~string
}
```

### Custom Constraints

```go
// Constraint: types with a String() method
type Stringer interface {
    String() string
}

func Print[T Stringer](items []T) {
    for _, item := range items {
        fmt.Println(item.String())
    }
}

// Constraint: types that are either int or string (union)
type IntOrString interface {
    int | string
}

func ProcessID[T IntOrString](id T) {
    fmt.Println(id)
}
```

## The `~T` Underlying Type Constraint

`~T` means "any type whose **underlying type** is T." This is the difference between a constraint that accepts only `int` and one that accepts `int` and all named types based on `int`.

```go
type Celsius float64
type Fahrenheit float64

// Without ~: only accepts float64 literally
func DoubleF(v float64) float64 { return v * 2 }
// DoubleF(Celsius(100))    // ❌ Celsius is not float64

// With ~: accepts float64 AND any named type based on float64
type FloatLike interface {
    ~float64
}

func Double[T FloatLike](v T) T { return v * 2 }
Double(3.14)            // ✓ float64
Double(Celsius(100))    // ✓ Celsius (underlying type is float64)
Double(Fahrenheit(32))  // ✓ Fahrenheit
```

**Why this matters**: If you define a type `type Duration int64`, you want `Duration` to work with functions constrained on integer types. Without `~`, you'd have to convert to `int64` everywhere.

```go
// The Ordered constraint uses ~ for all types:
type Ordered interface {
    ~int | ~float64 | ~string | ...
}
```

## Type Inference

The compiler infers type arguments when they can be determined from function arguments:

```go
func Map[T, U any](s []T, f func(T) U) []U { ... }

// Explicit (verbose):
Map[int, string]([]int{1, 2}, strconv.Itoa)

// Inferred (idiomatic):
Map([]int{1, 2}, strconv.Itoa) // T=int, U=string inferred from args
```

Type inference works for functions. For types, you must be explicit:

```go
// ❌ Type inference not supported for generic types:
s := Stack{} // compile error

// ✓ Must specify:
s := Stack[int]{}
```

## Real-World Use Cases

### Type-Safe Collections

```go
// Set implementation
type Set[T comparable] struct {
    m map[T]struct{}
}

func NewSet[T comparable](items ...T) *Set[T] {
    s := &Set[T]{m: make(map[T]struct{})}
    for _, item := range items {
        s.m[item] = struct{}{}
    }
    return s
}

func (s *Set[T]) Add(item T) { s.m[item] = struct{}{} }
func (s *Set[T]) Contains(item T) bool { _, ok := s.m[item]; return ok }
func (s *Set[T]) Remove(item T) { delete(s.m, item) }
func (s *Set[T]) Len() int { return len(s.m) }
```

### Functional Helpers

```go
func Filter[T any](s []T, predicate func(T) bool) []T {
    var result []T
    for _, v := range s {
        if predicate(v) {
            result = append(result, v)
        }
    }
    return result
}

func Reduce[T, U any](s []T, initial U, f func(U, T) U) U {
    acc := initial
    for _, v := range s {
        acc = f(acc, v)
    }
    return acc
}

// Usage:
evens := Filter([]int{1, 2, 3, 4, 5}, func(n int) bool { return n%2 == 0 })
sum := Reduce([]int{1, 2, 3, 4, 5}, 0, func(acc, n int) int { return acc + n })
```

### Result/Option Types

```go
type Result[T any] struct {
    value T
    err   error
}

func OK[T any](v T) Result[T]        { return Result[T]{value: v} }
func Err[T any](err error) Result[T] { return Result[T]{err: err} }

func (r Result[T]) Unwrap() (T, error) { return r.value, r.err }
func (r Result[T]) IsOK() bool         { return r.err == nil }
```

## When Interfaces Are Still Better

Not every abstraction should use generics. Interfaces remain the right tool when:

**1. You need runtime polymorphism** (different types at the same call site):

```go
// ✓ Interface: you need to hold different concrete types at runtime
type Animal interface{ Speak() string }
animals := []Animal{Dog{}, Cat{}, Bird{}}
for _, a := range animals {
    fmt.Println(a.Speak()) // runtime dispatch to different types
}

// ❌ Generics can't do this: T is fixed at instantiation time
func Feed[T Animal](a T) { fmt.Println(a.Speak()) }
// You can't have []T where T is "any Animal" - T must be one specific type
```

**2. You need to add methods to the constraint** (generics can't call methods not in the constraint):

```go
// Interface is clearer when you need to call specific methods:
type Serializer interface {
    Serialize() []byte
    Deserialize([]byte) error
}
```

**3. The abstraction is about behaviour, not data structure**:

Generics are primarily about data structures and algorithms. Interfaces are about behaviour contracts.

| Use Generics For                                        | Use Interfaces For                |
| ------------------------------------------------------- | --------------------------------- |
| Type-safe collections (Stack, Set, Queue)               | Polymorphic behaviour (io.Reader) |
| Algorithms on ordered/comparable types (Min, Max, Sort) | Plugin systems and extensibility  |
| Functional helpers (Map, Filter, Reduce)                | Mocking and testing               |
| Result/Option wrapper types                             | Runtime type dispatch             |

## Common Gotchas

### Can't Use Type Parameters as Map Keys (unless constrained)

```go
// ❌ compile error: T is not comparable
func Count[T any](s []T) map[T]int {
    // Can't use T as map key - not guaranteed comparable
}

// ✓ Constrain to comparable
func Count[T comparable](s []T) map[T]int {
    m := make(map[T]int)
    for _, v := range s {
        m[v]++
    }
    return m
}
```

### Zero Values of Type Parameters

When you need to return "nothing" for a generic type:

```go
// ✓ Declare zero value using var
func First[T any](s []T) (T, bool) {
    if len(s) == 0 {
        var zero T   // zero value of T (0 for int, "" for string, nil for pointer, etc.)
        return zero, false
    }
    return s[0], true
}
```

### Methods Can't Have Type Parameters

```go
type MyStruct struct{}

// ❌ Methods can't have their own type parameters
func (m MyStruct) DoThing[T any](v T) T { return v }

// ✓ Put the type parameter on the type instead
type MyStruct[T any] struct{}
func (m MyStruct[T]) DoThing(v T) T { return v }

// Or use a function:
func DoThing[T any](m MyStruct, v T) T { return v }
```

## Hands-On Exercise 1: Generic Ordered Map

Implement a map that keeps keys in sorted order.

```go
// Requirements:
// 1. OrderedMap[K, V] stores key-value pairs
// 2. Set(key K, value V) adds or updates
// 3. Get(key K) (V, bool) retrieves
// 4. Keys() []K returns keys in sorted order
// 5. K must be comparable and ordered
```

<details>
<summary>Solution</summary>

```go
package main

import (
    "fmt"
    "sort"

    "golang.org/x/exp/constraints"
)

type OrderedMap[K constraints.Ordered, V any] struct {
    keys   []K
    values map[K]V
}

func NewOrderedMap[K constraints.Ordered, V any]() *OrderedMap[K, V] {
    return &OrderedMap[K, V]{
        values: make(map[K]V),
    }
}

func (m *OrderedMap[K, V]) Set(key K, value V) {
    if _, exists := m.values[key]; !exists {
        m.keys = append(m.keys, key)
        sort.Slice(m.keys, func(i, j int) bool {
            return m.keys[i] < m.keys[j]
        })
    }
    m.values[key] = value
}

func (m *OrderedMap[K, V]) Get(key K) (V, bool) {
    v, ok := m.values[key]
    return v, ok
}

func (m *OrderedMap[K, V]) Keys() []K {
    result := make([]K, len(m.keys))
    copy(result, m.keys)
    return result
}

func main() {
    m := NewOrderedMap[string, int]()
    m.Set("banana", 3)
    m.Set("apple", 1)
    m.Set("cherry", 2)
    fmt.Println(m.Keys()) // [apple banana cherry]
    v, _ := m.Get("apple")
    fmt.Println(v) // 1
}
```

</details>

## Hands-On Exercise 2: Generic Channel Pipeline

Implement generic `Map` and `Filter` operations that work on channels.

```go
// Requirements:
// 1. ChanMap[T, U](ctx, in <-chan T, f func(T) U) <-chan U
//    transforms each value from in using f
// 2. ChanFilter[T](ctx, in <-chan T, pred func(T) bool) <-chan T
//    forwards only values that pass pred
// 3. Both must stop when ctx is cancelled
// 4. Both must close their output when input is exhausted
```

<details>
<summary>Solution</summary>

```go
package main

import (
    "context"
    "fmt"
)

func ChanMap[T, U any](ctx context.Context, in <-chan T, f func(T) U) <-chan U {
    out := make(chan U)
    go func() {
        defer close(out)
        for v := range in {
            select {
            case out <- f(v):
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}

func ChanFilter[T any](ctx context.Context, in <-chan T, pred func(T) bool) <-chan T {
    out := make(chan T)
    go func() {
        defer close(out)
        for v := range in {
            if pred(v) {
                select {
                case out <- v:
                case <-ctx.Done():
                    return
                }
            }
        }
    }()
    return out
}

func main() {
    ctx := context.Background()

    nums := make(chan int, 10)
    for i := 1; i <= 10; i++ {
        nums <- i
    }
    close(nums)

    evens := ChanFilter(ctx, nums, func(n int) bool { return n%2 == 0 })
    squares := ChanMap(ctx, evens, func(n int) int { return n * n })

    for v := range squares {
        fmt.Println(v) // 4, 16, 36, 64, 100
    }
}
```

</details>

## Interview Questions

### Q1: What problem do generics solve in Go, and what were the alternatives before Go 1.18?

Tests understanding of the design history and motivations.

<details>
<summary>Answer</summary>

Before generics, Go developers had three options for writing type-independent code:

1. **Code duplication**: Write `MinInt`, `MinFloat64`, `MinString` separately. Type-safe but explosive maintenance burden.

2. **`interface{}`**: Write functions that accept `interface{}` and use type assertions. Loses compile-time type safety, adds runtime overhead, requires boilerplate assertions, and shifts errors to runtime.

3. **Code generation**: Use `go generate` with tools like `genny` or `stringer` to stamp out typed versions. Correct but adds tooling complexity and generated files to check in.

Generics solve this by allowing type-parameterized functions and types where the compiler verifies type safety at instantiation time. Benefits:

- No code duplication
- Full type safety (errors at compile time)
- No runtime overhead from interface boxing (monomorphization in the compiler)
- No need for `.(Type)` assertions

The primary use cases are: type-safe collections (Stack, Set, Queue), algorithms (Min, Max, Sort, Map, Filter), and wrapper types (Result, Option).

</details>

### Q2: What is the `~T` constraint syntax and when do you need it?

Tests understanding of the type system's distinction between named types and their underlying types.

<details>
<summary>Answer</summary>

In Go, you can define a named type based on an existing type:

```go
type Celsius float64
type UserID int64
```

`Celsius` and `float64` are distinct types even though they have the same underlying type (`float64`). A constraint `float64` without `~` only accepts the literal `float64` type — not `Celsius`.

`~float64` means "any type whose underlying type is `float64`." It accepts both `float64` and any named type based on `float64` (like `Celsius`).

**When you need `~`**: When your constraint should work with named types that are semantically distinct but share the same underlying type. The `constraints.Ordered` interface uses `~int | ~float64 | ~string | ...` specifically so that user-defined types like `type Duration int64` or `type Celsius float64` satisfy the constraint.

```go
// Without ~: only int works
type IntOnly interface{ int }

// With ~: int, MyInt, UserID (underlying=int) all work
type IntLike interface{ ~int }

type UserID int
func foo[T IntLike](v T) T { return v * 2 }
foo(UserID(5)) // ✓ works with ~
```

</details>

### Q3: When should you use generics vs interfaces?

The design judgment question — tests whether you understand the appropriate tool for each problem.

<details>
<summary>Answer</summary>

**Use generics when**:

- You're writing a data structure that should work with any type (Stack, Set, Queue, Result)
- You're writing an algorithm that works on any comparable or ordered type (Min, Max, BinarySearch)
- You need type safety that `interface{}` can't provide
- You're writing functional helpers (Map, Filter, Reduce) where the input and output types matter
- Performance matters and you want to avoid interface boxing overhead

**Use interfaces when**:

- You need runtime polymorphism — different concrete types at the same call site
- You're defining a behaviour contract that many types implement (io.Reader, http.Handler)
- You're designing extension points (plugins, hooks)
- You need to store mixed types in a collection
- You're mocking in tests (interface + mock type)

**Key insight**: Generics are about data and algorithms; interfaces are about behaviour. An `io.Reader` isn't a generic collection — it's a behavioural contract. A `Stack[T]` isn't a behaviour — it's a data structure that happens to work with any type.

**Warning signs you're overusing generics**:

- The type parameter is always `any` and you don't use any type-specific operations
- You have one instantiation and never plan others
- The code is harder to read with the type parameter than without

</details>

### Q4: Why can't methods have their own type parameters in Go?

A subtler question about the design constraints of Go's generics implementation.

<details>
<summary>Answer</summary>

Go's generics specification explicitly excludes type parameters on methods (as opposed to the receiver type). This is a deliberate design decision, not an oversight.

**Why**: The Go team's primary concern is implementation complexity and interface satisfaction. If methods could have type parameters, it would break how interface satisfaction works. An interface defines a fixed method set. If `func (t T) Do[U any]()` were allowed, the interface would need to specify `Do` for all possible `U` — infinitely many methods.

Additionally, method type parameters would complicate the type system significantly (interaction with method sets, interface embedding, type assertions) for a use case that can usually be handled by:

1. **Putting the type parameter on the receiver type**:

   ```go
   type Container[T any] struct{ ... }
   func (c Container[T]) Process(v T) { ... }
   ```

2. **Using a package-level function**:

   ```go
   func Process[T any](c Container, v T) { ... }
   ```

3. **Using an interface parameter**:
   ```go
   func (c Container) Process(v any) { ... }
   ```

In practice, most uses of "generic methods" are better served by one of these alternatives.

</details>

## Key Takeaways

1. **Generics add type parameters**: `func F[T constraint](v T)` — type is inferred from arguments.
2. **Constraints are interfaces**: `any`, `comparable`, `~int | ~string`, or any interface type.
3. **`~T` for named types**: `~float64` accepts `Celsius`, `Temperature`, etc. whose underlying type is `float64`.
4. **Type inference for functions**: The compiler infers `T` from arguments; no need to write `F[int](...)` usually.
5. **Explicit for types**: `Stack[int]{}` — type inference doesn't work for generic types.
6. **`comparable` for map keys**: If you need `==` on your type parameter, constrain to `comparable`.
7. **Zero values**: Use `var zero T` to get the zero value for a type parameter.
8. **Methods can't have type params**: Put the type param on the struct, or use a function.
9. **Interfaces vs generics**: Interfaces for behaviour contracts; generics for type-safe data structures and algorithms.

## Next Steps

In [Lesson 08: Testing & Benchmarking](lesson-08-testing-and-benchmarking.md), you'll learn:

- Table-driven tests and subtests (`t.Run`) for organized, readable test suites
- The race detector and why it catches bugs that aren't otherwise apparent
- Writing benchmarks with `b.N` and interpreting the output
- Profiling with `pprof` to find allocation and CPU hotspots
