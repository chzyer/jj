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

// single-conn request multiplexer
type ServeMux struct {
	encoding    Encoding
	useEncoding bool
	state       state
	workChan    chan *Operation
	workGroup   sync.WaitGroup
	stopChan    chan struct{}
	writeChan   chan<- *WriteOp
}

func NewServeMux() *ServeMux {
	return &ServeMux{
		encoding: MsgPackEncoding{},
		stopChan: make(chan struct{}),
		workChan: make(chan *Operation, 10),
	}
}

type Operation struct {
	Version int
	Seq     int
	Path    string
}

type Response struct {
	Seq  int
	Data interface{}
}

func (s *ServeMux) SetWriteChan(ch chan<- *WriteOp) {
	s.writeChan = ch
}

func (s *ServeMux) Write(data interface{}) {
	s.writeChan <- &WriteOp{
		Encoding: s.encoding,
		Data:     data,
	}
}

func (s *ServeMux) Read(prot Protocol, buf []byte) error {
	var data *Operation
	if err := prot.ReadWithEncoding(s.encoding, buf, &data); err != nil {
		return logex.Trace(err)
	}
	s.workChan <- data
	return nil
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

		switch op.Path {
		case "ping":
			s.Write(&Response{
				Seq:  op.Seq,
				Data: "pong",
			})
		}
	}
}
