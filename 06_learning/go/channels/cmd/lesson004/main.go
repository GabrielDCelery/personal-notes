package main

import (
	"fmt"
	"time"
)

func main() {
	ch1 := make(chan string)
	ch2 := make(chan string)

	go func(ch chan<- string) {
		time.Sleep(100 * time.Millisecond)
		ch <- "from channel 1"
	}(ch1)

	go func(ch chan<- string) {
		time.Sleep(200 * time.Millisecond)
		ch <- "from channel 2"
	}(ch2)

	for range 2 {
		select {
		case msg := <-ch1:
			fmt.Println(msg)
		case msg := <-ch2:
			fmt.Println(msg)
		}
	}
}
