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

func NewReqPacket(path string, data interface{}) *Packet {
	return &Packet{
		Meta: NewReqMeta(path),
		Data: NewData(data),
	}
}

func NewRespPacket(seq uint64, data interface{}) *Packet {
	return &Packet{
		Meta: NewRespMeta(seq),
		Data: NewData(data),
	}
}

func (p *Packet) String() string {
	return fmt.Sprintf("meta:%+v data:%+v", p.Meta, p.Data)
}

type MetaType int

const (
	MetaReq MetaType = iota
	MetaResp
)

type Meta struct {
	Type    MetaType `json:"type"`
	Version int      `json:"version,omitempty"`
	Seq     uint64   `json:"seq"`
	Path    string   `json:"path,omitempty"`
	Error   string   `json:"error,omitempty"`
}

func NewReqMeta(path string) *Meta {
	return &Meta{
		Type: MetaReq,
		Path: path,
		Seq:  atomic.AddUint64(&metaSeq, 1),
	}
}

func NewRespMeta(seq uint64) *Meta {
	return &Meta{
		Type: MetaResp,
		Seq:  seq,
	}
}

func NewMetaError(seq uint64, err string) *Meta {
	return &Meta{
		Error: err,
		Type:  MetaResp,
		Seq:   seq,
	}
}
