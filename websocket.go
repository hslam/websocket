// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

// Package websocket implements a client and server for the WebSocket protocol as specified in RFC 6455.
package websocket

import (
	"io"
	"net"
	"net/http"
)

func Upgrade(w http.ResponseWriter, r *http.Request) *Conn {
	if r.Method != "GET" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, "405 must GET\n")
		return nil
	}
	if r.Header.Get("Upgrade") != "websocket" || r.Header.Get("Connection") != "Upgrade" {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "400 not websocket protocol\n")
		return nil
	}
	key := r.Header.Get("Sec-WebSocket-Key")
	if key == "" {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "400 bad Key\n")
		return nil
	}
	netConn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		return nil
	}
	conn := server(netConn, key)
	err = conn.handshake()
	if err != nil {
		return nil
	}
	return conn
}

func Dial(address, path string) (*Conn, error) {
	var err error
	var network = "tcp"
	netConn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	conn := client(netConn, address, path)
	err = conn.handshake()
	if err != nil {
		conn.Close()
		return nil, &net.OpError{
			Op:   "dial-http",
			Net:  network + " " + address,
			Addr: nil,
			Err:  err,
		}
	}
	return conn, nil
}

type Handler func(*Conn)

func (handler Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn := Upgrade(w, r)
	handler(conn)
}
