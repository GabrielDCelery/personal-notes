# Lesson 01: Interfaces & the Type System

Go's interface system is unlike anything in Java or C#. There are no `implements` keywords, no class hierarchies, and yet Go achieves polymorphism more flexibly than most OOP languages. Understanding how Go's type system actually works — especially the nil interface trap — is critical for interviews and for writing correct code.

## Duck Typing & Implicit Satisfaction

In most languages you declare that a type implements an interface. In Go, a type satisfies an interface automatically if it has the required methods. This is **structural typing** (often called duck typing): if it has the methods, it qualifies.

```go
type Stringer interface {
    String() string
}

type User struct {
    Name string
}

func (u User) String() string {
    return u.Name
}

// User satisfies Stringer - no declaration needed
var s Stringer = User{Name: "Alice"} // ✓ compiles
```

**Why this matters**: You can satisfy interfaces defined in packages you don't own. You can retrofit interfaces onto existing types without modifying them. This enables loose coupling without frameworks.

### Interface Satisfaction Rules

| Receiver type    | Value variable      | Pointer variable |
| ---------------- | ------------------- | ---------------- |
| Value `(t T)`    | ✓ satisfies         | ✓ satisfies      |
| Pointer `(t *T)` | ❌ does not satisfy | ✓ satisfies      |

```go
type Greeter interface {
    Greet() string
}

type Robot struct{}

func (r *Robot) Greet() string { return "Beep" } // pointer receiver

var g Greeter = &Robot{}  // ✓ pointer satisfies
var g2 Greeter = Robot{}  // ❌ compile error: Robot does not implement Greeter
```

**Rule**: A pointer to T has all of T's methods plus its own. T alone does not have pointer-receiver methods.

## Embedding (Composition Over Inheritance)

Go has no inheritance. Instead, you **embed** types to compose behaviour. This is not a workaround — it's idiomatic Go.

```go
type Animal struct {
    Name string
}

func (a Animal) Speak() string {
    return a.Name + " says something"
}

type Dog struct {
    Animal        // embedded - Dog "inherits" Animal's methods
    Breed string
}

func (d Dog) Speak() string { // Dog overrides Animal.Speak
    return d.Name + " barks"
}

d := Dog{Animal: Animal{Name: "Rex"}, Breed: "Labrador"}
fmt.Println(d.Speak())  // "Rex barks" - uses Dog's method
fmt.Println(d.Name)     // "Rex" - promoted field from Animal
```

### Interface Embedding

Interfaces can also embed other interfaces:

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

type ReadWriter interface {
    Reader  // embeds Reader
    Writer  // embeds Writer
}
```

This is how `io.ReadWriter`, `io.ReadWriteCloser`, and similar interfaces are built throughout the standard library.

## The Nil Interface Trap

This is one of Go's most common gotchas and a favourite interview question.

An interface value has **two components**: a type and a value. An interface is only `nil` if **both** are nil.

```go
var err error       // nil interface: type=nil, value=nil
fmt.Println(err == nil) // true

var p *MyError = nil
err = p             // non-nil interface: type=*MyError, value=nil
fmt.Println(err == nil) // FALSE - this surprises everyone
```

**Why?** Because `err` now has type information (`*MyError`), even though the underlying pointer is nil. The interface itself is not nil.

```go
// ❌ Classic bug: returning a typed nil
func getError(fail bool) error {
    var err *ValidationError = nil
    if fail {
        err = &ValidationError{msg: "bad input"}
    }
    return err // always returns non-nil interface!
}

// ✓ Correct: return untyped nil
func getError(fail bool) error {
    if fail {
        return &ValidationError{msg: "bad input"}
    }
    return nil // untyped nil = nil interface
}
```

**Memory model of an interface value:**

```
interface value:
┌──────────┬──────────┐
│  *type   │  *value  │
└──────────┴──────────┘
   nil means both fields are zero
```

## Type Assertions

A type assertion extracts the concrete value from an interface.

```go
var i interface{} = "hello"

// Single-return: panics if wrong type
s := i.(string)        // ✓ s = "hello"
n := i.(int)           // ❌ panics: interface conversion

// Two-return: safe, no panic
s, ok := i.(string)    // ok=true, s="hello"
n, ok := i.(int)       // ok=false, n=0 (zero value)
```

**When to use**: When you have an interface and need to access concrete-type methods not in the interface, or to check which concrete type is stored.

```go
type Animal interface{ Sound() string }
type Dog struct{}
func (d Dog) Sound() string { return "Woof" }
func (d Dog) Fetch() string { return "fetching!" }

var a Animal = Dog{}

// Need Fetch() which is not in Animal interface
if dog, ok := a.(Dog); ok {
    fmt.Println(dog.Fetch()) // ✓
}
```

## Type Switches

A type switch is the idiomatic way to handle multiple concrete types behind an interface.

```go
func describe(i interface{}) string {
    switch v := i.(type) {
    case int:
        return fmt.Sprintf("int: %d", v)
    case string:
        return fmt.Sprintf("string: %q", v)
    case bool:
        return fmt.Sprintf("bool: %v", v)
    case nil:
        return "nil"
    default:
        return fmt.Sprintf("unknown: %T", v)
    }
}
```

**Note**: Inside each `case`, `v` has the concrete type of that case. In `case int`, `v` is an `int`, not `interface{}`.

### Practical: JSON-like dynamic values

```go
type JSONValue interface{}

func printJSON(v JSONValue) {
    switch val := v.(type) {
    case map[string]JSONValue:
        for k, child := range val {
            fmt.Printf("%s: ", k)
            printJSON(child)
        }
    case []JSONValue:
        for _, item := range val {
            printJSON(item)
        }
    case string:
        fmt.Printf("%q\n", val)
    case float64:
        fmt.Printf("%g\n", val)
    case bool:
        fmt.Printf("%v\n", val)
    case nil:
        fmt.Println("null")
    }
}
```

## Comparable Types

Not all types in Go can be compared with `==`. Understanding this matters for map keys, interface comparisons, and generics constraints.

| Comparable                                       | Not Comparable                     |
| ------------------------------------------------ | ---------------------------------- |
| Basic types (`int`, `string`, `bool`, `float64`) | Slices                             |
| Arrays (if element type is comparable)           | Maps                               |
| Structs (if all fields are comparable)           | Functions                          |
| Pointers                                         | Structs with non-comparable fields |
| Interfaces                                       | —                                  |
| Channels                                         | —                                  |

```go
// Interface comparison: compares both type AND value
var a, b interface{} = 42, 42
fmt.Println(a == b) // true

var c, d interface{} = []int{1}, []int{1}
fmt.Println(c == d) // ❌ panics at runtime: slice is not comparable
```

**Gotcha**: Comparing interfaces with non-comparable underlying types panics at runtime, not at compile time. Use `reflect.DeepEqual` for deep equality checks on such types.

```go
// Safe comparison for unknown interface values:
import "reflect"
reflect.DeepEqual(c, d) // true, never panics
```

## The `any` Alias

Since Go 1.18, `any` is an alias for `interface{}`. They are identical — use `any` in new code.

```go
func printAll(values ...any) {
    for _, v := range values {
        fmt.Println(v)
    }
}
```

## Hands-On Exercise 1: Fixing the Nil Interface Bug

The following function has a nil interface trap. Identify and fix it.

```go
type AppError struct {
    Code    int
    Message string
}

func (e *AppError) Error() string {
    return fmt.Sprintf("error %d: %s", e.Code, e.Message)
}

func processRequest(id int) error {
    var appErr *AppError

    if id < 0 {
        appErr = &AppError{Code: 400, Message: "invalid id"}
    }

    // Some processing...
    if id == 0 {
        appErr = &AppError{Code: 404, Message: "not found"}
    }

    return appErr // Bug is here
}

func main() {
    err := processRequest(1) // valid request
    if err != nil {
        fmt.Println("Error:", err) // This prints even for valid requests!
    }
}
```

<details>
<summary>Solution</summary>

**Issue**: `return appErr` always returns a non-nil interface because `appErr` has type `*AppError`, even when its value is `nil`. The interface is `{type: *AppError, value: nil}` which is not equal to a nil interface.

**Fixed**:

```go
func processRequest(id int) error {
    if id < 0 {
        return &AppError{Code: 400, Message: "invalid id"}
    }
    if id == 0 {
        return &AppError{Code: 404, Message: "not found"}
    }
    return nil // ✓ untyped nil = nil interface
}
```

**Rule**: Never return a typed nil through an interface. Return `nil` directly, or return the interface type rather than a concrete type variable.

</details>

## Hands-On Exercise 2: Shape Area Calculator

Implement a shape system using interfaces and a type switch.

```go
// Requirements:
// 1. Define a Shape interface with Area() float64 and Perimeter() float64
// 2. Implement Circle, Rectangle, Triangle
// 3. Write a TotalArea(shapes []Shape) float64 function
// 4. Write a Describe(s Shape) string that uses a type switch to return
//    a human-readable description with type-specific fields
//    e.g. "Circle with radius 5.00" or "Rectangle 3.00x4.00"
```

<details>
<summary>Solution</summary>

```go
package main

import (
    "fmt"
    "math"
)

type Shape interface {
    Area() float64
    Perimeter() float64
}

type Circle struct {
    Radius float64
}

func (c Circle) Area() float64      { return math.Pi * c.Radius * c.Radius }
func (c Circle) Perimeter() float64 { return 2 * math.Pi * c.Radius }

type Rectangle struct {
    Width, Height float64
}

func (r Rectangle) Area() float64      { return r.Width * r.Height }
func (r Rectangle) Perimeter() float64 { return 2 * (r.Width + r.Height) }

type Triangle struct {
    A, B, C float64 // side lengths
}

func (t Triangle) Perimeter() float64 { return t.A + t.B + t.C }
func (t Triangle) Area() float64 {
    s := t.Perimeter() / 2 // Heron's formula
    return math.Sqrt(s * (s - t.A) * (s - t.B) * (s - t.C))
}

func TotalArea(shapes []Shape) float64 {
    total := 0.0
    for _, s := range shapes {
        total += s.Area()
    }
    return total
}

func Describe(s Shape) string {
    switch v := s.(type) {
    case Circle:
        return fmt.Sprintf("Circle with radius %.2f", v.Radius)
    case Rectangle:
        return fmt.Sprintf("Rectangle %.2fx%.2f", v.Width, v.Height)
    case Triangle:
        return fmt.Sprintf("Triangle sides %.2f, %.2f, %.2f", v.A, v.B, v.C)
    default:
        return fmt.Sprintf("Unknown shape: %T", v)
    }
}
```

</details>

## Interview Questions

### Q1: What is the nil interface trap and how do you avoid it?

Interviewers ask this because it's a runtime bug that's hard to spot in code review. It reveals whether you understand Go's interface internals. Many experienced Go developers have been caught by it.

<details>
<summary>Answer</summary>

An interface value in Go has two internal fields: a type pointer and a value pointer. The interface is only `nil` when **both** are nil. If you assign a typed nil pointer to an interface, the type field is set even though the value is nil — making the interface non-nil.

```go
var p *MyError = nil
var err error = p
fmt.Println(err == nil) // false — type is *MyError, value is nil
```

**Avoidance strategies**:

1. Return `nil` directly from functions that return interface types — don't return a typed nil variable
2. If you must build the error conditionally, use early returns rather than accumulating into a typed variable
3. Use `errors.As` to inspect concrete types safely rather than type-asserting directly

</details>

### Q2: Why does Go use structural typing (duck typing) instead of explicit interface declarations?

This question tests whether you understand the design philosophy behind Go, not just its syntax.

<details>
<summary>Answer</summary>

Structural typing enables **retroactive interface satisfaction**: a type can satisfy an interface defined after the type was written, with no changes to the type. This means:

- You can define small, focused interfaces at the point of use rather than in a central hierarchy
- Third-party types can satisfy your interfaces without modification
- Dependencies flow inward (your code depends on the interface, not on the concrete type)

The practical result is that Go naturally follows the Interface Segregation Principle and Dependency Inversion Principle without any extra ceremony. Compare this to Java where you must modify a class to add `implements Stringer` — that's a coupling from the concrete type to the interface.

The `io.Reader` / `io.Writer` interfaces are the canonical example: countless types from different packages (files, buffers, network connections, test mocks) satisfy them without any coordination.

</details>

### Q3: When should you use a value receiver vs a pointer receiver, and how does it affect interface satisfaction?

A classic Go question that trips up developers coming from other languages.

<details>
<summary>Answer</summary>

**Value receiver** (`func (t T) Method()`):

- Method operates on a copy — mutations don't affect the original
- Both `T` and `*T` satisfy interfaces requiring this method
- Use when the method doesn't need to modify the receiver and T is small/copyable

**Pointer receiver** (`func (t *T) Method()`):

- Method can modify the receiver
- Only `*T` satisfies interfaces requiring this method
- Use when the method mutates state, T is large (avoid copying), or T contains a mutex/sync primitive

**Interface satisfaction rule**: If any method in an interface requires a pointer receiver, only a pointer to the concrete type satisfies the interface. You cannot use a value.

```go
type Incrementer interface {
    Inc()
}

type Counter struct{ n int }
func (c *Counter) Inc() { c.n++ } // pointer receiver

var i Incrementer = &Counter{} // ✓
var i2 Incrementer = Counter{} // ❌ compile error
```

**Practical advice**: Be consistent within a type. If any method needs a pointer receiver (e.g., for mutation), make all methods pointer receivers. Mixed receiver types on the same type cause confusion.

</details>

### Q4: What is interface embedding and when is it useful?

Tests whether you understand Go's composition model and can design clean APIs.

<details>
<summary>Answer</summary>

Interfaces can embed other interfaces, combining their method sets. The embedding interface is satisfied by any type that satisfies all embedded interfaces.

```go
type Reader interface { Read(p []byte) (n int, err error) }
type Closer interface { Close() error }
type ReadCloser interface {
    Reader
    Closer
}
```

**When it's useful**:

1. **Building composed interfaces from small ones**: Start with `io.Reader`, `io.Writer`, `io.Closer` and compose them as needed (`io.ReadCloser`, `io.ReadWriteCloser`). Code that only needs reading takes `io.Reader`; code that needs both takes `io.ReadWriteCloser`.

2. **API design**: Define the minimal interface your function needs. Accept `io.Reader` not `*os.File`. This makes the function testable with any reader (buffer, string reader, mock).

3. **Struct embedding of interfaces**: You can embed an interface in a struct to create partial implementations — useful in testing where you want to implement only a few methods of a large interface.

```go
type BigInterface interface {
    A(); B(); C(); D(); E()
}

// In tests, only implement what you need:
type partialMock struct {
    BigInterface // provides nil implementations of A,B,C,D,E
}
func (m partialMock) A() { /* real implementation */ }
```

</details>

## Key Takeaways

1. **Implicit satisfaction**: Go types satisfy interfaces by having the required methods — no `implements` keyword.
2. **Pointer receivers**: Only pointer types satisfy interfaces that require pointer-receiver methods.
3. **Nil interface trap**: An interface holding a typed nil is not nil — both type and value must be nil.
4. **Composition via embedding**: Go achieves polymorphism through struct and interface embedding, not inheritance.
5. **Type assertions**: Extract concrete types from interfaces with `v, ok := i.(T)` — always use the two-value form to avoid panics.
6. **Type switches**: The idiomatic way to branch on concrete type; inside each case, the variable has the concrete type.
7. **Comparable types**: Interfaces with non-comparable underlying types panic when compared at runtime — not at compile time.
8. **`any`**: Alias for `interface{}` since Go 1.18 — prefer it in new code.
9. **Small interfaces**: Design interfaces with the fewest methods needed. The `io` package's one-method interfaces are the model.

## Next Steps

In [Lesson 02: Error Handling Patterns](lesson-02-error-handling-patterns.md), you'll learn:

- How `errors.Is` and `errors.As` traverse the error chain
- The difference between sentinel errors and custom error types
- When (and when not) to use `panic` and `recover`
- Wrapping errors with `%w` and the implications for callers
