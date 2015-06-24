package main

import (
	"net/url"

	"github.com/jj-io/jj/handlers/auth"
	"github.com/jj-io/jj/httprpc"
	"gopkg.in/logex.v1"
)

type Call struct {
	client *httprpc.Client
}

func NewCall(c *httprpc.Client) *Call {
	return &Call{
		client: c,
	}
}

func (c *Call) Register(email, pswd string) (resp *auth.RegisterResp, err error) {
	if err := c.client.Call(auth.RouterRegister, url.Values{
		"email":  {email},
		"secret": {pswd},
	}, &resp); err != nil {
		logex.Fatal(err)
	}
	return
}

func (c *Call) Login(email, pswd string) (loginResp *auth.LoginResp, err error) {
	if err := c.client.Call(auth.RouterLogin, url.Values{
		"email":  {email},
		"secret": {pswd},
	}, &loginResp); err != nil {
		logex.Fatal(err)
	}
	if loginResp.Result != 200 {
		return nil, logex.NewError(loginResp.Reason)
	}

	return
}
