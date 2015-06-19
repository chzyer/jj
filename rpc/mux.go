package rpc

import (
	"sync"

	"gopkg.in/logex.v1"
)

const state int
const (
	stateInit state = iota
	stateStart
)

// single-user request multiplexer
type ServeMux struct {
	encoding  Encoding
	state     state
	workChan  chan *Data
	workGroup sync.WaitGroup
}

type Data struct {
	Version int
	Path    string
}

func (s *ServeMux) Read(prot Protocol, buf []byte) error {
	var data *Data
	if err := prot.ReadWithEncoding(s.encoding, buf, &data); err != nil {
		return logex.Trace(err)
	}
	s.workChan <- data
	return nil
}

func (s *ServeMux) handleLoop() {
	s.workGroup.Add(1)
	defer s.workGroup.Done()

	for {
		select {
		case <-s.workChan:
		}
	}
}
