package rpcmux

import (
	"bytes"
	"io"
	"sync"
	"time"

	"github.com/jj-io/jj/rpc"
	"github.com/jj-io/jj/rpc/rpcenc"
	"github.com/jj-io/jj/rpc/rpcprot"

	"gopkg.in/logex.v1"
)

type state int

const (
	stateInit state = iota
	stateStart
)

var (
	ErrReceiveQuit = logex.Define("operation timeout, client quit")
)

type Request struct {
	Ctx  *Context
	Data *rpc.Data
	Meta *rpc.Meta
}

func NewRequest(p *rpc.Packet, ctx *Context) *Request {
	return &Request{
		Ctx:  ctx,
		Meta: p.Meta,
		Data: p.Data,
	}
}

func (r *Request) Params(v interface{}) error {
	return r.Data.Decode(r.Ctx.BodyEnc, v)
}

var _ rpc.Mux = &ClientMux{}

// single-conn request multiplexer
type ServeMux struct {
	prot        rpcprot.Protocol
	ctx         *Context
	useEncoding bool
	state       state
	workChan    chan *rpc.Packet
	workGroup   sync.WaitGroup
	stopChan    chan struct{}
	writeChan   chan *rpc.WriteItem
	Handler     *Handler
}

func NewServeMux() *ServeMux {
	ctx := NewContext(rpcenc.NewJSONEncoding(), rpcenc.NewJSONEncoding())
	sm := &ServeMux{
		ctx:       ctx,
		stopChan:  make(chan struct{}),
		workChan:  make(chan *rpc.Packet, 10),
		writeChan: make(chan *rpc.WriteItem),
	}
	go sm.handleLoop()
	return sm
}

func (s *ServeMux) SetHandler(h *Handler) {
	s.Handler = h
}

func (s *ServeMux) GetStopChan() <-chan struct{} {
	return s.stopChan
}

func (s *ServeMux) OnClosed() {
	close(s.stopChan)
}

func (s *ServeMux) Init(r io.Reader) {
	s.prot = rpcprot.NewProtocolV1(r, s)
}

func (s *ServeMux) Send(p *rpc.Packet) error {
	err := s.prot.Write(s.ctx.MetaEnc, s.ctx.BodyEnc, p)
	if err != nil {
		return logex.Trace(err)
	}
	return nil
}

func (s *ServeMux) Handle(buf *bytes.Buffer) error {
	var p rpc.Packet
	if err := s.prot.Read(buf, s.ctx.MetaEnc, &p); err != nil {
		return logex.Trace(err)
	}
	s.workChan <- &p
	return nil
}

func (s *ServeMux) Close() {
	close(s.stopChan)
	s.workGroup.Wait()
}

func (s *ServeMux) handlerWrap(h HandlerFunc, p *rpc.Packet, ctx *Context) {
	now := time.Now()
	h(NewResponseWriter(s, p), NewRequest(p, ctx))
	logex.Infof("request time: %v,%v", p.Meta.Path, time.Now().Sub(now))
}

func (s *ServeMux) handleLoop() {
	s.workGroup.Add(1)
	defer s.workGroup.Done()
	var op *rpc.Packet

	for {
		select {
		case op = <-s.workChan:
		case <-s.stopChan:
			return
		}

		logex.Info("comming:", op)

		handler := s.Handler.GetHandler(op.Meta.Path)
		go s.handlerWrap(handler, op, s.ctx)
	}
}

func (c *ServeMux) Write(b []byte) (n int, err error) {
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

func (c *ServeMux) WriteChan() <-chan *rpc.WriteItem {
	return c.writeChan
}
