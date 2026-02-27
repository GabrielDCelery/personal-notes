package main

import (
	"context"
	"fmt"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	var wg sync.WaitGroup
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	defer stop()

	for id := range 3 {
		wg.Add(1)
		go worker(ctx, id+1, &wg)
	}

	wg.Wait()

	fmt.Printf("all workers stopped, exitting\n")
}

func worker(ctx context.Context, id int, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("worked %d: shutting down\n", id)
			return
		case <-time.After(500 * time.Millisecond):
			fmt.Printf("worker %d: processing...\n", id)
		}
	}
}
