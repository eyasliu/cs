package cmdsrv

import (
	"fmt"
	"reflect"

	"github.com/gogf/gf/encoding/gjson"
	"github.com/gogf/gf/util/gvalid"
)

// HandlerFunc 消息处理函数，中间件和路由的函数签名
type HandlerFunc = func(*Context)

type PushHandlerFunc = func(*Context) error

// Context 处理函数的的上下文对象，中间件和路由的函数参数
type Context struct {
	*Response
	SID          string
	Srv          *Srv
	Server       ServerAdapter
	handlers     []HandlerFunc
	handlerIndex int
	handlerAbort bool
}

// Next 调用下一层中间件。
// 中间件的调用是按照洋葱模型调用，该方法应该只用在中间件函数使用
func (c *Context) Next() {
	if c.handlerAbort {
		return
	}
	if c.handlerIndex < len(c.handlers)-1 {
		c.handlerIndex++
		c.handlers[c.handlerIndex](c)
	} else {
		c.Abort()
	}
}

// Abort 中断更里层的中间件调用
// 该方法应该只用在中间件函数使用
func (c *Context) Abort() {
	c.handlerAbort = true
}

// Parse 解析并验证消息携带的参数，参考 Goframe 的 请求输入-对象处理 https://itician.org/pages/viewpage.action?pageId=1114185
// 支持 json 和 xml 数据流
// 支持将数据解析为 *struct/**struct/*[]struct/*[]*struct/*map/*[]map
// 如果目标值是 *struct/**struct/*[]struct/*[]*struct ，则会自动调用请求验证，参考 GoFrame 的 请求输入-请求校验 https://itician.org/pages/viewpage.action?pageId=1114244
func (c *Context) Parse(pointer interface{}, mapping ...map[string]string) error {
	var (
		rv = reflect.ValueOf(pointer)
		rk = rv.Kind()
	)
	if rk != reflect.Ptr {
		return fmt.Errorf(
			"parameter should be type of *struct/**struct/*[]struct/*[]*struct/*map[string]interface{}, but got: %v",
			rk,
		)
	}
	rv = rv.Elem()
	rk = rv.Kind()

	// parse request data
	data, err := gjson.LoadContent(c.Response.Request.RawData)
	if err != nil {
		return err
	}

	// if struct or array/slice, validate pointer
	switch rk {
	case reflect.Ptr, reflect.Struct:
		if err := data.GetStruct(".", pointer, mapping...); err != nil {
			return err
		}
		if err := gvalid.CheckStruct(pointer, nil); err != nil {
			return err
		}
	case reflect.Array, reflect.Slice:
		if err := data.GetStructs(".", pointer, mapping...); err != nil {
			return err
		}
		for i := 0; i < rv.Len(); i++ {
			if err := gvalid.CheckStruct(rv.Index(i), nil); err != nil {
				return err
			}
		}
	case reflect.Map:
		if err := data.ToMapToMapDeep(pointer); err != nil {
			return err
		}
	}

	return nil
}

// Get 获取当前会话的状态，
// 注意：这是会话的状态，而不是当前请求函数的状态（和HTTP那边不一样）
func (c *Context) Get(key string) interface{} {
	return c.Srv.state.Get(c.SID, key)
}

// Set 设置当前会话的状态
func (c *Context) Set(key string, v interface{}) {
	c.Srv.state.Set(c.SID, key, v)
}

// GetState 获取指定 sid 和 key 的状态值
func (c *Context) GetState(sid, key string) interface{} {
	return c.Srv.state.Get(sid, key)
}

// SetState 设置当前会话的状态
func (c *Context) SetState(sid, key string, v interface{}) {
	c.Srv.state.Set(sid, key, v)
}

// Err 响应错误，如果错误对象为空则忽略不处理
func (c *Context) Err(err error, code int) {
	if err != nil {
		c.Response.Code = code
		c.Response.Msg = err.Error()
	}
}

// OK 响应成功，参数是指定响应的 data 数据，如果不设置则默认为空对象
// c.OK() 响应 {}
// c.OK(nill) 响应 null
// c.OK(map[string]interface{}{"x": 1}) 响应 {"x": 1}
func (c *Context) OK(data ...interface{}) {
	c.Response.Code = 0
	c.Response.Msg = msgOk
	if len(data) == 0 {
		c.Response.Data = struct{}{}
	} else {
		c.Response.Data = data[0]
	}
}

// Resp 设置响应
func (c *Context) Resp(code int, msg string, data ...interface{}) {
	c.Response.Code = code
	c.Response.Msg = msg
	if len(data) > 0 {
		c.Response.Data = data[0]
	}
}

// Push 往当前会话推送消息
func (c *Context) Push(data *Response) error {
	ctx, err := c.Srv.callPushMiddleware(c, data)
	if err != nil {
		return err
	}
	// data.fill()
	// logger := c.Get(loggerCtxKey)
	// if logger != nil {
	// 	name, ok := c.Get(loggerNameCtxKey).(string)
	// 	if !ok {
	// 		name = "SRV"
	// 	}
	// 	logger.(printLogger).Debug(fmt.Sprintf("%s PUSH SID=%s CMD=%s SEQ=%s Code=%d Msg=%s %s",
	// 		name, c.SID, data.Cmd, data.Seqno, data.Code, data.Msg, getLogDataString(data.Data)))
	// }
	return c.Srv.PushServer(c.Server, c.SID, ctx.Response)
}

// PushSID 往指定SID会话推送消息
func (c *Context) PushSID(sid string, data *Response) error {
	ctx, err := c.Srv.callPushMiddleware(c, data)
	if err != nil {
		return err
	}
	// data.fill()
	// logger := c.Get(loggerCtxKey)
	// if logger != nil {
	// 	name, ok := c.Get(loggerNameCtxKey).(string)
	// 	if !ok {
	// 		name = "SRV"
	// 	}
	// 	logger.(printLogger).Debug(fmt.Sprintf("%s PUSH SID=%s CMD=%s SEQ=%s Code=%d Msg=%s %s",
	// 		name, sid, data.Cmd, data.Seqno, data.Code, data.Msg, getLogDataString(data.Data)))
	// }
	return c.Srv.Push(sid, ctx.Response)
}

// Close 关闭当前会话连接
func (c *Context) Close() error {
	return c.Srv.CloseWithServer(c.Server, c.SID)
}

// GetAllSID 获取目前生效的所有会话ID
func (c *Context) GetAllSID() []string {
	return c.Srv.GetAllSID()
}

// GetServerAllSID 获取当前适配器中生效的所有会话ID
func (c *Context) GetServerAllSID() []string {
	return c.Server.GetAllSID()
}

// Broadcast 广播消息，即给所有有效的会话推送消息
func (c *Context) Broadcast(data *Response) {
	// c.Srv.Broadcast(data)
	for _, server := range c.Srv.Server {
		for _, sid := range server.GetAllSID() {
			ctx, err := c.Srv.callPushMiddleware(c, data)
			if err == nil {
				go server.Write(sid, ctx.Response)
			}
		}
	}
}

func (c *Context) clone() *Context {
	return &Context{
		Response: &Response{
			Request: &Request{
				Cmd:     c.Request.Cmd,
				Seqno:   c.Request.Seqno,
				RawData: c.Request.RawData,
			},
			Cmd:   c.Cmd,
			Seqno: c.Seqno,
			Code:  c.Code,
			Msg:   c.Msg,
			Data:  struct{}{},
		},
		SID:          c.SID,
		Srv:          c.Srv,
		Server:       c.Server,
		handlers:     nil,
		handlerIndex: -1,
	}
}

// RouteNotFound 当路由没匹配到时的默认处理函数
func RouteNotFound(c *Context) {
	c.Resp(-1, msgUnsupportCmd)
}
