## go-socket
go-socket是一个Go(golang)编写的socket服务器、客户端开发框架包，使用非常简单，里面的两个示例代码验证通过

## 使用方式
最简单的使用方式，

#### 服务端示例

```golang
package main

import (
	"fmt"
	"github.com/loudbund/go-socket/socket_v1"
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

```

#### 客户端示例
```golang
package main

import (
	"fmt"
	"github.com/loudbund/go-socket/socket_v1"
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


```


