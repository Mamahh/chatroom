package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

//Client 定义结构体
type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	flag       int

	conn net.Conn
}

//NewClient 创建Client对象
func NewClient(serverip string, serverport int) *Client {
	client := &Client{
		ServerIp:   serverip,
		ServerPort: serverport,
		flag:       999,
	}

	//连接服务端
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverip, serverport))
	if err != nil {
		fmt.Println("net.Dial err:", err)
		return nil
	}

	client.conn = conn

	return client
}

//menu 菜单显示
func (client *Client) menu() bool {
	var flag int

	fmt.Println("1. 公聊模式")
	fmt.Println("2. 私聊模式")
	fmt.Println("3. 更新用户名")
	fmt.Println("0. 退出")

	fmt.Scanln(&flag)

	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println("模式选择出错！！")
		return false
	}

}

// Dealrespon 监听服务端回应消息的goroutine
func (client *Client) Dealrespon() {
	//一旦client.conn有数据，就会拷贝到标准输出
	io.Copy(os.Stdout, client.conn)
}

//PublicChat 公聊模式
func (client *Client) PublicChat() {
	var Chatmsg string
	fmt.Println("【公】>>>>>>>请输入聊天内容,exit退出")
	fmt.Scanln(&Chatmsg)

	for Chatmsg != "exit" {
		if len(Chatmsg) != 0 {
			SendMsg := Chatmsg + "\n"
			_, err := client.conn.Write([]byte(SendMsg))
			if err != nil {
				fmt.Println("conn.Write err:", err)
				break
			}
		}

		Chatmsg = ""
		fmt.Println("【公】>>>>>>>请输入聊天内容,exit退出")
		fmt.Scanln(&Chatmsg)
	}
}

//SelectUser 查询在线用户
func (client *Client) SelectUser() {
	SendMsg := "who\n"
	client.conn.Write([]byte(SendMsg))
}

//PrivateChat 私聊模式
func (client *Client) PrivateChat() {
	var RemoteUser string
	var Chatmsg string

	//显示在线用户列表
	client.SelectUser()
	fmt.Println(">>>>>>请输入要私聊的用户名：,exit退出")
	fmt.Scanln(&RemoteUser)

	for RemoteUser != "exit" {
		if len(RemoteUser) != 0 {
			fmt.Println("【私】>>>>>>>请输入聊天内容,exit退出")
			fmt.Scanln(&Chatmsg)

			for Chatmsg != "exit" {
				if len(Chatmsg) != 0 {
					SendMsg := Chatmsg + "\n"
					_, err := client.conn.Write([]byte("to|" + RemoteUser + "|" + SendMsg))
					if err != nil {
						fmt.Println("conn.Write err:", err)
						break
					}
				}

				Chatmsg = ""
				fmt.Println("【私】>>>>>>>请输入聊天内容,exit退出")
				fmt.Scanln(&Chatmsg)
			}

		}
		//显示在线用户列表
		client.SelectUser()
		fmt.Println(">>>>>>请输入要私聊的用户名：,exit退出")
		fmt.Scanln(&RemoteUser)
	}

}

//UpdateName 更新用户名
func (client *Client) UpdateName() bool {
	fmt.Println("请输入新的用户名：")
	fmt.Scanln(&client.Name)

	SendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(SendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}
	return true
}

//Run 运行的主函数
func (client *Client) Run() {
	for client.flag != 0 {
		for client.menu() != true {
		}
		switch client.flag {
		case 1:
			//公聊模式
			//fmt.Println("进入公聊模式...")
			client.PublicChat()
			break
		case 2:
			//私聊模式
			fmt.Println("进入私聊模式...")
			client.PrivateChat()
			break
		case 3:
			//更新用户名
			//fmt.Println("进行更改用户名...")
			client.UpdateName()
			break
		}

	}
}

var serverip string
var serverport int

func init() {
	flag.StringVar(&serverip, "ip", "127.0.0.1", "IP：（默认127.0.0.1）")
	flag.IntVar(&serverport, "port", 8888, "端口号：（默认8888）")
}

func main() {
	//参数解析函数
	flag.Parse()
	client := NewClient(serverip, serverport)

	if client == nil {
		fmt.Println(">>>>>>>>连接服务器失败<<<<<<<<<")
		return
	}

	fmt.Println(">>>>>>>>连接服务器成功<<<<<<<<<")

	//监听服务端的消息
	go client.Dealrespon()

	//启动业务流程
	client.Run()

}
