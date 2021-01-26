# cs http

HTTP 适配器，该适配器支持请求响应模式，服务端推送使用 SSE(Server-Sent Event)，所以需要客户端支持，会话 sid 使用 cookie 做保持

## 数据协议

#### 请求响应

请求响应模式只能通过 POST, PUT, DELETE 方法发起，因为只有这些请求才能在 HTTP BODY 放数据，数据类型是 json，遵循以下格式 ，对 http header 的 `Content-Type` 没有要求，无论设置成什么值，都会以 json 去解析 body 数据

**请求json**
```json
{
  "cmd":"register",
  "seqno":"unique_string",
  "data": {}
}
```

 * cmd 表示命令名，对应 cs 的路由
 * seqno 表示该请求的唯一标识，在响应中会原样返回
 * data 表示请求数据，可以是任意值，如 string, number, object, array, null


**响应**
```json
{
  "cmd":"register",
  "seqno":"unique_string",
  "code": 0,
  "msg":"ok",
  "data": {}
}
```

 * cmd 表示命令名，对应请求的 cmd
 * seqno 表示该请求的唯一标识，对应请求的 seqno
 * code 响应状态码，不等于 0 表示异常， -1 表示不支持请求的cmd，其他业务码根据业务适应
 * msg 响应说明，只在code不等于 0 时才有意义
 * data 表示响应数据，可能是任意值，如 string, number, object, array, null

#### 服务器推送

客户端通过 EventSource 连接上时必须要带上 Cookie，因为使用 Cookie 作为会话记录，否则将无法推送至对应客户端

## 使用示例

```go
import (
  "net/http"
  "github.com/eyasliu/cs"
  "github.com/eyasliu/cs/xhttp"
)

func main() {
  server := xhttp.New()
  http.Handle("/cmd", server)
  http.HandleFunc("/cmd1", server.Handler)
  srv := server.Srv()
  // http 不需要 srv.Run()

  log.Fatal(http.ListenAndServe(":8080", nil))
}
```

 - 该适配器是不需要调用 `srv.Run()`，事实上即使调用了也不会有异常，会启动一个无用的 goroutine, 并且您将无法关闭它
 - server 实现了 `net/http` 标准库的 `http.Handler` 接口，`server.Handler` 实现了 `http.HandlerFunc`，所以可以方便的用在其他 web 框架中

客户端在使用的时候

```sh
$ curl -XPOST --data '{"cmd":"register", "seqno": "unique_string", "data":{"name": "eyasliu"}}' http://127.0.0.1:8080/cmd
{"cmd":"register","seqno": "unique_string","code":0,"msg":"ok","data":{"timestamp": 1610960488}}
```
