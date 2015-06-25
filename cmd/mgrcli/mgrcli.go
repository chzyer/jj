package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bobappleyard/readline"
	"github.com/chzyer/flagx"
	"github.com/jj-io/jj/rpc"
	"github.com/jj-io/jj/rpc/rpcapi"
	"github.com/jj-io/jj/rpc/rpcenc"
	"github.com/jj-io/jj/rpc/rpclink"
	"github.com/jj-io/jj/rpc/rpcmux"
	"gopkg.in/logex.v1"
)

type Config struct {
	Command []string `flag:"cmd"`
	MgrHost string   `flag:"[0];def=:8682"`
}

func NewConfig() *Config {
	var c Config
	flagx.Parse(&c)
	return &c
}

var mux *rpcmux.ClientMux
var jsonEnc = rpcenc.NewJSONEncoding()

func getBody(data string) (obj interface{}) {
	buf := bytes.NewReader([]byte(data))
	if err := jsonEnc.Decode(buf, &obj); err != nil {
		logex.Error(err)
	}
	return
}

func process(cmd string) error {
	var path string
	var data string

	if idx := strings.Index(cmd, " "); idx < 0 {
		path = cmd
	} else {
		path = cmd[:idx]
		data = cmd[idx+1:]
	}

	if path == "bodyEnc" {
		enc, err := rpcenc.NewAesEncoding(mux.Ctx.BodyEnc, []byte(data))
		if err != nil {
			println("invalid enc key", data)
			return nil
		}
		mux.Ctx.BodyEnc = enc
		println("change bodyEnc to aes:", data)
		return nil
	}

	var body interface{}
	if data != "" {
		body = getBody(data)
		if body == nil {
			return nil
		}
	}

	packet := &rpc.Packet{
		Meta: rpc.NewReqMeta(path),
	}
	if body != nil {
		packet.Data = rpc.NewData(body)
	}

	resp, err := mux.Send(packet)
	if err != nil {
		return logex.Trace(err)
	}
	if resp.Meta.Error != "" {
		println(resp.Meta.Error)
		return nil
	}
	var v interface{}
	if err := resp.Data.Decode(mux.Ctx.BodyEnc, &v); err != nil {
		logex.Error(err)
		return nil
	}
	output(v)
	return nil
}

func output(v interface{}) {
	switch v := v.(type) {
	case []interface{}:
		fmt.Println("[")
		for _, s := range v {
			fmt.Print("    ", s, "\n")
		}
		fmt.Println("]")
	default:
		fmt.Println(v)
	}
}

var history = "/tmp/mgrcli.readline"

func main() {
	c := NewConfig()
	mux = rpcmux.NewClientMux(nil, nil)
	tcpLink := rpclink.NewTcpLink(mux)
	if err := rpcapi.Dial(c.MgrHost, tcpLink); err != nil {
		logex.Fatal(err)
	}

	for _, c := range c.Command {
		println(c)
		process(c)
	}

	go func() {
		<-mux.GetStopChan()
		os.Exit(1)
	}()
	if err := readline.LoadHistory(history); err != nil {
		logex.Error(err)
	}

	for {
		l, err := readline.String(">>> ")
		if err == io.EOF {
			break
		}
		if err != nil {
			println(err.Error())
			os.Exit(0)
		}
		if err := process(l); err != nil {
			println("bye!")
			os.Exit(1)
		}
		readline.AddHistory(l)
		if err := readline.SaveHistory(history); err != nil {
			logex.Error(err)
		}
	}
}
