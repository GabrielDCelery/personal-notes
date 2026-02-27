package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

func main() {
	signalCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	g, ctx := errgroup.WithContext(signalCtx)

	jobs := make(chan int)

	for workerID := range 3 {
		g.Go(func() error {
			return worker(ctx, workerID, jobs)
		})
	}

	g.Go(func() error {
		defer close(jobs)
		for job := range 10 {
			select {
			case <-ctx.Done():
				return nil
			case jobs <- job:
			}
		}
		return nil
	})

	err := g.Wait()

	if signalCtx.Err() != nil && err == nil {
		fmt.Printf("detected termination signal, shut down process\n")
		return
	}

	if err != nil {
		fmt.Printf("shutdown on worker error: %v\n", err)
		return
	}

	fmt.Printf("finished processing jobs\n")
}

func worker(ctx context.Context, workerID int, jobs <-chan int) error {
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("worker %d shutting down\n", workerID)
			return nil
		case job, ok := <-jobs:
			if !ok {
				return nil // jobs channel closed
			}
			if job == 7 {
				return fmt.Errorf("job %d failed", job)
			}
			fmt.Printf("worker %d processing job %d\n", workerID, job)
			time.Sleep(1 * time.Second)
		}
	}
}
