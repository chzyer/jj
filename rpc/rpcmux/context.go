package rpcmux

import (
	"fmt"

	"github.com/jj-io/jj/rpc"
	"gopkg.in/logex.v1"
)

type responseWriter struct {
	s  *ServeMux
	op *rpc.Packet
}

func NewResponseWriter(s *ServeMux, packet *rpc.Packet) *responseWriter {
	r := &responseWriter{
		s:  s,
		op: packet,
	}
	return r
}

func (w *responseWriter) routerList() []string {
	var s []string
	for k := range w.s.Handler.handlerMap {
		s = append(s, k)
	}
	return s
}

func (w *responseWriter) Responsef(fmt_ string, obj ...interface{}) error {
	return w.Response(fmt.Sprintf(fmt_, obj...))
}

func (w *responseWriter) Response(data interface{}) error {
	return w.s.Send(&rpc.Packet{
		Meta: &rpc.Meta{
			Seq: w.op.Meta.Seq,
		},
		Data: rpc.NewData(data),
	})
}

func (w *responseWriter) error(str string) error {
	return w.s.Send(&rpc.Packet{
		Meta: rpc.NewMetaError(w.op.Meta.Seq, str),
	})
}

func (w *responseWriter) ErrorInfo(info string) error {
	logex.DownLevel(1).Error(info)
	return logex.Trace(w.error(info))
}

func (w *responseWriter) Errorf(fmt_ string, obj ...interface{}) error {
	errInfo := fmt.Sprintf(fmt_, obj...)
	logex.DownLevel(1).Error(errInfo)
	return logex.Trace(w.error(errInfo))
}

func (w *responseWriter) Error(err error) error {
	logex.DownLevel(1).Error(err)
	return logex.Trace(w.error(err.Error()))
}
