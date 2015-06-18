package rpc

import (
	"encoding/binary"
	"io"

	"gopkg.in/logex.v1"
)

type Protocol interface {
}

type ProtocolV1 struct {
	encoding Encoding
	r        *io.LimitedReader
}

func NewProtocolV1(encoding Encoding, r io.Reader) *ProtocolV1 {
	return &ProtocolV1{
		encoding: encoding,
		r:        r,
	}
}

func (p1 *ProtocolV1) Read(buffer []byte) (n int, err error) {
	p1.r.N += 4
	n, err = p1.r.Read(buffer)
	if err != nil {
		return n, logex.Trace(err)
	}
	length := int64(binary.BigEndian.Uint32(buffer[:n]))
	p1.r.N += length
	n, err = p1.r.Read(buffer)
	return n + 4, logex.Trace(err)
}

func (p1 *ProtocolV1) Write(buffer []byte) (n int, err error) {
	lengthByte := len(buffer)
}
