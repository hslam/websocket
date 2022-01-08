// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package websocket

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"
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

func TestUpgradeHTTP(t *testing.T) {
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
				continue
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
	{
		clientHandshake := func(c *Conn) error {
			c.accept = accept(c.key)
			reqHeader := "GET " + c.path + " HTTP/1.1\r\n"
			reqHeader += "Host: " + c.address + "\r\n"
			reqHeader += "Origin: *\r\n"
			reqHeader += "Connection: Upgrade\r\n"
			reqHeader += "Upgrade: websocket\r\n"
			reqHeader += "Sec-WebSocket-Version: 13\r\n"
			reqHeader += "Sec-WebSocket-Key: " + c.key + "\r\n\r\n"
			_, err := c.conn.Write([]byte(reqHeader))
			if err != nil {
				return err
			}
			// Require successful HTTP response
			// before switching to websocket protocol.
			resp, err := http.ReadResponse(bufio.NewReader(c.conn), &http.Request{Method: "GET"})
			if err == nil {
				accept := resp.Header.Get("Sec-WebSocket-Accept")
				if resp.Status == status && accept == c.accept {
					return nil
				}
				err = errors.New("unexpected HTTP response: " + resp.Status)
			}
			return err
		}
		fakeDial := func(network, address, path string, config *tls.Config) (*Conn, error) {
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
			conn := client(netConn, true, address, path)
			conn.SetBufferedInput(bufferSize)
			conn.SetBufferedOutput(bufferSize)
			err = clientHandshake(conn)
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

		conn, err := fakeDial(network, addr, "/", nil)
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
		clientHandshake := func(c *Conn) error {
			c.accept = accept(c.key)
			reqHeader := "POST " + c.path + " HTTP/1.1\r\n"
			reqHeader += "Host: " + c.address + "\r\n"
			reqHeader += "Origin: *\r\n"
			reqHeader += "Connection: Upgrade\r\n"
			reqHeader += "Upgrade: websocket\r\n"
			reqHeader += "Sec-WebSocket-Version: 13\r\n"
			reqHeader += "Sec-WebSocket-Key: " + c.key + "\r\n\r\n"
			_, err := c.conn.Write([]byte(reqHeader))
			if err != nil {
				return err
			}
			// Require successful HTTP response
			// before switching to websocket protocol.
			resp, err := http.ReadResponse(bufio.NewReader(c.conn), &http.Request{Method: "GET"})
			if err == nil {
				accept := resp.Header.Get("Sec-WebSocket-Accept")
				if resp.Status == status && accept == c.accept {
					return nil
				}
				err = errors.New("unexpected HTTP response: " + resp.Status)
			}
			return err
		}
		fakeDial := func(network, address, path string, config *tls.Config) (*Conn, error) {
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
			err = clientHandshake(conn)
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

		_, err := fakeDial(network, addr, "/", nil)
		if err == nil {
			t.Error()
		}
	}
	{
		clientHandshake := func(c *Conn) error {
			c.accept = accept(c.key)
			reqHeader := "GET " + c.path + " HTTP/1.1\r\n"
			reqHeader += "Host: " + c.address + "\r\n"
			reqHeader += "Origin: *\r\n"
			reqHeader += "Sec-WebSocket-Version: 13\r\n"
			reqHeader += "Sec-WebSocket-Key: " + c.key + "\r\n\r\n"
			_, err := c.conn.Write([]byte(reqHeader))
			if err != nil {
				return err
			}
			// Require successful HTTP response
			// before switching to websocket protocol.
			resp, err := http.ReadResponse(bufio.NewReader(c.conn), &http.Request{Method: "GET"})
			if err == nil {
				accept := resp.Header.Get("Sec-WebSocket-Accept")
				if resp.Status == status && accept == c.accept {
					return nil
				}
				err = errors.New("unexpected HTTP response: " + resp.Status)
			}
			return err
		}
		fakeDial := func(network, address, path string, config *tls.Config) (*Conn, error) {
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
			err = clientHandshake(conn)
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

		_, err := fakeDial(network, addr, "/", nil)
		if err == nil {
			t.Error()
		}
	}
	{
		clientHandshake := func(c *Conn) error {
			c.accept = accept(c.key)
			reqHeader := "GET " + c.path + " HTTP/1.1\r\n"
			reqHeader += "Host: " + c.address + "\r\n"
			reqHeader += "Origin: *\r\n"
			reqHeader += "Connection: Upgrade\r\n"
			reqHeader += "Upgrade: websocket\r\n"
			reqHeader += "Sec-WebSocket-Version: 13\r\n"
			reqHeader += "\r\n"
			_, err := c.conn.Write([]byte(reqHeader))
			if err != nil {
				return err
			}
			// Require successful HTTP response
			// before switching to websocket protocol.
			resp, err := http.ReadResponse(bufio.NewReader(c.conn), &http.Request{Method: "GET"})
			if err == nil {
				accept := resp.Header.Get("Sec-WebSocket-Accept")
				if resp.Status == status && accept == c.accept {
					return nil
				}
				err = errors.New("unexpected HTTP response: " + resp.Status)
			}
			return err
		}
		fakeDial := func(network, address, path string, config *tls.Config) (*Conn, error) {
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
			err = clientHandshake(conn)
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

		_, err := fakeDial(network, addr, "/", nil)
		if err == nil {
			t.Error()
		}
	}
	{
		clientHandshake := func(c *Conn) error {
			c.accept = accept(c.key)
			reqHeader := "GET " + c.path + " HTTP/1.1\r\n"
			reqHeader += "Host: " + c.address + "\r\n"
			reqHeader += "Origin: *\r\n"
			reqHeader += "Connection: Upgrade\r\n"
			reqHeader += "Upgrade: websocket\r\n"
			reqHeader += "Sec-WebSocket-Version: 13\r\n"
			reqHeader += "\r\n"
			_, err := c.conn.Write([]byte(reqHeader))
			if err != nil {
				return err
			}
			// Require successful HTTP response
			// before switching to websocket protocol.
			resp, err := http.ReadResponse(bufio.NewReader(c.conn), &http.Request{Method: "GET"})
			if err == nil {
				accept := resp.Header.Get("Sec-WebSocket-Accept")
				if resp.Status == status && accept == c.accept {
					return nil
				}
				err = errors.New("unexpected HTTP response: " + resp.Status)
			}
			return err
		}
		fakeDial := func(network, address, path string, config *tls.Config) (*Conn, error) {
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
			conn.Close()
			err = clientHandshake(conn)
			if err != nil {
				return nil, &net.OpError{
					Op:   "dial-http",
					Net:  network + " " + address,
					Addr: nil,
					Err:  err,
				}
			}
			return conn, nil
		}
		_, err := fakeDial(network, addr, "/", nil)
		if err == nil {
			t.Error()
		}
	}
	l.Close()
	wg.Wait()
}

type testResponse struct {
	handlerHeader http.Header
	status        int
	conn          net.Conn
}

func (w *testResponse) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.conn, nil, errors.New("not support")
}

func (w *testResponse) Header() http.Header {
	return w.handlerHeader
}

func (w *testResponse) Write(data []byte) (n int, err error) {
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

func (w *testResponse) WriteHeader(code int) {
	w.status = code
}

func TestResponseWriter(t *testing.T) {
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
	Upgrade := func(conn net.Conn, config *tls.Config) (*Conn, error) {
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
		res := &testResponse{handlerHeader: req.Header, conn: conn}
		return upgradeHTTP(res, req, true)
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
				continue
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
	{
		_, err := Dial(network, addr, "/", nil)
		if err == nil {
			t.Error(err)
		}
	}
	l.Close()
	wg.Wait()
}
