package memredis

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"testing"
	"time"
)

func readData(conn *net.TCPConn) []byte {
	var resSize [4]byte
	conn.Read(resSize[:])
	responseDataSize := int32(binary.BigEndian.Uint32(resSize[:]))
	buf := make([]byte, responseDataSize)
	conn.Read(buf[:])
	return buf
}

func writeData(conn *net.TCPConn, data []byte) {
	size := make([]byte, 4)
	binary.BigEndian.PutUint32(size, uint32(len(data)))

	var buffer bytes.Buffer
	buffer.Write(size)
	buffer.Write(data)
	conn.Write(buffer.Bytes())
}

func TestProtocolHandler_SET(t *testing.T) {

	addr, _ := net.ResolveTCPAddr("tcp4", ":3009")
	conn, _ := net.DialTCP("tcp4", nil, addr)
	fmt.Println(conn)

	command := []byte("SET name jack 10")
	writeData(conn, command)

	//res, err := ioutil.ReadAll(conn)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Println(res)

	buf := readData(conn)
	fmt.Println(string(buf))
	if string(buf) != "true" {
		t.Error("SET error")
	}

	command = []byte("GET name")
	writeData(conn, command)
	buf = readData(conn)
	fmt.Println(string(buf))
	ok := bytes.Split(buf, separatorBytes)[0]
	if string(ok) != "true" {
		t.Error("GET error")
	} else {
		result := string(bytes.Split(buf, separatorBytes)[1])
		if result != "jack" {
			t.Error("GET value error")
		}
	}

	time.Sleep(10 * time.Second)

	command = []byte("GET name")
	writeData(conn, command)
	buf = readData(conn)
	fmt.Println("buf:", string(buf))
	ok = bytes.Split(buf, separatorBytes)[0]
	if string(ok) != "false" {
		t.Error("GET again error")
	}
}

func TestProtocolHandler_SADD(t *testing.T) {
	addr, _ := net.ResolveTCPAddr("tcp4", ":3009")
	conn, _ := net.DialTCP("tcp4", nil, addr)
	fmt.Println(conn)

	command := []byte("SADD names jack")
	writeData(conn, command)
	ok := readData(conn)
	if string(ok) != "true" {
		t.Error("SADD error")
	}

	command = []byte("SMEMBERS names")
	writeData(conn, command)
	buf := readData(conn)
	fmt.Println(string(buf))
	ok = bytes.Split(buf, separatorBytes)[0]
	if string(ok) != "true" {
		t.Error("SMEMBERS error")
	} else {
		result := bytes.Split(buf, separatorBytes)[1]
		fmt.Println(string(result))
		var resultList []string
		json.Unmarshal(result, &resultList)
		fmt.Println(resultList)
		if resultList[0] != "jack" {
			t.Error("SMEMBERS value ERROR")
		}
	}

	command = []byte("SADD names tony")
	writeData(conn, command)
	ok = readData(conn)
	if string(ok) != "true" {
		t.Error("SADD error")
	}

	command = []byte("SMEMBERS names")
	writeData(conn, command)
	buf = readData(conn)
	fmt.Println(string(buf))
	ok = bytes.Split(buf, separatorBytes)[0]
	if string(ok) != "true" {
		t.Error("SMEMBERS error")
	} else {
		result := bytes.Split(buf, separatorBytes)[1]
		fmt.Println(string(result))
		var resultList []string
		json.Unmarshal(result, &resultList)
		fmt.Println(resultList)
		if resultList[0] != "jack" || resultList[1] != "tony" {
			t.Error("SMEMBERS value ERROR")
		}
	}

}
