// +build linux

package memredis

import (
	"syscall"
	"fmt"
)

func event_base_create() (int, error) {
	epfd, e := syscall.EpollCreate1(0)
	if e != nil {
		fmt.Println("create event_base error: ", e)
		return -1, ErrCreateEventBase
	}
	return epfd, nil
}

func event_add(connFd int, event_base_fd int) error {
	var event syscall.EpollEvent
	event.Events = syscall.EPOLLIN | EPOLLET
	event.Fd = int32(connFd)
	if e := syscall.EpollCtl(event_base_fd, syscall.EPOLL_CTL_ADD, connFd, &event); e != nil {
		fmt.Println("add event error: ", e)
		return  ErrAddEvent
	}
	return nil
}

func event_wait(event_base_fd int) (int, *[MaxEpollEvents]int, error){
	var epoll_events [MaxEpollEvents]syscall.EpollEvent
	nevents, e := syscall.EpollWait(event_base_fd, epoll_events[:], -1)
	if e != nil {
		fmt.Println("epoll wait error:", e)
		return 0, nil, ErrWaitEvent
	}
	var eventFds [MaxEpollEvents] int
	for i := 0; i < nevents; i++ {
		eventFds[i] = int(epoll_events[i].Fd)
	}
	return nevents, &eventFds, nil
}