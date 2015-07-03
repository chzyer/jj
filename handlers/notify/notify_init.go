package notify

import (
	"github.com/jj-io/jj/handlers/mq"
	"github.com/jj-io/jj/model"
	"github.com/jj-io/jj/rpc"
	"github.com/jj-io/jj/rpc/rpcenc"
	"gopkg.in/logex.v1"
)

var (
	ErrEmptyDevice = logex.Define("device is empty")
	ErrEmptyUid    = logex.Define("uid is empty")
)

func Init(h rpc.Handler) {
	h.HandleFunc("init", InitHandler)
}

type InitParams struct {
	Uid    string `json:"uid"`
	Device string `json:"device"`
}

func InitHandler(w rpc.ResponseWriter, req *rpc.Request) {
	var params InitParams
	if err := req.Params(&params); err != nil {
		w.Error(err)
		return
	}

	var err error
	switch {
	case params.Uid == "":
		err = ErrEmptyUid
	case params.Device == "":
		err = ErrEmptyDevice
	}
	if err != nil {
		w.Error(err)
		return
	}

	token, err := model.Models.User.GetToken(params.Uid)
	if err != nil {
		w.Error(err)
		return
	}

	logex.Info("init:", token)

	enc, err := rpcenc.NewAesEncoding(req.Ctx.BodyEnc, []byte(token))
	if err != nil {
		w.Error(err)
		return
	}

	w.Response("success")
	req.Ctx.BodyEnc = enc

	tc := &mq.TopicChannel{
		Topic:   "to:" + params.Uid,
		Channel: params.Device,
	}
	getCtx(req).Subscribe(tc.String())
	packet := rpc.NewReqPacket(mq.PathSubscribe, tc)
	req.Gtx.(*Context).ToMqMux <- packet
	println("write to mq success")
	return
}
