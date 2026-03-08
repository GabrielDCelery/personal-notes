# TypeScript Arrays & Maps

## Why

- **Arrays are reference types** — Assigning an array to another variable does not copy it. Both point to the same data. Use spread (`[...arr]`) or `structuredClone` for independent copies.
- **map/filter/reduce create new arrays** — They don't mutate the original. This is safe but means chaining `.filter().map()` allocates intermediate arrays. For large datasets, use a single `reduce` or a `for` loop.
- **sort mutates in place** — Unlike most array methods, `sort()` and `reverse()` modify the original array. Use `toSorted()` and `toReversed()` (ES2023) for non-mutating versions.
- **Map over plain objects for dynamic keys** — `Map` preserves insertion order, supports any key type, has O(1) `.size`, and doesn't collide with prototype properties. Use plain objects for static, known-key structures.
- **Set for uniqueness** — `Set` gives O(1) add/has/delete. Much faster than `array.includes()` for repeated membership checks.
- **Immutable methods (ES2023)** — `toSorted()`, `toReversed()`, `toSpliced()`, `with()` return new arrays instead of mutating. Use when you need the original intact.

## Quick Reference — Arrays

| Use case        | Method                               |
| --------------- | ------------------------------------ |
| Add to end      | `arr.push(val)` / `[...arr, val]`    |
| Add to start    | `arr.unshift(val)` / `[val, ...arr]` |
| Remove from end | `arr.pop()`                          |
| Remove by index | `arr.splice(i, 1)`                   |
| Find element    | `arr.find(fn)`                       |
| Find index      | `arr.findIndex(fn)`                  |
| Check existence | `arr.includes(val)` / `arr.some(fn)` |
| Transform       | `arr.map(fn)`                        |
| Filter          | `arr.filter(fn)`                     |
| Reduce          | `arr.reduce(fn, init)`               |
| Sort (mutating) | `arr.sort(fn)`                       |
| Sort (copy)     | `arr.toSorted(fn)`                   |
| Flatten         | `arr.flat(depth)`                    |
| Unique values   | `[...new Set(arr)]`                  |

## Quick Reference — Map & Set

| Use case    | Method                         |
| ----------- | ------------------------------ |
| Create Map  | `new Map<K, V>()`              |
| Set / get   | `map.set(k, v)` / `map.get(k)` |
| Check key   | `map.has(k)`                   |
| Delete key  | `map.delete(k)`                |
| Size        | `map.size`                     |
| Create Set  | `new Set<T>()`                 |
| Add / check | `set.add(v)` / `set.has(v)`    |
| Delete      | `set.delete(v)`                |

## Arrays

### 1. Creating arrays

```typescript
const empty: number[] = [];
const nums = [1, 2, 3];
const filled = Array.from({ length: 5 }, (_, i) => i); // [0, 1, 2, 3, 4]
const repeated = new Array(3).fill(0); // [0, 0, 0]
```

### 2. Adding and removing

```typescript
const arr = [1, 2, 3];

arr.push(4); // [1, 2, 3, 4] — add to end (mutates)
arr.pop(); // [1, 2, 3] — remove from end (mutates)
arr.unshift(0); // [0, 1, 2, 3] — add to start (mutates)
arr.shift(); // [1, 2, 3] — remove from start (mutates)
arr.splice(1, 1); // [1, 3] — remove 1 element at index 1 (mutates)

// Non-mutating alternatives
const added = [...arr, 4];
const removed = arr.filter((_, i) => i !== 1);
```

### 3. find, findIndex, includes

```typescript
const users = [
  { id: 1, name: "Alice" },
  { id: 2, name: "Bob" },
];

users.find((u) => u.id === 2); // { id: 2, name: "Bob" } | undefined
users.findIndex((u) => u.id === 2); // 1 (-1 if not found)
[1, 2, 3].includes(2); // true
users.some((u) => u.name === "Bob"); // true
users.every((u) => u.id > 0); // true
```

### 4. map, filter, reduce

```typescript
const nums = [1, 2, 3, 4, 5];

nums.map((n) => n * 2); // [2, 4, 6, 8, 10]
nums.filter((n) => n > 3); // [4, 5]
nums.reduce((sum, n) => sum + n, 0); // 15

// flatMap — map + flatten one level
[["a", "b"], ["c"]].flat(); // ["a", "b", "c"]
users.flatMap((u) => u.tags); // flatten nested arrays
```

### 5. Sorting

```typescript
// sort mutates the original — use toSorted for a copy
const nums = [3, 1, 2];
nums.sort((a, b) => a - b); // [1, 2, 3] — ascending (mutates)
nums.sort((a, b) => b - a); // [3, 2, 1] — descending (mutates)

// ES2023 — non-mutating
const sorted = nums.toSorted((a, b) => a - b); // new array

// Sort objects by field
users.toSorted((a, b) => a.name.localeCompare(b.name));
```

### 6. Copying and slicing

```typescript
const arr = [1, 2, 3, 4, 5];

const copy = [...arr]; // shallow copy
const deep = structuredClone(arr); // deep copy (nested objects)
const slice = arr.slice(1, 3); // [2, 3] — does not mutate
```

### 7. Unique values

```typescript
const arr = [1, 2, 2, 3, 3, 3];
const unique = [...new Set(arr)]; // [1, 2, 3]

// Unique objects by key
function uniqueBy<T>(arr: T[], key: keyof T): T[] {
  const seen = new Set();
  return arr.filter((item) => {
    const k = item[key];
    if (seen.has(k)) return false;
    seen.add(k);
    return true;
  });
}
```

### 8. Group by (ES2024 or manual)

```typescript
// Object.groupBy (ES2024 / Node 21+)
const grouped = Object.groupBy(users, (u) => u.role);
// { admin: [...], user: [...] }

// Manual fallback
function groupBy<T>(arr: T[], fn: (item: T) => string): Record<string, T[]> {
  return arr.reduce<Record<string, T[]>>((acc, item) => {
    const key = fn(item);
    (acc[key] ??= []).push(item);
    return acc;
  }, {});
}
```

## Map

### 9. Map basics

```typescript
const map = new Map<string, number>();

map.set("a", 1);
map.set("b", 2);
map.get("a"); // 1
map.has("a"); // true
map.delete("a");
map.size; // 1
map.clear();

// Initialize from entries
const map = new Map([
  ["a", 1],
  ["b", 2],
]);
```

### 10. Iterating a Map

```typescript
const map = new Map([
  ["a", 1],
  ["b", 2],
  ["c", 3],
]);

for (const [key, value] of map) {
  console.log(key, value);
}

const keys = [...map.keys()]; // ["a", "b", "c"]
const values = [...map.values()]; // [1, 2, 3]
const entries = [...map.entries()];

// Convert to plain object
const obj = Object.fromEntries(map);
```

## Set

### 11. Set basics

```typescript
const set = new Set<string>();

set.add("a");
set.add("b");
set.add("a"); // no-op, already exists
set.has("a"); // true
set.delete("a");
set.size; // 1

// From array
const set = new Set([1, 2, 3, 2, 1]); // Set {1, 2, 3}
```

### 12. Set operations

```typescript
const a = new Set([1, 2, 3]);
const b = new Set([2, 3, 4]);

// Union
const union = new Set([...a, ...b]); // {1, 2, 3, 4}

// Intersection
const intersection = new Set([...a].filter((x) => b.has(x))); // {2, 3}

// Difference
const difference = new Set([...a].filter((x) => !b.has(x))); // {1}
```

## Patterns

### 13. Convert between Map and object

```typescript
// Object → Map
const obj = { a: 1, b: 2 };
const map = new Map(Object.entries(obj));

// Map → Object
const back = Object.fromEntries(map);
```

### 14. Frequency counter

```typescript
function countFrequency<T>(arr: T[]): Map<T, number> {
  const freq = new Map<T, number>();
  for (const item of arr) {
    freq.set(item, (freq.get(item) ?? 0) + 1);
  }
  return freq;
}

countFrequency(["a", "b", "a", "c", "a"]); // Map { "a" => 3, "b" => 1, "c" => 1 }
```
