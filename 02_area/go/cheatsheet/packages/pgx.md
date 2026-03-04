# pgx (PostgreSQL)

```sh
go get github.com/jackc/pgx/v5
```

## Setup

```go
import "github.com/jackc/pgx/v5/pgxpool"

pool, err := pgxpool.New(ctx, "postgres://user:pass@localhost:5432/dbname")
if err != nil {
    log.Fatal(err)
}
defer pool.Close()

if err := pool.Ping(ctx); err != nil {
    log.Fatal(err)
}
```

## Query multiple rows

```go
rows, err := pool.Query(ctx, "SELECT id, name, email FROM users WHERE active = $1", true)
if err != nil {
    return err
}
defer rows.Close()

// pgx.CollectRows — cleaner than manual scanning
users, err := pgx.CollectRows(rows, pgx.RowToStructByName[User])
```

## Query single row

```go
rows, err := pool.Query(ctx, "SELECT id, name FROM users WHERE id = $1", id)
if err != nil {
    return err
}

user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[User])
if errors.Is(err, pgx.ErrNoRows) {
    return nil, ErrNotFound
}
```

## Exec (insert / update / delete)

```go
tag, err := pool.Exec(ctx,
    "INSERT INTO users (name, email) VALUES ($1, $2)",
    "Alice", "alice@example.com",
)
if err != nil {
    return err
}

fmt.Println(tag.RowsAffected())
```

## Insert returning ID

```go
var id int
err := pool.QueryRow(ctx,
    "INSERT INTO users (name) VALUES ($1) RETURNING id",
    "Alice",
).Scan(&id)
```

## Transaction

```go
tx, err := pool.Begin(ctx)
if err != nil {
    return err
}
defer tx.Rollback(ctx) // no-op if committed

_, err = tx.Exec(ctx, "UPDATE accounts SET balance = balance - $1 WHERE id = $2", amount, fromID)
if err != nil {
    return err
}

_, err = tx.Exec(ctx, "UPDATE accounts SET balance = balance + $1 WHERE id = $2", amount, toID)
if err != nil {
    return err
}

return tx.Commit(ctx)
```

## Named struct scanning

```go
// Struct fields must match column names (case insensitive)
type User struct {
    ID    int    `db:"id"`
    Name  string `db:"name"`
    Email string `db:"email"`
}

rows, _ := pool.Query(ctx, "SELECT id, name, email FROM users")
users, err := pgx.CollectRows(rows, pgx.RowToStructByName[User])
```

## Batch queries

```go
batch := &pgx.Batch{}
batch.Queue("INSERT INTO items (name) VALUES ($1)", "a")
batch.Queue("INSERT INTO items (name) VALUES ($1)", "b")
batch.Queue("INSERT INTO items (name) VALUES ($1)", "c")

results := pool.SendBatch(ctx, batch)
defer results.Close()

for i := 0; i < batch.Len(); i++ {
    _, err := results.Exec()
    if err != nil {
        return err
    }
}
```
