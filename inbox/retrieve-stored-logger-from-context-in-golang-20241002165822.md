---
title: Retrieve stored logger from context in golang
author: GaborZeller
date: 2024-10-02T16-58-22Z
tags:
draft: true
---

# Retrieve stored logger from context in golang


```go
type contextKey string

const loggerKey = contextKey("logger")

func GetLoggerFromContext(ctx context.Context) *zap.Logger {
	if logger, ok := ctx.Value(loggerKey).(*zap.Logger); ok {
		cp := *logger
		return &cp
	}
	cp := *GetDefaultLogger()
	return &cp
}
```

- `contextKey` - private string type that prevents collision in context, it is a best practice over using regular strings
- `ctx.Value(loggerKey).(*zap.Logger)` - since `ctx.Value` returns an `interface{}` we need to use `type assertion`, we are telling go that we know what the value should be (if not the code will panic)
- the above example also creates a copy of the logger instead of returning them which is also a best practice to prevent unintended modifications
