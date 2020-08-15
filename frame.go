// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package websocket

import (
	"math/rand"
)

const (
	ContinuationFrame = 0x0
	TextFrame         = 0x1
	BinaryFrame       = 0x2
	CloseFrame        = 0x8
	PingFrame         = 0x9
	PongFrame         = 0xA
)

func (c *Conn) getFrame() *frame {
	return c.framePool.Get().(*frame)
}
func (c *Conn) putFrame(f *frame) {
	f.Reset()
	c.framePool.Put(f)
}
func (c *Conn) readFrame() (f *frame, err error) {
	f = c.getFrame()
	for {
		length := uint64(len(c.buffer))
		var i uint64 = 0
		for i < length {
			if length < 3 {
				break
			}
			var offset uint64
			offset, err = f.Unmarshal(c.buffer)
			if err != nil {
				return nil, err
			} else if offset == 0 {
				break
			} else {
				c.buffer = c.buffer[offset:]
				return
			}
		}
		var n int
		n, err = c.read(c.readBuffer)
		if err != nil {
			return nil, err
		}
		if n > 0 {
			c.buffer = append(c.buffer, c.readBuffer[:n]...)
		}
	}
	return
}

func (c *Conn) writeFrame(f *frame) error {
	if c.isClient {
		f.Mask = 1
		f.MaskingKey = maskingKey(c.random)
	}
	data, err := f.Marshal(nil)
	if err != nil {
		return err
	}
	c.write(data)
	c.putFrame(f)
	return nil
}

type frame struct {
	FIN                   byte
	RSV1                  byte
	RSV2                  byte
	RSV3                  byte
	Opcode                byte
	Mask                  byte
	PayloadLength         byte
	ExtendedPayloadLength uint64
	MaskingKey            []byte
	PayloadData           []byte
}

func (f *frame) Reset() {
	*f = frame{}
}

func (f *frame) Marshal(buf []byte) ([]byte, error) {
	var size uint64 = 2
	if f.Mask == 1 {
		size += 4
	}
	length := uint64(len(f.PayloadData))
	if length <= 125 {
	} else if length < 65536 {
		size += 2
	} else {
		size += 8
	}
	size += uint64(len(f.PayloadData))
	if uint64(cap(buf)) >= size {
		buf = buf[:size]
	} else {
		buf = make([]byte, size)
	}
	var offset uint64
	buf[0] = f.FIN<<7 + f.RSV1<<6 + f.RSV2<<5 + f.RSV3<<4 + f.Opcode
	if f.Mask == 0 {
		buf[1] = 0x00
	} else {
		buf[1] = 0x80
	}
	if length <= 125 {
		buf[1] |= byte(length)
		offset += 2
	} else if length < 65536 {
		buf[1] |= 126
		buf[2] = byte(length >> 8)
		buf[3] = byte(length)
		offset += 4
	} else {
		buf[1] |= 127
		buf[2] = byte(length >> 56)
		buf[3] = byte(length >> 48)
		buf[4] = byte(length >> 40)
		buf[5] = byte(length >> 32)
		buf[6] = byte(length >> 24)
		buf[7] = byte(length >> 16)
		buf[8] = byte(length >> 8)
		buf[9] = byte(length)
		offset += 10
	}
	if f.Mask == 0 {
		copy(buf[offset:], f.PayloadData)
		offset += uint64(len(f.PayloadData))
		return buf[:offset], nil
	}
	copy(buf[offset:offset+4], f.MaskingKey)
	offset += 4
	for i := 0; i < len(f.PayloadData); i++ {
		f.PayloadData[i] = f.PayloadData[i] ^ f.MaskingKey[i%4]
	}
	copy(buf[offset:], f.PayloadData)
	offset += uint64(len(f.PayloadData))
	return buf[:offset], nil
}
func (f *frame) Unmarshal(data []byte) (uint64, error) {
	var offset uint64
	if uint64(len(data)) < offset+1 {
		return 0, nil
	}
	f.FIN = data[0] >> 7
	f.RSV1 = data[0] >> 6 & 1
	f.RSV2 = data[0] >> 5 & 1
	f.RSV3 = data[0] >> 4 & 1
	f.Opcode = data[0] & 0xF
	offset++
	if uint64(len(data)) < offset+1 {
		return 0, nil
	}
	f.Mask = data[1] >> 7
	f.PayloadLength = byte(data[1] & 0x7F)
	offset++
	if f.PayloadLength <= 125 {
	} else if f.PayloadLength == 126 {
		if uint64(len(data)) < offset+2 {
			return 0, nil
		}
		f.ExtendedPayloadLength |= uint64(data[2]) << 8
		f.ExtendedPayloadLength |= uint64(data[3])
		offset += 2
	} else {
		if uint64(len(data)) < offset+8 {
			return 0, nil
		}
		f.ExtendedPayloadLength |= uint64(data[2]) << 56
		f.ExtendedPayloadLength |= uint64(data[3]) << 48
		f.ExtendedPayloadLength |= uint64(data[4]) << 40
		f.ExtendedPayloadLength |= uint64(data[5]) << 32
		f.ExtendedPayloadLength |= uint64(data[6]) << 24
		f.ExtendedPayloadLength |= uint64(data[7]) << 16
		f.ExtendedPayloadLength |= uint64(data[8]) << 8
		f.ExtendedPayloadLength |= uint64(data[9])
		offset += 8
	}
	if f.Mask == 0 {
		if f.ExtendedPayloadLength == 0 {
			if uint64(len(data)) < offset+uint64(f.PayloadLength) {
				return 0, nil
			}
			f.PayloadData = data[2 : 2+f.PayloadLength]
			offset += uint64(f.PayloadLength)
			return offset, nil
		}
		if uint64(len(data)) < offset+uint64(f.ExtendedPayloadLength) {
			return 0, nil
		}
		f.PayloadData = data[offset : offset+f.ExtendedPayloadLength]
		offset += uint64(f.ExtendedPayloadLength)
		return offset, nil
	}
	if f.ExtendedPayloadLength == 0 {
		if uint64(len(data)) < offset+4+uint64(f.PayloadLength) {
			return 0, nil
		}
		f.MaskingKey = data[2:6]
		f.PayloadData = data[6 : 6+f.PayloadLength]
		for i := 0; i < int(f.PayloadLength); i++ {
			f.PayloadData[i] = f.PayloadData[i] ^ f.MaskingKey[i%4]
		}
		offset += 4 + uint64(f.PayloadLength)
		return offset, nil
	}

	if uint64(len(data)) < offset+4+uint64(f.ExtendedPayloadLength) {
		return 0, nil
	}
	f.MaskingKey = data[offset : offset+4]
	offset += 4
	f.PayloadData = data[offset : offset+f.ExtendedPayloadLength]
	for i := 0; i < int(f.ExtendedPayloadLength); i++ {
		f.PayloadData[i] = f.PayloadData[i] ^ f.MaskingKey[i%4]
	}
	offset += uint64(f.ExtendedPayloadLength)
	return offset, nil
}

func maskingKey(random *rand.Rand) []byte {
	b := make([]byte, 4)
	for i := 0; i < 4; i++ {
		b[i] = byte(random.Intn(255))
	}
	return b
}
