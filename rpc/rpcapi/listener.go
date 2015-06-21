package rpcapi

import (
	"net"

	"github.com/jj-io/jj/rpc"

	"gopkg.in/logex.v1"
)

type NewLinkerFunc func() rpc.Linker

func Listen(addr, prot string, linkerFunc NewLinkerFunc) error {
	listener, err := net.Listen(prot, addr)
	if err != nil {
		return logex.Trace(err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			logex.Error(err)
			break
		}
		linker := linkerFunc()
		linker.Init(conn)
		go linker.Handle()
	}
	return nil
}
