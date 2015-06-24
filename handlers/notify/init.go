package notify

import (
	"github.com/jj-io/jj/model"
	"github.com/jj-io/jj/rpc"
	"github.com/jj-io/jj/rpc/rpcenc"
)

func Init(h rpc.Handler) {
	h.HandleFunc("init", InitHandler)
}

type InitParams struct {
	Uid string `json:"uid"`
}

func InitHandler(w rpc.ResponseWriter, req *rpc.Request) {
	var params InitParams
	if err := req.Params(&params); err != nil {
		w.Error(err)
		return
	}

	token, err := model.Models.User.GetToken(params.Uid)
	if err != nil {
		w.Error(err)
		return
	}

	enc, err := rpcenc.NewAesEncoding(req.Ctx.BodyEnc, []byte(token))
	if err != nil {
		w.Error(err)
		return
	}

	w.Response("success")
	req.Ctx.BodyEnc = enc
	return
}
