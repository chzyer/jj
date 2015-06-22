package rpcmux

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/jj-io/jj/rpc"
	"github.com/jj-io/jj/rpc/rpcenc"
	"github.com/jj-io/jj/rpc/rpclink"
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
	Enc  rpc.Encoding
	Data *rpcprot.Data
	Meta *rpcprot.Meta
}

func NewRequest(p *rpcprot.Packet, bodyEnc rpc.Encoding) *Request {
	return &Request{
		Enc:  bodyEnc,
		Meta: p.Meta,
		Data: p.Data,
	}
}

func (r *Request) Params(v interface{}) error {
	return r.Data.Decode(r.Enc, v)
}

type HandlerFunc func(rpc.ResponseWriter, *Request)

var _ rpclink.Mux = &ClientMux{}

// single-conn request multiplexer
type ServeMux struct {
	prot        rpcprot.Protocol
	metaEnc     rpc.Encoding
	bodyEnc     rpc.Encoding
	useEncoding bool
	state       state
	workChan    chan *rpcprot.Packet
	workGroup   sync.WaitGroup
	stopChan    chan struct{}
	writeChan   chan *rpclink.WriteItem
	handlerMap  map[string]HandlerFunc
}

func NewServeMux() *ServeMux {
	sm := &ServeMux{
		metaEnc:    rpcenc.NewJSONEncoding(),
		bodyEnc:    rpcenc.NewJSONEncoding(),
		stopChan:   make(chan struct{}),
		workChan:   make(chan *rpcprot.Packet, 10),
		handlerMap: make(map[string]HandlerFunc),
		writeChan:  make(chan *rpclink.WriteItem),
	}
	InitDebugHandler(sm)
	go sm.handleLoop()
	return sm
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

func (s *ServeMux) HandleFunc(path string, handlerFunc HandlerFunc) {
	s.handlerMap[path] = handlerFunc
}

func (s *ServeMux) Send(p *rpcprot.Packet) error {
	err := s.prot.Write(s.metaEnc, s.bodyEnc, p)
	if err != nil {
		return logex.Trace(err)
	}
	return nil
}

func (s *ServeMux) Handle(buf *bytes.Buffer) error {
	var p rpcprot.Packet
	if err := s.prot.Read(buf, s.metaEnc, &p); err != nil {
		return logex.Trace(err)
	}
	s.workChan <- &p
	return nil
}

func (s *ServeMux) Close() {
	close(s.stopChan)
	s.workGroup.Wait()
}

func (s *ServeMux) handlerWrap(h HandlerFunc, p *rpcprot.Packet, bodyEnc rpc.Encoding) {
	now := time.Now()
	h(NewResponseWriter(s, p), NewRequest(p, bodyEnc))
	logex.Infof("request time: %v,%v", p.Meta.Path, time.Now().Sub(now))
}

func (s *ServeMux) handleLoop() {
	s.workGroup.Add(1)
	defer s.workGroup.Done()
	var op *rpcprot.Packet

	for {
		select {
		case op = <-s.workChan:
		case <-s.stopChan:
			return
		}

		logex.Info("comming:", op)
		handler := s.handlerMap[op.Meta.Path]
		if handler == nil {
			handler = NotFoundHandler
			logex.Warn("unknown path: ", op.Meta.Path)
		}
		go s.handlerWrap(handler, op, s.bodyEnc)
	}
}

func NotFoundHandler(w rpc.ResponseWriter, data *Request) {
	w.ErrorInfo(fmt.Sprintf("path '%v' not found", data.Meta.Path))
}

type responseWriter struct {
	s  *ServeMux
	op *rpcprot.Packet
}

func NewResponseWriter(s *ServeMux, packet *rpcprot.Packet) *responseWriter {
	r := &responseWriter{
		s:  s,
		op: packet,
	}
	return r
}

func (w *responseWriter) Response(data interface{}) error {
	return w.s.Send(&rpcprot.Packet{
		Meta: &rpcprot.Meta{
			Seq: w.op.Meta.Seq,
		},
		Data: rpcprot.NewData(data),
	})
}

func (w *responseWriter) ErrorInfo(info string) error {
	logex.Error(info)
	return logex.Trace(w.s.Send(&rpcprot.Packet{
		Meta: rpcprot.NewMetaError(w.op.Meta.Seq, info),
	}))
}

func (w *responseWriter) Error(err error) error {
	logex.Error(err)
	return logex.Trace(w.s.Send(&rpcprot.Packet{
		Meta: rpcprot.NewMetaError(w.op.Meta.Seq, err.Error()),
	}))
}

func (c *ServeMux) Write(b []byte) (n int, err error) {
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

func (c *ServeMux) WriteChan() <-chan *rpclink.WriteItem {
	return c.writeChan
}
