// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package websocket

import (
	"github.com/hslam/buffer"
	"github.com/hslam/writer"
	"io"
	"math/rand"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

var gcing int32

func gc() {
	if atomic.CompareAndSwapInt32(&gcing, 0, 1) {
		defer atomic.StoreInt32(&gcing, 1)
		for i := 0; i < 24; i++ {
			time.Sleep(time.Millisecond * 125)
			runtime.GC()
		}
	}
}

// Conn represents a WebSocket connection.
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
	shared          bool
	scheduling      bool
	readBufferSize  int
	readBuffer      []byte
	writeBufferSize int
	writeBuffer     []byte
	buffer          []byte
	connBuffer      []byte
	readPool        *buffer.Pool
	writePool       *buffer.Pool
	closed          int32
}

// Read implements the net.Conn Read method.
func (c *Conn) Read(b []byte) (n int, err error) {
	if len(b) == 0 {
		return 0, nil
	}
	c.reading.Lock()
	if len(c.connBuffer) > 0 {
		if len(b) >= len(c.connBuffer) {
			n = copy(b, c.connBuffer)
			c.connBuffer = c.connBuffer[:0]
			c.reading.Unlock()
			return
		}
		n = copy(b, c.connBuffer[:len(b)])
		num := copy(c.connBuffer, c.connBuffer[len(b):])
		c.connBuffer = c.connBuffer[:num]
		c.reading.Unlock()
		return
	}
	f, err := c.readFrame(nil)
	if err == nil {
		length := len(f.PayloadData)
		if len(b) >= length {
			copy(b, f.PayloadData)
			c.putFrame(f)
			c.reading.Unlock()
			return length, nil
		}
		n = copy(b, f.PayloadData[:len(b)])
		c.connBuffer = append(c.connBuffer, f.PayloadData[len(b):]...)
		c.putFrame(f)
	}
	c.reading.Unlock()
	return
}

func (c *Conn) read(b []byte) (n int, err error) {
	return c.conn.Read(b)
}

// Write implements the net.Conn Write method.
func (c *Conn) Write(b []byte) (n int, err error) {
	if len(b) == 0 {
		return 0, nil
	}
	c.writing.Lock()
	f := c.getFrame()
	f.FIN = 1
	f.Opcode = BinaryFrame
	f.PayloadData = b
	err = c.writeFrame(f)
	if err == nil {
		n = len(b)
	}
	c.writing.Unlock()
	return
}

func (c *Conn) write(b []byte) (n int, err error) {
	return c.writer.Write(b)
}

// Close closes the connection.
func (c *Conn) Close() error {
	if !atomic.CompareAndSwapInt32(&c.closed, 0, 1) {
		return nil
	}
	if w, ok := c.writer.(*writer.Writer); ok {
		w.Close()
	}
	c.readBuffer = nil
	c.writeBuffer = nil
	c.buffer = nil
	c.connBuffer = nil
	go gc()
	return c.conn.Close()
}

// LocalAddr returns the local network address.
// The Addr returned is shared by all invocations of LocalAddr, so
// do not modify it.
func (c *Conn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
// The Addr returned is shared by all invocations of RemoteAddr, so
// do not modify it.
func (c *Conn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// SetDeadline implements the Conn SetDeadline method.
func (c *Conn) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

// SetReadDeadline implements the Conn SetReadDeadline method.
func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

// SetWriteDeadline implements the Conn SetWriteDeadline method.
func (c *Conn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}
