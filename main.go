package main

import (
	"flag"
	"fmt"
)

func server()  {
	server := NewServer("127.0.0.1", 8888)
	server.Start()
}

func client() {

	// 客户端建立测试

	// 使用命令行解析的方式

	flag.Parse()

	client:=NewClient("127.0.0.1", 8888)
	if client == nil {
		fmt.Println(">>>>> client connection failed！")
		return
	}
	fmt.Println(">>>>> client connection successful!")

	// 启动 客户端 的业务
	select {

	}

}
