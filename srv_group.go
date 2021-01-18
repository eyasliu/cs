package cmdsrv

type SrvGroup struct {
	parent     *SrvGroup
	srv        *Srv
	middleware []HandlerFunc
}

func (s *SrvGroup) Use(handlers ...HandlerFunc) *SrvGroup {
	s.middleware = append(s.middleware, handlers...)
	return s
}
func (s *SrvGroup) Group(handlers ...HandlerFunc) *SrvGroup {
	return &SrvGroup{
		parent:     s,
		srv:        s.srv,
		middleware: handlers,
	}
}
func (s *SrvGroup) Handle(cmd string, handlers ...HandlerFunc) *SrvGroup {
	s.srv.Handle(cmd, s.combineHandlers(handlers)...)
	return s
}

func (s *SrvGroup) combineHandlers(handlers []HandlerFunc) []HandlerFunc {
	size := len(s.middleware) + len(handlers)
	if s.parent != nil {
		size += len(s.parent.middleware)
	}
	hs := make([]HandlerFunc, 0, size)
	if s.parent != nil {
		hs = append(hs, s.parent.middleware...)
	}
	hs = append(hs, s.middleware...)
	hs = append(hs, handlers...)
	return hs
}
