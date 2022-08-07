package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
)

const (
	PackageLengthBytes = 4
	HeaderLengthBytes = 2
	ProtocolVersionBytes = 2
	OperationBytes = 4
	SequenceIDBytes = 4

	HeaderLength = PackageLengthBytes + HeaderLengthBytes + ProtocolVersionBytes + OperationBytes + SequenceIDBytes
)

//解包
func Depack(buffer []byte) []byte {
	length := len(buffer)

	var i int
	data := make([]byte, 32)

	for i = 0; i < length; i++ {
		if length < i + HeaderLength {
			break
		}

		messageLength := ByteToInt(buffer[i: i + PackageLengthBytes])
		if length < i + HeaderLength + messageLength {
			break
		}

		site := i + PackageLengthBytes
		headerLength := ByteToInt(buffer[site : site + HeaderLengthBytes])
		site += HeaderLengthBytes

		protocolVersion := ByteToInt16(buffer[site : site + ProtocolVersionBytes])
		site += ProtocolVersionBytes

		operation := ByteToInt(buffer[site : site + OperationBytes])
		site += OperationBytes

		sequenceID := ByteToInt(buffer[site : site + SequenceIDBytes])
		site += SequenceIDBytes

		fmt.Printf("packageLength: %d, headerLength: %d , protocolVersion: %d, operation: %d, sequenceID: %d \n", messageLength, headerLength, protocolVersion, operation, sequenceID)

		data = buffer[i + HeaderLength : i + HeaderLength + messageLength]
		break
	}

	if i == length {
		return make([]byte, 0)
	}

	return data
}

//byte数组转成32位整数
func ByteToInt(n []byte) int {
	bytesbuffer := bytes.NewBuffer(n)
	var x int32
	binary.Read(bytesbuffer, binary.BigEndian, &x)

	return int(x)
}

//转成16位整数
func ByteToInt16(n []byte) int {
	bytesbuffer := bytes.NewBuffer(n)
	var x int16
	binary.Read(bytesbuffer, binary.BigEndian, &x)

	return int(x)
}

//连接
func handleConnection(conn net.Conn) {
	tmpBuffer := make([]byte, 0)
	readerChannel := make(chan []byte, 10000)
	go reader(readerChannel)

	buffer := make([]byte, 1024)
	for{
		n, err := conn.Read(buffer)
		if err != nil {
			Log(conn.RemoteAddr().String(), "connection error: ", err)
			return
		}

		tmpBuffer = Depack(append(tmpBuffer, buffer[:n]...))
		readerChannel <- tmpBuffer  
	}
	defer conn.Close()
}

//获取数据
func reader(readerchannel chan []byte) {
	for{
		select {
		case data := <-readerchannel:
			Log(string(data))      //打印
		}
	}
}

//日志
func Log(v ...interface{}) {
	log.Println(v...)
}

//错误
func CheckErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func main() {
	netListen, err := net.Listen("tcp", "localhost:8080")
	CheckErr(err)
	defer netListen.Close()

	Log("Waiting for client ...")     
	for{
		conn, err := netListen.Accept()    
		if err != nil {
			Log(conn.RemoteAddr().String(), "Error", err)
			continue
		}
		Log(conn.RemoteAddr().String(), "tcp connection success")
		go handleConnection(conn)
	}
}