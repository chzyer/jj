package mq

import (
	"fmt"
	"sync"

	"gopkg.in/logex.v1"
)

type Mq struct {
	topics  map[string]*Topic
	pubChan chan *Msg
	gruad   sync.Mutex
}

func NewMq() *Mq {
	mq := &Mq{
		topics:  make(map[string]*Topic),
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

func (m *Mq) Channels(topic string) (channels []string) {
	m.gruad.Lock()
	defer m.gruad.Unlock()
	t := m.topics[topic]
	if t == nil {
		return nil
	}

	channels = make([]string, len(t.Chans))
	for idx := range t.Chans {
		channels[idx] = t.Chans[idx].Name
	}
	return
}

func (m *Mq) Topics() (topics []string) {
	m.gruad.Lock()
	defer m.gruad.Unlock()
	topics = make([]string, 0, len(m.topics))
	for k := range m.topics {
		topics = append(topics, k)
	}
	return
}

func (m *Mq) GetTopic(name string) *Topic {
	m.gruad.Lock()
	defer m.gruad.Unlock()

	topic, ok := m.topics[name]
	if !ok {
		topic = NewTopic(name)
		m.topics[name] = topic
	}
	return topic
}

func (m *Mq) Subscribe(client *MqClient, topic, channel string) {
	logex.Debugf("subscribe %v, %v", topic, channel)
	m.GetTopic(topic).AddSubscriber(channel, client)
	return
}

func (m *Mq) Unsubscribe(client *MqClient, topic, channel string) {
	m.GetTopic(topic).RemoveSubscriber(channel, client)
	logex.Debugf("unsubscribe %v, %v", topic, channel)
}

func (m *Mq) Publish(topic string, data []byte) {
	m.pubChan <- &Msg{
		Topic: topic,
		Data:  data,
	}
}

type Msg struct {
	Topic   string
	Channel string
	Data    []byte
}

func (m *Msg) TopicChannel() *TopicChannel {
	return &TopicChannel{m.Topic, m.Channel}
}

func (m *Msg) String() string {
	return fmt.Sprintf("msg:%v:%v", m.Topic, m.Channel)
}

func (m Msg) Clone(ch string) *Msg {
	m.Channel = ch
	return &m
}

type TopicChannel struct {
	Topic   string
	Channel string
}

func (t *TopicChannel) String() string {
	return t.Topic + ":" + t.Channel
}
