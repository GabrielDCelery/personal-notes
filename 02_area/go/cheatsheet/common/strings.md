# Go Strings & Strconv

## Quick Reference — strings

| Use case        | Method                                    |
| --------------- | ----------------------------------------- |
| Contains        | `strings.Contains(s, "sub")`              |
| Prefix / suffix | `strings.HasPrefix` / `strings.HasSuffix` |
| Split           | `strings.Split(s, ",")`                   |
| Join            | `strings.Join(slice, ",")`                |
| Replace         | `strings.ReplaceAll(s, old, new)`         |
| Trim whitespace | `strings.TrimSpace(s)`                    |
| To upper/lower  | `strings.ToUpper` / `strings.ToLower`     |
| Build strings   | `strings.Builder`                         |

## Quick Reference — strconv

| Use case       | Method                           |
| -------------- | -------------------------------- |
| Int → string   | `strconv.Itoa(42)`               |
| String → int   | `strconv.Atoi("42")`             |
| Float → string | `strconv.FormatFloat`            |
| String → float | `strconv.ParseFloat("3.14", 64)` |
| Bool → string  | `strconv.FormatBool(true)`       |
| String → bool  | `strconv.ParseBool("true")`      |

## Searching & Matching

### 1. Contains, prefix, suffix

```go
s := "Hello, World!"

strings.Contains(s, "World")    // true
strings.HasPrefix(s, "Hello")   // true
strings.HasSuffix(s, "!")       // true
strings.Count(s, "l")           // 3
strings.Index(s, "World")       // 7 (-1 if not found)
```

### 2. EqualFold — case-insensitive comparison

```go
strings.EqualFold("Go", "go") // true
```

## Splitting & Joining

### 3. Split

```go
parts := strings.Split("a,b,c", ",")     // ["a", "b", "c"]
parts := strings.SplitN("a,b,c", ",", 2) // ["a", "b,c"] — max 2 parts
parts := strings.Fields("  foo  bar  ")   // ["foo", "bar"] — split on whitespace
```

### 4. Join

```go
strings.Join([]string{"a", "b", "c"}, ", ") // "a, b, c"
```

## Transforming

### 5. Replace

```go
strings.ReplaceAll("foo bar foo", "foo", "baz") // "baz bar baz"
strings.Replace("foo bar foo", "foo", "baz", 1) // "baz bar foo" — replace first only
```

### 6. Trim

```go
strings.TrimSpace("  hello  ")           // "hello"
strings.Trim("***hello***", "*")         // "hello"
strings.TrimPrefix("hello.txt", "hello") // ".txt"
strings.TrimSuffix("hello.txt", ".txt")  // "hello"
strings.TrimLeft("000123", "0")          // "123"
```

### 7. Case conversion

```go
strings.ToUpper("hello")  // "HELLO"
strings.ToLower("HELLO")  // "hello"
strings.Title("hello world") // deprecated — use golang.org/x/text/cases
```

### 8. Map — transform each rune

```go
rot13 := strings.Map(func(r rune) rune {
    if r >= 'a' && r <= 'z' {
        return 'a' + (r-'a'+13)%26
    }
    return r
}, "hello") // "uryyb"
```

## Building Strings

### 9. strings.Builder (efficient concatenation)

```go
var b strings.Builder
for i := 0; i < 1000; i++ {
    b.WriteString("item ")
    b.WriteString(strconv.Itoa(i))
    b.WriteByte('\n')
}
result := b.String()
```

Use `strings.Builder` instead of `+=` in loops — avoids O(n^2) allocations.

### 10. fmt.Sprintf (convenient formatting)

```go
s := fmt.Sprintf("user:%s age:%d score:%.2f", name, age, score)
```

## Type Conversions (strconv)

### 11. String ↔ int

```go
// Int to string
s := strconv.Itoa(42) // "42"

// String to int
n, err := strconv.Atoi("42")
if err != nil {
    // handle invalid input
}

// With more control — base 10, 64-bit
n, err := strconv.ParseInt("42", 10, 64)

// Format with base
s := strconv.FormatInt(255, 16) // "ff"
```

### 12. String ↔ float

```go
// String to float
f, err := strconv.ParseFloat("3.14", 64)

// Float to string
s := strconv.FormatFloat(3.14, 'f', 2, 64) // "3.14"
// 'f' = decimal, 'e' = scientific, 'g' = auto
// 2 = precision, 64 = float64
```

### 13. String ↔ bool

```go
b, err := strconv.ParseBool("true")  // true
b, err := strconv.ParseBool("1")     // true
b, err := strconv.ParseBool("0")     // false

s := strconv.FormatBool(true) // "true"
```

### 14. Quoting (escaping for display/logs)

```go
strconv.Quote("hello\nworld")   // `"hello\nworld"`
strconv.Unquote(`"hello"`)      // hello, nil
```

## Patterns

### 15. Parse key=value pairs

```go
func parseKV(s string) map[string]string {
    result := make(map[string]string)
    for _, pair := range strings.Split(s, ",") {
        k, v, ok := strings.Cut(pair, "=")
        if ok {
            result[strings.TrimSpace(k)] = strings.TrimSpace(v)
        }
    }
    return result
}

// "host=localhost, port=5432" → {"host": "localhost", "port": "5432"}
```

`strings.Cut` (Go 1.18+) is cleaner than `SplitN` for key-value parsing.

### 16. Slugify a string

```go
func slugify(s string) string {
    s = strings.ToLower(s)
    s = strings.TrimSpace(s)
    var b strings.Builder
    for _, r := range s {
        if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
            b.WriteRune(r)
        } else if r == ' ' || r == '-' {
            b.WriteByte('-')
        }
    }
    return b.String()
}
```
