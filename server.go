package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	// 统计 在线用户的列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	// 消息广播的channel
	Message chan string
}

// 创建一个server的接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// 启动服务器的接口
func (this *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}
	// close listen socket
	defer listener.Close()

	// 启动监听Message的goroutine
	go this.ListenMessage()

	for {
		// accept
		// 用户上线
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}

		// do handler
		go this.Handler(conn)
	}

}

// 执行业务的 Handler
func (this *Server) Handler(conn net.Conn) {
	// 处理逻辑
	//fmt.Println("链接建立成功")

	user := NewUser(conn, this)
	user.Online()

	// 增加监听用户是否活跃的channel
	isLive := make(chan bool)

	// 接收客户端发送的消息
	go func() {
		buf := make([]byte, 4096) // 开辟一个4KB的切片buff 存储消息
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}
			// 提取用户发来的消息，并去除 \n
			msg := string(buf[:n-1])

			// 用户针对msg进行消息处理
			user.DoMessage(msg)

			// 用户发出任意动作都表示用户是活跃的
			isLive <- true
		}
	}()

	// 当前handler 阻塞
	// 增加定时器处理 实现超时强踢的功能
	for {
		select {
		case <-isLive:
			// 表示当前用户是活跃的， 需要重置定时器
			// 不做任何操作，直接进行定时器判断并重置定时器 进入到第二个case中
		// 读取定时器 channel
		case <-time.After(time.Second * 300):
			// 已经超时
			// 执行将当前的User强制关闭
			user.SendMsg("You're out!")
			close(user.C)
			// 关闭连接
			conn.Close()

			// 退出 handler
			return

		}

	}

}

// 广播消息的方法
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}

// 监听 Message 广播消息的 goroutine ，一旦有消息就发送给全部在线的User
func (this *Server) ListenMessage() {
	for {
		msg := <-this.Message
		// 将消息发送给全部在线的User
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}

		this.mapLock.Unlock()
	}

}

func main()  {
	server := NewServer("127.0.0.1", 8888)
	server.Start()
}