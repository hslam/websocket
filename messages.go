// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package websocket

import (
	"errors"
	"github.com/hslam/buffer"
	"github.com/hslam/writer"
	"unsafe"
)

// SetBufferedOutput sets the buffered writer with the buffer size.
func (c *Conn) SetBufferedOutput(writeBufferSize int) {
	c.writing.Lock()
	if w, ok := c.writer.(*writer.Writer); ok {
		w.Close()
	}
	if writeBufferSize > 0 {
		c.writer = writer.NewWriter(c.conn, writeBufferSize)
	} else {
		c.writer = c.conn
		writeBufferSize = bufferSize
	}
	c.writeBufferSize = writeBufferSize
	if c.shared {
		c.writePool = buffer.AssignPool(writeBufferSize)
	} else {
		c.writeBuffer = make([]byte, writeBufferSize)
	}
	c.writing.Unlock()
}

// SetBufferedInput sets the read buffer size.
func (c *Conn) SetBufferedInput(readBufferSize int) {
	if readBufferSize < 1 {
		readBufferSize = bufferSize
	}
	c.reading.Lock()
	c.readBufferSize = readBufferSize
	if c.shared {
		c.readPool = buffer.AssignPool(readBufferSize)
	} else {
		c.readBuffer = make([]byte, readBufferSize)
	}
	c.reading.Unlock()
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
