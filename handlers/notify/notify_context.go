package notify

import "github.com/jj-io/jj/rpc"

type Context struct {
	ToMqMux    chan *rpc.Packet
	ToDispatch chan *rpc.Packet
}

func NewContext(mqMux chan *rpc.Packet) *Context {
	return &Context{
		ToMqMux: mqMux,
	}
}

func getCtx(req *rpc.Request) *Context {
	return req.Gtx.(*Context)
}