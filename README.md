## go-socket
go-socket是一个Go(golang)编写的socket服务器、客户端开发框架包，使用非常简单，里面的两个示例代码验证通过

## 使用方式
最简单的使用方式，

#### 服务端示例

```golang
package main

import (
	"github.com/zj2010/go-socket/socket_v1"
	"time"
)

var (
	Server *socket_v1.Server
)

func main() {
	// 1、创建服务器
	Server = socket_v1.NewServer("0.0.0.0", 3333, func(Event socket_v1.HookEvent) {
		onHookEvent(Event)
	})

	// 演示用: 循环发消息
	go goTestSendMsg()

	// 处理其他逻辑
	select {}
}

// 处理数据,多线程转单线程处理
func onHookEvent(Event socket_v1.HookEvent) {
	// 事件处理在此处 ///////////////////////////////////////////////////////////////
	switch Event.EventType {
	case "message": // 1、消息事件
	case "offline": // 2、下线事件
	case "online": // 3、上线消息
	}
	// ////////////////////////////////////////////////////////////////////////////
}

// 发送数据给所有客户端
func goTestSendMsg() {
	for {
		_ = Server.SendMsg(nil, socket_v1.DataUnitSocket{
			Zlib:    1,
			CType:   1000,
			Content: []byte("hello"),
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
	"github.com/zj2010/go-socket/socket_v1"
	"time"
)

func main() {
	// 创建客户端连接
	C := socket_v1.NewClient("127.0.0.1", 3333, func(Msg socket_v1.DataUnitSocket, C *socket_v1.Client) {

		// 回调1：收到了消息，这里处理消息 ///////////////////////////////////////////
		onMsg(Msg)
		// ///////////////////////////////////////////

	}, func(C *socket_v1.Client) {

		// 回调2：连接失败回调

		fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "连接失败！5秒后重连")
		go func() {
			time.Sleep(time.Second * 5)
			C.Connect()
		}()

	}, func(C *socket_v1.Client) {

		// 回调3：连接成功回调
		fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "连接成功！")

	}, func(C *socket_v1.Client) {

		// 回调4：掉线回调
		fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "掉线了,5秒后重连")

		go func() {
			time.Sleep(time.Second * 5)
			C.Connect()
		}()

	})
	go C.Connect()

	// 处理其他逻辑
	select {}
}

// 收到日志消息
func onMsg(Msg socket_v1.DataUnitSocket) {
	fmt.Println(Msg.CType, string(Msg.Content))
}
```


