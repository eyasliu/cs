package cmdsrv

func fillPushResp(c *Context) {
	c.Response.fill()
	c.Next()
}
