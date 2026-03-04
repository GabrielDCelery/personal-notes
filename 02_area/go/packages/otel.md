# OpenTelemetry (otel)

```sh
go get go.opentelemetry.io/otel
go get go.opentelemetry.io/otel/sdk/trace
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc
```

## Concepts

| Term     | Meaning                                           |
| -------- | ------------------------------------------------- |
| Trace    | full journey of a request across services         |
| Span     | single unit of work within a trace                |
| Context  | carries the active span across function calls     |
| Exporter | sends traces to a backend (Jaeger, Datadog, etc.) |
| Tracer   | creates spans                                     |

## Setup (OTLP exporter)

```go
func initTracer(ctx context.Context) (*sdktrace.TracerProvider, error) {
    exporter, err := otlptracegrpc.New(ctx,
        otlptracegrpc.WithEndpoint("localhost:4317"),
        otlptracegrpc.WithInsecure(),
    )
    if err != nil {
        return nil, err
    }

    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
        sdktrace.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceName("my-service"),
        )),
    )

    otel.SetTracerProvider(tp)
    otel.SetTextMapPropagator(propagation.TraceContext{})

    return tp, nil
}

// In main
tp, err := initTracer(ctx)
if err != nil {
    log.Fatal(err)
}
defer tp.Shutdown(ctx)
```

## Create spans

```go
tracer := otel.Tracer("my-service")

func doWork(ctx context.Context) error {
    ctx, span := tracer.Start(ctx, "doWork")
    defer span.End()

    // always pass ctx downstream
    return callDB(ctx)
}
```

## Add attributes to span

```go
ctx, span := tracer.Start(ctx, "getUser")
defer span.End()

span.SetAttributes(
    attribute.Int("user.id", userID),
    attribute.String("user.name", name),
)
```

## Record errors

```go
ctx, span := tracer.Start(ctx, "fetchData")
defer span.End()

data, err := fetch(ctx)
if err != nil {
    span.RecordError(err)
    span.SetStatus(codes.Error, err.Error())
    return err
}
```

## HTTP middleware (propagate context)

```go
import "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

// Wrap handler — extracts incoming trace context and creates spans
mux := http.NewServeMux()
mux.HandleFunc("/items", listItems)

handler := otelhttp.NewHandler(mux, "my-service")
http.ListenAndServe(":8080", handler)
```

## Propagate context in outgoing HTTP calls

```go
import "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

client := &http.Client{
    Transport: otelhttp.NewTransport(http.DefaultTransport),
}

req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
resp, err := client.Do(req) // trace context injected into headers
```

## Add span events (structured log within span)

```go
span.AddEvent("cache miss", trace.WithAttributes(
    attribute.String("key", cacheKey),
))
```
