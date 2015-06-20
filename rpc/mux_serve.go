package rpc

import (
	"sync"

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

type Mux interface {
	Read(Protocol, []byte) error
	SetWriteChan(ch chan<- *WriteOp)
}

type ResponseWriter interface {
	Write(data interface{}) error
}

type HandlerFunc func(ResponseWriter, interface{})

// single-conn request multiplexer
type ServeMux struct {
	encoding    Encoding
	useEncoding bool
	state       state
	workChan    chan *Operation
	workGroup   sync.WaitGroup
	stopChan    chan struct{}
	writeChan   chan<- *WriteOp
	handlerMap  map[string]HandlerFunc
}

func NewServeMux() *ServeMux {
	sm := &ServeMux{
		encoding:   MsgPackEncoding{},
		stopChan:   make(chan struct{}),
		workChan:   make(chan *Operation, 10),
		handlerMap: make(map[string]HandlerFunc),
	}
	InitDebugHandler(sm)
	go sm.handleLoop()
	return sm
}

func (s *ServeMux) HandleFunc(path string, handlerFunc HandlerFunc) {
	s.handlerMap[path] = handlerFunc
}

type Operation struct {
	Version int         `msgpack:"version,omitempty"`
	Seq     int         `msgpack:"seq,omitempty"`
	Path    string      `msgpack:"path,omitempty"`
	Data    interface{} `msgpack:"data,omitempty"`
}

func (s *ServeMux) SetWriteChan(ch chan<- *WriteOp) {
	s.writeChan = ch
}

func (s *ServeMux) Write(data *Operation) error {
	op := &WriteOp{
		Encoding: s.encoding,
		Data:     data,
	}
	select {
	case s.writeChan <- op:
	default:
		return logex.Trace(ErrReceiveQuit)
	}
	return nil
}

func (s *ServeMux) Read(prot Protocol, buf []byte) error {
	var data *Operation
	if err := prot.ReadWithEncoding(s.encoding, buf, &data); err != nil {
		return logex.Trace(err)
	}
	s.workChan <- data
	return nil
}

func (s *ServeMux) Close() {
	close(s.stopChan)
	s.workGroup.Wait()
}

func (s *ServeMux) handleLoop() {
	s.workGroup.Add(1)
	defer s.workGroup.Done()
	var op *Operation

	for {
		select {
		case op = <-s.workChan:
		case <-s.stopChan:
			return
		}

		handler := s.handlerMap[op.Path]
		if handler != nil {
			go handler(&responseWriter{s, op}, op.Data)
		}
	}
}

type responseWriter struct {
	s  *ServeMux
	op *Operation
}

func (w *responseWriter) Write(data interface{}) error {
	return w.s.Write(&Operation{
		Seq:  w.op.Seq,
		Data: data,
	})
}
