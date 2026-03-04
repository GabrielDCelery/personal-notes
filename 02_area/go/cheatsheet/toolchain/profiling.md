# Go Profiling

## Quick Reference

| Use case        | Tool / Method         |
| --------------- | --------------------- |
| CPU profile     | `go test -cpuprofile` |
| Memory profile  | `go test -memprofile` |
| Benchmarks      | `go test -bench`      |
| Live profiling  | `net/http/pprof`      |
| Analyze profile | `go tool pprof`       |
| Execution trace | `go test -trace`      |
| View trace      | `go tool trace`       |

## Benchmarks

### 1. Write a benchmark

```go
func BenchmarkFoo(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Foo()
    }
}
```

### 2. Run benchmarks

```sh
go test -bench .                       # run all benchmarks
go test -bench BenchmarkFoo            # run specific
go test -bench . -benchmem             # include memory allocations
go test -bench . -count 5              # run 5 times for stable results
go test -bench . -benchtime 5s         # run for 5 seconds
```

### 3. Compare benchmarks

```sh
go install golang.org/x/perf/cmd/benchstat@latest

go test -bench . -count 10 > old.txt
# make changes
go test -bench . -count 10 > new.txt
benchstat old.txt new.txt
```

## CPU Profiling

### 4. Profile from tests

```sh
go test -cpuprofile cpu.prof -bench .
go tool pprof cpu.prof
```

### 5. Profile from running app

```go
import _ "net/http/pprof"

// Add to main (if not already serving HTTP)
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()
```

```sh
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=10
```

### 6. pprof interactive commands

```
(pprof) top 10              # top 10 CPU consumers
(pprof) list FunctionName   # annotated source
(pprof) web                 # open graph in browser
(pprof) svg > profile.svg   # save graph
```

## Memory Profiling

### 7. From tests

```sh
go test -memprofile mem.prof -bench .
go tool pprof mem.prof
```

### 8. From running app

```sh
# Heap — current allocations
go tool pprof http://localhost:6060/debug/pprof/heap

# Allocs — all allocations (including freed)
go tool pprof http://localhost:6060/debug/pprof/allocs
```

### 9. Common pprof flags

```sh
go tool pprof -alloc_space mem.prof    # total bytes allocated
go tool pprof -inuse_space mem.prof    # currently held bytes
go tool pprof -alloc_objects mem.prof  # total objects allocated
```

## Execution Trace

### 10. Capture and view trace

```sh
go test -trace trace.out -bench .
go tool trace trace.out
```

Opens a browser with:

- Goroutine analysis
- Network/sync blocking
- Scheduler latency
- GC pauses

### 11. Programmatic trace

```go
f, _ := os.Create("trace.out")
trace.Start(f)
defer trace.Stop()
```

## Other pprof Endpoints

### 12. Available profiles

```sh
# Goroutines
go tool pprof http://localhost:6060/debug/pprof/goroutine

# Block (contention)
go tool pprof http://localhost:6060/debug/pprof/block

# Mutex contention
go tool pprof http://localhost:6060/debug/pprof/mutex
```

> Enable block/mutex profiling in code:
>
> ```go
> runtime.SetBlockProfileRate(1)
> runtime.SetMutexProfileFraction(1)
> ```
