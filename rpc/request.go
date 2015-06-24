package rpc

type Request struct {
	Ctx  *Context
	Data *Data
	Meta *Meta
}

func NewRequest(p *Packet, ctx *Context) *Request {
	return &Request{
		Ctx:  ctx,
		Meta: p.Meta,
		Data: p.Data,
	}
}

func (r *Request) Params(v interface{}) error {
	return r.Data.Decode(r.Ctx.BodyEnc, v)
}
