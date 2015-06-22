package rpcenc

import (
	"bytes"
	"encoding/json"
	"io"

	"gopkg.in/logex.v1"

	"github.com/jj-io/jj/rpc"
)

type JSONEncoding struct{}

func NewJSONEncoding() *JSONEncoding {
	return &JSONEncoding{}
}

func (mp *JSONEncoding) Decode(r rpc.BufferReader, v interface{}) error {
	decoder := json.NewDecoder(r)
	err := decoder.Decode(v)
	br := decoder.Buffered().(*bytes.Reader)
	ch, err := br.ReadByte()
	if err != nil {
		if logex.Equal(err, io.EOF) {
			return nil
		}
		return logex.Trace(err)
	}
	if ch != byte(10) {
		br.UnreadByte()
	}
	if br.Len() > 0 {
		r.Prepand(br)
	}
	return err
}

func (mp *JSONEncoding) Encode(w rpc.BufferWriter, v interface{}) error {
	return json.NewEncoder(w).Encode(v)
}
