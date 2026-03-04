# Go Regex (regexp)

> Go uses RE2 syntax — no backreferences or lookaheads, but guaranteed linear time.

## Why

- **RE2 not PCRE** — Go deliberately uses RE2 for guaranteed O(n) matching. No catastrophic backtracking. The trade-off is no lookaheads or backreferences.
- **MustCompile for package-level vars** — Compile once, reuse everywhere. MustCompile panics on invalid patterns, which is fine at init time — you want to fail fast.
- **Compile for user input** — When the pattern comes from a user or config, use Compile to handle errors gracefully instead of panicking.
- **strings package first** — For simple operations (Contains, HasPrefix, Split), the strings package is faster and clearer than regex. Only reach for regexp when you need pattern matching.

## Quick Reference

| Use case          | Method                           |
| ----------------- | -------------------------------- |
| Match check       | `regexp.MatchString(pattern, s)` |
| Compile           | `regexp.MustCompile(pattern)`    |
| Find first        | `re.FindString(s)`               |
| Find all          | `re.FindAllString(s, -1)`        |
| Submatch (groups) | `re.FindStringSubmatch(s)`       |
| Replace           | `re.ReplaceAllString(s, repl)`   |
| Split             | `re.Split(s, -1)`                |

## Basics

### 1. Quick match check

```go
matched, err := regexp.MatchString(`^\d{3}-\d{4}$`, "123-4567")
// true, nil
```

### 2. Compile and reuse (preferred)

```go
// MustCompile panics on invalid pattern — use for package-level vars
var emailRe = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

emailRe.MatchString("alice@example.com")  // true
```

### 3. Compile with error handling

```go
re, err := regexp.Compile(`[invalid`)
if err != nil {
    // handle bad pattern
}
```

## Finding Matches

### 4. Find first match

```go
re := regexp.MustCompile(`\d+`)

re.FindString("abc 123 def 456")  // "123"
```

### 5. Find all matches

```go
re := regexp.MustCompile(`\d+`)

re.FindAllString("abc 123 def 456", -1)  // ["123", "456"]
re.FindAllString("abc 123 def 456", 1)   // ["123"] — limit to 1
```

### 6. Capture groups (submatches)

```go
re := regexp.MustCompile(`(\w+):(\d+)`)

match := re.FindStringSubmatch("host:8080")
// match[0] = "host:8080"  (full match)
// match[1] = "host"       (group 1)
// match[2] = "8080"       (group 2)
```

### 7. Named capture groups

```go
re := regexp.MustCompile(`(?P<name>\w+):(?P<port>\d+)`)

match := re.FindStringSubmatch("host:8080")
for i, name := range re.SubexpNames() {
    if name != "" {
        fmt.Printf("%s: %s\n", name, match[i])
    }
}
// name: host
// port: 8080
```

### 8. Find all with groups

```go
re := regexp.MustCompile(`(\w+)=(\w+)`)

matches := re.FindAllStringSubmatch("a=1 b=2 c=3", -1)
for _, m := range matches {
    fmt.Printf("%s -> %s\n", m[1], m[2])
}
// a -> 1
// b -> 2
// c -> 3
```

## Replacing

### 9. Simple replace

```go
re := regexp.MustCompile(`\d+`)

re.ReplaceAllString("abc 123 def 456", "NUM")  // "abc NUM def NUM"
```

### 10. Replace with capture group reference

```go
re := regexp.MustCompile(`(\w+)@(\w+)`)

re.ReplaceAllString("user@host", "${2}/${1}")  // "host/user"
```

### 11. Replace with function

```go
re := regexp.MustCompile(`\d+`)

result := re.ReplaceAllStringFunc("abc 2 def 3", func(s string) string {
    n, _ := strconv.Atoi(s)
    return strconv.Itoa(n * 10)
})
// "abc 20 def 30"
```

## Splitting

### 12. Split string by pattern

```go
re := regexp.MustCompile(`[,;\s]+`)

re.Split("a,b; c  d", -1)  // ["a", "b", "c", "d"]
```
