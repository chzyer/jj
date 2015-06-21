package rpcprot

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/jj-io/jj/rpc"

	"gopkg.in/logex.v1"
)

type Packet struct {
	Meta *Meta
	Data *Data
}

type Meta struct {
	Version int    `msgpack:"version,omitempty"`
	Seq     int    `msgpack:"seq,omitempty"`
	Path    string `msgpack:"path,omitempty"`
	Error   string `msgpack:"error,omitempty"`
}

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

func (p1 *ProtocolV1) Read(buf *bytes.Buffer, metaEnc rpc.Encoding, p *Packet) error {
	p1.r.N += 4
	var length int32
	if err := binary.Read(p1.r, binary.BigEndian, &length); err != nil {
		return logex.Trace(err)
	}
	p1.r.N += int64(length)

	n, err := buf.ReadFrom(p1.r)
	if err == nil && n != int64(length) {
		return logex.Trace(io.ErrUnexpectedEOF)
	}
	if err != nil {
		return logex.Trace(err)
	}

	if err := metaEnc.Decode(buf, &p.Meta); err != nil {
		return logex.Trace(err, string(buf))
	}

	p.Data = NewRawData(buf.Bytes())
	return nil
}

func (p1 *ProtocolV1) Write(metaEnc, bodyEnc rpc.Encoding, p *Packet) error {
	underBuf := make([]byte, 4, 512)
	buf := bytes.NewBuffer(underBuf)
	if err := metaEnc.Encode(buf, p.Meta); err != nil {
		return logex.Trace(err)
	}

	if p.Data != nil {
		if err := bodyEnc.Encode(buf, p.Data.underlay); err != nil {
			return logex.Trace(err)
		}
	}

	binary.BigEndian.PutUint32(underBuf[:4], uint32(buf.Len()-4))
	n, err := p1.w.Write(buf.Bytes())
	if err != nil {
		return logex.Trace(err)
	}
	if n != buf.Len() {
		return logex.Trace(io.ErrShortWrite)
	}
	return nil
}
