package notify

import (
	"time"

	"github.com/jj-io/jj/handlers/notify"
	"github.com/jj-io/jj/model"
	"github.com/jj-io/jj/rpc"
	"github.com/jj-io/jj/rpc/rpcapi"
	"github.com/jj-io/jj/rpc/rpclink"
	"github.com/jj-io/jj/rpc/rpcmux"
	"github.com/jj-io/jj/service"

	"github.com/chzyer/reflag"
	"gopkg.in/logex.v1"
)

var Name = "notify"

type Config struct {
	Mongo        string        `flag:"def=localhost:3000/jj"`
	Listen       string        `flag:"def=:8683;usage=listen port"`
	ReadTimeout  time.Duration `flag:"def=10s;usage=read timeout"`
	WriteTimeout time.Duration `flag:"def=1m;usage=write timeout"`
}

type NotifyService struct {
	*Config
}

func NewNotifyService(name string, args []string) service.Service {
	var c Config
	reflag.ParseFlag(&c, &reflag.FlagConfig{
		Name: name,
		Args: args,
	})
	return &NotifyService{
		Config: &c,
	}
}

func (a *NotifyService) Init() error {
	return model.Init(a.Config.Mongo)
}

func (a *NotifyService) Name() string {
	return Name
}

func (a *NotifyService) Run() error {
	logex.Infof("[notify] listen on %v", a.Listen)
	return rpcapi.Listen(a.Listen, "tcp", func() rpc.Linker {
		handler := rpcmux.NewPathHandler()
		mux := rpcmux.NewServeMux(handler)
		// fixme

		notify.Init(mux)
		return rpclink.NewTcpLink(mux)
	})
}
