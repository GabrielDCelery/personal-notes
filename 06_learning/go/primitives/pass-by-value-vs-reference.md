# Go: Pass by Value vs Reference

## Primitives — Value (always copied)

```go
func modify(n int, f float64, b bool, s string) {
    n = 99; f = 9.9; b = true; s = "changed"
}

n, f, b, s := 1, 1.1, false, "hello"
modify(n, f, b, s)
// n=1, f=1.1, b=false, s="hello" — all unchanged
```

Strings are immutable — even though a string header `{ptr, len}` is copied, you can never mutate the underlying bytes. Reassignment never escapes the function.

## Pointers — Reference

```go
func modify(n *int) { *n = 99 }

n := 1
modify(&n)
fmt.Println(n) // 99
```

## Arrays — Value (full copy)

```go
func modify(a [3]int) { a[0] = 99 }

a := [3]int{1, 2, 3}
modify(a)
fmt.Println(a[0]) // 1 — unchanged
```

## Slices — Value header, shared backing array

```go
func modifyElem(s []int)  { s[0] = 99 }
func grow(s []int)        { s = append(s, 99) }

s := []int{1, 2, 3}
modifyElem(s)
fmt.Println(s[0]) // 99 — visible

grow(s)
fmt.Println(len(s)) // 3 — NOT visible, need *[]int or return
```

## Maps — Reference type

```go
func modify(m map[string]int) { m["a"] = 99 }

m := map[string]int{"a": 1}
modify(m)
fmt.Println(m["a"]) // 99 — visible
```

## Structs — Value (full copy)

```go
type Point struct{ X, Y int }

func move(p Point)   { p.X = 99 }
func moveP(p *Point) { p.X = 99 }

p := Point{1, 2}
move(p)
fmt.Println(p.X) // 1 — unchanged

moveP(&p)
fmt.Println(p.X) // 99 — changed
```

## Channels — Reference type

```go
func send(ch chan int) { ch <- 99 }

ch := make(chan int, 1)
send(ch)
fmt.Println(<-ch) // 99 — visible, same underlying channel
```

Like maps, a channel variable is a pointer to an internal runtime structure. Closing or sending through a copy affects the same channel.

## Functions — Reference type

```go
func apply(fn func(int) int, n int) int { return fn(n) }

double := func(n int) int { return n * 2 }
fmt.Println(apply(double, 5)) // 10
```

Function values are pointers to code + closure data. Passing them around is cheap.

## Interfaces — Value containing `(type, pointer)` pair

```go
type Animal interface{ Sound() string }
type Dog struct{ Name string }
func (d *Dog) Sound() string { return "woof" }

func speak(a Animal) { fmt.Println(a.Sound()) }

d := &Dog{Name: "Rex"}
speak(d) // "woof" — interface wraps the pointer
```

An interface value is a pair `{type descriptor, data pointer}`. The underlying concrete value is usually accessed via a pointer, so mutations to it are visible.

## Quick Reference

| Type                     | Category            | Element mutations visible | Reassignment visible | Notes                          |
| ------------------------ | ------------------- | ------------------------- | -------------------- | ------------------------------ |
| `int`, `float64`, `bool` | Value               | N/A                       | No                   | Always copied                  |
| `string`                 | Value               | N/A (immutable)           | No                   | Header copied, bytes immutable |
| `*T`                     | Reference           | Yes                       | Yes                  | Dereference to mutate          |
| `[N]T`                   | Value               | No                        | No                   | Full array copied              |
| `[]T`                    | Value header        | **Yes**                   | No                   | Use `*[]T` to grow             |
| `map[K]V`                | Reference           | **Yes**                   | N/A                  | Never need `*map`              |
| `struct`                 | Value               | No                        | No                   | Use `*T` for mutation          |
| `chan T`                 | Reference           | **Yes**                   | N/A                  | Same channel across copies     |
| `func`                   | Reference           | **Yes** (via closure)     | N/A                  | Closures capture by reference  |
| `interface`              | Value `(type, ptr)` | **Yes** (usually)         | No                   | Depends on concrete type       |

**Rule of thumb**: if `make()` or `&` created it, copies share the same underlying data.
