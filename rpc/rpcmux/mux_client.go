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
	packet   *rpc.Packet
	callback func(*rpc.Packet)
	resp     chan *rpc.Packet
}

type ClientMux struct {
	prot        rpc.Protocol
	Gtx         rpc.Context
	ctxFunc     rpc.GenContext
	Ctx         *rpc.EncContext
	respChan    chan *rpc.Packet
	writeChan   chan *rpc.WriteItem
	stopChan    chan struct{}
	global      []*clientWriteContext
	globalGuard sync.Mutex
	handler     rpc.Handler
}

func NewClientMux(h rpc.Handler, ctxFunc rpc.GenContext) *ClientMux {
	cm := &ClientMux{
		Ctx: &rpc.EncContext{
			MetaEnc: rpcenc.NewJSONEncoding(),
			BodyEnc: rpcenc.NewJSONEncoding(),
		},
		ctxFunc:   ctxFunc,
		stopChan:  make(chan struct{}),
		respChan:  make(chan *rpc.Packet, 10),
		writeChan: make(chan *rpc.WriteItem, 10),
		handler:   h,
	}
	go cm.respLoop()
	return cm
}

func (c *ClientMux) Init(r io.Reader) {
	c.prot = rpcprot.NewProtocolV1(r, c)
	if c.ctxFunc != nil {
		c.Gtx = c.ctxFunc()
	}
}

func (s *ClientMux) GetStopChan() <-chan struct{} {
	return s.stopChan
}

func (c *ClientMux) clean() {
	<-c.stopChan
	if c.Gtx != nil {
		c.Gtx.Close()
	}
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

		if packet.Meta.Type == rpc.MetaReq {
			if c.handler == nil {
				logex.Error("receive req type packet, but handler is null")
				continue
			}
			h := c.handler.GetHandler(packet.Meta.Path)
			if h == nil {
				logex.Error("packet handler not found:", packet)
				continue
			}
			c.handlerWrap(h, packet)
			continue
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

		if op.callback != nil {
			op.callback(packet)
		} else if op.resp != nil {
			op.resp <- packet
		} else {
			logex.Error("unknown action for writeItem", op)
		}

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
	resp, err := c.Send(rpc.NewReqPacket(method, data))
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

func (c *ClientMux) SendAsync(w *rpc.Packet, cb func(*rpc.Packet)) error {
	item := &clientWriteContext{
		packet:   w,
		callback: cb,
	}
	c.globalGuard.Lock()
	c.global = append(c.global, item)
	c.globalGuard.Unlock()

	err := c.prot.Write(c.Ctx.MetaEnc, c.Ctx.BodyEnc, w)
	if err != nil {
		return logex.Trace(err)
	}
	return nil
}

func (c *ClientMux) WritePacket(p *rpc.Packet) error {
	err := c.prot.Write(c.Ctx.MetaEnc, c.Ctx.BodyEnc, p)
	if err != nil {
		return logex.Trace(err)
	}
	return nil
}

func (c *ClientMux) Send(w *rpc.Packet) (p *rpc.Packet, err error) {
	item := &clientWriteContext{
		packet: w,
		resp:   make(chan *rpc.Packet),
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
	case <-time.After(10 * time.Second):
		err = logex.Trace(ErrTimeout)
	case <-c.stopChan:
		err = logex.Trace(net.ErrWriteToConnected)
	}
	return
}

func (c *ClientMux) handlerWrap(h rpc.HandlerFunc, p *rpc.Packet) {
	now := time.Now()
	h(NewResponseWriter(c.handler, c, p), rpc.NewRequest(p, c.Ctx, c.Gtx))
	logex.Debug("request time: ", p.Meta.Path, ",", time.Now().Sub(now))
}
