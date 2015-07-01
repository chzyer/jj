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

func InitDebugHandler(handler rpc.Handler) {
	handler.HandleFunc(RouterDebugPing, PingHandler)
	handler.HandleFunc(RouterDebugSleep, SleepHandler)
	handler.HandleFunc(RouterHelp, HelpHandler)
}

func HelpHandler(w rpc.ResponseWriter, req *rpc.Request) {
	list := w.(*responseWriter).routerList()
	w.Response(list)
}

func PingHandler(w rpc.ResponseWriter, data *rpc.Request) {
	w.Response("pong")
}

func SleepHandler(w rpc.ResponseWriter, data *rpc.Request) {
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
