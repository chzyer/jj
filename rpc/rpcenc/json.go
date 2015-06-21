package rpcenc

import (
	"encoding/json"

	"github.com/jj-io/jj/rpc"
)

type JSONEncoding struct{}

func NewJSONEncoding() *JSONEncoding {
	return &JSONEncoding{}
}

func (mp *JSONEncoding) Decode(r rpc.BufferReader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}

func (mp *JSONEncoding) Encode(w rpc.BufferWriter, v interface{}) error {
	return json.NewEncoder(w).Encode(v)
}
