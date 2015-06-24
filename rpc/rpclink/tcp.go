package rpclink

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"sync/atomic"

	"github.com/jj-io/jj/rpc"

	"gopkg.in/logex.v1"
)

type TcpLink struct {
	mux       rpc.Mux
	conn      *net.TCPConn
	closeChan chan struct{}
	closed    int32
}

func NewTcpLink(mux rpc.Mux) *TcpLink {
	th := &TcpLink{
		mux:       mux,
		closeChan: make(chan struct{}, 1),
	}
	return th
}

func (th *TcpLink) Init(conn net.Conn) {
	th.conn = conn.(*net.TCPConn)
	logex.Debug("connect in:", th.conn.RemoteAddr())
	th.mux.Init(bufio.NewReader(conn))
}

func (th *TcpLink) Protocol() string {
	return "tcp"
}

func (th *TcpLink) Handle() {
	go th.HandleRead()
	go th.HandleWrite()
}

func (th *TcpLink) HandleWrite() {
	var (
		item *rpc.WriteItem
		err  error
		n    int
	)

	writeChan := th.mux.WriteChan()
	defer th.Close()

	for {
		select {
		case item = <-writeChan:
		case <-th.closeChan:
			return
		}
		n, err = th.conn.Write(item.Data)
		if err == nil && n != len(item.Data) {
			err = logex.Trace(io.ErrShortWrite)
		}
		select {
		case item.Resp <- err:
		case <-th.closeChan:
			return
		default:
		}
		if err != nil {
			logex.Error(err)
			break
		}
	}
}

func (th *TcpLink) HandleRead() {
	var (
		err    error
		buffer = bytes.NewBuffer(make([]byte, 0, 512))
	)
	defer th.Close()
	for {
		select {
		case <-th.closeChan:
			return
		default:
		}
		buffer.Reset()
		err = th.mux.Handle(buffer)
		if err != nil {
			if !logex.Equal(err, io.EOF) {
				logex.Error(err)
			}
			break
		}
	}
}

func (th *TcpLink) Close() {
	if !atomic.CompareAndSwapInt32(&th.closed, 0, 1) {
		return
	}
	logex.Debug("close tcplink")
	th.mux.OnClosed()
	th.conn.Close()
}
