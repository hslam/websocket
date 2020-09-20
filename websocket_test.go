package websocket

import (
	"net"
	"net/http"
	"sync"
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

func TestUpgrade(t *testing.T) {
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
	l, _ := net.Listen(network, addr)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			conn, err := l.Accept()
			if err != nil {
				break
			}
			ws, err := Upgrade(conn)
			if err != nil {
				t.Error(err)
				return
			}
			if ws != nil {
				wg.Add(1)
				go func() {
					defer wg.Done()
					Serve(ws)
				}()
			}
		}
	}()
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
	l.Close()
	wg.Wait()
}