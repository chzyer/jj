package rpcenc

import (
	"github.com/jj-io/jj/rpc"
	"gopkg.in/vmihailenco/msgpack.v2"
)

type MsgPackEncoding struct{}

func NewMsgPackEncoding() *MsgPackEncoding {
	return &MsgPackEncoding{}
}

func (mp *MsgPackEncoding) Decode(r rpc.BufferReader, v interface{}) error {
	return msgpack.NewDecoder(r).Decode(v)
}

func (mp *MsgPackEncoding) Encode(w rpc.BufferWriter, v interface{}) error {
	return msgpack.NewEncoder(w).Encode(v)
}
