# Kafka (confluent-go)

```sh
go get github.com/confluentinc/confluent-kafka-go/v2/kafka
```

## Producer

```go
p, err := kafka.NewProducer(&kafka.ConfigMap{
    "bootstrap.servers": "localhost:9092",
})
if err != nil {
    log.Fatal(err)
}
defer p.Close()

// Produce message
topic := "my-topic"
p.Produce(&kafka.Message{
    TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
    Key:            []byte("key"),
    Value:          []byte("hello"),
}, nil)

// Wait for all messages to be delivered
p.Flush(5000)
```

## Producer with delivery report

```go
go func() {
    for e := range p.Events() {
        switch ev := e.(type) {
        case *kafka.Message:
            if ev.TopicPartition.Error != nil {
                log.Printf("delivery failed: %v", ev.TopicPartition.Error)
            }
        }
    }
}()
```

## Consumer

```go
c, err := kafka.NewConsumer(&kafka.ConfigMap{
    "bootstrap.servers": "localhost:9092",
    "group.id":          "my-group",
    "auto.offset.reset": "earliest",
})
if err != nil {
    log.Fatal(err)
}
defer c.Close()

c.SubscribeTopics([]string{"my-topic"}, nil)

for {
    msg, err := c.ReadMessage(10 * time.Second)
    if err != nil {
        // timeout or error
        continue
    }
    fmt.Printf("key=%s value=%s\n", msg.Key, msg.Value)
}
```

## Consumer with graceful shutdown

```go
sigchan := make(chan os.Signal, 1)
signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

run := true
for run {
    select {
    case sig := <-sigchan:
        fmt.Printf("shutting down: %v\n", sig)
        run = false
    default:
        msg, err := c.ReadMessage(100 * time.Millisecond)
        if err != nil {
            continue
        }
        process(msg)
    }
}
```

## Commit offsets manually

```go
c, _ := kafka.NewConsumer(&kafka.ConfigMap{
    "bootstrap.servers":  "localhost:9092",
    "group.id":           "my-group",
    "enable.auto.commit": false,
})

msg, _ := c.ReadMessage(-1)
process(msg)
c.CommitMessage(msg) // only commit after successful processing
```
