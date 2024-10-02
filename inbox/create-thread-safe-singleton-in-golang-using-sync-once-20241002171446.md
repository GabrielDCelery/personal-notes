---
title: Create thread safe singleton in golang using sync once
author: GaborZeller
date: 2024-10-02T17-14-46Z
tags:
draft: true
---

# Create thread safe singleton in golang using sync once

```go
var (
	singleton     *zap.Logger
	singletonOnce sync.Once
)

func CreateSingletion() *zap.Logger {
	singletonOnce.Do(func() {
		singleton = SomeFuncCreatingTheInstance
	})
	return singleton
}

```

- `singletonOnce.Do` - makes the initialization thread safe and returns the pre-initialised instance for each subsequent call
- `lazy initialization` - the singleton is only created when needed
- `environment awareness` - the func passed into `Do` makes the initialization flexible
