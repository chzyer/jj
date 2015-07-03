package mgr

import "github.com/jj-io/jj/rpc"

type Context struct {
	ToMqChan chan *rpc.Packet
}

func NewContext(tomq chan *rpc.Packet) *Context {
	return &Context{
		ToMqChan: tomq,
	}
}

func (c *Context) Close() {}

func getCtx(req *rpc.Request) *Context {
	return req.Gtx.(*Context)
}
