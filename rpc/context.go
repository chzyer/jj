package rpc

type Context interface {
	Close()
}

type GenContext func() Context

type EncContext struct {
	MetaEnc Encoding
	BodyEnc Encoding
}

func NewEncContext(metaEnc, bodyEnc Encoding) *EncContext {
	ctx := &EncContext{
		MetaEnc: metaEnc,
		BodyEnc: bodyEnc,
	}
	return ctx
}
