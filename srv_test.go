package cmdsrv_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/eyasliu/cmdsrv"
	"github.com/gogf/gf/test/gtest"
)

type testAdapter struct {
	request  []*cmdsrv.Request
	response []*cmdsrv.Response
	state    map[string]interface{}
	stateMu  sync.RWMutex
}

func (a *testAdapter) Read() (string, *cmdsrv.Request, error) {
	if len(a.request) > 0 {
		m := a.request[0]
		a.request = a.request[1:len(a.request)]
		return "1", m, nil
	}
	return "", nil, errors.New("req empty")
}
func (*testAdapter) Write(sid string, resp *cmdsrv.Response) error {
	return nil
}
func (*testAdapter) Close(sid string) error {
	return nil
}
func (a *testAdapter) GetState(sid, key string) interface{} {
	return a.state[key]
}

func (a *testAdapter) SetState(sid, key string, v interface{}) {
	if a.state == nil {
		a.state = map[string]interface{}{}
	}
	a.stateMu.Lock()
	a.state[key] = v
	a.stateMu.Unlock()
}
func (a *testAdapter) GetAllSID() []string {
	return []string{"1"}
}

func TestSrv_MiddlewareCall(t *testing.T) {
	srv := cmdsrv.New(&testAdapter{
		request: []*cmdsrv.Request{
			{"a", "", nil},
			{"b", "", nil},
			{"c", "", nil},
			{"d", "", nil},
		},
	})

	gtest.C(t, func(t *gtest.T) {
		srv.Use(func(c *cmdsrv.Context) {
			c.Seqno = "a"
			c.Next()
			c.Seqno += "b"
			switch c.Cmd {
			case "a":
				t.Assert(c.Seqno, "acdb")
			case "b":
				t.Assert(c.Seqno, "acefdb")
			case "c":
				t.Assert(c.Seqno, "aceghfdb")
			case "d":
				t.Assert(c.Seqno, "aceijfdb")
			}
		}, func(c *cmdsrv.Context) {
			c.Seqno += "c"
			c.Next()
			c.Seqno += "d"
		})

		srv.Handle("a", func(c *cmdsrv.Context) {
			t.Assert(c.Seqno, "ac")
			t.Assert(c.Cmd, "a")
		})

		srvG1 := srv.Group(func(c *cmdsrv.Context) {
			c.Seqno += "e"
			c.Next()
			c.Seqno += "f"
		})
		srvG1.Handle("b", func(c *cmdsrv.Context) {
			t.Assert(c.Seqno, "ace")
			t.Assert(c.Cmd, "b")
		})

		srvG2 := srvG1.Group(func(c *cmdsrv.Context) {
			c.Seqno += "g"
			c.Next()
			c.Seqno += "h"
		})
		srvG2.Handle("c", func(c *cmdsrv.Context) {
			t.Assert(c.Seqno, "aceg")
			t.Assert(c.Cmd, "c")
		})
		srvG3 := srvG1.Group(func(c *cmdsrv.Context) {
			c.Seqno += "i"
			c.Next()
			c.Seqno += "j"
		})
		srvG3.Handle("d", func(c *cmdsrv.Context) {
			t.Assert(c.Seqno, "acei")
			t.Assert(c.Cmd, "d")
		})

		srv.Run()
		time.Sleep(50 * time.Millisecond)
	})
}

func TestSrv_Parse(t *testing.T) {
	srv := cmdsrv.New(&testAdapter{
		request: []*cmdsrv.Request{
			{"a", "1", []byte(`{"x":1}`)},
			{"b", "2", []byte(`[{"y":2},{"z":3}]`)},
		},
	})
	gtest.C(t, func(t *gtest.T) {
		srv.Handle("a", func(c *cmdsrv.Context) {
			var body *struct {
				Y int `p:"x"`
			}
			err1 := c.Parse(&body)
			t.Assert(err1, nil)
			t.Assert(body.Y, 1)

			var body2 *struct {
				X int `v:"min:10"`
			}
			err2 := c.Parse(&body2)
			t.AssertNE(err2, nil)
			t.Log(err2)
			t.Assert(body2.X, 1)

			var body3 map[string]int
			err3 := c.Parse(&body3)
			t.Assert(err3, nil)
			t.Logf("body3: %v", body3)
			t.AssertNE(body3, nil)
			t.Assert(body3["x"], 1)
		})
		srv.Handle("b", func(c *cmdsrv.Context) {
			var body []struct {
				Y int
				Z int
			}
			err := c.Parse(&body)
			t.Assert(err, nil)
			t.Assert(len(body), 2)
			t.Assert(body[0].Y, 2)
			t.Assert(body[1].Z, 3)

			var body2 map[string]int
			err2 := c.Parse(&body2)
			t.AssertNE(err2, nil)
		})
		srv.Run()
		time.Sleep(100 * time.Millisecond)
	})
}
func TestSrv_Response(t *testing.T) {
	srv := cmdsrv.New(&testAdapter{
		request: []*cmdsrv.Request{
			{"a", "", nil},
			{"b", "", nil},
			{"c", "", nil},
			{"d", "", nil},
			{"e", "", nil},
			{"f", "", nil},
		},
	})
	gtest.C(t, func(t *gtest.T) {
		srv.Use(func(c *cmdsrv.Context) {
			c.Next()
			switch c.Cmd {
			case "a":
				t.Assert(c.Code, 0)
				t.Assert(c.Msg, "ok")
				t.Assert(c.Data, struct{}{})
			case "b":
				t.Assert(c.Code, 0)
				t.Assert(c.Msg, "ok")
				t.Assert(c.Data, "str")
			case "c":
				t.Assert(c.Code, 0)
				t.Assert(c.Msg, "ok")
				t.Assert(c.Data, 123)
			case "d":
				t.Assert(c.Code, 11)
				t.Assert(c.Msg, "err1")
				t.Assert(c.Data, struct{}{})
			case "e":
				t.Assert(c.Code, 12)
				t.Assert(c.Msg, "msg2")
				t.Assert(c.Data, "data2")
			case "f":
				t.Assert(c.Code, -1)
				t.Assert(c.Msg, "unsupport cmd")
				t.Assert(c.Data, struct{}{})
			}
		})
		srv.Handle("a", func(c *cmdsrv.Context) {
			c.OK()
		})
		srv.Handle("b", func(c *cmdsrv.Context) {
			c.OK("str")
		})
		srv.Handle("c", func(c *cmdsrv.Context) {
			c.OK(123)
		})
		srv.Handle("d", func(c *cmdsrv.Context) {
			c.Err(errors.New("err1"), 11)
		})
		srv.Handle("e", func(c *cmdsrv.Context) {
			c.Resp(12, "msg2", "data2")
		})
		srv.Run()
		time.Sleep(100 * time.Millisecond)
	})
}

func TestSrv_State(t *testing.T) {
	srv := cmdsrv.New(&testAdapter{
		request: []*cmdsrv.Request{
			{"a", "1", []byte(`{"x":1}`)},
			{"b", "2", []byte(`[{"y":2}]`)},
		},
	})
	uid := 10
	srv.Use(func(c *cmdsrv.Context) {
		c.Set("uid", uid)
	})
	gtest.C(t, func(t *gtest.T) {
		srv.Handle("a", func(c *cmdsrv.Context) {
			t.Assert(c.Get("uid").(int), uid)
		})
		srv.Run()
		time.Sleep(50 * time.Millisecond)
	})
}
