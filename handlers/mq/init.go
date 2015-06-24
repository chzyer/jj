package mq

import (
	"github.com/jj-io/jj/internal/mq"
	"github.com/jj-io/jj/rpc"
)

var mqobj *mq.Mq
var (
	PathSubscribe   = "subscribe"
	PathUnsubscribe = "unsubscribe"
)

func InitMq() {
	mqobj = mq.NewMq()
}

func Init(h rpc.Handler) {
	h.HandleFunc(PathSubscribe, SubscribeHandler)
}

func getCtx(req *rpc.Request) *Context {
	return req.Gtx.(*Context)
}
