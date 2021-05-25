package main

import (
	"net"
	"strings"
)

//User 定义类
type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

//NewUser API
func NewUser(server *Server, conn net.Conn) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,

		server: server,
	}

	//启动监听当前channel消息的goroutine
	go user.ListenMessage()

	return user
}

//Online 用户上线
func (this *User) Online() {
	//用户上线,将用户添加进OnlineMap中
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	//广播用户上线的消息
	this.server.BroadCast(this, "已上线")
}

//Offline 用户下线
func (this *User) Offline() {
	//用户下线,将用户从OnlineMap移除
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	//广播用户下线的消息
	this.server.BroadCast(this, "已下线")
}

//SendMsg 发送消息
func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

//DoMessage 用户收发数据
func (this *User) DoMessage(msg string) {
	//输入指令 who 可以查询当前在线的用户列表
	if msg == "who" {
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			OnlineMsg := "[" + user.Addr + user.Name + ":" + "在线...\n"
			this.SendMsg(OnlineMsg)
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		Newname := strings.Split(msg, "|")[1]

		_, ok := this.server.OnlineMap[Newname]
		if ok {
			this.SendMsg("该名称已存在，请重新输入\n")
		} else {
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[Newname] = this
			this.server.mapLock.Unlock()

			this.Name = Newname
			this.SendMsg("已更新用户名" + ": " + this.Name + "\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		//判断接收用户名的合法性
		Remote_name := strings.Split(msg, "|")[1]
		if Remote_name == "" {
			this.SendMsg("用户名输入有误，请按`to|name|msg`格式发送" + "\n")
			return
		}

		//判断接收用户在线与否
		Remote_user, ok := this.server.OnlineMap[Remote_name]
		if !ok {
			this.SendMsg("该用户不在线" + "\n")
			return
		}

		//获取实际发送消息
		Real_msg := strings.Split(msg, "|")[2]
		Remote_user.SendMsg(this.Name + "对您说：" + Real_msg + "\n")

	} else {
		this.server.BroadCast(this, msg)
	}
}

//ListenMessage 监听是否有channel发送,有则将消息发往client
func (this *User) ListenMessage() {
	for {
		msg := <-this.C

		this.conn.Write([]byte(msg + "\n"))
	}
}
