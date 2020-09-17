// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package websocket

import (
	"github.com/hslam/writer"
	"io"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Conn struct {
	reading         sync.Mutex
	writing         sync.Mutex
	isClient        bool
	random          *rand.Rand
	conn            net.Conn
	writer          io.Writer
	key             string
	accept          string
	path            string
	address         string
	lowMemory       bool
	readBufferSize  int
	readBuffer      []byte
	writeBufferSize int
	writeBuffer     []byte
	buffer          []byte
	connBuffer      []byte
	closed          int32
}

func (c *Conn) Read(b []byte) (n int, err error) {
	c.reading.Lock()
	defer c.reading.Unlock()
	if len(c.connBuffer) > 0 {
		if len(b) >= len(c.connBuffer) {
			copy(b, c.connBuffer)
			c.connBuffer = c.connBuffer[:0]
			return len(c.connBuffer), nil
		}
		copy(b, c.connBuffer[:len(b)])
		num := copy(c.connBuffer, c.connBuffer[len(b):])
		c.connBuffer = c.connBuffer[:num]
		return len(b), nil
	}
	f, err := c.readFrame()
	if err != nil {
		return 0, err
	}
	length := len(f.PayloadData)
	if len(b) >= length {
		copy(b, f.PayloadData)
		c.putFrame(f)
		return length, nil
	}
	copy(b, f.PayloadData[:len(b)])
	c.connBuffer = append(c.connBuffer, f.PayloadData[len(b):]...)
	c.putFrame(f)
	return len(b), nil
}

func (c *Conn) read(b []byte) (n int, err error) {
	return c.conn.Read(b)
}

func (c *Conn) Write(b []byte) (n int, err error) {
	c.writing.Lock()
	defer c.writing.Unlock()
	f := c.getFrame()
	f.FIN = 1
	f.Opcode = BinaryFrame
	f.PayloadData = b
	err = c.writeFrame(f)
	if err != nil {
		return 0, err
	}
	return len(b), nil
}

func (c *Conn) write(b []byte) (n int, err error) {
	return c.writer.Write(b)
}

func (c *Conn) Close() error {
	if !atomic.CompareAndSwapInt32(&c.closed, 0, 1) {
		return nil
	}
	if w, ok := c.writer.(*writer.Writer); ok {
		w.Close()
	}
	c.writer = nil
	c.readBuffer = nil
	c.writeBuffer = nil
	c.buffer = nil
	c.connBuffer = nil
	return c.conn.Close()
}

func (c *Conn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *Conn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *Conn) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *Conn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}
