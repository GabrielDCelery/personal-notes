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
	signalCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	g, ctx := errgroup.WithContext(signalCtx)

	nums := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	generatorChan := generator(ctx, nums, g)
	transformChan := transform(ctx, generatorChan, g)
	save(ctx, transformChan, g)

	if err := g.Wait(); err != nil {
		fmt.Printf("pipeline error: %v\n", err)
		return
	}

	fmt.Printf("successfully finished processing\n")
}

func generator(ctx context.Context, nums []int, g *errgroup.Group) <-chan int {
	outChan := make(chan int)
	g.Go(func() error {
		defer close(outChan)
		for _, num := range nums {
			time.Sleep(100 * time.Millisecond)
			select {
			case <-ctx.Done():
				return nil
			case outChan <- num:
			}
		}
		return nil
	})
	return outChan
}

func transform(ctx context.Context, inChan <-chan int, g *errgroup.Group) <-chan int {
	outChan := make(chan int)

	g.Go(func() error {
		defer close(outChan)
		for num := range inChan {
			if num == 6 {
				return fmt.Errorf("transform error: number %d is invalid", num)
			}
			time.Sleep(100 * time.Millisecond)
			select {
			case <-ctx.Done():
				return nil
			case outChan <- num * 2:
			}
		}
		return nil
	})

	return outChan
}

func save(ctx context.Context, inChan <-chan int, g *errgroup.Group) {
	g.Go(func() error {
		for num := range inChan {
			select {
			case <-ctx.Done():
				return nil
			default:
				fmt.Printf("saved: %d\n", num)
			}
		}
		return nil
	})
}
