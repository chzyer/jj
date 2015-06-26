package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/bobappleyard/readline"
	"github.com/chzyer/flagx"
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
	flagx.Parse(&c)
	return &c
}

func usage() {
	println(`subscribe <topic> <channel>
unsubscribe <topic> <channel>
publish <topic> <message>`)
}

func getTopics(mux *rpcmux.ClientMux) []string {
	var subresp []string
	if err := mux.Call(mq.PathTopics, nil, &subresp); err != nil {
		logex.Fatal(err)
	}
	return subresp
}

func subscribe(mux *rpcmux.ClientMux, topic, channel string) {
	var subresp string
	if err := mux.Call(mq.PathSubscribe, &mq.TopicChannel{
		Topic:   topic,
		Channel: channel,
	}, &subresp); err != nil {
		logex.Fatal(err)
	}
}

func publish(mux *rpcmux.ClientMux, topic, msg string) {
	var resp string
	if err := mux.Call(mq.PathPublish, &mq.PublishParams{
		Topic: topic,
		Data:  msg,
	}, &resp); err != nil {
		logex.Fatal(err)
	}
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

	for {
		cmd, err := readline.String("> ")
		if err != nil {
			println(err.Error())
			os.Exit(1)
		}
		readline.AddHistory(cmd)
		switch cmd {
		case "topics":
			fmt.Println(getTopics(mux))
			continue
		}

		idx := strings.Index(cmd, " ")
		if idx < 0 {
			usage()
			continue
		}
		action := cmd[:idx]
		cmd = cmd[idx+1:]

		idx = strings.Index(cmd, " ")
		topic := cmd[:idx]
		cmd = cmd[idx+1:]

		switch action {
		case "subscribe":
			subscribe(mux, topic, cmd)
		case "publish":
			publish(mux, topic, cmd)
		default:
			usage()
			continue
		}
	}
}

func OnReceiveMsg(w rpc.ResponseWriter, req *rpc.Request) {
	var msg mq.MsgParams
	if err := req.Params(&msg); err != nil {
		logex.Error(err)
		return
	}
	println(fmt.Sprintf("topic: %v; msg: %v \n", msg.Topic, msg.Data))
	readline.RefreshLine()
}
