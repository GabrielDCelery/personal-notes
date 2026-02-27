package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	res, err := slowOperation(ctx)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Fatalf("context deadline exceeded")
		} else {
			log.Fatalf("failed to run slow operation, %v", err)
		}
	}
	fmt.Println(res)
}

func slowOperation(ctx context.Context) (string, error) {
	select {
	case <-time.After(300 * time.Millisecond):
		return "operation completed", nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}
