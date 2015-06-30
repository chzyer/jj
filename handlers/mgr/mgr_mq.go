package mgr

import (
	"github.com/jj-io/jj/handlers/mq"
	"github.com/jj-io/jj/rpc"
	"gopkg.in/logex.v1"
)

var (
	PathSendMsg = "mq.sendmsg"
)

// only publish for debug
func InitMqHandler(h rpc.Handler) {
	h.HandleFunc(PathSendMsg, HandlerSend)
}

type SendParams struct {
	Uid string `json:"uid"`
	Msg string `json:"msg"`
}

func HandlerSend(w rpc.ResponseWriter, req *rpc.Request) {
	var params SendParams
	if err := req.Params(&params); err != nil {
		w.Error(err)
		return
	}
	ctx := getCtx(req)
	msg := &mq.PublishParams{
		Topic: "to:" + params.Uid,
		Data:  params.Msg,
	}
	logex.Struct(msg)
	ctx.ToMqChan <- rpc.NewReqPacket(mq.PathPublish, msg)
	w.Response("success")
}
