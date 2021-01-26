package cs

import (
	"encoding/json"
)

// 内置命令
const (
	// CmdConnected on connection connected
	CmdConnected = "__cs_connected__"
	// CmdClosed on connection closed
	CmdClosed = "__cs_closed__"
	// CmdHeartbeat heartbeat message
	CmdHeartbeat = "__cs_heartbeat__"
)

// 默认消息
const (
	msgOk           = "ok"
	msgUnsupportCmd = "unsupport cmd"
)

// Request request message
type Request struct {
	Cmd     string          // message command, use for route
	Seqno   string          // seq number,the request id
	RawData json.RawMessage // request raw []byte data
}

// Response reply Request message
type Response struct {
	*Request             // reply the Request
	Cmd      string      // message command, use for route
	Seqno    string      // seq number,the request id
	Code     int         // response status code
	Msg      string      // response status message text
	Data     interface{} // response data
}

func (r *Response) fill() {
	if r.Code == 0 && r.Msg == "" {
		r.Msg = msgOk
	}
	if r.Seqno == "" {
		r.Seqno = randomString(12)
	}
}

// ServerAdapter defined integer to srv server
type ServerAdapter interface {
	// Write send response message to connect
	Write(sid string, resp *Response) error
	// Read read message form connect
	Read(*Srv) (sid string, req *Request, err error)
	// Close close specify connect
	Close(sid string) error

	// GetAllSID get server all sid
	GetAllSID() []string
}
