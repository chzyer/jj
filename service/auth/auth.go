package auth

import (
	"net/http"
	"time"

	"github.com/chzyer/reflag"
	"github.com/jj-io/jj/handlers/auth"
	"github.com/jj-io/jj/model"
	"github.com/jj-io/jj/service"
	"gopkg.in/logex.v1"
)

var Name = "auth"

type Config struct {
	Mongo        string        `flag:"def=localhost:3000/jj"`
	Listen       string        `flag:"def=:8681;usage=listen port"`
	ReadTimeout  time.Duration `flag:"def=10s;usage=read timeout"`
	WriteTimeout time.Duration `flag:"def=1m;usage=write timeout"`
}

type AuthService struct {
	*Config
}

func NewAuthService(name string, args []string) service.Service {
	var c Config
	reflag.ParseFlag(&c, &reflag.FlagConfig{
		Name: name,
		Args: args,
	})
	return &AuthService{
		Config: &c,
	}
}

func (a *AuthService) Init() error {
	return model.Init(a.Config.Mongo)
}

func (a *AuthService) Name() string {
	return Name
}

func (a *AuthService) Run() error {
	mux := http.NewServeMux()
	auth.InitHandler(mux)
	srv := &http.Server{
		Addr:         a.Listen,
		Handler:      mux,
		ReadTimeout:  a.ReadTimeout,
		WriteTimeout: a.WriteTimeout,
	}
	logex.Infof("[auth] listen on %v", a.Listen)
	err := srv.ListenAndServe()
	if err != nil {
		logex.Fatal(err)
	}
	return err
}
