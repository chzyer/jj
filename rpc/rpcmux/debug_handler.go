package rpcmux

import (
	"time"

	"github.com/jj-io/jj/rpc"
	"github.com/jj-io/jj/rpc/rpcprot"
)

func InitDebugHandler(mux *ServeMux) {
	mux.HandleFunc("debug.ping", Ping)
	mux.HandleFunc("debug.sleep", Sleep)
}

func Ping(w rpc.ResponseWriter, data *rpcprot.Data) {
	w.Response("pong")
}

type SleepData struct {
	Second int `msgpack:"second"`
}

func Sleep(w rpc.ResponseWriter, data *rpcprot.Data) {
	time.Sleep(100 * time.Microsecond)
	w.Response("")
}
