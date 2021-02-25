# Command Service JSSDK

github: https://github.com/eyasliu/cs

开箱即用的基于命令的消息处理框架，让 websocket 和 tcp 开发就像 http 那样简单

**JSSDK** 支持 http 和 websocket 连接

# install

via npm:

```sh
npm install -S cmdsrv-sdk
```

via yarn:

```sh
yarn add cmdsrv-sdk
```

via script:

```html
<script src="./cssdk.js"></script>
```

# Usage

```js
const wsUrl = 'ws://localhost:8080/ws'
const sseUrl = 'http://localhost:8080/sse'
const cs = new CS(wsUrl)

cs.resetUrl(sseUrl) // reset endpoint url


cs.on('cs.connected', e => {}) // 连接建立时触发
cs.on('cs.closed', e => {}) // 连接关闭时触发
cs.on('cs.message', e => {}) // 收到任意数据时触发

cs.on('register', data => {}) // 监听 cmd: register 命令的响应

cs.send('register', {uid: 10001}) // 发送命令和数据消息

cs.destroy() // 销毁连接

```

# API

## 方法

### CS(url, options?)

构造函数，实例化 CS,

 - `url` [string, object]连接url，支持 websocket 和 http url，如果为 object，则视为options参数
 - `options` [object] 配置选项，可选
    + `options.url` [string] 同上 url 参数
    + `options.headers` [object] 发起 http 请求和 sse 连接时自定义的 http headers
    + `options.withCredentials` 发起http请求和sse 连接时是否携带 cookie，默认为 true，如果为 false，则 sse 将不会正常工作
    + `options.wsMsgType` ['blob', 'text'] websocket 的消息类型，blob 为二进制消息, text 为文本消息
    + `options.wsHeartBeatTime` [number] websocket 心跳时间

### cs.resetUrl(url)

重新设置连接 url，会触发重连

### cs.adapterName 

返回当前连接的适配器类型， 固定 `ws` 或者 `http`

### cs.send(command, data?) 

往连接发送命令和数据

 - command 命令名称
 - data 发送命令时携带的参数


### cs.on(event, handler)

监听事件，事件见下文

### cs.destroy()

销毁连接和实例。在实例没有销毁前，如果连接因为意外断开了，会自动重连。如果已经销毁了，则断开连接，并且不再重连。


## 事件

### `cs.connected` 

连接建立成功后触发，即使是因为断线重连时也会触发

### `cs.closed`

连接断开会触发，即使是因为意外断线也会触发

### `cs.message`

收到服务器消息时触发，无论是什么消息。

### `[command]`

监听指定 command 的消息，收到指定的 command 时触发



## License

cs is released under the 
[MIT License](http://opensource.org/licenses/mit-license.php).

(c) 2020-2021 Eyas Liu <liuyuesongde@163.com>

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
the Software, and to permit persons to whom the Software is furnished to do so,
subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.