package rpc

import (
	"testing"
	"time"
)

func TestMux(t *testing.T) {
	go func() {
		serveMux := NewServeMux()
		tcpServeHandler := NewTcpHandler(NewProtocolV1, serveMux)
		Listen(":12346", tcpServeHandler)
	}()
	time.Sleep(time.Millisecond)
	clientMux := NewClientMux()
	tcpClientHandler := NewTcpHandler(NewProtocolV1, clientMux)
	if err := Dial(":12346", tcpClientHandler); err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{}, 1)
	go func() {
		_, err := clientMux.Write(&WriteOp{
			Encoding: MsgPackEncoding{},
			Data: &Operation{
				Version: 1,
				Seq:     1,
				Path:    "sleep",
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		done <- struct{}{}
	}()
	resp, err := clientMux.Write(&WriteOp{
		Encoding: MsgPackEncoding{},
		Data: &Operation{
			Version: 1,
			Seq:     2,
			Path:    "ping",
		},
	})
	if len(done) > 0 {
		t.Fatal("sleep not working")
	}

	if err != nil {
		t.Fatal(err)
	}

	if resp.Seq == 2 && resp.Path == "ping" && resp.Data.(string) == "pong" {
	} else {
		t.Fatal(resp)
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}
