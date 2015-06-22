package rpcmux

import (
	"time"

	"github.com/jj-io/jj/rpc"
)

func InitDebugHandler(mux *ServeMux) {
	mux.HandleFunc("debug.ping", Ping)
	mux.HandleFunc("debug.sleep", Sleep)
	mux.HandleFunc("help", Help)
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
