package rpcenc

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"

	"github.com/jj-io/jj/rpc"

	"gopkg.in/logex.v1"
)

type AesMsgPackEncoding struct {
	msgp   *MsgPackEncoding
	encode cipher.Stream
	decode cipher.Stream
}

var commonIV = []byte("b1d15254f0f0417d")

func NewAesMsgPackEncoding(key []byte) (*AesMsgPackEncoding, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, logex.Trace(err)
	}
	return &AesMsgPackEncoding{
		encode: cipher.NewCFBEncrypter(block, commonIV),
		decode: cipher.NewCFBDecrypter(block, commonIV),
	}, nil
}

func (mp *AesMsgPackEncoding) Decode(r rpc.BufferReader, v interface{}) error {
	buf := bytes.NewBuffer(make([]byte, 0, r.Len()))
	r.WriteTo(buf)
	mp.encode.XORKeyStream(buf.Bytes(), buf.Bytes())
	return mp.msgp.Decode(buf, v)
}

func (mp *AesMsgPackEncoding) Encode(w rpc.BufferWriter, v interface{}) error {
	buf := bytes.NewBuffer(make([]byte, 0, 512))
	if err := mp.msgp.Encode(buf, v); err != nil {
		return logex.Trace(err)
	}
	mp.decode.XORKeyStream(buf.Bytes(), buf.Bytes())
	buf.WriteTo(w)
	return nil
}
