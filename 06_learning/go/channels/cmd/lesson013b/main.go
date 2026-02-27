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

	g.Go(func() error {
		select {
		case <-time.After(5 * time.Second):
			fmt.Printf("worker finished processing\n")
			return nil
		case <-ctx.Done():
			fmt.Printf("worker shutting down\n")
			return nil
		}
	})

	g.Go(func() error {
		select {
		case <-time.After(5 * time.Second):
			fmt.Printf("worker finished processing\n")
			return nil
		case <-ctx.Done():
			fmt.Printf("worker shutting down\n")
			return nil
		}
	})

	g.Go(func() error {
		select {
		case <-time.After(3 * time.Second):
			return fmt.Errorf("worker failed\n")
		case <-ctx.Done():
			fmt.Printf("worker shutting down\n")
			return nil
		}
	})

	err := g.Wait()

	if err == nil && signalCtx.Err() != nil {
		fmt.Printf("shutdown: received signal\n")
	} else if err != nil {
		fmt.Printf("shutdown: worker error, reason: %v\n", err)
	} else {
		fmt.Printf("workers completed successfully\n")
	}

}
