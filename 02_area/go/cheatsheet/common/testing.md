# Go Testing

## Quick Reference

| Use case          | Method                                 |
| ----------------- | -------------------------------------- |
| Basic test        | `func TestXxx(t *testing.T)`           |
| Table-driven test | slice of structs + `t.Run`             |
| Subtest           | `t.Run("name", func(t *testing.T) {})` |
| Fail test         | `t.Errorf` / `t.Fatalf`                |
| Skip test         | `t.Skip`                               |
| Test HTTP handler | `httptest.NewRecorder`                 |
| Test HTTP server  | `httptest.NewServer`                   |
| Benchmark         | `func BenchmarkXxx(b *testing.B)`      |

## Basics

### 1. Basic test

```go
func TestAdd(t *testing.T) {
    got := Add(2, 3)
    want := 5

    if got != want {
        t.Errorf("Add(2, 3) = %d, want %d", got, want)
    }
}
```

### 2. t.Errorf vs t.Fatalf

```go
t.Errorf("bad value: %d", got) // marks failed, continues test
t.Fatalf("bad value: %d", got) // marks failed, stops test immediately
```

## Table-Driven Tests

### 3. Table-driven test (standard Go pattern)

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name string
        a, b int
        want int
    }{
        {"positive", 2, 3, 5},
        {"zero", 0, 0, 0},
        {"negative", -1, -2, -3},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Add(tt.a, tt.b)
            if got != tt.want {
                t.Errorf("got %d, want %d", got, tt.want)
            }
        })
    }
}
```

### 4. Table-driven with error cases

```go
func TestDivide(t *testing.T) {
    tests := []struct {
        name    string
        a, b    float64
        want    float64
        wantErr bool
    }{
        {"normal", 10, 2, 5, false},
        {"divide by zero", 10, 0, 0, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Divide(tt.a, tt.b)
            if (err != nil) != tt.wantErr {
                t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
            }
            if !tt.wantErr && got != tt.want {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

## HTTP Testing

### 5. Test an HTTP handler directly

```go
func TestGetItem(t *testing.T) {
    req := httptest.NewRequest(http.MethodGet, "/items/1", nil)
    w := httptest.NewRecorder()

    getItem(w, req)

    resp := w.Result()
    if resp.StatusCode != http.StatusOK {
        t.Errorf("got %d, want %d", resp.StatusCode, http.StatusOK)
    }

    body, _ := io.ReadAll(resp.Body)
    // assert body content
}
```

### 6. Test against a real HTTP server

```go
func TestAPI(t *testing.T) {
    srv := httptest.NewServer(http.HandlerFunc(getItem))
    defer srv.Close()

    resp, err := http.Get(srv.URL + "/items/1")
    if err != nil {
        t.Fatal(err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        t.Errorf("got %d, want %d", resp.StatusCode, http.StatusOK)
    }
}
```

## Mocking

### 7. Mock with interface

```go
type Store interface {
    GetUser(id int) (*User, error)
}

type MockStore struct {
    user *User
    err  error
}

func (m *MockStore) GetUser(id int) (*User, error) {
    return m.user, m.err
}

func TestHandler(t *testing.T) {
    store := &MockStore{user: &User{Name: "Alice"}}
    h := NewHandler(store)

    req := httptest.NewRequest(http.MethodGet, "/user/1", nil)
    w := httptest.NewRecorder()
    h.ServeHTTP(w, req)

    if w.Code != http.StatusOK {
        t.Errorf("got %d, want %d", w.Code, http.StatusOK)
    }
}
```

## Other

### 8. Setup and teardown

```go
func TestMain(m *testing.M) {
    // setup
    setup()

    code := m.Run()

    // teardown
    teardown()

    os.Exit(code)
}
```

### 9. Benchmark

```go
func BenchmarkAdd(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Add(2, 3)
    }
}
// run with: go test -bench=.
```

### 10. Skip slow tests

```go
func TestSlowOperation(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping in short mode")
    }
    // ...
}
// run short mode: go test -short
```
