package xhttp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/eyasliu/cs"
)

type SSEMsgType byte

const (
	SSEEvent   = 0
	SSEMessage = 1
)

type SSEConn struct {
	w         http.ResponseWriter
	flusher   http.Flusher
	msgType   SSEMsgType
	hbTime    time.Duration
	isClose   bool
	notifyErr chan error
}

func newSSEConn(w http.ResponseWriter, msgType SSEMsgType, heartbeatTime time.Duration) (*SSEConn, error) {
	s := &SSEConn{
		w:         w,
		msgType:   msgType,
		hbTime:    heartbeatTime,
		notifyErr: make(chan error),
	}
	flusher, ok := s.w.(http.Flusher)

	if !ok {
		return nil, errors.New("Streaming unsupported!")
	}
	s.flusher = flusher
	err := s.init()
	return s, err
}

func (s *SSEConn) init() error {

	s.w.Header().Set("Content-Type", "text/event-stream")
	s.w.Header().Set("Cache-Control", "no-cache")
	s.w.Header().Set("Connection", "keep-alive")
	// s.w.Header().Del("Content-Length")
	// retry
	_, err := fmt.Fprint(s.w, "retry: 10000\n\n")
	// _, err := writer(s.w, "retry: 10000\n\n")
	if err != nil {
		return err
	}
	s.flusher.Flush()

	// heartbeat
	go func(s *SSEConn) {
		for {
			time.Sleep(s.hbTime / 2)
			if s.isClose {
				break
			}
			_, err := fmt.Fprint(s.w, ": heartbeat\n\n")
			if err != nil {
				s.destroy(err)
				break
			}
			s.flusher.Flush()
		}
	}(s)
	return nil
}

func (s *SSEConn) Send(v ...*cs.Response) error {
	if s.w == nil {
		err := errors.New("connection is already closed")
		s.destroy(err)
		return err
	}
	for _, resp := range v {
		msg := ""
		if s.msgType == SSEEvent {
			if resp.Seqno != "" {
				msg += "id: " + resp.Seqno + "\n"
			}
			msg += "event: " + resp.Cmd + "\n"
			if resp.Data != nil {
				dataBt, err := json.Marshal(resp.Data)
				if err != nil {
					return err
				}
				msg += "data: " + string(dataBt) + "\n"
			}
		} else if s.msgType == SSEMessage {
			resp1 := &responseData{
				Cmd:   resp.Cmd,
				Seqno: resp.Seqno,
				Code:  resp.Code,
				Msg:   resp.Msg,
				Data:  resp.Data,
			}
			dataBt, err := json.Marshal(resp1)
			if err != nil {
				return err
			}
			msg += "data: " + string(dataBt) + "\n"
		} else {
			return errors.New("unsupport sse message type")
		}
		msg += "\n\n"
		_, err := fmt.Fprint(s.w, msg)
		s.flusher.Flush()
		if err != nil {
			s.destroy(err)
			return err
		}
	}
	return nil
}

func (s *SSEConn) destroy(err error) {
	s.notifyErr <- err
	s.isClose = true
}
