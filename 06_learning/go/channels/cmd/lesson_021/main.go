package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	g0 := make(chan int)
	g1, g2 := tee(ctx, g0)
	go func() {
		for n := range 50 {
			g0 <- n
		}
	}()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for n := range g1 {
			fmt.Printf("fast %d\n", n)
		}
	}()

	go func() {
		defer wg.Done()
		for n := range g2 {
			time.Sleep(100 * time.Millisecond)
			fmt.Printf("slow %d\n", n)
		}
	}()

	wg.Wait()

	fmt.Printf("done\n")

}

func tee[T any](ctx context.Context, in <-chan T) (<-chan T, <-chan T) {
	out1 := make(chan T)
	out2 := make(chan T)

	go func() {
		defer close(out1)
		defer close(out2)

		for {
			select {
			case <-ctx.Done():
				return
			case val, ok := <-in:
				if !ok {
					return
				}
				out1 <- val
				out2 <- val
			}
		}

	}()

	return out1, out2
}
