package main

import (
	"fmt"
	"garry.org/data_structure/threadpool"
	"time"
)

func main() {
	pool := threadpool.NewThreadPool(200, 1000)
	time.Sleep(20 * time.Second)
	task := &myTask{ID: 123}
	pool.Execute(task)
}

type myTask struct {
	ID int64
}

func (m *myTask) Run() {
	fmt.Println("Running my task ", m.ID)
}
