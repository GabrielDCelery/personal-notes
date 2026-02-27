# Lesson 16: database/sql Patterns

Go's `database/sql` package is a database-agnostic abstraction over SQL drivers. It handles connection pooling, prepared statements, and transactions — but it exposes enough rope to hang yourself with. Connection pool misconfiguration causes exhaustion under load. Missing `rows.Close()` leaks connections silently. Transaction rollback patterns that look correct can still leave transactions open. These are the things interviewers probe because they're the things that break in production.

## `sql.DB` Is a Pool, Not a Connection

This is the foundational misunderstanding. `sql.DB` manages a pool of underlying database connections — it is not a single connection.

```go
// sql.Open validates the driver name and DSN format, but does NOT connect
db, err := sql.Open("postgres", "postgres://user:pass@localhost/mydb?sslmode=disable")
if err != nil {
    return err   // DSN parse error, not a connection error
}
defer db.Close()

// ✓ Verify connectivity explicitly
if err := db.PingContext(ctx); err != nil {
    return fmt.Errorf("db unreachable: %w", err)
}
```

### Pool Configuration

```go
db.SetMaxOpenConns(25)          // max total connections (open or idle)
db.SetMaxIdleConns(10)          // max connections kept idle in pool
db.SetConnMaxLifetime(5 * time.Minute)   // close connections older than this
db.SetConnMaxIdleTime(1 * time.Minute)   // close idle connections older than this
```

| Setting           | Too Low                                           | Too High                                             |
| ----------------- | ------------------------------------------------- | ---------------------------------------------------- |
| `MaxOpenConns`    | Requests queue, latency spikes                    | Exceeds DB server connection limit                   |
| `MaxIdleConns`    | New connections opened per request (TCP overhead) | Idle connections waste DB resources                  |
| `ConnMaxLifetime` | High churn, frequent reconnects                   | Stale connections after DB restart, firewall timeout |
| `ConnMaxIdleTime` | —                                                 | Idle connections kept past firewall/LB timeout       |

**Gotcha: `MaxIdleConns > MaxOpenConns`**

```go
db.SetMaxOpenConns(5)
db.SetMaxIdleConns(10)   // ❌ capped to MaxOpenConns automatically — idle setting is silently ignored
```

**Gotcha: zero means unlimited**

```go
db.SetMaxOpenConns(0)   // ❌ unlimited open connections — DB server gets overwhelmed under load
```

Always set `MaxOpenConns` explicitly. A reasonable starting point for a web service: `MaxOpenConns(25)`, `MaxIdleConns(10)`, `ConnMaxLifetime(5*time.Minute)`.

---

## Query Patterns

Always use context variants — they respect cancellation and deadlines.

### `QueryContext` — Multiple Rows

```go
rows, err := db.QueryContext(ctx, "SELECT id, name FROM users WHERE active = $1", true)
if err != nil {
    return err
}
defer rows.Close()   // ✓ must close — releases connection back to pool

var users []User
for rows.Next() {
    var u User
    if err := rows.Scan(&u.ID, &u.Name); err != nil {
        return err
    }
    users = append(users, u)
}

// ✓ Must check rows.Err() — errors during iteration surface here, not in Next()
if err := rows.Err(); err != nil {
    return err
}
```

### `QueryRowContext` — Single Row

```go
var u User
err := db.QueryRowContext(ctx,
    "SELECT id, name FROM users WHERE id = $1", id,
).Scan(&u.ID, &u.Name)

if err == sql.ErrNoRows {
    return nil, ErrNotFound   // ✓ no row is not a fatal error
}
if err != nil {
    return nil, err
}
```

`QueryRow` returns a `*sql.Row` — the error (if any) is deferred to `Scan`. There is no `rows.Close()` needed; the connection is released after `Scan`.

### `ExecContext` — Writes

```go
result, err := db.ExecContext(ctx,
    "UPDATE users SET name = $1 WHERE id = $2", name, id,
)
if err != nil {
    return err
}

affected, err := result.RowsAffected()
if err != nil {
    return err
}
if affected == 0 {
    return ErrNotFound   // ✓ distinguish "not found" from "error"
}
```

---

## `sql.Rows` Gotchas

### Close After Error Check

```go
// ❌ Wrong — if QueryContext returns an error, rows is nil; defer rows.Close() panics
rows, err := db.QueryContext(ctx, query)
defer rows.Close()
if err != nil {
    return err
}

// ✓ Correct — defer after the error check
rows, err := db.QueryContext(ctx, query)
if err != nil {
    return err
}
defer rows.Close()
```

### `rows.Err()` Is Not Optional

```go
for rows.Next() {
    // iterate
}
// ✓ Always check — the loop exits on error OR on EOF; only rows.Err() distinguishes them
if err := rows.Err(); err != nil {
    return err
}
```

If the database connection drops mid-iteration, `rows.Next()` returns `false` (loop exits normally) but `rows.Err()` returns the connection error. Skipping this check silently returns partial results.

### Connection Leak Without Close

```go
func getUsers(db *sql.DB) ([]User, error) {
    rows, err := db.QueryContext(context.Background(), "SELECT id FROM users")
    if err != nil {
        return nil, err
    }
    // ❌ No defer rows.Close() — if function returns early (error in Scan),
    // rows are never closed, connection is never returned to pool
    var users []User
    for rows.Next() {
        var u User
        if err := rows.Scan(&u.ID); err != nil {
            return nil, err   // rows leaked
        }
        users = append(users, u)
    }
    return users, rows.Err()
}
```

---

## Transactions

### The Correct Pattern

```go
func transfer(ctx context.Context, db *sql.DB, fromID, toID int64, amount int) (err error) {
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    // ✓ Deferred rollback: no-op if Commit already succeeded
    defer func() {
        if rbErr := tx.Rollback(); rbErr != nil && rbErr != sql.ErrTxDone {
            err = fmt.Errorf("rollback failed: %w (original: %v)", rbErr, err)
        }
    }()

    if _, err = tx.ExecContext(ctx,
        "UPDATE accounts SET balance = balance - $1 WHERE id = $2", amount, fromID,
    ); err != nil {
        return err   // defer calls Rollback
    }

    if _, err = tx.ExecContext(ctx,
        "UPDATE accounts SET balance = balance + $1 WHERE id = $2", amount, toID,
    ); err != nil {
        return err   // defer calls Rollback
    }

    return tx.Commit()   // if Commit succeeds, Rollback is a no-op (returns sql.ErrTxDone)
}
```

**Why `sql.ErrTxDone` in the defer**: after `Commit()` succeeds, the transaction is done. Calling `Rollback()` returns `sql.ErrTxDone` — this is expected and should not be treated as an error.

### `sql.TxOptions`

```go
tx, err := db.BeginTx(ctx, &sql.TxOptions{
    Isolation: sql.LevelSerializable,   // or LevelRepeatableRead, LevelReadCommitted, etc.
    ReadOnly:  true,                    // advisory hint to the database
})
```

### Transactions Use a Dedicated Connection

A `sql.Tx` holds a single dedicated connection for its lifetime. All statements within the transaction use that connection — they are not distributed across the pool. This means:

```go
// ❌ Using db (the pool) inside a transaction uses a DIFFERENT connection — not part of tx
func doWork(ctx context.Context, db *sql.DB, tx *sql.Tx) error {
    tx.ExecContext(ctx, "INSERT ...")
    db.QueryContext(ctx, "SELECT ...")   // ❌ different connection, outside the transaction
    return nil
}

// ✓ Use tx for all work within the transaction
func doWork(ctx context.Context, tx *sql.Tx) error {
    tx.ExecContext(ctx, "INSERT ...")
    tx.QueryContext(ctx, "SELECT ...")   // ✓ same connection, within the transaction
    return nil
}
```

---

## Prepared Statements

`sql.Stmt` represents a server-side prepared statement — the query is parsed once, then executed multiple times with different parameters.

```go
stmt, err := db.PrepareContext(ctx, "SELECT id, name FROM users WHERE email = $1")
if err != nil {
    return err
}
defer stmt.Close()   // ✓ returns statement to pool

// Reuse across calls
for _, email := range emails {
    var u User
    err := stmt.QueryRowContext(ctx, email).Scan(&u.ID, &u.Name)
    // ...
}
```

**When prepared statements help**:

- The same query is executed many times in a tight loop
- The database supports efficient prepared statement caching (PostgreSQL, MySQL)

**When they add overhead**:

- One-off queries — prepare + execute is two round trips vs one
- Dynamic queries where the structure changes per call

**`sql.DB` already caches prepared statements** for `Query`/`Exec` calls internally (since Go 1.1), so explicit `Prepare` is mainly useful for very hot loops where you want full control.

---

## `sql.Null*` Types

SQL columns that allow NULL cannot be scanned into plain Go types — `Scan` returns an error.

```go
// ❌ Panics if deleted_at is NULL in the database
var deletedAt time.Time
rows.Scan(&deletedAt)

// ✓ Option 1: sql.Null* types
var deletedAt sql.NullTime
rows.Scan(&deletedAt)
if deletedAt.Valid {
    fmt.Println(deletedAt.Time)
}

// ✓ Option 2: pointer types (often cleaner in structs)
var deletedAt *time.Time
rows.Scan(&deletedAt)
if deletedAt != nil {
    fmt.Println(*deletedAt)
}
```

### Available `sql.Null*` Types

```go
sql.NullString  // {String string; Valid bool}
sql.NullInt64   // {Int64 int64; Valid bool}
sql.NullInt32   // {Int32 int32; Valid bool}
sql.NullFloat64 // {Float64 float64; Valid bool}
sql.NullBool    // {Bool bool; Valid bool}
sql.NullTime    // {Time time.Time; Valid bool}
sql.NullByte    // {Byte byte; Valid bool}
```

**Pointer types vs `sql.Null*`**: Pointer types (`*string`, `*int64`) are often cleaner in domain structs — they marshal to `null` in JSON natively, whereas `sql.NullString` marshals to `{"String":"...","Valid":true}` by default unless you add custom JSON marshaling.

---

## Context Cancellation

`database/sql` passes context to the driver, which can cancel in-flight queries when the context is cancelled:

```go
func getUser(ctx context.Context, db *sql.DB, id int64) (*User, error) {
    var u User
    err := db.QueryRowContext(ctx,   // ✓ ctx cancelled → query cancelled at DB level
        "SELECT id, name FROM users WHERE id = $1", id,
    ).Scan(&u.ID, &u.Name)
    if err != nil {
        if ctx.Err() != nil {
            return nil, fmt.Errorf("query cancelled: %w", ctx.Err())
        }
        if err == sql.ErrNoRows {
            return nil, ErrNotFound
        }
        return nil, err
    }
    return &u, nil
}
```

**What happens on cancellation**: the driver sends a query cancellation to the database (e.g., `pg_cancel_backend` in PostgreSQL). The `QueryRowContext` call returns an error. The connection is returned to the pool (not discarded) if it's still healthy.

**In HTTP handlers**: always pass `r.Context()` to database calls. When the client disconnects, the HTTP server cancels the request context, which propagates to the database query — long-running queries don't outlive their usefulness.

---

## Hands-On Exercise 1: Transaction Pattern

The following transaction function has bugs that can leave the transaction open. Identify and fix them.

```go
func createOrder(ctx context.Context, db *sql.DB, userID int64, items []Item) error {
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }

    var orderID int64
    err = tx.QueryRowContext(ctx,
        "INSERT INTO orders (user_id) VALUES ($1) RETURNING id", userID,
    ).Scan(&orderID)
    if err != nil {
        tx.Rollback()
        return err
    }

    for _, item := range items {
        _, err = tx.ExecContext(ctx,
            "INSERT INTO order_items (order_id, product_id, qty) VALUES ($1, $2, $3)",
            orderID, item.ProductID, item.Qty,
        )
        if err != nil {
            tx.Rollback()
            return err
        }
    }

    return tx.Commit()
}
```

<details>
<summary>Solution</summary>

**Issues**:

1. ❌ `tx.Rollback()` errors are ignored — if rollback fails, the transaction may be left open
2. ❌ No `defer tx.Rollback()` — if a future code path is added that returns without an explicit rollback, the transaction leaks
3. ❌ Each error path calls `Rollback()` manually — fragile, easy to miss a return path

**Fixed**:

```go
func createOrder(ctx context.Context, db *sql.DB, userID int64, items []Item) (err error) {
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    // ✓ Single deferred rollback covers all return paths
    defer func() {
        if rbErr := tx.Rollback(); rbErr != nil && rbErr != sql.ErrTxDone {
            err = fmt.Errorf("rollback: %w (original: %v)", rbErr, err)
        }
    }()

    var orderID int64
    if err = tx.QueryRowContext(ctx,
        "INSERT INTO orders (user_id) VALUES ($1) RETURNING id", userID,
    ).Scan(&orderID); err != nil {
        return err   // defer handles rollback
    }

    for _, item := range items {
        if _, err = tx.ExecContext(ctx,
            "INSERT INTO order_items (order_id, product_id, qty) VALUES ($1, $2, $3)",
            orderID, item.ProductID, item.Qty,
        ); err != nil {
            return err   // defer handles rollback
        }
    }

    return tx.Commit()   // on success, deferred Rollback returns sql.ErrTxDone (ignored)
}
```

**Named return `err`**: the deferred function captures the named return value, so if `tx.Rollback()` itself fails and `err` is non-nil (a prior error), we can wrap both errors.

</details>

## Hands-On Exercise 2: Connection Pool Misconfiguration

Review this server setup and identify the pool configuration problems:

```go
func main() {
    db, _ := sql.Open("postgres", dsn)

    db.SetMaxIdleConns(100)
    db.SetConnMaxLifetime(0)

    http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
        rows, err := db.QueryContext(r.Context(), "SELECT id, name FROM users")
        if err != nil {
            http.Error(w, err.Error(), 500)
            return
        }
        // ... process rows
    })

    http.ListenAndServe(":8080", nil)
}
```

<details>
<summary>Solution</summary>

**Issues**:

1. ❌ `sql.Open` error ignored — a bad DSN fails silently; the server starts but every request will fail
2. ❌ No `db.PingContext` — connectivity is never verified at startup
3. ❌ `MaxOpenConns` not set (defaults to 0 = unlimited) — under load, hundreds of connections could be opened, overwhelming the database
4. ❌ `MaxIdleConns(100)` is set, but with unlimited `MaxOpenConns`, `MaxIdleConns` is effectively uncapped in proportion — you'd keep 100 idle connections permanently
5. ❌ `ConnMaxLifetime(0)` means connections are never recycled — stale connections after a database restart, firewall timeout, or HA failover will produce errors rather than reconnecting
6. ❌ `rows.Close()` is missing in the handler — connections leak back to the pool
7. ❌ `http.ListenAndServe` error ignored and uses `DefaultServeMux` with no timeouts

**Fixed**:

```go
func main() {
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        log.Fatalf("open db: %v", err)
    }
    defer db.Close()

    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(10)
    db.SetConnMaxLifetime(5 * time.Minute)
    db.SetConnMaxIdleTime(1 * time.Minute)

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := db.PingContext(ctx); err != nil {
        log.Fatalf("ping db: %v", err)
    }

    mux := http.NewServeMux()
    mux.HandleFunc("GET /users", func(w http.ResponseWriter, r *http.Request) {
        rows, err := db.QueryContext(r.Context(), "SELECT id, name FROM users")
        if err != nil {
            http.Error(w, err.Error(), 500)
            return
        }
        defer rows.Close()   // ✓ always close
        // ... process rows
        if err := rows.Err(); err != nil {
            http.Error(w, err.Error(), 500)
        }
    })

    srv := &http.Server{
        Addr:              ":8080",
        Handler:           mux,
        ReadHeaderTimeout: 2 * time.Second,
        WriteTimeout:      10 * time.Second,
        IdleTimeout:       120 * time.Second,
    }
    log.Fatal(srv.ListenAndServe())
}
```

</details>

---

## Interview Questions

### Q1: Why is `sql.DB` safe to use concurrently, and what does it actually represent?

Tests the fundamental mental model. Developers who think `sql.DB` is a single connection make systematic errors — sharing it across goroutines thinking they're sharing one connection, or creating a new `sql.DB` per request.

<details>
<summary>Answer</summary>

`sql.DB` is a connection **pool**, not a connection. It manages a set of underlying database connections and distributes them across concurrent callers.

It is safe for concurrent use because:

- Each call to `QueryContext`, `ExecContext`, etc. acquires a connection from the pool, uses it, and returns it when done
- The pool itself is protected by internal synchronisation (`sync.Mutex`)
- Multiple goroutines can call methods on the same `sql.DB` simultaneously — each gets its own connection

**The correct pattern**: create one `sql.DB` at application startup and share it everywhere (pass it to handlers, repositories, etc.). Do not create a new `sql.DB` per request — that defeats connection pooling entirely.

```go
// ✓ One pool, shared across the application
db, _ := sql.Open("postgres", dsn)
server := &Server{db: db}
// All handlers share server.db
```

`sql.Open` does not open any connections — it configures the pool. Connections are opened lazily on demand and returned to the pool after use.

</details>

### Q2: What happens if you forget to call `rows.Close()`, and how does `defer` help?

A practical question about resource management — one of the most common Go `database/sql` bugs in production codebases.

<details>
<summary>Answer</summary>

`sql.Rows` holds a connection from the pool while it's open. If `rows.Close()` is never called, that connection is never returned to the pool.

In a long-running service, each leaked rows object means one permanently unavailable connection. With `SetMaxOpenConns(25)`, 25 concurrent leaks would exhaust the pool — subsequent requests block waiting for a connection, causing request timeouts.

The leak can happen when:

- An early return (on error) skips the explicit `rows.Close()` call
- A panic unwinds the stack before `Close()` is reached

**The fix**: always `defer rows.Close()` immediately after checking the error from `QueryContext`:

```go
rows, err := db.QueryContext(ctx, query)
if err != nil {
    return err
}
defer rows.Close()   // ✓ guaranteed to run regardless of how the function exits
```

`defer rows.Close()` after the error check ensures:

- If `err != nil`, rows is nil and we return before defer is registered
- If the function returns early due to a scan error, `defer` still fires
- If a panic occurs, `defer` still fires (during stack unwinding)

Note: `sql.Row` (from `QueryRowContext`) does not require `Close()` — the connection is released when `Scan` is called.

</details>

### Q3: Describe the correct pattern for database transactions in Go, and why is `defer tx.Rollback()` safe to call even after a successful `Commit`?

A code correctness question — transaction handling is consistently one of the most bug-prone areas in Go database code.

<details>
<summary>Answer</summary>

The canonical pattern:

```go
func doTransaction(ctx context.Context, db *sql.DB) (err error) {
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer func() {
        if rbErr := tx.Rollback(); rbErr != nil && rbErr != sql.ErrTxDone {
            err = fmt.Errorf("rollback failed: %w", rbErr)
        }
    }()

    // ... do work using tx ...

    return tx.Commit()
}
```

**Why `defer tx.Rollback()` is safe after `Commit`**:

After `Commit()` succeeds, the transaction is marked as done internally. Any subsequent call to `Rollback()` returns `sql.ErrTxDone`. The deferred function checks for this and ignores it — so the defer is a true no-op after a successful commit.

After `Commit()` fails (network error, constraint violation), the transaction may or may not have committed at the database. `Rollback()` attempts to clean up. If it fails too, the named return captures both errors.

**Why not just call `tx.Rollback()` at each error site**:

- Easy to forget a return path, leaving the transaction open
- Every new code path added to the function requires a manual rollback
- The defer approach covers all exit paths — current and future — with one declaration

</details>

### Q4: How does context cancellation interact with in-flight SQL queries?

Tests understanding of the context propagation model — critical for building services that respond correctly to client disconnects and timeouts.

<details>
<summary>Answer</summary>

When a context passed to `QueryContext`, `ExecContext`, or `BeginTx` is cancelled or its deadline expires, the `database/sql` package signals the driver to cancel the in-flight operation.

What happens at the driver level varies:

- **PostgreSQL (`lib/pq`, `pgx`)**: sends a cancellation request to the server (`pg_cancel_backend`). The query is aborted server-side, freeing DB resources.
- **MySQL**: closes the underlying connection (no graceful query cancel protocol).
- **SQLite**: checks for cancellation between statements.

From the Go side:

- The `QueryContext` / `ExecContext` call returns with a context error (`context.Canceled` or `context.DeadlineExceeded`)
- The connection is returned to the pool if it's still healthy; otherwise it's discarded
- If a transaction was in progress, it is automatically rolled back

**Practical implications**:

```go
// HTTP handler — passes r.Context() through to DB
func handler(w http.ResponseWriter, r *http.Request) {
    result, err := db.QueryContext(r.Context(), "SELECT ...")
    if err != nil {
        if r.Context().Err() != nil {
            // Client disconnected — no point writing a response
            return
        }
        http.Error(w, "db error", 500)
        return
    }
    // ...
}
```

When the HTTP client disconnects, the Go HTTP server cancels `r.Context()`. The in-flight DB query is cancelled, the goroutine handling the request exits cleanly, and the DB connection is returned to the pool. Without context propagation, the query would run to completion even though no client is waiting for the result.

</details>

---

## Key Takeaways

1. **`sql.DB` is a pool**: create one at startup, share it everywhere — never create one per request.
2. **`sql.Open` doesn't connect**: use `db.PingContext` at startup to verify connectivity.
3. **Always set `MaxOpenConns`**: the default (0 = unlimited) allows unbounded connections under load.
4. **`defer rows.Close()` after the error check**: always, without exception — leaking rows leaks connections.
5. **Check `rows.Err()` after the loop**: errors during iteration appear here, not in `rows.Next()`.
6. **Deferred `tx.Rollback()`**: one defer handles all exit paths; it's a no-op after successful commit (returns `sql.ErrTxDone`).
7. **Use `tx` for all work within a transaction**: queries on `db` (the pool) use different connections and are not part of the transaction.
8. **`sql.Null*` or pointer types for nullable columns**: scanning NULL into a plain Go type returns an error.
9. **Always pass context**: use `QueryContext`, `ExecContext`, `BeginTx` — never the non-context variants in production code.
10. **Context cancellation propagates to the DB**: passing `r.Context()` to queries ensures client disconnects cancel in-flight work server-side.

## Next Steps

This lesson completes the expanded Go refresher series. For further depth:

- [Go Channels Deep Dive](../channels/) — 23-lesson series covering channel patterns exhaustively
- [Go Primitives](../primitives/) — foundational Go data types and their internals
- [`database/sql` documentation](https://pkg.go.dev/database/sql) — full API reference
- [`pgx`](https://github.com/jackc/pgx) — a more featureful PostgreSQL driver with richer context and type support than `lib/pq`
- [sqlx](https://github.com/jmoiron/sqlx) — extends `database/sql` with named queries, struct scanning, and `In` clauses
