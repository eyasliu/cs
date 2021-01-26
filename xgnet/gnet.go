package xgnet

import (
	"errors"
	"sync"

	"github.com/eyasliu/cs"
	"github.com/panjf2000/gnet"
)

type GNet struct {
	*gnet.EventServer
	Config    *Config
	session   map[string]*Conn
	sessionMu sync.RWMutex
	receive   chan *reqMessage
	sidCount  uint64
}

func New(v interface{}) *GNet {
	var conf *Config
	if _conf, ok := v.(*Config); ok {
		conf = _conf
	} else if addr, ok := v.(string); ok {
		conf = &Config{
			Addr:    addr,
			Network: "tcp",
		}
	}
	// if conf.MsgPkg == nil {
	// 	conf.MsgPkg = &DefaultPkgProto{}
	// }

	return &GNet{
		Config:  conf,
		session: map[string]*Conn{},
		receive: make(chan *reqMessage, 50),
	}
}

// func (g *GNet) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {

// }

func (g *GNet) Read(s *cs.Srv) (string, *cs.Request, error) {
	m, ok := <-g.receive
	if !ok {
		return "", nil, errors.New("websocker server is shutdown")
	}
	return m.sid, m.data, nil
}
func (g *GNet) Write(sid string, resp *cs.Response) error {
	conn, ok := g.session[sid]
	if !ok {
		return errors.New("connection is already close")
	}
	return conn.Send(resp)
}
func (g *GNet) Close(sid string) error {
	g.sessionMu.RLock()
	conn, ok := g.session[sid]
	g.sessionMu.RUnlock()
	if !ok {
		return errors.New("conn is already close")
	}
	err := conn.destroy()
	if err != nil {
		return err
	}
	g.receive <- &reqMessage{
		data: &cs.Request{
			Cmd: cs.CmdClosed,
		},
		sid: sid,
	}
	g.sessionMu.Lock()
	delete(g.session, sid)
	g.sessionMu.Unlock()
	return nil
}
func (g *GNet) GetAllSID() []string {
	sids := make([]string, len(g.session))
	g.sessionMu.RLock()
	for sid := range g.session {
		sids = append(sids, sid)
	}
	g.sessionMu.RUnlock()
	return sids
}

func (g *GNet) Srv() (*cs.Srv, error) {
	return cs.New(g), nil
}

func (g *GNet) Run() error {
	err := gnet.Serve(g, g.Config.addrURI(), gnet.WithMulticore(true))
	return err
}
