package websocket

import (
	"net"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestConn(t *testing.T) {
	network := "tcp"
	addr := ":8080"
	Serve := func(conn *Conn) {
		for {
			msg, err := conn.ReadMessage(nil)
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
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		httpServer.Serve(l)
	}()
	conn, err := Dial(network, addr, "/", nil)
	if err != nil {
		t.Error(err)
	}
	{
		if err := conn.SetDeadline(time.Now().Add(time.Minute)); err != nil {
			t.Error()
		}
		if err := conn.SetWriteDeadline(time.Now().Add(time.Minute)); err != nil {
			t.Error()
		}
		if n, err := conn.Write(nil); err != nil {
			t.Error(err)
		} else if n > 0 {
			t.Error(n)
		}
		msg := "Hello World"
		if n, err := conn.Write([]byte(msg)); err != nil {
			t.Error(err)
		} else if n != len(msg) {
			t.Error(n)
		}
		if n, err := conn.Read(nil); err != nil {
			t.Error(err)
		} else if n > 0 {
			t.Error(n)
		}
		buf := make([]byte, 64)
		if err := conn.SetReadDeadline(time.Now().Add(time.Minute)); err != nil {
			t.Error()
		}
		n, err := conn.Read(buf)
		if err != nil {
			t.Error(err)
		} else if string(buf[:n]) != msg {
			t.Error(string(buf[:n]))
		}
		if addr := conn.LocalAddr(); addr == nil {
			t.Error()
		}
		if addr := conn.RemoteAddr(); addr == nil {
			t.Error()
		}
	}
	{
		msg := "Hello World"
		if n, err := conn.Write([]byte(msg)); err != nil {
			t.Error(err)
		} else if n != len(msg) {
			t.Error(n)
		}
		var str string
		buf := make([]byte, 1)
		for len(str) < len(msg) {
			n, err := conn.Read(buf)
			if err != nil {
				t.Error(err)
			} else if n > 0 {
				str += string(buf)
			}
		}
		if str != msg {
			t.Error(str)
		}
	}
	conn.Close()
	httpServer.Close()
	wg.Wait()
}
