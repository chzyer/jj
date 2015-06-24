package mq

import (
	"github.com/jj-io/jj/rpc"
	"gopkg.in/logex.v1"
)

var (
	ErrTopicChannelEmpty = logex.Define("topic or channel is empty")
)

type TopicChannel struct {
	Topic   string
	Channel string
}

func SubscribeHandler(w rpc.ResponseWriter, req *rpc.Request) {
	var tc TopicChannel
	if err := req.Params(&tc); err != nil {
		w.Error(err)
		return
	}

	if tc.Topic == "" || tc.Channel == "" {
		w.Error(ErrTopicChannelEmpty)
		return
	}

	ctx := getCtx(req)
	if err := ctx.MqClient.Subscribe(tc.Topic, tc.Channel); err != nil {
		w.Error(err)
		return
	}
	w.Response("success")
}
