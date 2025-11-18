---
title: Read file content as string in golang
author: GaborZeller
date: 2024-09-17T19-01-51Z
tags:
draft: true
---

# Read file content as string in golang

```go
func readFileAsString(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
```
