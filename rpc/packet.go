package rpc

import (
	"fmt"
	"sync/atomic"
)

var metaSeq uint64

type Packet struct {
	Meta *Meta
	Data *Data
}

func NewPacket(path string, data interface{}) *Packet {
	return &Packet{
		Meta: NewMeta(path),
	}
}

func (p *Packet) String() string {
	return fmt.Sprintf("meta:%+v data:%+v", p.Meta, p.Data)
}

type Meta struct {
	Version int    `json:"version,omitempty"`
	Seq     uint64 `json:"seq"`
	Path    string `json:"path,omitempty"`
	Error   string `json:"error,omitempty"`
}

func NewMeta(path string) *Meta {
	return &Meta{
		Path: path,
		Seq:  atomic.AddUint64(&metaSeq, 1),
	}
}

func NewMetaError(seq uint64, err string) *Meta {
	return &Meta{
		Error: err,
		Seq:   seq,
	}
}
