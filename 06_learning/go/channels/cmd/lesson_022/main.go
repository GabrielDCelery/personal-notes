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
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	chanOfChans := make(chan (<-chan int))
	go func() {
		defer close(chanOfChans)
		for id := range 3 {
			chanOfChans <- worker(id, ctx)
		}
	}()
	output := bridge(ctx, chanOfChans)
	for num := range output {
		fmt.Printf("received: %d\n", num)
	}
	fmt.Printf("done\n")
}

func bridge[T any](ctx context.Context, chanOfChans <-chan (<-chan T)) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		var wg sync.WaitGroup
		for ch := range chanOfChans {
			wg.Add(1)
			go func(ch <-chan T) {
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
						case out <- val:
						}
					}
				}
			}(ch)
		}
		wg.Wait()
	}()
	return out
}

func worker(id int, ctx context.Context) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range 5 {
			time.Sleep(100 * time.Millisecond)
			fmt.Printf("worker %d: %d\n", id, n)
			select {
			case <-ctx.Done():
				return
			case out <- n:
			}
		}
	}()
	return out
}
