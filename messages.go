// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package websocket

import (
	"errors"
	"github.com/hslam/writer"
	"unsafe"
)

// SetScheduling sets scheduling option.
func (c *Conn) SetScheduling(scheduling bool) {
	c.scheduling = scheduling
}

// SetConcurrency sets a callback func concurrency for writer.
func (c *Conn) SetConcurrency(concurrency func() int) {
	if concurrency == nil {
		if w, ok := c.writer.(*writer.Writer); ok {
			w.Close()
		}
		c.writing.Lock()
		c.writer = c.conn
		c.writing.Unlock()
		return
	}
	c.writing.Lock()
	if _, ok := c.writer.(*writer.Writer); !ok {
		c.writer = writer.NewWriter(c.writer, concurrency, 65536, c.scheduling || c.shared)
	}
	c.writing.Unlock()
}

// ReceiveMessage receives single frame from ws, unmarshaled and stores in v.
func (c *Conn) ReceiveMessage(v interface{}) (err error) {
	c.reading.Lock()
	c.buffer = c.buffer[:0]
	c.connBuffer = c.connBuffer[:0]
	var f *frame
	f, err = c.readFrame(nil)
	if err == nil {
		switch data := v.(type) {
		case *string:
			*data = *(*string)(unsafe.Pointer(&f.PayloadData))
			c.putFrame(f)
		case *[]byte:
			*data = f.PayloadData
			c.putFrame(f)
		default:
			err = errors.New("not supported")
		}
	}
	c.reading.Unlock()
	return
}

// SendMessage sends v marshaled as single frame to ws.
func (c *Conn) SendMessage(v interface{}) (err error) {
	switch data := v.(type) {
	case string:
		if len(data) > 0 {
			c.writing.Lock()
			f := c.getFrame()
			f.FIN = 1
			f.Opcode = TextFrame
			f.PayloadData = []byte(data)
			err = c.writeFrame(f)
			c.writing.Unlock()
		}
		return
	case *string:
		if len(*data) > 0 {
			c.writing.Lock()
			f := c.getFrame()
			f.FIN = 1
			f.Opcode = TextFrame
			f.PayloadData = []byte(*data)
			err = c.writeFrame(f)
			c.writing.Unlock()
		}
		return
	case []byte:
		if len(data) > 0 {
			c.writing.Lock()
			f := c.getFrame()
			f.FIN = 1
			f.Opcode = BinaryFrame
			f.PayloadData = data
			err = c.writeFrame(f)
			c.writing.Unlock()
		}
		return
	case *[]byte:
		if len(*data) > 0 {
			c.writing.Lock()
			f := c.getFrame()
			f.FIN = 1
			f.Opcode = BinaryFrame
			f.PayloadData = *data
			err = c.writeFrame(f)
			c.writing.Unlock()
		}
		return
	}
	return errors.New("not supported")
}

// ReadMessage reads single message from ws.
func (c *Conn) ReadMessage(buf []byte) (p []byte, err error) {
	c.reading.Lock()
	c.buffer = c.buffer[:0]
	c.connBuffer = c.connBuffer[:0]
	var f *frame
	f, err = c.readFrame(buf)
	if err == nil {
		p = f.PayloadData
		c.putFrame(f)
	}
	c.reading.Unlock()
	return
}

// WriteMessage writes single message to ws.
func (c *Conn) WriteMessage(b []byte) (err error) {
	if len(b) > 0 {
		c.writing.Lock()
		f := c.getFrame()
		f.FIN = 1
		f.Opcode = BinaryFrame
		f.PayloadData = b
		err = c.writeFrame(f)
		c.writing.Unlock()
	}
	return
}

// ReadTextMessage reads single text message from ws.
func (c *Conn) ReadTextMessage() (p string, err error) {
	c.reading.Lock()
	c.buffer = c.buffer[:0]
	c.connBuffer = c.connBuffer[:0]
	var f *frame
	f, err = c.readFrame(nil)
	if err == nil {
		p = *(*string)(unsafe.Pointer(&f.PayloadData))
		c.putFrame(f)
	}
	c.reading.Unlock()
	return
}

// WriteTextMessage writes single text message to ws.
func (c *Conn) WriteTextMessage(b string) (err error) {
	if len(b) > 0 {
		c.writing.Lock()
		f := c.getFrame()
		f.FIN = 1
		f.Opcode = TextFrame
		f.PayloadData = []byte(b)
		err = c.writeFrame(f)
		c.writing.Unlock()
	}
	return
}
