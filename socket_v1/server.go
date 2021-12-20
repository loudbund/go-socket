package socket_v1

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

// 1、结构体 -------------------------------------------------------------------------
type Server struct {
	Ip                   string
	Port                 int
	OnlineMap            map[string]*SocketUser    // 在线用户的列表
	ClientHeartTimeOut   int                 // 客户端超时时间 默认60秒
	OnHookEvent          func(Msg HookEvent) // hook回调消息
	ChanHookEvent        chan *HookEvent     // 所有消息，各个子连接传过来的
	chanBroadCastMessage chan DataUnitSocket // 消息广播的channel
	MapLock              sync.RWMutex        // 同步锁
}

//
type HookEvent struct {
	EventType string // 事件类型 online / offline / message
	User      *SocketUser
	Message   DataUnitSocket
}

// 2、全局变量 -------------------------------------------------------------------------

// 3、初始化函数 -------------------------------------------------------------------------

// 创建一个server的实例
func NewServer(ip string, port int, OnHookEvent func(Msg HookEvent)) *Server {
	server := &Server{
		Ip:                   ip,
		Port:                 port,
		OnlineMap:            make(map[string]*SocketUser),
		ClientHeartTimeOut:   60 * 3,
		chanBroadCastMessage: make(chan DataUnitSocket),
		ChanHookEvent:        make(chan *HookEvent),
		OnHookEvent:          OnHookEvent,
	}
	go server.goWaitNewClient()
	return server
}

// 消息发送
func (Me *Server) SendMsg(ClientId *string, Msg DataUnitSocket) error {
	if ClientId == nil {
		// 将msg发送给全部的在线User
		Me.MapLock.Lock()
		for _, cli := range Me.OnlineMap {
			cli.C <- Msg
		}
		Me.MapLock.Unlock()

		return nil
	} else {
		if user, ok := Me.OnlineMap[*ClientId]; ok {
			return user.SendSocketMsg(user.Conn, Msg)
		} else {
			return errors.New("用户不在线")
		}
	}
}

// ////////////////////////////////////////////////////

// 启动服务器的接口
func (Me *Server) goWaitNewClient() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", Me.Ip, Me.Port))
	if err != nil {
		log.Panic(err)
	}
	defer listener.Close()
	fmt.Printf("Waiting for clients , ip port %s:%d\n", Me.Ip, Me.Port)

	// 启动监听Message的goroutine
	go Me.goTranHookMessage()

	// 等待客户端连接
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}
		go Me.goWelcomeNewClient(conn)
	}
}

// 处理客户端连接
func (Me *Server) goWelcomeNewClient(conn net.Conn) {
	// 新用户来了
	user := NewUser(conn, Me)
	user.Online()
	fmt.Println("链接建立成功", user.ClientId, " 当前用户:", len(Me.OnlineMap))

	user.goListenClientMsg()
	// fmt.Println("用户守护进程已退出！")
}

// /////////////////////////////////////////////////

// 转发hook所有消息
func (Me *Server) goTranHookMessage() {
	for {
		select {
		case Event, ok := <-Me.ChanHookEvent:
			if !ok {
				return
			}
			// 推给应用的事件，除了心跳的所有事件
			if Event.Message.CType != 1 {
				Me.OnHookEvent(*Event)
			}
		}
	}
}
