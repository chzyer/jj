package rpc

import (
	"net"

	"gopkg.in/logex.v1"
)

func Dial(addr string, handler Handler) error {
	conn, err := net.Dial(handler.Protocol(), addr)
	if err != nil {
		return logex.Trace(err)
	}
	handler.Init(conn)
	handler.Handle()
	return nil
}
