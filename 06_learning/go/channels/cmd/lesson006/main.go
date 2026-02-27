package main

import (
	"fmt"
	"time"
)

func main() {
	done := make(chan struct{})
	results := make(chan int)

	go func(results chan<- int, done <-chan struct{}) {
		count := 0

		for {
			time.Sleep(100 * time.Millisecond)
			select {
			case <-done:
				return
			case results <- count:
				count += 1
			}
		}

	}(results, done)

	for range 5 {
		fmt.Println(<-results)
	}

	close(done)

	time.Sleep(10 * time.Millisecond)

	fmt.Println("worker stopped")
}
