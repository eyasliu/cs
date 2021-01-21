package cmdsrv

import (
	"errors"
	"sync"

	"github.com/gogf/gf/os/gcache"
)

// Srv 基于命令的消息处理框架
type Srv struct {
	Server     []ServerAdapter          // 服务器适配器
	isRunning  bool                     // 服务是否已经正在运行
	runErr     chan error               // 服务运行错误通知
	middleware []HandlerFunc            // 全局路由中间件
	routes     map[string][]HandlerFunc // 路由的处理函数
	state      *State                   // SID 会话的状态数据
}

// New 指定服务器实例化一个消息服务
func New(server ...ServerAdapter) *Srv {
	return &Srv{
		Server: server,
		runErr: make(chan error, 0),
		routes: map[string][]HandlerFunc{},
		state:  &State{cache: gcache.New()},
	}
}

var serverMu sync.RWMutex

// AddServer 增加服务适配器
func (s *Srv) AddServer(server ...ServerAdapter) *Srv {
	serverMu.Lock()
	s.Server = append(s.Server, server...)

	// 如果服务已经正在 running 了，增加的时候自动启动
	if s.isRunning {
		for _, ser := range server {
			go s.startServer(ser)
		}
	}
	serverMu.Unlock()
	return s
}

// Use 增加全局中间件
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

var routeMu sync.Mutex

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
	routeMu.Lock()
	srv.routes[cmd] = hs
	routeMu.Unlock()
	return s
}

// Push 往指定的会话 SID 连接推送消息
func (s *Srv) Push(sid string, resp *Response) error {
	server, err := s.getSidServer(sid)
	if err != nil {
		return err
	}
	return s.PushServer(server, sid, resp)
}

// PushServer 往指定适配器的 sid 推送消息
func (s *Srv) PushServer(server ServerAdapter, sid string, resp *Response) error {
	return server.Write(sid, resp)
}

// Broadcast 往所有可用的会话推送消息
func (s *Srv) Broadcast(resp *Response) {
	for _, server := range s.Server {
		for _, sid := range s.GetAllSID() {
			server.Write(sid, resp)
		}
	}
}

// Close 关闭指定会话 SID 的连接
func (s *Srv) Close(sid string) error {
	server, err := s.getSidServer(sid)
	if err != nil {
		return errors.New("the sid is already close")
	}
	return s.CloseByServer(server, sid)
}

// CloseByServer close specify server and specify sid
func (s *Srv) CloseByServer(server ServerAdapter, sid string) error {
	return server.Close(sid)
}

// GetState 获取指定会话的指定状态值
func (s *Srv) GetState(sid, key string) interface{} {
	return s.state.Get(sid, key)
}

// SetState 设置指定连接的状态
func (s *Srv) SetState(sid, key string, val interface{}) {
	s.state.Set(sid, key, val)
}

// NewContext 根据请求消息实例化上下文
// 应该在实现 adapter 时才有用
func (s *Srv) NewContext(server ServerAdapter, sid string, req *Request) *Context {
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
		Server: server,
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

// CallContext 调用上下文，触发上下文中间件
// 应该在实现 adapter 时才有用
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
func (s *Srv) startServer(server ServerAdapter) {
	for {
		sid, req, err := server.Read()
		if err != nil {
			s.runErr <- err
			return
		}
		if req == nil {
			s.runErr <- errors.New("unexpected request data")
			return
		}

		// handler cmd
		go func(sid string, req *Request) {
			ctx := s.NewContext(server, sid, req)

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

// GetAllSID 获取所有适配器的 SID
func (s *Srv) GetAllSID() []string {
	sids := []string{}
	for _, server := range s.Server {
		sids = append(sids, server.GetAllSID()...)
	}
	return sids
}

func (s *Srv) getSidServer(sid string) (ServerAdapter, error) {
	for _, server := range s.Server {
		for _, id := range server.GetAllSID() {
			if id == sid {
				return server, nil
			}
		}
	}
	return nil, errors.New("the sid is destroy")
}

// Run 开始接收命令消息，运行框架，会阻塞当前 goroutine
func (s *Srv) Run() error {
	serverMu.Lock()
	s.isRunning = true
	for _, server := range s.Server {
		go s.startServer(server)
	}
	serverMu.Unlock()
	err := <-s.runErr
	return err
	// if err := s.receive(); err != nil {
	// 	return err
	// }
	// return nil
}
