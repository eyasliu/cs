package xhttp

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync/atomic"

	"github.com/eyasliu/cmdsrv"
)

type HTTP struct {
	srv      *cmdsrv.Srv
	receive  chan *reqMessage
	sidKey   string
	sidCount uint64
}

func New() *HTTP {
	h := &HTTP{
		sidKey:  "sid",
		receive: make(chan *reqMessage, 2),
	}
	h.srv = cmdsrv.New(h)
	return h
}

func (h *HTTP) Handler(w http.ResponseWriter, req *http.Request) {
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

// ServeHTTP impl http.Handler to upgrade to websocket protocol
func (h *HTTP) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h.Handler(w, req)
}

func (h *HTTP) Srv() *cmdsrv.Srv {
	return h.srv
}

func (h *HTTP) Write(sid string, resp *cmdsrv.Response) error {
	return errors.New("HTTP Adapter unsupport Write")
}

func (h *HTTP) Read() (sid string, req *cmdsrv.Request, err error) {
	<-make(chan struct{})
	return "", nil, errors.New("HTTP Adapter unsupport Read")
}

func (h *HTTP) Close(sid string) error {
	return errors.New("HTTP Adapter unsupport Close")
}

func (h *HTTP) GetAllSID() []string {
	return []string{}
}
