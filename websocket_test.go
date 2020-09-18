package websocket

import (
	"net"
	"net/http"
	"testing"
)

func TestWebsocket(t *testing.T) {
	network := "tcp"
	addr := ":8080"
	Serve := func(conn *Conn) {
		for {
			msg, err := conn.ReadMessage()
			if err != nil {
				break
			}
			conn.WriteMessage(msg)
		}
		conn.Close()
	}

	httpServer := &http.Server{
		Addr:    addr,
		Handler: Handler(Serve),
	}
	l, _ := net.Listen(network, addr)
	go httpServer.Serve(l)
	conn, err := Dial(network, addr, "/", nil)
	if err != nil {
		t.Error(err)
	}
	msg := "Hello World"
	if err := conn.WriteMessage([]byte(msg)); err != nil {
		t.Error(err)
	}
	data, err := conn.ReadMessage()
	if err != nil {
		t.Error(err)
	} else if string(data) != msg {
		t.Error(string(data))
	}
	conn.Close()
	httpServer.Close()
}
