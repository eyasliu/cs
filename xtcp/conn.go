package xtcp

import (
	"encoding/json"
	"net"
	"sync"

	"github.com/eyasliu/cmdsrv"
)

// Conn tcp 连接对象
type Conn struct {
	Conn    net.Conn
	sid     string
	server  *TCP
	writeMu sync.Mutex
}

// Send 往连接推送消息，线程安全
func (c *Conn) Send(v ...*cmdsrv.Response) error {
	if len(v) == 0 {
		return nil
	}
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	for _, msg := range v {
		j := &responseData{
			Cmd:   msg.Cmd,
			Seqno: msg.Seqno,
			Code:  msg.Code,
			Msg:   msg.Msg,
			Data:  msg.Data,
		}
		bt, err := json.Marshal(j)
		if err != nil {
			return err
		}
		pkg, err := c.server.Config.Packer(bt)
		if err != nil {
			return err
		}
		if _, err := c.Conn.Write(pkg); err != nil {
			return err
		}
	}
	return nil
}

// func (c *Conn) Read() ([]*cmdsrv.Response, error) {
// 	_buf := make([]byte, 1024)
// 	buflen, err := conn.Conn.Read(_buf)
// 	if err != nil {
// 		return err
// 	}
// 	buf := _buf[:buflen]
// 	body, err := c.server.Config.Parser(c.sid, buf)
// 	if err != nil {
// 		// 解析异常，断开连接
// 		conn.Destroy()
// 		break
// 	}
// 	for _, body := range body {
// 		msg := &Message{
// 			Data: body,
// 			Conn: conn,
// 		}
// 		if conn.IsServer() {
// 			conn.server.recChan <- msg
// 		} else if conn.IsClient() {
// 			conn.client.recChan <- msg
// 		}
// 	}
// }
