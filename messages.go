// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package websocket

import (
	"errors"
	"github.com/hslam/writer"
)

func (c *Conn) SetConcurrency(concurrency func() int) {
	if concurrency == nil {
		return
	}
	c.writing.Lock()
	defer c.writing.Unlock()
	c.writer = writer.NewWriter(c.writer, concurrency, 65536, false)
}

func (c *Conn) ReadMsg(v interface{}) (err error) {
	c.reading.Lock()
	defer c.reading.Unlock()
	c.buffer = c.buffer[:0]
	c.connBuffer = c.connBuffer[:0]
	f, err := c.readFrame()
	if err != nil {
		return err
	}
	switch data := v.(type) {
	case *string:
		*data = string(f.PayloadData)
		c.putFrame(f)
		return nil
	case *[]byte:
		*data = f.PayloadData
		c.putFrame(f)
		return nil
	}
	return errors.New("not supported")
}

func (c *Conn) WriteMsg(b interface{}) (err error) {
	c.writing.Lock()
	defer c.writing.Unlock()
	f := c.getFrame()
	f.FIN = 1
	switch data := b.(type) {
	case string:
		f.Opcode = TextFrame
		f.PayloadData = []byte(data)
		return c.writeFrame(f)
	case *string:
		f.Opcode = TextFrame
		f.PayloadData = []byte(*data)
		return c.writeFrame(f)
	case []byte:
		f.Opcode = BinaryFrame
		f.PayloadData = data
		return c.writeFrame(f)
	case *[]byte:
		f.Opcode = BinaryFrame
		f.PayloadData = *data
		return c.writeFrame(f)
	}
	return errors.New("not supported")
}

func (c *Conn) ReadMessage() (p []byte, err error) {
	c.reading.Lock()
	defer c.reading.Unlock()
	c.buffer = c.buffer[:0]
	c.connBuffer = c.connBuffer[:0]
	f, err := c.readFrame()
	if err != nil {
		return nil, err
	}
	p = f.PayloadData
	c.putFrame(f)
	return
}

func (c *Conn) WriteMessage(b []byte) (err error) {
	c.writing.Lock()
	defer c.writing.Unlock()
	f := c.getFrame()
	f.FIN = 1
	f.Opcode = BinaryFrame
	f.PayloadData = b
	return c.writeFrame(f)
}

func (c *Conn) ReadTextMessage() (p string, err error) {
	c.reading.Lock()
	defer c.reading.Unlock()
	c.buffer = c.buffer[:0]
	c.connBuffer = c.connBuffer[:0]
	f, err := c.readFrame()
	if err != nil {
		return "", err
	}
	p = string(f.PayloadData)
	c.putFrame(f)
	return
}

func (c *Conn) WriteTextMessage(b string) (err error) {
	c.writing.Lock()
	defer c.writing.Unlock()
	f := c.getFrame()
	f.FIN = 1
	f.Opcode = TextFrame
	f.PayloadData = []byte(b)
	return c.writeFrame(f)
}
