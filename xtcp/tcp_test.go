package xtcp_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/eyasliu/cs"
	"github.com/eyasliu/cs/xtcp"
	"github.com/gogf/gf/test/gtest"
)

var prot = xtcp.DefaultPkgProto{}

func sendToTcp(url string, r interface{}) (map[string]interface{}, error) {
	c, err := net.Dial("tcp", url)

	if err != nil {
		return nil, err
	}
	defer c.Close()
	bt, _ := json.Marshal(r)
	pkg, _ := prot.Packer(bt)
	_, err = c.Write(pkg)
	if err != nil {
		return nil, err
	}

	_buf := make([]byte, 1024)
	buflen, err := c.Read(_buf)
	if err != nil {
		return nil, err
	}
	data := _buf[:buflen]
	datas, err := prot.Parser("1", data)
	if err != nil {
		return nil, err
	}
	if len(datas) == 0 {
		return nil, errors.New("receive err")
	}

	fmt.Println(string(datas[0]))
	res := map[string]interface{}{}
	err = json.Unmarshal(datas[0], &res)
	return res, err
}

func TestTcp(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		srv, err := xtcp.New("127.0.0.1:5670").Srv()
		t.Assert(err, nil)
		srv.Use(srv.AccessLogger("MYSRV"))

		data := map[string]interface{}{
			"cmd":   "register",
			"seqno": "12345",
			"data":  "asdfgh",
		}
		srv.Handle(data["cmd"].(string), func(c *cs.Context) {
			c.OK(c.RawData)
		})
		go srv.Run()
		time.Sleep(100 * time.Millisecond)

		res, err := sendToTcp("127.0.0.1:5670", data)

		t.Assert(err, nil)
		t.Assert(res["cmd"], data["cmd"])
		t.Assert(res["data"], data["data"])
		t.Assert(res["seqno"], data["seqno"])
	})
}
