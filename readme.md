# Cmd Srv

![Build Status](https://travis-ci.com/eyasliu/cmdsrv.svg)

开箱即用的基于命令的消息处理框架，让 websocket 和 tcp 开发就像 http 那样简单

# 使用示例


```go
package main
import (
    "net/http"
    "github.com/eyasliu/cmdsrv"
    "github.com/eyasliu/cmdsrv/xwebsocket"
)

func main() {
	// 初始化 websocket
	ws := xwebsocket.New()
	http.Handle("/ws", ws)

	srv := ws.New(ws)
	srv.Use(cmdsrv.AccessLogger("MYSRV")). // 打印请求响应日志
		Use(cmdsrv.Recover())             // 统一错误处理，消化 panic 错误

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
			if c.Server.GetState(sid, "uid") != nil {
				c.Srv.Push(sid, &cmdsrv.Response{
					Cmd:  "firend_online",
					Data: "your firend is online",
				})
			}
		}
	})

	// 分组
	group := srv.Group(func(c *cmdsrv.Context) {
		// 过滤指定请求
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
	go srv.Run()

	log.Fatal(http.ListenAndServe(":8080", nil))
}
```


### 适配器

用在 websocket

```go
import (
    "net/http"
    "github.com/eyasliu/cmdsrv/xwebsocket"
)

func main() {
    ws := xwebsocket.New()
    http.Handler("/ws", ws.Handler)
		srv := ws.Srv(ws)
		go srv.Run()

    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

用在 TCP，使用内置默认协议

```go
import (
    "net"
    "github.com/eyasliu/cmdsrv/xtcp"
)

func main() {
    server := xtcp.New("127.0.0.1:8520")
		srv, err := server.Srv()
		if err != nil {
			panic(err)
		}

		srv.Run() // 阻塞运行
}
```
用在 HTTP

```go
import (
    "net/http"
    "github.com/eyasliu/cmdsrv"
    "github.com/eyasliu/cmdsrv/xhttp"
)

func main() {
    server := xhttp.New()
    http.Handle("/cmd1", server)
    http.HandleFunc("/cmd1", server.Handler)
		srv := server.Srv()
		// http 不需要 srv.Run()

    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

```sh
$ curl -XPOST -H"Content-Type:application/json" --data '{"cmd":"register", "data":{"uid": 101, "name": "eyasliu"}}' http://localhost:8080/cmd
{"cmd":"register","data":{"timestamp": 1610960488}}
```

# 实现过程

在开发 websocket 和 tcp 的时候，对于长连接的消息处理都需要手动处理，并没有类似于 http 的路由那么方便，于是就想要实现一个可以处理该类消息的工具。

在长连接的开发中，经常遇到的一些问题：

 1. 每个连接会话在连接后都需要注册，以标识该连接的用途
 2. 每个请求都需要处理，并且保证一定有响应
 3. 往连接主动推送消息
 4. 给所有连接广播消息
 5. 消息处理的代码不够优雅，太多 switch case 等样板代码
 6. 请求的数据解析不好写

实现方案：

在 websocket 和 tcp 中，每个连接都抽象成一个字符串 `SID`, 即 Session ID, cmdsrv 只负责处理消息，不处理连接的任何状态，与连接和状态相关的操作全都以 interface 定义好，给各种工具去实现
