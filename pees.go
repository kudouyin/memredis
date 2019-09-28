package memredis

import (
	"fmt"
	"github.com/hashicorp/memberlist"
	"github.com/kudouyin/memredis/consistenthash"
	"sync"
	"time"
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

	// gossip
	gossipPort int
	seedNodeAddr string
	gossipList *memberlist.Memberlist
	updateTimer *time.Timer
}

type PeerOptions struct {
	Replicas int
	HashFn consistenthash.Hash
}

func (serverPeers *ServerPeers) joinGossip(){
	//hostName, _ := os.Hostname()
	config := memberlist.DefaultLocalConfig()
	config.Name = serverPeers.self
	config.BindPort = serverPeers.gossipPort
	config.AdvertisePort = serverPeers.gossipPort

	list, err := memberlist.Create(config)
	_, err = list.Join([]string{serverPeers.seedNodeAddr})
	if err != nil {
		fmt.Println("join error")
	}
	serverPeers.gossipList = list
}

func NewServerPeers(self string, gossipPort int, seedNodeAddr string, po *PeerOptions) *ServerPeers{
	sp := &ServerPeers{
		self: self,
		gossipPort:gossipPort,
		seedNodeAddr:seedNodeAddr,
		//commandSenders: make(map[string]*CommandSender),
	}
	if po != nil {
		sp.opts = *po
	}
	if sp.opts.Replicas == 0 {
		sp.opts.Replicas = defaultReplicas
	}
	sp.peers = consistenthash.New(sp.opts.Replicas, sp.opts.HashFn)
	sp.joinGossip()
	sp.updateTimer = time.AfterFunc(1000*time.Millisecond, func() {
		go sp.UpdatePeers()
	})
	return sp
}

func (serverPeers *ServerPeers) PickPeer(key string) (string, bool){
	serverPeers.mu.Lock()
	defer serverPeers.mu.Unlock()
	if serverPeers.peers.IsEmpty() {
		return "", false
	}
	fmt.Println("key is ", key, "self is ", serverPeers.self)
	peer := serverPeers.peers.Get(key)
	fmt.Println("find peer:", peer)
	if  peer != serverPeers.self {
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

// update hash by gossip
func (serverPeers *ServerPeers) UpdatePeers() {
	serverPeers.mu.Lock()
	defer serverPeers.mu.Unlock()

	if serverPeers.updateTimer != nil {
		serverPeers.updateTimer.Stop()
	}

	//compare whether peers is same with gossipList members, then decide to update hash ,
	//in this way maybe more efficient
	serverPeers.peers = consistenthash.New(serverPeers.opts.Replicas, serverPeers.opts.HashFn)
	for _, member := range serverPeers.gossipList.Members() {
		serverPeers.peers.Add(member.Name)
	}
	serverPeers.updateTimer = time.AfterFunc(1000*time.Millisecond, func() {
		go serverPeers.UpdatePeers()
	})
}

