package threadpool

// Worker type holds the job channel and passed worker threadpool
type Worker struct {
	jobChannel		chan interface{}
	workerPool		chan chan interface{}
	closeHandle		chan bool
}

func NewWorker (workerPool chan chan interface{}, closeHandle chan bool) *Worker {
	return &Worker{workerPool: workerPool, jobChannel: make(chan interface{}), closeHandle: closeHandle}
}

func (w Worker) Start() {
	go func() {
		for {
			// Put the worker to the worker threadpool
			w.workerPool <- w.jobChannel

			select {
				// Wait for the job
			case job := <-w.jobChannel:
			// Got the job
			w.executeJob(job)
			case <- w.closeHandle:
				return
			}
		}
	}()
}

func (w Worker) executeJob(job interface{}) {
	switch task := job.(type) {
	case Runnable :
		task.Run()
		break
	case callableTask:
		response := task.Task.Call()
		task.Handle.done = true
		task.Handle.response <- response
		break
	}
}