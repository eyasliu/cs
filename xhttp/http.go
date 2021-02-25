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

	"github.com/eyasliu/cs"
)

type HTTP struct {
	srv       *cs.Srv
	receive   chan *reqMessage
	session   map[string][]*SSEConn // http 模式可能出现一个会话多个连接的情况
	sessionMu sync.RWMutex
	sidKey    string
	sidCount  uint32
	hbTime    time.Duration
	msgType   SSEMsgType
}

var defaultHeartBeatTime = 10 * time.Second

func New() *HTTP {
	h := &HTTP{
		sidKey:  "sid",
		session: make(map[string][]*SSEConn),
		receive: make(chan *reqMessage, 2),
		hbTime:  defaultHeartBeatTime,
		msgType: SSEMessage,
	}
	return h
}

func (h *HTTP) Handler(w http.ResponseWriter, req *http.Request) {
	if h.srv == nil {
		w.WriteHeader(500)
		w.Write([]byte("srv not running"))
		return
	}
	sid := h.setSid(w, req)
	if sid == "" {
		w.WriteHeader(400)
		w.Write([]byte("invalid sid, must allow cookie to store sid"))
		return
	}
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

func (h *HTTP) Srv() *cs.Srv {
	if h.srv == nil {
		h.srv = cs.New(h)
	}
	return h.srv
}

func (h *HTTP) Write(sid string, resp *cs.Response) error {
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

func (h *HTTP) Read(srv *cs.Srv) (sid string, req *cs.Request, err error) {
	h.srv = srv
	<-make(chan struct{})
	return "", nil, errors.New("HTTP Adapter unsupport Read")
}

func (h *HTTP) Close(sid string) error {
	conns, ok := h.session[sid]
	if !ok {
		return errors.New("ths sid already close")
	}
	for _, conn := range conns {
		conn.destroy(nil)
	}
	h.sessionMu.Lock()
	delete(h.session, sid)
	h.sessionMu.Unlock()
	return nil
}

func (h *HTTP) GetAllSID() []string {
	sids := make([]string, 0, len(h.session))
	h.sessionMu.RLock()
	for sid := range h.session {
		sids = append(sids, sid)
	}
	h.sessionMu.RUnlock()
	return sids
}

func (h *HTTP) setSid(w http.ResponseWriter, req *http.Request) string {
	cookie, err := req.Cookie(h.sidKey)
	var sid string
	if err != nil || cookie == nil {
		atomic.AddUint32(&h.sidCount, 1)
		// 因为sid是存cookie的，而程序每次重启，这个计数器都会重置为 0
		// 只使用计数器会导致 sid 重复，需要加上其他变量，计数器可以保证在高并发时不会重复
		sid = fmt.Sprintf("http.%d-%d", time.Now().Unix(), h.sidCount)
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
	respData := &cs.Response{}
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		respData.Msg = err.Error()
	} else {
		err := json.Unmarshal(data, reqData)
		if err != nil {
			respData.Msg = err.Error()
		}
		ctx := h.srv.NewContext(h, sid, &cs.Request{
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
	h.sessionMu.Lock()

	conns, ok := h.session[sid]
	if !ok {
		conns = []*SSEConn{conn}
	} else {
		conns = append(conns, conn)
	}
	h.session[sid] = conns
	h.sessionMu.Unlock()
	<-conn.notifyErr

	h.sessionMu.Lock()
	conns, ok = h.session[sid]
	if !ok {
		return
	}
	nextConns := make([]*SSEConn, 0, len(conns)-1)
	for _, c := range conns {
		if c != conn {
			nextConns = append(nextConns, c)
		}
	}
	if len(nextConns) > 0 {
		h.session[sid] = nextConns
	} else {
		delete(h.session, sid)
	}
	h.sessionMu.Unlock()
}
