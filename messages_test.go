// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package websocket

import (
	"net"
	"net/http"
	"sync"
	"testing"
)

func TestMessage(t *testing.T) {
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
	}
	{
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
	}
	{
		msg := "Hello World"
		if err := conn.SendMessage([]byte(msg)); err != nil {
			t.Error(err)
		}
		var v string
		err := conn.ReceiveMessage(&v)
		if err != nil {
			t.Error(err)
		} else if v != msg {
			t.Error(v)
		}
	}
	{
		msg := "Hello World"
		if err := conn.WriteTextMessage(msg); err != nil {
			t.Error(err)
		}
		data, err := conn.ReadTextMessage()
		if err != nil {
			t.Error(err)
		} else if data != msg {
			t.Error(data)
		}
	}
	conn.Close()
	httpServer.Close()
	wg.Wait()
}
