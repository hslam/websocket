package websocket

import (
	"testing"
)

func BenchmarkFrameMarshal(b *testing.B) {
	buf := make([]byte, 64*1024)
	for i := 0; i < b.N; i++ {
		f := &frame{FIN: 1, Opcode: BinaryFrame, PayloadData: make([]byte, 512)}
		f.Marshal(buf)
	}
}

func BenchmarkFrameUnmarshal(b *testing.B) {
	buf := make([]byte, 64*1024)
	f := &frame{FIN: 1, Opcode: BinaryFrame, PayloadData: make([]byte, 512)}
	data, _ := f.Marshal(buf)
	for i := 0; i < b.N; i++ {
		var f2 = &frame{}
		f2.Unmarshal(data)
	}
}

func BenchmarkFrame(b *testing.B) {
	buf := make([]byte, 64*1024)
	f := &frame{FIN: 1, Opcode: BinaryFrame, PayloadData: make([]byte, 512)}
	for i := 0; i < b.N; i++ {
		data, _ := f.Marshal(buf)
		var f2 = &frame{}
		f2.Unmarshal(data)
	}
}
