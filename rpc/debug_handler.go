package rpc

import "time"

func InitDebugHandler(mux *ServeMux) {
	mux.HandleFunc("ping", Ping)
	mux.HandleFunc("sleep", Sleep)
}

func Ping(w ResponseWriter, data interface{}) {
	w.Write("pong")
}

func Sleep(w ResponseWriter, data interface{}) {
	time.Sleep(100 * time.Microsecond)
	w.Write("1second")
}
