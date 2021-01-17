// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package websocket

import (
	"errors"
	"github.com/hslam/writer"
	"unsafe"
)

// SetConcurrency sets a callback func concurrency for writer.
func (c *Conn) SetConcurrency(concurrency func() int) {
	if concurrency == nil {
		return
	}
	c.writing.Lock()
	defer c.writing.Unlock()
	c.writer = writer.NewWriter(c.writer, concurrency, 65536, false)
}

// ReceiveMessage receives single frame from ws, unmarshaled and stores in v.
func (c *Conn) ReceiveMessage(v interface{}) (err error) {
	c.reading.Lock()
	defer c.reading.Unlock()
	c.buffer = c.buffer[:0]
	c.connBuffer = c.connBuffer[:0]
	f, err := c.readFrame(nil)
	if err != nil {
		return err
	}
	switch data := v.(type) {
	case *string:
		*data = *(*string)(unsafe.Pointer(&f.PayloadData))
		c.putFrame(f)
		return nil
	case *[]byte:
		*data = f.PayloadData
		c.putFrame(f)
		return nil
	}
	return errors.New("not supported")
}

// SendMessage sends v marshaled as single frame to ws.
func (c *Conn) SendMessage(v interface{}) (err error) {
	switch data := v.(type) {
	case string:
		if len(data) == 0 {
			return nil
		}
		c.writing.Lock()
		defer c.writing.Unlock()
		f := c.getFrame()
		f.FIN = 1
		f.Opcode = TextFrame
		f.PayloadData = []byte(data)
		return c.writeFrame(f)
	case *string:
		if len(*data) == 0 {
			return nil
		}
		c.writing.Lock()
		defer c.writing.Unlock()
		f := c.getFrame()
		f.FIN = 1
		f.Opcode = TextFrame
		f.PayloadData = []byte(*data)
		return c.writeFrame(f)
	case []byte:
		if len(data) == 0 {
			return nil
		}
		c.writing.Lock()
		defer c.writing.Unlock()
		f := c.getFrame()
		f.FIN = 1
		f.Opcode = BinaryFrame
		f.PayloadData = data
		return c.writeFrame(f)
	case *[]byte:
		if len(*data) == 0 {
			return nil
		}
		c.writing.Lock()
		defer c.writing.Unlock()
		f := c.getFrame()
		f.FIN = 1
		f.Opcode = BinaryFrame
		f.PayloadData = *data
		return c.writeFrame(f)
	}
	return errors.New("not supported")
}

// ReadMessage reads single message from ws.
func (c *Conn) ReadMessage(buf []byte) (p []byte, err error) {
	c.reading.Lock()
	defer c.reading.Unlock()
	c.buffer = c.buffer[:0]
	c.connBuffer = c.connBuffer[:0]
	f, err := c.readFrame(buf)
	if err != nil {
		return nil, err
	}
	p = f.PayloadData
	c.putFrame(f)
	return
}

// WriteMessage writes single message to ws.
func (c *Conn) WriteMessage(b []byte) (err error) {
	if len(b) == 0 {
		return nil
	}
	c.writing.Lock()
	defer c.writing.Unlock()
	f := c.getFrame()
	f.FIN = 1
	f.Opcode = BinaryFrame
	f.PayloadData = b
	return c.writeFrame(f)
}

// ReadTextMessage reads single text message from ws.
func (c *Conn) ReadTextMessage() (p string, err error) {
	c.reading.Lock()
	defer c.reading.Unlock()
	c.buffer = c.buffer[:0]
	c.connBuffer = c.connBuffer[:0]
	f, err := c.readFrame(nil)
	if err != nil {
		return "", err
	}
	p = *(*string)(unsafe.Pointer(&f.PayloadData))
	c.putFrame(f)
	return
}

// WriteTextMessage writes single text message to ws.
func (c *Conn) WriteTextMessage(b string) (err error) {
	if len(b) == 0 {
		return nil
	}
	c.writing.Lock()
	defer c.writing.Unlock()
	f := c.getFrame()
	f.FIN = 1
	f.Opcode = TextFrame
	f.PayloadData = []byte(b)
	return c.writeFrame(f)
}
