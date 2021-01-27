package main

import (
	"errors"
	"net/http"

	"github.com/eyasliu/cs"
	"github.com/eyasliu/cs/xhttp"
	"github.com/eyasliu/cs/xwebsocket"
)

func assert(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	httpAdapter := xhttp.New()
	http.Handle("/sse", httpAdapter)

	wsAdapter := xwebsocket.New()
	http.Handle("/ws", wsAdapter)
	http.Handle("/", http.FileServer(http.Dir(`.`)))

	go http.ListenAndServe(":12000", nil)

	srv := cs.New(httpAdapter, wsAdapter)
	srv.Use(srv.AccessLogger("CHAT"))
	srv.Use(cs.Recover())

	srv.Handle("register", func(c *cs.Context) {
		var body struct {
			Name string `p:"name" v:"required#名称必填" json:"name"`
		}
		assert(c.Parse(&body))

		c.Set("name", body.Name)

		c.Push(&cs.Response{
			Cmd:  "welcome",
			Data: "welcome " + body.Name + " to my chat room",
		})

		c.Broadcast(&cs.Response{
			Cmd:  "user_online",
			Data: body,
		})
	})

	user := srv.Group(func(c *cs.Context) {
		if c.Get("name") == nil {
			c.Abort()
			c.Err(errors.New("you are not login"), 101)
			return
		}
		c.Next()
	})
	user.Handle("new_message", func(c *cs.Context) {
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
		pushMsg := &cs.Response{
			Cmd:  "push_message",
			Data: msg,
		}
		for _, sid := range c.GetAllSID() {
			if c.GetState(sid, "name") != nil {
				c.PushSID(sid, pushMsg)
			} else {
				c.PushSID(sid, &cs.Response{
					Cmd: "hide_message",
				})
			}
		}
	})
	user.Handle(cs.CmdClosed, func(c *cs.Context) {
		for _, sid := range c.GetAllSID() {
			if c.GetState(sid, "name") != nil {
				c.PushSID(sid, &cs.Response{
					Cmd: "user_offline",
					Data: map[string]interface{}{
						"name": c.Get("name"),
					},
				})
			}
		}
	})

	srv.Run()
}
