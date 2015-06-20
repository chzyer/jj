package mq

import "reflect"

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

func (m Msg) Clone(ch string) *Msg {
	m.channel = ch
	return &m
}

type MqClient struct {
	uuid     int
	mq       *Mq
	respChan chan *Msg
}

func NewMqClient(mq *Mq) *MqClient {
	return &MqClient{
		mq:       mq,
		respChan: make(chan *Msg),
	}
}

func (c *MqClient) Name() string {
	return ""
}

func (c *MqClient) Subscribe(topic, channel string) error {
	c.mq.Subscribe(c, topic, channel)
	return nil
}

func (c *MqClient) ToSelectCase() reflect.SelectCase {
	return reflect.SelectCase{
		Dir:  reflect.SelectSend,
		Chan: reflect.ValueOf(c.respChan),
	}
}
