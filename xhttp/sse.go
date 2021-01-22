package xhttp

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/eyasliu/cmdsrv"
)

type SSEMsgType byte

const (
	SSEEvent   = 0
	SSEMessage = 1
)

type SSEConn struct {
	w       http.ResponseWriter
	msgType SSEMsgType
	hbTime  time.Duration
	isClose bool
}

func newSSEConn(w http.ResponseWriter, msgType SSEMsgType, heartbeatTime time.Duration) (*SSEConn, error) {
	s := &SSEConn{
		w:       w,
		msgType: msgType,
		hbTime:  heartbeatTime,
	}
	err := s.init()
	return s, err
}

func (s *SSEConn) init() error {
	// retry
	_, err := s.w.Write([]byte("retry: 10000\n\n"))
	if err != nil {
		return err
	}

	// heartbeat
	go func(s *SSEConn) {
		for {
			time.Sleep(s.hbTime / 2)
			if s.isClose {
				break
			}
			_, err := s.w.Write([]byte(": heartbeat\n\n"))
			if err != nil {
				s.destroy()
				break
			}
		}
	}(s)
	return nil
}

func (s *SSEConn) Send(v ...*cmdsrv.Response) error {
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
			dataBt, err := json.Marshal(resp.Data)
			if err != nil {
				return err
			}
			msg += "data: " + string(dataBt) + "\n"
		} else {
			return errors.New("unsupport sse message type")
		}
		msg += "\n\n"
		_, err := s.w.Write([]byte(msg))
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SSEConn) destroy() {
	s.isClose = true
}
