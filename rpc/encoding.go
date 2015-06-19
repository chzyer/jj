package rpc

import (
	"crypto/aes"
	"crypto/cipher"

	"gopkg.in/logex.v1"
	"gopkg.in/vmihailenco/msgpack.v2"
)

type Encoding interface {
	Encode(v interface{}) ([]byte, error)
	Decode(b []byte, v interface{}) error
}

type MsgPackEncoding struct{}

func (mp MsgPackEncoding) Decode(b []byte, v interface{}) error {
	return msgpack.Unmarshal(b, v)
}

func (mp MsgPackEncoding) Encode(v interface{}) ([]byte, error) {
	return msgpack.Marshal(v)
}

type AesMsgPackEncoding struct {
	msgp   MsgPackEncoding
	encode cipher.Stream
	decode cipher.Stream
}

func NewAesMsgPackEncoding(key []byte) (*AesMsgPackEncoding, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, logex.Trace(err)
	}
	return &AesMsgPackEncoding{
		encode: cipher.NewCFBEncrypter(block, make([]byte, block.BlockSize())),
		decode: cipher.NewCFBDecrypter(block, make([]byte, block.BlockSize())),
	}, nil
}

func (mp *AesMsgPackEncoding) Decode(b []byte, v interface{}) error {
	bin := make([]byte, len(b))
	copy(bin, b)
	mp.encode.XORKeyStream(bin, bin)
	return mp.msgp.Decode(bin, v)
}

func (mp *AesMsgPackEncoding) Encode(v interface{}) ([]byte, error) {
	b, err := mp.msgp.Encode(v)
	if err != nil {
		return nil, logex.Trace(err)
	}
	mp.decode.XORKeyStream(b, b)
	return b, nil
}
