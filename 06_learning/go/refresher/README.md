# Go Interview Prep

A focused lesson series for senior Go developers preparing for technical interviews. Each lesson covers interview-critical concepts, common gotchas, and practical exercises.

## Target Audience

Senior developers (2+ years Go experience) who need:

- Quick refreshers on concepts they may have forgotten
- Deep dives into topics that regularly appear in interviews
- Understanding of edge cases and runtime behaviour
- Practical challenges to test understanding

## Lesson Index

### Language & Type System

| Lesson | Topic | Key Concepts |
| ------ | ----- | ------------ |
| [01](lesson-01-interfaces-and-type-system.md) | Interfaces & the Type System | Duck typing, embedding, nil interface pitfalls, type assertions |
| [02](lesson-02-error-handling-patterns.md) | Error Handling Patterns | `errors.Is/As`, sentinel errors, `%w` wrapping, `panic/recover` |

### Concurrency (Beyond Channels)

> Channels are covered in depth in [`go/channels/`](../channels/) — this group focuses on what's not covered there.

| Lesson | Topic | Key Concepts |
| ------ | ----- | ------------ |
| [03](lesson-03-goroutines-and-scheduler.md) | Goroutines & the Go Scheduler | M:N threading, GOMAXPROCS, goroutine leaks, work stealing |
| [04](lesson-04-sync-primitives-and-patterns.md) | Sync Primitives & Patterns | Mutex, RWMutex, WaitGroup, Once, atomic, sync.Map |

### Memory & Runtime

| Lesson | Topic | Key Concepts |
| ------ | ----- | ------------ |
| [05](lesson-05-memory-model-and-escape-analysis.md) | Memory Model & Escape Analysis | Happens-before, heap vs stack, escape analysis, GC tuning |
| [06](lesson-06-context-and-cancellation.md) | Context & Cancellation | `context.Context` contract, propagation, `WithValue` anti-patterns |

### Modern Go

| Lesson | Topic | Key Concepts |
| ------ | ----- | ------------ |
| [07](lesson-07-generics.md) | Generics | Type parameters, constraints, `~T`, interfaces vs generics |
| [08](lesson-08-testing-and-benchmarking.md) | Testing & Benchmarking | Table-driven tests, `t.Run`, race detector, `pprof` |

### Toolchain

| Lesson | Topic | Key Concepts |
| ------ | ----- | ------------ |
| [09](lesson-09-module-system-and-toolchain.md) | Module System & Toolchain | `go.mod`, workspaces, build tags, `go vet`, `staticcheck` |

### Standard Library Deep Dives

| Lesson | Topic | Key Concepts |
| ------ | ----- | ------------ |
| [10](lesson-10-time-operations.md) | Time Operations | `time.Time`, monotonic vs wall clock, formatting, timezones, timers, testable time |

## How to Use This Series

Work through lessons in order — later lessons assume knowledge from earlier ones (especially concurrency concepts). Each lesson takes 30–60 minutes and includes:

- Concept explanations with real-world context
- Code examples annotated with ✓ / ❌
- 2 hands-on exercises with solutions
- 4 interview questions with full answers

## Related Resources

- [Go Channels Deep Dive](../channels/) — 23-lesson series covering channels exhaustively
- [Go Primitives](../primitives/) — foundational Go data types
- [Go Language Specification](https://go.dev/ref/spec)
- [Effective Go](https://go.dev/doc/effective_go)
