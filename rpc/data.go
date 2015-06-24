package rpc

import (
	"bytes"
	"fmt"

	"gopkg.in/logex.v1"
)

type Data struct {
	underlay interface{}
	buf      []byte
}

func NewData(d interface{}) *Data {
	if d == nil {
		return nil
	}
	return &Data{
		underlay: d,
	}
}

func NewRawData(buf []byte) *Data {
	return &Data{
		buf: buf,
	}
}

func (d *Data) Write(buf BufferWriter, enc Encoding) error {
	if err := enc.Encode(buf, d.underlay); err != nil {
		return logex.Trace(err)
	}
	return nil
}

func (d *Data) Decode(enc Encoding, v interface{}) error {
	buf := bytes.NewReader(d.buf)
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
