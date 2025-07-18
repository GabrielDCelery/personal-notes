---
title: Reading files with Javascript, Go and Rust
author: GaborZeller
date: 2025-06-22T15-46-33Z
tags:
draft: true
---

# Reading files with Javascript, Go and Rust

Notes on what exactly happens when reading a text file line by line.

```go
package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	file, err := os.Open("/path/to/file")

	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	buffer := make([]byte, 1)

	for {
		bytesRead, err := file.Read(buffer)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			log.Fatal(err)
		}
		fmt.Print(string(buffer[:bytesRead]))
	}
}
```

```go
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

func main() {
	file, err := os.Open("/path/to/file")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n') // Read until newline
		if err != nil {
			break // End of file or error
		}
		fmt.Print(line) // Process the line
	}
}
```

```go
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

func main() {
	file, err := os.Open("/path/to/file")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)
	}
}
```

```go
file, err := os.Open("/home/gaze/projects/practice/read-line-by-line/sample3.txt")
if err != nil {
	log.Fatal(err)
}
defer file.Close()
```

In go we use the `os` package `os.Open` method which under the hood uses the `open(2)` system call to create a file descriptor.
