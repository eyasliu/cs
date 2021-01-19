package cmdsrv

import (
	"runtime/debug"
)

// Recover 错误处理中间件
// 当处理函数发生 panic 时在该中间件恢复，并根据panic 的内容默认处理响应数据
func Recover() HandlerFunc {
	return func(c *Context) {
		defer func() {
			if data := recover(); data != nil {
				debug.PrintStack()
				c.Response.Code = -2
				if err, ok := data.(error); ok {
					c.Response.Msg = err.Error()
				} else if s, ok := data.(string); ok {
					c.Response.Msg = s
				} else if r, ok := data.(*Response); ok {
					c.Response.Code = r.Code
					c.Response.Msg = r.Msg
					if r.Data != nil {
						c.Response.Data = r.Data
					}
				}
			}
		}()
		c.Next()
	}
}
