package memredis

import (
	"syscall"
	"net"
	"fmt"
	"strconv"
)

func Run () {
	port := 3011
	addr := syscall.SockaddrInet4{Port: port}
	copy(addr.Addr[:], net.ParseIP("0.0.0.0").To4())
	fmt.Println(addr.Addr[:])
	peers := NewServerPeers("0.0.0.0"+strconv.Itoa(port) , nil)
	//peers.SetPeers("0.0.0.0:3022", "0.0.0.0:3023")

	cacheTable := NewCacheTable()
	protocolHandler := NewProtocolHandler(peers, cacheTable)

	//init worker pool
	workerPool := NewWorkerPool(workerln)
	workerPool.Run(protocolHandler)

	socketServer := &SocketServer{&addr}
	socketServer.Serve()
}