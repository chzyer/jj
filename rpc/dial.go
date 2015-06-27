package rpc

import (
	"net"

	"gopkg.in/logex.v1"
)

func Dial(addr string, linker Linker) error {
	conn, err := net.Dial(linker.Protocol(), addr)
	if err != nil {
		return logex.Trace(err)
	}
	linker.Init(conn)
	linker.Handle()
	return nil
}
