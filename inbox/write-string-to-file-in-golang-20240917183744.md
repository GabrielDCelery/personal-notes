---
title: Write string to file in golang
author: GaborZeller
date: 2024-09-17T19-37-44Z
tags:
draft: true
---

# Write string to file in golang

```go
func writeStringToFile(filePath string, content string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(content)
	if err != nil {
		return err
	}
	return nil
}
```
