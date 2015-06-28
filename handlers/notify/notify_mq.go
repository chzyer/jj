package notify

import (
	hmq "github.com/jj-io/jj/handlers/mq"
	"github.com/jj-io/jj/internal/mq"
	"github.com/jj-io/jj/rpc"
	"gopkg.in/logex.v1"
)

func InitMqHandler(h rpc.Handler) {
	h.HandleFunc(hmq.PathMsg, MqMsgHandler)
}

func MqMsgHandler(w rpc.ResponseWriter, req *rpc.Request) {
	var msg mq.Msg
	if err := req.Params(&msg); err != nil {
		logex.Info(err)
		return
	}

	logex.Info(msg)
}
