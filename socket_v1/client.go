package socket_v1

import (
	"fmt"
	"net"
	"time"
)

// 客户端结构体
type Client struct {
	socketMsg
	ip            string // ip地址
	port          int    // 端口
	heartBeatTime int    // 心跳时间

	conn     net.Conn // 连接句柄
	ClientId string   // 我的客户端id

	OnMessage     func(msg UDataSocket, C *Client) // 消息回调
	OnConnect     func(C *Client)                  // 上线回调
	OnConnectFail func(C *Client)                  // 上线回调
	OnDisConnect  func(C *Client)                  // 掉线回调
}

// 对外函数1：初始化一个客户端
func NewClient(ip string, port int, OnMessage func(msg UDataSocket, C *Client), OnConnectFail, OnConnect, OnDisConnect func(C *Client)) *Client {
	return &Client{
		ip:            ip,
		port:          port,
		heartBeatTime: 5,

		conn:     nil,
		ClientId: "",

		OnMessage:     OnMessage,
		OnConnect:     OnConnect,
		OnConnectFail: OnConnectFail,
		OnDisConnect:  OnDisConnect,
	}
}

// 对外函数2：连接服务器
func (Me *Client) Connect() {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", Me.ip, Me.port))
	if err != nil {
		fmt.Println("net.Dial error:", err)
		Me.OnConnectFail(Me) // 回调连接失败事件
	} else {
		Me.conn = conn
		// 运行
		go Me.runWaitMsg()
	}
}

// 对外函数3：延时后重连
func (Me *Client) ReConnect(second int) {
	select {
	case <-time.After(time.Duration(second) * time.Second):
		Me.Connect()
	}
}

// 对外函数4：消息发送
func (Me *Client) SendMsg(msg UDataSocket) error {
	return sendSocketMsg(Me.conn, msg)
}

// 对外函数5：主动断开连接
func (Me *Client) DisConnect() {
	_ = Me.conn.Close()
}

// 内部函数1：通讯保持
func (Me *Client) runWaitMsg() {
	defer func() { fmt.Println(utilDateTime(), "退出 waitMsg") }()
	defer func() { _ = Me.conn.Close(); Me.conn = nil }()

	// 1、启动心跳
	stopHeartBeat := make(chan bool)
	go Me.heartBeat(stopHeartBeat)

	// 2、发送问候消息
	_ = sendSocketMsg(Me.conn, UDataSocket{0, 7, []byte("hello test msg from client")})

	// 3、循环接收指令
	Me.OnConnect(Me) // 回调连接成功事件
	if err := Me.getSocketMsg(Me.conn, func(msg *UDataSocket) bool {
		// 收到反馈的问候消息
		if msg.CType == 8 {
			fmt.Println(msg.Zlib, msg.CType, string(msg.Content))
		}

		Me.OnMessage(*msg, Me) // 回调收到消息事件
		return true
	}); err != nil {
		stopHeartBeat <- true

		Me.OnDisConnect(Me) // 回调连接断开事件
		return              // 退出用户运行协程
	}
}

// 内部函数1：心跳保持
func (Me *Client) heartBeat(stopHeartBeat chan bool) {
	defer func() { fmt.Println(utilDateTime(), "退出 heartBeat") }()

	T := time.NewTicker(time.Second * time.Duration(Me.heartBeatTime))
	for {
		select {
		case <-stopHeartBeat:
			return
		case <-T.C:
			// 发送心跳消息，CType为1
			if err := Me.SendMsg(UDataSocket{CType: 1}); err != nil {
				// fmt.Println("退出心跳协程,停止发送心跳")
				// return // 退出心跳协程,停止发送心跳
			}
		}
	}
}
