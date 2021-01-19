package xwebsocket

import (
	"encoding/json"
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
		receive: make(chan *reqMessage, 50),
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

	fmt.Println("connection")
}

func (ws *WS) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ws.Handler(w, req)
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

func (ws *WS) GetAllSID() []string {
	sids := make([]string, len(ws.session))
	for sid := range ws.session {
		sids = append(sids, sid)
	}
	return sids
}

func (ws *WS) newConn(sid string, conn *websocket.Conn) {
	ws.session[sid] = &Conn{
		Conn: conn,
	}
	ws.receive <- &reqMessage{msgType: websocket.TextMessage, data: &cmdsrv.Request{
		Cmd: cmdsrv.CmdConnected,
	}, sid: sid}
	for {
		messageType, payload, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		r := &cmdsrv.Request{}
		if len(payload) == 0 { // heartbeat
			r.Cmd = cmdsrv.CmdHeartbeat
			ws.receive <- &reqMessage{msgType: messageType, data: r, sid: sid}
			continue
		}
		if err = json.Unmarshal(payload, r); err != nil {
			continue
		}
		ws.receive <- &reqMessage{msgType: messageType, data: r, sid: sid}
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
