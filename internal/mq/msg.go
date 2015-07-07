package mq

import "fmt"

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
