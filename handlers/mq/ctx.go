package mq

import (
	"github.com/jj-io/jj/internal/mq"
	"github.com/jj-io/jj/rpc"
)

type Context struct {
	*mq.MqClient
}

func NewContext() rpc.Context {
	return &Context{}
}
