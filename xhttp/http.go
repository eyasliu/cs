package xhttp

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/eyasliu/cmdsrv"
)

type HTTP struct {
	srv       *cmdsrv.Srv
	receive   chan *reqMessage
	session   map[string][]*SSEConn // http 模式可能出现一个会话多个连接的情况
	sessionMu sync.RWMutex
	sidKey    string
	sidCount  uint64
	hbTime    time.Duration
	msgType   SSEMsgType
}

var defaultHeartBeatTime = 10 * time.Second

func New() *HTTP {
	h := &HTTP{
		sidKey:  "sid",
		receive: make(chan *reqMessage, 2),
		hbTime:  defaultHeartBeatTime,
		msgType: SSEEvent,
	}
	h.srv = cmdsrv.New(h)
	return h
}

func (h *HTTP) Handler(w http.ResponseWriter, req *http.Request) {
	sid := h.setSid(w, req)
	if req.Method == "GET" {
		h.invokeSSE(sid, w, req)
	} else if req.Method == "POST" || req.Method == "PUT" || req.Method == "DELETE" {
		h.invokeHandle(sid, w, req)
	}
}

// ServeHTTP impl http.Handler to upgrade to websocket protocol
func (h *HTTP) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h.Handler(w, req)
}

func (h *HTTP) Srv() *cmdsrv.Srv {
	return h.srv
}

func (h *HTTP) Write(sid string, resp *cmdsrv.Response) error {
	conns, ok := h.session[sid]
	if !ok {
		return errors.New("connection is already close")
	}
	for _, conn := range conns {
		if err := conn.Send(resp); err != nil {
			return err
		}
	}
	return nil
}

func (h *HTTP) Read() (sid string, req *cmdsrv.Request, err error) {
	<-make(chan struct{})
	return "", nil, errors.New("HTTP Adapter unsupport Read")
}

func (h *HTTP) Close(sid string) error {
	conns, ok := h.session[sid]
	if !ok {
		return errors.New("ths sid already close")
	}
	for _, conn := range conns {
		conn.destroy()
	}
	h.sessionMu.Lock()
	delete(h.session, sid)
	h.sessionMu.Unlock()
	return nil
}

func (h *HTTP) GetAllSID() []string {
	return []string{}
}

func (h *HTTP) setSid(w http.ResponseWriter, req *http.Request) string {
	cookie, err := req.Cookie(h.sidKey)
	var sid string
	if err != nil || cookie == nil {
		atomic.AddUint64(&h.sidCount, 1)
		sid := fmt.Sprintf("http.%d", h.sidCount)
		cookie = &http.Cookie{
			Name:     h.sidKey,
			Value:    sid,
			HttpOnly: true,
			// Expires: time.Now().Add(24 * time.Hour)
		}
		http.SetCookie(w, cookie)
	} else {
		sid = cookie.Value
	}
	return sid
}

func (h *HTTP) invokeHandle(sid string, w http.ResponseWriter, req *http.Request) {
	reqData := &requestData{}
	respData := &cmdsrv.Response{}
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		respData.Msg = err.Error()
	} else {
		err := json.Unmarshal(data, reqData)
		if err != nil {
			respData.Msg = err.Error()
		}
		ctx := h.srv.NewContext(h, sid, &cmdsrv.Request{
			Cmd:     reqData.Cmd,
			Seqno:   reqData.Seqno,
			RawData: reqData.Data,
		})
		h.srv.CallContext(ctx)
		respData = ctx.Response
	}

	resp := &responseData{
		Cmd:   respData.Cmd,
		Seqno: respData.Seqno,
		Code:  respData.Code,
		Msg:   respData.Msg,
		Data:  respData.Data,
	}
	respBt, _ := json.Marshal(resp)

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write(respBt)
}

func (h *HTTP) invokeSSE(sid string, w http.ResponseWriter, req *http.Request) {
	conn, err := newSSEConn(w, h.msgType, h.hbTime)
	if err != nil {
		return
	}
	header := w.Header()
	header.Set("Content-Type", "text/event-stream")
	header.Set("Cache-Control", "no-cache")
	header.Set("Connection", "keep-alive")
	h.sessionMu.Lock()
	conns, ok := h.session[sid]
	if !ok {
		conns = []*SSEConn{conn}
	} else {
		conns = append(conns, conn)
	}
	h.session[sid] = conns
	h.sessionMu.Unlock()
}
