// +build darwin

package memredis

import (
	"fmt"
	"syscall"
)

func event_base_create() (int, error){
	epfd, e := syscall.Kqueue()
	if e != nil {
		fmt.Println("Kqueue_create: ", e)
		return -1, ErrCreateEventBase
	}
	return epfd, nil
}

func event_add(connFd int, event_base_fd int) error {
	var changes [] syscall.Kevent_t = make([]syscall.Kevent_t, 1)
	changes[0].Ident = uint64(connFd)
	changes[0].Filter = syscall.EVFILT_READ
	changes[0].Flags = syscall.EV_ADD | syscall.EV_CLEAR
	changes[0].Fflags = 0
	changes[0].Data = 0
	fmt.Println(changes)

	if _, e := syscall.Kevent(event_base_fd, changes, nil, nil); e != nil {
		fmt.Println("kevent 0: ", e)
		return ErrAddEvent
	}
	return nil
}

func event_wait(event_base_fd int) (int, *[MaxEpollEvents]int, error){
	var kevents [MaxEpollEvents]syscall.Kevent_t
	nevents, e := syscall.Kevent(event_base_fd, nil, kevents[:], nil)
	if e != nil {
		fmt.Println("kevent: ", e)
		return 0, nil, ErrWaitEvent
	}
	var eventFds [MaxEpollEvents] int
	for i := 0; i < nevents; i++ {
		eventFds[i] = int(kevents[i].Ident)
	}
	return nevents, &eventFds, nil
}