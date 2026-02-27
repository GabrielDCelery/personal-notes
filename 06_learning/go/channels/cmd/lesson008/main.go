package main

import (
	"fmt"
	"sync"
)

const numOfJobs = 5

func main() {
	var wg sync.WaitGroup
	jobs := make(chan int)
	results := make(chan int)

	for workerID := range 3 {
		wg.Add(1)
		go worker(workerID, jobs, results, &wg)
	}

	go func(wg *sync.WaitGroup, results chan<- int) {
		wg.Wait()
		close(results)
	}(&wg, results)

	go func(jobs chan<- int) {
		for num := range numOfJobs {
			jobs <- (num + 1)
		}

		close(jobs)
	}(jobs)

	for result := range results {
		fmt.Printf("result: %d\n", result)
	}
}

func worker(workerID int, jobs <-chan int, results chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		fmt.Printf("worker %d is processing job %d\n", workerID, job)
		results <- 2 * job
	}
}
