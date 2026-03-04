# Gin

```sh
go get github.com/gin-gonic/gin
```

## Setup

```go
import "github.com/gin-gonic/gin"

r := gin.Default() // includes Logger and Recoverer middleware

r.GET("/health", func(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"status": "ok"})
})

r.Run(":8080")
```

## Routes

```go
r.GET("/items", listItems)
r.POST("/items", createItem)
r.PUT("/items/:id", updateItem)
r.DELETE("/items/:id", deleteItem)

// Route param
func getItem(c *gin.Context) {
    id := c.Param("id")
    // ...
}

// Query param
func listItems(c *gin.Context) {
    page := c.DefaultQuery("page", "1")
    limit := c.Query("limit") // "" if not set
}
```

## JSON response

```go
func getItem(c *gin.Context) {
    item := Item{ID: 1, Name: "Alice"}
    c.JSON(http.StatusOK, item)
}

// Quick map response
c.JSON(http.StatusOK, gin.H{"message": "created"})

// Error response
c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
```

## Bind request body

```go
type CreateRequest struct {
    Name  string `json:"name" binding:"required"`
    Email string `json:"email" binding:"required,email"`
}

func createItem(c *gin.Context) {
    var req CreateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    // use req.Name, req.Email
}
```

## Middleware

```go
// Global
r.Use(gin.Logger())
r.Use(gin.Recovery())

// Custom middleware
func authMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
            return
        }
        c.Next()
    }
}
```

## Route groups

```go
api := r.Group("/api/v1")
api.Use(authMiddleware())
{
    api.GET("/users", listUsers)
    api.POST("/users", createUser)
    api.GET("/users/:id", getUser)
}

// Public
r.POST("/login", login)
```

## Store and retrieve context values

```go
// Set in middleware
c.Set("user", user)

// Get in handler
user, exists := c.Get("user")
if !exists {
    c.AbortWithStatus(http.StatusUnauthorized)
    return
}
u := user.(User)
```

## gin vs chi

|                      | chi                       | gin                |
| -------------------- | ------------------------- | ------------------ |
| API style            | `http.Handler` compatible | own `*gin.Context` |
| Middleware ecosystem | any `net/http` middleware | gin-specific       |
| Binding/validation   | manual                    | built-in           |
| Performance          | fast                      | faster             |
| Learning curve       | low                       | low                |
