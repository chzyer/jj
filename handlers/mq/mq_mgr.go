package mq

import "github.com/jj-io/jj/rpc"

func TopicsHandler(w rpc.ResponseWriter, req *rpc.Request) {
	topics := getCtx(req).Topics()
	w.Response(topics)
}
