package main

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"
)

func main() {
	g, ctx := errgroup.WithContext(context.Background())

	g.Go(func() error {
		select {
		case <-time.After(200 * time.Millisecond):
			return fmt.Errorf("failed to run worker")
		case <-ctx.Done():
			fmt.Printf("closing worker\n")
			return nil
		}
	})

	g.Go(func() error {
		select {
		case <-time.After(500 * time.Millisecond):
			fmt.Printf("finished fetching data")
			return nil
		case <-ctx.Done():
			fmt.Printf("closing worker\n")
			return nil
		}
	})

	g.Go(func() error {
		select {
		case <-time.After(500 * time.Millisecond):
			fmt.Printf("finished fetching data")
			return nil
		case <-ctx.Done():
			fmt.Printf("closing worker\n")
			return nil
		}
	})

	if err := g.Wait(); err != nil {
		fmt.Printf("encountered error, shut down workers\n")
	}
}
