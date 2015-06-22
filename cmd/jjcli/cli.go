package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/chzyer/reflag"
	"github.com/jj-io/jj/rpc/rpcapi"
	"github.com/jj-io/jj/rpc/rpcenc"
	"github.com/jj-io/jj/rpc/rpclink"
	"github.com/jj-io/jj/rpc/rpcmux"
	"github.com/jj-io/jj/rpc/rpcprot"
	"gopkg.in/logex.v1"
)

type Config struct {
	MgrHost string `flag:"[0];def=:8682"`
}

func NewConfig() *Config {
	var c Config
	reflag.Parse(&c)
	return &c
}

var mux *rpcmux.ClientMux
var jsonEnc = rpcenc.NewJSONEncoding()

func process(cmd string) error {
	resp, err := mux.Send(&rpcprot.Packet{
		Meta: rpcprot.NewMeta(cmd),
	})
	if err != nil {
		return logex.Trace(err)
	}
	if resp.Meta.Error != "" {
		println(resp.Meta.Error)
		return nil
	}
	var v interface{}
	if err := resp.Data.Decode(jsonEnc, &v); err != nil {
		logex.Error(err)
		return nil
	}
	fmt.Println(v)
	return nil
}

func main() {
	c := NewConfig()
	mux = rpcmux.NewClientMux()
	tcpLink := rpclink.NewTcpLink(mux)
	if err := rpcapi.Dial(c.MgrHost, tcpLink); err != nil {
		logex.Fatal(err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	go func() {
		<-mux.GetStopChan()
		os.Exit(1)
	}()
	print("please input: ")
	for scanner.Scan() {
		err := process(scanner.Text())
		if err != nil {
			println("bye!")
			os.Exit(0)
		}
		print("please input: ")
	}
}
