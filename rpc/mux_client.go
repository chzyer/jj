package rpc

import (
	"sync"
	"time"

	"io"

	"gopkg.in/logex.v1"
)

var _ Mux = &ClientMux{}

type ClientMux struct {
	encoding    Encoding
	respChan    chan *Operation
	writeChan   chan<- *WriteOp
	global      []*WriteOp
	globalGuard sync.Mutex
}

func NewClientMux() *ClientMux {
	cm := &ClientMux{
		encoding: MsgPackEncoding{},
		respChan: make(chan *Operation, 10),
	}
	go cm.respLoop()
	return cm
}

func (c *ClientMux) Read(prot Protocol, buf []byte) error {
	var data *Operation
	if err := prot.ReadWithEncoding(c.encoding, buf, &data); err != nil {
		return logex.Trace(err)
	}
	c.respChan <- data
	return nil
}

func (c *ClientMux) respLoop() {
	var (
		o  *Operation
		op *WriteOp
	)
	for {
		op = nil
		select {
		case o = <-c.respChan:
		}

		c.globalGuard.Lock()
		for i := 0; i < len(c.global); i++ {
			if c.global[i].Data.Seq == o.Seq {
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
			logex.Info("operation not found:", o)
			continue
		}

		op.resp <- o
	}
}

func (c *ClientMux) SetWriteChan(ch chan<- *WriteOp) {
	c.writeChan = ch
}

func (c *ClientMux) Write(w *WriteOp) (op *Operation, err error) {
	w.resp = make(chan *Operation)
	w.err = make(chan error)
	c.globalGuard.Lock()
	c.global = append(c.global, w)
	c.globalGuard.Unlock()
	c.writeChan <- w
	select {
	case op = <-w.resp:
	case err = <-w.err:
		err = logex.Trace(err)
	case <-time.After(time.Second):
		err = logex.Trace(io.ErrNoProgress)
	}
	return
}
