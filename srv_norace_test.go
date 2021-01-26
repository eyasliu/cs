// +build !race test

package cmdsrv_test

import (
	"github.com/eyasliu/cmdsrv"
	"github.com/gogf/gf/test/gtest"
	"testing"
	"time"
)

func TestSrv_MutilServer(t *testing.T) {
	server1 := &testAdapter{request: []*cmdsrv.Request{{"a", "1", []byte{1}}}, sid: "1"}
	server2 := &testAdapter{request: []*cmdsrv.Request{{"a", "2", []byte{2}}}, sid: "2"}
	server3 := &testAdapter{request: []*cmdsrv.Request{{"a", "3", []byte{3}}}, sid: "3"}
	server4 := &testAdapter{request: []*cmdsrv.Request{{"a", "4", []byte{4}}}, sid: "4"}
	server5 := &testAdapter{request: []*cmdsrv.Request{{"a", "5", []byte{5}}}, sid: "5"}
	gtest.C(t, func(t *gtest.T) {
		srv := cmdsrv.New(server1, server2)
		seqnos := []string{} // 这个变量会产生数据竞争，数据竞争就会导致里面数组的顺序不确定，但这只是测试代码
		srv.Handle("a", func(c *cmdsrv.Context) {
			seqnos = append(seqnos, c.Seqno)
			c.OK()
		})
		srv.AddServer(server3)
		go srv.Run()

		time.Sleep(50 * time.Millisecond)
		t.AssertIN(seqnos, []string{"1", "2", "3"}) // 1 + 2 + 3

		srv.AddServer(server4)
		time.Sleep(50 * time.Millisecond)
		t.AssertIN(seqnos, []string{"1", "2", "3", "4"}) // 1 + 2 + 3 + 4

		srv.AddServer(server5)
		time.Sleep(50 * time.Millisecond)
		t.AssertIN(seqnos, []string{"1", "2", "3", "4", "5"}) // 1 + 2 + 3 + 4 + 5

	})
}
