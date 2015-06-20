package rpc

import (
	"sync"
	"time"

	"gopkg.in/logex.v1"
)

type state int

const (
	stateInit state = iota
	stateStart
)

type Mux interface {
	Read(Protocol, []byte) error
	SetWriteChan(ch chan<- *WriteOp)
}

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
	sm := &ServeMux{
		encoding: MsgPackEncoding{},
		stopChan: make(chan struct{}),
		workChan: make(chan *Operation, 10),
	}
	go sm.handleLoop()
	return sm
}

type Operation struct {
	Version int
	Seq     int
	Path    string
	Data    interface{}
}

func (s *ServeMux) SetWriteChan(ch chan<- *WriteOp) {
	s.writeChan = ch
}

func (s *ServeMux) Write(data *Operation) {
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

		go func(op *Operation) {
			switch op.Path {
			case "ping":
				s.Write(&Operation{
					Path: op.Path,
					Seq:  op.Seq,
					Data: "pong",
				})
			case "sleep":
				time.Sleep(100 * time.Microsecond)
				s.Write(&Operation{
					Path: op.Path,
					Seq:  op.Seq,
					Data: "1second",
				})
			}
		}(op)
	}
}
