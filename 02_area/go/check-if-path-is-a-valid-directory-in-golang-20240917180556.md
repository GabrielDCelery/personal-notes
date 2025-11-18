---
title: Check if path is a valid directory in golang
author: GaborZeller
date: 2024-09-17T19-05-56Z
tags:
draft: true
---

# Check if path is a valid directory in golang

```go
func isValidDirectory(path string) bool {
	dirInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return dirInfo.IsDir()
}
```
