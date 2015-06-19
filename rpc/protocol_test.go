package rpc

import (
	"bytes"
	"testing"
)

func TestProtocol(t *testing.T) {
	buffer := bytes.NewBuffer(nil)
	p := NewProtocolV1(buffer, buffer)
	data := []byte("hello")
	n, err := p.Write(data)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(data) {
		t.Fatal("short writen")
	}
	readed := make([]byte, 512)
	n, err = p.Read(readed)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(data) {
		t.Fatal("short readed")
	}
	if string(readed[:n]) != string(data) {
		t.Fatal("result not except")
	}
}
