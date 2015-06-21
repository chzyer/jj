package rpcenc

import (
	"bytes"
	"testing"
)

type MsgpackTest struct {
	Version int    `msgpack:"Version,omitempty"`
	Seq     int    `msgpack:"seq,omitempty"`
	Path    string `msgpack:"path,omitempty"`
	Error   string `msgpack:"error,omitempty"`
}

func TestMsgpackEncoding(t *testing.T) {
	var obj MsgpackTest
	obj.Version = 12346
	enc := NewMsgPackEncoding()
	buf := bytes.NewBuffer(nil)
	err := enc.Encode(buf, obj)
	if err != nil {
		t.Fatal(err)
	}
	var obj2 *MsgpackTest
	err = enc.Decode(buf, &obj2)
	if err != nil {
		t.Fatal(err)
	}
	if obj.Version != obj2.Version {
		t.Fatal("result not except")
	}
}
