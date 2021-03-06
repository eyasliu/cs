package cs

func fillPushResp(c *Context) error {
	c.Response.fill()
	return nil
}

type internalPanic byte

const (
	internalExitPanic internalPanic = 1 // 应用退出
)

func internalPanicHandler(c *Context) {
	defer func() {
		if data := recover(); data != nil {
			if _, ok := data.(internalPanic); !ok {
				panic(data)
			}
		}
	}()
	c.Next()
}
