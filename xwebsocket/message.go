package xwebsocket

import "github.com/eyasliu/cmdsrv"

type reqMessage struct {
	sid     string
	msgType int
	data    *cmdsrv.Request
}
