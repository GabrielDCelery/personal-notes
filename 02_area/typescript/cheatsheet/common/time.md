# TypeScript Time & Dates

## Why

- **Date is mutable** — Calling `setHours()` modifies the original object. This causes bugs when you pass dates around. Always create new Date objects instead of mutating, or use `Temporal` (when stable).
- **Month is 0-indexed** — `new Date(2024, 0, 15)` is January 15th. `new Date(2024, 11, 25)` is December 25th. This is a constant source of bugs. Use ISO strings or explicit month values.
- **Date.now() for timestamps** — Returns milliseconds since epoch as a number. Faster than `new Date().getTime()` when you just need a timestamp, not a Date object.
- **toISOString for serialization** — Always use ISO 8601 format (`2024-01-15T14:30:00.000Z`) for APIs and storage. It's unambiguous, sortable, and timezone-aware.
- **Intl.DateTimeFormat for display** — Don't build format strings manually. Intl handles locale-specific formatting (date order, month names, AM/PM) correctly.

## Quick Reference

| Use case           | Method                             |
| ------------------ | ---------------------------------- |
| Current time       | `new Date()` or `Date.now()`       |
| From ISO string    | `new Date("2024-01-15T14:30:00Z")` |
| To ISO string      | `date.toISOString()`               |
| Unix timestamp ms  | `Date.now()` or `date.getTime()`   |
| Unix timestamp s   | `Math.floor(Date.now() / 1000)`    |
| From Unix ms       | `new Date(ms)`                     |
| From Unix s        | `new Date(seconds * 1000)`         |
| Format for display | `Intl.DateTimeFormat`              |
| Add time           | manual arithmetic on ms            |
| Compare            | `date.getTime()` comparison        |

## Creating Dates

### 1. Current time

```typescript
const now = new Date();
const timestamp = Date.now(); // ms since epoch — no object allocation
```

### 2. From strings

```typescript
new Date("2024-01-15"); // midnight UTC
new Date("2024-01-15T14:30:00Z"); // explicit UTC
new Date("2024-01-15T14:30:00+02:00"); // with timezone offset
```

### 3. From components (beware: month is 0-indexed)

```typescript
new Date(2024, 0, 15); // Jan 15, 2024 — local time
new Date(2024, 0, 15, 14, 30, 0); // Jan 15, 2024 14:30:00 — local time

// Safer: use Date.UTC for explicit UTC
new Date(Date.UTC(2024, 0, 15, 14, 30)); // UTC
```

### 4. From Unix timestamp

```typescript
new Date(1705312200000); // from milliseconds
new Date(1705312200 * 1000); // from seconds
```

## Formatting

### 5. ISO string (APIs and storage)

```typescript
const date = new Date();
date.toISOString(); // "2024-01-15T14:30:00.000Z" — always UTC
```

### 6. Intl.DateTimeFormat (display)

```typescript
const date = new Date("2024-01-15T14:30:00Z");

new Intl.DateTimeFormat("en-GB").format(date);
// "15/01/2024"

new Intl.DateTimeFormat("en-US", {
  dateStyle: "medium",
  timeStyle: "short",
}).format(date);
// "Jan 15, 2024, 2:30 PM"

new Intl.DateTimeFormat("en-US", {
  year: "numeric",
  month: "2-digit",
  day: "2-digit",
}).format(date);
// "01/15/2024"
```

### 7. Quick format helpers

```typescript
function formatDate(date: Date): string {
  return date.toISOString().split("T")[0]; // "2024-01-15"
}

function formatDateTime(date: Date): string {
  return date.toISOString().replace("T", " ").slice(0, 19); // "2024-01-15 14:30:00"
}
```

## Extracting Components

### 8. Get date and time parts

```typescript
const d = new Date("2024-01-15T14:30:45.123Z");

d.getFullYear(); // 2024
d.getMonth(); // 0 (January — 0-indexed!)
d.getDate(); // 15
d.getDay(); // 1 (Monday — 0=Sunday)
d.getHours(); // local hours
d.getMinutes(); // local minutes
d.getSeconds(); // local seconds
d.getMilliseconds(); // 123
d.getTime(); // 1705328445123 (ms since epoch)

// UTC variants
d.getUTCHours(); // 14
d.getUTCMonth(); // 0
```

## Date Arithmetic

### 9. Add/subtract time

```typescript
const now = new Date();

// Add hours
const later = new Date(now.getTime() + 2 * 60 * 60 * 1000);

// Helper for common operations
function addDays(date: Date, days: number): Date {
  return new Date(date.getTime() + days * 24 * 60 * 60 * 1000);
}

function addHours(date: Date, hours: number): Date {
  return new Date(date.getTime() + hours * 60 * 60 * 1000);
}

const tomorrow = addDays(now, 1);
const yesterday = addDays(now, -1);
```

### 10. Difference between dates

```typescript
const start = new Date("2024-01-15");
const end = new Date("2024-03-20");

const diffMs = end.getTime() - start.getTime();
const diffDays = Math.floor(diffMs / (24 * 60 * 60 * 1000)); // 65
const diffHours = Math.floor(diffMs / (60 * 60 * 1000));
```

## Comparing Dates

### 11. Compare two dates

```typescript
const a = new Date("2024-01-15");
const b = new Date("2024-03-20");

a < b; // true
a.getTime() === b.getTime(); // exact equality
a.getTime() > b.getTime(); // after

// Sort dates
const dates = [b, a];
dates.sort((x, y) => x.getTime() - y.getTime()); // ascending
```

## Measuring Elapsed Time

### 12. Measure execution time

```typescript
// Simple — millisecond precision
const start = Date.now();
await doWork();
console.log(`Took ${Date.now() - start}ms`);

// Precise — sub-millisecond
const start = performance.now();
await doWork();
console.log(`Took ${(performance.now() - start).toFixed(2)}ms`);
```

## Timers

### 13. setTimeout and setInterval

```typescript
// One-shot
const timer = setTimeout(() => console.log("fired"), 5000);
clearTimeout(timer); // cancel

// Repeating
const interval = setInterval(() => console.log("tick"), 1000);
clearInterval(interval); // stop

// Promise-based (Node.js)
import { setTimeout } from "node:timers/promises";
await setTimeout(1000); // sleep 1 second
```

### 14. setInterval with cleanup

```typescript
function startPolling(fn: () => Promise<void>, intervalMs: number): () => void {
  const id = setInterval(() => fn().catch(console.error), intervalMs);
  return () => clearInterval(id); // return cleanup function
}

const stopPolling = startPolling(checkHealth, 30_000);
// later...
stopPolling();
```

## Unix Timestamps

### 15. Convert to/from Unix

```typescript
// To Unix
const ms = Date.now(); // milliseconds
const seconds = Math.floor(Date.now() / 1000); // seconds

// From Unix
const fromMs = new Date(1705312200000);
const fromSec = new Date(1705312200 * 1000);
```
