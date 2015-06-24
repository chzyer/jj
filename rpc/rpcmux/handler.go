package rpcmux

import (
	"fmt"

	"gopkg.in/logex.v1"

	"github.com/jj-io/jj/rpc"
)

type HandlerFunc func(rpc.ResponseWriter, *Request)

type Handler struct {
	handlerMap map[string]HandlerFunc
}

func NewHandler() *Handler {
	h := &Handler{
		handlerMap: make(map[string]HandlerFunc),
	}
	InitDebugHandler(h)
	return h
}

func (h *Handler) HandleFunc(path string, handlerFunc HandlerFunc) {
	if h == nil {
		return
	}
	h.handlerMap[path] = handlerFunc
}

func (h *Handler) GetHandler(path string) (handler HandlerFunc) {
	if h != nil {
		handler = h.handlerMap[path]
	}
	if handler == nil {
		handler = NotFoundHandler
		logex.Warn("unknown path: ", path)
	}
	return handler
}

func NotFoundHandler(w rpc.ResponseWriter, data *Request) {
	w.ErrorInfo(fmt.Sprintf("path '%v' not found", data.Meta.Path))
}
