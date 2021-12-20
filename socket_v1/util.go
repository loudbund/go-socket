package socket_v1

import (
	"sync"
	"time"
)

// 1、结构体 -------------------------------------------------------------------------

// 2、全局变量 -------------------------------------------------------------------------
var lastAutoId int64
var lastAutiIdLock sync.RWMutex

// 3、初始化函数 -------------------------------------------------------------------------
func init() {
	lastAutoId = time.Now().UnixNano()
}

/**-------------------------
  // 名称：生成11位单机不重复字串 (耗时短 100000 条需要40毫秒)
  ***-----------------------*/
func utilUuidShort() string {
	dem := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-_"
	retBuf := make([]byte, 0)
	// int64位时间
	lastAutiIdLock.Lock()
	lastAutoId++
	T1 := lastAutoId
	lastAutiIdLock.Unlock()
	Mask := int64(63)
	for i := 0; i < 64; i += 6 {
		retBuf = append(retBuf, dem[T1>>i&Mask])
	}
	return string(retBuf)
}

/**-------------------------
// 名称：获取当前时间字串
// 参数: 无
// 返回："2021-08-25 11:16:20"
***-----------------------*/
func utilDateTime(T ...time.Time) string {
	timeObj := time.Now()
	if len(T) > 0 {
		timeObj = T[0]
	}
	return timeObj.Format("2006-01-02 15:04:05")
}
