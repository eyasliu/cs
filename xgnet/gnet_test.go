package xgnet_test

import (
	"testing"

	"github.com/eyasliu/cs/xgnet"
	"github.com/gogf/gf/test/gtest"
)

func TestGnet(t *testing.T) {
	server := xgnet.New("127.0.0.1:5672")
	gtest.C(t, func(t *gtest.T) {
		srv, err := server.Srv()
		srv.Handle("register")
		t.Assert(err, nil)
	})

}
