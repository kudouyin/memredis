package memredis

import (
	"syscall"
	"bytes"
	"time"
	"strconv"
	"fmt"
	"encoding/json"
	"encoding/binary"
)
var separatorBytes = []byte(" ")

type ProtocolHandler struct {
	peerPicker PeerPicker
	cachetable *CacheTable
}

func NewProtocolHandler(peerPicker PeerPicker, table *CacheTable) *ProtocolHandler {
	return &ProtocolHandler{
		peerPicker: peerPicker,
		cachetable: table,
	}
}


func (protocolHandler *ProtocolHandler) handle(connFd int) {
	params := protocolHandler.readCommand(connFd)
	if params != nil {
		key := string(params[1])
		peer, ok := protocolHandler.peerPicker.PickPeer(key)
		if !ok {
			fmt.Println("cannot find peer")
			ok, result := protocolHandler.Exec(params)
			protocolHandler.writeResult(connFd, ok, result)
		} else {
			fmt.Println("find peer", peer)
			protocolHandler.transmit(peer)
		}
		check()
	}
}

func check() {
	for k, v := range Cachetable.items {
		fmt.Println("cache ", k, v.data)
	}
}

func (protocolHandler *ProtocolHandler) transmit(peer string) {

}

//	[x][x][x][x][x][x][x][x]
//	|  (int32) ||  (binary)
//	|  4-byte  ||  N-byte
//	---------------------------
//	   size         data
func (protocolHandler *ProtocolHandler) readCommand(connFd int) (params [][]byte){
	var lenSlice [4]byte
	_, e := syscall.Read(connFd, lenSlice[:])
	if e != nil {
		fmt.Println(e)
		return nil
	}
	size :=  int32(binary.BigEndian.Uint32(lenSlice[:]))
	fmt.Println("size:", size)
	if size > 0 {
		buf := make([]byte, size)
		nbytes, err := syscall.Read(connFd, buf[:])
		if err != nil {
			return nil
		}
		if nbytes > 0 {
			params := bytes.Split(buf[:nbytes], separatorBytes)
			return params
		}
	}
	return nil
}

func (ProtocolHandler *ProtocolHandler) writeResult(connFd int, ok bool, result string) {

	var buffer bytes.Buffer
	buffer.Write([]byte(strconv.FormatBool(ok)))
	buffer.Write(separatorBytes)
	buffer.Write(([]byte(result)))
	syscall.Write(connFd, buffer.Bytes())
}


func (protocolHandler *ProtocolHandler) lookupCache(key string) (interface{}, bool){
	value, err := protocolHandler.cachetable.Get(key)
	if err != nil {
		return nil, false
	}
	return value, true

}

func (protocolHandler *ProtocolHandler) Exec(params [][]byte) (bool, string){
	switch  {
	case bytes.Equal(params[0], []byte("SET")):
		return protocolHandler.SET(params), ""
	case bytes.Equal(params[0], []byte("SADD")):
		return protocolHandler.SADD(params), ""
	case bytes.Equal(params[0], []byte("GET")):
		return protocolHandler.GET(params)

	}
	return false, ""
}

func (protocolHandler *ProtocolHandler) SET(params [][]byte) (ok bool){
	key := string(params[1])
	lifeSpan, err:= strconv.Atoi(string(params[2]))
	if err != nil {
		fmt.Println("参数转换错误")
		return false
	}
	value := string(params[3])
	newLifeSpan := time.Duration(lifeSpan) * time.Second

	fmt.Println("all params is ", key, newLifeSpan, value)
	return protocolHandler.cachetable.Set(key, newLifeSpan, value)
}

func (protocolHandler *ProtocolHandler) SADD(params [][]byte) (ok bool){
	key := string(params[1])
	lifeSpan, err:= strconv.Atoi(string(params[2]))
	if err != nil {
		fmt.Println("参数转换错误")
		return false
	}
	value := string(params[3])
	newLifeSpan := time.Duration(lifeSpan) * time.Second

	fmt.Println("all params is ", key, newLifeSpan, value)
	return protocolHandler.cachetable.SAdd(key, newLifeSpan, value)
}

func (protocolHandler *ProtocolHandler) GET(params [][]byte) (bool, string){
	key := string(params[1])
	item, err := protocolHandler.cachetable.Get(key)
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
