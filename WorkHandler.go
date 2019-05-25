package memredis

import (
	"fmt"
	"syscall"
	"strconv"
)


func echo(fd int) {
	defer syscall.Close(fd)
	var buf [32 * 1024]byte
	for {
		//var buffer bytes.Buffer
		nbytes, e := syscall.Read(fd, buf[:])
		if nbytes > 0 {
			fmt.Printf(">>> %s\n", buf[:nbytes])
			s := "<html><head><title>Example></title></head><body><p>" + string(buf[:nbytes]) + " hello </p></body></html>"
			//buffer.Write(buf[:nbytes])
			//buffer.Write([]byte(" hello"))
			//syscall.Write(fd, buffer.Bytes())
			output := "HTTP/1.1 200 OK\r\n" + "Content-Type: text/html\r\n" + "Content-Length: " + strconv.Itoa(len(s)) + "\r\n\r\n" + s
			syscall.Write(fd, []byte(output))

			fmt.Printf("<<< %s\n", output)
		}
		if e != nil {
			break
		}
	}
}

type WorkerHandlerImpl struct {}

func (wh *WorkerHandlerImpl) handle(connFd int) {
	fmt.Println("this is a worker handler, fd is ", connFd)
	defer syscall.Close(connFd)
	var buf [32 * 1024]byte
	nbytes, e := syscall.Read(connFd, buf[:])
	if nbytes > 0 {
		fmt.Printf(">>> %s\n", buf[:nbytes])
		s := "<html><head><title>Example></title></head><body><p>" + string(buf[:nbytes]) + " hello </p></body></html>"
		//buffer.Write(buf[:nbytes])
		//buffer.Write([]byte(" hello"))
		//syscall.Write(fd, buffer.Bytes())
		output := "HTTP/1.1 200 OK\r\n" + "Content-Type: text/html\r\n" + "Content-Length: " + strconv.Itoa(len(s)) + "\r\n\r\n" + s
		syscall.Write(connFd, []byte(output))

		fmt.Printf("<<< %s\n", output)
	}
	if e != nil {
		fmt.Println("read error")
	}
}
