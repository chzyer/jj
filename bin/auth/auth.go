package main

import (
	"io"
	"net/http"
	"time"

	"github.com/chzyer/reflag"
	"gopkg.in/logex.v1"
)

type Config struct {
	Listen       string        `flag:"def=:8681;usage=listen port"`
	ReadTimeout  time.Duration `flag:"def=10s;usage=read timeout"`
	WriteTimeout time.Duration `flag:"def=1m;usage=write timeout"`
}

func NewConfig() *Config {
	var c Config
	reflag.Parse(&c)
	return &c
}

func Register(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "hello\n")
}

func Login(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "login")
}

func main() {
	conf := NewConfig()
	mux := http.NewServeMux()
	mux.HandleFunc("/user/register.v1", Register)
	mux.HandleFunc("/user/login.v1", Login)

	srv := &http.Server{
		Addr:         conf.Listen,
		Handler:      mux,
		ReadTimeout:  conf.ReadTimeout,
		WriteTimeout: conf.WriteTimeout,
	}
	err := srv.ListenAndServe()
	if err != nil {
		logex.Fatal(err)
	}
}
