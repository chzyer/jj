package rpcenc

import (
	"bytes"
	"reflect"
	"testing"
)

type AesTest struct {
	Version int    `json:"version,omitempty"`
	Seq     int    `json:"seq,omitempty"`
	Path    string `json:"path,omitempty"`
	Error   string `json:"error,omitempty"`
}

func TestAesEncoding(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	a := &AesTest{
		Version: 1,
		Seq:     12,
		Path:    "Hello",
		Error:   "error",
	}

	jencd := NewJSONEncoding()
	enc, err := NewAesEncoding(jencd, key)
	if err != nil {
		t.Fatal(err)
	}
	buf := bytes.NewBuffer(nil)
	if err := enc.Encode(buf, &a); err != nil {
		t.Fatal(err)
	}

	var b *AesTest
	reader := bytes.NewReader(buf.Bytes())
	if err := enc.Decode(reader, &b); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(a, b) {
		t.Fatal("result not expect")
	}
}
