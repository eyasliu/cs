package cmdsrv

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
	Cmd     string // message command, use for route
	Seqno   string // seq number,the request id
	RawData []byte // request raw []byte data
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
