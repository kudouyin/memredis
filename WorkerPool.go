package memredis

import (
	"fmt"
	"syscall"
	"os"
)

var connFdChan chan int = make(chan int)


type WorkerPool struct {
	workerlen int
	WorkerQueue chan *Worker
}

func NewWorkerPool(worklen int) *WorkerPool {
	return &WorkerPool{
		workerlen: worklen,
		WorkerQueue: make(chan *Worker, workerln),
	}
}
func (wp *WorkerPool) Run(handler WorkerHandler) {
	for i:= 0; i < wp.workerlen; i ++ {
		//fmt.Println("new a worker, index: ", i)
		worker := NewWorker(handler)
		// 放入workerQueue
		wp.WorkerQueue <- worker
		worker.Run()
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
			fmt.Println(worker)

			// register to event base
			var event syscall.EpollEvent
			event.Events = syscall.EPOLLIN | EPOLLET
			event.Fd = int32(connFd)
			fmt.Println(worker.event_base_fd)
			if e := syscall.EpollCtl(worker.event_base_fd, syscall.EPOLL_CTL_ADD, connFd, &event); e != nil {
				fmt.Println("worker epoll_ctl error: ", e)
				os.Exit(1)
			}
			wp.WorkerQueue <- worker
		}
	}
}
