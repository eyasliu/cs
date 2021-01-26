package cs

import (
	"encoding/json"
	"fmt"
)

type printLogger interface {
	Debug(...interface{})
}

type consoleLogger struct{}

func (consoleLogger) Debug(a ...interface{}) {
	fmt.Println(a...)
}

const loggerCtxKey = "__logger_middleware_instence__"
const loggerNameCtxKey = "__logger_middleware_name__"

func getLogDataString(v interface{}) string {
	data := ""
	switch v.(type) {
	case nil:
		data = "nil"
	case string:
		data = v.(string)
	case []byte:
		data = string(v.([]byte))
	default:
		bt, _ := json.Marshal(v)
		data = string(bt)
	}
	return data
}

// AccessLogger 打印请求响应中间件
// 2 个可选参数，如果参数是 printLogger 接口类型则用于设置打印日志的 Logger 实例，如果是 string 类型则用于设置日志前缀
// AccessLogger("MySRV") 设置名称
// AccessLogger("MySRV", logger) 设置名称和打日志的实例
// AccessLogger(logger, "MySRV") 设置名称和打日志的实例
// AccessLogger(logger) 设置打日志的实例
// AccessLogger(123) 无效参数，不会产生异常，等价于没有参数
func (s *Srv) AccessLogger(args ...interface{}) HandlerFunc {
	var logger printLogger = consoleLogger{}
	name := "SRV"
	for _, v := range args {
		if n, ok := v.(string); ok {
			name = n
		} else if l, ok := v.(printLogger); ok {
			logger = l
		}
	}

	s.UsePush(func(c *Context) error {
		logger.Debug(fmt.Sprintf("%s PUSH SID=%s CMD=%s %s", name, c.SID, c.Cmd, getLogDataString(c.Data)))
		return nil
	})

	return func(c *Context) {
		if c.Get(loggerCtxKey) == nil {
			c.Set(loggerCtxKey, logger)
			c.Set(loggerNameCtxKey, name)
		}
		if (c.Cmd == CmdConnected ||
			c.Cmd == CmdClosed ||
			c.Cmd == CmdHeartbeat) &&
			(len(c.handlers) == len(c.Srv.middleware)) { // 有路由监听的话就要打印
			c.Next()
			return
		}
		logger.Debug(fmt.Sprintf("%s RECV SID=%s CMD=%s SEQ=%s %s", name, c.SID, c.Cmd, c.Seqno, string(c.RawData)))
		c.Next()
		logger.Debug(fmt.Sprintf("%s RESP SID=%s CMD=%s SEQ=%s %s", name, c.SID, c.Cmd, c.Seqno, getLogDataString(c.Data)))
	}
}
