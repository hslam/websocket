// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

// Package websocket implements a client and server for the WebSocket protocol as specified in RFC 6455.
package websocket

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
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

func UpgradeConn(conn net.Conn) *Conn {
	var b = bufio.NewReader(conn)
	req, err := http.ReadRequest(b)
	if err != nil {
		return nil
	}
	res := &response{handlerHeader: req.Header, conn: conn}
	return Upgrade(res, req)
}

type response struct {
	handlerHeader http.Header
	status        int
	conn          net.Conn
}

func (w *response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.conn, nil, nil
}

func (w *response) Header() http.Header {
	return w.handlerHeader
}

func (w *response) Write(data []byte) (n int, err error) {
	h := make([]byte, 0, 1024)
	h = append(h, fmt.Sprintf("HTTP/1.1 %03d %s\r\n", w.status, http.StatusText(w.status))...)
	h = append(h, fmt.Sprintf("Date: %s\r\n", time.Now().UTC().Format(http.TimeFormat))...)
	h = append(h, fmt.Sprintf("Content-Length: %d\r\n", len(data))...)
	h = append(h, "Content-Type: text/plain; charset=utf-8\r\n"...)
	h = append(h, "\r\n"...)
	h = append(h, data...)
	n, err = w.conn.Write(h)
	return len(data), err
}

func (w *response) WriteHeader(code int) {
	w.status = code
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
