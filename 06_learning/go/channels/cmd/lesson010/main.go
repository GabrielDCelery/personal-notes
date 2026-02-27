package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	gen := generator(ctx)
	for range 5 {
		value := <-gen
		fmt.Printf("%d\n", value)
	}
	cancel()
}

func generator(ctx context.Context) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		counter := 0
		for {
			time.Sleep(100 * time.Millisecond)
			select {
			case <-ctx.Done():
				fmt.Printf("generator stopped\n")
				return
			case out <- counter:
				counter += 1
			}
		}
	}()
	return out
}
