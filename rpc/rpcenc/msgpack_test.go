package rpcenc

import (
	"bytes"
	"testing"

	"github.com/jj-io/jj/rpc"

	"gopkg.in/logex.v1"
	"gopkg.in/vmihailenco/msgpack.v2"
)

type MsgpackTest struct {
	Version int    `msgpack:"Version,omitempty"`
	Seq     int    `msgpack:"seq,omitempty"`
	Path    string `msgpack:"path,omitempty"`
	Error   string `msgpack:"error,omitempty"`
}

type Meta struct {
	Version int    `json:"version,omitempty"`
	Seq     int    `json:"seq,omitempty"`
	Path    string `json:"path,omitempty"`
	Error   string `json:"error,omitempty"`
}

func TestMeta(t *testing.T) {
	return
	data := []byte{
		147, 170, 100, 101, 98, 117, 103,
		46, 112, 105, 110, 103, 0, 1, 118,
	}
	var obj *Meta
	if err := msgpack.Unmarshal(data, &obj); err != nil {
		t.Error(err)
	}
	logex.Struct(obj)
}

func TestMsgpackEncoding(t *testing.T) {
	var obj MsgpackTest
	obj.Version = 12346
	enc := NewMsgPackEncoding()
	buf := rpc.NewBuffer(bytes.NewBuffer(nil))
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
