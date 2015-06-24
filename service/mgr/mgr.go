package mgr

import (
	"time"

	"github.com/jj-io/jj/handlers/mgr"
	"github.com/jj-io/jj/rpc"
	"github.com/jj-io/jj/rpc/rpcapi"
	"github.com/jj-io/jj/rpc/rpclink"
	"github.com/jj-io/jj/rpc/rpcmux"
	"github.com/jj-io/jj/service"

	"github.com/chzyer/reflag"
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
	ReadTimeout  time.Duration `flag:"def=10s;usage=read timeout"`
	WriteTimeout time.Duration `flag:"def=1m;usage=write timeout"`
}

type MgrService struct {
	*Config
}

func NewMgrService(name string, args []string) service.Service {
	var c Config
	reflag.ParseFlag(&c, &reflag.FlagConfig{
		Name: name,
		Args: args,
	})
	return &MgrService{
		Config: &c,
	}
}

func (a *MgrService) Name() string { return Name }

func (a *MgrService) Run() error {
	logex.Infof("[mgr] listen on %v", a.Listen)
	return rpcapi.Listen(a.Listen, "tcp", func() rpc.Linker {
		mux := rpcmux.NewServeMux(mgrHandler, nil)
		return rpclink.NewTcpLink(mux)
	})
}
