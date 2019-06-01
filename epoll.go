// +build linux

package memredis

import (
	"syscall"
	"fmt"
	"os"
)

func event_base_create() int {
	epfd, e := syscall.EpollCreate1(0)
	if e != nil {
		fmt.Println("create event_base error: ", e)
		os.Exit(1)
	}
	return epfd
}

func event_add(connFd int, event_base_fd int) {
	var event syscall.EpollEvent
	event.Events = syscall.EPOLLIN | EPOLLET
	event.Fd = int32(connFd)
	if e := syscall.EpollCtl(event_base_fd, syscall.EPOLL_CTL_ADD, connFd, &event); e != nil {
		fmt.Println("add event error: ", e)
		os.Exit(1)
	}
}

func event_wait(event_base_fd int) (int, [MaxEpollEvents]syscall.EpollEvent){
	var events [MaxEpollEvents]syscall.EpollEvent
	nevents, e := syscall.EpollWait(event_base_fd, events[:], -1)
	if e != nil {
		fmt.Println("epoll wait error:", e)
		os.Exit(1)
	}
	return nevents, events
}