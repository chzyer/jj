package rpc

type Request struct {
	Gtx  Context
	Ctx  *EncContext
	Data *Data
	Meta *Meta
}

func NewRequest(p *Packet, ctx *EncContext, gtx Context) *Request {
	return &Request{
		Gtx:  gtx,
		Ctx:  ctx,
		Meta: p.Meta,
		Data: p.Data,
	}
}

func (r *Request) Params(v interface{}) error {
	return r.Data.Decode(r.Ctx.BodyEnc, v)
}
