package threadpool

import (
	"sync"
	"time"
)


type ScheduledThreadPool struct {
	workers     chan chan interface{}
	tasks       *sync.Map
	noOfWorkers int
	counter     uint64
	counterLock sync.Mutex
	closeHandle chan bool
}

func NewScheduledThreadPool(noOfWorkers int) *ScheduledThreadPool {
	pool := &ScheduledThreadPool{}
	pool.noOfWorkers = noOfWorkers
	pool.workers = make(chan chan interface{}, noOfWorkers)
	pool.tasks = new(sync.Map)
	pool.closeHandle = make(chan bool)
	pool.createPool()
	return pool
}

func (stf *ScheduledThreadPool) createPool() {
	for i := 0; i < stf.noOfWorkers; i++ {
		worker := NewWorker(stf.workers, stf.closeHandle)
		worker.Start()
	}

	go stf.dispatch()
}

func (stf *ScheduledThreadPool) dispatch() {
	for {
		select {
		case <-stf.closeHandle:
			//Stop the scheduler
			return
		default:
			go stf.intervalRunner()    
			time.Sleep(time.Second * 1) 
		}
	}
}

func (stf *ScheduledThreadPool) intervalRunner() {

	stf.updateCounter()

	currentTasksToRun, ok := stf.tasks.Load(stf.counter)

	if ok {
		currentTasksSet := currentTasksToRun.(*Set)

		// For each tasks , get a worker from the threadpool and run the task
		for _,val := range currentTasksSet.GetAll() {
			go func(job interface{}) {
				worker :=<-stf.workers
				worker <-job
			}(val)
		}
	}
}

func (stf *ScheduledThreadPool) updateCounter() {
	stf.counterLock.Lock()
	defer stf.counterLock.Unlock()
	stf.counter++
}

func (stf *ScheduledThreadPool) ScheduleOnce(task Runnable, delay time.Duration) {
	scheduleTime := stf.counter + uint64(delay.Seconds())
	existingTasks, ok := stf.tasks.Load(scheduleTime)

	// Create new set if no tasks are already there
	if !ok {
		existingTasks = NewSet()
		stf.tasks.Store(scheduleTime, existingTasks)
	}
	// Add task
	existingTasks.(*Set).Add(task)
}

func (stf *ScheduledThreadPool) Close() {
	close(stf.closeHandle)
}