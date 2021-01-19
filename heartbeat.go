package cmdsrv

import (
	"sync"
	"time"
)

func Heartbeat(timeout time.Duration) HandlerFunc {
	heartbeatTime := sync.Map{}
	var srv *Srv
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
		if srv == nil {
			srv = c.Srv
		}
		// TIP: Store syncMap may block current goroutine longtime, should I use new goroutine to Store?
		heartbeatTime.Store(c.SID, time.Now().Unix())
		c.Next()
	}
}
