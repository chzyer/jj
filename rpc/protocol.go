package rpc

import (
	"encoding/binary"
	"io"

	"gopkg.in/logex.v1"
)

type Protocol interface {
	io.Reader
	io.Writer
	ReadWithEncoding(encoding Encoding, buffer []byte, v interface{}) error
	WriteWithEncoding(encoding Encoding, v interface{}) error
}

type NewProtocolFunc func(r io.Reader, w io.Writer) Protocol

type ProtocolV1 struct {
	r *io.LimitedReader
	w io.Writer
}

func NewProtocolV1(r io.Reader, w io.Writer) Protocol {
	return &ProtocolV1{
		r: &io.LimitedReader{r, 0},
		w: w,
	}
}

func (p1 *ProtocolV1) ReadWithEncoding(encoding Encoding, buffer []byte, v interface{}) error {
	n, err := p1.Read(buffer)
	if err != nil {
		return logex.Trace(err)
	}
	if err := encoding.Decode(buffer[:n], v); err != nil {
		return logex.Trace(err)
	}
	return nil
}

func (p1 *ProtocolV1) Read(buffer []byte) (n int, err error) {
	p1.r.N += 4
	n, err = p1.r.Read(buffer)
	if n != 4 && err == nil {
		err = logex.Trace(io.EOF)
	}
	if err != nil {
		return n, logex.Trace(err)
	}

	length := int64(binary.BigEndian.Uint32(buffer[:n]))
	p1.r.N += length
	n, err = p1.r.Read(buffer)
	return n, logex.Trace(err)
}

func (p1 *ProtocolV1) WriteWithEncoding(encoding Encoding, v interface{}) error {
	buf, err := encoding.Encode(v)
	if err != nil {
		return logex.Trace(err)
	}
	if _, err := p1.Write(buf); err != nil {
		return logex.Trace(err)
	}
	return nil
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
	if err == nil && n != len(buffer) {
		err = logex.Trace(io.ErrShortWrite)
	}
	if err != nil {
		return n, logex.Trace(err)
	}
	return
}
