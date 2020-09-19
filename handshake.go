// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package websocket

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"
)

const (
	guid   = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	status = "101 Switching Protocols"
)

func server(conn net.Conn, shared bool, key string) *Conn {
	var random = rand.New(rand.NewSource(time.Now().UnixNano()))
	var readBufferSize = bufferSize
	var writeBufferSize = bufferSize
	readBufferSize += 14
	writeBufferSize += 14
	var readBuffer []byte
	var writeBuffer []byte
	var readPool *sync.Pool
	var writePool *sync.Pool
	if shared {
		readPool = assignPool(readBufferSize)
		writePool = assignPool(writeBufferSize)
	} else {
		readBuffer = make([]byte, readBufferSize)
		writeBuffer = make([]byte, writeBufferSize)
	}
	return &Conn{
		conn:            conn,
		writer:          conn,
		random:          random,
		shared:          shared,
		readBufferSize:  readBufferSize,
		writeBufferSize: writeBufferSize,
		readBuffer:      readBuffer,
		writeBuffer:     writeBuffer,
		readPool:        readPool,
		writePool:       writePool,
		key:             key,
	}
}

func client(conn net.Conn, shared bool, address, path string) *Conn {
	var random = rand.New(rand.NewSource(time.Now().UnixNano()))
	var readBufferSize = bufferSize
	var writeBufferSize = bufferSize
	readBufferSize += 14
	writeBufferSize += 14
	var readBuffer []byte
	var writeBuffer []byte
	var readPool *sync.Pool
	var writePool *sync.Pool
	if shared {
		readPool = assignPool(readBufferSize)
		writePool = assignPool(writeBufferSize)
	} else {
		readBuffer = make([]byte, readBufferSize)
		writeBuffer = make([]byte, writeBufferSize)
	}
	return &Conn{
		isClient:        true,
		conn:            conn,
		writer:          conn,
		random:          random,
		shared:          shared,
		readBufferSize:  readBufferSize,
		writeBufferSize: writeBufferSize,
		readBuffer:      readBuffer,
		writeBuffer:     writeBuffer,
		readPool:        readPool,
		writePool:       writePool,
		key:             key(random),
		address:         address,
		path:            path,
	}
}

func (c *Conn) handshake() error {
	if c.isClient {
		return c.clientHandshake()
	}
	return c.serverHandshake()
}

func (c *Conn) clientHandshake() error {
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
	accept := resp.Header.Get("Sec-WebSocket-Accept")
	if err == nil && resp.Status == status && accept == c.accept {
		return nil
	}
	return errors.New("unexpected HTTP response: " + resp.Status)
}

func (c *Conn) serverHandshake() error {
	c.accept = accept(c.key)
	respHeader := "HTTP/1.1 " + status + "\r\n"
	respHeader += "Upgrade: websocket\r\n"
	respHeader += "Connection: Upgrade\r\n"
	respHeader += "Sec-WebSocket-Accept: " + c.accept + "\r\n\r\n"
	_, err := c.conn.Write([]byte(respHeader))
	return err
}

func key(random *rand.Rand) string {
	b := make([]byte, 16)
	for i := 0; i < 16; i++ {
		b[i] = byte(random.Intn(255))
	}
	return base64.StdEncoding.EncodeToString(b)
}

func accept(key string) string {
	text := key + guid
	h := sha1.New()
	h.Write([]byte(text))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
