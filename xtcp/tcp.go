package xtcp

import (
	"encoding/json"
	"errors"
	"log"
	"net"

	"github.com/eyasliu/cmdsrv"
)

// TCP 适配器
type TCP struct {
	Config   *Config
	session  map[string]*Conn
	receive  chan *reqMessage
	sidCount uint64
}

// New 创建 TCP 适配器
func New(addr string) *TCP {
	conf := &Config{
		Addr:    addr,
		Network: "tcp",
		MsgPkg:  defaultPkg,
	}
	return &TCP{
		Config:  conf,
		session: map[string]*Conn{},
		receive: make(chan *reqMessage, 50),
	}
}

// Srv 使用该适配器创建命令消息服务
func (t *TCP) Srv() *cmdsrv.Srv {
	return cmdsrv.New(t)
}

// Read 实现 cmdsrv.ServerAdapter 接口，读取消息，每次返回一条，循环读取
func (t *TCP) Read() (string, *cmdsrv.Request, error) {
	m, ok := <-t.receive
	if !ok {
		return "", nil, errors.New("websocker server is shutdown")
	}
	return m.sid, m.data, nil
}

// Write 实现 cmdsrv.ServerAdapter 接口，给连接推送消息
func (t *TCP) Write(sid string, resp *cmdsrv.Response) error {
	conn, ok := t.session[sid]
	if !ok {
		return errors.New("connection is already close")
	}
	return conn.Send(resp)
}

// Close 实现 cmdsrv.ServerAdapter 接口，关闭指定连接
func (t *TCP) Close(sid string) error {
	return t.destroyConn(sid)
}

// GetAllSID 实现 cmdsrv.ServerAdapter 接口，获取当前服务所有SID，用于遍历连接
func (t *TCP) GetAllSID() []string {
	sids := make([]string, len(t.session))
	for sid := range t.session {
		sids = append(sids, sid)
	}
	return sids
}

// 初始化 tcp 连接
func (t *TCP) newConn(sid string, conn net.Conn) {
	t.session[sid] = &Conn{
		Conn:   conn,
		server: t,
	}
	t.receive <- &reqMessage{
		data: &cmdsrv.Request{
			Cmd: cmdsrv.CmdConnected,
		},
		sid: sid,
	}
	for {
		messageType, payload, err := conn.Read()
		if err != nil {
			log.Println(err)
			return
		}
		r := &cmdsrv.Request{}
		if len(payload) == 0 { // heartbeat
			r.Cmd = cmdsrv.CmdHeartbeat
			ws.receive <- &reqMessage{data: r, sid: sid}
			continue
		}
		if err = json.Unmarshal(payload, r); err != nil {
			continue
		}
		ws.receive <- &reqMessage{data: r, sid: sid}
	}
}

// 销毁指定连接
func (t *TCP) destroyConn(sid string) error {
	conn, ok := t.session[sid]
	if !ok {
		return errors.New("conn is already close")
	}
	err := conn.Close()
	if err != nil {
		return err
	}
	t.receive <- &reqMessage{
		data: &cmdsrv.Request{
			Cmd: cmdsrv.CmdClosed,
		},
		sid: sid,
	}
	delete(t.session, sid)
	return nil
}
