package mq

import (
	"reflect"
	"sync"
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

func (t *Topic) RemoveSubscriber(channel string, s Subscriber) {
	t.GetChan(channel).RemoveSubscriber(s)
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
	mutex      sync.Mutex
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

		ch.mutex.Lock()
		sc := make([]reflect.SelectCase, len(ch.selectCase))
		for i := 0; i < len(ch.selectCase); i++ {
			sc[i] = ch.selectCase[i]
			sc[i].Send = val
		}
		ch.mutex.Unlock()
		reflect.Select(sc)

	}
}

func (ch *Channel) AddSubscriber(s Subscriber) {
	ch.mutex.Lock()
	ch.subscriber = append(ch.subscriber, s)
	ch.selectCase = append(ch.selectCase, s.ToSelectCase())
	ch.mutex.Unlock()
}

func (ch *Channel) RemoveSubscriber(s Subscriber) (idx int) {
	ch.mutex.Lock()
	defer ch.mutex.Unlock()

	idx = -1
	for i := range ch.subscriber {
		if ch.subscriber[i] == s {
			idx = i
			break
		}
	}
	if idx < 0 {
		return
	}

	ch.subscriber = append(ch.subscriber[:idx], ch.subscriber[idx+1:]...)
	ch.selectCase = append(ch.selectCase[:idx], ch.selectCase[idx+1:]...)
	return
}

func (ch *Channel) Write(data []byte) {
	ch.underlay <- data
}
