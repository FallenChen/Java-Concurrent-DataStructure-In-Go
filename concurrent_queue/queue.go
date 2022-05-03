package concurrentqueue

import "context"



type Queue interface {

	Enqueue(interface{}) error

	Dequeue() (interface{}, error)

	// DequeueOrWaitForNextElement dequeues an element (if exist) or waits until the next element gets enqueued and returns it.
	// Multiple calls to DequeueOrWaitForNextElement() would enqueue multiple "listeners" for future enqueued elements.
	DequeueOrWaitForNextElement() (interface{}, error)
	// When the passed context expires this function exits and returns the context' error
	DequeueOrWaitForNextElementContext(context.Context) (interface{}, error)

	GetLen()	int

	GetCap()	int

	Lock()

	Unlock()

	IsLocked()	bool
}