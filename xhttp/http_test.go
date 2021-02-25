// +build !race test

package xhttp_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/eyasliu/cs"
	"github.com/eyasliu/cs/xhttp"
	"github.com/gogf/gf/test/gtest"
)

func sendToHttp(url string, r interface{}) (map[string]interface{}, error) {
	bt, _ := json.Marshal(r)
	client := http.DefaultClient
	resp, err := client.Post(url, "application/json", bytes.NewReader(bt))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode > 300 {

		return nil, fmt.Errorf("http status code fail: %d", resp.StatusCode)
	}
	resBt, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	res := map[string]interface{}{}
	err = json.Unmarshal(resBt, &res)
	return res, err
}

func TestHttpSrv(t *testing.T) {
	h := xhttp.New()
	http.Handle("/cmd1", h)
	http.HandleFunc("/cmd2", h.Handler)
	go http.ListenAndServe(":5673", nil)

	srv := h.Srv()
	srv.Use(srv.AccessLogger("MYSRV")) // 打印请求响应日志

	gtest.C(t, func(t *gtest.T) {
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

		res, err := sendToHttp("http://127.0.0.1:5673/cmd1", data)
		t.Log(res, err)
		t.Assert(err, nil)
		t.Assert(res["cmd"], data["cmd"])
		t.Assert(res["data"], data["data"])
		t.Assert(res["seqno"], data["seqno"])

		res, err = sendToHttp("http://127.0.0.1:5673/cmd2", data)

		t.Assert(err, nil)
		t.Assert(res["cmd"], data["cmd"])
		t.Assert(res["data"], data["data"])
		t.Assert(res["seqno"], data["seqno"])
	})

}
