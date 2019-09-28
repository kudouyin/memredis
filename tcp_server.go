package memredis

import (
	"net"
	"runtime"
	"fmt"
)

type TCPHandler interface {
	Handle(net.Conn)
}


func TCPServer(listener net.Listener, handler TCPHandler) {

	for {
		clientConn, err := listener.Accept()

		if err != nil {
			fmt.Println("conn error")
			runtime.Gosched()
			continue
		}
		go handler.Handle(clientConn)
	}
}