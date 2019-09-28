package memredis

import (
	"fmt"
	"syscall"
)

var connFdChan chan int = make(chan int)


type WorkerPool struct {
	workerlen int
	WorkerQueue chan *Worker
}

func NewWorkerPool(worklen int) *WorkerPool {
	return &WorkerPool{
		workerlen: worklen,
		// buffered channel
		WorkerQueue: make(chan *Worker, workerln),
	}
}
func (wp *WorkerPool) Run(handler WorkerHandler) {
	for i:= 0; i < wp.workerlen; i ++ {
		//fmt.Println("new a worker, index: ", i)
		worker, err := NewWorker(handler)
		if err != nil {
			fmt.Printf("create %d worker failed\n", i)
		}else{
			// 放入workerQueue
			wp.WorkerQueue <- worker
			worker.Run()
		}
	}
	go wp.Dispatch()
}

func (wp *WorkerPool) Dispatch() {
	for {
		select {
		case connFd:= <- connFdChan:
			fmt.Println("recevie a new fd: ", connFd)
			// get a worker'
			worker := <- wp.WorkerQueue

			// register to event base
			err := event_add(connFd, worker.event_base_fd)
			if err != nil {
				fmt.Printf("connFd: %d cannot be added into worker's event loop, so will be closed\n", connFd)
				syscall.Close(connFd)
			}

			wp.WorkerQueue <- worker
		}
	}
}
