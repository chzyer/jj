package rpcprot

import (
	"bytes"
	"fmt"

	"github.com/jj-io/jj/rpc"
)

type Data struct {
	underlay interface{}
	buf      []byte
}

func NewData(d interface{}) *Data {
	return &Data{
		underlay: d,
	}
}

func NewRawData(buf []byte) *Data {
	return &Data{
		buf: buf,
	}
}

func (d *Data) Decode(enc rpc.Encoding, v interface{}) error {
	buf := rpc.NewBuffer(bytes.NewBuffer(d.buf))
	err := enc.Decode(buf, v)
	if err != nil {
		return err
	}
	d.underlay = v
	return nil
}

func (d *Data) String() string {
	if d.underlay != nil {
		return fmt.Sprintf("%v", d.underlay)
	}
	return fmt.Sprintf("%v", d.buf)
}
