package cmdsrv

func fillPushResp(c *Context) error {
	c.Response.fill()
	return nil
}