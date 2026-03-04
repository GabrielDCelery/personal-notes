# Go Interfaces

## Why

- **Implicit implementation** — No `implements` keyword. If your type has the right methods, it satisfies the interface. This decouples packages — the consumer defines what it needs, not the producer.
- **Define interfaces where they're used, not where they're implemented** — Keep interfaces small and in the consumer package. This follows the dependency inversion principle.
- **Small interfaces** — One or two methods is ideal. io.Reader, io.Writer, and error are each one method. The smaller the interface, the more types satisfy it.
- **Interface for testing** — The main reason to use interfaces in Go. Define an interface for external dependencies (DB, HTTP client, email sender), then swap in a mock during tests.
- **Compile-time check** — `var _ MyInterface = (*MyStruct)(nil)` catches missing methods at compile time instead of runtime. Costs nothing, catches bugs early.

## Quick Reference

| Use case             | Notes                                  |
| -------------------- | -------------------------------------- |
| Define interface     | `type Name interface { Method() }`     |
| Implement implicitly | no `implements` keyword needed         |
| Check implementation | `var _ MyInterface = (*MyStruct)(nil)` |
| Type assertion       | `v, ok := x.(ConcreteType)`            |
| Type switch          | `switch v := x.(type) { case ... }`    |
| Empty interface      | `any` (alias for `interface{}`)        |
| Compose interfaces   | embed one interface in another         |

## Defining & Implementing

### 1. Define and implement an interface

```go
type Storer interface {
    Save(data []byte) error
    Load(id string) ([]byte, error)
}

type FileStore struct {
    dir string
}

func (f *FileStore) Save(data []byte) error {
    return os.WriteFile(f.dir+"/data", data, 0644)
}

func (f *FileStore) Load(id string) ([]byte, error) {
    return os.ReadFile(f.dir + "/" + id)
}
```

### 2. Use the interface (depend on abstraction)

```go
func backup(s Storer, data []byte) error {
    return s.Save(data)
}

// Works with any Storer implementation
backup(&FileStore{dir: "/tmp"}, []byte("hello"))
```

### 3. Compile-time implementation check

```go
// Will fail to compile if *FileStore doesn't implement Storer
var _ Storer = (*FileStore)(nil)
```

## Type Assertions & Switches

### 4. Type assertion

```go
var s Storer = &FileStore{}

fs, ok := s.(*FileStore)
if ok {
    fmt.Println(fs.dir)
}
```

### 5. Type switch

```go
func describe(i any) {
    switch v := i.(type) {
    case int:
        fmt.Println("int:", v)
    case string:
        fmt.Println("string:", v)
    case bool:
        fmt.Println("bool:", v)
    default:
        fmt.Printf("unknown: %T\n", v)
    }
}
```

## Common Patterns

### 6. Interface composition

```go
type Reader interface {
    Read() ([]byte, error)
}

type Writer interface {
    Write([]byte) error
}

type ReadWriter interface {
    Reader
    Writer
}
```

### 7. Mocking with interfaces (for testing)

```go
type EmailSender interface {
    Send(to, subject, body string) error
}

// Production implementation
type SMTPSender struct{}
func (s *SMTPSender) Send(to, subject, body string) error { /* ... */ }

// Test mock
type MockSender struct {
    Sent []string
}
func (m *MockSender) Send(to, subject, body string) error {
    m.Sent = append(m.Sent, to)
    return nil
}

// Test
func TestWelcomeEmail(t *testing.T) {
    mock := &MockSender{}
    sendWelcome(mock, "user@example.com")
    if len(mock.Sent) != 1 {
        t.Fatal("expected one email")
    }
}
```

### 8. Standard library interfaces worth knowing

```go
// io.Reader
type Reader interface {
    Read(p []byte) (n int, err error)
}

// io.Writer
type Writer interface {
    Write(p []byte) (n int, err error)
}

// fmt.Stringer (controls fmt output)
type Stringer interface {
    String() string
}

// error
type error interface {
    Error() string
}

// io.Closer
type Closer interface {
    Close() error
}
```

### 9. Implement fmt.Stringer

```go
type User struct {
    Name string
    Age  int
}

func (u User) String() string {
    return fmt.Sprintf("%s (%d)", u.Name, u.Age)
}

u := User{Name: "Alice", Age: 30}
fmt.Println(u) // Alice (30)
```
