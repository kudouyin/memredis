package memredis

import (
	"syscall"
	"fmt"
	"os"
)

const (
	LISTENN = 100
	EPOLLET        = 1 << 31
	MaxEpollEvents = 320
	workerln = 1
)

type SocketServer struct {
	Addr *syscall.SockaddrInet4
}

func (s *SocketServer) Serve() {
	//create socket and bind addr
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer syscall.Close(fd)

	if err = syscall.SetNonblock(fd, true); err != nil {
		fmt.Println("setnonblock1: ", err)
		os.Exit(1)
	}

	fmt.Println("addr:", s.Addr)
	syscall.Bind(fd, s.Addr)
	syscall.Listen(fd, LISTENN)

	// create epoll fd
	epfd, err:= event_base_create()
	if err != nil {
		os.Exit(1)
	}
	defer syscall.Close(epfd)

	err = event_add(fd, epfd)
	if err != nil {
		os.Exit(1)
	}
	for {
		nevents, eventFds, _:= event_wait(epfd)
		for ev := 0; ev < nevents; ev ++ {
			if int(eventFds[ev]) == fd {
				connFd, _, err := syscall.Accept(fd)
				defer syscall.Close(connFd)
				if err != nil {
					fmt.Println("accept error: ", err)
					continue
				}
				syscall.SetNonblock(connFd, true)
				fmt.Println("accept success")
				connFdChan <- connFd
			}
		}
	}

}
