package websocket

import "errors"

func (c *Conn) ReadMessage(v interface{}) (err error) {
	f, err := c.readFrame()
	if err != nil {
		return err
	}
	switch data := v.(type) {
	case *string:
		*data = string(f.PayloadData)
		return nil
	case *[]byte:
		*data = f.PayloadData
		return nil
	}
	return errors.New("not supported")
}

func (c *Conn) WriteMessage(b interface{}) (err error) {
	switch data := b.(type) {
	case string:
		f := &frame{FIN: 1, Opcode: TextFrame, PayloadData: []byte(data)}
		return c.writeFrame(f)
	case *string:
		f := &frame{FIN: 1, Opcode: TextFrame, PayloadData: []byte(*data)}
		return c.writeFrame(f)
	case []byte:
		f := &frame{FIN: 1, Opcode: BinaryFrame, PayloadData: data}
		return c.writeFrame(f)
	case *[]byte:
		f := &frame{FIN: 1, Opcode: BinaryFrame, PayloadData: *data}
		return c.writeFrame(f)
	}
	return errors.New("not supported")
}

func (c *Conn) ReadBinaryMessage() (p []byte, err error) {
	f, err := c.readFrame()
	if err != nil {
		return nil, err
	}
	return f.PayloadData, nil
}

func (c *Conn) WriteBinaryMessage(b []byte) (err error) {
	f := &frame{FIN: 1, Opcode: BinaryFrame, PayloadData: b}
	return c.writeFrame(f)
}

func (c *Conn) ReadTextMessage() (p string, err error) {
	f, err := c.readFrame()
	if err != nil {
		return "", err
	}
	return string(f.PayloadData), nil
}

func (c *Conn) WriteTextMessage(b string) (err error) {
	f := &frame{FIN: 1, Opcode: TextFrame, PayloadData: []byte(b)}
	return c.writeFrame(f)
}
