package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	for id := range 10 {
		wg.Add(1)
		<-ticker.C
		go func() {
			defer wg.Done()
			makeRequest(id)
		}()
	}
	wg.Wait()
	fmt.Printf("all requests completed\n")
}

func makeRequest(id int) {
	tm := time.Now().Format("15:04:05.000")
	fmt.Printf("request %d sent at %v\n", id, tm)
}
