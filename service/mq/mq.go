package mq

import (
	"time"

	"gopkg.in/logex.v1"

	"github.com/chzyer/reflag"
	"github.com/jj-io/jj/handlers/mq"
	"github.com/jj-io/jj/rpc"
	"github.com/jj-io/jj/rpc/rpcapi"
	"github.com/jj-io/jj/rpc/rpclink"
	"github.com/jj-io/jj/rpc/rpcmux"
	"github.com/jj-io/jj/service"
)

var (
	Name      = "mq"
	Desc      = "message queue and boardcast message to each channel"
	mqHandler = rpcmux.NewPathHandler()
)

func init() {
	mq.Init(mqHandler)
}

type Config struct {
	Listen       string        `flag:"def=:8684;usage=listen port"`
	ReadTimeout  time.Duration `flag:"def=10s;usage=read timeout"`
	WriteTimeout time.Duration `flag:"def=1m;usage=write timeout"`
}

type MqService struct {
	*Config
}

func NewMqService(name string, args []string) service.Service {
	var c Config
	reflag.ParseFlag(&c, &reflag.FlagConfig{
		Name: name,
		Args: args,
	})
	return &MqService{
		Config: &c,
	}
}

func (a *MqService) Init() error {
	return nil
}

func (a *MqService) Name() string {
	return Name
}

func (a *MqService) Run() error {
	logex.Infof("[mq] listen on %v", a.Listen)
	mq.InitMq()
	return rpcapi.Listen(a.Listen, "tcp", func() rpc.Linker {
		var mux *rpcmux.ClientMux
		mux = rpcmux.NewClientMux(mqHandler, func() rpc.Context {
			return mq.NewContext(mux)
		})
		return rpclink.NewTcpLink(mux)
	})
}
