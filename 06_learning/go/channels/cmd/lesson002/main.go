package main

import "fmt"

func main() {
	ch := make(chan int, 3)

	// NOTE: Sending values to the channel that is not in a goroutine works because it is a buffered channel
	for i := range 3 {
		ch <- (i + 1)
	}

	close(ch)

	for value := range ch {
		fmt.Println(value)
	}
}
