package cs

// SrvGroup 路由组，用于实现分组路由
type SrvGroup struct {
	parent     *SrvGroup
	srv        *Srv
	middleware []HandlerFunc
}

// Use 在当前分组添加中间件
func (s *SrvGroup) Use(handlers ...HandlerFunc) *SrvGroup {
	s.middleware = append(s.middleware, handlers...)
	return s
}

// Group 基于当前分组继续创建分组路由
func (s *SrvGroup) Group(handlers ...HandlerFunc) *SrvGroup {
	return &SrvGroup{
		parent:     s,
		srv:        s.srv,
		middleware: handlers,
	}
}

// Handle 注册路由
func (s *SrvGroup) Handle(cmd string, handlers ...HandlerFunc) *SrvGroup {
	s.srv.Handle(cmd, s.combineHandlers(handlers)...)
	return s
}

// 在注册路由前用于计算路由组的处理函数，不包含顶级的中间件
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
