package memredis

import (
	"syscall"
	"bytes"
	"time"
	"strconv"
	"fmt"
)
var separatorBytes = []byte(" ")

type CommandHandler struct {
	peerPicker PeerPicker
	cachetable *CacheTable
}

func NewCommandHandler(peerPicker PeerPicker, table *CacheTable) *CommandHandler {
	return &CommandHandler{
		peerPicker: peerPicker,
		cachetable: table,
	}
}


func (commandHandler *CommandHandler) handle(connFd int) {
	params := commandHandler.readCommand(connFd)
	key := string(params[1])
	peer, ok := commandHandler.peerPicker.PickPeer(key)
	if !ok {
		fmt.Println("cannot find peer")
		isok := commandHandler.Exec(params)
		commandHandler.writeResult(connFd, isok)
	}else{
		fmt.Println("find peer", peer)
		commandHandler.transmit(peer)
	}
	check()
}

func check() {
	for k, v := range Cachetable.items {
		fmt.Println("cache ", k, v.data)
	}
}

func (commandHandler *CommandHandler) transmit(peer string) {

}

func (commandHandler *CommandHandler) readCommand(connFd int) (params [][]byte){
	var buf [32 * 1024]byte
	nbytes, err := syscall.Read(connFd, buf[:])
	if err != nil {
		return nil
	}
	if nbytes > 0{
		params := bytes.Split(buf[:nbytes], separatorBytes)
		fmt.Println("params is ", string(params[1]))
		return params
	}
	return nil
}

func (CommandHandler *CommandHandler) writeResult(connFd int, ok bool) {
	syscall.Write(connFd, []byte(strconv.FormatBool(ok)))
}


func (commandHandler *CommandHandler) lookupCache(key string) (interface{}, bool){
	value, err := commandHandler.cachetable.Get(key)
	if err != nil {
		return nil, false
	}
	return value, true

}

func (commandHandler *CommandHandler) Exec(params [][]byte) (ok bool){
	switch  {
	case bytes.Equal(params[0], []byte("SET")):
		return commandHandler.SET(params)
	}
	return false
}

func (commandHandler *CommandHandler) SET(params [][]byte) (ok bool){
	key := string(params[1])
	lifeSpan, err:= strconv.Atoi(string(params[2]))
	if err != nil {
		fmt.Println("参数转换错误")
		return false
	}
	value := string(params[3])
	fmt.Println("params is ", key, time.Duration(lifeSpan), value)
	return commandHandler.cachetable.Set(key, time.Duration(lifeSpan), value)
}
