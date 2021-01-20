package cmdsrv

import (
	"errors"

	"github.com/gogf/gf/os/gcache"
)

// Srv 基于命令的消息处理框架
type Srv struct {
	Server     ServerAdapter            // 服务器适配器
	middleware []HandlerFunc            // 全局路由中间件
	routes     map[string][]HandlerFunc // 路由的处理函数
	state      *State                   // SID 会话的状态数据
}

// New 指定服务器实例化一个消息服务
func New(server ServerAdapter) *Srv {
	return &Srv{
		Server: server,
		routes: map[string][]HandlerFunc{},
		state:  &State{cache: gcache.New()},
	}
}

// Use 增加一个全局中间件
func (s *Srv) Use(handlers ...HandlerFunc) *Srv {
	s.middleware = append(s.middleware, handlers...)
	return s
}

// Group 路由分组，指定该分组下的中间件
func (s *Srv) Group(handlers ...HandlerFunc) *SrvGroup {
	srv := &SrvGroup{
		parent:     nil,
		middleware: handlers,
		srv:        s,
	}
	return srv
}

// Handle 注册路由，cmd 是命令， handlers 是该路由的处理函数
func (s *Srv) Handle(cmd string, handlers ...HandlerFunc) *Srv {
	srv := s
	if len(handlers) == 0 {
		return s
	}
	hs, ok := srv.routes[cmd]
	if ok {
		hs = append(hs, handlers...)
	} else {
		hs = handlers
	}
	srv.routes[cmd] = hs
	return s
}

// Push 往指定的会话 SID 连接推送消息
func (s *Srv) Push(sid string, resp *Response) error {
	return s.Server.Write(sid, resp)
}

// Broadcast 往所有可用的会话推送消息
func (s *Srv) Broadcast(resp *Response) {
	for _, sid := range s.Server.GetAllSID() {
		s.Server.Write(sid, resp)
	}
}

// Close 关闭指定会话 SID 的连接
func (s *Srv) Close(sid string) error {
	return s.Server.Close(sid)
}

// GetState 获取指定会话的指定状态值
func (s *Srv) GetState(sid, key string) interface{} {
	return s.state.Get(sid, key)
}

// SetState 设置指定连接的状态
func (s *Srv) SetState(sid, key string, val interface{}) {
	s.state.Set(sid, key, val)
}

func (s *Srv) NewContext(sid string, req *Request) *Context {
	ctx := &Context{
		Response: &Response{
			Request: req,
			Cmd:     req.Cmd,
			Seqno:   req.Seqno,
			Code:    -1,
			Msg:     msgUnsupportCmd,
			Data:    struct{}{},
		},
		SID:    sid,
		Srv:    s,
		Server: s.Server,
	}

	// call internal hooks
	switch req.Cmd {
	case CmdConnected:
		s.onSidConnected(sid)
	case CmdClosed:
		s.onSidClosed(sid)
	}

	routeHandlers, ok := s.routes[req.Cmd]
	var handlers []HandlerFunc
	if ok {
		handlers = make([]HandlerFunc, 0, len(s.middleware)+len(routeHandlers))
		handlers = append(handlers, s.middleware...)
		handlers = append(handlers, routeHandlers...)
	} else {
		handlers = make([]HandlerFunc, 0, len(s.middleware)+1)
		handlers = append(handlers, s.middleware...)
		handlers = append(handlers, RouteNotFound)
	}
	ctx.handlers = handlers
	ctx.handlerIndex = -1
	return ctx
}

func (s *Srv) CallContext(ctx *Context) {
	for !ctx.handlerAbort && ctx.handlerIndex < len(ctx.handlers) {
		ctx.Next()
	}
}

// 当有新的会话SID产生时触发，依赖内置命令 CmdConnected 实现
func (s *Srv) onSidConnected(sid string) {}

// 当有会话SID关闭时触发，依赖内置命令 CmdClosed 实现
func (s *Srv) onSidClosed(sid string) {
	s.state.destroySid(sid)
}

// 接收服务器适配器产生的消息，并执行路由处理函数
func (s *Srv) receive() error {
	for {
		sid, req, err := s.Server.Read()
		if err != nil {
			return err
		}
		if req == nil {
			return errors.New("unexpected request data")
		}

		// handler cmd
		go func(sid string, req *Request) {
			ctx := s.NewContext(sid, req)

			// internal will not response
			if req.Cmd != CmdConnected &&
				req.Cmd != CmdClosed &&
				req.Cmd != CmdHeartbeat {
				defer func() {
					ctx.Push(ctx.Response)
				}()
			}
			s.CallContext(ctx)
		}(sid, req)
	}
}

// Run 开始接收命令消息，运行框架，会阻塞当前 goroutine
func (s *Srv) Run() error {
	return s.receive()
	// if err := s.receive(); err != nil {
	// 	return err
	// }
	// return nil
}
