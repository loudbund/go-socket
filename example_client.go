package main

import (
	"fmt"
	"github.com/zhoujianstudio/go-socket/socket_v1"
	"time"
)

// 1.1、收到了消息回调函数，这里处理消息
func OnMessage(Msg socket_v1.UDataSocket, C *socket_v1.Client) {
	onMsg(Msg)
}

// 1.2、连接失败回调函数
func OnConnectFail(C *socket_v1.Client) {
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "连接失败！5秒后重连")
	go C.ReConnect(5) // 延时5秒后重连
}

// 1.3、连接成功回调函数
func OnConnect(C *socket_v1.Client) {
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "连接成功！")
}

// 1.4、掉线回调函数
func OnDisConnect(C *socket_v1.Client) {
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "掉线了,5秒后重连")
	go C.ReConnect(5) // 延时5秒后重连
}

// 2、消息处理
func onMsg(Msg socket_v1.UDataSocket) {
	fmt.Println(Msg.CType, string(Msg.Content))
}

// 6、主函数 -------------------------------------------------------------------------
func main() {
	serverIp := "127.0.0.1"
	serverPort := 2222

	// 创建客户端连接
	Client := socket_v1.NewClient(serverIp, serverPort, OnMessage, OnConnectFail, OnConnect, OnDisConnect)
	go Client.Connect()

	// 处理其他逻辑
	select {}
}
