package memredis

import (
	"fmt"
)

type WorkerHandler interface {
	handle(connFd int)
}


type Worker struct {
	event_base_fd int
	eventChan chan int
	handler WorkerHandler
}

func NewWorker(handler WorkerHandler) (*Worker, error) {
	worker := &Worker{
		eventChan: make(chan int),
		handler: handler,
	}
	epfd, err := event_base_create()
	if err != nil {
		return nil, err
	}
	worker.event_base_fd = epfd
	//fmt.Println("new worker and epollfd success")

	return worker, nil
}

func (w *Worker) Run() {
	go func() {
		for {
			fmt.Println("into for")
			nevents, events, _ := event_wait(w.event_base_fd)
			for ev := 0; ev < nevents; ev++ {
				w.handler.handle(int(events[ev].Fd))
			}
		}
	}()
}
