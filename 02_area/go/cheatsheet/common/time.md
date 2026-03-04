# Go Time

> Go uses a reference time for formatting: `Mon Jan 2 15:04:05 MST 2006` (1/2 3:04:05 PM 2006, Unix time 1136239445).

## Quick Reference

| Use case         | Method                                    |
| ---------------- | ----------------------------------------- |
| Current time     | `time.Now()`                              |
| Format           | `t.Format("2006-01-02")`                  |
| Parse            | `time.Parse("2006-01-02", s)`             |
| Duration         | `5 * time.Second`                         |
| Add duration     | `t.Add(2 * time.Hour)`                    |
| Difference       | `t2.Sub(t1)`                              |
| Sleep            | `time.Sleep(100 * time.Millisecond)`      |
| Timer (once)     | `time.NewTimer(5 * time.Second)`          |
| Ticker (repeat)  | `time.NewTicker(1 * time.Second)`         |
| Timeout context  | `context.WithTimeout(ctx, 5*time.Second)` |

## Formatting and Parsing

### 1. Format a time

```go
now := time.Now()

now.Format("2006-01-02")                  // "2025-03-04"
now.Format("2006-01-02 15:04:05")         // "2025-03-04 14:30:00"
now.Format(time.RFC3339)                  // "2025-03-04T14:30:00Z"
now.Format("02 Jan 2006")                 // "04 Mar 2025"
now.Format("3:04 PM")                     // "2:30 PM"
```

### 2. Common format constants

```go
time.RFC3339      // "2006-01-02T15:04:05Z07:00"
time.RFC3339Nano  // "2006-01-02T15:04:05.999999999Z07:00"
time.DateTime     // "2006-01-02 15:04:05"
time.DateOnly     // "2006-01-02"
time.TimeOnly     // "15:04:05"
```

### 3. Parse a string to time

```go
t, err := time.Parse("2006-01-02", "2025-03-04")
t, err := time.Parse(time.RFC3339, "2025-03-04T14:30:00Z")

// Parse in a specific timezone
loc, _ := time.LoadLocation("Europe/London")
t, err := time.ParseInLocation("2006-01-02 15:04", "2025-03-04 14:30", loc)
```

## Duration and Arithmetic

### 4. Create durations

```go
d := 5 * time.Second
d := 2*time.Hour + 30*time.Minute
d := time.Duration(n) * time.Millisecond  // from variable
```

### 5. Duration methods

```go
d := 90 * time.Second
d.Seconds()     // 90.0
d.Minutes()     // 1.5
d.String()      // "1m30s"
d.Milliseconds() // 90000
```

### 6. Add and subtract time

```go
now := time.Now()

future := now.Add(24 * time.Hour)         // add duration
past := now.Add(-2 * time.Hour)           // subtract duration
nextMonth := now.AddDate(0, 1, 0)         // add years, months, days
lastYear := now.AddDate(-1, 0, 0)
```

### 7. Difference between times

```go
elapsed := time.Since(start)              // time.Now().Sub(start)
remaining := time.Until(deadline)         // deadline.Sub(time.Now())
diff := t2.Sub(t1)                        // arbitrary difference
```

## Comparing Times

### 8. Compare two times

```go
t1.Before(t2)   // true if t1 < t2
t1.After(t2)    // true if t1 > t2
t1.Equal(t2)    // true if same instant (handles different timezones)
t1.IsZero()     // true if zero value
```

## Time Zones

### 9. Work with locations

```go
utc := time.Now().UTC()

loc, err := time.LoadLocation("America/New_York")
nyTime := utc.In(loc)

_, offset := nyTime.Zone()  // "EST", -18000
```

## Extract Components

### 10. Get date and time parts

```go
t := time.Now()

t.Year()        // 2025
t.Month()       // time.March
t.Day()         // 4
t.Hour()        // 14
t.Minute()      // 30
t.Second()      // 0
t.Weekday()     // time.Tuesday
t.YearDay()     // 63

year, month, day := t.Date()
hour, min, sec := t.Clock()
```

## Timers and Tickers

### 11. Timer (fires once)

```go
timer := time.NewTimer(5 * time.Second)
<-timer.C  // blocks until timer fires

// Cancel before it fires
if !timer.Stop() {
    <-timer.C  // drain channel if already fired
}
```

### 12. Ticker (fires repeatedly)

```go
ticker := time.NewTicker(1 * time.Second)
defer ticker.Stop()

for range ticker.C {
    fmt.Println("tick")
}
```

### 13. Ticker with done channel

```go
ticker := time.NewTicker(500 * time.Millisecond)
defer ticker.Stop()

for {
    select {
    case <-ticker.C:
        fmt.Println("tick")
    case <-ctx.Done():
        return
    }
}
```

## Measuring Elapsed Time

### 14. Benchmark a block of code

```go
start := time.Now()
doWork()
slog.Info("completed", "took", time.Since(start))
```

## Unix Timestamps

### 15. Convert to/from Unix

```go
ts := time.Now().Unix()          // seconds
tsMs := time.Now().UnixMilli()   // milliseconds
tsNs := time.Now().UnixNano()    // nanoseconds

t := time.Unix(ts, 0)            // from seconds
t := time.UnixMilli(tsMs)        // from milliseconds
```
