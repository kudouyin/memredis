// +build darwin

package memredis

import (
	"fmt"
	"os"
	"syscall"
)

func event_base_create() int{
	epfd, e := syscall.Kqueue()
	if e != nil {
		fmt.Println("Kqueue_create: ", e)
		os.Exit(1)
	}
	return epfd
}

func event_add(connFd int, event_base_fd int) {
	var changes [] syscall.Kevent_t = make([]syscall.Kevent_t, 1)
	changes[0].Ident = uint64(connFd)
	changes[0].Filter = syscall.EVFILT_READ
	changes[0].Flags = syscall.EV_ADD | syscall.EV_CLEAR
	changes[0].Fflags = 0
	changes[0].Data = 0
	fmt.Println(changes)


	if _, e := syscall.Kevent(event_base_fd, changes, nil, nil); e != nil {
		fmt.Println("kevent 0: ", e)
		os.Exit(1)
	}
}

func event_wait(event_base_fd int) (int, [MaxEpollEvents]syscall.EpollEvent){
	var events [MaxEpollEvents]syscall.Kevent_t
	nevents, e := syscall.Kevent(event_base_fd, nil, events[:], nil)
	if e != nil {
		fmt.Println("kevent: ", e)
		os.Exit(1)
	}
	return nevents, events
}