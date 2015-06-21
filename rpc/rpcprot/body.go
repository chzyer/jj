package rpcprot

import "gopkg.in/vmihailenco/msgpack.v2"

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

func (d *Data) Unmarshal(v interface{}) error {
	return msgpack.Unmarshal(d.buf, v)
}
