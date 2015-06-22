package rpcmux

import (
	"testing"
	"time"

	"github.com/jj-io/jj/rpc"
	"github.com/jj-io/jj/rpc/rpcapi"
	"github.com/jj-io/jj/rpc/rpclink"
	"github.com/jj-io/jj/rpc/rpcprot"

	"gopkg.in/logex.v1"
)

func TestMux(t *testing.T) {
	go func() {
		rpcapi.Listen(":12346", "tcp", func() rpc.Linker {
			serveMux := NewServeMux()
			return rpclink.NewTcpLink(serveMux)
		})
	}()
	time.Sleep(time.Millisecond)
	clientMux := NewClientMux()
	tcpClient := rpclink.NewTcpLink(clientMux)
	if err := rpcapi.Dial(":12346", tcpClient); err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{}, 1)
	go func() {
		_, err := clientMux.Send(&rpcprot.Packet{
			Meta: &rpcprot.Meta{
				Version: 1,
				Seq:     1,
				Path:    "debug.sleep",
			},
			Data: rpcprot.NewData(&SleepData{
				Millisecond: 100,
			}),
		})
		if err != nil {
			t.Fatal(err)
			return
		}
		done <- struct{}{}
	}()
	resp, err := clientMux.Send(&rpcprot.Packet{
		Meta: &rpcprot.Meta{
			Version: 1,
			Seq:     2,
			Path:    "debug.ping",
		},
	})

	if err != nil {
		logex.Error(err)
		t.Fatal(err)
	}

	if resp.Meta.Seq == 2 {
	} else {
		logex.Error(err)
		t.Fatal(resp)
	}

	if len(done) > 0 {
		t.Fatal("sleep not working")
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}
