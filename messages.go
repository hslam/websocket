package websocket

import (
	"errors"
	"github.com/hslam/autowriter"
)

type Batch interface {
	Concurrency() int
}

func (c *Conn) SetBatch(batch Batch) {
	c.writer = autowriter.NewAutoWriter(c.writer, false, 65536, 4, batch)
}
func (c *Conn) ReadMessage(v interface{}) (err error) {
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

func (c *Conn) WriteMessage(b interface{}) (err error) {
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

func (c *Conn) ReadBinaryMessage() (p []byte, err error) {
	f, err := c.readFrame()
	if err != nil {
		return nil, err
	}
	p = f.PayloadData
	c.putFrame(f)
	return
}

func (c *Conn) WriteBinaryMessage(b []byte) (err error) {
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
