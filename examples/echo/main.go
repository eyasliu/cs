package main

import (
	"net/http"

	"github.com/eyasliu/cs"
	"github.com/eyasliu/cs/xhttp"
	"github.com/eyasliu/cs/xwebsocket"
)

func main() {
	// http
	httpAdapter := xhttp.New()
	http.Handle("/http", httpAdapter)

	wsAdapter := xwebsocket.New()
	http.Handle("/ws", wsAdapter)
	http.Handle("/", http.FileServer(http.Dir(`.`)))

	go http.ListenAndServe(":13000", nil)

	srv := cs.New(httpAdapter, wsAdapter)
	srv.Use(srv.AccessLogger("ECHOSRV"))

	srv.Use(func(c *cs.Context) {
		c.OK(c.RawData)
		c.Abort()
	})

	srv.Run()
}
