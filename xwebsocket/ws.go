package xwebsocket

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/eyasliu/cmdsrv"

	"github.com/gorilla/websocket"
)

type WS struct {
	upgrader websocket.Upgrader
	session  map[string]*Conn
	receive  chan *reqMessage
	sidCount uint64
}

var _ cmdsrv.ServerAdapter = &WS{}

func New() *WS {
	return &WS{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		session: make(map[string]*Conn),
	}
}

// Handler impl http.HandlerFunc to upgrade to websocket protocol
func (ws *WS) Handler(w http.ResponseWriter, req *http.Request) {
	conn, err := ws.upgrader.Upgrade(w, req, nil)
	if err != nil {
		return
	}
	ws.sidCount++

	sid := fmt.Sprintf("%d", ws.sidCount)

	defer ws.destroyConn(sid)
	ws.newConn(sid, conn)
}

func (ws *WS) Read() (string, *cmdsrv.Request, error) {
	m, ok := <-ws.receive
	if !ok {
		return "", nil, errors.New("websocker server is shutdown")
	}
	return m.sid, m.data, nil
}

func (ws *WS) Write(sid string, resp *cmdsrv.Response) error {
	conn, ok := ws.session[sid]
	if !ok {
		return errors.New("connection is already close")
	}
	return conn.Send(resp)
}

func (ws *WS) Close(sid string) error {
	return ws.destroyConn(sid)
}

func (ws *WS) GetState(sid, key string) interface{} {
	conn, ok := ws.session[sid]
	if !ok {
		return nil
	}
	data := conn.state[key]
	return data
}

func (ws *WS) SetState(sid, key string, v interface{}) {
	conn, ok := ws.session[sid]
	if !ok {
		return
	}
	if conn.state == nil {
		conn.state = make(map[string]interface{})
	}
	conn.state[key] = v
}

func (ws *WS) newConn(sid string, conn *websocket.Conn) {
	ws.session[sid] = &Conn{
		Conn:  conn,
		state: map[string]interface{}{},
	}
	ws.receive <- &reqMessage{msgType: websocket.TextMessage, data: &cmdsrv.Request{
		Cmd: cmdsrv.CmdConnected,
	}, sid: sid}
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Println(err)
			return
		}
	}
}

func (ws *WS) destroyConn(sid string) error {
	conn, ok := ws.session[sid]
	if !ok {
		return errors.New("conn is already close")
	}
	err := conn.Close()
	if err != nil {
		return err
	}
	ws.receive <- &reqMessage{msgType: websocket.TextMessage, data: &cmdsrv.Request{
		Cmd: cmdsrv.CmdClosed,
	}, sid: sid}
	delete(ws.session, sid)
	return nil
}
