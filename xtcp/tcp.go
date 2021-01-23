package xtcp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"

	"github.com/eyasliu/cmdsrv"
)

// TCP 适配器
type TCP struct {
	Config    *Config
	listener  net.Listener
	session   map[string]*Conn
	sessionMu sync.RWMutex
	receive   chan *reqMessage
	sidCount  uint64
}

// New 创建 TCP 适配器，必需指定地址或者配置，使用默认的私有协议解析数据包
// 默认私有协议包结构: 4byte标识数据长度 + 任意byte 数据
//
// 可使用 string 使用 tcp 网络类型指定监听的地址
// xtcp.New("127.0.0.1:8520")
//
// 使用配置对象配置tcp服务，定义网络和数据包私有协议
//
// xtcp.New(&xtcp.Config{
// 	  Addr: "127.0.0.1:8520",
// 	  Network: "tcpv6",
// 	  MsgPkg: &yourPrivateProtocol{}, //
// })
//
// xtcp.New(&xtcp.Config{
// 	  Addr: "127.0.0.1:8520",
// 	  Network: "tcpv6",
// 	  Packer: func([]byte) ([]byte, error) {},
// 	  Parser(string, []byte) ([][]byte, error),
// })
func New(v interface{}) *TCP {
	var conf *Config
	if _conf, ok := v.(*Config); ok {
		conf = _conf
	} else if addr, ok := v.(string); ok {
		conf = &Config{
			Addr:    addr,
			Network: "tcp",
		}
	}
	if conf.MsgPkg == nil {
		conf.MsgPkg = &DefaultPkgProto{}
	}

	return &TCP{
		Config:  conf,
		session: map[string]*Conn{},
		receive: make(chan *reqMessage, 50),
	}
}

// Srv 使用该适配器创建命令消息服务
func (t *TCP) Srv() (*cmdsrv.Srv, error) {
	err := t.listen()
	if err != nil {
		return nil, err
	}
	go t.accept()
	return cmdsrv.New(t), nil
}

// Read 实现 cmdsrv.ServerAdapter 接口，读取消息，每次返回一条，循环读取
func (t *TCP) Read(s *cmdsrv.Srv) (string, *cmdsrv.Request, error) {
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
	t.sessionMu.RLock()
	for sid := range t.session {
		sids = append(sids, sid)
	}
	t.sessionMu.RUnlock()
	return sids
}

// Run 启动 TCP 服务器，监听连接请求
func (t *TCP) Run() error {
	err := t.listen()
	if err != nil {
		return err
	}
	t.accept()
	return nil
}

func (t *TCP) listen() error {
	listener, err := net.Listen(t.Config.Network, t.Config.Addr)
	t.listener = listener
	return err
}

func (t *TCP) accept() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			continue
		}
		atomic.AddUint64(&t.sidCount, 1)

		sid := fmt.Sprintf("tcp.%d", t.sidCount)
		t.newConn(sid, conn)
	}
}

// 初始化 tcp 连接
func (t *TCP) newConn(sid string, netconn net.Conn) {
	conn := &Conn{
		Conn:   netconn,
		server: t,
	}
	t.sessionMu.Lock()
	t.session[sid] = conn
	t.sessionMu.Unlock()
	t.receive <- &reqMessage{
		data: &cmdsrv.Request{
			Cmd: cmdsrv.CmdConnected,
		},
		sid: sid,
	}
	for {
		_buf := make([]byte, 1024)
		buflen, err := netconn.Read(_buf)
		if err != nil {
			// data err, close socket
			t.destroyConn(sid)
			return
		}
		buf := _buf[:buflen]
		payloads, err := t.Config.MsgPkg.Parser(sid, buf)

		for _, payload := range payloads {

			if len(payload) == 0 { // heartbeat
				t.receive <- &reqMessage{data: &cmdsrv.Request{
					Cmd: cmdsrv.CmdHeartbeat,
				}, sid: sid}
				continue
			}
			r := &requestData{}
			if err = json.Unmarshal(payload, r); err != nil {
				continue
			}
			t.receive <- &reqMessage{data: &cmdsrv.Request{
				Cmd:     r.Cmd,
				Seqno:   r.Seqno,
				RawData: r.Data,
			}, sid: sid}
		}
	}
}

// 销毁指定连接
func (t *TCP) destroyConn(sid string) error {
	t.sessionMu.RLock()
	conn, ok := t.session[sid]
	t.sessionMu.RUnlock()
	if !ok {
		return errors.New("conn is already close")
	}
	err := conn.Conn.Close()
	if err != nil {
		return err
	}
	t.receive <- &reqMessage{
		data: &cmdsrv.Request{
			Cmd: cmdsrv.CmdClosed,
		},
		sid: sid,
	}
	t.sessionMu.Lock()
	delete(t.session, sid)
	t.sessionMu.Unlock()
	return nil
}
