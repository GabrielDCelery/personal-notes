# Prometheus

```sh
go get github.com/prometheus/client_golang/prometheus
go get github.com/prometheus/client_golang/prometheus/promhttp
```

## Metric types

| Type | Use case |
|---|---|
| Counter | monotonically increasing value (requests, errors) |
| Gauge | value that goes up and down (active connections, queue size) |
| Histogram | distribution of values (request duration, payload size) |
| Summary | similar to histogram, pre-calculated quantiles |

## Counter

```go
var requestsTotal = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "http_requests_total",
        Help: "Total number of HTTP requests",
    },
    []string{"method", "path", "status"},
)

func init() {
    prometheus.MustRegister(requestsTotal)
}

// Increment
requestsTotal.WithLabelValues("GET", "/items", "200").Inc()
requestsTotal.WithLabelValues("POST", "/items", "400").Add(5)
```

## Gauge

```go
var activeConnections = prometheus.NewGauge(prometheus.GaugeOpts{
    Name: "active_connections",
    Help: "Number of active connections",
})

func init() {
    prometheus.MustRegister(activeConnections)
}

activeConnections.Inc()
activeConnections.Dec()
activeConnections.Set(42)
```

## Histogram (most common for latency)

```go
var requestDuration = prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name:    "http_request_duration_seconds",
        Help:    "HTTP request duration in seconds",
        Buckets: prometheus.DefBuckets, // .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10
    },
    []string{"method", "path"},
)

func init() {
    prometheus.MustRegister(requestDuration)
}

// Observe a value
start := time.Now()
// ... handle request ...
requestDuration.WithLabelValues("GET", "/items").Observe(time.Since(start).Seconds())
```

## Expose /metrics endpoint

```go
import "github.com/prometheus/client_golang/prometheus/promhttp"

mux := http.NewServeMux()
mux.Handle("/metrics", promhttp.Handler())
mux.HandleFunc("/items", listItems)

http.ListenAndServe(":8080", mux)
```

## Middleware to track HTTP metrics

```go
func metricsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
        next.ServeHTTP(rw, r)

        duration := time.Since(start).Seconds()
        status := strconv.Itoa(rw.status)

        requestsTotal.WithLabelValues(r.Method, r.URL.Path, status).Inc()
        requestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
    })
}

type responseWriter struct {
    http.ResponseWriter
    status int
}

func (rw *responseWriter) WriteHeader(status int) {
    rw.status = status
    rw.ResponseWriter.WriteHeader(status)
}
```

## Custom registry (avoid global state in tests)

```go
reg := prometheus.NewRegistry()

counter := prometheus.NewCounter(prometheus.CounterOpts{
    Name: "my_counter",
})
reg.MustRegister(counter)

// Use custom registry with handler
promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
```
