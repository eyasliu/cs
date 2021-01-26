# cs tcp

TCP 的适配器实现

## 数据协议

#### tcp 数据包协议

该实现有个默认的tcp数据包协议实现，内部已基于该协议处理好了数据边界问题，无需考虑粘包半包问题

```
header               + data
[4字节标识data长度]     [任意长度]
```

支持自定义数据包协议，需要实现 `xtcp.MsgPkg` interface，在实例化的时候通过 `xtcp.Config` 指定

```go
// MsgPkg tcp 消息的编解码，处理封包解包
type MsgPkg interface {
  // Packer tcp 数据包的封装函数，传入的数据是需要发送的业务数据，返回发送给 tcp 的数据
  // data 是序列化后的响应数据
  // tcpProtoPkg 是根据私有协议封装好的tcp数据包
  // err 标识封包失败，该数据不会被发送
  Packer(data []byte) (tcpProtoPkg []byte, err error)      
  
  // Parser 将收到的数据包，根据私有协议转换成业务数据，在这里处理粘包,半包等数据包问题，返回处理好的数据包
  // sid 是会话ID，是tcp连接的标识
  // recvPkg 是tcp连接当次收到的数据
  // datas 根据私有解析解析好的数据，是个数组，表示可能解析出了 0 个或多个
  // err 是解析失败，如果 != nil 将会断开该连接 socket
	Parser(sid string, recvPkg []byte) (datas [][]byte, err error) 
}
```

#### 业务数据协议

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


## 使用示例

 1. 指定配置实例化 tcp 服务
 2. 通过 tcp 服务生成 cs 实例

```go
package main
import (
  "net"
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
- `Srv()` 会自动启动 tcp server，返回的 err 变量则是 tcp 启动可能会发生的异常
- `New()` 方法支持 `string` 和 `*xtcp.Config` 两种类型


**指定ip和端口**, 默认使用 `tcp` 网络类型
```go
server := xtcp.New("127.0.0.1:8520")
```

**使用配置对象**
```go
server := xtcp.New(&xtcp.Config{
  Addr: "/var/path/app.socket",
  Network: "unix", // tcp 的网络类型，可选值为 "tcp", "tcp4", "tcp6", "unix" or "unixpacket"
  // 不指定 MsgPkg 则使用内置默认数据协议
})
```

```go
server := xtcp.New(&xtcp.Config{
  Addr: "/var/path/app.socket",
  Network: "unix", // tcp 的网络类型，可选值为 "tcp", "tcp4", "tcp6", "unix" or "unixpacket"
  MsgPkg: &yourProtocolImpl{} // 指定自己的私有协议
})
```

