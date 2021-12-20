package socket_v1

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
)

// 1、结构体 -------------------------------------------------------------------------
type DataUnitSocket struct {
	Zlib    int    // 是否压缩 1:压缩
	CType   int    // 内容类型 1:客户端请求消息 2:服务端表接口消息 4:服务端表内容数据 200:服务端发送结束
	Content []byte // 发送内容
}
type sendUnit struct {
	SendFlag          int    // 消息最前面标记
	Zlib              int    // 压缩标记
	CType             int    // 内容类型
	ContentLength     int    // 原内容大小
	ContentTranLength int    // 发送内容大小
	ContentTran       []byte // 发送的内容
}
type SocketMsg struct {
}

// 2、全局变量 -------------------------------------------------------------------------
var sendFlag = 398359203 // 消息最前面标记

// 3、初始化函数 -------------------------------------------------------------------------

// 5、私有函数 -------------------------------------------------------------------------

func (Me *SocketMsg) SendSocketMsg(conn net.Conn, data DataUnitSocket) error {
	if conn == nil {
		return errors.New("发送失败:conn is nil")
	}

	b := bytes.NewBuffer([]byte{})
	// 需要压缩数据的，先压缩数据
	ContentSend := data.Content
	if data.Zlib == 1 {
		ContentSend = Me.zlibCompress(data.Content)
	}

	// 整合传输数据体
	sendData := sendUnit{
		SendFlag:          sendFlag,
		Zlib:              data.Zlib,
		CType:             data.CType,
		ContentLength:     len(data.Content),
		ContentTranLength: len(ContentSend),
		ContentTran:       ContentSend,
	}

	// 拼凑发送数据发送
	b.Write(Me.int2Bytes(sendData.SendFlag))
	b.Write(Me.int2Bytes(sendData.Zlib))
	b.Write(Me.int2Bytes(sendData.CType))
	b.Write(Me.int2Bytes(sendData.ContentLength))
	b.Write(Me.int2Bytes(sendData.ContentTranLength))
	b.Write(sendData.ContentTran)
	if n, err := conn.Write(b.Bytes()); err != nil || n != len(ContentSend)+20 {
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("数据未全部发送")
		}
		return err
	}
	return nil
}

// 读取socket消息
func (Me *SocketMsg) getSocketMsg(conn net.Conn, fSuccess func(msg *DataUnitSocket) bool) error {
	// 循环
	for {
		// 读取头数据
		buffHeader := make([]byte, 20)  // 建立一个slice
		_, err := conn.Read(buffHeader) // 读取客户端传来的内容
		if err != nil {
			return err
		}
		revSendFlag := Me.bytes2Int(buffHeader[0:4])
		revZlib := Me.bytes2Int(buffHeader[4:8])
		revCType := Me.bytes2Int(buffHeader[8:12])
		if revSendFlag != sendFlag {
			fmt.Println("传输码校验失败")
			return errors.New("传输码校验失败")
		}

		// 读取内容数据
		ContentTranLength := Me.bytes2Int(buffHeader[16:20])
		if revCType == 1 && ContentTranLength > 1*1024 {
			fmt.Println("头消息不能超过1K")
			return errors.New("头消息不能超过1K")
		}
		ContentTran, err := Me.readSocketSizeData(conn, ContentTranLength)
		if err != nil {
			return err
		}

		// 头数据和内容数据整合成传输数据体
		revData := sendUnit{
			SendFlag:          revSendFlag,
			Zlib:              revZlib,
			CType:             revCType,
			ContentLength:     Me.bytes2Int(buffHeader[12:16]),
			ContentTranLength: ContentTranLength,
			ContentTran:       ContentTran,
		}

		// 传输数据体校验
		if revData.ContentTranLength != len(revData.ContentTran) {
			fmt.Println("传输内容长度校验失败")
		}

		// 内容数据解压处理
		content := revData.ContentTran
		if revData.Zlib == 1 {
			content = Me.zlibUnCompress(revData.ContentTran)
		}
		// 校验原始内容长度
		if revData.ContentLength != len(content) {
			fmt.Println("内容长度校验失败")
		}
		// 回调
		continueRead := fSuccess(&DataUnitSocket{revData.Zlib, revData.CType, content})
		if !continueRead {
			break
		}
	}
	return nil
}

// 读取指定长度数据
func (Me *SocketMsg) readSocketSizeData(conn net.Conn, length int) ([]byte, error) {
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

// 进行zlib压缩
func (Me *SocketMsg) zlibCompress(src []byte) []byte {
	var in bytes.Buffer
	w := zlib.NewWriter(&in)
	_, _ = w.Write(src)
	_ = w.Close()
	return in.Bytes()
}

// 进行zlib解压缩
func (Me *SocketMsg) zlibUnCompress(compressSrc []byte) []byte {
	b := bytes.NewReader(compressSrc)
	var out bytes.Buffer
	r, _ := zlib.NewReader(b)
	_, _ = io.Copy(&out, r)
	return out.Bytes()
}

// 整形转换成字节
func (Me *SocketMsg) int2Bytes(n int) []byte {
	x := int32(n)

	bytesBuffer := bytes.NewBuffer([]byte{})
	_ = binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}
func (Me *SocketMsg) int64ToBytes(i int64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

// 字节转换成整形
func (Me *SocketMsg) bytes2Int(b []byte) int {
	bytesBuffer := bytes.NewBuffer(b)

	var x int32
	_ = binary.Read(bytesBuffer, binary.BigEndian, &x)

	return int(x)
}

func (Me *SocketMsg) bytes2Int64(buf []byte) int64 {
	return int64(binary.BigEndian.Uint64(buf))
}
