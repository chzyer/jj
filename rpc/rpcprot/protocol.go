package rpcprot

import (
	"bytes"

	"github.com/jj-io/jj/rpc"
)

type Protocol interface {
	Read(buf *bytes.Buffer, metaEnc rpc.Encoding, p *Packet) error
	Write(metaEnc, bodyEnc rpc.Encoding, p *Packet) error
}
