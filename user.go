package main

import (
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

// 创建一个用户的API
func NewUser(conn net.Conn,server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,

		server: server,
	}

	// 启动监听当前user channel 消息的go routine
	go user.ListenMessage()

	return user

}

// 监听当前User channel 的方法，一旦有消息，就直接发送给对端客户端
func (this *User) ListenMessage() {
	for true {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))
	}
}

// 用户上线的功能
func (this *User) Online() {

	// 用户上线 将用户加入到Server的 map中
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this

	this.server.mapLock.Unlock()
	//并 广播当前用户的上线消息
	this.server.BroadCast(this, "is Online")

}

// 用户下线的功能
func (this *User) Offline() {

	// 用户下线 将用户从Server的 map中删除
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap,this.Name)

	this.server.mapLock.Unlock()
	//并 广播当前用户的下线消息
	this.server.BroadCast(this, "is Offline")

}

// 给当前的User对应的客户端发送消息
func (this *User) SendMsg(msg string)  {
	this.conn.Write([]byte(msg))

}


// 用户处理消息的功能
func (this *User) DoMessage(msg string) {
	// 判断是否是查找命令
	if msg=="who" {
		// 查询当前在线用户都有哪些
		this.server.mapLock.Lock()
		for _,user:=range this.server.OnlineMap{
			onLineMsg :="["+user.Addr+"]"+user.Name+":"+"Online...\n"
			this.SendMsg(onLineMsg)
		}
		this.server.mapLock.Unlock()
	}else if len(msg)>7 && msg[:7]== "rename|" {
		// 消息格式：rename|张三
		newName := strings.Split(msg,"|")[1]

		// 判断 name 是否存在
		_,ok:= this.server.OnlineMap[newName]
		if ok {
			this.SendMsg("the Name has been used.\n")
		}else {
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap,this.Name)
			this.server.OnlineMap[newName] = this
			this.server.mapLock.Unlock()

			this.Name= newName
			this.SendMsg("Update the name:"+this.Name+ "\n")

		}

	}else if len(msg)>4 &&msg[:3]=="to|" {
		// 私聊功能：向指定用户发送消息 to|user|msg content

		//1. 获取对方的用户名
		remoteName := strings.Split(msg,"|")[1]
		if remoteName==""{
			this.SendMsg("message format has error,please use \"to|user|msg content\" format!\n")
			return
		}

		//2. 根据用户名 得到对方User对象
		remoteUser,ok:= this.server.OnlineMap[remoteName]
		if !ok {
			this.SendMsg("user is not exist!\n")
			return
		}

		//3. 获取消息内容，通过对方的user对象将消息内容发送过去
		content := strings.Split(msg,"|")[2]
		if content=="" {
			this.SendMsg("there is no message content!\n")
			return
		}
		remoteUser.SendMsg(this.Name+" say: " +content)



	}else {
		// 当前用户 直接广播消息
		this.server.BroadCast(this,msg)
	}

}
