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
	ctx, close := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer close()
	g1 := generator(ctx)
	g2 := generator(ctx)
	g3 := generator(ctx)

	merged := merge(ctx, orDone(ctx, g1), orDone(ctx, g2), orDone(ctx, g3))

	for n := range merged {
		fmt.Printf("received %d\n", n)
	}

	fmt.Printf("done\n")
}

func generator(ctx context.Context) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		nums := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
		for _, num := range nums {
			time.Sleep(50 * time.Millisecond)
			select {
			case <-ctx.Done():
				return
			case out <- num:
			}
		}
	}()
	return out
}

func orDone[T any](ctx context.Context, ch <-chan T) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case val, ok := <-ch:
				if !ok {
					return
				}
				select {
				case <-ctx.Done():
					return
				case out <- val:
				}
			}
		}
	}()
	return out
}

func merge[T any](ctx context.Context, chs ...<-chan T) <-chan T {
	merged := make(chan T)
	var wg sync.WaitGroup
	wg.Add(len(chs))
	for _, ch := range chs {
		go func(<-chan T) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case val, ok := <-ch:
					if !ok {
						return
					}
					select {
					case <-ctx.Done():
						return
					case merged <- val:
					}
				}
			}
		}(ch)
	}
	go func() {
		defer close(merged)
		wg.Wait()
	}()
	return merged
}
