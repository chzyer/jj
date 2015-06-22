package rpcmux

import (
	"bytes"
	"io"
	"sync"
	"time"

	"github.com/jj-io/jj/rpc"
	"github.com/jj-io/jj/rpc/rpcenc"
	"github.com/jj-io/jj/rpc/rpclink"
	"github.com/jj-io/jj/rpc/rpcprot"

	"net"

	"gopkg.in/logex.v1"
)

var (
	ErrTimeout = logex.Define("timeout")
)

var _ rpclink.Mux = &ClientMux{}

type clientWriteContext struct {
	packet *rpcprot.Packet
	resp   chan *rpcprot.Packet
	err    chan error
}

type ClientMux struct {
	metaEnc     rpc.Encoding
	bodyEnc     rpc.Encoding
	prot        rpcprot.Protocol
	respChan    chan *rpcprot.Packet
	writeChan   chan *rpclink.WriteItem
	stopChan    chan struct{}
	global      []*clientWriteContext
	globalGuard sync.Mutex
}

func NewClientMux() *ClientMux {
	cm := &ClientMux{
		metaEnc:   rpcenc.NewJSONEncoding(),
		bodyEnc:   rpcenc.NewJSONEncoding(),
		stopChan:  make(chan struct{}),
		respChan:  make(chan *rpcprot.Packet, 10),
		writeChan: make(chan *rpclink.WriteItem, 10),
	}
	go cm.respLoop()
	return cm
}

func (c *ClientMux) Init(r io.Reader) {
	c.prot = rpcprot.NewProtocolV1(r, c)
}

func (s *ClientMux) GetStopChan() <-chan struct{} {
	return s.stopChan
}

func (c *ClientMux) OnClosed() {
	close(c.stopChan)
}

func (c *ClientMux) Handle(buf *bytes.Buffer) error {
	var data rpcprot.Packet
	if err := c.prot.Read(buf, c.metaEnc, &data); err != nil {
		return logex.Trace(err)
	}
	c.respChan <- &data
	return nil
}

func (c *ClientMux) respLoop() {
	var (
		packet *rpcprot.Packet
		op     *clientWriteContext
	)
	for {
		op = nil
		select {
		case packet = <-c.respChan:
			logex.Debug("client rece:", packet)
		case <-c.stopChan:
			return
		}

		c.globalGuard.Lock()
		for i := 0; i < len(c.global); i++ {
			if c.global[i].packet.Meta.Seq == packet.Meta.Seq {
				op = c.global[i]
				if i == 0 {
					c.global = c.global[1:]
				} else {
					c.global = append(c.global[:i], c.global[i+1:]...)
				}
				break
			}
		}
		c.globalGuard.Unlock()

		if op == nil {
			// TODO: add stat
			logex.Info("receive packet which not found sender:", packet)
			continue
		}

		op.resp <- packet
	}
}

func (c *ClientMux) WriteChan() <-chan *rpclink.WriteItem {
	return c.writeChan
}

func (c *ClientMux) Write(b []byte) (n int, err error) {
	wi := &rpclink.WriteItem{
		Data: b,
		Resp: make(chan error),
	}
	c.writeChan <- wi
	err = <-wi.Resp
	if err != nil {
		return 0, logex.Trace(err)
	}
	return len(b), nil
}

func (c *ClientMux) Send(w *rpcprot.Packet) (p *rpcprot.Packet, err error) {
	item := &clientWriteContext{
		packet: w,
		resp:   make(chan *rpcprot.Packet),
		err:    make(chan error),
	}

	c.globalGuard.Lock()
	c.global = append(c.global, item)
	c.globalGuard.Unlock()

	err = c.prot.Write(c.metaEnc, c.bodyEnc, w)
	if err != nil {
		return nil, logex.Trace(err)
	}

	select {
	case p = <-item.resp:
	case err = <-item.err:
		err = logex.Trace(err)
	case <-time.After(10 * time.Second):
		err = logex.Trace(ErrTimeout)
	case <-c.stopChan:
		err = logex.Trace(net.ErrWriteToConnected)
	}
	return
}
