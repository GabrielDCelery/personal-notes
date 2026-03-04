# Mockery

```sh
go install github.com/vektra/mockery/v2@latest
```

> Automatically generates testify mocks from interfaces. Eliminates writing mock structs by hand.

## Basic usage

```go
// Given this interface
type UserStore interface {
    GetUser(ctx context.Context, id int) (*User, error)
    CreateUser(ctx context.Context, user *User) error
    DeleteUser(ctx context.Context, id int) error
}
```

```sh
# Generate mock for a single interface
mockery --name=UserStore

# Generate all interfaces in a package
mockery --all

# Output goes to mocks/ by default
```

## .mockery.yaml (recommended config)

```yaml
with-expecter: true # generates type-safe EXPECT() API
mockname: "Mock{{.InterfaceName}}"
outpkg: mocks
dir: mocks
filename: "{{.InterfaceName | snakecase}}.go"
```

## Using generated mock (standard API)

```go
import "myapp/mocks"

func TestGetUser(t *testing.T) {
    store := mocks.NewMockUserStore(t) // t auto-asserts expectations on cleanup

    store.On("GetUser", mock.Anything, 1).
        Return(&User{Name: "Alice"}, nil)

    svc := NewUserService(store)
    user, err := svc.GetUser(ctx, 1)

    require.NoError(t, err)
    assert.Equal(t, "Alice", user.Name)
}
```

## Using Expecter API (type-safe, with-expecter: true)

```go
store := mocks.NewMockUserStore(t)

// Type-safe — catches wrong arg types at compile time
store.EXPECT().
    GetUser(mock.Anything, 1).
    Return(&User{Name: "Alice"}, nil)
```

## Return error

```go
store.EXPECT().
    GetUser(mock.Anything, 99).
    Return(nil, ErrNotFound)
```

## Called multiple times

```go
store.EXPECT().
    GetUser(mock.Anything, mock.Anything).
    Return(&User{Name: "Alice"}, nil).
    Times(3)

// Or
store.EXPECT().
    GetUser(mock.Anything, mock.Anything).
    Return(&User{Name: "Alice"}, nil).
    Once()
```

## Run custom logic on call

```go
store.EXPECT().
    CreateUser(mock.Anything, mock.Anything).
    RunAndReturn(func(ctx context.Context, u *User) error {
        u.ID = 42 // simulate DB assigning ID
        return nil
    })
```

## Mockery vs manual mocks

|             | Manual mock                   | Mockery                        |
| ----------- | ----------------------------- | ------------------------------ |
| Setup       | write struct by hand          | `mockery --name=Foo`           |
| Type safety | runtime                       | compile-time with Expecter     |
| Maintenance | update when interface changes | regenerate                     |
| Best for    | simple interfaces             | large interfaces, many methods |
