package xgnet

import (
	"sync"

	"github.com/panjf2000/gnet"
)

type GNet struct {
	*gnet.EventServer
	Config    *Config
	session   map[string]gnet.Conn
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
		session: map[string]gnet.Conn{},
		receive: make(chan *reqMessage, 50),
	}
}

// func (g *GNet) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {

// }

func (g *GNet) Run() error {
	err := gnet.Serve(g, g.Config.addrURI(), gnet.WithMulticore(true))
	return err
}
