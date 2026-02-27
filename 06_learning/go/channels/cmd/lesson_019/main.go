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
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	nums := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	genChan := generator(ctx, nums)
	transChan, transErrChan := transform(ctx, genChan)
	doneChan, saveErrChan := save(ctx, transChan)
	mergedErrChan := mergeErrorChannels(ctx, transErrChan, saveErrChan)

	for {
		select {
		case err, ok := <-mergedErrChan:
			if !ok {
				mergedErrChan = nil
				continue
			}
			fmt.Printf("error: %v\n", err)
			cancel()
		case <-doneChan:
			fmt.Printf("finished processing\n")
			return
		}
	}
}

func mergeErrorChannels(ctx context.Context, errChans ...<-chan error) <-chan error {
	merged := make(chan error)

	var wg sync.WaitGroup
	wg.Add(len(errChans))

	for _, errChan := range errChans {
		go func(e <-chan error) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case err, ok := <-e:
					if !ok {
						return
					}
					merged <- err
					return
				}
			}
		}(errChan)
	}

	go func() {
		defer close(merged)
		wg.Wait()
	}()

	return merged
}

func generator(ctx context.Context, nums []int) <-chan int {
	outChan := make(chan int)
	go func() {
		defer close(outChan)
		for _, num := range nums {
			select {
			case <-ctx.Done():
				return
			case outChan <- num:
			}
		}
	}()
	return outChan
}

func transform(ctx context.Context, inChan <-chan int) (<-chan int, <-chan error) {
	outChan := make(chan int)
	errChan := make(chan error)
	go func() {
		defer close(errChan)
		defer close(outChan)
		for {
			time.Sleep(100 * time.Millisecond)
			select {
			case <-ctx.Done():
				return
			case num, ok := <-inChan:
				if !ok {
					return
				}
				if num == 6 {
					errChan <- fmt.Errorf("number %d is invalid", num)
					return
				}
				outChan <- num * 2
			}
		}
	}()
	return outChan, errChan

}

func save(ctx context.Context, inChan <-chan int) (<-chan struct{}, <-chan error) {
	doneChan := make(chan struct{})
	errChan := make(chan error)
	go func() {
		defer close(errChan)
		defer close(doneChan)
		for {
			time.Sleep(100 * time.Millisecond)
			select {
			case <-ctx.Done():
				return
			case num, ok := <-inChan:
				if !ok {
					return
				}
				fmt.Printf("saved %d\n", num)
			}
		}
	}()
	return doneChan, errChan
}
