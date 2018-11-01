package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:8088")
	checkErr(err)
	defer conn.Close()

	go msgSend(conn) //发送消息

	buf := make([]byte, 1024)
	for {
		length, err := conn.Read(buf) //接收服务器信息
		if err != nil {
			fmt.Println("you disconnected server")
			os.Exit(0)
		}
		if length > 0 {
			fmt.Printf("%s\n", buf[:length])
		}
	}
}

//发送消息
func msgSend(conn net.Conn) {
	for {
		reader := bufio.NewReader(os.Stdin)
		data, _, _ := reader.ReadLine() //从标准输入读取信息
		//data like : 127.0.0.1:15716#你好
		input := string(data)
		if strings.ToLower(input) == "quit" {
			conn.Close()
			break
		}
		_, err := conn.Write(data) //发送给服务器
		if err != nil {
			conn.Close()
			fmt.Printf("connection error: %s", err.Error())
			break
		}
	}
}
