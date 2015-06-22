package rpcprot

import (
	"bytes"

	"github.com/jj-io/jj/rpc"
)

type Data struct {
	underlay interface{}
	buf      *rpc.Buffer
}

func NewData(d interface{}) *Data {
	return &Data{
		underlay: d,
	}
}

func NewRawData(buf []byte) *Data {
	return &Data{
		buf: rpc.NewBuffer(bytes.NewBuffer(buf)),
	}
}

func (d *Data) Decode(enc rpc.Encoding, v interface{}) error {
	return enc.Decode(d.buf, v)
}
