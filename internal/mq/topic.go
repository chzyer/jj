package mq

import (
	"reflect"
	"sync/atomic"
)

type Subscriber interface {
	Name() string
	ToSelectCase() reflect.SelectCase
}

type Topic struct {
	Name       string
	buffer     chan []byte
	writeState int32
	hasChan    chan struct{}
	Chans      []*Channel
}

func NewTopic(name string) *Topic {
	t := &Topic{
		Name:    name,
		hasChan: make(chan struct{}),
		buffer:  make(chan []byte),
	}
	go t.writeBufferLoop()
	return t
}

func (t *Topic) AddSubscriber(channel string, s Subscriber) {
	t.GetChan(channel).AddSubscriber(s)
}

func (t *Topic) writeBufferLoop() {
	var (
		data []byte
	)
	<-t.hasChan
	for {
		select {
		case data = <-t.buffer:
			t.Publish(data)
		}
	}
}

func (t *Topic) ChanIdx(name string) int {
	for i := 0; i < len(t.Chans); i++ {
		if t.Chans[i].Name == name {
			return i
		}
	}
	return -1
}

func (t *Topic) GetChan(name string) (ch *Channel) {
	if idx := t.ChanIdx(name); idx < 0 {
		ch = NewChannel(t.Name, name)
		t.Chans = append(t.Chans, ch)
		if atomic.CompareAndSwapInt32(&t.writeState, 0, 1) {
			t.hasChan <- struct{}{}
		}
	} else {
		ch = t.Chans[idx]
	}
	return
}

func (t *Topic) Publish(data []byte) {
	if len(t.Chans) == 0 {
		t.buffer <- data
		return
	}
	for i := 0; i < len(t.Chans); i++ {
		t.Chans[i].Write(data)
	}
}

type Channel struct {
	Name       string
	Topic      string
	subscriber []Subscriber
	selectCase []reflect.SelectCase
	underlay   chan []byte
}

func NewChannel(topic, name string) *Channel {
	ch := &Channel{
		Name:     name,
		underlay: make(chan []byte, 10),
	}
	go ch.moveLoop()
	return ch
}

func (ch *Channel) moveLoop() {
	var (
		data []byte
	)
	for {
		select {
		case data = <-ch.underlay:

		}
		val := reflect.ValueOf(&Msg{
			Topic:   ch.Topic,
			channel: ch.Name,
			Data:    data,
		})

		for i := 0; i < len(ch.selectCase); i++ {
			ch.selectCase[i].Send = val
		}
		reflect.Select(ch.selectCase)
	}
}

func (ch *Channel) AddSubscriber(s Subscriber) {
	ch.subscriber = append(ch.subscriber, s)
	ch.selectCase = append(ch.selectCase, s.ToSelectCase())
}

func (ch *Channel) Write(data []byte) {
	ch.underlay <- data
}
