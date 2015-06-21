package rpc

import (
	"io"
	"net"
)

type Encoding interface {
	Encode(w BufferWriter, v interface{}) error
	Decode(r BufferReader, v interface{}) error
}

type BufferReader interface {
	io.Reader
	Len() int
	WriteTo(w io.Writer) (int64, error)
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
}
