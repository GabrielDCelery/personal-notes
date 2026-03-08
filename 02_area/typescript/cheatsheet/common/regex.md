# TypeScript Regex

> JavaScript uses a PCRE-like engine — supports lookaheads, lookbehinds, named groups, and backreferences. Unlike Go's RE2, backtracking is possible, so catastrophic patterns can hang.

## Why

- **Literal vs constructor** — `/pattern/` is checked at parse time. `new RegExp(pattern)` compiles at runtime — use it when the pattern comes from a variable. Literal syntax is preferred for static patterns.
- **Flags matter** — `g` (global) changes behavior of `exec`, `match`, and `test` by advancing `lastIndex`. Without `g`, methods only find the first match.
- **exec has state when using /g** — A regex with the `g` flag maintains `lastIndex` between calls. This is useful for iterating matches but causes bugs if you reuse the regex across unrelated calls.
- **matchAll over exec loop** — `string.matchAll(re)` returns an iterator of all matches with groups. Cleaner than manually looping `re.exec()`. Requires the `g` flag.
- **Named groups** — `(?<name>...)` makes capture groups self-documenting and less fragile than positional `$1`, `$2`.

## Quick Reference

| Use case          | Method                            |
| ----------------- | --------------------------------- |
| Test match        | `re.test(s)`                      |
| First match       | `s.match(re)` (without g)        |
| All matches       | `s.matchAll(re)` (with g)        |
| Replace first     | `s.replace(re, repl)`             |
| Replace all       | `s.replaceAll(re, repl)` (with g)|
| Split by pattern  | `s.split(re)`                     |
| From variable     | `new RegExp(pattern, flags)`      |
| Named group       | `(?<name>...)`                    |

## Basics

### 1. Test if a string matches

```typescript
const re = /^\d{3}-\d{4}$/;
re.test("123-4567"); // true
re.test("abc");      // false
```

### 2. Literal vs constructor

```typescript
// Literal — preferred for static patterns
const re = /\d+/g;

// Constructor — when pattern is dynamic
const userInput = "hello";
const re = new RegExp(userInput, "gi");
```

### 3. Common flags

```typescript
/pattern/g;   // global — find all matches
/pattern/i;   // case-insensitive
/pattern/m;   // multiline — ^ and $ match line boundaries
/pattern/s;   // dotAll — . matches newlines
/pattern/u;   // unicode — correct handling of surrogate pairs
/pattern/gi;  // combine flags
```

## Finding Matches

### 4. First match with groups

```typescript
const match = "host:8080".match(/(\w+):(\d+)/);
if (match) {
  match[0]; // "host:8080" — full match
  match[1]; // "host" — group 1
  match[2]; // "8080" — group 2
}
```

### 5. All matches with matchAll

```typescript
const re = /(\w+)=(\w+)/g;
const str = "a=1 b=2 c=3";

for (const match of str.matchAll(re)) {
  console.log(`${match[1]} -> ${match[2]}`);
}
// a -> 1
// b -> 2
// c -> 3

// Or collect into array
const matches = [...str.matchAll(re)];
```

### 6. Named capture groups

```typescript
const re = /(?<name>\w+):(?<port>\d+)/;
const match = "host:8080".match(re);

if (match?.groups) {
  match.groups.name; // "host"
  match.groups.port; // "8080"
}
```

### 7. Named groups with matchAll

```typescript
const re = /(?<key>\w+)=(?<value>\w+)/g;

for (const { groups } of "a=1 b=2".matchAll(re)) {
  console.log(`${groups!.key} -> ${groups!.value}`);
}
```

## Replacing

### 8. Simple replace

```typescript
"abc 123 def 456".replace(/\d+/, "NUM");    // "abc NUM def 456" — first only
"abc 123 def 456".replaceAll(/\d+/g, "NUM"); // "abc NUM def NUM"
// or
"abc 123 def 456".replace(/\d+/g, "NUM");    // same with g flag
```

### 9. Replace with capture group reference

```typescript
"2024-01-15".replace(/(\d{4})-(\d{2})-(\d{2})/, "$2/$3/$1");
// "01/15/2024"

// Named groups
"user@host".replace(/(?<user>\w+)@(?<host>\w+)/, "$<host>/$<user>");
// "host/user"
```

### 10. Replace with function

```typescript
"abc 2 def 3".replace(/\d+/g, (match) => {
  return String(Number(match) * 10);
});
// "abc 20 def 30"

// With groups
"John Smith".replace(/(?<first>\w+) (?<last>\w+)/, (_, first, last) => {
  return `${last}, ${first}`;
});
// "Smith, John"
```

## Splitting

### 11. Split by pattern

```typescript
"a,b; c  d".split(/[,;\s]+/); // ["a", "b", "c", "d"]
"one1two2three".split(/\d/);   // ["one", "two", "three"]
```

## Lookahead & Lookbehind

### 12. Lookahead and lookbehind

```typescript
// Positive lookahead — match "foo" followed by "bar"
"foobar foobaz".match(/foo(?=bar)/g); // ["foo"]

// Negative lookahead — match "foo" NOT followed by "bar"
"foobar foobaz".match(/foo(?!bar)/g); // ["foo"]

// Positive lookbehind — match digits preceded by "$"
"$100 €200".match(/(?<=\$)\d+/g); // ["100"]

// Negative lookbehind
"$100 €200".match(/(?<!\$)\d+/g); // ["200"]
```

## Patterns

### 13. Escape user input for regex

```typescript
function escapeRegex(s: string): string {
  return s.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

const userInput = "price is $10.00";
const re = new RegExp(escapeRegex(userInput));
```

### 14. Common validation patterns

```typescript
const patterns = {
  email: /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/,
  uuid: /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i,
  ipv4: /^(?:(?:25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(?:25[0-5]|2[0-4]\d|[01]?\d\d?)$/,
  slug: /^[a-z0-9]+(?:-[a-z0-9]+)*$/,
  isoDate: /^\d{4}-\d{2}-\d{2}$/,
} as const;
```
