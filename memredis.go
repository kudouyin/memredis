package memredis

import (
	"syscall"
	"net"
)

func Run () {
	addr := syscall.SockaddrInet4{Port: 3022}
	copy(addr.Addr[:], net.ParseIP("0.0.0.0").To4())
	peers := NewServerPeers("0.0.0.0:3018", nil)
	//peers.SetPeers("0.0.0.0:3015", "0.0.0.0:3011")

	commandHandler := NewCommandHandler(peers, Cachetable)

	//init worker pool
	workerpool := NewWorkerPool(workerln)
	workerpool.Run(commandHandler)

	socket_server := &SocketServer{&addr}
	socket_server.Serve()
}