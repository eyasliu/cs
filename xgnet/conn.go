package xgnet

import (
	"github.com/eyasliu/cmdsrv"
)

type Conn struct {
}

func (c *Conn) Send(v *cmdsrv.Response) error {
	return nil
}

func (c *Conn) destroy() error {
	return nil
}
