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

func (mp *JSONEncoding) Decode(r *bytes.Reader, v interface{}) error {
	decoder := json.NewDecoder(r)
	err := decoder.Decode(v)
	br := decoder.Buffered().(*bytes.Reader)
	if br.Len() > 0 {
		r.Seek(int64(-br.Len()), 1)
	}

	ch, err := r.ReadByte()
	if err != nil {
		if logex.Equal(err, io.EOF) {
			return nil
		}
		return logex.Trace(err)
	}
	if ch != byte(10) {
		r.UnreadByte()
	}
	return err
}

func (mp *JSONEncoding) Encode(w rpc.BufferWriter, v interface{}) error {
	return json.NewEncoder(w).Encode(v)
}
