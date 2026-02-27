package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	sem := make(chan struct{}, 3)

	for id := range 10 {
		wg.Add(1)
		sem <- struct{}{}
		go func(id int) {
			defer wg.Done()
			defer func() { <-sem }()
			process(id)
		}(id)
	}

	wg.Wait()

	fmt.Printf("all jobs completed\n")
}

func process(id int) {
	fmt.Printf("starting job %d\n", id)
	time.Sleep(1 * time.Second)
	fmt.Printf("finished job %d\n", id)
}
