package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

const (
	PackageLengthBytes = 4
	HeaderLengthBytes = 2
	ProtocolVersionBytes = 2
	OperationBytes = 4
	SequenceIDBytes = 4

	HeaderLength = PackageLengthBytes + HeaderLengthBytes + ProtocolVersionBytes + OperationBytes + SequenceIDBytes
)

//封
func Enpack(message []byte) []byte {
	b := append(Int32ToBytes(len(message)), Int16ToBytes(0)...)
	b = append(b, Int16ToBytes(8)...)
	b = append(b, Int32ToBytes(99)...)
	b = append(b, Int32ToBytes(10)...)
	b = append(b, message...)

	return b
}

//32位整数转成字节
func Int32ToBytes(n int) []byte {
	x := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	_ = binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}

//16位整数转成字节
func Int16ToBytes(n int) []byte {
	x := int16(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	_ = binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}

//发送实验请求
func send(conn net.Conn)  {
	for i := 0; i < 10; i++ {
		session := GetSession()
		words := "{\"ID\":\""+strconv.Itoa(i)+"\",\"Session\":\""+session+"20170914165908\",\"Meta\":\"golang\",\"Content\":\"message\"}"
		_, _ = conn.Write(Enpack([]byte(words)))

		fmt.Println(words)
	}

	fmt.Println("send over")
	defer conn.Close()
}

//用当前时间做标识
func GetSession() string {
	gs1 := time.Now().Unix()
	gs2 := strconv.FormatInt(gs1, 10)
	return gs2
}

func connectAndSend()  {
	server := "localhost:8080"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", server)
	if err != nil{
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil{
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}

	fmt.Println("Connect success")

	send(conn)
}

func main() {
	connectAndSend()
}