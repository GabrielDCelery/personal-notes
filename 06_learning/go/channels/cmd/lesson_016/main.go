package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	jobs := make(chan int, 10)
	results := make(chan int, 10)

	for workerID := range 3 {
		wg.Add(1)
		go worker(workerID, jobs, results, &wg)
	}

	for job := range 10 {
		jobs <- job
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	close(jobs)

	var collected []int

	for result := range results {
		collected = append(collected, result)
	}
	fmt.Printf("results: %v\n", collected)
}

func worker(workerID int, jobs <-chan int, results chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		fmt.Printf("worker %d processing job %d\n", workerID, job)
		results <- job * job
	}
}
