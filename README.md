## go-socket
go-socket是一个Go(golang)编写的socket服务器、客户端开发框架包，使用非常简单，里面的两个示例代码验证通过

## 安装
go get github.com/loudbund/go-socket

## 引入
```golang
import "github.com/loudbund/go-socket/socket_v1"
```
## 消息协议格式
1. socket_v1.UDataSocket 这个是应用程序方使用的数据协议结构体，收到消息和发送消息都用这个结构
2. Ctype为消息类型：建议应用从100开始使用；已被使用的有1,7,8，分别是 1：心跳，7：客户端连上后自动给服务器发条消息 8：服务器回复客户端的7类型消息
3. Content为传输内容：较大的传输时， Zlib可以开启，这样可以减少传输字节，当然也会消耗cpu性能

```golang
// 结构体1：(外部用)传输数据上层结构体
type UDataSocket struct {
	Zlib    int    // 是否压缩 1:压缩
	CType   int    // 内容类型 1:客户端请求消息 2:服务端表接口消息 4:服务端表内容数据 200:服务端发送结束
	Content []byte // 发送内容
}
```
## 传输协议
数据传输也是按这个顺序组装的
1. 先读取前面20个字节，再根据ContentTranLength的值读取ContentTran的内容
2. SendFlag将会被校验，不匹配则中止接收抛出异常
3. Zlib开启时，ContentTran为unitDataSend.ContentTran的压缩内容，程序收到后要解压
4. ContentLength未为压缩内容长度，ContentTran解开后的内容长度如果不一致，则抛出异常
```golang
// 结构体2：(内部用)传输数据底层结构体
type unitDataSend struct {
	SendFlag          int    // 消息最前面标记
	Zlib              int    // 压缩标记 (同 UDataSocket.Zlib)
	CType             int    // 内容类型 (同 UDataSocket.CType)
	ContentLength     int    // 原内容大小
	ContentTranLength int    // 发送内容大小
	ContentTran       []byte // 发送的内容 (同 UDataSocket.Content)
}
```
## 传输协议 SendFlag设置
```golang
socket_v1.SetSendFlag(xxxxx)
```

## 服务端示例

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
    socket_v1.SetSendFlag(123456)
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

## 客户端示例
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
    socket_v1.SetSendFlag(123456)
	Client := socket_v1.NewClient(serverIp, serverPort, OnMessage, OnConnectFail, OnConnect, OnDisConnect)
	go Client.Connect()

	// 处理其他逻辑
	select {}
}


```


