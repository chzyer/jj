package mq

import (
	"sync"

	"github.com/jj-io/jj/internal"

	"gopkg.in/logex.v1"
)

var (
	ErrMqClosing = logex.Define("mq is closing")
)

type Mq struct {
	topics  map[string]*Topic
	pubChan chan *Msg
	rat     *internal.Rat
	gruad   sync.Mutex
}

func NewMq() *Mq {
	mq := &Mq{
		topics:  make(map[string]*Topic),
		pubChan: make(chan *Msg, 10),
		rat:     internal.NewRat(),
	}
	go mq.pubLoop()
	return mq
}

func (m *Mq) pubLoop() {
	m.rat.Birth()
	defer m.rat.Kill()

	var msg *Msg
	for {
		select {
		case msg = <-m.pubChan:
		case <-m.rat.C:
			break
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
		topic = NewTopic(name, m.rat)
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

func (m *Mq) Publish(topic string, data []byte) error {
	select {
	case m.pubChan <- &Msg{
		Topic: topic,
		Data:  data,
	}:
	case <-m.rat.C:
		return ErrMqClosing
	}
	return nil
}

func (m *Mq) Close() {
	m.rat.Shoo()
}
