package memredis

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"syscall"
	"time"
)
var separatorBytes = []byte(" ")

type ProtocolHandler struct {
	peerPicker PeerPicker
	cacheTable *CacheTable
}

func NewProtocolHandler(peerPicker PeerPicker, table *CacheTable) *ProtocolHandler {
	return &ProtocolHandler{
		peerPicker: peerPicker,
		cacheTable: table,
	}
}


func (protocolHandler *ProtocolHandler) handle(connFd int) {
	data := protocolHandler.readData(connFd)
	if data != nil {
		params := bytes.Split(data, separatorBytes)
		if params != nil {
			key := string(params[1])
			peer, ok := protocolHandler.peerPicker.PickPeer(key)
			if !ok {
				fmt.Println("cannot find key in other peer, will exec command in this peer")
				ok, result := protocolHandler.Exec(params)
				data := []byte(strconv.FormatBool(ok) + " " + result)
				protocolHandler.writeData(connFd, data)
			} else {
				fmt.Println("find peer", peer)
				protocolHandler.transmit(connFd, peer, data)
			}
			protocolHandler.check()
		}
	}
}

func (protocolHandler *ProtocolHandler)check() {
	for k, v := range protocolHandler.cacheTable.items {
		fmt.Println("cache ", k, v.data)
	}
}

func ipToAddrByte(ip string) [4]byte {
	bits := strings.Split(ip, ".")
	var ipBytes [4]byte
	for  i := 0; i < 4; i ++ {
		field, _ := strconv.Atoi(bits[i])
		ipBytes[i] = uint8(field)
	}
	fmt.Println("peer ip is:", ipBytes)
	return ipBytes
}

func (protocolHandler *ProtocolHandler) transmit(connFd int, peer string, data []byte) {
	fmt.Println("proxy to other peer")
	addrs := strings.Split(peer, ":")
	port, _ := strconv.Atoi(addrs[1])
	address := syscall.SockaddrInet4{
		Addr: ipToAddrByte(addrs[0]),
		Port: port,

	}
	peerFd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		fmt.Println(err)
		responseData := []byte(strconv.FormatBool(false))
		protocolHandler.writeData(connFd, responseData)
		return
	}
	syscall.Connect(peerFd, &address)
	protocolHandler.writeData(peerFd, data)
	responseData := protocolHandler.readData(peerFd)
	protocolHandler.writeData(connFd, responseData)
}

//	[x][x][x][x][x][x][x][x]
//	|  (int32) ||  (binary)
//	|  4-byte  ||  N-byte
//	---------------------------
//	   size         data
func (protocolHandler *ProtocolHandler) readData(connFd int) (data []byte){
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
			data = buf[:nbytes]
			return data
		}
	}
	return nil
}

//	[x][x][x][x][x][x][x][x]
//	|  (int32) ||  (binary)
//	|  4-byte  ||  N-byte
//	---------------------------
//	   size         data(success, result)
func (ProtocolHandler *ProtocolHandler) writeData(connFd int, data []byte) {
	size := make([]byte, 4)
	binary.BigEndian.PutUint32(size, uint32(len(data)))
	fmt.Println("write size: ", size)

	var buffer bytes.Buffer
	buffer.Write(size)
	buffer.Write(data)
	syscall.Write(connFd, buffer.Bytes())
}


func (protocolHandler *ProtocolHandler) lookupCache(key string) (interface{}, bool){
	value, err := protocolHandler.cacheTable.Get(key)
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
	value := string(params[2])
	newLifeSpan := time.Duration(0)
	fmt.Println("param length:", len(params))
	if len(params) == 4 {
		lifeSpan, err := strconv.Atoi(string(params[3]))
		if err != nil {
			fmt.Println("参数转换错误")
			return false
		}
		newLifeSpan = time.Duration(lifeSpan) * time.Second
	}

	fmt.Println("all params is ", key, newLifeSpan, value)
	return protocolHandler.cacheTable.Set(key, newLifeSpan, value)
}

func (protocolHandler *ProtocolHandler) SADD(params [][]byte) (ok bool){
	key := string(params[1])
	value := string(params[2])
	newLifeSpan := time.Duration(0)
	if len(params) == 4 {
		lifeSpan, err := strconv.Atoi(string(params[3]))
		if err != nil {
			fmt.Println("参数转换错误")
			return false
		}
		newLifeSpan = time.Duration(lifeSpan) * time.Second
	}

	fmt.Println("all params is ", key, newLifeSpan, value)
	return protocolHandler.cacheTable.SAdd(key, newLifeSpan, value)
}

func (protocolHandler *ProtocolHandler) GET(params [][]byte) (bool, string){
	key := string(params[1])
	item, err := protocolHandler.cacheTable.Get(key)
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
