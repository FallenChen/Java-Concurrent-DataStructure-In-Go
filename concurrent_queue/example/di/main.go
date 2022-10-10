package main

import (
	"fmt"
	"github.com/enriquebris/goconcurrentqueue"
)

func main() {
	var (
		queue          goconcurrentqueue.Queue
		dummyCondition = true
	)

	if dummyCondition {
		queue = goconcurrentqueue.NewFIFO()
	} else {
		queue = goconcurrentqueue.NewFixedFIFO(10)
	}

	fmt.Printf("queue's length: %v\n", queue.GetLen())
	workWithQueue(queue)
	fmt.Printf("queue's length: %v\n", queue.GetLen())
}

func workWithQueue(queue goconcurrentqueue.Queue) error {
	// do some work

	// enqueue an item
	if err := queue.Enqueue("test value"); err != nil {
		return err
	}

	return nil
}
