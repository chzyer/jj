package mq

import "github.com/jj-io/jj/rpc"

func TopicsHandler(w rpc.ResponseWriter, req *rpc.Request) {
	topics := getCtx(req).Topics()
	w.Response(topics)
}

func ChannelsHandler(w rpc.ResponseWriter, req *rpc.Request) {
	var params TopicChannel
	if err := req.Params(&params); err != nil {
		w.Error(err)
		return
	}

	channels := getCtx(req).Channels(params.Topic)
	w.Response(channels)
}
