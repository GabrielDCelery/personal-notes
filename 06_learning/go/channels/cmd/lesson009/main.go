package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup

	results := make(chan int)

	gen := generator()

	for range 3 {
		wg.Add(1)
		sq := square(gen)
		go func() {
			defer wg.Done()
			for s := range sq {
				results <- s
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		fmt.Println(result)
	}
}

func generator() <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for i := range 9 {
			out <- (i + 1)
		}
	}()
	return out
}

func square(generator <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for num := range generator {
			out <- num * num
		}
	}()
	return out
}
