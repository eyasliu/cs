package cmdsrv

import (
	"encoding/json"
	"fmt"
)

// 内置命令
const (
	// CmdConnected on connection connected
	CmdConnected = "__cmdsrv_connected__"
	// CmdClosed on connection closed
	CmdClosed = "__cmdsrv_closed__"
	// CmdHeartbeat heartbeat message
	CmdHeartbeat = "__cmdsrv_heartbeat__"
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

func (r *Response) MarshalJSON() ([]byte, error) {
	var data json.RawMessage
	var err error
	switch r.Data.(type) {
	case string:
		data = json.RawMessage(r.Data.(string))
	case []byte:
		data = json.RawMessage(r.Data.([]byte))
	default:
		data, err = json.Marshal(r.Data)
		if err != nil {
			return nil, err
		}
	}
	s := fmt.Sprintf(`{"cmd":"%s","seqno":"%s","code":%d,"msg":"%s","data":%s}`,
		r.Cmd,
		r.Seqno,
		r.Code,
		r.Msg,
		string(data),
	)
	return []byte(s), nil
}

// ServerAdapter defined integer to srv server
type ServerAdapter interface {
	// Write send response message to connect
	Write(sid string, resp *Response) error
	// Read read message form connect
	Read() (sid string, req *Request, err error)
	// Close close specify connect
	Close(sid string) error

	// GetAllSID get server all sid
	GetAllSID() []string
}
