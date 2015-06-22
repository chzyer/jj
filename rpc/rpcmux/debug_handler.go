package rpcmux

import (
	"time"

	"gopkg.in/logex.v1"

	"github.com/jj-io/jj/rpc"
)

func InitDebugHandler(mux *ServeMux) {
	mux.HandleFunc("debug.ping", Ping)
	mux.HandleFunc("debug.sleep", Sleep)
}

func Ping(w rpc.ResponseWriter, data *Request) {
	w.Response("pong")
}

type SleepData struct {
	Second      int `msgpack:"second"`
	Millisecond int `msgpack:"millisecond"`
}

func Sleep(w rpc.ResponseWriter, data *Request) {
	var params SleepData
	data.Params(&params)

	if params.Millisecond > 0 {
		logex.Infof("sleep %v ms", params.Millisecond)
		time.Sleep(time.Duration(params.Millisecond) * time.Microsecond)
	}

	if params.Second > 0 {
		logex.Infof("sleep %v second", params.Second)
		time.Sleep(time.Duration(params.Second) * time.Second)
	}

	w.Response(nil)
}
