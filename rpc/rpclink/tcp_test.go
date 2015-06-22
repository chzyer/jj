package rpclink

import (
	"bytes"
	"io"
	"testing"
	"time"

	"gopkg.in/logex.v1"
)
import "net"

type MockMux struct {
	r  io.Reader
	ch chan *WriteItem
}

func NewMockMux() *MockMux {
	return &MockMux{
		ch: make(chan *WriteItem),
	}
}

func (mm *MockMux) OnClosed() {

}

func (mm *MockMux) Init(r io.Reader) {
	mm.r = r
}

func (mm *MockMux) Handle(buf *bytes.Buffer) error {
	b := make([]byte, 512)
	n, err := mm.r.Read(b)
	if err != nil {
		return logex.Trace(err)
	}

	mm.ch <- &WriteItem{
		Data: b[:n],
	}
	return nil
}

func (mm *MockMux) WriteChan() <-chan *WriteItem {
	return mm.ch
}

func TestTcpLink(t *testing.T) {
	tcp := NewTcpLink(NewMockMux())
	ln, err := net.Listen(tcp.Protocol(), ":12347")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			t.Fatal(err)
		}
		tcp.Init(conn)
		tcp.Handle()
	}()
	time.Sleep(time.Microsecond)
	client, err := net.Dial(tcp.Protocol(), ":12347")
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Microsecond)
	client.Write([]byte("hello"))
	buf := make([]byte, 1024)
	n, err := client.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	if string(buf[:n]) != "hello" {
		t.Fatal("result not except")
	}

}
