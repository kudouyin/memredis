package memredis

import (
	"syscall"
	"net"
	"fmt"
)

func Run () {
	port := 3003
	addr := syscall.SockaddrInet4{Port: port}
	copy(addr.Addr[:], net.ParseIP("0.0.0.0").To4())
	fmt.Println(addr.Addr[:])
	peers := NewServerPeers("0.0.0.0:3003" , nil)
	//peers.SetPeers("0.0.0.0:3022", "0.0.0.0:3023")

	commandHandler := NewCommandHandler(peers, Cachetable)

	//init worker pool
	workerpool := NewWorkerPool(workerln)
	workerpool.Run(commandHandler)

	socket_server := &SocketServer{&addr}
	socket_server.Serve()
}