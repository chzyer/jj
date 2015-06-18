package rpc

import (
	"bufio"
	"io"
	"net"
	"time"

	"encoding/binary"

	"gopkg.in/logex.v1"
	"gopkg.in/vmihailenco/msgpack.v2"
)

type TcpHandler struct {
	timeout time.Duration
	conn    *net.TCPConn
}

func (th *TcpHandler) Init(conn net.Conn) {
	th.conn = conn.(*net.TCPConn)
}

type Hello struct {
	Uid string `msg:"uid"`
}

func (th *TcpHandler) Handle() {
	buffer := make([]byte, 1<<10)
	pr := NewProtocolReader(bufio.NewReader(th.conn))
	for {
		n, err := pr.Read(buffer)
		if err != nil {
			logex.Error(err)
			break
		}
		var hello *Hello
		if err := msgpack.Unmarshal(buffer[:n], &hello); err != nil {
			logex.Trace(err)
			return
		}
		logex.Struct(hello)
	}
	th.Close()
}

func (th *TcpHandler) Close() {
	th.conn.Close()
}

type ProtocolReader struct {
	r *io.LimitedReader
}

func NewProtocolReader(r io.Reader) *ProtocolReader {
	return &ProtocolReader{
		r: &io.LimitedReader{r, 0},
	}
}

func (pr *ProtocolReader) Read(buffer []byte) (n int, err error) {
	pr.r.N += 4
	n, err = pr.r.Read(buffer)
	if err != nil {
		return n, logex.Trace(err)
	}
	length := int64(binary.BigEndian.Uint32(buffer[:n]))
	pr.r.N += length
	n, err = pr.r.Read(buffer)
	return n, logex.Trace(err)
}
