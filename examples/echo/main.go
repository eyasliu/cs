package main

import (
	"net/http"

	"github.com/eyasliu/cmdsrv"
	"github.com/eyasliu/cmdsrv/xhttp"
	"github.com/eyasliu/cmdsrv/xwebsocket"
)

func main() {
	// http
	httpAdapter := xhttp.New()
	http.Handle("/http", httpAdapter)

	wsAdapter := xwebsocket.New()
	http.Handle("/ws", wsAdapter)
	http.Handle("/", http.FileServer(http.Dir(`.`)))

	go http.ListenAndServe(":13000", nil)

	srv := cmdsrv.New(httpAdapter, wsAdapter)
	srv.Use(cmdsrv.AccessLogger("ECHOSRV"))

	srv.Use(func(c *cmdsrv.Context) {
		c.OK(c.RawData)
		c.Abort()
	})

	srv.Run()
}
