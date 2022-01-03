package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int
}

func NewClient(serverIp string, serverPort int) *Client {
	// 创建客户端对象
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}

	// 连接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial error: ", err)
		return nil
	}
	client.conn = conn

	// 返回对象
	return client
}

// 通过命令行解析 得到 ip 和 端口号
var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器ip地址(默认127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口(默认8888)")
}

// client 菜单显示
func (client *Client) menu() bool {
	var flag int

	fmt.Println("1.广播模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")

	fmt.Scanln(&flag)
	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println(">>>>> please input legal number！")
		return false
	}

}

func (client *Client) Run() {
	for client.flag != 0 {
		for client.menu() != true {

		}
		// 根据不同的模式处理不同的业务
		switch client.flag {
		case 1:
			// 广播模式
			client.BroadCastMsg()
			break
		case 2:
			// 私聊模式
			client.PrivateChat()
			break
		case 3:
			// 更新用户名
			client.UpdateUserName()
			break

		}

	}
}

// 更新用户名
func (client *Client) UpdateUserName() bool {
	fmt.Println(">>> please input user name:")

	fmt.Scanln(&client.Name)
	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Writer err:", err)
		return false
	}
	return true
}

// 广播模式
func (client *Client) BroadCastMsg() {
	// 提示用户输入消息
	var chatMsg string
	fmt.Println("请输入广播消息,exit退出")
	fmt.Scanln(&chatMsg)
	for chatMsg != "exit" {
		// 发送给服务器
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn Writer err: ", err)
				break
			}
		}
		chatMsg = ""
		fmt.Println("input message ,\"exit\" will exit this process!")
		fmt.Scanln(&chatMsg)
	}
}

// 查询当前在线的用户
func (client *Client) SelectUser() {
	sendMsg := "who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn Writer err: ", err)
		return
	}
}

// 私聊模式
func (client *Client) PrivateChat() {
	var remoteUserName string
	var chatMsg string

	fmt.Println("current online users:")
	client.SelectUser()
	fmt.Println("please input userName,\"exit\" will exit this process!")
	fmt.Scanln(&remoteUserName)

	for remoteUserName != "exit" {
		fmt.Println("please input chat message,\"exit\" will exit this process!")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {
			// 发送给服务器
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteUserName + "|" + chatMsg + "\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn Writer err: ", err)
					break
				}
			}
			chatMsg = ""
			fmt.Println("please input chat message,\"exit\" will exit this process!")
			fmt.Scanln(&chatMsg)
		}

		client.SelectUser()
		fmt.Println("please input userName,\"exit\" will exit this process!")
		fmt.Scanln(&remoteUserName)

	}

}

// 处理服务器回应的消息
func (client *Client) DealResponse() {
	// 一旦服务器端回应消息就直接copy到stdout标准输出上，且永久阻塞监听
	io.Copy(os.Stdout, client.conn)

}

func main() {

	// 客户端建立测试

	// 使用命令行解析的方式

	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>> client connection failed！")
		return
	}
	fmt.Println(">>>>> client connection successful!")

	// 单独开启一个go routine 去处理server的回执消息
	go client.DealResponse()

	// 启动 客户端 的业务
	client.Run()

}
