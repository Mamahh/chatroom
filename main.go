package main

func main() {
	//设置默认Ip,Port
	server := NewServer("127.0.0.1", 8888)
	server.Start()
}
