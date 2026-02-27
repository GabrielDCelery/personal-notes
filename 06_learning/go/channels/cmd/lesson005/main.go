package main

import (
	"fmt"
	"time"
)

func main() {
	ch := make(chan string)

	go func(ch chan<- string) {
		time.Sleep(500 * time.Millisecond)
		ch <- "operation completed"
	}(ch)

	select {
	case msg := <-ch:
		fmt.Println(msg)
	case <-time.After(600 * time.Millisecond):
		fmt.Println("operation took too long")
	}
}
