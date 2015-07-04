package main

import (
	"os"

	"gopkg.in/logex.v1"

	"github.com/chzyer/flagx"
	"github.com/jj-io/jj/internal/rl"
	"github.com/jj-io/jj/rpc"
	"github.com/jj-io/jj/rpc/rpclink"
	"github.com/jj-io/jj/rpc/rpcmux"
)

type Config struct {
	Command []string `flag:"cmd"`
	Host    string   `flag:"[0];def=:8683"`
}

func NewConfig() *Config {
	var c Config
	flagx.Parse(&c)
	return &c
}

func init() {
	rl.Init()
}

func main() {
	c := NewConfig()
	mux := rpcmux.NewClientMux(nil, nil)
	tcpLink := rpclink.NewTcpLink(mux)
	if err := rpc.Dial(c.Host, tcpLink); err != nil {
		logex.Fatal(err)
	}

	go func() {
		<-mux.GetStopChan()
		os.Exit(1)
	}()

	for {
		l := rl.Readline(">")
		if err := process(mux, l); err != nil {
			println("bye")
			os.Exit(1)
		}
	}
}

func process(mux *rpcmux.ClientMux, l string) error {

	return nil
}
