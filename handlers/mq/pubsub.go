package mq

import (
	"github.com/jj-io/jj/rpc"
	"gopkg.in/logex.v1"
)

var (
	ErrTopicEmpty   = logex.Define("topic is empty")
	ErrChannelEmpry = logex.Define("channel is emtpy")
	ErrMsgEmpty     = logex.Define("msg is empty")
)

type TopicChannel struct {
	Topic   string `json:"topic"`
	Channel string `json:"channel"`
}

func NewTopicChannel(req *rpc.Request) (*TopicChannel, error) {
	var tc TopicChannel
	if err := req.Params(&tc); err != nil {
		return nil, err
	}
	if tc.Topic == "" {
		return nil, ErrTopicEmpty
	}
	if tc.Channel == "" {
		return nil, ErrChannelEmpry
	}
	return &tc, nil
}

type PublishParams struct {
	Topic string `json:"topic"`
	Msg   string `json:"msg"`
}

func UnsubscribeHandler(w rpc.ResponseWriter, req *rpc.Request) {
	tc, err := NewTopicChannel(req)
	if err != nil {
		w.Error(err)
		return
	}
	if err := getCtx(req).Unsubscribe(tc.Topic, tc.Channel); err != nil {
		w.Error(err)
		return
	}
	w.Response("success")
}

func SubscribeHandler(w rpc.ResponseWriter, req *rpc.Request) {
	tc, err := NewTopicChannel(req)
	if err != nil {
		w.Error(err)
		return
	}
	if err := getCtx(req).Subscribe(tc.Topic, tc.Channel); err != nil {
		w.Error(err)
		return
	}
	w.Response("success")
}

func PublishHandler(w rpc.ResponseWriter, req *rpc.Request) {
	var params PublishParams
	if err := req.Params(&params); err != nil {
		w.Error(err)
		return
	}

	if params.Topic == "" {
		w.Error(ErrTopicEmpty)
		return
	}
	if params.Msg == "" {
		w.Error(ErrMsgEmpty)
		return
	}

	getCtx(req).Publish(params.Topic, []byte(params.Msg))
	w.Response("success")
	return
}
