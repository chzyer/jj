package rpcprot

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/jj-io/jj/rpc/rpcenc"

	"gopkg.in/logex.v1"
)

func TestV1(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 0, 512))
	buf2 := bytes.NewBuffer(make([]byte, 0, 512))
	prot := NewProtocolV1(buf, buf)
	metaEnc := rpcenc.NewJSONEncoding()
	bodyEnc := rpcenc.NewJSONEncoding()
	p1 := &Packet{
		Meta: &Meta{
			Version: 4,
			Seq:     3,
		},
		Data: NewData("hello"),
	}
	if err := prot.Write(metaEnc, bodyEnc, p1); err != nil {
		t.Fatal(err)
	}

	var p Packet
	if err := prot.Read(buf2, metaEnc, &p); err != nil {
		logex.Error(err)
		t.Fatal(err)
	}

	if !reflect.DeepEqual(p1.Meta, p.Meta) {
		t.Fatal("result not except")
	}

	var pData string
	if err := p.Data.Decode(bodyEnc, &pData); err != nil {
		t.Fatal(err)
	}
	if pData != p1.Data.underlay.(string) {
		t.Fatal("data not except")
	}
}
