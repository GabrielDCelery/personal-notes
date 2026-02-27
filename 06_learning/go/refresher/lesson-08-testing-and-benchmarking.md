# Lesson 08: Testing & Benchmarking

Go's testing toolchain is built into the language, not bolted on as an afterthought. The patterns — table-driven tests, subtests, interface-based mocking — are idiomatic to Go and will come up in any senior interview. Equally important is the performance side: knowing how to write a benchmark that measures what you think it measures, and how to profile a real service to find actual bottlenecks.

## The `testing` Package

Every Go test is a function in a `_test.go` file:

```go
// math_test.go
package math_test  // black-box testing (can also use package math for white-box)

import (
    "testing"
    "your/package/math"
)

func TestAdd(t *testing.T) {
    result := math.Add(2, 3)
    if result != 5 {
        t.Errorf("Add(2,3) = %d; want 5", result)
    }
}
```

```sh
go test ./...                    # run all tests
go test -v ./...                 # verbose output
go test -run TestAdd ./...       # run specific test by name (regex)
go test -count=1 ./...           # disable test result caching
go test -timeout 30s ./...       # test timeout
```

## Table-Driven Tests

The idiomatic Go pattern for testing multiple inputs. Instead of writing many similar test functions, define a slice of test cases and loop over them.

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"both positive", 2, 3, 5},
        {"negative + positive", -1, 1, 0},
        {"both negative", -2, -3, -5},
        {"zeros", 0, 0, 0},
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            result := Add(tc.a, tc.b)
            if result != tc.expected {
                t.Errorf("Add(%d, %d) = %d; want %d", tc.a, tc.b, result, tc.expected)
            }
        })
    }
}
```

**Why table-driven tests?**

- Adding a new test case is one line
- Failed test output identifies exactly which case failed: `--- FAIL: TestAdd/both_negative`
- No duplication of test structure

## Subtests: `t.Run`

`t.Run` creates a named subtest. It enables:

- Running specific subtests: `go test -run TestAdd/both_negative`
- Parallel subtests
- Clear failure messages

```go
func TestParse(t *testing.T) {
    t.Run("valid input", func(t *testing.T) {
        v, err := Parse("42")
        if err != nil { t.Fatal(err) }
        if v != 42 { t.Errorf("got %d; want 42", v) }
    })

    t.Run("invalid input", func(t *testing.T) {
        _, err := Parse("abc")
        if err == nil { t.Error("expected error for invalid input") }
    })
}
```

### Parallel Subtests

```go
func TestHTTPEndpoints(t *testing.T) {
    tests := []struct{ name, path string }{ ... }

    for _, tc := range tests {
        tc := tc // ❌ before Go 1.22: capture loop variable
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel() // run subtests concurrently
            // test tc.path
        })
    }
}
```

**Note**: In Go 1.22+, loop variables are no longer shared — `tc := tc` is no longer needed. In older versions, it's required to avoid the closure capturing the same variable.

## Test Helpers: `t.Helper()`

When you extract test logic into a helper function, call `t.Helper()` so failures report the caller's line, not the helper's line.

```go
// ❌ Without t.Helper(): error points to assertNil, not the test case
func assertNil(t *testing.T, err error) {
    if err != nil {
        t.Fatalf("unexpected error: %v", err) // line 10: points here
    }
}

// ✓ With t.Helper(): error points to TestSomething where assertNil was called
func assertNil(t *testing.T, err error) {
    t.Helper() // marks this function as a test helper
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}

func TestSomething(t *testing.T) {
    err := doSomething()
    assertNil(t, err) // failure reported here, not inside assertNil
}
```

## `t.Fatal` vs `t.Error`

| Method                | Behaviour                                       |
| --------------------- | ----------------------------------------------- |
| `t.Error(msg)`        | Mark test failed, continue execution            |
| `t.Errorf(fmt, args)` | Same with formatting                            |
| `t.Fatal(msg)`        | Mark test failed, stop current test immediately |
| `t.Fatalf(fmt, args)` | Same with formatting                            |

```go
// Use Fatal when subsequent code depends on the preceding check
result, err := DoThing()
if err != nil {
    t.Fatalf("DoThing failed: %v", err) // stop here; result is unusable
}
// Only reaches here if err == nil
if result.Value != 42 {
    t.Errorf("got %d; want 42", result.Value)
}
```

## Mocking via Interfaces

Go's testing approach to mocking uses interfaces and handwritten fakes — not reflection-based mock frameworks (though `testify/mock` is common in industry).

```go
// Production code depends on interface, not concrete type
type EmailSender interface {
    Send(to, subject, body string) error
}

type UserService struct {
    email EmailSender
}

func (s *UserService) Register(email string) error {
    // ... create user ...
    return s.email.Send(email, "Welcome!", "Thanks for signing up")
}

// Test fake - handwritten
type fakeEmailSender struct {
    sent []struct{ to, subject, body string }
    err  error
}

func (f *fakeEmailSender) Send(to, subject, body string) error {
    f.sent = append(f.sent, struct{ to, subject, body string }{to, subject, body})
    return f.err
}

func TestUserService_Register(t *testing.T) {
    sender := &fakeEmailSender{}
    svc := &UserService{email: sender}

    err := svc.Register("alice@example.com")
    if err != nil { t.Fatal(err) }

    if len(sender.sent) != 1 {
        t.Fatalf("expected 1 email sent; got %d", len(sender.sent))
    }
    if sender.sent[0].to != "alice@example.com" {
        t.Errorf("email sent to wrong address")
    }
}
```

### Using `testify`

[testify](https://github.com/stretchr/testify) is the most popular third-party test library:

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestSomething(t *testing.T) {
    result, err := DoSomething()

    require.NoError(t, err)           // like t.Fatal: stops on failure
    assert.Equal(t, 42, result.Value) // like t.Error: continues on failure
    assert.Contains(t, result.Tags, "important")
    assert.Len(t, result.Items, 3)
}
```

`require` vs `assert`: `require` stops the test on failure (like `Fatal`); `assert` continues (like `Error`).

## The Race Detector

```sh
go test -race ./...
go run -race main.go
```

The race detector instruments memory accesses and detects concurrent reads/writes without synchronization. It adds ~5-10x overhead — use in CI, not production.

```go
// This data race will be caught:
func TestRace(t *testing.T) {
    m := make(map[string]int)
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(n int) {
            defer wg.Done()
            m[fmt.Sprintf("key%d", n)] = n // DATA RACE: concurrent map writes
        }(i)
    }
    wg.Wait()
}
```

Race detector output:

```
==================
WARNING: DATA RACE
Write at 0x... by goroutine 8:
  TestRace.func1()
      race_test.go:12
Previous write at 0x... by goroutine 7:
  TestRace.func1()
      race_test.go:12
==================
```

**Always run tests with `-race` in CI.** Many race conditions don't trigger on every run — the race detector's instrumentation catches them reliably.

## Writing Benchmarks

Benchmarks measure performance — execution time and allocations per operation.

```go
func BenchmarkAdd(b *testing.B) {
    for i := 0; i < b.N; i++ { // b.N is set by the testing framework
        Add(1000000, 2000000)
    }
}
```

```sh
go test -bench=. ./...              # run all benchmarks
go test -bench=BenchmarkAdd ./...   # specific benchmark
go test -bench=. -benchmem ./...    # include memory allocation stats
go test -bench=. -count=5 ./...     # run each benchmark 5 times
go test -bench=. -benchtime=10s ./... # run for 10 seconds
```

### Reading Benchmark Output

```
BenchmarkAdd-8        1000000000    0.29 ns/op
BenchmarkJSON/encode-8  2000000     650 ns/op    256 B/op    2 allocs/op
```

| Column        | Meaning                                                |
| ------------- | ------------------------------------------------------ |
| `-8`          | GOMAXPROCS value                                       |
| `1000000000`  | Number of iterations (`b.N`)                           |
| `0.29 ns/op`  | Average time per operation                             |
| `256 B/op`    | Bytes allocated per operation (with `-benchmem`)       |
| `2 allocs/op` | Number of allocations per operation (with `-benchmem`) |

### Benchmark Setup Code

Code before and after the timer should be excluded:

```go
func BenchmarkJSON(b *testing.B) {
    data := generateLargeStruct() // setup - not part of benchmark

    b.ResetTimer() // start timing after setup
    for i := 0; i < b.N; i++ {
        json.Marshal(data)
    }
}

// For benchmarks with per-iteration setup:
func BenchmarkDB(b *testing.B) {
    for i := 0; i < b.N; i++ {
        b.StopTimer()
        setupDB() // not benchmarked
        b.StartTimer()
        queryDB() // benchmarked
    }
}
```

### Preventing Compiler Optimization

The compiler can eliminate computations whose results are unused. Prevent this with a sink variable:

```go
var result int // package-level sink

func BenchmarkCompute(b *testing.B) {
    var r int
    for i := 0; i < b.N; i++ {
        r = ExpensiveCompute(i) // assign to local
    }
    result = r // assign to package var, preventing dead code elimination
}
```

## Profiling with `pprof`

`pprof` is Go's built-in profiling tool. Two ways to use it:

### Option 1: Profile Tests/Benchmarks

```sh
# CPU profile
go test -bench=. -cpuprofile=cpu.prof ./...
go tool pprof cpu.prof

# Memory profile
go test -bench=. -memprofile=mem.prof ./...
go tool pprof mem.prof
```

### Option 2: Profile a Running Server

```go
import _ "net/http/pprof" // registers /debug/pprof/* endpoints

func main() {
    go http.ListenAndServe(":6060", nil) // profiling server
    // your actual server
}
```

```sh
# Capture 30s CPU profile from running server:
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Heap profile:
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine profile:
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

### Reading pprof Output

```
(pprof) top10         # top 10 functions by CPU time
(pprof) list funcName # annotated source for a function
(pprof) web           # open interactive graph in browser (requires graphviz)
```

```
Showing nodes accounting for 2.13s, 91.01% of 2.34s total
  flat  flat%   sum%   cum   cum%
 1.05s 44.87% 44.87%  1.05s 44.87%  runtime.mallocgc
 0.42s 17.95% 62.82%  0.42s 17.95%  encoding/json.(*encodeState).marshal
```

- `flat`: time spent in this function itself (not its callees)
- `cum`: time spent in this function AND its callees

High `runtime.mallocgc` means allocation pressure — look for escaping variables or unnecessary heap allocations.

## Test Organization

```
mypackage/
├── user.go
├── user_test.go           # white-box tests (same package)
└── user_integration_test.go  # black-box tests (package mypackage_test)
```

```go
// White-box test: same package, can access unexported identifiers
package mypackage

func TestInternalHelper(t *testing.T) {
    result := internalHelper() // can access unexported function
}

// Black-box test: _test suffix, can only use exported API
package mypackage_test

func TestPublicAPI(t *testing.T) {
    result := mypackage.PublicFunction()
}
```

## Hands-On Exercise 1: Table-Driven Test with Error Cases

Write table-driven tests for a URL validator function.

```go
// Function to test:
func ValidateURL(rawURL string) error // returns nil for valid URLs

// Requirements:
// 1. Test at least 8 cases: valid URLs, missing scheme, empty, whitespace-only,
//    localhost, IP address, URL with path/query, invalid characters
// 2. Use t.Run for each case
// 3. Verify that valid URLs return nil and invalid URLs return non-nil error
```

<details>
<summary>Solution</summary>

```go
package main

import (
    "errors"
    "net/url"
    "strings"
    "testing"
)

func ValidateURL(rawURL string) error {
    rawURL = strings.TrimSpace(rawURL)
    if rawURL == "" {
        return errors.New("empty URL")
    }
    u, err := url.Parse(rawURL)
    if err != nil {
        return err
    }
    if u.Scheme == "" {
        return errors.New("missing scheme")
    }
    if u.Scheme != "http" && u.Scheme != "https" {
        return errors.New("scheme must be http or https")
    }
    if u.Host == "" {
        return errors.New("missing host")
    }
    return nil
}

func TestValidateURL(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid http", "http://example.com", false},
        {"valid https with path", "https://example.com/path?q=1", false},
        {"valid localhost", "http://localhost:8080", false},
        {"valid IP", "http://192.168.1.1", false},
        {"empty string", "", true},
        {"whitespace only", "   ", true},
        {"missing scheme", "example.com", true},
        {"invalid scheme", "ftp://example.com", true},
        {"scheme only", "https://", true},
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            err := ValidateURL(tc.input)
            if tc.wantErr && err == nil {
                t.Errorf("ValidateURL(%q): expected error, got nil", tc.input)
            }
            if !tc.wantErr && err != nil {
                t.Errorf("ValidateURL(%q): unexpected error: %v", tc.input, err)
            }
        })
    }
}
```

</details>

## Hands-On Exercise 2: Benchmark Two Implementations

Benchmark two implementations of string concatenation and analyze which is faster and why.

```go
// Implementation 1: naive concatenation in a loop
func concatPlus(parts []string) string {
    result := ""
    for _, p := range parts {
        result += p
    }
    return result
}

// Implementation 2: strings.Builder
func concatBuilder(parts []string) string {
    var b strings.Builder
    for _, p := range parts {
        b.WriteString(p)
    }
    return b.String()
}

// Write benchmarks for both with 100 strings of 100 chars each.
// Use -benchmem. Explain the results.
```

<details>
<summary>Solution</summary>

```go
package main

import (
    "strings"
    "testing"
)

var parts []string

func init() {
    for i := 0; i < 100; i++ {
        parts = append(parts, strings.Repeat("x", 100))
    }
}

func BenchmarkConcatPlus(b *testing.B) {
    var r string
    for i := 0; i < b.N; i++ {
        r = concatPlus(parts)
    }
    _ = r
}

func BenchmarkConcatBuilder(b *testing.B) {
    var r string
    for i := 0; i < b.N; i++ {
        r = concatBuilder(parts)
    }
    _ = r
}
```

**Expected output (`go test -bench=. -benchmem`)**:

```
BenchmarkConcatPlus-8       50000    25000 ns/op    530000 B/op    99 allocs/op
BenchmarkConcatBuilder-8   500000     2500 ns/op     12288 B/op     1 allocs/op
```

**Explanation**: `+=` on strings creates a new string on each iteration (strings are immutable in Go). 100 concatenations = 99 intermediate allocations growing from 100 bytes to 10,000 bytes. `strings.Builder` uses an internal `[]byte` that grows geometrically, requiring ~1 allocation total and a single final copy for the `String()` call.

The lesson: string concatenation in a loop is O(n²) in both time and allocations. Always use `strings.Builder` or `bytes.Buffer`.

</details>

## Interview Questions

### Q1: What is the table-driven test pattern and why is it preferred?

A straightforward question — tests whether your test habits are idiomatic Go.

<details>
<summary>Answer</summary>

Table-driven tests define test cases as a slice of structs and loop over them with subtests. Each struct contains inputs and expected outputs.

```go
tests := []struct {
    name     string
    input    int
    expected string
}{
    {"zero", 0, "zero"},
    {"positive", 5, "five"},
}
for _, tc := range tests {
    t.Run(tc.name, func(t *testing.T) {
        got := numToWord(tc.input)
        if got != tc.expected {
            t.Errorf("got %q; want %q", got, tc.expected)
        }
    })
}
```

**Why it's preferred**:

1. **Adding cases is trivial** — one line per case
2. **Clear failure messages** — `--- FAIL: TestFoo/positive` pinpoints the case
3. **No code duplication** — test logic is written once
4. **Selective runs** — `go test -run TestFoo/positive` runs a specific case
5. **Easy code review** — reviewers can see all cases at a glance

The alternative (separate `TestFooZero`, `TestFooPositive` functions) leads to duplicated structure and makes adding cases burdensome.

</details>

### Q2: How does Go test mocking work without a mock framework?

Tests understanding of Go's interface-based design and how it enables testability.

<details>
<summary>Answer</summary>

Go's mocking approach relies on interfaces. The production code depends on an interface (not a concrete type). In tests, you provide a fake implementation of that interface.

```go
// 1. Define the interface (or use an existing one)
type Database interface {
    GetUser(id int) (*User, error)
}

// 2. Production code depends on the interface
type Service struct{ db Database }

// 3. In tests, implement the interface with a fake
type fakeDB struct{ users map[int]*User }
func (f *fakeDB) GetUser(id int) (*User, error) {
    u, ok := f.users[id]
    if !ok { return nil, errors.New("not found") }
    return u, nil
}

// 4. Inject the fake in tests
svc := &Service{db: &fakeDB{users: map[int]*User{1: {Name: "Alice"}}}}
```

**Advantages over reflection-based mocks**:

- No runtime panics from type mismatches
- Compiler catches missing method implementations
- Clear, readable test code
- No mock framework dependency

**When `testify/mock` is appropriate**: When the interface has many methods and you only want to mock a few for a specific test. `testify/mock` reduces boilerplate in that case.

**The `embed interface` trick** for large interfaces in tests:

```go
type partialMock struct {
    LargeInterface // provides nil implementations for all methods
}
func (m *partialMock) TheOneMethodICarAbout() { /* implement */ }
```

</details>

### Q3: How do you write a reliable benchmark in Go, and what pitfalls should you avoid?

Tests practical benchmarking knowledge.

<details>
<summary>Answer</summary>

**Basic structure**:

```go
func BenchmarkFoo(b *testing.B) {
    setup() // outside the loop
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        foo() // what you're measuring
    }
}
```

**Key pitfalls**:

1. **Dead code elimination**: The compiler may optimize away computations whose results are unused. Assign the result to a package-level `var result T` variable to prevent this.

2. **Setup in the loop**: Database setup, HTTP server start, test data generation — these inflate the benchmark. Move them before the loop and call `b.ResetTimer()` after setup.

3. **Caching effects**: If the function under test caches results, the second and subsequent iterations are artificially fast. Use `b.N`-varying inputs or clear caches per iteration.

4. **System noise**: Run benchmarks multiple times (`-count=5`) and use [benchstat](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat) to compare results statistically.

5. **`-benchmem` is mandatory**: Always include allocation stats. A faster `ns/op` that allocates more can be slower under real load due to GC pressure.

6. **Comparing two implementations**: Run both in the same `go test -bench=.` invocation, then use `benchstat` to compare with statistical significance.

</details>

### Q4: How do you use `pprof` to find and fix a performance problem in production?

A practical question that tests production experience.

<details>
<summary>Answer</summary>

**Step 1: Expose the pprof endpoint** (add once, always available):

```go
import _ "net/http/pprof"
go http.ListenAndServe(":6060", nil)
```

**Step 2: Identify the symptom** — high CPU, high memory, high latency, goroutine growth.

**Step 3: Capture the right profile**:

- High CPU → `go tool pprof http://...:6060/debug/pprof/profile?seconds=30`
- High memory → `go tool pprof http://...:6060/debug/pprof/heap`
- Goroutine leak → `go tool pprof http://...:6060/debug/pprof/goroutine`

**Step 4: Analyze**:

```sh
(pprof) top10      # top functions by CPU/memory
(pprof) list foo   # annotated source of function foo
(pprof) web        # flame graph (requires graphviz)
```

**Step 5: Identify and fix**:

- High `mallocgc` → reduce allocations: use `sync.Pool`, avoid interface boxing, pre-allocate slices
- Specific function dominating CPU → algorithm improvement, caching, concurrency
- Many goroutines → goroutine leak, fix with context cancellation

**Step 6: Benchmark before and after** to confirm improvement, then ship.

**Production note**: CPU profiles have ~2% overhead. Memory profiles are instantaneous snapshots. Both are safe to capture in production during an incident.

</details>

## Key Takeaways

1. **Table-driven tests**: Define cases as a struct slice, loop with `t.Run` — adding a case is one line.
2. **`t.Helper()`**: Call in test helper functions so failures point to the caller, not the helper.
3. **`t.Fatal` vs `t.Error`**: Fatal stops the test; Error continues — use Fatal when remaining assertions would be nonsensical.
4. **Mocking via interfaces**: Depend on interfaces, inject fakes in tests — no mock framework required.
5. **`-race` flag**: Always run tests with the race detector in CI; it catches races that aren't otherwise apparent.
6. **`b.N` is dynamic**: The testing framework runs the benchmark function with increasing `b.N` until results stabilize.
7. **Prevent dead code elimination**: Assign benchmark results to a package-level variable.
8. **`-benchmem`**: Always include — allocation count matters as much as time per op.
9. **pprof**: Use CPU profiles for hot code, heap profiles for allocations, goroutine profiles for leak detection.
10. **`b.ResetTimer()`**: Exclude setup time from benchmark measurements.

## Next Steps

In [Lesson 09: Module System & Toolchain](lesson-09-module-system-and-toolchain.md), you'll learn:

- How `go.mod` and `go.sum` work together and what each guarantees
- Major version suffixes (`/v2`) and why they exist
- Go workspaces (`go.work`) for multi-module development
- Build tags, `go:generate`, and the static analysis tools every Go project should run
