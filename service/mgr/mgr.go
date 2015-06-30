package mgr

import (
	"time"

	"github.com/jj-io/jj/handlers/mgr"
	"github.com/jj-io/jj/rpc"
	"github.com/jj-io/jj/rpc/rpclink"
	"github.com/jj-io/jj/rpc/rpcmux"
	"github.com/jj-io/jj/service"

	"github.com/chzyer/flagx"
	"gopkg.in/logex.v1"
)

var (
	Name       = "mgr"
	Desc       = "process requests from client"
	mgrHandler = rpcmux.NewPathHandler()
)

func init() {
	mgr.Init(mgrHandler)
}

type Config struct {
	Listen       string        `flag:"def=:8682;usage=listen port"`
	MqAddr       string        `flag:"def=:8684;usage=mq addr"`
	ReadTimeout  time.Duration `flag:"def=10s;usage=read timeout"`
	WriteTimeout time.Duration `flag:"def=1m;usage=write timeout"`
}

type MgrService struct {
	*Config
	ToMqChan chan *rpc.Packet
}

func NewMgrService(name string, args []string) service.Service {
	var c Config
	flagx.ParseFlag(&c, &flagx.FlagConfig{
		Name: name,
		Args: args,
	})
	return &MgrService{
		Config:   &c,
		ToMqChan: make(chan *rpc.Packet),
	}
}

func (a *MgrService) Name() string { return Name }

func (a *MgrService) runMq() error {
	handler := rpcmux.NewPathHandler()
	mux := rpcmux.NewClientMux(handler, nil)
	mux.Gtx = mgr.NewContext(a.ToMqChan)
	linker := rpclink.NewTcpLink(mux)
	if err := rpc.Dial(a.MqAddr, linker); err != nil {
		return logex.Trace(err)
	}
	a.sendToMqLoop(mux)
	return nil

}

func (a *MgrService) sendToMqLoop(mux *rpcmux.ClientMux) {
	var packet *rpc.Packet
	for {
		select {
		case packet = <-a.ToMqChan:
			resp, err := mux.Send(packet)
			if err != nil {
				logex.Error(err)
				continue
			}
			logex.Info(resp)
		}

	}
}

func (a *MgrService) Run() error {
	go a.runMq()
	logex.Infof("[mgr] listen on %v", a.Listen)
	return rpc.Listen(a.Listen, "tcp", func() rpc.Linker {
		mux := rpcmux.NewServeMux(mgrHandler, nil)
		mux.Gtx = mgr.NewContext(a.ToMqChan)
		return rpclink.NewTcpLink(mux)
	})
}
