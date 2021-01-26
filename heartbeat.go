package cs

import (
	"sync"
	"time"
)

// Heartbeat 会话心跳维护，如果会话在指定周期内没有发送任何数据，则关闭该连接会话
// timeout 心跳过期时长，指定当会话间隔
// 重置心跳过期时间，当接收到了会话的任意命令，都会重置
func Heartbeat(timeout time.Duration, srv *Srv) HandlerFunc {
	heartbeatTime := sync.Map{}
	// check heartbeat timeout
	go func() {
		for {
			time.Sleep(timeout / 2)
			if srv == nil {
				continue
			}
			to := time.Now().Add(-1 * timeout).Unix()
			heartbeatTime.Range(func(key, val interface{}) bool {
				if hbTime, ok := val.(int64); !ok || hbTime < to {
					sid := key.(string)
					srv.Close(sid)
					heartbeatTime.Delete(sid)
				}
				return true
			})
		}
	}()
	return func(c *Context) {
		// TIP: Store syncMap may block current goroutine longtime, should I use new goroutine to Store?
		heartbeatTime.Store(c.SID, time.Now().Unix())
		c.Next()
	}
}

// Heartbeat 会话心跳维护，如果会话在指定周期内没有发送任何数据，则关闭该连接会话
// timeout 心跳过期时长，指定当会话间隔
// 重置心跳过期时间，当接收到了会话的任意命令，都会重置
func (srv *Srv) Heartbeat(timeout time.Duration) HandlerFunc {
	return Heartbeat(timeout, srv)
}
