package websocket

import (
	"io"
	"math/rand"
	"net"
	"sync"
	"time"
)

type Conn struct {
	isClient    bool
	random      *rand.Rand
	conn        net.Conn
	writer      io.Writer
	key         string
	accept      string
	path        string
	address     string
	readBuffer  []byte
	buffer      []byte
	connBuffer  []byte
	frameBuffer []byte
	framePool   *sync.Pool
}

func (c *Conn) Read(b []byte) (n int, err error) {
	if len(c.connBuffer) > 0 {
		if len(b) >= len(c.connBuffer) {
			copy(b, c.connBuffer)
			c.connBuffer = c.connBuffer[len(b):]
			return len(c.connBuffer), nil
		}
		copy(b, c.connBuffer[:len(b)])
		c.connBuffer = c.connBuffer[len(b):]
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
