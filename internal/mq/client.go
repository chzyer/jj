package mq

import (
	"reflect"
	"sync"
)

type MqClient struct {
	uuid     int
	mq       *Mq
	sub      map[string][]string
	gruad    sync.Mutex
	RespChan chan *Msg
	StopChan chan struct{}
}

func NewMqClient(mq *Mq) *MqClient {
	return &MqClient{
		mq:       mq,
		sub:      make(map[string][]string, 1<<10),
		RespChan: make(chan *Msg),
		StopChan: make(chan struct{}),
	}
}

func (c *MqClient) Name() string {
	return ""
}

func (c *MqClient) Publish(topic string, msg []byte) {
	c.mq.Publish(topic, msg)
}

func (c *MqClient) Topics() []string {
	return c.mq.Topics()
}

func (c *MqClient) Subscribe(topic, channel string) error {
	c.mq.Subscribe(c, topic, channel)
	c.gruad.Lock()
	c.sub[topic] = append(c.sub[topic], channel)
	c.gruad.Unlock()
	return nil
}

func (c *MqClient) Unsubscribe(topic, channel string) error {
	c.mq.Unsubscribe(c, topic, channel)
	c.gruad.Lock()
	channels := c.sub[topic]
	idx := -1
	for i := range channels {
		if channels[i] == channel {
			idx = i
			break
		}
	}
	if idx > 0 {
		c.sub[topic] = append(channels[:idx], channels[idx+1:]...)
	}
	c.gruad.Unlock()
	return nil
}

func (c *MqClient) ToSelectCase() *reflect.SelectCase {
	return &reflect.SelectCase{
		Dir:  reflect.SelectSend,
		Chan: reflect.ValueOf(c.RespChan),
	}
}

func (c *MqClient) Stop() {
	c.gruad.Lock()
	defer c.gruad.Unlock()

	for topic, channels := range c.sub {
		for i := range channels {
			c.mq.Unsubscribe(c, topic, channels[i])
		}
	}
	close(c.StopChan)
}
