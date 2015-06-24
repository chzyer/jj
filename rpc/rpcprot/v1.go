package rpcprot

import (
	"bytes"
	"encoding/binary"
	"io"
	"io/ioutil"

	"github.com/jj-io/jj/rpc"

	"gopkg.in/logex.v1"
)

const (
	version     = 1
	MaxBodySize = 10 << 10
)

var (
	ErrBodyTooLarge = logex.Define("body too large")
)

type ProtocolV1 struct {
	r *io.LimitedReader
	w io.Writer
}

func NewProtocolV1(r io.Reader, w io.Writer) *ProtocolV1 {
	return &ProtocolV1{
		r: &io.LimitedReader{r, 0},
		w: w,
	}
}

func (p1 *ProtocolV1) Read(buf *bytes.Buffer, metaEnc rpc.Encoding, p *rpc.Packet) error {
	p1.r.N += 4
	var length int32
	if err := binary.Read(p1.r, binary.BigEndian, &length); err != nil {
		return logex.Trace(err)
	}
	if length > MaxBodySize {
		return logex.Trace(ErrBodyTooLarge)
	}
	p1.r.N += int64(length)

	n, err := buf.ReadFrom(p1.r)
	if err == nil && n != int64(length) {
		return logex.Trace(io.ErrUnexpectedEOF)
	}
	if err != nil {
		return logex.Trace(err)
	}

	br := bytes.NewReader(buf.Bytes()[:length])
	if err := metaEnc.Decode(br, &p.Meta); err != nil {
		return logex.Trace(err, length, buf.Bytes())
	}
	data, _ := ioutil.ReadAll(br)

	p.Data = rpc.NewRawData(data)
	return nil
}

func (p1 *ProtocolV1) Write(metaEnc, bodyEnc rpc.Encoding, p *rpc.Packet) error {
	underBuf := make([]byte, 4, 512)
	buf := bytes.NewBuffer(underBuf)
	if err := metaEnc.Encode(buf, p.Meta); err != nil {
		return logex.Trace(err)
	}

	if p.Data != nil {
		if err := p.Data.Write(buf, bodyEnc); err != nil {
			return logex.Trace(err)
		}
	}

	logex.Debug("write: ", buf.Bytes())

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
