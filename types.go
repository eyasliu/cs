package cmdsrv

const (
	CmdConnected = "__cmdsrv_connected__" // on connection connected
	CmdClosed    = "__cmdsrv_closed__"    // on connection closed
	CmdHeartbeat = "__cmdsrv_heartbeat__" // receive heartbeat message

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

	// GetState get sid state data
	GetState(sid string, key string) interface{}
	// SetState set sid state data
	SetState(sid string, key string, v interface{})

	// GetAllSid get server all sid
	GetAllSid() []string
}
