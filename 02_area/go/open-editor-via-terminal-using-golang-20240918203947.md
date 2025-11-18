---
title: Open editor via terminal using golang
author: GaborZeller
date: 2024-09-18T20-39-47Z
tags:
draft: true
---

# Open editor via terminal using golang

```go
	cmd := exec.Command("editor you want to use", "path to file")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Fatalln(err)
	}

```
