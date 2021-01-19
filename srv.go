package cmdsrv

import "errors"

type Srv struct {
	Server     ServerAdapter
	middleware []HandlerFunc
	routes     map[string][]HandlerFunc
}

// type SrvGroup interface {
// 	Use(handlers ...HandlerFunc) SrvGroup
// 	Group(handlers ...HandlerFunc) SrvGroup
// 	Handle(cmd string, handlers ...HandlerFunc) SrvGroup
// }

func New(server ServerAdapter) *Srv {
	return &Srv{
		Server: server,
		routes: map[string][]HandlerFunc{},
	}
}

func (s *Srv) Use(handlers ...HandlerFunc) *Srv {
	s.middleware = append(s.middleware, handlers...)
	return s
}

func (s *Srv) Group(handlers ...HandlerFunc) *SrvGroup {
	srv := &SrvGroup{
		parent:     nil,
		middleware: handlers,
		srv:        s,
	}
	return srv
}

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

func (s *Srv) Push(sid string, resp *Response) error {
	return s.Server.Write(sid, resp)
}

func (s *Srv) Broadcast(resp *Response) {
	for _, sid := range s.Server.GetAllSID() {
		s.Server.Write(sid, resp)
	}
}

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

			// internal will not response
			if req.Cmd != CmdConnected &&
				req.Cmd != CmdClosed &&
				req.Cmd != CmdHeartbeat {
				defer func() {
					ctx.Push(ctx.Response)
				}()
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

			for !ctx.handlerAbort && ctx.handlerIndex < len(ctx.handlers) {
				ctx.Next()
			}
		}(sid, req)
	}
}

func (s *Srv) Run() error {
	if err := s.receive(); err != nil {
		return err
	}
	return nil
}
