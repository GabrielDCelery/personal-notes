# Lesson 02: Error Handling Patterns

Go's error handling is explicit and intentional — errors are values, not exceptions. This design forces you to think about failure at every step. But the real interview depth is in the error _chain_: how modern Go lets callers inspect wrapped errors, and how to design error types that provide useful context without leaking internals.

## Errors Are Values

In Go, `error` is just an interface:

```go
type error interface {
    Error() string
}
```

Any type with an `Error() string` method satisfies it. This simplicity means:

- Errors can carry structured data (not just messages)
- Errors can be compared, wrapped, stored in maps, passed to goroutines
- Standard control flow (`if err != nil`) handles them — no exception unwinding

```go
f, err := os.Open("file.txt")
if err != nil {
    return fmt.Errorf("opening file: %w", err) // wrap with context
}
defer f.Close()
```

## Sentinel Errors

A **sentinel error** is a package-level error variable that callers compare against. They are the oldest pattern and still common in the standard library.

```go
// Standard library examples
var (
    io.EOF          = errors.New("EOF")
    os.ErrNotExist  = errors.New("file does not exist")
    sql.ErrNoRows   = errors.New("sql: no rows in result set")
)

// Usage
data, err := io.ReadAll(r)
if err == io.EOF {  // ❌ fragile - breaks if error is wrapped
    // done
}
if errors.Is(err, io.EOF) { // ✓ works even through wrapping
    // done
}
```

**When to use sentinel errors**:

- Simple, fixed conditions that callers need to detect
- Public API contracts where callers branch on specific errors
- No additional context needed beyond the error identity

**Gotcha**: Never compare sentinel errors with `==` — use `errors.Is`. A wrapped sentinel won't match `==` but will match `errors.Is`.

## Custom Error Types

When callers need structured data from an error (e.g., which field failed, what HTTP status to return), create a custom type:

```go
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed on %s: %s", e.Field, e.Message)
}

// Caller extracts structured data:
var valErr *ValidationError
if errors.As(err, &valErr) {
    fmt.Printf("Field %q: %s\n", valErr.Field, valErr.Message)
}
```

### Error Types vs Sentinel Errors

|                  | Sentinel Error             | Custom Error Type                      |
| ---------------- | -------------------------- | -------------------------------------- |
| **Comparison**   | `errors.Is`                | `errors.As`                            |
| **Carries data** | No                         | Yes                                    |
| **Use when**     | Identity matters           | Data matters                           |
| **Example**      | `io.EOF`, `os.ErrNotExist` | HTTP status codes, validation failures |

## Error Wrapping with `%w`

Before Go 1.13, errors were opaque once wrapped: `fmt.Errorf("context: %v", err)` produced a new error with no link to the original. Go 1.13 introduced `%w` to preserve the chain.

```go
// ❌ Wraps but loses original error - callers can't inspect it
return fmt.Errorf("loading config: %v", err)

// ✓ Wraps AND preserves - callers can use errors.Is/As
return fmt.Errorf("loading config: %w", err)
```

The `%w` verb calls `errors.New` under the hood and stores a reference to the wrapped error. The resulting error's `Unwrap()` method returns the original.

### `errors.Is` — Identity Check

`errors.Is` traverses the error chain looking for a match:

```go
// errors.Is checks: is this error, or any error it wraps, equal to target?
func Is(err, target error) bool

// Example chain: ErrPermission wraps ErrNotExist wraps ErrInternal
errors.Is(err, ErrInternal)    // true - found at depth 2
errors.Is(err, io.EOF)         // false - not in chain
```

You can implement `Is` on your error type for custom matching logic:

```go
type HTTPError struct {
    Code int
}

func (e *HTTPError) Is(target error) bool {
    t, ok := target.(*HTTPError)
    if !ok {
        return false
    }
    return e.Code == t.Code || t.Code == 0 // 0 means "any HTTP error"
}

// Usage:
errors.Is(err, &HTTPError{Code: 404}) // matches only 404
errors.Is(err, &HTTPError{})          // matches any HTTPError
```

### `errors.As` — Type Extraction

`errors.As` traverses the chain looking for a type match, and if found, sets the target pointer:

```go
// errors.As checks: is any error in the chain assignable to target?
func As(err error, target any) bool

var netErr *net.OpError
if errors.As(err, &netErr) {
    fmt.Println("network error:", netErr.Op, netErr.Net)
}
```

### `errors.Unwrap` — Manual Chain Traversal

```go
// errors.Unwrap returns the next error in the chain, or nil
wrapped := fmt.Errorf("outer: %w", inner)
errors.Unwrap(wrapped) // returns inner
```

You rarely call `Unwrap` directly — use `errors.Is` and `errors.As` instead. They call `Unwrap` internally and traverse the full chain.

## Building an Error Chain

```go
// Low-level layer
func readConfig(path string) error {
    _, err := os.ReadFile(path)
    if err != nil {
        return fmt.Errorf("readConfig %q: %w", path, err) // wrap with context
    }
    return nil
}

// Mid-level layer
func loadApp(configPath string) error {
    if err := readConfig(configPath); err != nil {
        return fmt.Errorf("loadApp: %w", err) // wrap again
    }
    return nil
}

// Caller
err := loadApp("/etc/app.yaml")
// err chain: "loadApp: readConfig \"/etc/app.yaml\": open /etc/app.yaml: no such file or directory"

errors.Is(err, os.ErrNotExist) // true - os.ErrNotExist is at the bottom of the chain
```

## `panic` and `recover`

Go's `panic` is not for ordinary error handling. It's for **truly unrecoverable situations** — programming bugs, impossible states.

```go
// ✓ Appropriate panic use cases
func MustCompile(expr string) *regexp.Regexp {
    re, err := regexp.Compile(expr)
    if err != nil {
        panic(err) // called with a literal string - if it panics, that's a bug
    }
    return re
}

// ❌ Wrong - panicking on user/network errors
func readFile(path string) []byte {
    data, err := os.ReadFile(path)
    if err != nil {
        panic(err) // wrong: file not existing is not a bug
    }
    return data
}
```

### `recover`

`recover` stops a panic and returns the panic value. It only works **inside a deferred function**.

```go
func safeDiv(a, b int) (result int, err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("recovered from panic: %v", r)
        }
    }()
    return a / b, nil // panics if b == 0
}
```

**When to use `recover`**:

- At the boundary between your code and external/untrusted code (e.g., executing a plugin, running user-supplied functions)
- In HTTP handlers to prevent one bad request from crashing the server
- In library code that calls into user-supplied callbacks

**Never use `panic`/`recover` as a substitute for error returns within your own package.**

### HTTP middleware example (the only common legitimate use)

```go
func recoverMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("panic: %v\n%s", err, debug.Stack())
                http.Error(w, "Internal Server Error", 500)
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```

## When NOT to Panic

| Situation                      | Use panic?    | Why                              |
| ------------------------------ | ------------- | -------------------------------- |
| File not found                 | ❌            | Expected condition, return error |
| Network timeout                | ❌            | Expected condition, return error |
| Nil pointer in your code       | ✓             | Programming bug                  |
| Index out of bounds            | ✓             | Programming bug                  |
| Invalid argument to public API | ✓ (sometimes) | Document it clearly              |
| Unreachable code path          | ✓             | `panic("unreachable")`           |

## Hands-On Exercise 1: Error Chain Inspector

Write a function that describes the full error chain and checks for specific error types.

```go
// Given this setup:
type DBError struct {
    Query string
    Err   error
}
func (e *DBError) Error() string { return fmt.Sprintf("db error on %q: %v", e.Query, e.Err) }
func (e *DBError) Unwrap() error { return e.Err }

var ErrConnectionFailed = errors.New("connection failed")

// Write:
// 1. A function that builds this chain:
//    "service layer" wraps DBError wraps ErrConnectionFailed
// 2. A function DescribeChain(err error) that prints each error in the chain
// 3. Demonstrate that errors.Is(err, ErrConnectionFailed) returns true
```

<details>
<summary>Solution</summary>

```go
package main

import (
    "errors"
    "fmt"
)

type DBError struct {
    Query string
    Err   error
}

func (e *DBError) Error() string { return fmt.Sprintf("db error on %q: %v", e.Query, e.Err) }
func (e *DBError) Unwrap() error { return e.Err }

var ErrConnectionFailed = errors.New("connection failed")

func buildChain() error {
    base := ErrConnectionFailed
    dbErr := &DBError{Query: "SELECT * FROM users", Err: base}
    return fmt.Errorf("service layer: %w", dbErr)
}

func DescribeChain(err error) {
    depth := 0
    for err != nil {
        fmt.Printf("%s[%d] %T: %v\n", spaces(depth), depth, err, err)
        err = errors.Unwrap(err)
        depth++
    }
}

func spaces(n int) string {
    s := ""
    for i := 0; i < n*2; i++ {
        s += " "
    }
    return s
}

func main() {
    err := buildChain()
    DescribeChain(err)
    fmt.Println("Is ErrConnectionFailed:", errors.Is(err, ErrConnectionFailed)) // true

    var dbErr *DBError
    if errors.As(err, &dbErr) {
        fmt.Println("DB query was:", dbErr.Query)
    }
}
```

</details>

## Hands-On Exercise 2: Retry with Error Classification

Implement a retry function that distinguishes between retryable and permanent errors.

```go
// Requirements:
// 1. Define a RetryableError type that wraps another error and marks it as retryable
// 2. Define a sentinel: var ErrPermanent = errors.New("permanent error")
// 3. Implement Retry(attempts int, fn func() error) error that:
//    - Retries only if errors.As(err, &RetryableError{}) matches
//    - Returns immediately on permanent/non-retryable errors
//    - Returns the last error if all attempts fail
// Test with a function that fails twice then succeeds
```

<details>
<summary>Solution</summary>

```go
package main

import (
    "errors"
    "fmt"
)

type RetryableError struct {
    Err error
}

func (e *RetryableError) Error() string { return fmt.Sprintf("retryable: %v", e.Err) }
func (e *RetryableError) Unwrap() error { return e.Err }

var ErrPermanent = errors.New("permanent error")

func Retry(attempts int, fn func() error) error {
    var lastErr error
    for i := 0; i < attempts; i++ {
        lastErr = fn()
        if lastErr == nil {
            return nil
        }
        var retryable *RetryableError
        if !errors.As(lastErr, &retryable) {
            return lastErr // non-retryable: return immediately
        }
        fmt.Printf("attempt %d failed (retryable): %v\n", i+1, lastErr)
    }
    return fmt.Errorf("all %d attempts failed: %w", attempts, lastErr)
}

func main() {
    count := 0
    err := Retry(5, func() error {
        count++
        if count < 3 {
            return &RetryableError{Err: fmt.Errorf("temporary network hiccup")}
        }
        return nil // succeed on attempt 3
    })
    fmt.Println("final error:", err) // nil
    fmt.Println("attempts:", count)  // 3

    // Permanent error stops immediately
    err = Retry(5, func() error {
        return ErrPermanent
    })
    fmt.Println("permanent:", err) // returns after 1 attempt
}
```

</details>

## Interview Questions

### Q1: What is the difference between `errors.Is` and `errors.As`?

Asked to test whether you understand error chains and modern Go error handling (post-1.13). Candidates who still use `==` or type assertions directly are working with outdated patterns.

<details>
<summary>Answer</summary>

- `errors.Is(err, target)` checks **identity**: is `target` anywhere in the error chain? Uses `==` for comparison, or a custom `Is(error) bool` method if defined. Use this to check for sentinel errors.

- `errors.As(err, &target)` checks **type**: is any error in the chain assignable to the type of `target`? If found, sets `target` to that error value. Use this to extract structured data from a specific error type.

Both traverse the chain by calling `Unwrap()` repeatedly. `errors.Is` and `errors.As` replaced direct `==` comparisons and type assertions because those break when errors are wrapped.

```go
// Sentinel: use errors.Is
errors.Is(err, os.ErrNotExist)

// Custom type: use errors.As
var pathErr *os.PathError
errors.As(err, &pathErr)
fmt.Println(pathErr.Path) // access structured data
```

</details>

### Q2: When should you use `panic` vs returning an error?

A question that separates developers who understand Go's design philosophy from those who treat panic like exceptions.

<details>
<summary>Answer</summary>

**Return an error** when:

- The failure is a normal operational condition (file not found, network timeout, validation failure)
- The caller might reasonably handle and recover from the failure
- The situation is caused by external inputs or environment

**Use `panic`** when:

- The situation represents a programming bug that should never happen in correct code
- An API contract is violated (e.g., calling a function with a nil argument that must not be nil)
- The program cannot safely continue and there is no meaningful recovery
- In `init()` or package-level setup where startup errors should terminate the program

**`recover`** is appropriate at:

- HTTP server handler boundaries (prevent one bad request from crashing the server)
- Goroutine boundaries when running untrusted code
- Public library APIs that call user-supplied callbacks

The `Must*` pattern is idiomatic for wrapping functions that should only be called with valid inputs at program startup:

```go
var compiled = regexp.MustCompile(`^[a-z]+$`) // panics at startup if regex is invalid
```

</details>

### Q3: How do you design an error hierarchy for a large application?

Tests API design skills and understanding of how callers interact with errors.

<details>
<summary>Answer</summary>

Good error design follows these principles:

**1. Layer errors at package boundaries**: Each package wraps errors with its own context. Don't let `*os.PathError` leak from your HTTP handler — wrap it with a domain error.

**2. Define sentinel errors for conditions callers branch on**:

```go
var ErrNotFound = errors.New("not found")
var ErrUnauthorized = errors.New("unauthorized")
```

**3. Use structured types for errors that need data**:

```go
type ValidationError struct {
    Field   string
    Message string
}
```

**4. Implement `Is` for custom matching if needed** (e.g., matching by HTTP status code range).

**5. Wrap, don't replace**: Use `%w` so the chain is preserved for callers who want to inspect it.

**6. Don't expose implementation details**: Return `ErrNotFound` rather than `*sql.ErrNoRows` from your repository layer — the caller shouldn't need to know you use SQL.

```go
// Repository layer
func (r *UserRepo) FindByID(id int) (*User, error) {
    u, err := r.db.QueryRow(...)
    if errors.Is(err, sql.ErrNoRows) {
        return nil, ErrNotFound // translate implementation error to domain error
    }
    if err != nil {
        return nil, fmt.Errorf("findByID %d: %w", id, err)
    }
    return u, nil
}
```

</details>

### Q4: What is the difference between `fmt.Errorf("msg: %v", err)` and `fmt.Errorf("msg: %w", err)`?

A concise question testing knowledge of Go 1.13 error wrapping.

<details>
<summary>Answer</summary>

- `%v` converts the error to a string and includes it in the message. The resulting error has **no reference** to the original — `errors.Is` and `errors.As` cannot traverse to it. The original error is lost.

- `%w` wraps the error, producing a new error that implements `Unwrap() error` returning the original. The chain is preserved, so `errors.Is` and `errors.As` can inspect the original.

```go
err := os.ErrNotExist

// With %v - chain is broken
wrapped := fmt.Errorf("context: %v", err)
errors.Is(wrapped, os.ErrNotExist) // false - original is lost

// With %w - chain is preserved
wrapped = fmt.Errorf("context: %w", err)
errors.Is(wrapped, os.ErrNotExist) // true - Unwrap finds it
```

**Rule**: Always use `%w` unless you explicitly want to hide the original error from callers (e.g., to prevent implementation details from leaking in a public API). Even then, consider returning a domain error sentinel instead.

</details>

## Key Takeaways

1. **Errors are values**: The `error` interface is just `Error() string` — any type satisfying it can carry structured data.
2. **`errors.Is` for identity**: Use it for sentinel errors; it traverses the chain unlike `==`.
3. **`errors.As` for type**: Use it to extract structured data from a specific error type in the chain.
4. **`%w` for wrapping**: Use `fmt.Errorf("context: %w", err)` to preserve the error chain for callers.
5. **Sentinel vs struct**: Sentinels for identity, custom structs for data — choose based on what callers need.
6. **Panic is for bugs**: Not for user errors, network failures, or any expected failure mode.
7. **Recover at boundaries**: HTTP handlers and goroutine launchers are the legitimate places for `recover`.
8. **Layer your errors**: Each package translates lower-level errors to domain errors. Don't let `sql.ErrNoRows` leak from your HTTP handler.
9. **Never expose implementation**: Return `ErrNotFound` not `*pq.Error` — callers shouldn't know your storage technology.

## Next Steps

In [Lesson 03: Goroutines & the Go Scheduler](lesson-03-goroutines-and-scheduler.md), you'll learn:

- How Go's M:N threading model works and what `GOMAXPROCS` controls
- The goroutine lifecycle and what happens when goroutines leak
- Work stealing and why it matters for performance
- Practical techniques to detect and fix goroutine leaks
