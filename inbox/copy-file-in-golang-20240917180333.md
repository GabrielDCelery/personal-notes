---
title: Copy file in golang
author: GaborZeller
date: 2024-09-17T19-03-33Z
tags:
draft: true
---

# Copy file in golang

```go
func copyFile(source string, destination string) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	destinationFile, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer destinationFile.Close()
	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}
	return nil
}
```
