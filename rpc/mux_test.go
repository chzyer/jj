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

	resp, err := clientMux.Write(&WriteOp{
		Encoding: MsgPackEncoding{},
		Data: &Operation{
			Version: 1,
			Seq:     1,
			Path:    "ping",
		},
	})

	if err != nil {
		t.Fatal(err)
	}
	if resp.Seq == 1 && resp.Path == "ping" && resp.Data.(string) == "pong" {
	} else {
		t.Fatal(resp)
	}
}
