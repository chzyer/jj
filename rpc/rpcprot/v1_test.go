package rpcprot

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/jj-io/jj/rpc/rpcenc"

	"gopkg.in/logex.v1"
)

func TestV1(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 0, 512))
	buf2 := bytes.NewBuffer(make([]byte, 0, 512))
	prot := NewProtocolV1(buf, buf)
	metaEnc := rpcenc.NewMsgPackEncoding()
	bodyEnc := rpcenc.NewMsgPackEncoding()
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

	if strings.HasSuffix(p1.Data.underlay.(string), string(p.Data.buf)) {
		t.Fatal("data not except")
	}
}
