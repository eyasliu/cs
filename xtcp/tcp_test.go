package xtcp_test

import (
	"encoding/json"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/eyasliu/cmdsrv"
	"github.com/eyasliu/cmdsrv/xtcp"
	"github.com/gogf/gf/test/gtest"
)

var prot = xtcp.DefaultPkgProto{}

func sendToTcp(url string, r interface{}) (map[string]interface{}, error) {
	c, _, err := net.Dial(url, nil)

	if err != nil {
		return nil, err
	}
	defer c.Close()
	bt, _ := json.Marshal(r)
	pkg, _ := prot.Packer(bt)
	err = c.Send(bt)
	if err != nil {
		return nil, err
	}

	_, data, err := c.ReadMessage()
	if err != nil {
		return nil, err
	}
	fmt.Println(string(data))
	res := map[string]interface{}{}
	err = json.Unmarshal(data, &res)
	return res, err
}

func TestTcp(t *testing.T) {
	tcp := xtcp.New("127.0.0.1:5670")
	go tcp.Run()

	srv := tcp.Srv().Use(cmdsrv.AccessLogger("MYSRV"))
	gtest.C(t, func(t *gtest.T) {
		srv.Handle("register", func(c *cmdsrv.Context) {
			c.OK("login_success")
		})
		go srv.Run()
		time.Sleep(100 * time.Millisecond)

		res, err := sendToTcp("ws://127.0.0.1:5679/ws", map[string]interface{}{
			"cmd": "register",
		})

		t.Assert(err, nil)
		t.Assert(res["data"], "login_success")
	})
}
