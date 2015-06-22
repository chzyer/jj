package rpc

import (
	"bytes"
	"io"
	"net"
)

type Buffer struct {
	*bytes.Buffer
	r io.Reader
}

func NewBuffer(buf *bytes.Buffer) *Buffer {
	return &Buffer{
		Buffer: buf,
		r:      io.MultiReader(buf),
	}
}

func (b *Buffer) Read(buf []byte) (int, error) {
	return b.r.Read(buf)
}

func (b *Buffer) Prepand(r io.Reader) {
	b.r = io.MultiReader(r, b.Buffer)
}

type Encoding interface {
	Encode(w BufferWriter, v interface{}) error
	Decode(r BufferReader, v interface{}) error
}

type BufferReader interface {
	io.Reader
	Len() int
	WriteTo(w io.Writer) (int64, error)
	Prepand(r io.Reader)
}

type BufferWriter interface {
	io.Writer
}

type Linker interface {
	Init(net.Conn)
	Handle()
	Protocol() string
}

type ResponseWriter interface {
	Response(data interface{}) error
	Error(err error) error
	ErrorInfo(string) error
}
