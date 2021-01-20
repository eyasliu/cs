package xtcp_test

import (
	"encoding/json"
	"errors"
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
		srv.Use(cmdsrv.AccessLogger("MYSRV"))

		srv.Handle("register", func(c *cmdsrv.Context) {
			c.OK("login_success")
		})
		go srv.Run()
		time.Sleep(1000 * time.Millisecond)

		res, err := sendToTcp("127.0.0.1:5670", map[string]interface{}{
			"cmd": "register",
		})

		t.Assert(err, nil)
		t.Assert(res["data"], "login_success")
	})
}
