package rpc

import (
	"bufio"
	"net"

	"gopkg.in/logex.v1"
)

type TcpHandler struct {
	protGen   NewProtocolFunc
	mux       Mux
	conn      *net.TCPConn
	writeChan chan *WriteOp
	closeChan chan struct{}
}

func NewTcpHandler(protGen NewProtocolFunc, mux Mux) *TcpHandler {
	th := &TcpHandler{
		mux:       mux,
		protGen:   protGen,
		writeChan: make(chan *WriteOp, 10),
		closeChan: make(chan struct{}, 1),
	}
	mux.SetWriteChan(th.writeChan)
	return th
}

func (th *TcpHandler) Init(conn net.Conn) {
	th.conn = conn.(*net.TCPConn)
}

func (th *TcpHandler) Protocol() string {
	return "tcp"
}

func (th *TcpHandler) Handle() {
	go th.HandleRead()
	go th.HandleWrite()
}

type WriteOp struct {
	Encoding Encoding
	Data     *Operation
	resp     chan *Operation
	err      chan error
}

func (th *TcpHandler) HandleWrite() {
	var (
		buf  *WriteOp
		prot = th.protGen(nil, th.conn)
	)
	defer th.Close()

	for {
		select {
		case buf = <-th.writeChan:
		case <-th.closeChan:
			return
		}
		err := prot.WriteWithEncoding(buf.Encoding, buf.Data)
		if err != nil {
			logex.Error(err)
			break
		}
	}
}

func (th *TcpHandler) HandleRead() {
	var (
		err    error
		buffer = make([]byte, 1<<10)
		prot   = th.protGen(bufio.NewReader(th.conn), nil)
	)
	defer th.Close()

	for {
		err = th.mux.Read(prot, buffer)
		if err != nil {
			logex.Error(err)
			break
		}
	}
}

func (th *TcpHandler) Close() {
	th.conn.Close()
}
