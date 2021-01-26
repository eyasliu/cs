package xgnet

import (
	"github.com/eyasliu/cs"
)

type Conn struct {
}

func (c *Conn) Send(v *cs.Response) error {
	return nil
}

func (c *Conn) destroy() error {
	return nil
}
