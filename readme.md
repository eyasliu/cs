# Cmd Srv

[![Build Status](https://travis-ci.com/eyasliu/cs.svg)](https://travis-ci.com/eyasliu/cs)
[![Go Doc](https://godoc.org/github.com/eyasliu/cs?status.svg)](https://godoc.org/github.com/eyasliu/cs)
[![Code Coverage](https://codecov.io/gh/eyasliu/cs/branch/master/graph/badge.svg)](https://codecov.io/gh/eyasliu/cs/branch/master)
[![License](https://img.shields.io/github/license/eyasliu/cs.svg?style=flat)](https://github.com/eyasliu/cs)

开箱即用的基于命令的消息处理框架，让 websocket 和 tcp 开发就像 http 那样简单

# 使用示例


```go
package main
import (
  "net/http"
  "github.com/eyasliu/cs"
  "github.com/eyasliu/cs/xwebsocket"
)

func main() {
  // 初始化 websocket
  ws := xwebsocket.New()
  http.Handle("/ws", ws)

  srv := ws.Srv()
  srv.Use(cs.AccessLogger("MYSRV")). // 打印请求响应日志
            Use(cs.Recover()) // 统一错误处理，消化 panic 错误

  srv.Handle("register", func(c *cs.Context) {
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
    c.Broadcast(&cs.Response{
      Cmd:  "someone_online",
      Data: body,
    })

    // 往当前连接主动推送消息
    c.Push(&cs.Response{
      Cmd:  "welcome",
      Data: "welcome to register my server",
    })

    // 遍历所有在线会话，获取其他会话的状态，并往指定会话推送消息
    for _, sid := range c.GetAllSID() {
      if c.Srv.GetState(sid, "uid") != nil {
        c.Srv.Push(sid, &cs.Response{
          Cmd:  "firend_online",
          Data: "your firend is online",
        })
      }
    }
  })

  // 分组
  group := srv.Group(func(c *cs.Context) {
    // 过滤指定请求
    if _, ok := c.Get("uid").(int); !ok {
      c.Err(errors.New("unregister session"), 101)
      return
    }
    c.Next()
  })

  group.Handle("userinfo", func(c *cs.Context) {
    uid := c.Get("uid").(int) // 中间件已处理过，可大胆断言
    c.OK(map[string]interface{}{
      "uid": uid,
    })
  })
  go srv.Run()

  http.ListenAndServe(":8080", nil)
}
```


### 适配器

[用在 websocket](./xwebsocket)

```go
import (
  "net/http"
  "github.com/eyasliu/cs/xwebsocket"
)

func main() {
  ws := xwebsocket.New()
  http.Handler("/ws", ws.Handler)
  srv := ws.Srv(ws)
  go srv.Run()

  http.ListenAndServe(":8080", nil)
}
```

[用在 TCP](./xtcp)，使用内置默认协议

```go
import (
  "github.com/eyasliu/cs/xtcp"
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

[使用 gnet](./gnet), [gnet](https://github.com/panjf2000/gnet) 是一个性能非常高的网络框架，尤其在超高并发情况下性能更优，使用 gnet 作为 tcp 服务的底层框架

```go
import (
  "github.com/eyasliu/cs/xgnet"
)

func main() {
  server := xgnet.New("127.0.0.1:8520")
  srv, err := server.Srv()
  if err != nil {
    panic(err)
  }

  srv.Run() // 阻塞运行
}
```

[用在 HTTP](./xhttp)，支持请求响应，支持服务器主动推送

```go
import (
  "net/http"
  "github.com/eyasliu/cs"
  "github.com/eyasliu/cs/xhttp"
)

func main() {
  server := xhttp.New()
  http.Handle("/cmd", server)
  http.HandleFunc("/cmd2", server.Handler)
  srv := server.Srv()
  go http.ListenAndServe(":8080", nil)
	
  srv.Run()
}
```

多个适配器混用，让 websocket, tcp, http 共用同一套逻辑

```go
import (
  "net/http"
  "github.com/eyasliu/cs"
  "github.com/eyasliu/cs/xhttp"
  "github.com/eyasliu/cs/xwebsocket"
  "github.com/eyasliu/cs/xtcp"
)

func main() {
  // http adapter
  server := xhttp.New()
  http.Handle("/cmd", server)
  http.HandleFunc("/cmd2", server.Handler)
  
  // websocket adapter
  ws := xwebsocket.New()
  http.Handle("/ws", server)

  // tcp adapter
  tcp := xtcp.New()

  // boot srv
  go tcp.Run()
  go http.ListenAndServe(":8080", nil)

  srv := cs.New(server, ws, tcp)
  srv.Run() // 阻塞运行
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

在 websocket 和 tcp 中，每个连接都抽象成一个字符串 `SID`, 即 Session ID, cs 只负责处理消息，不处理连接的任何状态，与连接和状态相关的操作全都以 interface 定义好，给各种工具去实现

## API

[GoDoc](https://pkg.go.dev/github.com/eyasliu/cs)