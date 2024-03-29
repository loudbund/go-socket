package socket_v2

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"unsafe"
)

// 结构体1：(外部用)传输数据上层结构体
type UDataSocket struct {
	Zlib    int8   // 是否压缩 1:压缩
	CType   int16  // 内容类型 1:客户端请求消息 2:服务端表接口消息 4:服务端表内容数据 200:服务端发送结束
	Content []byte // 发送内容
}

// 结构体2：(内部用)传输数据的header头信息,共15个字节
type unitDataHeader struct {
	SendFlag          int32 // 消息最前面标记
	Zlib              int8  // 压缩标记 (同 UDataSocket.Zlib)
	CType             int16 // 内容类型 (同 UDataSocket.CType)
	ContentLength     int32 // 原内容大小
	ContentTranLength int32 // 发送内容大小
}

// 结构体3：(内部用)传输数据底层结构体
type unitDataSend struct {
	SendFlag          int32  // 消息最前面标记
	Zlib              int8   // 压缩标记 (同 UDataSocket.Zlib)
	CType             int16  // 内容类型 (同 UDataSocket.CType)
	ContentLength     int32  // 原内容大小
	ContentTranLength int32  // 发送内容大小
	ContentTran       []byte // 发送的内容 (同 UDataSocket.Content)
}

// 结构体4：本模块封装用结构体
type socketMsg struct {
	SendFlag int
}

// 内部函数1：发送socket消息
func (Me *socketMsg) sendSocketMsg(conn net.Conn, data UDataSocket) error {
	if conn == nil {
		return errors.New("发送失败:conn is nil")
	}

	b := bytes.NewBuffer([]byte{})
	// 需要压缩数据的，先压缩数据
	ContentSend := data.Content
	if data.Zlib == 1 {
		ContentSend = utilZLibCompress(data.Content)
	}

	// 整合传输数据体
	sendDataHeader := unitDataHeader{
		SendFlag:          int32(Me.SendFlag),
		Zlib:              data.Zlib,
		CType:             data.CType,
		ContentLength:     int32(len(data.Content)),
		ContentTranLength: int32(len(ContentSend)),
	}

	// 拼凑发送数据发送
	b.Write((*(*[15]byte)(unsafe.Pointer(&sendDataHeader)))[:])
	b.Write(ContentSend)
	if n, err := conn.Write(b.Bytes()); err != nil || n != len(ContentSend)+15 {
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("数据未全部发送")
		}
		return err
	}
	return nil
}

// 内部函数2：读取socket消息
func (Me *socketMsg) getSocketMsg(conn net.Conn, fSuccess func(msg *UDataSocket) bool) error {
	// 循环
	for {
		// 读取头数据
		buffHeader := make([]byte, 15)  // 建立一个slice
		_, err := conn.Read(buffHeader) // 读取客户端传来的内容
		if err != nil {
			return err
		}
		headerData := *(*unitDataHeader)(unsafe.Pointer(&buffHeader[0]))
		if int(headerData.SendFlag) != Me.SendFlag {
			fmt.Println("传输码校验失败")
			return errors.New("传输码校验失败")
		}

		// 读取内容数据
		if headerData.Zlib == 1 && headerData.ContentTranLength > 1*1024 {
			fmt.Println("头消息不能超过1K")
			return errors.New("头消息不能超过1K")
		}
		ContentTran, err := Me.readSocketSizeData(conn, int(headerData.ContentTranLength))
		if err != nil {
			return err
		}

		// 头数据和内容数据整合成传输数据体
		revData := unitDataSend{
			SendFlag:          headerData.SendFlag,
			Zlib:              headerData.Zlib,
			CType:             headerData.CType,
			ContentLength:     headerData.ContentLength,
			ContentTranLength: headerData.ContentTranLength,
			ContentTran:       ContentTran,
		}
		fmt.Println(999, revData)

		// 传输数据体校验
		if int(revData.ContentTranLength) != len(revData.ContentTran) {
			fmt.Println("传输内容长度校验失败")
			return errors.New("传输内容长度校验失败")
		}

		// 内容数据解压处理
		content := revData.ContentTran
		if revData.Zlib == 1 {
			content = utilZLibUnCompress(revData.ContentTran)
		}
		// 校验原始内容长度
		if int(revData.ContentLength) != len(content) {
			fmt.Println("内容长度校验失败")
			return errors.New("内容长度校验失败")
		}
		// 回调
		continueRead := fSuccess(&UDataSocket{revData.Zlib, revData.CType, content})
		if !continueRead {
			break
		}
	}
	return nil
}

// 内部函数3：读取指定长度数据
func (Me *socketMsg) readSocketSizeData(conn net.Conn, length int) ([]byte, error) {
	var retBuff []byte

	for {
		ContentTran := make([]byte, length) // 建立一个slice
		// 读取传输内容
		if n, err := conn.Read(ContentTran); err == nil {
			if n != length {
				retBuff = append(retBuff, ContentTran[:n]...)
				length -= n
			} else {
				retBuff = append(retBuff, ContentTran[:n]...)
				break
			}
		} else {
			return nil, err
		}
	}
	return retBuff, nil
}
