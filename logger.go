package cmdsrv

import "fmt"

type printLogger interface {
	Debug(...interface{})
}

type consoleLogger struct{}

func (consoleLogger) Debug(a ...interface{}) {
	fmt.Println(a...)
}

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
		logger.Debug(fmt.Sprintf("%s RECV CMD=%s SEQ=%s %s", name, c.Cmd, c.Seqno, string(c.RawData)))
		c.Next()
		logger.Debug(fmt.Sprintf("%s RESP CMD=%s SEQ=%s %v", name, c.Cmd, c.Seqno, c.Response.Data))
	}
}
