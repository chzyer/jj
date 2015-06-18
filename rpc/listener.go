package rpc

import (
	"net"

	"gopkg.in/logex.v1"
)

type Handler interface {
	Init(net.Conn)
	Handle()
}

func Listen(addr string, handler Handler) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return logex.Trace(err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			logex.Error(err)
			break
		}
		handler.Init(conn)
		go handler.Handle()
	}
	return nil
}
