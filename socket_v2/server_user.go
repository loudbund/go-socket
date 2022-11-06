package socket_v2

import (
	"fmt"
	"net"
	"reflect"
	"time"
)

// 结构体1： 单个客户端用户结构体
type serverUser struct {
	socketMsg
	ClientId string // 客户端id(服务器给客户端分配的唯一标志)
	Addr     string // 客户端地址,ip和port

	conn   net.Conn         // 客户端连接
	server *Server          // Server指针，主要用来将自己加入和移除Server用户列表里
	c      chan UDataSocket // 发消息channel，Server将要消息推送到channel里
}

// 内部函数1：创建一个用户
func newUser(conn net.Conn, server *Server) *serverUser {
	user := &serverUser{
		ClientId:  utilUuidShort(),
		conn:      conn,
		Addr:      conn.RemoteAddr().String(),
		server:    server,
		c:         make(chan UDataSocket, 10),
		socketMsg: socketMsg{SendFlag: server.SendFlag},
	}
	return user
}

// //////////////////////////////////////////////////////////////////

// 内部函数2：保持连接
func (Me *serverUser) goListenClientMsg() {
	// 1、监听用户是否活跃的channel
	isLive := make(chan bool)

	// 2、接收消息
	go func() {
		if err := Me.getSocketMsg(Me.conn, func(msg *UDataSocket) bool {
			// 用户的任意消息，代表当前用户是一个活跃的
			isLive <- true

			// 收到问候的消息
			if msg.CType == 7 {
				fmt.Println(Me.ClientId, msg.Zlib, msg.CType, string(msg.Content))
				_ = Me.sendSocketMsg(Me.conn, UDataSocket{0, 8, []byte("hello test msg from server")})
			}

			// 3、消息发给主进程
			Me.server.ChanHookEvent <- &HookEvent{"message", Me, *msg}
			return true
		}); err != nil {
			// fmt.Println("消息接收终止，退出消息接收协程")
			isLive <- false
			Me.offline()
		}
	}()

	// 3、等待心跳
	Me.waitHeartBeet(isLive)
}

// 内部函数3：等待心跳
func (Me *serverUser) waitHeartBeet(isLive chan bool) {
	// 阻塞住
	for {
		select {
		case msg := <-Me.c:
			if !reflect.DeepEqual(msg, reflect.Zero(reflect.TypeOf(msg)).Interface()) {
				if err := Me.sendSocketMsg(Me.conn, msg); err != nil {
					Me.offline()
					return // 退出socket协程 // fmt.Println("消息发送失败，用户进程阻塞终止，退出用户协程")
				}
			}

		case live, ok := <-isLive:
			if !live || !ok { // fmt.Println("收到掉线通知，用户进程阻塞终止，退出用户协程")
				return
			}
			_ = Me.conn.SetDeadline(time.Now().Add(time.Duration(Me.server.ClientHeartTimeOut) * time.Second))

			// 当前用户是活跃的，应该重置定时器
			// 不做任何事情，为了激活select，更新下面的定时器

		case <-time.After(time.Second * time.Duration(Me.server.ClientHeartTimeOut)):
			Me.offline()

			return // runtime.Goexit() // fmt.Println("心跳过期，用户进程阻塞终止，退出用户协程")
		}
	}
}

// 内部函数4：用户的上线业务
func (Me *serverUser) online() {

	// 1、用户上线,将用户加入到onlineMap中
	Me.server.onlineMapLock.Lock()
	Me.server.onlineMap[Me.ClientId] = Me
	Me.server.onlineMapLock.Unlock()

	// 2、事件发给server
	Me.server.ChanHookEvent <- &HookEvent{"online", Me, UDataSocket{}}
}

// 内部函数5：用户的下线业务
func (Me *serverUser) offline() {

	// 1、用户下线，将用户从onlineMap里移除
	Me.server.onlineMapLock.Lock()
	if _, ok := Me.server.onlineMap[Me.ClientId]; ok {
		_ = Me.conn.Close()                      // 释放资源 - 关闭socket链接
		close(Me.c)                              // 释放资源 - 销毁用的资源
		delete(Me.server.onlineMap, Me.ClientId) // 移除用户
		fmt.Println(Me.ClientId, "退出成功", "当前在线", len(Me.server.onlineMap))
	}
	Me.server.onlineMapLock.Unlock()

	// 2、事件发给server
	Me.server.ChanHookEvent <- &HookEvent{"offline", Me, UDataSocket{}}
}
