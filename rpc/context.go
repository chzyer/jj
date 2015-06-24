package rpc

type Context struct {
	MetaEnc Encoding
	BodyEnc Encoding
}

func NewContext(metaEnc, bodyEnc Encoding) *Context {
	ctx := &Context{
		MetaEnc: metaEnc,
		BodyEnc: bodyEnc,
	}
	return ctx
}
