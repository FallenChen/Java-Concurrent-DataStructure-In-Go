package threadpool

import "fmt"

var (
	ErrQueueFull = fmt.Errorf("queue is full, not able add the task")
)

type ThreadPool struct {
	queueSize	int64
	// number of workers
	noOfWorkers	int

	jobQueue	chan interface{}
	workerPool	chan chan interface{}
	// Channel used to stop all the workers
	closeHandle	chan bool
}

// NewThreadPool creates thread threadpool
func NewThreadPool(noOfWorkers int, queueSize int64) *ThreadPool {
	threadPool := &ThreadPool{queueSize: queueSize, noOfWorkers: noOfWorkers}
	threadPool.jobQueue = make(chan interface{}, queueSize)
	threadPool.workerPool = make(chan chan interface{}, noOfWorkers)
	threadPool.closeHandle = make(chan bool)
	threadPool.createPool()
	return threadPool
}

func (t *ThreadPool) createPool() {
	for i:=0; i<t.noOfWorkers; i++ {
		worker := NewWorker(t.workerPool, t.closeHandle)
		worker.Start()
	}
	go t.dispatch()
}

// dispatch listens to the jobqueue and handles the jobs to the workers
func (t *ThreadPool) dispatch() {
	for {
		select {
		case job := <-t.jobQueue:
			func(job interface{}) {
				// find a worker for the job
				jobChannle := <-t.workerPool
				// submit job to the worker
				jobChannle <- job
			}(job)

		case <- t.closeHandle:
			return
		}
	}
}

func (t *ThreadPool) submitTask (task interface{}) error {
	if len(t.jobQueue) == int(t.queueSize) {
		return ErrQueueFull
	}
	t.jobQueue <- task
	return nil
}

// Execute submits the job to available worker
func (t *ThreadPool) Execute(task Runnable) error {
	return t.submitTask(task)
}

func (t *ThreadPool) ExecuteFuture(task Callable) (*Future, error) {
	// Create future and task
	handle := &Future{response: make(chan interface{})}
	futureTask := callableTask{Task: task, Handle: handle}
	err := t.submitTask(futureTask)
	if err != nil {
		return nil, err
	}
	return futureTask.Handle, nil

}

func (t *ThreadPool) Close() {
	close(t.closeHandle) // Stops all the routines
	close(t.workerPool)  // Closes the Job threadpool
	close(t.jobQueue)    // Closes the job Queue
}