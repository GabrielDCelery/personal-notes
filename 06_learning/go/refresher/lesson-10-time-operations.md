# Lesson 10: Time Operations

Go's `time` package looks simple on the surface — create a time, add a duration, format it. In practice it's a source of subtle bugs: wall clock vs monotonic mismatches, timezone traps, timer resets that don't work the way you expect, and code that's impossible to test because `time.Now()` is hardcoded everywhere.

## `time.Time` and `time.Duration`

### The Basics

`time.Time` is a value type representing an instant in time. `time.Duration` is an `int64` counting nanoseconds.

```go
now := time.Now()                        // current local time
utc := time.Now().UTC()                  // current UTC time
t   := time.Date(2024, time.January, 15, // specific instant
                 10, 30, 0, 0, time.UTC)

// Duration literals — always use typed constants
d := 5 * time.Second                    // ✓ time.Duration(5_000_000_000)
d := 5 * 1000 * time.Millisecond        // ✓ same value, different expression
d := time.Duration(5) * time.Second     // ✓ explicit conversion from int

raw := 5000000000                        // ❌ untyped int — won't compile as Duration
d   := time.Duration(raw)               // ✓ explicit cast required
```

### Time Arithmetic

```go
t1 := time.Now()
t2 := t1.Add(2 * time.Hour)            // add duration to time → time.Time
t3 := t1.Add(-30 * time.Minute)        // subtract duration

gap := t2.Sub(t1)                       // difference between two times → time.Duration
fmt.Println(gap)                        // "2h0m0s"

since := time.Since(t1)                // time.Now().Sub(t1) shorthand
until := time.Until(t2)                // t2.Sub(time.Now()) shorthand
```

### The Zero Value

```go
var t time.Time                         // zero value: January 1, year 1, 00:00:00 UTC
t.IsZero()                              // ✓ true — use this to check for unset time

if t == (time.Time{}) { ... }           // ❌ works, but IsZero() is idiomatic

// Common gotcha: zero time in structs
type Event struct {
    Name      string
    StartTime time.Time                 // zero if not set — always check IsZero()
}
```

### Comparing Times

```go
t1.Before(t2)       // t1 < t2
t1.After(t2)        // t1 > t2
t1.Equal(t2)        // ✓ correct comparison — handles monotonic clock

t1 == t2            // ❌ wrong — compares internal representation including location;
                    //    two times representing the same instant may not be ==
```

**Gotcha**: Never use `==` to compare `time.Time` values. Use `.Equal()`. Two times can represent the same instant but have different `Location` values, making `==` return `false`.

---

## Monotonic vs Wall Clock

This is one of the most misunderstood parts of the `time` package, and a reliable interview differentiator.

### Why Two Clocks?

The system clock (wall clock) can jump — NTP adjustments, daylight saving, leap seconds, manual changes. If you're measuring elapsed time and the clock jumps backward, your elapsed time becomes negative.

Go's answer (since 1.9): `time.Now()` returns a `time.Time` that contains **both** a wall clock reading **and** a monotonic clock reading. The monotonic reading is guaranteed to always increase.

```
time.Now() → {
    wall:      2024-01-15 10:30:00.000 UTC   // can jump
    monotonic: +1234567890ns since process start  // never jumps
}
```

### How Go Uses Them

```go
t1 := time.Now()
// ... do work ...
t2 := time.Now()

elapsed := t2.Sub(t1)   // ✓ uses monotonic readings — accurate even if wall clock adjusted
```

`Sub`, `Since`, `Until`, and comparisons between two `time.Now()` results all use the monotonic component automatically.

### When the Monotonic Clock Is Stripped

Operations that don't make sense with a monotonic reading strip it:

```go
t := time.Now()                        // has monotonic

t.Round(0)                             // strips monotonic
t.Truncate(0)                          // strips monotonic
t.Add(0)                               // keeps monotonic
t.UTC()                                // keeps monotonic
t.In(loc)                              // keeps monotonic

// Marshaling/unmarshaling strips monotonic
b, _ := json.Marshal(t)
var t2 time.Time
json.Unmarshal(b, &t2)                 // t2 has no monotonic clock

// time.Date() never has monotonic
fixed := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)  // wall only
```

### The Comparison Gotcha

```go
t1 := time.Now()
t2 := t1.Round(0)   // same wall time, but monotonic stripped from t2

t1.Equal(t2)        // ✓ true — Equal() uses wall time when one side has no monotonic
t1 == t2            // ❌ false — struct comparison includes the monotonic field
```

This is the most common source of test failures involving time. If you marshal a `time.Time` to JSON and back, then compare it to the original with `==`, it will fail. Use `.Equal()`.

```go
// In tests, strip monotonic explicitly to avoid surprises:
now := time.Now().Round(0)             // consistent: no monotonic, wall only
```

---

## Formatting and Parsing

### The Reference Time

Go doesn't use `%Y-%m-%d` style format strings. It uses a specific reference time:

```
Mon Jan 2 15:04:05 MST 2006
```

Which corresponds to:
```
01/02 03:04:05PM '06 -0700
```

Mnemonic: `1 2 3 4 5 6 7` — month, day, hour, minute, second, year, timezone offset.

```go
t := time.Date(2024, time.March, 15, 9, 30, 0, 0, time.UTC)

t.Format("2006-01-02")                  // "2024-03-15"
t.Format("2006-01-02 15:04:05")         // "2024-03-15 09:30:00"
t.Format("Jan 2, 2006")                 // "Mar 15, 2024"
t.Format("3:04 PM")                     // "9:30 AM"
t.Format(time.RFC3339)                  // "2024-03-15T09:30:00Z"
t.Format(time.RFC3339Nano)              // "2024-03-15T09:30:00.000000000Z"
```

### Parsing

```go
t, err := time.Parse("2006-01-02", "2024-03-15")
// t is in UTC (no location info in the string)

t, err := time.ParseInLocation("2006-01-02 15:04:05", "2024-03-15 09:30:00", loc)
// t is in loc's timezone
```

### Common Mistakes

```go
// ❌ Using wrong reference values in format string
t.Format("2006-01-02")   // ✓ correct
t.Format("YYYY-MM-DD")   // ❌ wrong — outputs "YYYY-MM-DD" literally
t.Format("2006-01-30")   // ❌ wrong — 30 is not a valid day reference; use 02

// ❌ Parsing without timezone when you care about timezone
t, _ := time.Parse("2006-01-02 15:04:05", "2024-03-15 09:30:00")
// t.Location() is UTC — not your local time, not the user's timezone

// ✓ Use ParseInLocation when timezone matters
t, _ := time.ParseInLocation("2006-01-02 15:04:05", "2024-03-15 09:30:00", time.Local)
```

### Pre-defined Formats

```go
time.RFC3339         // "2006-01-02T15:04:05Z07:00"   — use for APIs and logs
time.RFC3339Nano     // nanosecond precision
time.DateTime        // "2006-01-02 15:04:05"          — Go 1.20+
time.DateOnly        // "2006-01-02"                   — Go 1.20+
time.TimeOnly        // "15:04:05"                     — Go 1.20+
time.RFC822          // "02 Jan 06 15:04 MST"
time.RFC1123         // "Mon, 02 Jan 2006 15:04:05 MST" — HTTP headers
```

**Rule of thumb**: For any data that crosses a system boundary (APIs, databases, logs), use `time.RFC3339`. It's unambiguous, sortable as a string, and universally parseable.

---

## Locations and Timezones

### `time.Location`

```go
utc   := time.UTC                              // built-in
local := time.Local                            // machine's local timezone

loc, err := time.LoadLocation("America/New_York")  // IANA timezone database
loc, err := time.LoadLocation("Europe/London")
loc, err := time.LoadLocation("Asia/Tokyo")

// Fixed offset (no DST)
fixed := time.FixedZone("UTC+5", 5*60*60)     // +05:00
```

### Converting Between Timezones

```go
t := time.Now().UTC()
ny, _ := time.LoadLocation("America/New_York")
tNY := t.In(ny)                               // same instant, different representation

// Both represent the same moment:
t.Equal(tNY)   // true
t.Unix()       // same value as tNY.Unix()
```

### Common Traps

**Trap 1: `time.Local` in tests**

```go
// ❌ Test passes on your machine, fails in CI (different timezone)
t, _ := time.Parse("2006-01-02", "2024-03-15")
// t is always UTC from time.Parse — but developer assumes local

// ✓ Explicit about timezone
t, _ := time.ParseInLocation("2006-01-02", "2024-03-15", time.UTC)
```

**Trap 2: DST assumptions**

```go
t := time.Date(2024, time.March, 10, 2, 30, 0, 0, nyLoc)
// March 10, 2024 at 2:30 AM Eastern — this time doesn't exist
// Clocks spring forward: 2:00 AM → 3:00 AM
// Go handles this by adjusting to 3:30 AM, but the behavior is surprising
```

**Trap 3: `LoadLocation` requires timezone data**

```go
loc, err := time.LoadLocation("America/New_York")
// On minimal Docker containers (scratch, alpine without tzdata), this fails
// err: "unknown time zone America/New_York"

// Fix: import the embedded timezone database (Go 1.15+)
import _ "time/tzdata"    // embeds IANA database in the binary
```

**Trap 4: Storing timezone in databases**

```go
// ❌ Store only UTC timestamp — lose user's local context
// ❌ Store only local time string — ambiguous during DST transitions
// ✓ Store UTC timestamp + timezone name separately
type Appointment struct {
    StartUTC  time.Time  // store as UTC in DB
    Timezone  string     // "America/New_York" — reconstruct local on read
}
```

---

## Timers and Tickers

### One-shot: `time.Timer`

```go
timer := time.NewTimer(2 * time.Second)
<-timer.C                               // blocks until 2s elapsed

// ✓ Always stop timers you don't fully drain
timer := time.NewTimer(5 * time.Second)
select {
case <-timer.C:
    // fired
case <-ctx.Done():
    timer.Stop()                        // ✓ prevent goroutine leak
    // don't drain timer.C here — Stop() may or may not have drained it
}
```

### The `Reset` Gotcha

`timer.Reset()` is subtle and widely misused:

```go
// ❌ Incorrect reset pattern — can receive on an already-fired timer
timer.Stop()
timer.Reset(newDuration)    // race: timer.C may still have a value

// ✓ Correct reset pattern
if !timer.Stop() {
    select {
    case <-timer.C:         // drain if it fired before Stop()
    default:
    }
}
timer.Reset(newDuration)
```

**Why**: `Stop()` returns `false` if the timer already fired. If `false`, the timer's channel already has a value. If you don't drain it before `Reset()`, the next `<-timer.C` reads the stale value from the previous expiry, not the new one.

**Practical advice**: In most cases, it's simpler to create a new timer than to reset an existing one. Only optimize to reuse timers when benchmarks show it matters.

### Repeating: `time.Ticker`

```go
ticker := time.NewTicker(1 * time.Second)
defer ticker.Stop()                     // ✓ always Stop() tickers — they don't stop themselves

for {
    select {
    case t := <-ticker.C:
        fmt.Println("tick at", t)
    case <-ctx.Done():
        return
    }
}
```

**Gotcha**: Unlike `Timer`, `Ticker` does NOT buffer dropped ticks. If your handler takes longer than the tick interval, ticks are dropped silently — the ticker adjusts to maintain the interval cadence, not to deliver every tick.

```go
ticker := time.NewTicker(100 * time.Millisecond)
// handler takes 150ms → some ticks are dropped — this is by design
```

### `time.AfterFunc`

Runs a function in a new goroutine after a duration:

```go
t := time.AfterFunc(5*time.Second, func() {
    fmt.Println("fired in separate goroutine")
})
t.Stop()    // ✓ cancel it if no longer needed
```

### `time.After` — The Leak

```go
// ❌ Common in select statements — leaks timer until it fires
select {
case msg := <-ch:
    process(msg)
case <-time.After(5 * time.Second):
    return ErrTimeout
}
// If msg arrives before 5s, the timer is NOT garbage collected until 5s passes.
// In a hot loop, this creates thousands of leaked timers.

// ✓ Use context timeout instead
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()
select {
case msg := <-ch:
    process(msg)
case <-ctx.Done():
    return ErrTimeout
}

// ✓ Or reuse a timer explicitly
timer := time.NewTimer(5 * time.Second)
defer timer.Stop()
select {
case msg := <-ch:
    process(msg)
case <-timer.C:
    return ErrTimeout
}
```

---

## Testable Time: The Clock Interface Pattern

Code that calls `time.Now()` directly is untestable — you can't control what it returns. The solution is to inject a clock.

### The Pattern

```go
// Define a minimal interface
type Clock interface {
    Now() time.Time
    Since(time.Time) time.Duration
}

// Real implementation wraps the standard library
type realClock struct{}

func (realClock) Now() time.Time                  { return time.Now() }
func (realClock) Since(t time.Time) time.Duration  { return time.Since(t) }

// Fake implementation for tests
type fakeClock struct {
    current time.Time
}

func (f *fakeClock) Now() time.Time                  { return f.current }
func (f *fakeClock) Since(t time.Time) time.Duration  { return f.current.Sub(t) }
func (f *fakeClock) Advance(d time.Duration)           { f.current = f.current.Add(d) }
```

### Usage

```go
// Production code depends on the interface
type RateLimiter struct {
    clock    Clock
    lastSeen map[string]time.Time
    window   time.Duration
}

func NewRateLimiter(clock Clock, window time.Duration) *RateLimiter {
    return &RateLimiter{clock: clock, lastSeen: make(map[string]time.Time), window: window}
}

func (r *RateLimiter) Allow(key string) bool {
    now := r.clock.Now()
    if last, ok := r.lastSeen[key]; ok && r.clock.Since(last) < r.window {
        return false
    }
    r.lastSeen[key] = now
    return true
}

// Wiring in main
limiter := NewRateLimiter(realClock{}, time.Minute)

// Test
func TestRateLimiter(t *testing.T) {
    clk := &fakeClock{current: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}
    limiter := NewRateLimiter(clk, time.Minute)

    if !limiter.Allow("user1") { t.Fatal("first request should be allowed") }
    if limiter.Allow("user1")  { t.Fatal("immediate retry should be blocked") }

    clk.Advance(61 * time.Second)   // ✓ control time precisely

    if !limiter.Allow("user1") { t.Fatal("after window, should be allowed again") }
}
```

### Alternatives

| Approach | When to Use |
| -------- | ----------- |
| Clock interface (above) | Complex domain logic with multiple time-dependent methods |
| `func() time.Time` field | Simpler — just inject `time.Now` as a function |
| `github.com/jonboulle/clockwork` | Third-party clock interface with fake; saves boilerplate |
| Accept `time.Time` as parameter | Simplest — push `time.Now()` up to the caller |

```go
// Simplest approach: pass time as a parameter
func IsExpired(t time.Time, now time.Time) bool {
    return now.After(t)
}

// Caller decides what "now" means — trivially testable
IsExpired(expiryTime, time.Now())              // production
IsExpired(expiryTime, time.Unix(0, 0).UTC())   // test
```

---

## Hands-On Exercise 1: Timer and Context Timeout

The following function has a timer leak and an incorrect reset. Fix both issues.

```go
func processWithRetry(ch <-chan string, maxAttempts int) (string, error) {
    for i := 0; i < maxAttempts; i++ {
        timer := time.NewTimer(500 * time.Millisecond)
        select {
        case msg := <-ch:
            timer.Reset(500 * time.Millisecond)  // reset for next iteration
            if validate(msg) {
                return msg, nil
            }
        case <-timer.C:
            continue
        }
    }
    return "", errors.New("max attempts reached")
}
```

<details>
<summary>Solution</summary>

**Issues**:

1. ❌ `timer.Reset()` is called after the timer fired (inside `case <-timer.C`) — the timer is already expired, so no drain is needed, but the reset logic is wrong because it resets in the `msg` branch without stopping first.
2. ❌ When `msg` is received and `validate` returns `false`, the old timer is never stopped — it leaks until it fires.
3. ❌ The reset inside the `msg` branch is pointless — a new loop iteration will create a new timer anyway, making the reset both incorrect and unnecessary.

**Fixed**:

```go
func processWithRetry(ch <-chan string, maxAttempts int) (string, error) {
    for i := 0; i < maxAttempts; i++ {
        timer := time.NewTimer(500 * time.Millisecond)
        select {
        case msg := <-ch:
            timer.Stop()            // ✓ always stop the timer when done with it
            if validate(msg) {
                return msg, nil
            }
        case <-timer.C:
            // timer fired naturally — no Stop() needed, channel already drained
        }
    }
    return "", errors.New("max attempts reached")
}
```

**Why create a new timer per iteration**: Reusing a timer with `Reset` is error-prone. Creating a new timer per iteration is clear, correct, and the GC handles cleanup. Only optimize this if profiling shows it matters.

</details>

## Hands-On Exercise 2: Untestable Time

The following code is impossible to test deterministically. Refactor it to be testable, then write a test for the expiry logic.

```go
type Session struct {
    ID        string
    CreatedAt time.Time
    TTL       time.Duration
}

func (s *Session) IsExpired() bool {
    return time.Now().After(s.CreatedAt.Add(s.TTL))
}

func (s *Session) TimeRemaining() time.Duration {
    return time.Until(s.CreatedAt.Add(s.TTL))
}
```

<details>
<summary>Solution</summary>

**Problem**: Both methods call `time.Now()` / `time.Until()` directly. Any test is racing against real time.

**Refactored** — simplest approach (push `now` as a parameter):

```go
type Session struct {
    ID        string
    CreatedAt time.Time
    TTL       time.Duration
}

func (s *Session) IsExpired(now time.Time) bool {
    return now.After(s.CreatedAt.Add(s.TTL))
}

func (s *Session) TimeRemaining(now time.Time) time.Duration {
    return s.CreatedAt.Add(s.TTL).Sub(now)
}
```

**Test**:

```go
func TestSession(t *testing.T) {
    created := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
    session := &Session{
        ID:        "abc",
        CreatedAt: created,
        TTL:       30 * time.Minute,
    }

    // 10 minutes in — not expired
    now := created.Add(10 * time.Minute)
    if session.IsExpired(now) {
        t.Error("session should not be expired at 10 minutes")
    }
    if got := session.TimeRemaining(now); got != 20*time.Minute {
        t.Errorf("expected 20m remaining, got %v", got)
    }

    // 31 minutes in — expired
    now = created.Add(31 * time.Minute)
    if !session.IsExpired(now) {
        t.Error("session should be expired at 31 minutes")
    }
    if got := session.TimeRemaining(now); got >= 0 {
        t.Errorf("expected negative remaining time, got %v", got)
    }
}
```

**Production callers** pass `time.Now()`:

```go
session.IsExpired(time.Now())
session.TimeRemaining(time.Now())
```

</details>

---

## Interview Questions

### Q1: What is the monotonic clock reading in `time.Time` and why does it exist?

Interviewers ask this to separate developers who've read the docs from those who've actually debugged time-related issues in production. It reveals understanding of operating system fundamentals and Go's pragmatic design choices.

<details>
<summary>Answer</summary>

The wall clock can jump — NTP adjustments, DST changes, leap seconds, or manual intervention can move it backward or forward. If you measure elapsed time using only the wall clock, a backward jump produces a negative duration.

Since Go 1.9, `time.Now()` stores two readings:

- **Wall clock**: the real-world time, used for display and absolute time comparisons
- **Monotonic clock**: nanoseconds since an arbitrary process-start epoch, guaranteed to only increase

`Sub`, `Since`, and `Until` use the monotonic component when both operands have one, ensuring elapsed time measurements are always accurate regardless of wall clock adjustments.

The monotonic reading is stripped by operations that produce times independent of `time.Now()`: marshaling/unmarshaling, `time.Date()`, `Round(0)`, `Truncate(0)`. After stripping, comparisons fall back to wall clock.

**Key implication for tests**: If you call `time.Now()`, marshal it to JSON, unmarshal it, and then compare with `==` to the original, you get `false` — the marshal/unmarshal strips the monotonic clock. Always use `.Equal()` for time comparisons.

</details>

### Q2: Why does `time.After` in a `select` loop cause a memory leak, and what's the fix?

A practical question — tests whether the developer understands channel semantics and Go's runtime behavior. This comes up in real code reviews constantly.

<details>
<summary>Answer</summary>

`time.After(d)` creates a `time.Timer` internally and returns its channel. The timer holds a reference in the Go runtime until it fires. If the other `select` case triggers first, the timer is not stopped — it keeps the timer alive in memory until `d` elapses.

In a hot loop, each iteration allocates a new timer that may not be collected for up to `d`. With a 5-second timeout and 1000 RPS, you're holding up to 5000 live timers simultaneously.

```go
// ❌ Leaks one timer per iteration if ch delivers faster than timeout
for {
    select {
    case msg := <-ch:
        process(msg)
    case <-time.After(5 * time.Second):
        return ErrTimeout
    }
}
```

**Fix 1 — context timeout** (preferred when a deadline applies to the whole loop):

```go
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()
for {
    select {
    case msg := <-ch:
        process(msg)
    case <-ctx.Done():
        return ErrTimeout
    }
}
```

**Fix 2 — reuse a timer** (when you want per-iteration timeout):

```go
timer := time.NewTimer(5 * time.Second)
defer timer.Stop()
for {
    // Reset timer for each iteration (correctly)
    if !timer.Stop() {
        select { case <-timer.C: default: }
    }
    timer.Reset(5 * time.Second)

    select {
    case msg := <-ch:
        process(msg)
    case <-timer.C:
        return ErrTimeout
    }
}
```

</details>

### Q3: What happens when you call `time.LoadLocation("America/New_York")` in a scratch Docker container?

A production-readiness question. It distinguishes developers who've deployed Go services in minimal containers from those who've only developed locally.

<details>
<summary>Answer</summary>

It returns an error: `"unknown time zone America/New_York"`.

`time.LoadLocation` reads timezone data from the OS — typically `/usr/share/zoneinfo` on Linux. Minimal containers (`scratch`, `alpine` without the `tzdata` package) don't include this data.

**Fix 1 — embed the timezone database** (Go 1.15+):

```go
import _ "time/tzdata"
```

This blank import embeds the full IANA timezone database into the binary at compile time. No OS timezone data needed. Binary size increases by ~450KB.

**Fix 2 — install tzdata in the container**:

```dockerfile
# Alpine
RUN apk add --no-cache tzdata

# Debian/Ubuntu
RUN apt-get install -y tzdata
```

**Fix 3 — use `time.FixedZone` for known offsets** (not recommended — loses DST handling):

```go
eastern := time.FixedZone("ET", -5*60*60)  // ❌ wrong during EDT (UTC-4)
```

**Recommendation**: Always use `import _ "time/tzdata"` in services that handle user-facing times. It's a minor binary size trade-off for correct, portable timezone handling.

</details>

### Q4: How do you design time-dependent code to be testable?

A design question — tests whether the candidate thinks about testability as a first-class concern. Expected from senior developers.

<details>
<summary>Answer</summary>

The core problem: `time.Now()` is a global function. Code that calls it directly cannot be tested with controlled time.

**Three main approaches, in order of preference**:

**1. Pass time as a parameter** (simplest — no abstraction needed):

```go
func IsExpired(expiry, now time.Time) bool {
    return now.After(expiry)
}
// Test: pass any time you want
```

Best when only one or two methods need time. No interface, no indirection.

**2. Inject a `func() time.Time`** (lightweight injection):

```go
type Cache struct {
    now func() time.Time
}
func NewCache() *Cache { return &Cache{now: time.Now} }
// Test: c.now = func() time.Time { return fixedTime }
```

Best when a struct has a few methods using `time.Now()` and you don't want a full interface.

**3. Clock interface** (most flexible):

```go
type Clock interface {
    Now() time.Time
    Since(time.Time) time.Duration
}
// Real: wraps time package. Fake: stores and advances a fixed time.
```

Best for complex domain logic where you need to advance time, simulate timeouts, or test multiple time-dependent interactions in sequence.

**What to avoid**:

- ❌ Mocking the `time` package via monkey-patching (fragile, not idiomatic Go)
- ❌ Using `time.Sleep` in tests to wait for time-dependent behaviour (slow and flaky)
- ❌ Making time-dependent logic untestable by hardcoding `time.Now()`

The Go standard library itself uses the parameter-passing approach in many places. Prefer simplicity: only introduce a Clock interface when parameter passing becomes unwieldy.

</details>

---

## Key Takeaways

1. **`time.Time` is a value type**: use `.Equal()` for comparisons, never `==` — two times at the same instant with different locations or clock readings won't be `==`.
2. **Monotonic clock**: `time.Now()` includes both wall and monotonic readings; `Sub`/`Since`/`Until` use monotonic for accuracy. Marshaling strips it.
3. **Format reference time**: Go uses `Mon Jan 2 15:04:05 MST 2006` — memorise `1 2 3 4 5 6 7`. Use `time.RFC3339` for any data crossing system boundaries.
4. **Timezone data**: always `import _ "time/tzdata"` in services deployed to minimal containers; `time.LoadLocation` fails without OS timezone data.
5. **`time.After` leaks**: every call allocates a timer that lives until expiry; use context timeouts or an explicit `time.NewTimer` with `defer Stop()` in loops.
6. **`timer.Reset` is subtle**: if `Stop()` returns `false`, drain the channel before resetting; in practice, creating a new timer is simpler and correct.
7. **Tickers drop ticks**: if your handler is slower than the tick interval, intermediate ticks are silently discarded — this is by design.
8. **Testable time**: push `time.Now()` to the boundary — pass it as a parameter, inject it as a `func() time.Time`, or use a Clock interface. Never call `time.Now()` deep in domain logic.
9. **UTC in storage**: always store timestamps as UTC; store the timezone name separately if you need to reconstruct the user's local time.
10. **`time.Since(t1)` is shorthand** for `time.Now().Sub(t1)` — prefer it for readability; `time.Until(t2)` is shorthand for `t2.Sub(time.Now())`.

## Next Steps

This lesson completes the core Go refresher series. For further depth:

- [Go Channels Deep Dive](../channels/) — 23-lesson series covering channel patterns exhaustively
- [Go Primitives](../primitives/) — foundational Go data types and their performance characteristics
- [time package documentation](https://pkg.go.dev/time) — full API reference
- [jonboulle/clockwork](https://github.com/jonboulle/clockwork) — a popular Clock interface library used in production codebases
