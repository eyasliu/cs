package xwebsocket

import (
	"sync"

	"github.com/eyasliu/cmdsrv"
	"github.com/gorilla/websocket"
)

type Conn struct {
	*websocket.Conn
	state   map[string]interface{}
	writeMu sync.Mutex
}

func (c *Conn) Send(v ...*cmdsrv.Response) error {
	if len(v) == 0 {
		return nil
	}
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	for _, msg := range v {
		if err := c.Conn.WriteJSON(msg); err != nil {
			return err
		}
	}
	return nil
}
