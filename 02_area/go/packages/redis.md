# go-redis

```sh
go get github.com/redis/go-redis/v9
```

## Setup

```go
import "github.com/redis/go-redis/v9"

rdb := redis.NewClient(&redis.Options{
    Addr:     "localhost:6379",
    Password: "",
    DB:       0,
})

// verify connection
if err := rdb.Ping(ctx).Err(); err != nil {
    log.Fatal(err)
}
```

## Get / Set

```go
// Set with expiry
err := rdb.Set(ctx, "key", "value", 10*time.Minute).Err()

// Get
val, err := rdb.Get(ctx, "key").Result()
if errors.Is(err, redis.Nil) {
    // key does not exist
}

// Set only if not exists
set, err := rdb.SetNX(ctx, "lock", "1", 30*time.Second).Result()
```

## Delete / Expire

```go
rdb.Del(ctx, "key1", "key2")

rdb.Expire(ctx, "key", 5*time.Minute)

ttl, err := rdb.TTL(ctx, "key").Result()
```

## JSON (store struct)

```go
data, _ := json.Marshal(user)
rdb.Set(ctx, "user:1", data, time.Hour)

val, _ := rdb.Get(ctx, "user:1").Bytes()
json.Unmarshal(val, &user)
```

## Hash

```go
// Set fields
rdb.HSet(ctx, "user:1", "name", "Alice", "email", "alice@example.com")

// Get field
name, err := rdb.HGet(ctx, "user:1", "name").Result()

// Get all
fields, err := rdb.HGetAll(ctx, "user:1").Result()
// fields is map[string]string
```

## List

```go
rdb.LPush(ctx, "queue", "item1", "item2") // push left
rdb.RPush(ctx, "queue", "item3")          // push right

val, err := rdb.LPop(ctx, "queue").Result()  // pop left
val, err := rdb.RPop(ctx, "queue").Result()  // pop right

// blocking pop (wait up to 5s)
val, err := rdb.BLPop(ctx, 5*time.Second, "queue").Result()
```

## Pub / Sub

```go
// Publish
rdb.Publish(ctx, "channel", "message")

// Subscribe
sub := rdb.Subscribe(ctx, "channel")
defer sub.Close()

for msg := range sub.Channel() {
    fmt.Println(msg.Payload)
}
```

## Pipeline (batch commands)

```go
pipe := rdb.Pipeline()

pipe.Set(ctx, "key1", "val1", 0)
pipe.Set(ctx, "key2", "val2", 0)
pipe.Incr(ctx, "counter")

_, err := pipe.Exec(ctx)
```
