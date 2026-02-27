package main

import (
	"fmt"
	"time"
)

func main() {
	urgent := make(chan int)
	normal := make(chan int)

	go func() {
		for id := range 3 {
			time.Sleep(300 * time.Millisecond)
			urgent <- id
		}
		close(urgent)
	}()

	go func() {
		for id := range 10 {
			time.Sleep(100 * time.Millisecond)
			normal <- id
		}
		close(normal)
	}()

	urgentCount := 0
	normalCount := 0

	for {
		if urgent == nil && normal == nil {
			fmt.Printf("finished processing, exitting...\n")
			return
		}

		select {
		case msg, ok := <-urgent:
			if !ok {
				urgent = nil
				continue
			}
			urgentCount += 1
			fmt.Printf("processing urgent message %d, porcessed: %d\n", msg, urgentCount)
		default:
			select {
			case msg, ok := <-normal:
				if !ok {
					normal = nil
					continue
				}
				normalCount += 1
				fmt.Printf("processing normal message %d, porcessed: %d\n", msg, normalCount)
			case msg, ok := <-urgent:
				if !ok {
					urgent = nil
					continue
				}
				urgentCount += 1
				fmt.Printf("processing urgent message %d, porcessed: %d\n", msg, urgentCount)
			}
		}
	}
}
