package memredis

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"testing"
	"time"
)

type CR struct {
	command string
	response string
}

var set_expire_commands = []CR{
	CR{"SET name jack 10", "true"},
	CR{"GET name", "true jack"},
	// sleep 10s
	CR{"GET name", "false Key not found in cache"},
	CR{"SET name tony", "true"},
	CR{"GET name", "true tony"},
}

var setnx_commands = []CR{
	CR{"SETNX name1 jack", "true"},
	CR{"GET name1", "true jack"},
	CR{"SETNX name1 tony", "true"},
	CR{"GET name1", "true jack"},
}

var sadd_smembers_commands = []CR{
	CR{"SADD names jack", "true"},
	CR{"SMEMBERS names", "true [\"jack\"]"},
	CR{"SADD names tony", "true"},
	CR{"SMEMBERS names","true [\"jack\",\"tony\"]"},

}


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

func commonTest(t *testing.T, crList []CR) {
	addr, _ := net.ResolveTCPAddr("tcp4", ":3009")
	conn, _ := net.DialTCP("tcp4", nil, addr)
	for i:= 0; i < len(crList); i += 1 {
		writeData(conn, []byte(crList[i].command))
		buf := readData(conn)
		fmt.Println(string(buf))
		if string(buf) != crList[i].response {
			t.Error(i, " case fail")
		}
	}
}

func TestProtocolHandler_SADD_COMMON(t *testing.T) {
	commonTest(t, sadd_smembers_commands)
}

func TestProtocolHandler_SETNX_COMMON(t *testing.T) {
	commonTest(t, setnx_commands)
}


func TestProtocolHandler_SET_EXPIRE(t *testing.T) {
	addr, _ := net.ResolveTCPAddr("tcp4", ":3009")
	conn, _ := net.DialTCP("tcp4", nil, addr)
	for i:= 0; i < 2; i += 1 {
		writeData(conn, []byte(set_expire_commands[i].command))
		buf := readData(conn)
		fmt.Println(string(buf))
		if string(buf) != set_expire_commands[i].response {
			t.Error(i, " case fail")
		}
	}
	time.Sleep(10*time.Second)
	for i := 2; i < len(set_expire_commands); i += 1 {
		writeData(conn, []byte(set_expire_commands[i].command))
		buf := readData(conn)
		fmt.Println(string(buf))
		if string(buf) != set_expire_commands[i].response {
			t.Error(i, " case fail")
		}
	}
}