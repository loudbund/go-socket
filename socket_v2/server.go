package socket_v2

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

// 结构体1：服务结构体数据
type Server struct {
	Ip                   string
	Port                 int
	ClientHeartTimeOut   int                    // 客户端超时时间 默认60秒
	OnHookEvent          func(Msg HookEvent)    // hook回调消息
	ChanHookEvent        chan *HookEvent        // 所有消息，各个子连接传过来的
	chanBroadCastMessage chan UDataSocket       // 消息广播的channel
	onlineMap            map[string]*serverUser // 在线用户的列表
	onlineMapLock        sync.RWMutex           // 同步锁
	SendFlag             int                    // socket验证标记
}

// 结构体2：hook消息结构体
type HookEvent struct {
	EventType string // 事件类型 online / offline / message
	User      *serverUser
	Message   UDataSocket
}

// 对外函数1：创建一个server的实例
func NewServer(ip string, port int, OnHookEvent func(Msg HookEvent)) *Server {
	server := &Server{
		Ip:                   ip,
		Port:                 port,
		onlineMap:            make(map[string]*serverUser),
		ClientHeartTimeOut:   60 * 3,
		chanBroadCastMessage: make(chan UDataSocket),
		ChanHookEvent:        make(chan *HookEvent),
		OnHookEvent:          OnHookEvent,
		SendFlag:             398359203,
	}
	go server.goWaitNewClient()
	return server
}

// 对外函数2：连接服务器
func (Me *Server) Set(opt string, value interface{}) {
	if opt == "SendFlag" {
		Me.SendFlag = value.(int)
	}
}

// 对外函数2：消息发送，ClientId为nil，发给所有客户端
func (Me *Server) SendMsg(ClientId *string, Msg UDataSocket) error {
	if ClientId == nil {
		// 将msg发送给全部的在线User
		Me.onlineMapLock.Lock()
		for _, cli := range Me.onlineMap {
			cli.c <- Msg
		}
		Me.onlineMapLock.Unlock()

		return nil
	} else {
		Me.onlineMapLock.Lock()
		user, ok := Me.onlineMap[*ClientId]
		Me.onlineMapLock.Unlock()

		if ok {
			if err := user.sendSocketMsg(user.conn, Msg); err != nil {
				user.offline()
				return err
			}
			return nil
		} else {
			return errors.New("用户不在线")
		}
	}
}

// 内部函数1：启动服务器的接口
func (Me *Server) goWaitNewClient() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", Me.Ip, Me.Port))
	if err != nil {
		log.Panic(err)
	}
	defer func() { _ = listener.Close() }()
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

// 内部函数2：处理客户端连接
func (Me *Server) goWelcomeNewClient(conn net.Conn) {
	// 新用户来了
	user := newUser(conn, Me)
	user.online()
	fmt.Println("链接建立成功", user.ClientId, " 当前用户:", len(Me.onlineMap))

	user.goListenClientMsg()
	// fmt.Println("用户守护进程已退出！")
}

// 内部函数3：转发hook所有消息
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
