package memredis

import (
	"syscall"
	"fmt"
	"os"
)

type WorkerHandler interface {
	handle(connFd int)
}


type Worker struct {
	event_base_fd int
	eventChan chan int
	handler WorkerHandler
}

func NewWorker(handler WorkerHandler) *Worker {
	worker := &Worker{
		eventChan: make(chan int),
		handler: handler,
	}
	epfd, e := syscall.EpollCreate1(0)
	if e != nil {
		fmt.Println("epoll_create1 error: ", e)
		os.Exit(1)
	}
	worker.event_base_fd = epfd
	//fmt.Println("new worker and epollfd success")

	return worker
}

func (w *Worker) Run() {
	go func() {
		for {
			fmt.Println("into for")
			var events [MaxEpollEvents]syscall.EpollEvent
			nevents, e := syscall.EpollWait(w.event_base_fd, events[:], -1)
			if e != nil {
				fmt.Println("epoll wait error:", e)
				os.Exit(1)
			}
			for ev := 0; ev < nevents; ev++ {
				w.handler.handle(int(events[ev].Fd))
			}
		}
	}()
}
