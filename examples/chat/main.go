package main

import (
	"errors"
	"net/http"

	"github.com/eyasliu/cmdsrv"
	"github.com/eyasliu/cmdsrv/xhttp"
	"github.com/eyasliu/cmdsrv/xwebsocket"
)

func assert(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	httpAdapter := xhttp.New()
	http.Handle("/http", httpAdapter)

	wsAdapter := xwebsocket.New()
	http.Handle("/ws", wsAdapter)
	http.Handle("/", http.FileServer(http.Dir(`.`)))

	go http.ListenAndServe(":12000", nil)

	srv := cmdsrv.New(httpAdapter, wsAdapter)
	srv.Use(cmdsrv.AccessLogger("CHAT"))
	srv.Use(cmdsrv.Recover())

	srv.Handle("register", func(c *cmdsrv.Context) {
		var body struct {
			Name string `p:"name" v:"required|min:4#名称必填|名称最小4个字符"`
		}
		assert(c.Parse(&body))

		c.Set("name", body.Name)

		c.Push(&cmdsrv.Response{
			Cmd:  "welcome",
			Data: "welcome " + body.Name + " to my chat room",
		})

		c.Broadcast(&cmdsrv.Response{
			Cmd:  "user_online",
			Data: body,
		})
	})

	user := srv.Group(func(c *cmdsrv.Context) {
		if c.Get("name") == nil {
			c.Abort()
			c.Err(errors.New("you are not login"), 101)
			return
		}
		c.Next()
	})
	user.Handle("message", func(c *cmdsrv.Context) {
		var body struct {
			Message string `v:"required#消息不能为空"`
		}
		assert(c.Parse(&body))
		c.OK()

		name := c.Get("name").(string)
		msg := map[string]interface{}{
			"name":    name,
			"message": body.Message,
		}
		pushMsg := &cmdsrv.Response{
			Cmd:  "push_message",
			Data: msg,
		}
		for _, sid := range c.GetAllSID() {
			if c.Srv.GetState(sid, "name") != nil {
				c.Srv.Push(sid, pushMsg)
			} else {
				c.Srv.Push(sid, &cmdsrv.Response{
					Cmd: "hide_message",
				})
			}
		}
	})

	srv.Run()
}
