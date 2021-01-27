package xwebsocket

import (
	"encoding/json"
	"sync"

	"github.com/eyasliu/cs"
	"github.com/gorilla/websocket"
)

// Conn websocket 连接对象
type Conn struct {
	*websocket.Conn
	writeMu sync.Mutex
	msgType int
}

type reqMessage struct {
	sid     string
	msgType int
	data    *cs.Request
}

type requestData struct {
	Cmd   string          `json:"cmd"`   // message command, use for route
	Seqno string          `json:"seqno"` // seq number,the request id
	Data  json.RawMessage `json:"data"`  // response data
}

type responseData struct {
	Cmd   string      `json:"cmd"`   // message command, use for route
	Seqno string      `json:"seqno"` // seq number,the request id
	Code  int         `json:"code"`  // response status code
	Msg   string      `json:"msg"`   // response status message text
	Data  interface{} `json:"data"`  // response data
}

// Send 往连接推送消息，线程安全
func (c *Conn) Send(v ...*cs.Response) error {
	if len(v) == 0 {
		return nil
	}
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	for _, msg := range v {
		resp := &responseData{
			Cmd:   msg.Cmd,
			Seqno: msg.Seqno,
			Code:  msg.Code,
			Msg:   msg.Msg,
			Data:  msg.Data,
		}

		j, err := json.Marshal(resp)
		if err != nil {
			return err
		}

		if err := c.Conn.WriteMessage(c.msgType, json.RawMessage(j)); err != nil {
			return err
		}
	}
	return nil
}
