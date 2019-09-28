package memredis

import (
	"fmt"
	"net"
	"strconv"
	"syscall"
)

func Run (port *int, gossipPort *int, seedNodeAddr *string) {
	addr := syscall.SockaddrInet4{Port: *port}
	copy(addr.Addr[:], net.ParseIP("0.0.0.0").To4())
	fmt.Println(addr.Addr)

	addrString := "0.0.0.0:"+ strconv.Itoa(*port)
	peers := NewServerPeers(addrString, *gossipPort, *seedNodeAddr,nil)
	// when use gossip to manage nodes, we will not to set peer by manual
	//peers.SetPeers(addrString)

	cacheTable := NewCacheTable()
	protocolHandler := NewProtocolHandler(peers, cacheTable)

	//init worker pool
	workerPool := NewWorkerPool(workerln)
	workerPool.Run(protocolHandler)

	socketServer := &SocketServer{&addr}
	socketServer.Serve()
}