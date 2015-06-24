package main

import (
	"time"

	"github.com/chzyer/reflag"
	"github.com/jj-io/jj/handlers/mq"
	"github.com/jj-io/jj/rpc"
	"github.com/jj-io/jj/rpc/rpcapi"
	"github.com/jj-io/jj/rpc/rpclink"
	"github.com/jj-io/jj/rpc/rpcmux"
	"gopkg.in/logex.v1"
)

type Config struct {
	Command []string `flag:"cmd"`
	Host    string   `flag:"[0];def=:8684"`
}

func NewConfig() *Config {
	var c Config
	reflag.Parse(&c)
	return &c
}

func main() {
	c := NewConfig()
	handler := rpcmux.NewPathHandler()
	handler.HandleFunc(mq.PathMsg, OnReceiveMsg)
	mux := rpcmux.NewClientMux(handler, nil)
	link := rpclink.NewTcpLink(mux)
	if err := rpcapi.Dial(c.Host, link); err != nil {
		logex.Fatal(err)
	}

	var subresp string
	if err := mux.Call(mq.PathSubscribe, &mq.TopicChannel{
		Topic:   "hello",
		Channel: "ch1",
	}, &subresp); err != nil {
		logex.Fatal(err)
	}

	var resp string
	if err := mux.Call(mq.PathPublish, &mq.PublishParams{
		Topic: "hello",
		Data:  "msg here!",
	}, &resp); err != nil {
		logex.Fatal(err)
	}

	time.Sleep(time.Second)
	println(resp, subresp)
}

func OnReceiveMsg(w rpc.ResponseWriter, req *rpc.Request) {
	var msg mq.MsgParams
	if err := req.Params(&msg); err != nil {
		logex.Error(err)
		return
	}
	println("comming:", msg.Topic, msg.Data)
}
