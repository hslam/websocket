// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

// Package websocket implements a client and server for the WebSocket protocol as specified in RFC 6455.
package websocket

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"
)

var responsePool = &sync.Pool{New: func() interface{} {
	return make([]byte, 1024)
}}

// UpgradeHTTP upgrades the HTTP server connection to the WebSocket protocol.
func UpgradeHTTP(w http.ResponseWriter, r *http.Request) (*Conn, error) {
	if r.Method != "GET" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, "405 must GET\n")
		return nil, errors.New("405 must GET")
	}
	if r.Header.Get("Upgrade") != "websocket" || r.Header.Get("Connection") != "Upgrade" {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "400 not websocket protocol\n")
		return nil, errors.New("400 not websocket protocol")
	}
	key := r.Header.Get("Sec-WebSocket-Key")
	if key == "" {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "400 bad Key\n")
		return nil, errors.New("400 bad Key")
	}
	netConn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		return nil, err
	}
	conn := server(netConn, false, key)
	err = conn.handshake()
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// Upgrade upgrades the net.Conn conn to the WebSocket protocol.
func Upgrade(conn net.Conn, config *tls.Config) (*Conn, error) {
	if config != nil {
		tlsConn := tls.Server(conn, config)
		if err := tlsConn.Handshake(); err != nil {
			conn.Close()
			return nil, err
		}
		conn = tlsConn
	}
	var b = bufio.NewReader(conn)
	req, err := http.ReadRequest(b)
	if err != nil {
		return nil, err
	}
	res := &response{handlerHeader: req.Header, conn: conn}
	return UpgradeHTTP(res, req)
}

type response struct {
	handlerHeader http.Header
	status        int
	conn          net.Conn
}

func (w *response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.conn, bufio.NewReadWriter(bufio.NewReader(w.conn), bufio.NewWriter(w.conn)), nil
}

func (w *response) Header() http.Header {
	return w.handlerHeader
}

func (w *response) Write(data []byte) (n int, err error) {
	h := responsePool.Get().([]byte)[:0]
	h = append(h, fmt.Sprintf("HTTP/1.1 %03d %s\r\n", w.status, http.StatusText(w.status))...)
	h = append(h, fmt.Sprintf("Date: %s\r\n", time.Now().UTC().Format(http.TimeFormat))...)
	h = append(h, fmt.Sprintf("Content-Length: %d\r\n", len(data))...)
	h = append(h, "Content-Type: text/plain; charset=utf-8\r\n"...)
	h = append(h, "\r\n"...)
	h = append(h, data...)
	n, err = w.conn.Write(h)
	responsePool.Put(h)
	return len(data), err
}

func (w *response) WriteHeader(code int) {
	w.status = code
}

// Dial opens a new client connection to a WebSocket.
func Dial(network, address, path string, config *tls.Config) (*Conn, error) {
	var err error
	netConn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	if config != nil {
		config.ServerName = address
		tlsConn := tls.Client(netConn, config)
		if err = tlsConn.Handshake(); err != nil {
			tlsConn.Close()
			return nil, err
		}
		netConn = tlsConn
	}
	conn := client(netConn, false, address, path)
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

// Handler represents a http.Handler.
type Handler func(*Conn)

// ServeHTTP implements the http.Handler interface for a WebSocket
func (handler Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := UpgradeHTTP(w, r)
	if err == nil {
		handler(conn)
	}
}
