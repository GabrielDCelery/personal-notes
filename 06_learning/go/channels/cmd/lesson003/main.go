package main

import "fmt"

func main() {
	//NOTE: The bidirectional chan int in main converts automatically when passed to the child functions
	ch := make(chan int)

	go producer(ch)

	consumer(ch)
}

// NOTE: producer can only send and close (appropriate for a producer)
func producer(ch chan<- int) {
	for i := range 5 {
		ch <- (i + 1)
	}
	close(ch)
}

// NOTE: consumer can only receive (can't accidentally close or send)
func consumer(ch <-chan int) {
	for value := range ch {
		fmt.Println(value)
	}
}
