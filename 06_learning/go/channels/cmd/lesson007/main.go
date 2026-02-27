package main

import (
	"fmt"
)

const numOfJobs = 5

func main() {
	jobs := make(chan int, 10)
	results := make(chan int, 10)

	for workerID := range 3 {
		go worker(workerID, jobs, results)
	}

	for num := range numOfJobs {
		jobs <- (num + 1)
	}

	close(jobs)

	for range numOfJobs {
		result := <-results
		fmt.Printf("result: %d\n", result)
	}

	close(results)
}

func worker(workerID int, jobs <-chan int, results chan<- int) {
	for job := range jobs {
		fmt.Printf("worker %d is processing job %d\n", workerID, job)
		results <- 2 * job
	}
}
