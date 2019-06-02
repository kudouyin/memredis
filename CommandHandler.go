package memredis

import (
	"syscall"
	"bytes"
	"time"
	"strconv"
	"fmt"
	"encoding/json"
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
		ok, result := commandHandler.Exec(params)
		commandHandler.writeResult(connFd, ok, result)
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
		return params
	}
	return nil
}

func (CommandHandler *CommandHandler) writeResult(connFd int, ok bool, result string) {

	var buffer bytes.Buffer
	buffer.Write([]byte(strconv.FormatBool(ok)))
	buffer.Write(separatorBytes)
	buffer.Write(([]byte(result)))
	syscall.Write(connFd, buffer.Bytes())
}


func (commandHandler *CommandHandler) lookupCache(key string) (interface{}, bool){
	value, err := commandHandler.cachetable.Get(key)
	if err != nil {
		return nil, false
	}
	return value, true

}

func (commandHandler *CommandHandler) Exec(params [][]byte) (bool, string){
	switch  {
	case bytes.Equal(params[0], []byte("SET")):
		return commandHandler.SET(params), ""
	case bytes.Equal(params[0], []byte("SADD")):
		return commandHandler.SADD(params), ""
	case bytes.Equal(params[0], []byte("GET")):
		return commandHandler.GET(params)

	}
	return false, ""
}

func (commandHandler *CommandHandler) SET(params [][]byte) (ok bool){
	key := string(params[1])
	lifeSpan, err:= strconv.Atoi(string(params[2]))
	if err != nil {
		fmt.Println("参数转换错误")
		return false
	}
	value := string(params[3])
	newLifeSpan := time.Duration(lifeSpan) * time.Second

	fmt.Println("all params is ", key, newLifeSpan, value)
	return commandHandler.cachetable.Set(key, newLifeSpan, value)
}

func (commandHandler *CommandHandler) SADD(params [][]byte) (ok bool){
	key := string(params[1])
	lifeSpan, err:= strconv.Atoi(string(params[2]))
	if err != nil {
		fmt.Println("参数转换错误")
		return false
	}
	value := string(params[3])
	newLifeSpan := time.Duration(lifeSpan) * time.Second

	fmt.Println("all params is ", key, newLifeSpan, value)
	return commandHandler.cachetable.SAdd(key, newLifeSpan, value)
}

func (commandHandler *CommandHandler) GET(params [][]byte) (bool, string){
	key := string(params[1])
	item, err := commandHandler.cachetable.Get(key)
	if err != nil {
		fmt.Println(err)
		return false, ""
	}
	fmt.Println("get result before: ", item.data)
	mjson, err := json.Marshal(item.data)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println("get result: ", mjson)
	return true, string(mjson)
}
