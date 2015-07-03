package notify

import (
	"sync"

	"gopkg.in/logex.v1"

	"github.com/jj-io/jj/internal/mq"
	"github.com/jj-io/jj/rpc"
	"github.com/jj-io/jj/rpc/rpcmux"
)

type CtxGroup struct {
	group map[string]*Context
	gruad sync.Mutex
}

func NewCtxGroup() *CtxGroup {
	return &CtxGroup{
		group: make(map[string]*Context),
	}
}

func (cg *CtxGroup) GetCtx(name string) *Context {
	cg.gruad.Lock()
	ctx := cg.group[name]
	cg.gruad.Unlock()
	return ctx
}

func (cg *CtxGroup) AddCtx(name string, c *Context) {
	cg.gruad.Lock()
	cg.group[name] = c
	cg.gruad.Unlock()
}

func (cg *CtxGroup) DelCtx(name string) {
	cg.gruad.Lock()
	delete(cg.group, name)
	cg.gruad.Unlock()
}

type MqCtx struct {
	ToMqMux chan *rpc.Packet
}

func NewMqCtx(mqMux chan *rpc.Packet) *MqCtx {
	return &MqCtx{
		ToMqMux: mqMux,
	}
}

func (c *MqCtx) Close() {}

type Context struct {
	group       *CtxGroup
	uid         string
	ToMqMux     chan *rpc.Packet
	IncomingMsg chan *mq.Msg
}

func NewContext(group *CtxGroup, mqMux chan *rpc.Packet) *Context {
	ctx := &Context{
		group:       group,
		ToMqMux:     mqMux,
		IncomingMsg: make(chan *mq.Msg),
	}
	return ctx
}

func (c *Context) HandleIncomingLoop(mux *rpcmux.ClientMux) {
	var msg *mq.Msg
	for {
		select {
		case msg = <-c.IncomingMsg:
		}
		logex.Debug("some msg is coming:", msg)
		mux.WritePacket(rpc.NewReqPacket("test", msg))
	}
}

func (c *Context) OnIncomingMsg(msg *mq.Msg) {
	c.IncomingMsg <- msg
}

// call from mq handler
func (c *Context) Dispatch(tc *mq.TopicChannel, msg *mq.Msg) {
	targetCtx := c.group.GetCtx(tc.String())
	if targetCtx == nil {
		logex.Errorf("dispatch msg not found, %v", msg)
		return
	}
	targetCtx.OnIncomingMsg(msg)
}

func (c *Context) Subscribe(uid string) {
	c.uid = uid
	c.group.AddCtx(c.uid, c)
}

func (c *Context) Unsubscribe() {
	if c.uid == "" {
		return
	}
	c.group.DelCtx(c.uid)
}

func (c *Context) Close() {
	c.Unsubscribe()
}

func getCtx(req *rpc.Request) *Context {
	return req.Gtx.(*Context)
}
