package rpcmux

import (
	"fmt"

	"gopkg.in/logex.v1"

	"github.com/jj-io/jj/rpc"
)

type PathHandler struct {
	handlerMap map[string]rpc.HandlerFunc
}

func NewPathHandler() *PathHandler {
	h := &PathHandler{
		handlerMap: make(map[string]rpc.HandlerFunc),
	}
	InitDebugHandler(h)
	return h
}

func (h *PathHandler) ListPath() []string {
	var s []string
	for k := range h.handlerMap {
		s = append(s, k)
	}
	return s
}

func (h *PathHandler) HandleFunc(path string, handlerFunc rpc.HandlerFunc) {
	if h == nil {
		return
	}
	h.handlerMap[path] = handlerFunc
}

func (h *PathHandler) GetHandler(path string) (handler rpc.HandlerFunc) {
	if h != nil {
		handler = h.handlerMap[path]
	}
	if handler == nil {
		handler = NotFoundHandler
		logex.Warn("unknown path: ", path)
	}
	return handler
}

func NotFoundHandler(w rpc.ResponseWriter, data *rpc.Request) {
	w.ErrorInfo(fmt.Sprintf("path '%v' not found", data.Meta.Path))
}
