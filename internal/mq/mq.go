package mq

import (
	"reflect"
	"sync"
)

type Mq struct {
	Topics  map[string]*Topic
	pubChan chan *Msg
}

func NewMq() *Mq {
	mq := &Mq{
		Topics:  make(map[string]*Topic),
		pubChan: make(chan *Msg, 10),
	}
	go mq.pubLoop()
	return mq
}

func (m *Mq) pubLoop() {
	var msg *Msg
	for {
		select {
		case msg = <-m.pubChan:
		}

		m.GetTopic(msg.Topic).Publish(msg.Data)
	}
}

func (m *Mq) GetTopic(name string) *Topic {
	topic, ok := m.Topics[name]
	if !ok {
		topic = NewTopic(name)
		m.Topics[name] = topic
	}
	return topic
}

func (m *Mq) Subscribe(client *MqClient, topic, channel string) {
	m.GetTopic(topic).AddSubscriber(channel, client)
	return
}

func (m *Mq) Unsubscribe(client *MqClient, topic, channel string) {
	m.GetTopic(topic).RemoveSubscriber(channel, client)
}

func (m *Mq) Publish(topic string, data []byte) {
	m.pubChan <- &Msg{
		Topic: topic,
		Data:  data,
	}
}

type Msg struct {
	Topic   string
	channel string
	Data    []byte
}

func (m *Msg) Channel() string {
	return m.channel
}

func (m Msg) Clone(ch string) *Msg {
	m.channel = ch
	return &m
}

type TopicChannel struct {
	Topic   string
	Channel string
}

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
