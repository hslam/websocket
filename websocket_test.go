// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

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
	{
		conn, err := Dial(network, addr, "/", nil)
		if err != nil {
			t.Error(err)
		}
		msg := "Hello World"
		if err := conn.WriteMessage([]byte(msg)); err != nil {
			t.Error(err)
		}
		data, err := conn.ReadMessage(nil)
		if err != nil {
			t.Error(err)
		} else if string(data) != msg {
			t.Error(string(data))
		}
		conn.Close()
	}
	{
		_, err := Dial(network, addr, "/", testSkipVerifyTLSConfig())
		if err == nil {
			t.Error()
		}
	}
	httpServer.Close()
	{
		_, err := Dial(network, addr, "/", nil)
		if err == nil {
			t.Error()
		}
	}
	wg.Wait()
}

func TestUpgrade(t *testing.T) {
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
			ws, err := Upgrade(conn, nil)
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
	data, err := conn.ReadMessage(nil)
	if err != nil {
		t.Error(err)
	} else if string(data) != msg {
		t.Error(string(data))
	}
	conn.Close()
	l.Close()
	wg.Wait()
}

func TestUpgradeTLS(t *testing.T) {
	network := "tcp"
	addr := ":8080"
	Serve := func(conn *Conn) {
		var msg []byte
		var err error
		for err == nil {
			msg, err = conn.ReadMessage(nil)
			if err != nil {
				break
			}
			err = conn.WriteMessage(msg)
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
			wg.Add(1)
			go func() {
				defer wg.Done()
				ws, err := Upgrade(conn, testTLSConfig())
				if ws != nil && err == nil {
					Serve(ws)
				}
			}()
		}
	}()
	{
		_, err := Dial(network, addr, "/", nil)
		if err == nil {
			t.Error(err)
		}
	}
	{
		conn, err := Dial(network, addr, "/", testSkipVerifyTLSConfig())
		if err != nil {
			t.Error(err)
		}
		msg := "Hello World"
		if err := conn.WriteMessage([]byte(msg)); err != nil {
			t.Error(err)
		}
		data, err := conn.ReadMessage(nil)
		if err != nil {
			t.Error(err)
		} else if string(data) != msg {
			t.Error(string(data))
		}
		conn.Close()
	}
	l.Close()
	wg.Wait()
}
