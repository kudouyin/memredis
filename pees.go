package memredis

import (
	"github.com/kudouyin/memredis/consistenthash"
	"sync"
	"fmt"
)
var defaultReplicas = 50

type PeerPicker interface {
	PickPeer(key string) (string, bool)
}


type ServerPeers struct {
	mu sync.Mutex
	self string
	peers *consistenthash.Map
	//commandSenders map[string]*CommandSender
	opts PeerOptions
}

type PeerOptions struct {
	Replicas int
	HashFn consistenthash.Hash
}

func NewServerPeers(self string, po *PeerOptions) *ServerPeers{
	sp := &ServerPeers{
		self: self,
		//commandSenders: make(map[string]*CommandSender),
	}
	if po != nil {
		sp.opts = *po
	}
	if sp.opts.Replicas == 0 {
		sp.opts.Replicas = defaultReplicas
	}
	sp.peers = consistenthash.New(sp.opts.Replicas, sp.opts.HashFn)
	return sp
}

func (serverPeers *ServerPeers) PickPeer(key string) (string, bool){
	serverPeers.mu.Lock()
	defer serverPeers.mu.Unlock()
	if serverPeers.peers.IsEmpty() {
		return "", false
	}
	fmt.Println("key is ", key, "self is ", serverPeers.self)
	if peer := serverPeers.peers.Get(key); peer != serverPeers.self {
		return peer, true
	}
	return "", false
}

func (serverPeers *ServerPeers) SetPeers(peeraddrs ...string) {
	serverPeers.mu.Lock()
	defer serverPeers.mu.Unlock()
	serverPeers.peers = consistenthash.New(serverPeers.opts.Replicas, serverPeers.opts.HashFn)
	serverPeers.peers.Add(peeraddrs...)
	//serverPeers.commandSenders = make(map[string]*CommandSender, len(peeraddrs))
	//for _, peeraddr := range peeraddrs {
	//	serverPeers.commandSenders[peeraddr] = &CommandSender{
	//		Addr: peeraddr,
	//	}
	//}
}
