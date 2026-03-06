# sqlx

```sh
go get github.com/jmoiron/sqlx
```

> Extends `database/sql` with struct scanning, named queries, and slice binding. Works with any `database/sql` driver.

## Setup

```go
import (
    "github.com/jmoiron/sqlx"
    _ "github.com/lib/pq"
)

db, err := sqlx.Connect("postgres", "postgres://user:pass@localhost/dbname?sslmode=disable")
if err != nil {
    log.Fatal(err)
}
defer db.Close()
```

## Struct scanning

```go
type User struct {
    ID    int    `db:"id"`
    Name  string `db:"name"`
    Email string `db:"email"`
}

// Multiple rows
var users []User
err := db.SelectContext(ctx, &users, "SELECT id, name, email FROM users WHERE active = $1", true)

// Single row
var u User
err := db.GetContext(ctx, &u, "SELECT id, name, email FROM users WHERE id = $1", 1)
if errors.Is(err, sql.ErrNoRows) {
    // not found
}
```

## Named queries (use struct fields as params)

```go
u := User{Name: "Alice", Email: "alice@example.com"}

_, err := db.NamedExecContext(ctx,
    "INSERT INTO users (name, email) VALUES (:name, :email)",
    u,
)
```

## Named query with map

```go
_, err := db.NamedExecContext(ctx,
    "UPDATE users SET name = :name WHERE id = :id",
    map[string]any{"name": "Alice", "id": 1},
)
```

## In queries (slice binding)

```go
ids := []int{1, 2, 3}

query, args, err := sqlx.In("SELECT * FROM users WHERE id IN (?)", ids)
if err != nil {
    return err
}

// Rebind for your driver (postgres uses $1, $2...)
query = db.Rebind(query)

var users []User
err = db.SelectContext(ctx, &users, query, args...)
```

## Transaction

```go
tx, err := db.BeginTxx(ctx, nil)
if err != nil {
    return err
}
defer tx.Rollback()

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

## sqlx vs database/sql vs pgx

|                      | database/sql | sqlx           | pgx           |
| -------------------- | ------------ | -------------- | ------------- |
| Struct scanning      | manual Scan  | `Get`/`Select` | `CollectRows` |
| Named params         | no           | yes            | no            |
| Performance          | baseline     | baseline       | faster        |
| PG-specific features | no           | no             | yes           |
