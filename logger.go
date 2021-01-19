package cmdsrv

import "fmt"

type printLogger interface {
	Debug(...interface{})
}

type consoleLogger struct{}

func (consoleLogger) Debug(a ...interface{}) {
	fmt.Println(a...)
}

// AccessLogger 打印请求响应中间件
// 2 个可选参数，如果参数是 printLogger 接口类型则用于设置打印日志的 Logger 实例，如果是 string 类型则用于设置日志前缀
// AccessLogger("MySRV") 设置名称
// AccessLogger("MySRV", logger) 设置名称和打日志的实例
// AccessLogger(logger, "MySRV") 设置名称和打日志的实例
// AccessLogger(logger) 设置打日志的实例
// AccessLogger(123) 无效参数，不会产生异常，等价于没有参数
func AccessLogger(args ...interface{}) HandlerFunc {
	var logger printLogger = consoleLogger{}
	name := "SRV"
	for _, v := range args {
		if n, ok := v.(string); ok {
			name = n
		} else if l, ok := v.(printLogger); ok {
			logger = l
		}
	}
	return func(c *Context) {
		if c.Cmd == CmdConnected ||
			c.Cmd == CmdClosed ||
			c.Cmd == CmdHeartbeat {
			c.Next()
			return
		}
		logger.Debug(fmt.Sprintf("%s RECV CMD=%s SEQ=%s %s", name, c.Cmd, c.Seqno, string(c.RawData)))
		c.Next()
		logger.Debug(fmt.Sprintf("%s RESP CMD=%s SEQ=%s %v", name, c.Cmd, c.Seqno, c.Response.Data))
	}
}
