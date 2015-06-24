package rpcmux

import (
	"bytes"
	"io"
	"sync"
	"time"

	"github.com/jj-io/jj/rpc"
	"github.com/jj-io/jj/rpc/rpcenc"
	"github.com/jj-io/jj/rpc/rpcprot"

	"net"

	"gopkg.in/logex.v1"
)

var (
	ErrTimeout = logex.Define("timeout")
)

var _ rpc.Mux = &ClientMux{}

type clientWriteContext struct {
	packet *rpc.Packet
	resp   chan *rpc.Packet
	err    chan error
}

type ClientCtx struct {
	MetaEnc rpc.Encoding
	BodyEnc rpc.Encoding
}

type ClientMux struct {
	prot        rpcprot.Protocol
	Ctx         *ClientCtx
	respChan    chan *rpc.Packet
	writeChan   chan *rpc.WriteItem
	stopChan    chan struct{}
	global      []*clientWriteContext
	globalGuard sync.Mutex
}

func NewClientMux() *ClientMux {
	cm := &ClientMux{
		Ctx: &ClientCtx{
			MetaEnc: rpcenc.NewJSONEncoding(),
			BodyEnc: rpcenc.NewJSONEncoding(),
		},
		stopChan:  make(chan struct{}),
		respChan:  make(chan *rpc.Packet, 10),
		writeChan: make(chan *rpc.WriteItem, 10),
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
	var data rpc.Packet
	if err := c.prot.Read(buf, c.Ctx.MetaEnc, &data); err != nil {
		return logex.Trace(err)
	}
	c.respChan <- &data
	return nil
}

func (c *ClientMux) respLoop() {
	var (
		packet *rpc.Packet
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
			logex.Info("receive packet which sender is not found:", packet)
			continue
		}

		op.resp <- packet
	}
}

func (c *ClientMux) WriteChan() <-chan *rpc.WriteItem {
	return c.writeChan
}

func (c *ClientMux) Write(b []byte) (n int, err error) {
	wi := &rpc.WriteItem{
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

func (c *ClientMux) Call(method string, data, result interface{}) *rpc.Error {
	resp, err := c.Send(&rpc.Packet{
		Meta: rpc.NewMeta(method),
		Data: rpc.NewData(data),
	})
	if err != nil {
		return rpc.NewError(err, false)
	}
	if resp.Meta.Error != "" {
		return rpc.NewError(err, true)
	}
	if err := resp.Data.Decode(c.Ctx.BodyEnc, result); err != nil {
		return rpc.NewError(err, true)
	}
	return nil
}

func (c *ClientMux) Send(w *rpc.Packet) (p *rpc.Packet, err error) {
	item := &clientWriteContext{
		packet: w,
		resp:   make(chan *rpc.Packet),
		err:    make(chan error),
	}

	c.globalGuard.Lock()
	c.global = append(c.global, item)
	c.globalGuard.Unlock()

	err = c.prot.Write(c.Ctx.MetaEnc, c.Ctx.BodyEnc, w)
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
