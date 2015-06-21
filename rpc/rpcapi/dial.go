package rpcapi

import (
	"net"

	"github.com/jj-io/jj/rpc"

	"gopkg.in/logex.v1"
)

func Dial(addr string, handler rpc.Linker) error {
	conn, err := net.Dial(handler.Protocol(), addr)
	if err != nil {
		return logex.Trace(err)
	}
	handler.Init(conn)
	handler.Handle()
	return nil
}
