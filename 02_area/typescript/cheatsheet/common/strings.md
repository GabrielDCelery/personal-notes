# TypeScript Strings

## Why

- **Strings are immutable** — Every operation returns a new string. Concatenating in a loop with `+=` creates intermediate strings. For many joins, use array `.join()` instead.
- **Template literals over concatenation** — `` `Hello ${name}` `` is more readable than `"Hello " + name`. Also supports multi-line strings without `\n`.
- **padStart/padEnd for formatting** — Useful for IDs, timestamps, table output. Better than manual string math.
- **Intl for locale-aware operations** — `toUpperCase()` works for ASCII but fails for Turkish İ/i. Use `Intl.Collator` for locale-aware sorting and comparison.
- **String.raw for regex and paths** — Tagged template that doesn't process escape sequences. `String.raw`\`\n\``returns the literal`\n`, not a newline.

## Quick Reference

| Use case         | Method                                   |
| ---------------- | ---------------------------------------- |
| Contains         | `s.includes("sub")`                      |
| Starts/ends with | `s.startsWith()` / `s.endsWith()`        |
| Find index       | `s.indexOf("sub")`                       |
| Split            | `s.split(",")`                           |
| Join             | `arr.join(",")`                          |
| Replace first    | `s.replace("old", "new")`                |
| Replace all      | `s.replaceAll("old", "new")`             |
| Trim whitespace  | `s.trim()` / `trimStart()` / `trimEnd()` |
| Upper / lower    | `s.toUpperCase()` / `s.toLowerCase()`    |
| Pad              | `s.padStart(n, "0")` / `s.padEnd()`      |
| Substring        | `s.slice(start, end)`                    |
| Repeat           | `s.repeat(n)`                            |

## Searching

### 1. Contains, starts with, ends with

```typescript
const s = "Hello, World!";

s.includes("World"); // true
s.startsWith("Hello"); // true
s.endsWith("!"); // true
s.indexOf("World"); // 7 (-1 if not found)
```

### 2. Case-insensitive search

```typescript
const query = "hello";
const text = "Hello World";

text.toLowerCase().includes(query.toLowerCase()); // true

// For repeated comparisons, use Intl.Collator
const collator = new Intl.Collator("en", { sensitivity: "base" });
collator.compare("hello", "Hello") === 0; // true (equal)
```

## Splitting & Joining

### 3. Split

```typescript
"a,b,c".split(","); // ["a", "b", "c"]
"a,b,c".split(",", 2); // ["a", "b"] — limit
"  foo  bar  ".split(/\s+/); // ["", "foo", "bar", ""]
"  foo  bar  ".trim().split(/\s+/); // ["foo", "bar"]
```

### 4. Join

```typescript
["a", "b", "c"].join(", "); // "a, b, c"
["a", "b", "c"].join(""); // "abc"
```

## Transforming

### 5. Replace

```typescript
"foo bar foo".replace("foo", "baz"); // "baz bar foo" — first only
"foo bar foo".replaceAll("foo", "baz"); // "baz bar baz"
"foo bar foo".replace(/foo/g, "baz"); // "baz bar baz" — regex

// With capture groups
"2024-01-15".replace(/(\d{4})-(\d{2})-(\d{2})/, "$2/$3/$1"); // "01/15/2024"
```

### 6. Trim

```typescript
"  hello  ".trim(); // "hello"
"  hello  ".trimStart(); // "hello  "
"  hello  ".trimEnd(); // "  hello"

// Trim specific characters — no built-in, use replace
"***hello***".replace(/^\*+|\*+$/g, ""); // "hello"
```

### 7. Case conversion

```typescript
"hello".toUpperCase(); // "HELLO"
"HELLO".toLowerCase(); // "hello"

// No built-in title case — simple version
function titleCase(s: string): string {
  return s.replace(/\b\w/g, (c) => c.toUpperCase());
}
titleCase("hello world"); // "Hello World"
```

### 8. Padding

```typescript
"42".padStart(6, "0"); // "000042"
"42".padEnd(6, "."); // "42...."
"5".padStart(2, "0"); // "05" — useful for dates/times
```

## Template Literals

### 9. Multi-line and expressions

```typescript
const name = "Alice";
const age = 30;

const msg = `Hello ${name}, you are ${age} years old`;

const sql = `
  SELECT id, name
  FROM users
  WHERE age > ${minAge}
  ORDER BY name
`;
```

### 10. Tagged templates

```typescript
function sql(strings: TemplateStringsArray, ...values: unknown[]) {
  const text = strings.reduce((acc, str, i) => acc + `$${i}` + str);
  return { text, values };
}

const id = "user-123";
const query = sql`SELECT * FROM users WHERE id = ${id}`;
// { text: "SELECT * FROM users WHERE id = $1", values: ["user-123"] }
```

## Type Conversions

### 11. String to number

```typescript
parseInt("42", 10); // 42 — always pass radix
parseInt("0xff", 16); // 255
parseFloat("3.14"); // 3.14
Number("42"); // 42 — stricter, NaN on "42abc"
+"42"; // 42 — shorthand for Number()
```

### 12. Number to string

```typescript
(42).toString(); // "42"
(255).toString(16); // "ff"
(3.14159).toFixed(2); // "3.14"
String(42); // "42"
```

## Patterns

### 13. Slugify

```typescript
function slugify(s: string): string {
  return s
    .toLowerCase()
    .trim()
    .replace(/[^\w\s-]/g, "")
    .replace(/[\s_]+/g, "-")
    .replace(/^-+|-+$/g, "");
}

slugify("Hello World! 123"); // "hello-world-123"
```

### 14. Truncate with ellipsis

```typescript
function truncate(s: string, maxLength: number): string {
  if (s.length <= maxLength) return s;
  return s.slice(0, maxLength - 3) + "...";
}

truncate("Hello World", 8); // "Hello..."
```
