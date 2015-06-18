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
	w        io.Writer
}

func NewProtocolV1(encoding Encoding, r io.Reader, w io.Writer) *ProtocolV1 {
	return &ProtocolV1{
		r: &io.LimitedReader{r, 0},
		w: w,

		encoding: encoding,
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
	return n, logex.Trace(err)
}

func (p1 *ProtocolV1) Write(buffer []byte) (n int, err error) {
	lengthByte := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthByte, uint32(len(buffer)))
	n, err = p1.w.Write(lengthByte)
	if n != 4 && err == nil {
		err = io.ErrShortWrite
	}
	if err != nil {
		return n, logex.Trace(err)
	}

	n, err = p1.w.Write(buffer)
	if err != nil {
		return n, logex.Trace(err)
	}
	return
}
