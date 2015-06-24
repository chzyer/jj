package mq

import (
	"github.com/jj-io/jj/internal/mq"
	"github.com/jj-io/jj/rpc"
)

var mqobj *mq.Mq
var (
	PathSubscribe   = "subscribe"
	PathUnsubscribe = "unsubscribe"
	PathPublish     = "publish"
)

func InitMq() {
	mqobj = mq.NewMq()
}

func Init(h rpc.Handler) {
	h.HandleFunc(PathSubscribe, SubscribeHandler)
	h.HandleFunc(PathUnsubscribe, UnsubscribeHandler)
	h.HandleFunc(PathPublish, PublishHandler)
}

func getCtx(req *rpc.Request) *Context {
	return req.Gtx.(*Context)
}
