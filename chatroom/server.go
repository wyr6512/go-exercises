package main

import (
	"fmt"
	"net"
	"strings"
)

type msgInfo struct {
	msg  string   //消息
	conn net.Conn //当前连接
}

var clients = make(map[string]net.Conn)     //客户端列表
var messageQueue = make(chan msgInfo, 1024) //消息队列

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:8088")
	checkErr(err)
	defer listener.Close()

	fmt.Println("server is waiting")

	go processMsg() //消息处理

	for {
		conn, err := listener.Accept()
		checkErr(err)
		defer conn.Close()

		addr := conn.RemoteAddr().String() //客户端地址
		clients[addr] = conn

		go storeInfo(conn) //消息转存
	}
}

//消息转存
func storeInfo(conn net.Conn) {
	buf := make([]byte, 1024)
	defer func(conn net.Conn) {
		addr := fmt.Sprintf("%s", conn.RemoteAddr()) //客户端地址
		delete(clients, addr)                        //连接关闭时删除
		conn.Close()
		for client, con := range clients {
			_, err := con.Write([]byte(addr + " quit"))
			if err != nil {
				fmt.Printf("send to client %s failed\n", client)
			}
		}
	}(conn)

	for {
		length, err := conn.Read(buf)
		if err != nil {
			break
		}
		if length > 0 {
			messageQueue <- msgInfo{string(buf[:length]), conn} //转存到消息队列
		}
	}
}

//处理信息
func processMsg() {
	for {
		select {
		case msg := <-messageQueue:
			doProcessMsg(msg)
		}
	}
}

//解析信息并转发到对应客户端
func doProcessMsg(info msgInfo) {
	contents := strings.Split(info.msg, "#")
	if len(contents) > 1 { //消息解析
		addr := strings.Trim(contents[0], " ")
		message := strings.Join(contents[1:], "#")
		if conn, ok := clients[addr]; ok {
			_, err := conn.Write([]byte(message))
			if err != nil {
				fmt.Printf("send to client %s failed\n", addr)
			}
		}
	} else if strings.ToLower(info.msg) == "list" { //客户端列表
		var list []string
		for client, _ := range clients {
			list = append(list, client)
		}
		_, err := info.conn.Write([]byte(strings.Join(list, "\n")))
		if err != nil {
			fmt.Printf("send to client %s failed\n", info.conn.RemoteAddr().String())
		}
	}
}
