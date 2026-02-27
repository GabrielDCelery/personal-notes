package main

import "fmt"

func main() {
	ch := make(chan string)
	go func(ch chan string) {
		defer close(ch)
		ch <- "hello from goroutine"
	}(ch)
	fmt.Println(<-ch)

}
