package main

import (
	"github.com/jj-io/jj/handlers/mgr"
	"github.com/jj-io/jj/rpc"
	"github.com/jj-io/jj/rpc/rpcenc"
	"github.com/jj-io/jj/rpc/rpclink"
	"github.com/jj-io/jj/rpc/rpcmux"
	"github.com/jj-io/readline"
	"gopkg.in/logex.v1"
)

type MgrCli struct {
	Mux  *rpcmux.ClientMux
	Link *rpclink.TcpLink
	Addr []string
}

func NewMgrCli(mgrAddr []string) (*MgrCli, error) {
	cli := &MgrCli{
		Addr: mgrAddr,
	}
	cli.Mux = rpcmux.NewClientMux(nil, nil)
	cli.Link = rpclink.NewTcpLink(cli.Mux)
	if err := cli.connect(); err != nil {
		return nil, err
	}
	return cli, nil
}

func (m *MgrCli) connect() error {
	return logex.Trace(rpc.Dial(m.Addr[0], m.Link))
}

func (m *MgrCli) Call(method string, data, result interface{}) error {
	err := m.Mux.Call(method, data, result)
	if err.IsUserError {
		return logex.Trace(err)
	}
	readline.Exit(err.Error())
	return nil
}

func (m *MgrCli) Ping() error {
	var pong string
	if err := m.Call(rpcmux.RouterDebugPing, nil, &pong); err != nil {
		return err
	}
	if pong != "pong" {
		readline.Exit("unexcept ping result")
	}
	return nil
}

func (m *MgrCli) SendInit(uid string, token string) error {
	resp, err := m.Mux.Send(rpc.NewReqPacket(
		mgr.RouterInit,
		&mgr.InitParams{uid},
	))
	if err != nil {
		readline.Exit(err.Error())
	}
	if resp.Meta.Error != "" {
		return logex.NewError(resp.Meta.Error)
	}
	m.Mux.Ctx.BodyEnc, err = rpcenc.NewAesEncoding(m.Mux.Ctx.BodyEnc, []byte(token))
	if err != nil {
		readline.Exit("invalid token:" + err.Error())
	}
	return nil
}
