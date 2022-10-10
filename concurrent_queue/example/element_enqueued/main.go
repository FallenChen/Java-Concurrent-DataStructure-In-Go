package main

import (
	"fmt"
	"github.com/enriquebris/goconcurrentqueue"
	"time"
)

func main() {
	var (
		fifo = goconcurrentqueue.NewFIFO()
		done = make(chan struct{})
	)

	go func() {
		fmt.Println("1 - Waiting for next enqueued element")
		value, _ := fifo.DequeueOrWaitForNextElement()
		fmt.Println("2 - Dequeued element: %v\n", value)

		done <- struct{}{}
	}()

	fmt.Println("3 - Go to sleep for 3 seconds")
	time.Sleep(3 * time.Second)

	fmt.Println("4 - Enqueue element")
	fifo.Enqueue(100)

	<-done

}
