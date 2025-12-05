---
title: Passing boolean value into function
author: GaborZeller
date: 2025-12-05
tags: go
draft: true
---

# The problem

Passing a boolean as a function parameter is generally considered a code smell and should be avoided in most cases, though there are exceptions.

## Why It's Usually Bad

1. Reduces readability at call site:

```go
// What does true mean here?
user := CreateUser("John", true)

// What about these?
ProcessData(data, false, true, false)
```

2. Violates Single Responsibility Principle:

```go
// This function does two different things
func SaveUser(user User, sendEmail bool) {
	db.Save(user)
	if sendEmail {
		email.Send(user)
	}
}

```

## Better Alternatives

1. Split into separate functions:

```go
// Bad
func CreateUser(name string, isAdmin bool) User { ... }

// Good
func CreateUser(name string) User { ... }
func CreateAdmin(name string) User { ... }
```

2. Use an enum/constant type:

```go
type UserRole int

const (
	RoleUser UserRole = iota
	RoleAdmin
)

func CreateUser(name string, role UserRole) User { ... }

// Clear at call site
user := CreateUser("John", RoleAdmin)
```

3. Use options pattern:

```go

type UserOption func(*User)

func WithAdmin() UserOption {
	return func(u *User) { u.IsAdmin = true }
}

func CreateUser(name string, opts ...UserOption) User {
	user := User{Name: name}
	for _, opt := range opts {
		opt(&user)
	}
	return user
}

// Clear and extensible
user := CreateUser("John", WithAdmin())
```

4. Use a struct for configuration:

```go

type UserConfig struct {
	Name string
	IsAdmin bool
	Active bool
}

func CreateUser(config UserConfig) User { ... }

// Named fields make it clear
user := CreateUser(UserConfig{
	Name: "John",
	IsAdmin: true,
	Active: false,
})

```

## When Booleans Are Acceptable

1. The boolean represents a clear, binary state:

```go
// Clear meaning
func SetEnabled(enabled bool)
func IsValid() bool
func Contains(item string, caseSensitive bool)
```

2. Well-named at call site:

```go
const (
	CaseSensitive = true
	CaseInsensitive = false
)

result := strings.Contains(text, "hello", CaseInsensitive)
```

3. Private/internal functions:

```go
// Internal helper - less critical
func (p *Parser) parse(data string, strict bool) error
```

## Rule of Thumb

- One boolean parameter: Often okay if well-named
- Multiple boolean parameters: Almost always bad
- Boolean that changes behavior significantly: Split into separate functions
- Public API: Avoid booleans, use explicit types
