package socket_v1

import (
	"fmt"
	"net"
	"time"
)

// 1、结构体 -------------------------------------------------------------------------
type Client struct {
	SocketMsg
	ip            string // ip地址
	port          int    // 端口
	heartBeatTime int    // 心跳时间

	conn     net.Conn // 连接句柄
	ClientId string   // 我的客户端id

	OnMessage     func(msg DataUnitSocket, C *Client) // 消息回调
	OnConnect     func(C *Client)                     // 上线回调
	OnConnectFail func(C *Client)                     // 上线回调
	OnDisConnect  func(C *Client)                     // 掉线回调
}

// 2、全局变量 -------------------------------------------------------------------------

// 3、初始化函数 -------------------------------------------------------------------------

// 开放的函数 -------------------------------------------------------------------------

// 初始化一个客户端
func NewClient(ip string, port int, OnMessage func(msg DataUnitSocket, C *Client), OnConnectFail, OnConnect, OnDisConnect func(C *Client)) *Client {
	C := &Client{
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
	return C
}

// 连接服务器
func (Me *Client) Connect() {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", Me.ip, Me.port))
	if err != nil {
		fmt.Println("net.Dial error:", err)
		Me.OnConnectFail(Me)
	} else {
		Me.conn = conn
		// 运行
		go Me.runWaitMsg()
	}
}

// 消息发送
func (Me *Client) SendMsg(msg DataUnitSocket) error {
	return Me.SendSocketMsg(Me.conn, msg)
}

// 内部函数 -------------------------------------------------------------------------

// 运行
func (Me *Client) runWaitMsg() {
	defer func() { fmt.Println(utilDateTime(), "退出 waitMsg") }()
	defer func() { _ = Me.conn.Close(); Me.conn = nil }()

	// 心跳
	stopHeartBeat := make(chan bool)
	go Me.heartBeat(stopHeartBeat)

	// 发送问候消息
	_ = Me.SendSocketMsg(Me.conn, DataUnitSocket{0, 7, []byte("hello test msg from client")})

	// 循环接收指令
	Me.OnConnect(Me)
	if err := Me.getSocketMsg(Me.conn, func(msg *DataUnitSocket) bool {
		// 收到反馈的问候消息
		if msg.CType == 8 {
			fmt.Println(msg.Zlib, msg.CType, string(msg.Content))
		}

		// 回调
		Me.OnMessage(*msg, Me)
		return true
	}); err != nil {
		stopHeartBeat <- true

		Me.OnDisConnect(Me)
		return // 退出用户运行协程
	}
}

// 心跳
func (Me *Client) heartBeat(stopHeartBeat chan bool) {
	defer func() { fmt.Println(utilDateTime(), "退出 heartBeat") }()

	T := time.NewTicker(time.Second * time.Duration(Me.heartBeatTime))
	for {
		select {
		case <-stopHeartBeat:
			return
		case <-T.C:
			// 发送心跳
			if err := Me.SendMsg(DataUnitSocket{CType: 1}); err != nil {
				// fmt.Println("退出心跳协程,停止发送心跳")
				// return // 退出心跳协程,停止发送心跳
			}
		}
	}
}
