# Go Database (database/sql)

## Quick Reference

| Use case                  | Method                                      |
| ------------------------- | ------------------------------------------- |
| Open connection           | `sql.Open` + `db.Ping`                      |
| Query multiple rows       | `db.QueryContext` + `rows.Scan`             |
| Query single row          | `db.QueryRowContext` + `row.Scan`           |
| Execute (insert/update)   | `db.ExecContext`                            |
| Transaction               | `db.BeginTx` + `tx.Commit/Rollback`        |
| Prepared statement        | `db.PrepareContext`                         |
| Check no rows             | `errors.Is(err, sql.ErrNoRows)`             |

## Setup

### 1. Open and verify connection

```go
import (
    "database/sql"
    _ "github.com/lib/pq" // postgres driver
)

db, err := sql.Open("postgres", "postgres://user:pass@localhost/dbname?sslmode=disable")
if err != nil {
    return err
}
defer db.Close()

if err := db.PingContext(ctx); err != nil {
    return err
}
```

### 2. Connection pool settings

```go
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(25)
db.SetConnMaxLifetime(5 * time.Minute)
```

## Querying

### 3. Query multiple rows

```go
rows, err := db.QueryContext(ctx, "SELECT id, name FROM users WHERE active = $1", true)
if err != nil {
    return err
}
defer rows.Close()

var users []User
for rows.Next() {
    var u User
    if err := rows.Scan(&u.ID, &u.Name); err != nil {
        return err
    }
    users = append(users, u)
}

if err := rows.Err(); err != nil {
    return err
}
```

### 4. Query single row

```go
var u User
err := db.QueryRowContext(ctx, "SELECT id, name FROM users WHERE id = $1", id).
    Scan(&u.ID, &u.Name)

if errors.Is(err, sql.ErrNoRows) {
    return nil, ErrNotFound
}
if err != nil {
    return nil, err
}
```

## Writing

### 5. Insert / Update / Delete

```go
result, err := db.ExecContext(ctx,
    "INSERT INTO users (name, email) VALUES ($1, $2)",
    "Alice", "alice@example.com",
)
if err != nil {
    return err
}

rowsAffected, _ := result.RowsAffected()
```

### 6. Insert and get returned ID (Postgres)

```go
var id int
err := db.QueryRowContext(ctx,
    "INSERT INTO users (name) VALUES ($1) RETURNING id",
    "Alice",
).Scan(&id)
```

## Transactions

### 7. Transaction with rollback on error

```go
tx, err := db.BeginTx(ctx, nil)
if err != nil {
    return err
}
defer tx.Rollback() // no-op if already committed

_, err = tx.ExecContext(ctx, "UPDATE accounts SET balance = balance - $1 WHERE id = $2", amount, fromID)
if err != nil {
    return err
}

_, err = tx.ExecContext(ctx, "UPDATE accounts SET balance = balance + $1 WHERE id = $2", amount, toID)
if err != nil {
    return err
}

return tx.Commit()
```

## Nullable Values

### 8. Handle NULL columns

```go
var (
    id    int
    name  string
    email sql.NullString // use sql.Null* for nullable columns
)

err := db.QueryRowContext(ctx, "SELECT id, name, email FROM users WHERE id = $1", id).
    Scan(&id, &name, &email)

if email.Valid {
    fmt.Println(email.String)
}
```

## Prepared Statements

### 9. Reusable prepared statement

```go
stmt, err := db.PrepareContext(ctx, "SELECT id, name FROM users WHERE id = $1")
if err != nil {
    return err
}
defer stmt.Close()

var u User
err = stmt.QueryRowContext(ctx, 42).Scan(&u.ID, &u.Name)
```
