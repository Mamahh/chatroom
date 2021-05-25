package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

//Server 创建结构体
type Server struct {
	Ip        string
	Port      int
	OnlineMap map[string]*User
	mapLock   sync.RWMutex
	Message   chan string
}

//NewServer new一个对象
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

//ListenMessager 监听Message广播消息channel的goroutine,一旦有消息就发送给全部的在线user
func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

//BroadCast 广播用户上线
func (this *Server) BroadCast(user *User, msg string) {
	Sendmsg := "[" + user.Addr + "]" + user.Name + ":" + msg

	this.Message <- Sendmsg
}

//Hander 业务处理函数
func (this *Server) Hander(conn net.Conn) {
	//当前连接的业务
	//fmt.Println("连接建立成功")

	user := NewUser(this, conn)

	//上线
	user.Online()

	//添加活跃标志位
	isLive := make(chan bool)

	//添加群聊功能
	go func() {
		buf := make([]byte, 4096)

		for {
			n, err := conn.Read(buf)
			if n == 0 {
				//下线
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("send msg err:", err)
				return
			}

			//截取用户消息发送的"\n"
			msg := string(buf[:n-1])
			user.DoMessage(msg)

			//发消息时往channel发true，维持活跃度
			isLive <- true
		}
	}()

	//当前handler阻塞
	for {
		select {
		case <-isLive:
			{
				//仅为了刷新select，重置超时定时器作用
			}
		case <-time.After(time.Second * 300):
			{
				//强制下线
				user.SendMsg("连接超时，您已被强制下线" + "\n")

				//销毁资源，关闭channel
				close(user.C)

				//关闭conn连接
				conn.Close()

				//退出当前handle
				return
			}
		}
	}
}

//Start 启动服务器的接口
func (this *Server) Start() {
	//socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}

	//启动监听message的goroutine
	go this.ListenMessager()

	//close listener socket
	defer listener.Close()

	for {
		//accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("accept connect err:", err)
			continue
		}

		go this.Hander(conn)
	}

}
