package xwebsocket_test

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/eyasliu/cmdsrv"
	"github.com/eyasliu/cmdsrv/xwebsocket"
	"github.com/gogf/gf/test/gtest"
	"github.com/gorilla/websocket"
)

func sendToWs(url string, r interface{}) (map[string]interface{}, error) {
	c, _, err := websocket.DefaultDialer.Dial(url, nil)

	if err != nil {
		return nil, err
	}
	defer c.Close()
	bt, _ := json.Marshal(r)
	err = c.WriteMessage(websocket.TextMessage, bt)
	if err != nil {
		return nil, err
	}

	_, data, err := c.ReadMessage()
	if err != nil {
		return nil, err
	}
	// fmt.Println(string(data))
	res := map[string]interface{}{}
	err = json.Unmarshal(data, &res)
	return res, err
}

func TestWS(t *testing.T) {
	ws := xwebsocket.New()
	http.Handle("/ws1", ws)
	http.HandleFunc("/ws2", ws.Handler)

	go http.ListenAndServe(":5679", nil)

	srv := ws.Srv().Use(cmdsrv.AccessLogger("MYSRV")) // 打印请求响应日志

	gtest.C(t, func(t *gtest.T) {
		data := map[string]interface{}{
			"cmd":   "register",
			"seqno": "12345",
			"data":  "asdfgh",
		}
		srv.Handle(data["cmd"].(string), func(c *cmdsrv.Context) {
			c.OK(c.RawData)
		})

		go srv.Run()
		time.Sleep(100 * time.Millisecond)

		res, err := sendToWs("ws://127.0.0.1:5679/ws1", data)

		t.Assert(err, nil)
		t.Assert(res["cmd"], data["cmd"])
		t.Assert(res["data"], data["data"])
		t.Assert(res["seqno"], data["seqno"])

		res, err = sendToWs("ws://127.0.0.1:5679/ws2", data)

		t.Assert(err, nil)
		t.Assert(res["cmd"], data["cmd"])
		t.Assert(res["data"], data["data"])
		t.Assert(res["seqno"], data["seqno"])
	})

}

func ExampleFull() {
	// 初始化 websocket
	ws := xwebsocket.New()
	http.Handle("/ws", ws)

	srv := cmdsrv.New(ws)

	srv.Use(cmdsrv.AccessLogger("MYSRV")) // 打印请求响应日志
	srv.Use(cmdsrv.Recover())             // 统一错误处理，消化 panic 错误

	srv.Handle("register", func(c *cmdsrv.Context) {
		// 定义请求数据
		var body struct {
			UID  int    `p:"uid" v:"required"`
			Name string `p:"name" v:"required|min:4#必需指定名称|名称长度必需大于4位"`
		}
		// 解析请求数据
		if err := c.Parse(&body); err != nil {
			c.Err(err, 401)
			return
		}
		// 设置会话状态数据
		c.Set("uid", body.UID)
		c.Set("name", body.Name)

		// 响应消息
		c.OK(map[string]interface{}{
			"timestamp": time.Now().Unix(),
		})

		// 给所有连接广播消息
		c.Broadcast(&cmdsrv.Response{
			Cmd:  "someone_online",
			Data: body,
		})

		// 往当前连接主动推送消息
		c.Push(&cmdsrv.Response{
			Cmd:  "welcome",
			Data: "welcome to register my server",
		})

		// 遍历所有在线会话，获取其他会话的状态，并往指定会话推送消息
		for _, sid := range c.GetAllSID() {
			if c.Srv.GetState(sid, "uid") != nil {
				c.Srv.Push(sid, &cmdsrv.Response{
					Cmd:  "firend_online",
					Data: "your firend is online",
				})
			}
		}
	})

	// 分组
	group := srv.Group(func(c *cmdsrv.Context) {
		// 过滤指定会话
		if _, ok := c.Get("uid").(int); !ok {
			c.Err(errors.New("unregister session"), 101)
			return
		}
		c.Next()
	})

	group.Handle("userinfo", func(c *cmdsrv.Context) {
		uid := c.Get("uid").(int) // 中间件已处理过，可大胆断言
		c.OK(map[string]interface{}{
			"uid": uid,
		})
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
