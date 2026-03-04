# Testify

```sh
go get github.com/stretchr/testify
```

> Assertion library, mock framework, and test suite. Most widely used Go testing package.

## assert vs require

```go
// assert — marks test as failed, continues execution
assert.Equal(t, expected, got)

// require — marks test as failed, stops execution immediately
require.NoError(t, err) // use for setup steps that must succeed
```

## Common assertions

```go
import "github.com/stretchr/testify/assert"

assert.Equal(t, expected, actual)
assert.NotEqual(t, unexpected, actual)

assert.Nil(t, val)
assert.NotNil(t, val)

assert.True(t, condition)
assert.False(t, condition)

assert.NoError(t, err)
assert.Error(t, err)
assert.ErrorIs(t, err, ErrNotFound)
assert.ErrorAs(t, err, &target)

assert.Contains(t, "hello world", "world")    // string contains
assert.Contains(t, []int{1, 2, 3}, 2)         // slice contains
assert.Contains(t, map[string]int{"a": 1}, "a") // map has key

assert.Len(t, slice, 3)
assert.Empty(t, slice)
assert.NotEmpty(t, slice)

assert.ElementsMatch(t, []int{1, 2, 3}, []int{3, 1, 2}) // order independent
```

## Custom message

```go
assert.Equal(t, 5, result, "Add(2, 3) should equal 5")
assert.Equal(t, 5, result, "expected %d got %d", 5, result)
```

## Require (stop on failure)

```go
import "github.com/stretchr/testify/require"

user, err := store.GetUser(1)
require.NoError(t, err)        // stop if error
require.NotNil(t, user)        // stop if nil
assert.Equal(t, "Alice", user.Name) // safe to continue
```

## Mocks

```go
import "github.com/stretchr/testify/mock"

type MockStore struct {
    mock.Mock
}

func (m *MockStore) GetUser(id int) (*User, error) {
    args := m.Called(id)
    return args.Get(0).(*User), args.Error(1)
}

// In test
func TestHandler(t *testing.T) {
    store := new(MockStore)

    // Set expectation
    store.On("GetUser", 1).Return(&User{Name: "Alice"}, nil)

    h := NewHandler(store)
    // ... call handler ...

    // Verify all expectations were met
    store.AssertExpectations(t)
}
```

## Mock with any argument

```go
store.On("GetUser", mock.Anything).Return(&User{Name: "Alice"}, nil)
```

## Mock called multiple times

```go
store.On("GetUser", 1).Return(&User{Name: "Alice"}, nil).Once()
store.On("GetUser", 1).Return(nil, ErrNotFound).Once()
// first call → Alice, second call → error
```

## Suite (setup/teardown)

```go
import "github.com/stretchr/testify/suite"

type UserSuite struct {
    suite.Suite
    db *sql.DB
}

func (s *UserSuite) SetupSuite() {
    s.db = connectTestDB()
}

func (s *UserSuite) TearDownSuite() {
    s.db.Close()
}

func (s *UserSuite) SetupTest() {
    // runs before each test
    clearDB(s.db)
}

func (s *UserSuite) TestCreateUser() {
    // use s.Assert() or s.Require()
    s.Require().NoError(createUser(s.db, "Alice"))
}

// Required to run the suite
func TestUserSuite(t *testing.T) {
    suite.Run(t, new(UserSuite))
}
```
