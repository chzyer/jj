package notify

import (
	"time"

	"github.com/jj-io/jj/handlers/notify"
	"github.com/jj-io/jj/model"
	"github.com/jj-io/jj/rpc"
	"github.com/jj-io/jj/rpc/rpclink"
	"github.com/jj-io/jj/rpc/rpcmux"
	"github.com/jj-io/jj/service"

	"github.com/chzyer/flagx"
	"gopkg.in/logex.v1"
)

var (
	Name          = "notify"
	Desc          = "send message to client, consumers to mq service"
	notifyHandler = rpcmux.NewPathHandler()
)

func init() {
	notify.Init(notifyHandler)
}

type Config struct {
	Mongo        string        `flag:"def=localhost:3000/jj"`
	Listen       string        `flag:"def=:8683;usage=listen port"`
	MqAddr       string        `flag:"def=:8684;usage=mq addr"`
	ReadTimeout  time.Duration `flag:"def=10s;usage=read timeout"`
	WriteTimeout time.Duration `flag:"def=1m;usage=write timeout"`
}

type NotifyService struct {
	*Config
}

func NewNotifyService(name string, args []string) service.Service {
	var c Config
	flagx.ParseFlag(&c, &flagx.FlagConfig{
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

func (a *NotifyService) RunMqFetcher() {
	handler := rpcmux.NewPathHandler()
	notify.Init(handler)
	mux := rpcmux.NewClientMux(handler, nil)
	linker := rpclink.NewTcpLink(mux)
	rpc.Dial(a.MqAddr, linker)

}

func (a *NotifyService) Run() error {
	logex.Infof("[notify] listen on %v", a.Listen)
	return rpc.Listen(a.Listen, "tcp", func() rpc.Linker {
		mux := rpcmux.NewServeMux(notifyHandler, nil)
		return rpclink.NewTcpLink(mux)
	})
}
