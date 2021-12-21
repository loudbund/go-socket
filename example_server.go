package main

import (
	"fmt"
	"github.com/zhoujianstudio/go-socket/socket_v1"
	"time"
)

var (
	Server *socket_v1.Server
)

// 1、处理数据,多线程转单线程处理
func onHookEvent(Event socket_v1.HookEvent) {
	// 事件处理在此处 ///////////////////////////////////////////////////////////////
	switch Event.EventType {
	case "message": // 1、消息事件
		fmt.Println("message:", string(Event.Message.Content))
	case "offline": // 2、下线事件
		fmt.Println("message:", string(Event.Message.Content))
	case "online": // 3、上线消息
		fmt.Println("message:", string(Event.Message.Content))
	}
	// ////////////////////////////////////////////////////////////////////////////
}

func main() {
	port := 2222

	// 1、创建服务器
	Server = socket_v1.NewServer("0.0.0.0", port, func(Event socket_v1.HookEvent) {
		onHookEvent(Event)
	})

	// 演示用: 循环发消息
	for {
		_ = Server.SendMsg(nil, socket_v1.UDataSocket{
			Zlib:    1,               // 是否压缩传输 1:压缩 0:不压缩
			CType:   1000,            // 指令编号
			Content: []byte("hello"), // 指令内容
		})
		time.Sleep(time.Second)
	}
}
