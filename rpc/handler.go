package rpc

import (
	"bufio"
	"net"
	"time"
)

type TcpHandler struct {
	pgen      NewProtocolFunc
	mux       ServeMux
	timeout   time.Duration
	conn      *net.TCPConn
	writeChan chan []byte
	closeChan chan struct{}
}

func (th *TcpHandler) Init(conn net.Conn) {
	th.conn = conn.(*net.TCPConn)
}

type Hello struct {
	Uid string `msg:"uid"`
}

func (th *TcpHandler) Handle() {
	go th.HandleRead()
	go th.HandleWrite()
}

func (th *TcpHandler) HandleWrite() {
	var (
		buf []byte
	)
	defer th.Close()

	for {
		select {
		case buf = <-th.writeChan:
		case <-th.closeChan:
			return
		}
	}
}

func (th *TcpHandler) HandleRead() {
	var (
		err    error
		buffer = make([]byte, 1<<10)
		prot   = th.pgen(bufio.NewReader(th.conn), th.conn)
	)
	defer th.Close()

	for {
		err := th.mux.Read(prot, buffer)
		if err != nil {
			break
		}
	}
}

func (th *TcpHandler) Close() {
	th.conn.Close()
}
