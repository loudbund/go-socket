package main

import (
	"fmt"
	"github.com/zj2010/go-socket/socket_v1"
	"time"
)

// 1、结构体 -------------------------------------------------------------------------

// 2、全局变量 -------------------------------------------------------------------------

// 3、初始化函数 -------------------------------------------------------------------------

// 4、开放的函数 -------------------------------------------------------------------------

// 5、内部函数 -------------------------------------------------------------------------

// 6、主函数 -------------------------------------------------------------------------
func main() {
	serverIp := "127.0.0.1"
	serverPort := 2222

	// 创建客户端连接
	Client := socket_v1.NewClient(serverIp, serverPort, func(Msg socket_v1.UDataSocket, C *socket_v1.Client) {

		// 1、收到了消息，这里处理消息 OnMessage ///////////////////////////////////////////
		onMsg(Msg)
		// ///////////////////////////////////////////

	}, func(C *socket_v1.Client) {

		// 2、连接失败回调 OnConnectFail

		fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "连接失败！5秒后重连")
		go func() {
			time.Sleep(time.Second * 5)
			C.Connect()
		}()

	}, func(C *socket_v1.Client) {

		// 3、连接成功回调 OnConnect
		fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "连接成功！")

	}, func(C *socket_v1.Client) {

		// 4、掉线回调 OnDisConnect
		fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "掉线了,5秒后重连")

		go func() {
			time.Sleep(time.Second * 5)
			C.Connect()
		}()

	})
	go Client.Connect()

	// 处理其他逻辑
	select {}
}

// 收到日志消息
func onMsg(Msg socket_v1.UDataSocket) {
	fmt.Println(Msg.CType, string(Msg.Content))
}
