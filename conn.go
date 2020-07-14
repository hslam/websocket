package websocket

import (
	"math/rand"
	"net"
	"time"
)

type Conn struct {
	isClient    bool
	random      *rand.Rand
	conn        net.Conn
	key         string
	accept      string
	path        string
	address     string
	readBuffer  []byte
	buffer      []byte
	connBuffer  []byte
	frameBuffer []byte
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
	if len(b) >= len(f.PayloadData) {
		copy(b, f.PayloadData)
		return len(f.PayloadData), nil
	}
	copy(b, f.PayloadData[:len(b)])
	c.connBuffer = append(c.connBuffer, f.PayloadData[len(b):]...)
	return len(b), nil
}

func (c *Conn) Write(b []byte) (n int, err error) {
	f := &frame{FIN: 1, Opcode: BinaryFrame, PayloadData: b}
	err = c.writeFrame(f)
	if err != nil {
		return 0, err
	}
	return len(b), nil
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
