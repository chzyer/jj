package rpcmux

import (
	"time"

	"github.com/jj-io/jj/rpc"
)

var (
	RouterDebugPing  = "debug.ping"
	RouterDebugSleep = "debug.sleep"
	RouterHelp       = "help"
)

func InitDebugHandler(handler *Handler) {
	handler.HandleFunc(RouterDebugPing, Ping)
	handler.HandleFunc(RouterDebugSleep, Sleep)
	handler.HandleFunc(RouterHelp, Help)
}

func Help(w rpc.ResponseWriter, req *Request) {
	list := w.(*responseWriter).routerList()
	w.Response(list)
}

func Ping(w rpc.ResponseWriter, data *Request) {
	w.Response("pong")
}

func Sleep(w rpc.ResponseWriter, data *Request) {
	var params string
	if err := data.Params(&params); err != nil {
		w.Error(err)
		return
	}

	duration, err := time.ParseDuration(params)
	if err != nil {
		w.Errorf("%v: %v", err, params)
		return
	}

	time.Sleep(duration)
	w.Responsef("sleep %v", duration)
}
