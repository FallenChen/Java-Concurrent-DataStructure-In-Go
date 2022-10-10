package main

import (
	"context"
	"fmt"
	"github.com/enriquebris/goconcurrentqueue"
	"time"
)

func main() {
	var (
		fifo        = goconcurrentqueue.NewFIFO()
		ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	)
	defer cancel()

	fmt.Println("1 - Waiting for next enqueued element")
	_, err := fifo.DequeueOrWaitForNextElementContext(ctx)

	if err != nil {
		fmt.Printf("2 - Failed waiting for new element: %v\n", err)
		return
	}
}