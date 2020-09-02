// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package websocket

import (
	"errors"
	"github.com/hslam/autowriter"
)

func (c *Conn) SetBatch(concurrency func() int) {
	if concurrency == nil {
		return
	}
	c.writer = autowriter.NewAutoWriter(c.writer, false, 65536, 4, concurrency)
}

func (c *Conn) ReadMsg(v interface{}) (err error) {
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
	f, err := c.readFrame()
	if err != nil {
		return nil, err
	}
	p = f.PayloadData
	c.putFrame(f)
	return
}

func (c *Conn) WriteMessage(b []byte) (err error) {
	f := c.getFrame()
	f.FIN = 1
	f.Opcode = BinaryFrame
	f.PayloadData = b
	return c.writeFrame(f)
}

func (c *Conn) ReadTextMessage() (p string, err error) {
	f, err := c.readFrame()
	if err != nil {
		return "", err
	}
	p = string(f.PayloadData)
	c.putFrame(f)
	return
}

func (c *Conn) WriteTextMessage(b string) (err error) {
	f := c.getFrame()
	f.FIN = 1
	f.Opcode = TextFrame
	f.PayloadData = []byte(b)
	return c.writeFrame(f)
}
