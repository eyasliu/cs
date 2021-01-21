# cmdsrv http

HTTP 适配器，该适配器比较特殊，因为 HTTP 协议的限制只能响应数据，不能推送数据，所以使用该适配器时，cmdsrv 的 `Context` 的主动推送数据的函数都将失效，只能响应数据

## 数据协议

只能通过 POST, PUT, DELETE 方法请求，因为只有这些请求才能在 HTTP BODY 放数据，数据类型是 json，遵循以下格式 ，对 http header 的 `Content-Type` 没有要求，无论设置成什么值，都会以 json 去解析 body 数据

**请求json**
```json
{
  "cmd":"register",
  "seqno":"unique_string",
  "data": {}
}
```

 * cmd 表示命令名，对应 cmdsrv 的路由
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

## 使用示例

```go
import (
  "net/http"
  "github.com/eyasliu/cmdsrv"
  "github.com/eyasliu/cmdsrv/xhttp"
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
