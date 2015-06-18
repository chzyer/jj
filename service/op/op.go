package op

import (
	"time"

	"github.com/jj-io/jj/service"

	"github.com/chzyer/reflag"
	"gopkg.in/logex.v1"
)

var Name = "op"

type Config struct {
	Listen       string        `flag:"def=:8682;usage=listen port"`
	ReadTimeout  time.Duration `flag:"def=10s;usage=read timeout"`
	WriteTimeout time.Duration `flag:"def=1m;usage=write timeout"`
}

type OpService struct {
	*Config
}

func NewOpService(name string, args []string) service.Service {
	var c Config
	reflag.ParseFlag(&c, &reflag.FlagConfig{
		Name: name,
		Args: args,
	})
	return &OpService{
		Config: &c,
	}
}

func (a *OpService) Name() string { return Name }

func (a *OpService) Run() error {
	logex.Infof("[op] listen on %v", a.Listen)
	return nil
}
