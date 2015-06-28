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

	ToMqMux chan *rpc.Packet
}

func NewNotifyService(name string, args []string) service.Service {
	var c Config
	flagx.ParseFlag(&c, &flagx.FlagConfig{
		Name: name,
		Args: args,
	})
	return &NotifyService{
		Config:  &c,
		ToMqMux: make(chan *rpc.Packet, 100),
	}
}

func (a *NotifyService) Init() error {
	return model.Init(a.Config.Mongo)
}

func (a *NotifyService) Name() string {
	return Name
}

func (a *NotifyService) RunMqFetcher() error {
	handler := rpcmux.NewPathHandler()
	notify.InitMqHandler(handler)
	mux := rpcmux.NewClientMux(handler, nil)
	mux.Gtx = notify.NewContext(a.ToMqMux)
	linker := rpclink.NewTcpLink(mux)
	if err := rpc.Dial(a.MqAddr, linker); err != nil {
		return logex.Trace(err)
	}
	go func() {
		for {
			select {
			case data := <-a.ToMqMux:
				resp, err := mux.Send(data)
				if err != nil {
					logex.Error(err)
				}
				logex.Info("to mq mux", resp)
			}
		}
	}()
	return nil
}

func (a *NotifyService) Run() error {
	logex.Infof("[notify] listen on %v", a.Listen)
	handler := rpcmux.NewPathHandler()
	notify.Init(handler)
	return rpc.Listen(a.Listen, "tcp", func() rpc.Linker {
		mux := rpcmux.NewServeMux(notifyHandler, nil)
		return rpclink.NewTcpLink(mux)
	})
}
